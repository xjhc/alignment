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
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/lobby"
)


// GameLifecycleManager unifies lobby and session management with event-driven architecture
type GameLifecycleManager struct {
	// Lobby management
	lobbies map[string]*lobby.Lobby
	tokens  map[string]*lobby.JoinToken

	// Game session management
	gameSessions map[string]map[string]interfaces.PlayerActorInterface
	gameActors   map[string]interfaces.GameActorInterface

	// Synchronization
	mutex sync.RWMutex
	ctx   context.Context

	// Dependencies
	datastore   interfaces.DataStore
	broadcaster interfaces.Broadcaster
	supervisor  interfaces.SupervisorInterface
	eventBus    *events.EventBus

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
) *GameLifecycleManager {
	glm := &GameLifecycleManager{
		lobbies:      make(map[string]*lobby.Lobby),
		tokens:       make(map[string]*lobby.JoinToken),
		gameSessions: make(map[string]map[string]interfaces.PlayerActorInterface),
		gameActors:   make(map[string]interfaces.GameActorInterface),
		ctx:          ctx,
		datastore:    datastore,
		broadcaster:  broadcaster,
		supervisor:   supervisor,
		eventBus:     eventBus,
		eventChannel: make(chan events.Event, 100), // Buffered to prevent blocking
		stopChannel:  make(chan struct{}),
	}

	// Subscribe to relevant events
	eventBus.Subscribe("player_disconnected", glm.eventChannel)
	eventBus.Subscribe("game_ended", glm.eventChannel)

	// Start event processing goroutine
	go glm.processEvents()

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
	case events.GameEndedEvent:
		glm.handleGameEnded(e)
	default:
		log.Printf("GameLifecycleManager: Unknown event type: %T", event)
	}
}

// handlePlayerDisconnected removes disconnected players from lobbies/games
func (glm *GameLifecycleManager) handlePlayerDisconnected(event events.PlayerDisconnectedEvent) {
	log.Printf("GameLifecycleManager: Handling player disconnection: %s", event.PlayerID)

	// Handle lobby disconnection
	if event.LobbyID != "" {
		glm.mutex.RLock()
		lobby, exists := glm.lobbies[event.LobbyID]
		glm.mutex.RUnlock()

		if exists {
			lobby.RemovePlayer(event.PlayerID)

			// Publish player left event
			glm.eventBus.Publish(events.PlayerLeftLobbyEvent{
				PlayerID: event.PlayerID,
				LobbyID:  event.LobbyID,
			})

			// Check if lobby is now empty and should be cleaned up
			glm.mutex.Lock()
			if len(lobby.GetPlayerActors()) == 0 {
				delete(glm.lobbies, event.LobbyID)
				log.Printf("GameLifecycleManager: Cleaned up empty lobby %s", event.LobbyID)
			}
			glm.mutex.Unlock()
		}
	}

	// Handle game disconnection
	if event.GameID != "" {
		glm.mutex.Lock()
		if session, exists := glm.gameSessions[event.GameID]; exists {
			delete(session, event.PlayerID)
			log.Printf("GameLifecycleManager: Removed player %s from game session %s", event.PlayerID, event.GameID)
		}
		glm.mutex.Unlock()

		// Notify the GameActor about the disconnection
		glm.mutex.RLock()
		gameActor, exists := glm.gameActors[event.GameID]
		glm.mutex.RUnlock()

		if exists {
			// Send player disconnection to GameActor
			disconnectAction := core.Action{
				Type:     "PLAYER_DISCONNECTED",
				PlayerID: event.PlayerID,
			}
			// Post action asynchronously
			go func() {
				resultChan := gameActor.PostAction(disconnectAction)
				result := <-resultChan
				if result.Error != nil {
					log.Printf("GameLifecycleManager: Error handling player disconnection: %v", result.Error)
				}
			}()
		}
	}
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

// CreateLobbyViaHTTP creates a lobby via HTTP and returns join credentials
func (glm *GameLifecycleManager) CreateLobbyViaHTTP(hostPlayerName, lobbyName, playerAvatar string) (string, string, string, error) {
	lobbyID := uuid.New().String()
	hostPlayerID := fmt.Sprintf("player_%s_%d", hostPlayerName, time.Now().UnixNano())

	// Generate session token
	sessionToken, err := glm.generateSessionTokenWithLobbyInfo(lobbyID, hostPlayerID, hostPlayerName, playerAvatar, lobbyName, true)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	// Create the lobby immediately (not waiting for WebSocket connection)
	// This ensures it appears in the lobby list right away
	glm.mutex.Lock()
	if lobbyName == "" {
		lobbyName = hostPlayerName + "'s Game"
	}

	// Create a placeholder lobby without a PlayerActor (will be added when host connects)
	newLobby := &lobby.Lobby{
		ID:           lobbyID,
		Name:         lobbyName,
		HostPlayerID: hostPlayerID,
		Players:      make(map[string]interfaces.PlayerActorInterface),
		MaxPlayers:   8,
		MinPlayers:   2,
		CreatedAt:    time.Now(),
		Status:       "WAITING_FOR_HOST",
	}

	glm.lobbies[lobbyID] = newLobby
	glm.mutex.Unlock()

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
	glm.mutex.Lock()
	targetLobby, exists := glm.lobbies[lobbyID]

	if !exists {
		// This should not happen in the regular flow anymore, as the lobby
		// is created via HTTP first. But as a safeguard:
		glm.mutex.Unlock()
		return fmt.Errorf("lobby not found: %s", lobbyID)
	}

	// Lock the specific lobby for state changes
	targetLobby.Lock()

	playerID := playerActor.GetPlayerID()
	// Check if this is the host connecting for the first time
	if targetLobby.Status == "WAITING_FOR_HOST" && targetLobby.HostPlayerID == playerID {
		// Host is connecting - transition lobby from placeholder to active
		targetLobby.Status = "WAITING"
		log.Printf("GameLifecycleManager: Host %s connected, lobby %s is now active", playerID, lobbyID)
	} else if targetLobby.Status == "WAITING_FOR_HOST" {
		// If another player tries to join before the host, reject them.
		targetLobby.Unlock()
		glm.mutex.Unlock()
		return fmt.Errorf("lobby is not accepting new players yet")
	}
	targetLobby.Unlock() // Unlock the lobby after status check/update
	glm.mutex.Unlock() // Unlock the manager after getting the lobby ref

	// Add player to lobby (this will handle its own locking and broadcasting)
	err := targetLobby.AddPlayer(playerActor)
	if err != nil {
		return err
	}

	// Transition player actor to lobby state
	return playerActor.TransitionToLobby(lobbyID)
}

// StartGame atomically transitions a lobby to an active game
func (glm *GameLifecycleManager) StartGame(lobbyID string, hostPlayerID string) error {
	log.Printf("GameLifecycleManager: Starting game for lobby %s from host %s", lobbyID, hostPlayerID)

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

	// Mark as starting
	lobby.SetStatus("STARTING")

	// Copy players for game creation
	playerActors := lobby.GetPlayerActors()

	// Create the game
	err := glm.createGameFromLobby(lobbyID, playerActors)
	if err != nil {
		// Revert lobby state on failure
		lobby.SetStatus("WAITING")
		return fmt.Errorf("failed to create game: %w", err)
	}

	// Remove lobby from manager
	glm.mutex.Lock()
	delete(glm.lobbies, lobbyID)
	glm.mutex.Unlock()

	return nil
}

// createGameFromLobby handles the atomic transition from lobby to game
func (glm *GameLifecycleManager) createGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	log.Printf("GameLifecycleManager: Creating game from lobby %s with %d players", lobbyID, len(playerActors))

	// Convert PlayerActors to core.Players map
	players := make(map[string]*core.Player)
	for playerID, actor := range playerActors {
		players[playerID] = &core.Player{
			ID:          playerID,
			Name:        actor.GetPlayerName(),
			ControlType: "HUMAN",
			IsAlive:     true,
		}
	}

	// Add the AI player
	aiPlayerID := "ai-nexus-" + uuid.New().String()[:8]
	players[aiPlayerID] = &core.Player{
		ID:          aiPlayerID,
		Name:        "NEXUS",
		JobTitle:    "AI Assistant",
		ControlType: "AI",
		IsAlive:     true,
		Alignment:   "AI", // Start with AI alignment
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
		if err := glm.datastore.AppendEvent(gameID, event); err != nil {
			log.Printf("CRITICAL: Failed to persist event %s for %s: %v", event.ID, gameID, err)
			// Don't fail the whole process, but log critically
		}
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

		// Send the correctly initialized snapshot.
		// This snapshot now contains the correct phase (SITREP) and role data.
		snapshotEvent := gameActor.CreatePlayerStateUpdateEvent(playerID)
		playerActor.SendServerMessage(snapshotEvent)
	}

	// 3. Then broadcast the granular events to the appropriate players.
	for _, event := range initialEvents {
		// Private events are sent to a specific player
		if event.PlayerID != "" {
			if playerActor, exists := playerActors[event.PlayerID]; exists {
				playerActor.SendServerMessage(event)
			}
		} else { // Public events are broadcast to all players in the game
			for _, playerActor := range playerActors {
				playerActor.SendServerMessage(event)
			}
		}
	}

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


// Helper methods for token management (copied from original LobbyManager)
func (glm *GameLifecycleManager) generateSessionTokenWithLobbyInfo(lobbyID, playerID, playerName, playerAvatar, lobbyName string, isHost bool) (string, error) {
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

// Stop gracefully shuts down the manager
func (glm *GameLifecycleManager) Stop() {
	log.Println("GameLifecycleManager: Shutting down")
	close(glm.stopChannel)
}

// JoinLobby creates credentials for joining an existing lobby via HTTP
func (glm *GameLifecycleManager) JoinLobby(lobbyID, playerName, playerAvatar string) (string, string, error) {
	// Check if lobby exists
	glm.mutex.RLock()
	lobby, exists := glm.lobbies[lobbyID]
	glm.mutex.RUnlock()

	if !exists {
		return "", "", fmt.Errorf("lobby not found")
	}

	// Check if lobby can accept players
	if !lobby.CanStart() && len(lobby.GetPlayerActors()) >= lobby.MaxPlayers {
		return "", "", fmt.Errorf("lobby is full")
	}

	// Generate player ID and token
	playerID := fmt.Sprintf("player_%s_%d", playerName, time.Now().UnixNano())
	sessionToken, err := glm.generateSessionTokenWithLobbyInfo(lobbyID, playerID, playerName, playerAvatar, "", false)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	return playerID, sessionToken, nil
}

// SendActionToGame forwards an action to the appropriate GameActor
func (glm *GameLifecycleManager) SendActionToGame(gameID string, action core.Action) error {
	glm.mutex.RLock()
	gameActor, exists := glm.gameActors[gameID]
	glm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("game not found: %s", gameID)
	}

	// Send action to GameActor (this returns a channel, but we don't wait for the result)
	resultChan := gameActor.PostAction(action)
	go func() {
		// Handle the result asynchronously to prevent blocking
		result := <-resultChan
		if result.Error != nil {
			log.Printf("GameLifecycleManager: Error processing action in game %s: %v", gameID, result.Error)
		}
	}()

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

	// Check if token matches player and is not expired
	return token.PlayerID == playerID &&
		   (token.LobbyID == gameID || gameID == "") && // Allow empty gameID for lobby connections
		   time.Now().Before(token.ExpiresAt)
}

// GetPlayerInfo implements the TokenValidator interface
func (glm *GameLifecycleManager) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	// Find the token for this player
	for _, token := range glm.tokens {
		if token.PlayerID == playerID && (token.LobbyID == gameID || gameID == "") {
			return token.PlayerName, token.PlayerAvatar, nil
		}
	}

	return "", "", fmt.Errorf("player info not found")
}

// GetLobbyList returns a list of active lobbies for the HTTP API
func (glm *GameLifecycleManager) GetLobbyList() []interface{} {
	glm.mutex.RLock()
	defer glm.mutex.RUnlock()

	lobbies := make([]interface{}, 0, len(glm.lobbies))
	for _, lobby := range glm.lobbies {
		players := lobby.GetPlayerActors()
		lobbies = append(lobbies, map[string]interface{}{
			"id":           lobby.ID,
			"name":         lobby.Name,
			"player_count": len(players),
			"max_players":  lobby.MaxPlayers,
			"min_players":  lobby.MinPlayers,
			"can_join":     (lobby.Status == "WAITING" || lobby.Status == "WAITING_FOR_HOST") && len(players) < lobby.MaxPlayers,
			"status":       lobby.Status,
		})
	}

	return lobbies
}