package comms

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/actors"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// WebSocketManager handles WebSocket connections via PlayerActors
type WebSocketManager struct {
	playerActors   map[string]*actors.PlayerActor
	actorsMutex    sync.RWMutex
	ctx            context.Context
	tokenValidator TokenValidator

	// Dependencies for PlayerActors
	lobbyManager   interfaces.LobbyManagerInterface
	sessionManager interfaces.SessionManagerInterface
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
	}
}

// SetDependencies injects the required managers
func (wsm *WebSocketManager) SetDependencies(lobbyManager interfaces.LobbyManagerInterface, sessionManager interfaces.SessionManagerInterface) {
	wsm.lobbyManager = lobbyManager
	wsm.sessionManager = sessionManager
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

	if !wsm.tokenValidator.ValidateSession(gameID, playerID, sessionToken) {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Get player information
	playerName, _, err := wsm.tokenValidator.GetPlayerInfo(gameID, playerID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get player info: %v", err), http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
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
	playerActor := actors.NewPlayerActor(wsm.ctx, playerID, playerName, sessionToken, conn)
	playerActor.SetDependencies(wsm.lobbyManager, wsm.sessionManager)
	
	wsm.playerActors[playerID] = playerActor
	wsm.actorsMutex.Unlock()

	// Start the PlayerActor
	playerActor.Start()

	// In the REST-then-WebSocket flow, automatically join the lobby associated with the token
	// This removes the need for the client to send a separate JOIN_GAME action
	err = wsm.joinLobbyAutomatically(gameID, playerActor)
	if err != nil {
		log.Printf("WebSocketManager: Failed to auto-join lobby %s for player %s: %v", gameID, playerID, err)
		playerActor.Stop()
		wsm.actorsMutex.Lock()
		delete(wsm.playerActors, playerID)
		wsm.actorsMutex.Unlock()
		conn.Close()
		return
	}

	log.Printf("WebSocketManager: Created PlayerActor for %s (%s) and joined lobby %s", playerID, playerName, gameID)
}

// joinLobbyAutomatically handles the automatic lobby joining in REST-then-WebSocket flow
func (wsm *WebSocketManager) joinLobbyAutomatically(lobbyID string, playerActor *actors.PlayerActor) error {
	if wsm.lobbyManager == nil {
		return fmt.Errorf("lobby manager not initialized")
	}

	// Let the LobbyManager handle all the logic for joining
	// This includes checking if the player is the host, lobby status, etc.
	err := wsm.lobbyManager.JoinLobbyWithActor(lobbyID, playerActor)
	if err != nil {
		return fmt.Errorf("failed to auto-join lobby %s: %w", lobbyID, err)
	}

	log.Printf("WebSocketManager: Player %s automatically joined lobby %s", playerActor.GetPlayerID(), lobbyID)
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
	defer wsm.actorsMutex.Unlock()
	if actor, exists := wsm.playerActors[playerID]; exists {
		actor.Stop()
		delete(wsm.playerActors, playerID)
		log.Printf("WebSocketManager: Removed PlayerActor for %s", playerID)
	}
}

// BroadcastToGame sends an event to all PlayerActors in a specific game/lobby
func (wsm *WebSocketManager) BroadcastToGame(gameID string, event core.Event) error {
	wsm.actorsMutex.RLock()
	var actorsToNotify []*actors.PlayerActor
	for _, actor := range wsm.playerActors {
		// For now, we'll rely on the PlayerActor's state to determine if it should receive this event
		// The GameActor or LobbyManager should only send events to relevant players
		actorsToNotify = append(actorsToNotify, actor)
	}
	wsm.actorsMutex.RUnlock()

	// Send to all actors - they will filter based on their state and context
	for _, actor := range actorsToNotify {
		actor.SendServerMessage(event)
	}

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

// Custom errors
var (
	ErrClientDisconnected = fmt.Errorf("client disconnected")
	ErrPlayerNotFound     = fmt.Errorf("player not found")
)