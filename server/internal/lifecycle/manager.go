package lifecycle

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/helpers"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/lobby"
	"github.com/xjhc/alignment/server/internal/store"
)

// Timeout constants for lobby cleanup
const (
	waitingForHostTimeout = 30 * time.Minute // Time to wait for host to connect
	emptyLobbyTimeout     = 10 * time.Minute // Time to wait before cleaning up empty lobbies
	staleLobbyTimeout     = 60 * time.Minute // Time since last activity before cleanup

	// Grace period for player reconnection
	playerReconnectionGracePeriod = 2 * time.Minute // Time to wait before abandoning disconnected players
)

// ActiveUserTrackerInterface defines the interface for tracking active users
type ActiveUserTrackerInterface interface {
	RemoveUserSessionByGameAndPlayer(gameID, playerID string)
}


// GameLifecycleManager unifies lobby and session management with event-driven architecture
type GameLifecycleManager struct {
	// Lobby management
	lobbies map[string]*lobby.Lobby
	tokens  map[string]*lobby.JoinToken

	// Game session management
	gameSessions map[string]map[string]interfaces.PlayerActorInterface
	gameActors   map[string]interfaces.GameActorInterface

	// Countdown management
	countdownTimers map[string]*time.Timer
	countdownCancel map[string]context.CancelFunc

	// Disconnected player tracking for grace period
	disconnectedPlayers map[string]*time.Timer // gameID:playerID -> grace period timer
	disconnectedCancel  map[string]context.CancelFunc // gameID:playerID -> cancel function

	// Synchronization
	mutex sync.RWMutex
	ctx   context.Context

	// Dependencies
	datastore         interfaces.DataStore
	broadcaster       interfaces.Broadcaster
	supervisor        interfaces.SupervisorInterface
	eventBus          *events.EventBus
	postgresStore     *store.PostgresStore
	activeUserTracker ActiveUserTrackerInterface

	// Event handling
	eventChannel chan events.Event
	stopChannel  chan struct{}
}

// NewGameLifecycleManager creates a new unified game lifecycle manager
func NewGameLifecycleManager(
	ctx context.Context,
	datastore interfaces.DataStore,
	broadcaster interfaces.Broadcaster,
	supervisor interfaces.SupervisorInterface,
	eventBus *events.EventBus,
	postgresStore *store.PostgresStore,
	activeUserTracker ActiveUserTrackerInterface,
) *GameLifecycleManager {
	glm := &GameLifecycleManager{
		lobbies:             make(map[string]*lobby.Lobby),
		tokens:              make(map[string]*lobby.JoinToken),
		gameSessions:        make(map[string]map[string]interfaces.PlayerActorInterface),
		gameActors:          make(map[string]interfaces.GameActorInterface),
		countdownTimers:     make(map[string]*time.Timer),
		countdownCancel:     make(map[string]context.CancelFunc),
		disconnectedPlayers: make(map[string]*time.Timer),
		disconnectedCancel:  make(map[string]context.CancelFunc),
		ctx:                 ctx,
		datastore:           datastore,
		broadcaster:         broadcaster,
		supervisor:          supervisor,
		eventBus:            eventBus,
		postgresStore:       postgresStore,
		activeUserTracker:   activeUserTracker,
		eventChannel:        make(chan events.Event, 100), // Buffered to prevent blocking
		stopChannel:         make(chan struct{}),
	}

	// Subscribe to relevant events
	eventBus.Subscribe("player_disconnected", glm.eventChannel)
	eventBus.Subscribe("game_ended", glm.eventChannel)

	// Start event processing goroutine with panic protection
	helpers.GoSafe(ctx, func(_ context.Context) { glm.processEvents() })

	// Start periodic cleanup goroutine for stale lobbies and games with panic protection
	helpers.GoSafe(ctx, func(_ context.Context) { glm.periodicCleanup() })

	return glm
}

// processEvents handles incoming events in a separate goroutine
func (glm *GameLifecycleManager) processEvents() {
	log.Println("GameLifecycleManager: Event processing started")
	defer log.Println("GameLifecycleManager: Event processing stopped")

	for {
		select {
		case <-glm.ctx.Done():
			return
		case <-glm.stopChannel:
			return
		case event := <-glm.eventChannel:
			glm.handleEvent(event)
		}
	}
}

// handleEvent processes individual events
func (glm *GameLifecycleManager) handleEvent(event events.Event) {
	switch e := event.(type) {
	case events.PlayerDisconnectedEvent:
		glm.handlePlayerDisconnected(e)
	case events.PlayerAbandonedGameEvent:
		glm.handlePlayerAbandoned(e)
	case events.GameEndedEvent:
		glm.handleGameEnded(e)
	default:
		log.Printf("GameLifecycleManager: Unknown event type: %T", event)
	}
}

// handlePlayerDisconnected handles player disconnection with grace period for games
func (glm *GameLifecycleManager) handlePlayerDisconnected(event events.PlayerDisconnectedEvent) {
	// Generate correlation ID for tracking this disconnection flow
	correlationID := uuid.New().String()[:8]
	log.Printf("[Disconnect-%s] handlePlayerDisconnected: Player %s disconnected from lobby=%s, game=%s", correlationID, event.PlayerID, event.LobbyID, event.GameID)

	// Remove from active user tracker
	if glm.activeUserTracker != nil {
		if event.LobbyID != "" {
			glm.activeUserTracker.RemoveUserSessionByGameAndPlayer(event.LobbyID, event.PlayerID)
			log.Printf("[Disconnect-%s] handlePlayerDisconnected: Removed player %s from active user tracker for lobby %s", correlationID, event.PlayerID, event.LobbyID)
		}
		if event.GameID != "" {
			glm.activeUserTracker.RemoveUserSessionByGameAndPlayer(event.GameID, event.PlayerID)
			log.Printf("[Disconnect-%s] handlePlayerDisconnected: Removed player %s from active user tracker for game %s", correlationID, event.PlayerID, event.GameID)
		}
	}

	// Handle lobby disconnection (immediate cleanup - no grace period for lobbies)
	if event.LobbyID != "" {
		log.Printf("[Disconnect-%s] handlePlayerDisconnected: Handling lobby disconnection", correlationID)
		glm.handleLobbyDisconnection(event)
	}

	// Handle game disconnection (with grace period)
	if event.GameID != "" {
		log.Printf("[Disconnect-%s] handlePlayerDisconnected: Handling game disconnection", correlationID)
		glm.handleGameDisconnection(event)
	}
	
	log.Printf("[Disconnect-%s] handlePlayerDisconnected: Completed disconnection handling for player %s", correlationID, event.PlayerID)
}

// handleLobbyDisconnection handles disconnection from lobbies (with grace period)
func (glm *GameLifecycleManager) handleLobbyDisconnection(event events.PlayerDisconnectedEvent) {
	correlationID := uuid.New().String()[:8]
	log.Printf("[Disconnect-%s] handleLobbyDisconnection: Processing lobby disconnect for player %s in lobby %s", correlationID, event.PlayerID, event.LobbyID)

	glm.mutex.RLock()
	lobby, exists := glm.lobbies[event.LobbyID]
	glm.mutex.RUnlock()

	if exists {
		// Mark player as disconnected but keep their slot - this will broadcast LOBBY_STATE_UPDATE
		lobby.DisconnectPlayer(event.PlayerID)
		log.Printf("[Disconnect-%s] handleLobbyDisconnection: Player %s marked as disconnected in lobby, state broadcasted", correlationID, event.PlayerID)

		// Publish player disconnected event (not left)
		glm.eventBus.Publish(events.PlayerDisconnectedFromLobbyEvent{
			PlayerID: event.PlayerID,
			LobbyID:  event.LobbyID,
		})

		// Start grace period timer
		glm.startLobbyDisconnectionGracePeriod(event.LobbyID, event.PlayerID)

		log.Printf("[Disconnect-%s] handleLobbyDisconnection: Player %s disconnected from lobby %s, starting grace period", correlationID, event.PlayerID, event.LobbyID)
	} else {
		log.Printf("[Disconnect-%s] handleLobbyDisconnection: Lobby %s not found", correlationID, event.LobbyID)
	}
}

// startLobbyDisconnectionGracePeriod starts a grace period timer for a disconnected lobby player
func (glm *GameLifecycleManager) startLobbyDisconnectionGracePeriod(lobbyID, playerID string) {
	gracePeriodKey := fmt.Sprintf("lobby:%s:%s", lobbyID, playerID)

	// Check if this player already has a grace period running
	glm.mutex.Lock()
	if _, exists := glm.disconnectedPlayers[gracePeriodKey]; exists {
		log.Printf("GameLifecycleManager: Grace period already running for player %s in lobby %s", playerID, lobbyID)
		glm.mutex.Unlock()
		return
	}
	glm.mutex.Unlock()

	// Create a cancellable context for this grace period
	ctx, cancel := context.WithCancel(glm.ctx)

	glm.mutex.Lock()
	glm.disconnectedCancel[gracePeriodKey] = cancel
	glm.mutex.Unlock()

	// Start the grace period timer in a goroutine with panic protection
	helpers.GoSafe(ctx, func(ctx context.Context) { glm.runLobbyGracePeriodTimer(ctx, lobbyID, playerID) })
}

// runLobbyGracePeriodTimer runs the lobby grace period timer and handles expiration
func (glm *GameLifecycleManager) runLobbyGracePeriodTimer(ctx context.Context, lobbyID, playerID string) {
	gracePeriodKey := fmt.Sprintf("lobby:%s:%s", lobbyID, playerID)

	defer func() {
		glm.mutex.Lock()
		delete(glm.disconnectedPlayers, gracePeriodKey)
		delete(glm.disconnectedCancel, gracePeriodKey)
		glm.mutex.Unlock()
	}()

	timer := time.NewTimer(playerReconnectionGracePeriod)
	defer timer.Stop()

	glm.mutex.Lock()
	glm.disconnectedPlayers[gracePeriodKey] = timer
	glm.mutex.Unlock()

	select {
	case <-ctx.Done():
		// Grace period was cancelled (player reconnected)
		log.Printf("GameLifecycleManager: Lobby grace period cancelled for player %s in lobby %s (reconnected)", playerID, lobbyID)
		return
	case <-timer.C:
		// Grace period expired - permanently remove the player
		log.Printf("GameLifecycleManager: Lobby grace period expired for player %s in lobby %s, removing permanently", playerID, lobbyID)
		glm.removePermanentlyFromLobby(lobbyID, playerID)
	}
}

// removePermanentlyFromLobby permanently removes a player from a lobby after grace period expires
func (glm *GameLifecycleManager) removePermanentlyFromLobby(lobbyID, playerID string) {
	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	glm.mutex.RUnlock()

	if !exists {
		return
	}

	// Check if the disconnecting player was the host
	wasHost := lobby.HostPlayerID == playerID

	// Permanently remove the player
	lobby.RemovePlayer(playerID)

	// Publish player left event
	glm.eventBus.Publish(events.PlayerLeftLobbyEvent{
		PlayerID: playerID,
		LobbyID:  lobbyID,
	})

	// Handle host transfer if needed
	if wasHost && len(lobby.GetPlayerActors()) > 0 {
		lobby.Lock()
		newHostID := lobby.TransferHostToNextPlayer()
		lobby.Unlock()

		if newHostID != "" {
			log.Printf("GameLifecycleManager: Transferred host from %s to %s in lobby %s after grace period", playerID, newHostID, lobbyID)

			// Broadcast host transfer event to all players in lobby
			hostTransferEvent := core.Event{
				ID:        uuid.New().String(),
				Type:      core.EventHostTransferred,
				GameID:    lobbyID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"previous_host_id": playerID,
					"new_host_id":      newHostID,
				},
			}

			playerActors := lobby.GetPlayerActors()
			for _, actor := range playerActors {
				actor.SendServerMessage(hostTransferEvent)
			}
		}
	}

	// Check if lobby should be cleaned up
	glm.checkLobbyCleanup(lobbyID, lobby)
}

// checkLobbyCleanup checks if a lobby needs to be cleaned up and does so if necessary
func (glm *GameLifecycleManager) checkLobbyCleanup(lobbyID string, lobby *lobby.Lobby) {
	// If countdown is running and lobby can no longer start, cancel countdown
	glm.mutex.RLock()
	cancelFunc, countdownRunning := glm.countdownCancel[lobbyID]
	glm.mutex.RUnlock()

	if countdownRunning && !lobby.CanStart() {
		log.Printf("GameLifecycleManager: Cancelling countdown for lobby %s due to insufficient players", lobbyID)
		cancelFunc()
	}

	// Check if lobby is now empty and should be cleaned up
	glm.mutex.Lock()
	if len(lobby.GetPlayerActors()) == 0 {
		// Cancel any running countdown for empty lobby
		if cancelFunc, exists := glm.countdownCancel[lobbyID]; exists {
			cancelFunc()
			delete(glm.countdownCancel, lobbyID)
			delete(glm.countdownTimers, lobbyID)
		}
		delete(glm.lobbies, lobbyID)
		log.Printf("GameLifecycleManager: Cleaned up empty lobby %s", lobbyID)
	}
	glm.mutex.Unlock()
}

// handleGameDisconnection handles disconnection from games (with grace period)
func (glm *GameLifecycleManager) handleGameDisconnection(event events.PlayerDisconnectedEvent) {
	correlationID := uuid.New().String()[:8]
	gracePeriodKey := fmt.Sprintf("%s:%s", event.GameID, event.PlayerID)

	log.Printf("[Disconnect-%s] handleGameDisconnection: Processing game disconnect for player %s in game %s", correlationID, event.PlayerID, event.GameID)

	// Check if this player is already in a grace period (duplicate disconnect event)
	glm.mutex.RLock()
	if _, exists := glm.disconnectedPlayers[gracePeriodKey]; exists {
		glm.mutex.RUnlock()
		log.Printf("[Disconnect-%s] handleGameDisconnection: Player %s already in grace period for game %s", correlationID, event.PlayerID, event.GameID)
		return
	}
	glm.mutex.RUnlock()

	// Remove the player from the active game session (but keep them in the game state)
	glm.mutex.Lock()
	if session, exists := glm.gameSessions[event.GameID]; exists {
		delete(session, event.PlayerID)
		log.Printf("[Disconnect-%s] handleGameDisconnection: Removed player %s from game session %s (grace period started)", correlationID, event.PlayerID, event.GameID)
	}
	glm.mutex.Unlock()

	// Notify the GameActor about the disconnection (set connection status to DISCONNECTED)
	glm.mutex.RLock()
	gameActor, exists := glm.gameActors[event.GameID]
	glm.mutex.RUnlock()

	if exists {
		log.Printf("[Disconnect-%s] handleGameDisconnection: Found GameActor, dispatching connection status change", correlationID)
		
		// Send SET_PLAYER_CONNECTION_STATUS action to GameActor
		disconnectAction := core.Action{
			Type:     core.ActionSetPlayerConnectionStatus,
			PlayerID: event.PlayerID,
			GameID:   event.GameID,
			Payload: map[string]interface{}{
				"connection_status": "DISCONNECTED",
			},
		}

		// Post action synchronously to ensure the status is updated and broadcasted
		log.Printf("[Disconnect-%s] handleGameDisconnection: Posting SET_PLAYER_CONNECTION_STATUS action", correlationID)
		resultChan := gameActor.PostAction(disconnectAction)
		result := <-resultChan
		if result.Error != nil {
			log.Printf("[Disconnect-%s] handleGameDisconnection: ERROR - Failed to set player connection status: %v", correlationID, result.Error)
		} else {
			log.Printf("[Disconnect-%s] handleGameDisconnection: Successfully set connection status to DISCONNECTED and broadcasted to other players", correlationID)
		}

		// Start grace period timer
		glm.startGracePeriodTimer(event.GameID, event.PlayerID)
	} else {
		log.Printf("[Disconnect-%s] handleGameDisconnection: GameActor not found for game %s", correlationID, event.GameID)
	}
}

// startGracePeriodTimer starts a grace period timer for a disconnected player
func (glm *GameLifecycleManager) startGracePeriodTimer(gameID, playerID string) {
	gracePeriodKey := fmt.Sprintf("%s:%s", gameID, playerID)

	// Create grace period context
	gracePeriodCtx, cancel := context.WithCancel(glm.ctx)

	glm.mutex.Lock()
	glm.disconnectedCancel[gracePeriodKey] = cancel
	glm.mutex.Unlock()

	log.Printf("GameLifecycleManager: Starting %v grace period for player %s in game %s",
		playerReconnectionGracePeriod, playerID, gameID)

	// Start grace period timer in a goroutine with panic protection
	helpers.GoSafe(gracePeriodCtx, func(ctx context.Context) { glm.runGracePeriodTimer(ctx, gameID, playerID) })
}

// runGracePeriodTimer runs the grace period timer and handles expiration
func (glm *GameLifecycleManager) runGracePeriodTimer(ctx context.Context, gameID, playerID string) {
	gracePeriodKey := fmt.Sprintf("%s:%s", gameID, playerID)

	defer func() {
		glm.mutex.Lock()
		delete(glm.disconnectedPlayers, gracePeriodKey)
		delete(glm.disconnectedCancel, gracePeriodKey)
		glm.mutex.Unlock()
	}()

	timer := time.NewTimer(playerReconnectionGracePeriod)
	defer timer.Stop()

	glm.mutex.Lock()
	glm.disconnectedPlayers[gracePeriodKey] = timer
	glm.mutex.Unlock()

	select {
	case <-ctx.Done():
		// Grace period was cancelled (player reconnected)
		log.Printf("GameLifecycleManager: Grace period cancelled for player %s in game %s (reconnected)", playerID, gameID)
		return
	case <-timer.C:
		// Grace period expired - abandon the player
		log.Printf("GameLifecycleManager: Grace period expired for player %s in game %s, abandoning", playerID, gameID)
		glm.abandonDisconnectedPlayer(gameID, playerID)
	}
}

// abandonDisconnectedPlayer abandons a player whose grace period has expired
func (glm *GameLifecycleManager) abandonDisconnectedPlayer(gameID, playerID string) {
	// Notify the GameActor to abandon the player
	glm.mutex.RLock()
	gameActor, exists := glm.gameActors[gameID]
	session := glm.gameSessions[gameID]
	glm.mutex.RUnlock()

	if exists {
		// Send ABANDON_PLAYER action to GameActor
		abandonAction := core.Action{
			Type:     core.ActionAbandonPlayer,
			PlayerID: playerID,
			GameID:   gameID,
			Payload: map[string]interface{}{
				"reason": "grace_period_expired",
			},
		}

		// Post action asynchronously with panic protection
		helpers.GoSafe(glm.ctx, func(_ context.Context) {
			resultChan := gameActor.PostAction(abandonAction)
			result := <-resultChan
			if result.Error != nil {
				log.Printf("GameLifecycleManager: Error abandoning player: %v", result.Error)
			}
		})

		// Check if this was the last player out and trigger garbage collection
		if session != nil && len(session) == 0 {
			log.Printf("GameLifecycleManager: Last player left game %s, triggering garbage collection", gameID)
			glm.mutex.Lock()
			delete(glm.gameSessions, gameID)
			delete(glm.gameActors, gameID)
			glm.mutex.Unlock()

			// Stop the GameActor via supervisor (safer than calling Stop directly)
			if glm.supervisor != nil {
				glm.supervisor.RemoveGame(gameID)
			}

			// Publish game ended event for any other cleanup
			glm.eventBus.Publish(events.GameEndedEvent{
				GameID: gameID,
				Reason: "abandoned", // All players left
			})
		}
	}
}

// cancelGracePeriod cancels the grace period timer for a reconnecting player
func (glm *GameLifecycleManager) cancelGracePeriod(gameID, playerID string) {
	gracePeriodKey := fmt.Sprintf("%s:%s", gameID, playerID)

	glm.mutex.Lock()
	if cancelFunc, exists := glm.disconnectedCancel[gracePeriodKey]; exists {
		cancelFunc()
		delete(glm.disconnectedCancel, gracePeriodKey)
		delete(glm.disconnectedPlayers, gracePeriodKey)
		log.Printf("GameLifecycleManager: Cancelled grace period for player %s in game %s", playerID, gameID)
	}
	glm.mutex.Unlock()
}

// handlePlayerAbandoned removes abandoned players from game sessions (similar to disconnect but for active abandonment)
func (glm *GameLifecycleManager) handlePlayerAbandoned(event events.PlayerAbandonedGameEvent) {
	log.Printf("GameLifecycleManager: Handling player abandonment: %s from game %s", event.PlayerID, event.GameID)

	// Remove from active user tracker
	if glm.activeUserTracker != nil {
		glm.activeUserTracker.RemoveUserSessionByGameAndPlayer(event.GameID, event.PlayerID)
	}

	// Remove from game session tracking
	glm.mutex.Lock()
	if session, exists := glm.gameSessions[event.GameID]; exists {
		delete(session, event.PlayerID)
		log.Printf("GameLifecycleManager: Removed abandoned player %s from game session %s", event.PlayerID, event.GameID)

		// Check if game session is now empty
		if len(session) == 0 {
			log.Printf("GameLifecycleManager: Game session %s is now empty after abandonment, cleaning up", event.GameID)
			delete(glm.gameSessions, event.GameID)

			// Publish game ended event if no players remain
			if glm.eventBus != nil {
				glm.eventBus.Publish(events.GameEndedEvent{
					GameID: event.GameID,
					Reason: "all_players_abandoned",
				})
			}
		}
	}
	glm.mutex.Unlock()

	log.Printf("GameLifecycleManager: Successfully handled abandonment for player %s", event.PlayerID)
}

// ForceLogoutPlayer publishes a force logout event for a player due to unrecoverable state
func (glm *GameLifecycleManager) ForceLogoutPlayer(playerID string, reason string) {
	if glm.eventBus == nil {
		log.Printf("GameLifecycleManager: Warning - EventBus not available, cannot force logout for player %s", playerID)
		return
	}

	log.Printf("GameLifecycleManager: Publishing force logout event for player %s: %s", playerID, reason)
	glm.eventBus.Publish(events.ForceLogoutEvent{
		PlayerID: playerID,
		Reason:   reason,
	})
}

// handleGameEnded cleans up finished games
func (glm *GameLifecycleManager) handleGameEnded(event events.GameEndedEvent) {
	log.Printf("GameLifecycleManager: Handling game end: %s (reason: %s)", event.GameID, event.Reason)

	glm.mutex.Lock()
	defer glm.mutex.Unlock()

	// Clean up game session
	delete(glm.gameSessions, event.GameID)
	delete(glm.gameActors, event.GameID)

	log.Printf("GameLifecycleManager: Cleaned up game %s", event.GameID)
}

// periodicCleanup runs periodically to clean up stale lobbies and sessions
func (glm *GameLifecycleManager) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Run cleanup every 5 minutes
	defer ticker.Stop()

	log.Println("GameLifecycleManager: Periodic cleanup started")
	defer log.Println("GameLifecycleManager: Periodic cleanup stopped")

	for {
		select {
		case <-glm.ctx.Done():
			return
		case <-glm.stopChannel:
			return
		case <-ticker.C:
			glm.cleanupStaleLobbies()
			glm.cleanupStaleGameSessions()
			glm.cleanupExpiredTokens()
		}
	}
}

// cleanupStaleLobbies removes lobbies that have been inactive for too long
func (glm *GameLifecycleManager) cleanupStaleLobbies() {
	const maxInactivityDuration = 30 * time.Minute // Clean up lobbies inactive for 30+ minutes
	const maxEmptyDuration = 2 * time.Minute       // Clean up empty lobbies after 2 minutes

	now := time.Now()
	var staleLobbyIDs []string

	glm.mutex.RLock()
	for lobbyID, lobby := range glm.lobbies {
		lobby.RLock()
		playerCount := len(lobby.Players)
		status := lobby.Status
		createdAt := lobby.CreatedAt
		lastActivity := lobby.LastActivity
		lobby.RUnlock()

		// CRITERIA FOR DELETION:
		shouldDelete := false
		reason := ""

		// 1. Lobby is waiting for host for too long
		if status == "WAITING_FOR_HOST" && now.Sub(createdAt) > waitingForHostTimeout {
			shouldDelete = true
			reason = "waiting for host timeout"
		}
		// 2. Lobby is empty and has existed for a while
		if playerCount == 0 && now.Sub(createdAt) > emptyLobbyTimeout {
			shouldDelete = true
			reason = "empty lobby timeout"
		}
		// 3. Lobby has seen no activity (joins/leaves) for a long time, regardless of player count
		if now.Sub(lastActivity) > staleLobbyTimeout {
			shouldDelete = true
			reason = "no recent activity"
		}

		if shouldDelete {
			staleLobbyIDs = append(staleLobbyIDs, lobbyID)
			log.Printf("[GC] Marking lobby %s for removal: %s (players: %d, created: %v ago, last activity: %v ago)",
				lobbyID, reason, playerCount, now.Sub(createdAt), now.Sub(lastActivity))
		}
	}
	glm.mutex.RUnlock()

	// Remove stale lobbies
	if len(staleLobbyIDs) > 0 {
		glm.mutex.Lock()
		for _, lobbyID := range staleLobbyIDs {
			if lobby, exists := glm.lobbies[lobbyID]; exists {
				// Notify remaining players that the lobby is being closed
				closeEvent := core.Event{
					ID:        fmt.Sprintf("lobby_closed_%d", time.Now().UnixNano()),
					Type:      "LOBBY_CLOSED",
					GameID:    lobbyID,
					Timestamp: time.Now(),
					Payload: map[string]interface{}{
						"reason": "inactivity",
						"message": "Lobby closed due to inactivity",
					},
				}

				players := lobby.GetPlayerActors()
				for _, actor := range players {
					actor.SendServerMessage(closeEvent)
					// Transition players back to idle state
					actor.TransitionToIdle()
				}

				// Cancel any running countdown
				if cancelFunc, exists := glm.countdownCancel[lobbyID]; exists {
					cancelFunc()
					delete(glm.countdownCancel, lobbyID)
					delete(glm.countdownTimers, lobbyID)
				}

				delete(glm.lobbies, lobbyID)
			}
		}
		glm.mutex.Unlock()
		log.Printf("GameLifecycleManager: Cleaned up %d stale lobbies", len(staleLobbyIDs))
	}
}

// cleanupStaleGameSessions removes game sessions that might be leaking
func (glm *GameLifecycleManager) cleanupStaleGameSessions() {
	var staleGameIDs []string

	glm.mutex.RLock()
	for gameID, session := range glm.gameSessions {
		// Check if there's no corresponding GameActor (orphaned session)
		if _, hasActor := glm.gameActors[gameID]; !hasActor {
			staleGameIDs = append(staleGameIDs, gameID)
			log.Printf("GameLifecycleManager: Found orphaned game session %s (no corresponding GameActor)", gameID)
			continue
		}

		// For sessions with GameActors, we can't easily determine creation time
		// without adding more tracking, so we rely on the GameActor's own health monitoring
		_ = session // Keep the session alive if it has a GameActor
	}
	glm.mutex.RUnlock()

	// Clean up orphaned sessions
	if len(staleGameIDs) > 0 {
		glm.mutex.Lock()
		for _, gameID := range staleGameIDs {
			delete(glm.gameSessions, gameID)
		}
		glm.mutex.Unlock()
		log.Printf("GameLifecycleManager: Cleaned up %d orphaned game sessions", len(staleGameIDs))
	}
}

// CreateLobbyViaHTTP creates a lobby via HTTP and returns join credentials
func (glm *GameLifecycleManager) CreateLobbyViaHTTP(userID, hostPlayerName, lobbyName, playerAvatar string, isPrivate bool) (string, string, string, error) {
	lobbyID := uuid.New().String()
	// Use the userID as the playerID to ensure consistency
	hostPlayerID := userID

	glm.mutex.Lock()
	defer glm.mutex.Unlock()

	// Generate session token
	sessionToken, err := glm.generateSessionTokenWithLobbyInfo_unsafe(lobbyID, hostPlayerID, hostPlayerName, playerAvatar, lobbyName, true, isPrivate)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create the lobby immediately (not waiting for WebSocket connection)
	// This ensures it appears in the lobby list right away
	if lobbyName == "" {
		lobbyName = hostPlayerName + "'s Game"
	}

	// Create a placeholder lobby without a PlayerActor (will be added when host connects)
	newLobby := &lobby.Lobby{
		ID:           lobbyID,
		Name:         lobbyName,
		HostPlayerID: hostPlayerID,
		Players:      make(map[string]*lobby.PlayerInfo),
		PlayerActors: make(map[string]interfaces.PlayerActorInterface),
		MaxPlayers:   8,
		MinPlayers:   2,
		CreatedAt:    time.Now(),
		Status:       "WAITING_FOR_HOST",
		IsPrivate:    isPrivate,
	}

	glm.lobbies[lobbyID] = newLobby

	// Publish lobby created event
	glm.eventBus.Publish(events.LobbyCreatedEvent{
		LobbyID:      lobbyID,
		HostPlayerID: hostPlayerID,
		LobbyName:    lobbyName,
	})

	log.Printf("GameLifecycleManager: Created lobby %s for host %s (waiting for connection)", lobbyID, hostPlayerID)

	return lobbyID, hostPlayerID, sessionToken, nil
}

// JoinLobbyWithActor adds a player actor to a lobby, creating it if needed
func (glm *GameLifecycleManager) JoinLobbyWithActor(lobbyID string, playerActor interfaces.PlayerActorInterface) error {
	// Generate correlation ID for tracking this connection flow
	correlationID := uuid.New().String()[:8]
	playerID := playerActor.GetPlayerID()
	
	log.Printf("[Connection-%s] JoinLobbyWithActor: Starting join process for player %s to lobby %s", correlationID, playerID, lobbyID)
	
	glm.mutex.Lock()
	targetLobby, exists := glm.lobbies[lobbyID]

	if !exists {
		// This should not happen in the regular flow anymore, as the lobby
		// is created via HTTP first. But as a safeguard:
		glm.mutex.Unlock()
		log.Printf("[Connection-%s] JoinLobbyWithActor: FAILED - lobby not found: %s", correlationID, lobbyID)
		return fmt.Errorf("lobby not found: %s", lobbyID)
	}
	
	log.Printf("[Connection-%s] JoinLobbyWithActor: Found existing lobby %s", correlationID, lobbyID)

	// Lock the specific lobby for state changes
	targetLobby.Lock()

	// Check if this is the host connecting for the first time
	if targetLobby.Status == "WAITING_FOR_HOST" && targetLobby.HostPlayerID == playerID {
		// Host is connecting - transition lobby from placeholder to active
		targetLobby.Status = "WAITING"
		log.Printf("[Connection-%s] JoinLobbyWithActor: Host %s connected, lobby %s is now active", correlationID, playerID, lobbyID)
	} else if targetLobby.Status == "WAITING_FOR_HOST" {
		// If another player tries to join before the host, reject them.
		targetLobby.Unlock()
		glm.mutex.Unlock()
		log.Printf("[Connection-%s] JoinLobbyWithActor: FAILED - lobby %s is waiting for host, rejecting player %s", correlationID, lobbyID, playerID)
		return fmt.Errorf("lobby is not accepting new players yet")
	}
	targetLobby.Unlock() // Unlock the lobby after status check/update
	glm.mutex.Unlock() // Unlock the manager after getting the lobby ref

	log.Printf("[Connection-%s] JoinLobbyWithActor: Status validated, checking for blocked players", correlationID)

	// Check for blocked players before allowing the join
	err := glm.checkForBlockedPlayers(playerActor.GetPlayerID(), targetLobby)
	if err != nil {
		log.Printf("[Connection-%s] JoinLobbyWithActor: FAILED - blocked player check failed: %v", correlationID, err)
		return err
	}

	log.Printf("[Connection-%s] JoinLobbyWithActor: Blocked player check passed", correlationID)

	// Check if this is a reconnection (player exists but is disconnected)
	if targetLobby.IsPlayerDisconnected(playerID) {
		log.Printf("[Connection-%s] JoinLobbyWithActor: Player %s is reconnecting to lobby %s", correlationID, playerID, lobbyID)

		// Cancel the grace period timer
		gracePeriodKey := fmt.Sprintf("lobby:%s:%s", lobbyID, playerID)
		glm.mutex.Lock()
		if cancelFunc, exists := glm.disconnectedCancel[gracePeriodKey]; exists {
			cancelFunc()
			delete(glm.disconnectedCancel, gracePeriodKey)
			delete(glm.disconnectedPlayers, gracePeriodKey)
			log.Printf("[Connection-%s] JoinLobbyWithActor: Cancelled grace period for reconnecting player %s in lobby %s", correlationID, playerID, lobbyID)
		}
		glm.mutex.Unlock()

		// Reconnect the player - this will automatically broadcast state update to all players
		targetLobby.ReconnectPlayer(playerID, playerActor)
		log.Printf("[Connection-%s] JoinLobbyWithActor: Player %s reconnected to lobby, state broadcasted", correlationID, playerID)

		// Transition player actor to lobby state
		err := playerActor.TransitionToLobby(lobbyID)
		if err != nil {
			log.Printf("[Connection-%s] JoinLobbyWithActor: FAILED - transition to lobby failed: %v", correlationID, err)
			return err
		}
		
		log.Printf("[Connection-%s] JoinLobbyWithActor: SUCCESS - Player %s reconnected to lobby %s and received state", correlationID, playerID, lobbyID)
		return nil
	}

	log.Printf("[Connection-%s] JoinLobbyWithActor: Adding new player %s to lobby %s", correlationID, playerID, lobbyID)

	// Add player to lobby (this will handle its own locking and broadcasting)
	err = targetLobby.AddPlayer(playerActor)
	if err != nil {
		log.Printf("[Connection-%s] JoinLobbyWithActor: FAILED - AddPlayer failed: %v", correlationID, err)
		return err
	}

	log.Printf("[Connection-%s] JoinLobbyWithActor: Player %s added to lobby, transitioning to lobby state", correlationID, playerID)

	// Transition player actor to lobby state
	err = playerActor.TransitionToLobby(lobbyID)
	if err != nil {
		log.Printf("[Connection-%s] JoinLobbyWithActor: FAILED - transition to lobby failed: %v", correlationID, err)
		return err
	}
	
	log.Printf("[Connection-%s] JoinLobbyWithActor: SUCCESS - Player %s joined lobby %s", correlationID, playerID, lobbyID)
	return nil
}

// StartGame initiates a 3-second countdown before starting the game
func (glm *GameLifecycleManager) StartGame(lobbyID string, hostPlayerID string) error {
	log.Printf("GameLifecycleManager: Starting countdown for lobby %s from host %s", lobbyID, hostPlayerID)

	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	if !exists {
		glm.mutex.RUnlock()
		return fmt.Errorf("lobby not found")
	}
	glm.mutex.RUnlock()

	// Verify host first
	if lobby.HostPlayerID != hostPlayerID {
		return fmt.Errorf("only the host can start the game")
	}

	// Check if lobby can start
	if !lobby.CanStart() {
		return fmt.Errorf("lobby cannot start: not enough players or invalid state")
	}

	// Check if countdown is already running
	glm.mutex.Lock()
	if _, running := glm.countdownTimers[lobbyID]; running {
		glm.mutex.Unlock()
		return fmt.Errorf("game start countdown already in progress")
	}
	glm.mutex.Unlock()

	// Start the countdown
	return glm.startCountdown(lobbyID)
}

// startCountdown begins a 3-second countdown and broadcasts updates
func (glm *GameLifecycleManager) startCountdown(lobbyID string) error {
	lobby, exists := glm.lobbies[lobbyID]
	if !exists {
		return fmt.Errorf("lobby not found during countdown start")
	}

	// Set lobby status to countdown
	lobby.SetStatus("COUNTDOWN")

	// Create countdown context
	countdownCtx, cancel := context.WithCancel(glm.ctx)

	glm.mutex.Lock()
	glm.countdownCancel[lobbyID] = cancel
	glm.mutex.Unlock()

	// Broadcast countdown initiated event
	playerActors := lobby.GetPlayerActors()
	countdownEvent := core.Event{
		ID:        uuid.New().String(),
		Type:      core.EventGameStartCountdownStart,
		GameID:    lobbyID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"duration": 3,
		},
	}

	for _, actor := range playerActors {
		actor.SendServerMessage(countdownEvent)
	}

	// Start countdown timer in a goroutine with panic protection
	helpers.GoSafe(countdownCtx, func(ctx context.Context) { glm.runCountdown(ctx, lobbyID, 3) })

	return nil
}

// runCountdown handles the countdown timer and broadcasts updates
func (glm *GameLifecycleManager) runCountdown(ctx context.Context, lobbyID string, duration int) {
	defer func() {
		glm.mutex.Lock()
		delete(glm.countdownTimers, lobbyID)
		delete(glm.countdownCancel, lobbyID)
		glm.mutex.Unlock()
	}()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	remaining := duration

	for remaining > 0 {
		select {
		case <-ctx.Done():
			// Countdown was cancelled
			log.Printf("GameLifecycleManager: Countdown cancelled for lobby %s", lobbyID)
			glm.broadcastCountdownCancel(lobbyID)
			return
		case <-ticker.C:
			remaining--

			// Broadcast countdown update
			glm.broadcastCountdownUpdate(lobbyID, remaining)

			if remaining == 0 {
				// Countdown complete - start the actual game
				glm.finalizeGameStart(lobbyID)
				return
			}
		}
	}
}

// broadcastCountdownUpdate sends countdown update to all players in lobby
func (glm *GameLifecycleManager) broadcastCountdownUpdate(lobbyID string, remaining int) {
	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	glm.mutex.RUnlock()

	if !exists {
		return
	}

	updateEvent := core.Event{
		ID:        uuid.New().String(),
		Type:      core.EventGameStartCountdownUpdate,
		GameID:    lobbyID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"remaining": remaining,
		},
	}

	playerActors := lobby.GetPlayerActors()
	for _, actor := range playerActors {
		actor.SendServerMessage(updateEvent)
	}
}

// broadcastCountdownCancel sends countdown cancellation to all players in lobby
func (glm *GameLifecycleManager) broadcastCountdownCancel(lobbyID string) {
	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	glm.mutex.RUnlock()

	if !exists {
		return
	}

	// Reset lobby status
	lobby.SetStatus("WAITING")

	cancelEvent := core.Event{
		ID:        uuid.New().String(),
		Type:      core.EventGameStartCountdownCancel,
		GameID:    lobbyID,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{},
	}

	playerActors := lobby.GetPlayerActors()
	for _, actor := range playerActors {
		actor.SendServerMessage(cancelEvent)
	}
}

// finalizeGameStart completes the game start process after countdown
func (glm *GameLifecycleManager) finalizeGameStart(lobbyID string) {
	log.Printf("GameLifecycleManager: Finalizing game start for lobby %s", lobbyID)

	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	glm.mutex.RUnlock()

	if !exists {
		log.Printf("GameLifecycleManager: Lobby %s not found during finalization", lobbyID)
		return
	}

	// Final check - ensure lobby still has enough players
	playerCount := len(lobby.GetPlayerActors())
	status := lobby.Status
	minPlayers := lobby.MinPlayers
	log.Printf("GameLifecycleManager: Lobby %s final check - players: %d, minPlayers: %d, status: %s",
		lobbyID, playerCount, minPlayers, status)

	if !lobby.CanStart() {
		log.Printf("GameLifecycleManager: Lobby %s no longer eligible to start (players: %d/%d, status: %s)",
			lobbyID, playerCount, minPlayers, status)
		glm.broadcastCountdownCancel(lobbyID)
		return
	}

	// Mark as starting
	lobby.SetStatus("STARTING")

	// Copy the players for game creation
	playerActors := lobby.GetPlayerActors()

	// Create the game
	err := glm.createGameFromLobby(lobbyID, playerActors)
	if err != nil {
		// Revert lobby state on failure
		lobby.SetStatus("WAITING")
		log.Printf("GameLifecycleManager: Failed to create game for lobby %s: %v", lobbyID, err)
		return
	}

	// Remove lobby from manager
	glm.mutex.Lock()
	delete(glm.lobbies, lobbyID)
	glm.mutex.Unlock()

	log.Printf("GameLifecycleManager: Successfully started game for lobby %s", lobbyID)
}

// createGameFromLobby handles the atomic transition from lobby to game
func (glm *GameLifecycleManager) createGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	log.Printf("GameLifecycleManager: Creating game from lobby %s with %d players", lobbyID, len(playerActors))

	// Create temporary game state to get default starting tokens
	tempGameState := core.NewGameState(lobbyID, time.Now())
	startingTokens := tempGameState.Settings.StartingTokens

	// Convert PlayerActors to core.Players map
	players := make(map[string]*core.Player)
	currentTime := time.Now()

	for playerID, actor := range playerActors {
		players[playerID] = &core.Player{
			ID:                playerID,
			Name:              actor.GetPlayerName(),
			JobTitle:          "", // Will be set during role assignment
			ControlType:       "HUMAN",
			Status:            core.PlayerStatusAlive,
			IsAlive:           true,
			ConnectionStatus:  "CONNECTED", // Initial connection status
			Tokens:            startingTokens, // Use game settings for starting tokens
			ProjectMilestones: 0,
			StatusMessage:     "",
			JoinedAt:          currentTime,
			Alignment:         "HUMAN", // Default alignment before role assignment
		}
	}

	// Add the AI player
	aiPlayerID := "ai-nexus-" + uuid.New().String()[:8]
	players[aiPlayerID] = &core.Player{
		ID:                aiPlayerID,
		Name:              "NEXUS",
		JobTitle:          "AI Assistant",
		ControlType:       "AI",
		Status:            core.PlayerStatusAlive,
		IsAlive:           true,
		ConnectionStatus:  "CONNECTED", // AI is always connected
		Tokens:            startingTokens, // AI also uses game settings for starting tokens
		ProjectMilestones: 0,
		StatusMessage:     "",
		JoinedAt:          currentTime,
		Alignment:         "AI", // Start with AI alignment
	}

	gameID := lobbyID // The lobby ID becomes the game ID

	// Create GameActor via Supervisor
	gameActor, err := glm.supervisor.CreateGameWithPlayers(gameID, players)
	if err != nil {
		return fmt.Errorf("failed to create game actor: %w", err)
	}

	// Store game session and actor
	glm.mutex.Lock()
	glm.gameSessions[gameID] = playerActors
	glm.gameActors[gameID] = gameActor
	glm.mutex.Unlock()

	// --- START FIX: Synchronous Initialization ---

	// 1. Send an INITIALIZE_GAME action to the new actor and wait for the response.
	// This ensures the GameActor's internal state (roles, phase) is set *before* we proceed.
	initAction := core.Action{
		Type:     core.ActionType("INITIALIZE_GAME"),
		GameID:   gameID,
		PlayerID: "SYSTEM",
	}

	responseChan := gameActor.PostAction(initAction)
	var initialEvents []core.Event
	select {
	case result := <-responseChan:
		if result.Error != nil {
			return fmt.Errorf("failed to initialize game actor: %w", result.Error)
		}
		initialEvents = result.Events
	case <-time.After(5 * time.Second): // Add a timeout to prevent hanging
		return fmt.Errorf("timeout waiting for game actor initialization")
	}

	// Persist all events generated during game initialization
	for _, event := range initialEvents {
		if err := glm.datastore.AppendEvent(glm.ctx, gameID, event); err != nil {
			log.Printf("CRITICAL: Failed to persist event %s for %s: %v", event.ID, gameID, err)
			// Don't fail the whole process, but log critically
		}
	}

	// Assign corporate mandate for this game
	mandateAction := core.Action{
		Type:     core.ActionType("ASSIGN_CORPORATE_MANDATE"),
		GameID:   gameID,
		PlayerID: "SYSTEM",
	}
	mandateResponseChan := gameActor.PostAction(mandateAction)
	select {
	case result := <-mandateResponseChan:
		if result.Error != nil {
			log.Printf("WARNING: Failed to assign corporate mandate for game %s: %v", gameID, result.Error)
			// Don't fail the whole process, but log the warning
		} else {
			// Persist mandate assignment events
			for _, event := range result.Events {
				if err := glm.datastore.AppendEvent(glm.ctx, gameID, event); err != nil {
					log.Printf("CRITICAL: Failed to persist mandate event %s for %s: %v", event.ID, gameID, err)
				}
			}
		}
	case <-time.After(2 * time.Second):
		log.Printf("WARNING: Timeout waiting for corporate mandate assignment for game %s", gameID)
		// Don't fail the whole process, continue without mandate
	}

	// 2. Now that the GameActor's state is fully initialized, we can safely
	//    transition players and send them the correct snapshot.
	for playerID, playerActor := range playerActors {
		// Transition the player actor's internal state
		err := playerActor.TransitionToGame(gameID)
		if err != nil {
			log.Printf("GameLifecycleManager: Failed to transition player %s to game: %v", playerID, err)
			continue // Skip to next player
		}

		// Send GAME_STARTED event to signal UI transition
		gameStartedEvent := core.Event{
			ID:        uuid.New().String(),
			Type:      core.EventGameStarted,
			GameID:    gameID,
			PlayerID:  playerID, // Private event for this player
			Timestamp: time.Now(),
			Payload:   map[string]interface{}{"game_id": gameID},
		}
		playerActor.SendServerMessage(gameStartedEvent)

		// Send the correctly initialized snapshot.
		// This snapshot now contains the correct phase (SITREP) and role data.
		snapshotEvent := gameActor.CreatePlayerStateUpdateEvent(playerID)
		playerActor.SendServerMessage(snapshotEvent)
	}

	// 3. REMOVED: Do not send the granular events again. The snapshot is sufficient.
	//    The snapshot already contains the state resulting from these events.

	// --- END FIX ---

	// Publish game started event (external listeners)
	playerIDs := make([]string, 0, len(playerActors))
	for id := range playerActors {
		playerIDs = append(playerIDs, id)
	}

	glm.eventBus.Publish(events.GameStartedEvent{
		GameID:    gameID,
		LobbyID:   lobbyID,
		PlayerIDs: playerIDs,
	})

	log.Printf("GameLifecycleManager: Successfully created and started game %s", gameID)
	return nil
}


// generateSessionTokenWithLobbyInfo creates a session token for a player in a lobby
func (glm *GameLifecycleManager) generateSessionTokenWithLobbyInfo(lobbyID, playerID, playerName, playerAvatar, lobbyName string, isHost, isPrivate bool) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	joinToken := &lobby.JoinToken{
		Token:        token,
		LobbyID:      lobbyID,
		PlayerID:     playerID,
		PlayerName:   playerName,
		PlayerAvatar: playerAvatar,
		LobbyName:    lobbyName,
		IsHost:       isHost,
		IsPrivate:    isPrivate,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
	}

	glm.mutex.Lock()
	defer glm.mutex.Unlock()
	glm.tokens[token] = joinToken
	return token, nil
}

// generateSessionTokenWithLobbyInfo_unsafe is a private helper that assumes the caller holds the lock.
func (glm *GameLifecycleManager) generateSessionTokenWithLobbyInfo_unsafe(lobbyID, playerID, playerName, playerAvatar, lobbyName string, isHost, isPrivate bool) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)

	joinToken := &lobby.JoinToken{
		Token:        token,
		LobbyID:      lobbyID,
		PlayerID:     playerID,
		PlayerName:   playerName,
		PlayerAvatar: playerAvatar,
		LobbyName:    lobbyName,
		IsHost:       isHost,
		IsPrivate:    isPrivate,
		ExpiresAt:    time.Now().Add(30 * time.Minute),
	}

	glm.tokens[token] = joinToken
	return token, nil
}

// GetLobbyByID returns lobby information (for HTTP API)
func (glm *GameLifecycleManager) GetLobbyByID(lobbyID string) (*lobby.Lobby, error) {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	lobby, exists := glm.lobbies[lobbyID]
	if !exists {
		return nil, fmt.Errorf("lobby not found")
	}

	return lobby, nil
}

// ValidateSessionToken validates and returns token information
func (glm *GameLifecycleManager) ValidateSessionToken(token string) (interface{}, error) {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	joinToken, exists := glm.tokens[token]
	if !exists {
		return nil, fmt.Errorf("invalid token")
	}

	if time.Now().After(joinToken.ExpiresAt) {
		return nil, fmt.Errorf("token expired")
	}

	return joinToken, nil
}

// cleanupExpiredTokens removes tokens that have expired
func (glm *GameLifecycleManager) cleanupExpiredTokens() {
	glm.mutex.Lock()
	defer glm.mutex.Unlock()

	now := time.Now()
	var expiredTokens []string
	for tokenStr, token := range glm.tokens {
		if now.After(token.ExpiresAt) {
			expiredTokens = append(expiredTokens, tokenStr)
		}
	}

	if len(expiredTokens) > 0 {
		for _, tokenStr := range expiredTokens {
			delete(glm.tokens, tokenStr)
		}
		log.Printf("GameLifecycleManager: Cleaned up %d expired tokens", len(expiredTokens))
	}
}

// Stop gracefully shuts down the manager
func (glm *GameLifecycleManager) Stop() {
	log.Println("GameLifecycleManager: Shutting down")
	close(glm.stopChannel)
}

// JoinLobby creates credentials for joining an existing lobby via HTTP
func (glm *GameLifecycleManager) JoinLobby(lobbyID, userID, playerName, playerAvatar string) (string, string, error) {
	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	glm.mutex.RUnlock() // Release manager lock after getting lobby reference

	if !exists {
		return "", "", fmt.Errorf("lobby not found")
	}

	// Now lock the specific lobby to read its state safely
	lobby.RLock() // Using read lock
	status := lobby.Status
	playerCount := len(lobby.Players)
	maxPlayers := lobby.MaxPlayers
	isPrivate := lobby.IsPrivate
	lobby.RUnlock()

	// Now do checks without holding a lock
	if status != "WAITING" && status != "WAITING_FOR_HOST" {
		return "", "", fmt.Errorf("lobby is not accepting new players")
	}

	if playerCount >= maxPlayers {
		return "", "", fmt.Errorf("lobby is full")
	}

	playerID := userID
	sessionToken, err := glm.generateSessionTokenWithLobbyInfo(lobbyID, playerID, playerName, playerAvatar, "", false, isPrivate)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	return playerID, sessionToken, nil
}

// JoinAsSpectator allows a user to join a running game as a spectator
func (glm *GameLifecycleManager) JoinAsSpectator(gameID, userID, spectatorName string) (string, string, error) {
	glm.mutex.RLock()
	_, exists := glm.gameActors[gameID]
	glm.mutex.RUnlock()

	if !exists {
		return "", "", fmt.Errorf("game not found")
	}

	// Check if the game is in progress
	// Note: We can't easily check the game state here without adding more complexity
	// For now, we'll let the GameActor handle the validation

	// Generate a unique spectator ID
	spectatorID := fmt.Sprintf("spectator-%s", userID)

	// Generate session token for spectator
	sessionToken, err := glm.generateSessionTokenWithLobbyInfo(gameID, spectatorID, spectatorName, "", "", false, false)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	return spectatorID, sessionToken, nil
}

// SendActionToGame forwards an action to the appropriate GameActor
func (glm *GameLifecycleManager) SendActionToGame(gameID string, action core.Action) (chan interfaces.ProcessActionResult, error) {
	glm.mutex.RLock()
	gameActor, exists := glm.gameActors[gameID]
	glm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("game not found: %s", gameID)
	}

	// Return the response channel directly to enable request-response pattern
	resultChan := gameActor.PostAction(action)
	return resultChan, nil
}

// BroadcastEventsToGame broadcasts events to all players in a game session
func (glm *GameLifecycleManager) BroadcastEventsToGame(gameID string, events []core.Event) error {
	glm.mutex.RLock()
	playerActors, sessionExists := glm.gameSessions[gameID]
	glm.mutex.RUnlock()

	if !sessionExists {
		return fmt.Errorf("could not find session for game %s", gameID)
	}

	// Broadcast events to players in the game session
	for _, event := range events {
		if event.PlayerID != "" { // Private event for a specific player
			if actor, ok := playerActors[event.PlayerID]; ok {
				actor.SendServerMessage(event)
			}
		} else { // Public event for all players in the game
			for _, actor := range playerActors {
				actor.SendServerMessage(event)
			}
		}
	}
	return nil
}

// ValidateSession implements the TokenValidator interface
func (glm *GameLifecycleManager) ValidateSession(gameID, playerID, sessionToken string) bool {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	token, exists := glm.tokens[sessionToken]
	if !exists {
		return false
	}

	// Check if token is expired - if so, trigger force logout for cleanup
	if !time.Now().Before(token.ExpiresAt) {
		// Schedule force logout outside of the read lock with panic protection
		helpers.GoSafe(glm.ctx, func(_ context.Context) {
			glm.ForceLogoutPlayer(playerID, "session_expired")
		})
		return false
	}

	// Check if token matches player and game
	return token.PlayerID == playerID &&
		   (token.LobbyID == gameID || gameID == "") // Allow empty gameID for lobby connections
}

// GetPlayerInfo implements the TokenValidator interface
func (glm *GameLifecycleManager) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	// Look up player info from the token
	for _, token := range glm.tokens {
		if token.PlayerID == playerID && (token.LobbyID == gameID || gameID == "") {
			// Return name and avatar from the token
			return token.PlayerName, token.PlayerAvatar, nil
		}
	}

	return "", "", fmt.Errorf("player info not found")
}

// GetInitialStateForPlayer returns the appropriate initial state snapshot for a connecting player
func (glm *GameLifecycleManager) GetInitialStateForPlayer(gameID, playerID string) (interface{}, error) {
	correlationID := uuid.New().String()[:8]
	log.Printf("[InitialState-%s] Getting initial state for player %s in game/lobby %s", correlationID, playerID, gameID)

	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	// Check if it's a lobby first
	if lobby, exists := glm.lobbies[gameID]; exists {
		log.Printf("[InitialState-%s] Found lobby %s, generating lobby state snapshot", correlationID, gameID)
		
		lobby.RLock()
		defer lobby.RUnlock()
		
		// Create lobby state snapshot
		var playerInfos []interface{}
		for _, playerInfo := range lobby.Players {
			playerInfos = append(playerInfos, map[string]interface{}{
				"id":               playerInfo.ID,
				"name":             playerInfo.Name,
				"avatar":           playerInfo.Avatar,
				"joinedAt":         playerInfo.JoinedAt,
				"connectionStatus": playerInfo.ConnectionStatus,
			})
		}

		snapshot := map[string]interface{}{
			"type": "LOBBY_STATE_UPDATE",
			"payload": map[string]interface{}{
				"lobby_id":      lobby.ID,
				"players":       playerInfos,
				"host_id":       lobby.HostPlayerID,
				"can_start":     len(lobby.Players) >= lobby.MinPlayers && (lobby.Status == "WAITING" || lobby.Status == "COUNTDOWN"),
				"name":          lobby.Name,
				"max_players":   lobby.MaxPlayers,
				"game_settings": lobby.GameSettings,
			},
		}

		log.Printf("[InitialState-%s] Generated lobby snapshot with %d players", correlationID, len(playerInfos))
		return snapshot, nil
	}

	// Check if it's an active game
	if gameActor, exists := glm.gameActors[gameID]; exists {
		log.Printf("[InitialState-%s] Found active game %s, generating game state snapshot", correlationID, gameID)
		
		// Get game state snapshot from the GameActor
		stateSnapshot := gameActor.CreatePlayerStateUpdateEvent(playerID)
		
		snapshot := map[string]interface{}{
			"type":    stateSnapshot.Type,
			"payload": stateSnapshot.Payload,
		}

		log.Printf("[InitialState-%s] Generated game state snapshot for player %s", correlationID, playerID)
		return snapshot, nil
	}

	log.Printf("[InitialState-%s] No lobby or game found for ID %s", correlationID, gameID)
	return nil, fmt.Errorf("no lobby or game found with ID: %s", gameID)
}

// GetLobbyList returns a list of active lobbies for the HTTP API
func (glm *GameLifecycleManager) GetLobbyList() []interface{} {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	lobbies := make([]interface{}, 0, len(glm.lobbies))
	for _, lobby := range glm.lobbies {
		// Use fine-grained locking to read lobby state safely
		lobby.RLock()
		if lobby.Status == "WAITING" && !lobby.IsPrivate {
			playerActors := len(lobby.Players) // Read directly to avoid extra lock
			lobbies = append(lobbies, map[string]interface{}{
				"id":            lobby.ID,
				"name":          lobby.Name,
				"player_count":  playerActors,
				"max_players":   lobby.MaxPlayers,
				"min_players":   lobby.MinPlayers,
				"can_join":      (lobby.Status == "WAITING" || lobby.Status == "WAITING_FOR_HOST") && playerActors < lobby.MaxPlayers,
				"status":        lobby.Status,
				"game_settings": lobby.GameSettings,
			})
		}
		lobby.RUnlock()
	}

	return lobbies
}

// GetGameActor returns the game actor for a given game ID
func (glm *GameLifecycleManager) GetGameActor(gameID string) (interfaces.GameActorInterface, bool) {
	return glm.supervisor.GetActor(gameID)
}

// ReconnectPlayerToGame re-adds a reconnecting player to an active game session
func (glm *GameLifecycleManager) ReconnectPlayerToGame(gameID string, playerActor interfaces.PlayerActorInterface) error {
	// Generate correlation ID for tracking this reconnection flow
	correlationID := uuid.New().String()[:8]
	playerID := playerActor.GetPlayerID()
	
	log.Printf("[Connection-%s] ReconnectPlayerToGame: Starting reconnection for player %s to game %s", correlationID, playerID, gameID)
	
	glm.mutex.Lock()
	defer glm.mutex.Unlock()

	// Check if the game session exists
	session, exists := glm.gameSessions[gameID]
	if !exists {
		log.Printf("[Connection-%s] ReconnectPlayerToGame: FAILED - game session not found: %s", correlationID, gameID)
		return fmt.Errorf("game session not found: %s", gameID)
	}
	
	log.Printf("[Connection-%s] ReconnectPlayerToGame: Found existing game session %s", correlationID, gameID)

	// Cancel the grace period timer if it exists
	glm.cancelGracePeriod(gameID, playerID)
	log.Printf("[Connection-%s] ReconnectPlayerToGame: Cancelled grace period for player %s", correlationID, playerID)

	// Add the player back to the game session
	session[playerID] = playerActor
	log.Printf("[Connection-%s] ReconnectPlayerToGame: Player %s added back to game session", correlationID, playerID)

	// Notify the GameActor about the reconnection (set connection status to CONNECTED)
	glm.mutex.RLock()
	gameActor, exists := glm.gameActors[gameID]
	glm.mutex.RUnlock()

	if exists {
		log.Printf("[Connection-%s] ReconnectPlayerToGame: Found GameActor, sending connection status update and state snapshot", correlationID)
		
		// Send SET_PLAYER_CONNECTION_STATUS action to GameActor
		reconnectAction := core.Action{
			Type:     core.ActionSetPlayerConnectionStatus,
			PlayerID: playerID,
			GameID:   gameID,
			Payload: map[string]interface{}{
				"connection_status": "CONNECTED",
			},
		}

		// Post action synchronously to ensure connection status is updated first
		log.Printf("[Connection-%s] ReconnectPlayerToGame: Posting connection status action to GameActor", correlationID)
		resultChan := gameActor.PostAction(reconnectAction)
		result := <-resultChan
		if result.Error != nil {
			log.Printf("[Connection-%s] ReconnectPlayerToGame: ERROR - Failed to set player connection status: %v", correlationID, result.Error)
		} else {
			log.Printf("[Connection-%s] ReconnectPlayerToGame: Successfully updated connection status in GameActor", correlationID)
		}

		// Now send the player a complete state snapshot to ensure they have current game state
		log.Printf("[Connection-%s] ReconnectPlayerToGame: Sending game state snapshot to reconnected player", correlationID)
		stateSnapshot := gameActor.CreatePlayerStateUpdateEvent(playerID)
		playerActor.SendServerMessage(stateSnapshot)
		log.Printf("[Connection-%s] ReconnectPlayerToGame: Game state snapshot sent to player %s", correlationID, playerID)
	} else {
		log.Printf("[Connection-%s] ReconnectPlayerToGame: WARNING - GameActor not found for game %s", correlationID, gameID)
	}

	log.Printf("[Connection-%s] ReconnectPlayerToGame: SUCCESS - Player %s reconnected to game session %s", correlationID, playerID, gameID)
	return nil
}

// RehydrateGamesFromStore scans Redis for active games and restarts their actors
func (glm *GameLifecycleManager) RehydrateGamesFromStore() error {
	log.Println("GameLifecycleManager: Starting game rehydration from persistent store")
	
	activeGameIDs, err := glm.datastore.ListActiveGames(glm.ctx)
	if err != nil {
		return fmt.Errorf("could not list active games from datastore: %w", err)
	}

	log.Printf("GameLifecycleManager: Found %d active games to rehydrate", len(activeGameIDs))
	
	successCount := 0
	errorCount := 0
	
	for _, gameID := range activeGameIDs {
		log.Printf("GameLifecycleManager: Rehydrating game: %s", gameID)
		
		// Load the game state snapshot to understand the current state
		gameState, err := glm.datastore.GetLatestSnapshot(glm.ctx, gameID)
		if err != nil {
			log.Printf("GameLifecycleManager: Error loading snapshot for game %s: %v", gameID, err)
			errorCount++
			continue
		}
		
		// Skip games that are already completed
		if gameState.Phase.Type == core.PhaseGameOver {
			log.Printf("GameLifecycleManager: Skipping completed game %s", gameID)
			continue
		}
		
		// Extract players from the game state to reconstruct the player map
		players := make(map[string]*core.Player)
		for playerID, player := range gameState.Players {
			players[playerID] = player
		}
		
		// Create the GameActor via Supervisor
		// The supervisor will spawn the actor, which will then trigger its own recovery logic
		gameActor, err := glm.supervisor.CreateGameWithPlayers(gameID, players)
		if err != nil {
			log.Printf("GameLifecycleManager: Error rehydrating game %s: %v", gameID, err)
			errorCount++
			continue
		}
		
		// Track the rehydrated game in our internal state
		glm.mutex.Lock()
		glm.gameActors[gameID] = gameActor
		// Note: We don't populate gameSessions since the actual player WebSocket connections
		// will be established when players reconnect. The GameActor itself maintains
		// the authoritative game state.
		glm.mutex.Unlock()
		
		log.Printf("GameLifecycleManager: Successfully rehydrated game %s", gameID)
		successCount++
	}
	
	log.Printf("GameLifecycleManager: Game rehydration completed - Success: %d, Errors: %d", 
		successCount, errorCount)
	
	if errorCount > 0 {
		log.Printf("GameLifecycleManager: Warning - %d games failed to rehydrate", errorCount)
	}
	
	return nil
}

// checkForBlockedPlayers checks if the joining player or existing players have blocked each other
func (glm *GameLifecycleManager) checkForBlockedPlayers(joiningPlayerID string, lobby *lobby.Lobby) error {
	if glm.postgresStore == nil {
		// If PostgreSQL is not available, allow all joins (graceful degradation)
		return nil
	}

	// Get the list of blocked players for the joining player
	joiningPlayerBlocked, err := glm.postgresStore.GetBlockedPlayers(joiningPlayerID)
	if err != nil {
		log.Printf("Warning: Failed to get blocked players for %s: %v", joiningPlayerID, err)
		// Don't block the join if we can't check - graceful degradation
		return nil
	}

	// Check each existing player in the lobby
	lobby.RLock()
	defer lobby.RUnlock()

	for existingPlayerID := range lobby.Players {
		// Check if the joining player has blocked this existing player
		for _, blockedID := range joiningPlayerBlocked {
			if blockedID == existingPlayerID {
				return fmt.Errorf("cannot join lobby: you have blocked a player in this game")
			}
		}

		// Check if the existing player has blocked the joining player
		existingPlayerBlocked, err := glm.postgresStore.GetBlockedPlayers(existingPlayerID)
		if err != nil {
			log.Printf("Warning: Failed to get blocked players for %s: %v", existingPlayerID, err)
			continue // Skip this check if we can't verify
		}

		for _, blockedID := range existingPlayerBlocked {
			if blockedID == joiningPlayerID {
				return fmt.Errorf("cannot join lobby: another player has blocked you")
			}
		}
	}

	return nil
}