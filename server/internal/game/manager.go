package game

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// SessionManager manages active game sessions and coordinates with GameActors
type SessionManager struct {
	gameSessions map[string]map[string]interfaces.PlayerActorInterface
	gameMutex    sync.RWMutex

	gameActors map[string]interfaces.GameActorInterface
	mutex      sync.RWMutex
	ctx        context.Context

	// Dependencies
	datastore   interfaces.DataStore
	broadcaster interfaces.Broadcaster
	supervisor  interfaces.SupervisorInterface
}

// NewSessionManager creates a new session manager
func NewSessionManager(ctx context.Context, datastore interfaces.DataStore, broadcaster interfaces.Broadcaster, supervisor interfaces.SupervisorInterface) *SessionManager {
	return &SessionManager{
		gameSessions: make(map[string]map[string]interfaces.PlayerActorInterface),
		gameActors:   make(map[string]interfaces.GameActorInterface),
		ctx:          ctx,
		datastore:    datastore,
		broadcaster:  broadcaster,
		supervisor:   supervisor,
	}
}

// SetBroadcaster sets the broadcaster dependency (for dependency injection)
func (sm *SessionManager) SetBroadcaster(broadcaster interfaces.Broadcaster) {
	sm.broadcaster = broadcaster
}

// CreateGameFromLobby implements the atomic transition from lobby to game
func (sm *SessionManager) CreateGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	log.Printf("[SessionManager] Creating game from lobby %s with %d players", lobbyID, len(playerActors))

	// Convert PlayerActors to core.Players for GameActor initialization
	gamePlayers := make(map[string]*core.Player)
	for playerID, playerActor := range playerActors {
		gamePlayers[playerID] = &core.Player{
			ID:       playerID,
			Name:     playerActor.GetPlayerName(),
			JobTitle: "Employee", // Default job title
			IsAlive:  true,
		}
	}

	// Create the GameActor with pre-populated players
	gameActor, err := sm.supervisor.CreateGameWithPlayers(lobbyID, gamePlayers)
	if err != nil {
		return fmt.Errorf("failed to create game actor: %w", err)
	}

	// Store the GameActor and PlayerActor sessions
	sm.mutex.Lock()
	sm.gameActors[lobbyID] = gameActor
	sm.mutex.Unlock()

	sm.gameMutex.Lock()
	sm.gameSessions[lobbyID] = playerActors
	sm.gameMutex.Unlock()

	// Send an INITIALIZE_GAME action to the new actor and wait for the response,
	// which will contain the initial state updates for all players.
	initAction := core.Action{
		Type:     core.ActionType("INITIALIZE_GAME"),
		GameID:   lobbyID,
		PlayerID: "SYSTEM",
	}

	responseChan := gameActor.PostAction(initAction)
	var stateUpdateEvents []core.Event
	select {
	case result := <-responseChan:
		if result.Error != nil {
			return fmt.Errorf("failed to initialize game actor: %w", result.Error)
		}
		stateUpdateEvents = result.Events
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for game actor initialization")
	}

	// Persist all events generated during game initialization
	for _, event := range stateUpdateEvents {
		if err := sm.datastore.AppendEvent(lobbyID, event); err != nil {
			log.Printf("CRITICAL: Failed to persist event %s for %s: %v", event.ID, lobbyID, err)
			// Don't fail the whole process, but log critically
		}
	}

	// Send initial game state snapshots to all players after events are applied
	for playerID, playerActor := range playerActors {
		// Send transition message with the current game state
		playerGameState := gameActor.CreatePlayerStateUpdateEvent(playerID)
		transitionMessage := interfaces.TransitionToGame{
			GameID:    lobbyID,
			GameState: playerGameState.Payload["game_state"],
		}
		playerActor.SendServerMessage(transitionMessage)
		
		// Also send the game state update event for initialization
		playerActor.SendServerMessage(playerGameState)
	}

	// Then broadcast granular events to the appropriate players
	for _, event := range stateUpdateEvents {
		switch event.Type {
		case core.EventRoleAssigned:
			// Private event - send only to the specific player
			if playerActor, exists := playerActors[event.PlayerID]; exists {
				playerActor.SendServerMessage(event)
			}
		case core.EventGameStarted:
			// Public event - broadcast to all players  
			for _, playerActor := range playerActors {
				playerActor.SendServerMessage(event)
			}
		default:
			// Handle other event types as they are implemented
			log.Printf("SessionManager: Broadcasting event %s to all players", event.Type)
			for _, playerActor := range playerActors {
				playerActor.SendServerMessage(event)
			}
		}
	}

	log.Printf("SessionManager: Successfully created game %s and sent transition messages", lobbyID)
	return nil
}

// JoinGame handles a player joining an existing game (for reconnection)
func (sm *SessionManager) JoinGame(gameID string, playerActor interfaces.PlayerActorInterface) error {
	sm.mutex.RLock()
	gameActor, exists := sm.gameActors[gameID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("game not found: %s", gameID)
	}

	// Add player to the session
	sm.gameMutex.Lock()
	if _, ok := sm.gameSessions[gameID]; !ok {
		sm.gameSessions[gameID] = make(map[string]interfaces.PlayerActorInterface)
	}
	sm.gameSessions[gameID][playerActor.GetPlayerID()] = playerActor
	sm.gameMutex.Unlock()

	// Transition the player to the game
	err := playerActor.TransitionToGame(gameID)
	if err != nil {
		return fmt.Errorf("failed to transition player to game: %w", err)
	}

	// Send current game state snapshot
	snapshotEvent := gameActor.CreatePlayerStateUpdateEvent(playerActor.GetPlayerID())
	playerActor.SendServerMessage(snapshotEvent)

	// Send chat history from datastore
	historicalEvents, err := sm.datastore.LoadEvents(gameID, 0)
	if err != nil {
		log.Printf("SessionManager: Error loading event history for reconnecting player %s: %v", playerActor.GetPlayerID(), err)
		return err
	}

	var chatHistory []core.ChatMessage
	for _, event := range historicalEvents {
		if event.Type == core.EventChatMessage {
			if payload, ok := event.Payload["message"].(map[string]interface{}); ok {
				playerName, _ := payload["player_name"].(string)
				content, _ := payload["content"].(string)

				chatMsg := core.ChatMessage{
					ID:         event.ID,
					PlayerID:   event.PlayerID,
					PlayerName: playerName,
					Message:    content,
					Timestamp:  event.Timestamp,
					IsSystem:   false,
				}
				chatHistory = append(chatHistory, chatMsg)
			}
		}
	}

	chatHistoryEvent := core.Event{
		ID:        fmt.Sprintf("chat_history_%d", time.Now().UnixNano()),
		Type:      core.EventType("CHAT_HISTORY_SNAPSHOT"),
		GameID:    gameID,
		PlayerID:  playerActor.GetPlayerID(),
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"messages": chatHistory},
	}
	playerActor.SendServerMessage(chatHistoryEvent)

	return nil
}

// LeaveGame handles a player leaving a game
func (sm *SessionManager) LeaveGame(gameID string, playerID string) error {
	sm.gameMutex.Lock()
	if session, ok := sm.gameSessions[gameID]; ok {
		delete(session, playerID)
	}
	sm.gameMutex.Unlock()

	// Send leave action to the GameActor
	leaveAction := core.Action{
		Type:     core.ActionLeaveGame,
		PlayerID: playerID,
		GameID:   gameID,
	}

	return sm.SendActionToGame(gameID, leaveAction)
}

// SendActionToGame forwards an action to the appropriate GameActor and handles persistence/broadcasting
func (sm *SessionManager) SendActionToGame(gameID string, action core.Action) error {
	sm.mutex.RLock()
	gameActor, exists := sm.gameActors[gameID]
	sm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("game not found: %s", gameID)
	}

	// Post the action and get the response channel (non-blocking)
	responseChan := gameActor.PostAction(action)

	// Wait for the result with a timeout
	var events []core.Event
	select {
	case result := <-responseChan:
		if result.Error != nil {
			return fmt.Errorf("failed to process action: %w", result.Error)
		}
		events = result.Events
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout waiting for response from GameActor %s", gameID)
	}

	// Persist all events
	for _, event := range events {
		if err := sm.datastore.AppendEvent(gameID, event); err != nil {
			log.Printf("SessionManager: CRITICAL - FAILED TO PERSIST EVENT %s: %v", event.ID, err)
			return fmt.Errorf("failed to persist event: %w", err)
		}
	}

	// Look up players from internal state
	sm.gameMutex.RLock()
	playerActors, sessionExists := sm.gameSessions[gameID]
	sm.gameMutex.RUnlock()

	if !sessionExists {
		return fmt.Errorf("no active session found for game %s to broadcast events", gameID)
	}

	// Broadcast events to all players in the game
	for _, event := range events {
		if playerActor, exists := playerActors[event.PlayerID]; exists {
			playerActor.SendServerMessage(event)
		} else if event.PlayerID == "" { // Public event
			for _, pa := range playerActors {
				pa.SendServerMessage(event)
			}
		}
	}

	return nil
}

// RemoveGame removes a game from management (called when game ends)
func (sm *SessionManager) RemoveGame(gameID string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, exists := sm.gameActors[gameID]; exists {
		delete(sm.gameActors, gameID)
		log.Printf("SessionManager: Removed game %s", gameID)
	}

	// Also remove from supervisor
	sm.supervisor.RemoveGame(gameID)

	// Also remove from sessions
	sm.gameMutex.Lock()
	delete(sm.gameSessions, gameID)
	sm.gameMutex.Unlock()
}

// GetGameActor returns a GameActor by ID
func (sm *SessionManager) GetGameActor(gameID string) (interfaces.GameActorInterface, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	gameActor, exists := sm.gameActors[gameID]
	return gameActor, exists
}

// GetStats returns statistics about active games
func (sm *SessionManager) GetStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	return map[string]interface{}{
		"active_games": len(sm.gameActors),
	}
}
