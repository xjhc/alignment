package comms

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/semaphore"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/actors"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/party"
	"github.com/xjhc/alignment/server/internal/store"
)

// WebSocketManager handles WebSocket connections via PlayerActors
type WebSocketManager struct {
	playerActors   map[string]*actors.PlayerActor
	actorsMutex    sync.RWMutex
	ctx            context.Context
	tokenValidator TokenValidator

	// Dependencies for PlayerActors
	lifecycleManager interfaces.GameLifecycleManagerInterface
	eventBus         *events.EventBus
	postgresStore    *store.PostgresStore
	partyManager     *party.PartyManager
	actionSemaphore  *semaphore.Weighted
}

// TokenValidator validates sessions and provides player information
type TokenValidator interface {
	ValidateSession(gameId, playerId, sessionToken string) bool
	GetPlayerInfo(gameId, playerId string) (string, string, error) // Returns playerName, avatar, error
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking
		return true
	},
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(ctx context.Context, tokenValidator TokenValidator) *WebSocketManager {
	return &WebSocketManager{
		playerActors:   make(map[string]*actors.PlayerActor),
		ctx:            ctx,
		tokenValidator: tokenValidator,
		partyManager:   party.NewPartyManager(),
	}
}

// SetDependencies injects the required managers
func (wsm *WebSocketManager) SetDependencies(lifecycleManager interfaces.GameLifecycleManagerInterface, eventBus *events.EventBus) {
	wsm.lifecycleManager = lifecycleManager
	wsm.eventBus = eventBus
}

// SetPostgresStore sets the PostgreSQL store for presence tracking
func (wsm *WebSocketManager) SetPostgresStore(postgresStore *store.PostgresStore) {
	wsm.postgresStore = postgresStore
	
	// Subscribe to events for presence tracking
	if wsm.eventBus != nil && postgresStore != nil {
		wsm.startPresenceEventListeners()
	}
}

// SetActionSemaphore sets the global action semaphore
func (wsm *WebSocketManager) SetActionSemaphore(sem *semaphore.Weighted) {
	wsm.actionSemaphore = sem
}

// startPresenceEventListeners sets up event listeners for presence tracking and player management
func (wsm *WebSocketManager) startPresenceEventListeners() {
	// Create channels for different event types
	playerJoinedCh := make(chan events.Event, 10)
	playerLeftCh := make(chan events.Event, 10)
	gameStartedCh := make(chan events.Event, 10)
	playerDisconnectedCh := make(chan events.Event, 10)
	forceLogoutCh := make(chan events.Event, 10)
	
	// Subscribe to events
	wsm.eventBus.Subscribe("player_joined_lobby", playerJoinedCh)
	wsm.eventBus.Subscribe("player_left_lobby", playerLeftCh)
	wsm.eventBus.Subscribe("game_started", gameStartedCh)
	wsm.eventBus.Subscribe("player_disconnected", playerDisconnectedCh)
	wsm.eventBus.Subscribe("force_logout", forceLogoutCh)
	
	// Start goroutines to handle events
	go wsm.listenForPlayerJoinedEvents(playerJoinedCh)
	go wsm.listenForPlayerLeftEvents(playerLeftCh)
	go wsm.listenForGameStartedEvents(gameStartedCh)
	go wsm.listenForPlayerDisconnectedEvents(playerDisconnectedCh)
	go wsm.listenForForceLogoutEvents(forceLogoutCh)
}

// updatePlayerPresence updates a player's presence status in the database
func (wsm *WebSocketManager) updatePlayerPresence(playerID, status string, lobbyID, gameID *string) {
	if wsm.postgresStore == nil {
		return // Presence tracking not available
	}

	if err := wsm.postgresStore.UpdatePlayerPresence(playerID, status, lobbyID, gameID); err != nil {
		log.Printf("WebSocketManager: Failed to update presence for player %s: %v", playerID, err)
	}
}

// UpdatePlayerState updates a player's presence based on their actor state
func (wsm *WebSocketManager) UpdatePlayerState(playerID string) {
	wsm.actorsMutex.RLock()
	actor, exists := wsm.playerActors[playerID]
	wsm.actorsMutex.RUnlock()
	
	if !exists {
		return
	}
	
	state := actor.GetState()
	lobbyID := actor.GetLobbyID()
	gameID := actor.GetGameID()
	
	var status string
	var lobbyPtr, gamePtr *string
	
	switch state {
	case interfaces.StateIdle:
		status = "online"
		lobbyPtr = nil
		gamePtr = nil
	case interfaces.StateInLobby:
		status = "in_lobby"
		if lobbyID != "" {
			lobbyPtr = &lobbyID
		}
		gamePtr = nil
	case interfaces.StateInGame:
		status = "in_game"
		lobbyPtr = nil
		if gameID != "" {
			gamePtr = &gameID
		}
	default:
		status = "online"
		lobbyPtr = nil
		gamePtr = nil
	}
	
	wsm.updatePlayerPresence(playerID, status, lobbyPtr, gamePtr)
}

// Event listeners for presence tracking

func (wsm *WebSocketManager) listenForPlayerJoinedEvents(ch chan events.Event) {
	for event := range ch {
		if e, ok := event.(events.PlayerJoinedLobbyEvent); ok {
			wsm.updatePlayerPresence(e.PlayerID, "in_lobby", &e.LobbyID, nil)
		}
	}
}

func (wsm *WebSocketManager) listenForPlayerLeftEvents(ch chan events.Event) {
	for event := range ch {
		if e, ok := event.(events.PlayerLeftLobbyEvent); ok {
			wsm.updatePlayerPresence(e.PlayerID, "online", nil, nil)
		}
	}
}

func (wsm *WebSocketManager) listenForGameStartedEvents(ch chan events.Event) {
	for event := range ch {
		if e, ok := event.(events.GameStartedEvent); ok {
			// Update all players in the game to "in_game" status
			for _, playerID := range e.PlayerIDs {
				wsm.updatePlayerPresence(playerID, "in_game", nil, &e.GameID)
			}
		}
	}
}

func (wsm *WebSocketManager) listenForPlayerDisconnectedEvents(ch chan events.Event) {
	for event := range ch {
		if e, ok := event.(events.PlayerDisconnectedEvent); ok {
			wsm.updatePlayerPresence(e.PlayerID, "offline", nil, nil)
		}
	}
}

func (wsm *WebSocketManager) listenForForceLogoutEvents(ch chan events.Event) {
	for event := range ch {
		if e, ok := event.(events.ForceLogoutEvent); ok {
			log.Printf("WebSocketManager: Processing force logout event for player %s: %s", e.PlayerID, e.Reason)
			wsm.ForceLogoutPlayer(e.PlayerID, e.Reason)
		}
	}
}

// Start is now a no-op since PlayerActors manage themselves
func (wsm *WebSocketManager) Start() {
	log.Println("WebSocketManager: Ready to handle connections")
}

// HandleWebSocket handles WebSocket connection upgrades and creates PlayerActors
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("gameId")
	playerID := r.URL.Query().Get("playerId")
	sessionToken := r.URL.Query().Get("sessionToken")

	if gameID == "" || playerID == "" || sessionToken == "" {
		http.Error(w, "Missing required parameters: gameId, playerId, sessionToken", http.StatusBadRequest)
		return
	}

	// Check session validity
	sessionValid := wsm.tokenValidator.ValidateSession(gameID, playerID, sessionToken)
	
	// Get player information regardless of session validity (for SESSION_EXPIRED handling)
	playerName, playerAvatar, err := wsm.tokenValidator.GetPlayerInfo(gameID, playerID)
	if err != nil && sessionValid {
		// Only error out if session is valid but we can't get player info
		http.Error(w, fmt.Sprintf("Failed to get player info: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Always upgrade to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	
	// If session is invalid, send SESSION_EXPIRED and close
	if !sessionValid {
		// Create a temporary PlayerActor to send the session expired event
		tempActor := actors.NewPlayerActor(wsm.ctx, playerID, playerName, playerAvatar, sessionToken, conn)
		tempActor.Start()
		wsm.sendSessionExpiredAndClose(tempActor, gameID, "session_invalid", "Your session has expired. Please log in again.")
		return
	}

	// Check if PlayerActor already exists (reconnection)
	wsm.actorsMutex.Lock()
	existingActor, exists := wsm.playerActors[playerID]
	if exists {
		// Stop the existing actor first
		existingActor.Stop()
		delete(wsm.playerActors, playerID)
	}

	// Create new PlayerActor
	playerActor := actors.NewPlayerActor(wsm.ctx, playerID, playerName, playerAvatar, sessionToken, conn)
	playerActor.SetDependencies(wsm.lifecycleManager, wsm.eventBus, wsm.partyManager)
	
	// Set the global action semaphore for admission control
	if wsm.actionSemaphore != nil {
		playerActor.SetActionSemaphore(wsm.actionSemaphore)
	}

	wsm.playerActors[playerID] = playerActor
	wsm.actorsMutex.Unlock()

	// Update player presence to "online"
	wsm.updatePlayerPresence(playerID, "online", nil, nil)

	// Start the PlayerActor
	playerActor.Start()

	// Determine if this is a reconnection to an active game or joining a lobby
	// The initial state snapshot is now sent atomically by JoinLobbyWithActor or ReconnectPlayerToGame
	wsm.handleConnectionRoutingLogic(gameID, playerActor)

	log.Printf("WebSocketManager: Created PlayerActor for %s (%s) and joined lobby %s", playerID, playerName, gameID)
}

// handleConnectionRoutingLogic determines if this is a reconnection to a game or joining a lobby
func (wsm *WebSocketManager) handleConnectionRoutingLogic(gameID string, playerActor *actors.PlayerActor) {
	if wsm.lifecycleManager == nil {
		log.Printf("WebSocketManager: Lifecycle manager not initialized")
		wsm.sendSessionExpiredAndClose(playerActor, gameID, "server_error", "Server configuration error")
		return
	}

	// First, check if this is an active game (for reconnection)
	gameActor, gameExists := wsm.lifecycleManager.GetGameActor(gameID)
	if gameExists {
		// Check if this is a spectator (player ID starts with "spectator-")
		if isSpectator := len(playerActor.GetPlayerID()) > 10 && playerActor.GetPlayerID()[:10] == "spectator-"; isSpectator {
			log.Printf("WebSocketManager: Spectator %s connecting to active game %s", playerActor.GetPlayerID(), gameID)
			
			// Transition the player actor to the spectating state
			err := playerActor.TransitionToSpectating(gameID)
			if err != nil {
				log.Printf("WebSocketManager: Failed to transition spectator %s to spectating state: %v", playerActor.GetPlayerID(), err)
				wsm.sendSessionExpiredAndClose(playerActor, gameID, "transition_failed", "Failed to transition to spectating")
				return
			}
			
			// Add the spectator to the game actor
			gameActor.AddSpectator(playerActor)
			
			log.Printf("WebSocketManager: Successfully connected spectator %s to game %s", playerActor.GetPlayerID(), gameID)
			return
		}
		
		// This is a regular player reconnection to an active game
		log.Printf("WebSocketManager: Player %s reconnecting to active game %s", playerActor.GetPlayerID(), gameID)
		
		err := wsm.lifecycleManager.ReconnectPlayerToGame(gameID, playerActor)
		if err != nil {
			log.Printf("WebSocketManager: Failed to reconnect player %s to game %s: %v", playerActor.GetPlayerID(), gameID, err)
			wsm.sendSessionExpiredAndClose(playerActor, gameID, "reconnection_failed", "Failed to reconnect to game")
			return
		}
		
		// Transition the player actor to the game state
		err = playerActor.TransitionToGame(gameID)
		if err != nil {
			log.Printf("WebSocketManager: Failed to transition player %s to game state: %v", playerActor.GetPlayerID(), err)
			wsm.sendSessionExpiredAndClose(playerActor, gameID, "transition_failed", "Failed to transition to game")
			return
		}
		
		// Send the current game state to the reconnecting player
		gameStateEvent := gameActor.CreatePlayerStateUpdateEvent(playerActor.GetPlayerID())
		playerActor.SendServerMessage(gameStateEvent)
		
		log.Printf("WebSocketManager: Successfully reconnected player %s to game %s", playerActor.GetPlayerID(), gameID)
		return
	}

	// Not an active game, try to join as a lobby
	err := wsm.joinLobbyAutomatically(gameID, playerActor)
	if err != nil {
		log.Printf("WebSocketManager: Failed to auto-join lobby %s for player %s: %v", gameID, playerActor.GetPlayerID(), err)
		
		// Check if this is a "lobby not found" error vs other errors
		if err.Error() == "lobby not found: "+gameID {
			wsm.sendSessionExpiredAndClose(playerActor, gameID, "lobby_not_found", "The lobby you were trying to join no longer exists. Please join a new game.")
		} else {
			wsm.sendSessionExpiredAndClose(playerActor, gameID, "join_failed", "Failed to join lobby")
		}
		return
	}
}

// sendSessionExpiredAndClose sends a session expired event and closes the connection
func (wsm *WebSocketManager) sendSessionExpiredAndClose(playerActor *actors.PlayerActor, gameID, reason, message string) {
	sessionExpiredEvent := core.Event{
		ID:        "session_expired_" + playerActor.GetPlayerID(),
		Type:      "SESSION_EXPIRED",
		GameID:    gameID,
		PlayerID:  playerActor.GetPlayerID(),
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"reason":  reason,
			"message": message,
		},
	}
	
	// Send the event before closing
	playerActor.SendServerMessage(sessionExpiredEvent)
	
	// Give the message time to be sent
	time.Sleep(100 * time.Millisecond)
	
	// Stop the player actor and clean up
	playerActor.Stop()
	wsm.actorsMutex.Lock()
	delete(wsm.playerActors, playerActor.GetPlayerID())
	wsm.actorsMutex.Unlock()
}

// joinLobbyAutomatically handles the automatic lobby joining in REST-then-WebSocket flow
func (wsm *WebSocketManager) joinLobbyAutomatically(gameIDOrLobbyID string, playerActor *actors.PlayerActor) error {
	if wsm.lifecycleManager == nil {
		return fmt.Errorf("lifecycle manager not initialized")
	}

	// First, check if this is an active game (for reconnection)
	gameActor, gameExists := wsm.lifecycleManager.GetGameActor(gameIDOrLobbyID)
	if gameExists {
		// This is a reconnection to an active game
		log.Printf("WebSocketManager: Player %s reconnecting to active game %s", playerActor.GetPlayerID(), gameIDOrLobbyID)
		
		// Transition the player actor to the game
		err := playerActor.TransitionToGame(gameIDOrLobbyID)
		if err != nil {
			return fmt.Errorf("failed to transition player to game %s: %w", gameIDOrLobbyID, err)
		}
		
		// Re-add the player to the game session to fix state corruption
		err = wsm.lifecycleManager.ReconnectPlayerToGame(gameIDOrLobbyID, playerActor)
		if err != nil {
			log.Printf("WebSocketManager: Warning - failed to reconnect player to game session: %v", err)
			// Don't fail the reconnection if session tracking fails, but log it
		}
		
		// Send a PLAYER_RECONNECTED event to notify other players
		reconnectedEvent := core.Event{
			ID:        "reconnect_" + playerActor.GetPlayerID() + "_" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Type:      core.EventPlayerReconnected,
			GameID:    gameIDOrLobbyID,
			PlayerID:  "", // Public event for all players
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"player_id":   playerActor.GetPlayerID(),
				"player_name": playerActor.GetPlayerName(),
			},
		}
		
		// Broadcast to all players in the game
		err = wsm.BroadcastToGame(gameIDOrLobbyID, reconnectedEvent)
		if err != nil {
			log.Printf("WebSocketManager: Warning - failed to broadcast reconnection event: %v", err)
		}
		
		// Send a game state snapshot to the reconnecting player
		snapshotEvent := gameActor.CreatePlayerStateUpdateEvent(playerActor.GetPlayerID())
		playerActor.SendServerMessage(snapshotEvent)
		
		log.Printf("WebSocketManager: Player %s successfully reconnected to game %s", playerActor.GetPlayerID(), gameIDOrLobbyID)
		return nil
	}

	// If not a game, try to join as a lobby
	err := wsm.lifecycleManager.JoinLobbyWithActor(gameIDOrLobbyID, playerActor)
	if err != nil {
		return fmt.Errorf("lobby not found: %s", gameIDOrLobbyID)
	}

	log.Printf("WebSocketManager: Player %s automatically joined lobby %s", playerActor.GetPlayerID(), gameIDOrLobbyID)
	
	// For lobby joins, the AddPlayer function in the lobby will automatically send
	// a LOBBY_STATE_UPDATE event to all players in the lobby, including the newly joined player.
	// This ensures that the reconnecting player receives the complete current lobby state.
	
	return nil
}

// GetPlayerActor returns a PlayerActor by ID
func (wsm *WebSocketManager) GetPlayerActor(playerID string) (*actors.PlayerActor, bool) {
	wsm.actorsMutex.RLock()
	defer wsm.actorsMutex.RUnlock()
	actor, exists := wsm.playerActors[playerID]
	return actor, exists
}

// RemovePlayerActor removes a PlayerActor (called when they disconnect)
func (wsm *WebSocketManager) RemovePlayerActor(playerID string) {
	wsm.actorsMutex.Lock()
	if actor, exists := wsm.playerActors[playerID]; exists {
		actor.Stop()
		delete(wsm.playerActors, playerID)
		wsm.actorsMutex.Unlock()
		
		// Update player presence to "offline"
		wsm.updatePlayerPresence(playerID, "offline", nil, nil)
		
		log.Printf("WebSocketManager: Removed PlayerActor for %s", playerID)
	} else {
		wsm.actorsMutex.Unlock()
	}
}

// BroadcastToGame sends an event to all PlayerActors in a specific game/lobby
func (wsm *WebSocketManager) BroadcastToGame(gameID string, event core.Event) error {
	wsm.actorsMutex.RLock()
	var actorsToNotify []*actors.PlayerActor
	for _, actor := range wsm.playerActors {
		// Check if the player is in the target game or lobby
		actorGameID := actor.GetGameID()
		actorLobbyID := actor.GetLobbyID()

		// Include players who are either in the game or in a lobby that matches the gameID
		if actorGameID == gameID || actorLobbyID == gameID {
			actorsToNotify = append(actorsToNotify, actor)
		}
	}
	wsm.actorsMutex.RUnlock()

	// Send only to relevant actors
	for _, actor := range actorsToNotify {
		actor.SendServerMessage(event)
	}

	log.Printf("WebSocketManager: Broadcasted %s to %d players for game %s", event.Type, len(actorsToNotify), gameID)
	return nil
}

// SendToPlayer sends an event to a specific PlayerActor
func (wsm *WebSocketManager) SendToPlayer(gameID, playerID string, event core.Event) error {
	wsm.actorsMutex.RLock()
	actor, exists := wsm.playerActors[playerID]
	wsm.actorsMutex.RUnlock()

	if !exists {
		return ErrPlayerNotFound
	}

	// Validate that the player is actually in the specified game/lobby
	actorGameID := actor.GetGameID()
	actorLobbyID := actor.GetLobbyID()
	if actorGameID != gameID && actorLobbyID != gameID {
		log.Printf("WebSocketManager: Player %s not in game %s (playerGame=%s, playerLobby=%s)",
			playerID, gameID, actorGameID, actorLobbyID)
		return fmt.Errorf("player not in specified game")
	}

	actor.SendServerMessage(event)
	return nil
}

// GetStats returns statistics about connected players
func (wsm *WebSocketManager) GetStats() map[string]interface{} {
	wsm.actorsMutex.RLock()
	defer wsm.actorsMutex.RUnlock()

	stats := make(map[string]interface{})
	stats["connected_players"] = len(wsm.playerActors)

	stateCounts := make(map[string]int)
	for _, actor := range wsm.playerActors {
		state := actor.GetState().String()
		stateCounts[state]++
	}
	stats["player_states"] = stateCounts

	return stats
}

// ForceLogoutPlayer sends a FORCE_LOGOUT event to a specific player and disconnects them
func (wsm *WebSocketManager) ForceLogoutPlayer(playerID string, reason string) {
	wsm.actorsMutex.RLock()
	actor, exists := wsm.playerActors[playerID]
	wsm.actorsMutex.RUnlock()

	if exists {
		// Send FORCE_LOGOUT event before disconnecting
		forceLogoutEvent := core.Event{
			ID:        "force_logout_" + playerID + "_" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Type:      core.EventForceLogout,
			GameID:    "",
			PlayerID:  playerID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"reason":  reason,
				"message": "Server has detected an unrecoverable state and is forcing logout: " + reason,
			},
		}

		actor.SendServerMessage(forceLogoutEvent)

		// Give the message time to be sent
		time.Sleep(100 * time.Millisecond)

		log.Printf("WebSocketManager: Forced logout for player %s: %s", playerID, reason)
	}

	// Remove the player actor
	wsm.RemovePlayerActor(playerID)
}

// Custom errors
var (
	ErrClientDisconnected = fmt.Errorf("client disconnected")
	ErrPlayerNotFound     = fmt.Errorf("player not found")
)
