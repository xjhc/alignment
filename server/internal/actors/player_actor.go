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
	"github.com/xjhc/alignment/server/internal/events"
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
	stopOnce      sync.Once // <-- ADD THIS

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies (will be injected)
	lifecycleManager interfaces.GameLifecycleManagerInterface
	eventBus         *events.EventBus
}

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
func (pa *PlayerActor) SetDependencies(lifecycleManager interfaces.GameLifecycleManagerInterface, eventBus *events.EventBus) {
	pa.lifecycleManager = lifecycleManager
	pa.eventBus = eventBus
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
	// Use sync.Once to ensure the shutdown logic runs exactly once.
	pa.stopOnce.Do(func() {
		log.Printf("[PlayerActor/%s] Stopping", pa.playerID)

		// Publish disconnection event if we have event bus
		if pa.eventBus != nil {
			pa.eventBus.Publish(events.PlayerDisconnectedEvent{
				PlayerID: pa.playerID,
				LobbyID:  pa.lobbyID,
				GameID:   pa.gameID,
			})
		}

		pa.cancel()
		close(pa.send)
		close(pa.shutdown)
	})
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
	// Prevent sending on a closed channel
	if pa.ctx.Err() != nil {
		log.Printf("[PlayerActor/%s] Dropped message, context is done.", pa.playerID)
		return
	}

	// Marshal the message to JSON bytes.
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("[PlayerActor/%s] Failed to marshal server message: %v", pa.playerID, err)
		return
	}

	select {
	case pa.send <- data:
		// Message sent successfully
	case <-pa.ctx.Done():
		log.Printf("[PlayerActor/%s] Dropped server message, context is done.", pa.playerID)
	default:
		log.Printf("[PlayerActor/%s] Send channel is full, dropping server message.", pa.playerID)
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

	if pa.lifecycleManager == nil {
		pa.sendError("Lifecycle manager not available")
		return
	}

	// In the new architecture, lobby creation happens via HTTP, not here
	pa.sendError("Lobby creation should be done via HTTP API, not WebSocket")
	return
}

// handleJoinLobby joins an existing lobby
func (pa *PlayerActor) handleJoinLobby(action core.Action) {
	lobbyID, ok := action.Payload["lobby_id"].(string)
	if !ok || lobbyID == "" {
		pa.sendError("Missing lobby_id in join request")
		return
	}

	if pa.lifecycleManager == nil {
		pa.sendError("Lifecycle manager not available")
		return
	}

	err := pa.lifecycleManager.JoinLobbyWithActor(lobbyID, pa)
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
		// Publish disconnection event for lifecycle manager to handle
		if pa.eventBus != nil {
			pa.eventBus.Publish(events.PlayerDisconnectedEvent{
				PlayerID: pa.playerID,
				LobbyID:  pa.lobbyID,
				GameID:   "",
			})
		}
		pa.TransitionToIdle()
	case StateInGame:
		// Publish disconnection event for lifecycle manager to handle
		if pa.eventBus != nil {
			pa.eventBus.Publish(events.PlayerDisconnectedEvent{
				PlayerID: pa.playerID,
				LobbyID:  "",
				GameID:   pa.gameID,
			})
		}
		pa.TransitionToIdle()
	default:
		pa.sendError(fmt.Sprintf("Cannot leave game in state %s", state))
	}
}

// handleStartGame starts the game (only available to lobby host)
func (pa *PlayerActor) handleStartGame(action core.Action) {
	if pa.lifecycleManager == nil {
		pa.sendError("Lifecycle manager not available")
		return
	}

	// Launch the potentially long-running StartGame process in a new goroutine
	// to prevent blocking the PlayerActor's main processing loop.
	go func() {
		log.Printf("[PlayerActor/%s] Dispatching START_GAME for lobby %s", pa.playerID, pa.lobbyID)
		err := pa.lifecycleManager.StartGame(pa.lobbyID, pa.playerID)
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
	// Handle special non-game actions first
	switch action.Type {
	case "ping":
		// Handle WebSocket heartbeat - just respond with pong
		pa.sendPong()
		return
	case core.ActionLeaveGame:
		pa.handleLeaveGame(action)
		return
	}

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

	// Handle client action types that need to be mapped to core action types
	actionType := action.Type
	switch action.Type {
	case "POST_CHAT_MESSAGE":
		actionType = core.ActionSendMessage
	case "UPDATE_STATUS":
		actionType = core.ActionSetSlackStatus
	}

	// List of valid game actions that should be forwarded to the SessionManager
	validGameActions := map[core.ActionType]bool{
		core.ActionSendMessage:         true,
		core.ActionSubmitVote:          true,
		core.ActionSubmitNightAction:   true,
		core.ActionMineTokens:          true,
		core.ActionSubmitPulseCheck:    true,
		core.ActionUseAbility:          true,
		core.ActionAttemptConversion:   true,
		core.ActionExtendDiscussion:    true,
		core.ActionRunAudit:            true,
		core.ActionOverclockServers:    true,
		core.ActionIsolateNode:         true,
		core.ActionPerformanceReview:   true,
		core.ActionReallocateBudget:    true,
		core.ActionPivot:               true,
		core.ActionDeployHotfix:        true,
		core.ActionSetSlackStatus:      true,
		core.ActionProjectMilestones:   true,
		core.ActionReconnect:           true,
	}

	// Check if this is a valid game action
	if !validGameActions[actionType] {
		pa.sendError(fmt.Sprintf("Action %s not allowed in InGame state", action.Type))
		return
	}

	// Forward valid game actions to GameLifecycleManager
	if pa.lifecycleManager == nil {
		pa.sendError("Lifecycle manager not available")
		return
	}

	// Update the action type if it was mapped
	action.Type = actionType
	// Ensure the action is properly attributed to this player and game
	action.PlayerID = pa.playerID
	action.GameID = pa.gameID

	err := pa.lifecycleManager.SendActionToGame(pa.gameID, action)
	if err != nil {
		log.Printf("[PlayerActor/%s] Failed to send action to game: %v", pa.playerID, err)
		pa.sendError(fmt.Sprintf("Failed to process action: %v", err))
	}
}

// handleServerMessage processes messages from server components
func (pa *PlayerActor) handleServerMessage(message interface{}) {
	var data []byte
	var err error

	// Marshal the message to JSON. If it's already []byte, use it directly.
	if bytes, ok := message.([]byte); ok {
		data = bytes
	} else {
		data, err = json.Marshal(message)
		if err != nil {
			log.Printf("[PlayerActor/%s] Failed to marshal server message: %v", pa.playerID, err)
			return
		}
	}

	// Send the marshalled data to the client
	select {
	case pa.send <- data:
	case <-pa.ctx.Done():
		log.Printf("[PlayerActor/%s] Dropped server message, context is done.", pa.playerID)
	default:
		log.Printf("[PlayerActor/%s] Send channel is full, dropping server message.", pa.playerID)
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

	// Publish disconnection events for the lifecycle manager to handle
	switch state {
	case StateInLobby:
		log.Printf("[PlayerActor/%s] Leaving lobby %s", pa.playerID, pa.lobbyID)
		if pa.eventBus != nil {
			pa.eventBus.Publish(events.PlayerDisconnectedEvent{
				PlayerID: pa.playerID,
				LobbyID:  pa.lobbyID,
				GameID:   "",
			})
		}
	case StateInGame:
		log.Printf("[PlayerActor/%s] Leaving game %s", pa.playerID, pa.gameID)
		if pa.eventBus != nil {
			pa.eventBus.Publish(events.PlayerDisconnectedEvent{
				PlayerID: pa.playerID,
				LobbyID:  "",
				GameID:   pa.gameID,
			})
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
	pa.SendServerMessage(event)
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

// sendPong responds to a ping with a pong message
func (pa *PlayerActor) sendPong() {
	event := core.Event{
		Type:      "pong",
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{},
	}

	pa.sendEvent(event)
}