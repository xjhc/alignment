package actors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/lobby"
)

// Use PlayerState from interfaces package
type PlayerState = interfaces.PlayerState

const (
	StateIdle    = interfaces.StateIdle
	StateInLobby = interfaces.StateInLobby
	StateInGame  = interfaces.StateInGame
)

// PlayerActor manages a single player's session and WebSocket connection
type PlayerActor struct {
	playerID     string
	playerName   string
	sessionToken string
	conn         *websocket.Conn
	send         chan []byte
	state        PlayerState
	stateMutex   sync.RWMutex

	// Current context
	lobbyID string
	gameID  string

	// Communication channels
	mailbox       chan interface{} // From client WebSocket
	serverMailbox chan interface{} // From server components
	shutdown      chan struct{}

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies (will be injected)
	lobbyManager   interfaces.LobbyManagerInterface
	sessionManager interfaces.SessionManagerInterface
}

// Message types are now defined in interfaces package

// NewPlayerActor creates a new PlayerActor for a WebSocket connection
func NewPlayerActor(ctx context.Context, playerID, playerName, sessionToken string, conn *websocket.Conn) *PlayerActor {
	actorCtx, cancel := context.WithCancel(ctx)

	return &PlayerActor{
		playerID:      playerID,
		playerName:    playerName,
		sessionToken:  sessionToken,
		conn:          conn,
		send:          make(chan []byte, 256),
		state:         StateIdle,
		mailbox:       make(chan interface{}, 100),
		serverMailbox: make(chan interface{}, 100),
		shutdown:      make(chan struct{}),
		ctx:           actorCtx,
		cancel:        cancel,
	}
}

// SetDependencies injects the required managers
func (pa *PlayerActor) SetDependencies(lobbyManager interfaces.LobbyManagerInterface, sessionManager interfaces.SessionManagerInterface) {
	pa.lobbyManager = lobbyManager
	pa.sessionManager = sessionManager
}

// Start begins the PlayerActor's processing loops
func (pa *PlayerActor) Start() {
	log.Printf("[PlayerActor/%s] Starting", pa.playerID)

	// Start WebSocket read pump
	go pa.readPump()

	// Start WebSocket write pump
	go pa.writePump()

	// Start main processing loop
	go pa.processLoop()
}

// Stop gracefully shuts down the PlayerActor
func (pa *PlayerActor) Stop() {
	log.Printf("[PlayerActor/%s] Stopping", pa.playerID)
	pa.cancel()
	close(pa.send)
	close(pa.shutdown)
}

// GetPlayerID returns the player's ID
func (pa *PlayerActor) GetPlayerID() string {
	return pa.playerID
}

// GetPlayerName returns the player's name
func (pa *PlayerActor) GetPlayerName() string {
	return pa.playerName
}

// GetSessionToken returns the player's session token
func (pa *PlayerActor) GetSessionToken() string {
	return pa.sessionToken
}

// GetState returns the current player state
func (pa *PlayerActor) GetState() PlayerState {
	pa.stateMutex.RLock()
	defer pa.stateMutex.RUnlock()
	return pa.state
}

func (pa *PlayerActor) GetGameID() string {
	pa.stateMutex.RLock()
	defer pa.stateMutex.RUnlock()
	return pa.gameID
}

func (pa *PlayerActor) GetLobbyID() string {
	pa.stateMutex.RLock()
	defer pa.stateMutex.RUnlock()
	return pa.lobbyID
}

// isConnectionValid checks if the player's connection is in a valid state
func (pa *PlayerActor) isConnectionValid() bool {
	// Check if WebSocket connection is still active
	if pa.conn == nil {
		return false
	}

	// Check if actor hasn't been cancelled
	select {
	case <-pa.ctx.Done():
		return false
	default:
	}

	// Additional validation could include checking session token expiry
	// but that's handled by the LobbyManager's token validation
	return true
}

// SendServerMessage sends a message from server components to this player
func (pa *PlayerActor) SendServerMessage(message interface{}) {
	select {
	case pa.serverMailbox <- message:
	case <-pa.ctx.Done():
		log.Printf("[PlayerActor/%s] Attempted to send server message to stopped actor", pa.playerID)
	}
}

// TransitionToLobby transitions the player to lobby state
func (pa *PlayerActor) TransitionToLobby(lobbyID string) error {
	pa.stateMutex.Lock()
	defer pa.stateMutex.Unlock()

	if pa.state != StateIdle {
		return fmt.Errorf("invalid state transition from %s to InLobby", pa.state)
	}

	pa.state = StateInLobby
	pa.lobbyID = lobbyID
	log.Printf("[PlayerActor/%s] Transitioned to InLobby (lobby: %s)", pa.playerID, lobbyID)

	return nil
}

// TransitionToGame transitions the player to game state
func (pa *PlayerActor) TransitionToGame(gameID string) error {
	pa.stateMutex.Lock()
	defer pa.stateMutex.Unlock()

	if pa.state != StateInLobby {
		return fmt.Errorf("invalid state transition from %s to InGame", pa.state)
	}

	pa.state = StateInGame
	pa.gameID = gameID
	pa.lobbyID = "" // Clear lobby reference
	log.Printf("[PlayerActor/%s] Transitioned to InGame (game: %s)", pa.playerID, gameID)

	return nil
}

// TransitionToIdle transitions the player back to idle state
func (pa *PlayerActor) TransitionToIdle() error {
	pa.stateMutex.Lock()
	defer pa.stateMutex.Unlock()

	oldState := pa.state
	pa.state = StateIdle
	pa.lobbyID = ""
	pa.gameID = ""
	log.Printf("[PlayerActor/%s] Transitioned from %s to Idle", pa.playerID, oldState)

	return nil
}

// readPump handles incoming WebSocket messages from the client
func (pa *PlayerActor) readPump() {
	defer func() {
		pa.conn.Close()
		pa.Stop()
	}()

	// Set read deadline and pong handler for heartbeat
	pa.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	pa.conn.SetPongHandler(func(string) error {
		pa.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		select {
		case <-pa.ctx.Done():
			return
		default:
			_, message, err := pa.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[PlayerActor/%s] WebSocket error: %v", pa.playerID, err)
				}
				return
			}

			// Parse and route the action
			var action core.Action
			if err := json.Unmarshal(message, &action); err != nil {
				log.Printf("[PlayerActor/%s] Failed to parse action: %v", pa.playerID, err)
				continue
			}

			// Set the player ID on the action
			action.PlayerID = pa.playerID

			// Send to processing loop
			select {
			case pa.mailbox <- action:
			case <-pa.ctx.Done():
				return
			}
		}
	}
}

// writePump handles outgoing WebSocket messages to the client
func (pa *PlayerActor) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		pa.conn.Close()
	}()

	for {
		select {
		case message, ok := <-pa.send:
			pa.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				pa.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := pa.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-pa.ctx.Done():
			pa.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		case <-ticker.C:
			select {
			case <-pa.ctx.Done():
				return
			default:
				pa.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := pa.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}
}

// processLoop is the main message processing loop
func (pa *PlayerActor) processLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PlayerActor/%s] Panic recovered: %v", pa.playerID, r)
		}
		pa.handleDisconnect()
	}()

	for {
		select {
		case <-pa.ctx.Done():
			return
		case <-pa.shutdown:
			return
		case msg := <-pa.mailbox:
			// Handle client actions
			if action, ok := msg.(core.Action); ok {
				pa.handleClientAction(action)
			}
		case msg := <-pa.serverMailbox:
			// Handle server messages
			pa.handleServerMessage(msg)
		}
	}
}

// handleClientAction routes client actions based on current state
func (pa *PlayerActor) handleClientAction(action core.Action) {
	pa.stateMutex.RLock()
	currentState := pa.state
	pa.stateMutex.RUnlock()

	// Validate connection state before processing actions
	if !pa.isConnectionValid() {
		pa.sendError("Invalid connection state")
		return
	}

	log.Printf("[PlayerActor/%s] Handling action %s in state %s", pa.playerID, action.Type, currentState)

	// Route actions based on current state to enforce state machine
	switch currentState {
	case StateIdle:
		pa.handleIdleAction(action)
	case StateInLobby:
		pa.handleLobbyAction(action)
	case StateInGame:
		pa.handleGameAction(action)
	default:
		pa.sendError(fmt.Sprintf("Invalid player state: %s", currentState))
	}
}

// handleIdleAction handles actions valid in Idle state
func (pa *PlayerActor) handleIdleAction(action core.Action) {
	switch action.Type {
	case core.ActionCreateGame:
		pa.handleCreateLobby(action)
	case core.ActionJoinGame:
		pa.handleJoinLobby(action)
	default:
		pa.sendError(fmt.Sprintf("Action %s not allowed in Idle state", action.Type))
	}
}

// handleLobbyAction handles actions valid in InLobby state
func (pa *PlayerActor) handleLobbyAction(action core.Action) {
	switch action.Type {
	case core.ActionLeaveGame:
		pa.handleLeaveGame(action)
	case core.ActionStartGame:
		pa.handleStartGame(action)
	case core.ActionSendMessage:
		pa.handleLobbyChat(action)
	default:
		pa.sendError(fmt.Sprintf("Action %s not allowed in InLobby state", action.Type))
	}
}

// handleCreateLobby creates a new lobby
func (pa *PlayerActor) handleCreateLobby(action core.Action) {
	lobbyName, _ := action.Payload["lobby_name"].(string)
	if lobbyName == "" {
		lobbyName = fmt.Sprintf("%s's Game", pa.playerName)
	}

	if pa.lobbyManager == nil {
		pa.sendError("Lobby manager not available")
		return
	}

	lobbyID, err := pa.lobbyManager.CreateLobby(pa, lobbyName)
	if err != nil {
		pa.sendError(fmt.Sprintf("Failed to create lobby: %v", err))
		return
	}

	// The lobby manager will call TransitionToLobby on success
	log.Printf("[PlayerActor/%s] Created lobby %s", pa.playerID, lobbyID)
}

// handleJoinLobby joins an existing lobby
func (pa *PlayerActor) handleJoinLobby(action core.Action) {
	lobbyID, ok := action.Payload["lobby_id"].(string)
	if !ok || lobbyID == "" {
		pa.sendError("Missing lobby_id in join request")
		return
	}

	if pa.lobbyManager == nil {
		pa.sendError("Lobby manager not available")
		return
	}

	err := pa.lobbyManager.JoinLobbyWithActor(lobbyID, pa)
	if err != nil {
		pa.sendError(fmt.Sprintf("Failed to join lobby: %v", err))
		return
	}

	log.Printf("[PlayerActor/%s] Joined lobby %s", pa.playerID, lobbyID)
}

// handleLeaveGame handles leaving current context (lobby or game)
func (pa *PlayerActor) handleLeaveGame(action core.Action) {
	state := pa.GetState()

	switch state {
	case StateInLobby:
		if pa.lobbyManager != nil {
			pa.lobbyManager.LeaveLobby(pa.lobbyID, pa.playerID)
		}
		pa.TransitionToIdle()
	case StateInGame:
		if pa.sessionManager != nil {
			pa.sessionManager.LeaveGame(pa.gameID, pa.playerID)
		}
		pa.TransitionToIdle()
	default:
		pa.sendError(fmt.Sprintf("Cannot leave game in state %s", state))
	}
}

// handleStartGame starts the game (only available to lobby host)
func (pa *PlayerActor) handleStartGame(action core.Action) {
	if pa.lobbyManager == nil {
		pa.sendError("Lobby manager not available")
		return
	}

	// BUG FIX: Launch the potentially long-running StartGame process in a new goroutine
	// to prevent blocking the PlayerActor's main processing loop.
	go func() {
		log.Printf("[PlayerActor/%s] Dispatching START_GAME for lobby %s", pa.playerID, pa.lobbyID)
		err := pa.lobbyManager.StartGame(pa.playerID, pa.lobbyID)
		if err != nil {
			// If an error occurs (e.g., not enough players), send it back to the client.
			// This is safe to call from a goroutine as it sends to a channel.
			pa.sendError(fmt.Sprintf("Failed to start game: %v", err))
		}
	}()
	log.Printf("[PlayerActor/%s] Dispatched START_GAME action for lobby %s", pa.playerID, pa.lobbyID)
}

// handleLobbyChat handles chat messages in lobby
func (pa *PlayerActor) handleLobbyChat(action core.Action) {
	message, ok := action.Payload["message"].(string)
	if !ok || message == "" {
		pa.sendError("Missing message in chat action")
		return
	}

	// TODO: Implement lobby chat broadcasting
	log.Printf("[PlayerActor/%s] Lobby chat: %s", pa.playerID, message)
}

// handleGameAction forwards actions to the game
func (pa *PlayerActor) handleGameAction(action core.Action) {
	// Validate that player is actually in a game
	if pa.gameID == "" {
		pa.sendError("Player not in a game")
		return
	}

	// Validate that the action's gameID matches the player's current game
	if action.GameID != "" && action.GameID != pa.gameID {
		pa.sendError("Action gameID does not match player's current game")
		return
	}

	// Handle special game-level actions that don't go to GameActor
	switch action.Type {
	case core.ActionLeaveGame:
		pa.handleLeaveGame(action)
		return
	case core.ActionSubmitVote, core.ActionSubmitNightAction, core.ActionMineTokens, core.ActionSendMessage:
		// Valid game actions - forward to SessionManager
		if pa.sessionManager == nil {
			pa.sendError("Session manager not available")
			return
		}

		// Ensure the action is properly attributed to this player and game
		action.PlayerID = pa.playerID
		action.GameID = pa.gameID

		err := pa.sessionManager.SendActionToGame(pa.gameID, action)
		if err != nil {
			log.Printf("[PlayerActor/%s] Failed to send action to game: %v", pa.playerID, err)
		}
	default:
		// Invalid actions for InGame state
		pa.sendError(fmt.Sprintf("Action %s not allowed in InGame state", action.Type))
	}
}

// handleServerMessage processes messages from server components
func (pa *PlayerActor) handleServerMessage(message interface{}) {
	switch msg := message.(type) {
	case lobby.LobbyStateUpdate:
		pa.sendLobbyStateUpdate(msg)
	case interfaces.GameStateSnapshot:
		pa.sendGameStateSnapshot(msg)
	case interfaces.TransitionToGame:
		pa.handleTransitionToGame(msg)
	case core.Event:
		pa.sendEvent(msg)
	default:
		log.Printf("[PlayerActor/%s] Unknown server message type: %T", pa.playerID, message)
	}
}

// handleTransitionToGame handles the atomic transition from lobby to game
func (pa *PlayerActor) handleTransitionToGame(transition interfaces.TransitionToGame) {
	log.Printf("[PlayerActor/%s] Received TransitionToGame for game %s. Forwarding snapshot.", pa.playerID, transition.GameID)
	err := pa.TransitionToGame(transition.GameID)
	if err != nil {
		log.Printf("[PlayerActor/%s] Failed to transition to game: %v", pa.playerID, err)
		return
	}

	// First, send a generic GAME_STARTED event to signal the UI to transition.
	gameStartedEvent := core.Event{
		Type:      core.EventGameStarted,
		GameID:    transition.GameID,
		PlayerID:  pa.playerID, // Private, to this player
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"game_id": transition.GameID},
	}
	pa.sendEvent(gameStartedEvent)

	// THEN, send the game state snapshot with role info etc.
	pa.sendGameStateSnapshot(interfaces.GameStateSnapshot{
		GameID:    transition.GameID,
		GameState: transition.GameState,
	})
}

// handleDisconnect cleans up when player disconnects
func (pa *PlayerActor) handleDisconnect() {
	state := pa.GetState()
	log.Printf("[PlayerActor/%s] Disconnecting in state %s", pa.playerID, state)

	switch state {
	case StateInLobby:
		if pa.lobbyManager != nil {
			pa.lobbyManager.LeaveLobby(pa.lobbyID, pa.playerID)
		}
	case StateInGame:
		if pa.sessionManager != nil {
			pa.sessionManager.LeaveGame(pa.gameID, pa.playerID)
		}
	}
}

// sendLobbyStateUpdate sends lobby state to client
func (pa *PlayerActor) sendLobbyStateUpdate(update lobby.LobbyStateUpdate) {
	event := core.Event{
		Type:      "LOBBY_STATE_UPDATE",
		GameID:    update.LobbyID,
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"lobby_id":    update.LobbyID,
			"players":     update.Players,
			"host_id":     update.HostID,
			"can_start":   update.CanStart,
			"name":        update.LobbyName,
			"max_players": 8, // TODO: Make configurable
		},
	}

	pa.sendEvent(event)
}

// sendGameStateSnapshot sends game state to client
func (pa *PlayerActor) sendGameStateSnapshot(snapshot interfaces.GameStateSnapshot) {
	// BUG FIX: The snapshot now contains the already-prepared player-specific view.
	// We just need to wrap it in the GAME_STATE_UPDATE event.
	event := core.Event{
		Type:      "GAME_STATE_UPDATE",
		GameID:    snapshot.GameID,
		PlayerID:  pa.playerID, // Private event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"game_state": snapshot.GameState,
		},
	}

	pa.sendEvent(event)
}

// sendEvent sends a core event to the client
func (pa *PlayerActor) sendEvent(event core.Event) {
	// Handle nil connection (for testing)
	if pa.conn == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("[PlayerActor/%s] Failed to marshal event: %v", pa.playerID, err)
		return
	}

	select {
	case pa.send <- data:
	case <-pa.ctx.Done():
		log.Printf("[PlayerActor/%s] Dropped message, context is done.", pa.playerID)
	default:
		log.Printf("[PlayerActor/%s] Send channel is full, dropping message.", pa.playerID)
	}
}

// sendError sends an error message to the client
func (pa *PlayerActor) sendError(message string) {
	log.Printf("Sending error to player %s: %s", pa.playerID, message)
	event := core.Event{
		Type:      core.EventSystemMessage,
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": message,
			"error":   true,
		},
	}

	pa.sendEvent(event)
}
