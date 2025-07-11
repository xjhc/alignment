package actors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/helpers"
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

// Input validation constants
const (
	MAX_PLAYER_NAME_LENGTH = 50
	MAX_CHAT_MESSAGE_LENGTH = 280
	MAX_STATUS_MESSAGE_LENGTH = 100
	BATCH_ID_CACHE_TTL        = 30 * time.Second
)

// Regular expressions for input validation
var (
	// Allow alphanumeric, spaces, basic punctuation, and common emoji
	safeCharPattern = regexp.MustCompile(`^[\w\s\.\,\!\?\-\(\)\[\]\{\}'"@#&*+=|\\/:;<>~` + "`" + `\r\n\t]+$`)
	// For player names, be more restrictive
	playerNamePattern = regexp.MustCompile(`^[\w\s\-\.]+$`)
)

// PlayerActor manages a single player's session and WebSocket connection
type PlayerActor struct {
	playerID     string
	playerName   string
	playerAvatar string
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

	// Deduplication cache
	recentBatchIDs map[string]time.Time

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Rate limiting for chat messages
	chatLimiter *rate.Limiter
	// Rate limiting for general actions
	generalLimiter *rate.Limiter

	// Dependencies (will be injected)
	lifecycleManager interfaces.GameLifecycleManagerInterface
	eventBus         *events.EventBus
	partyManager     interfaces.PartyManagerInterface
	actionSemaphore  *semaphore.Weighted
}

// NewPlayerActor creates a new PlayerActor for a WebSocket connection
func NewPlayerActor(ctx context.Context, playerID, playerName, playerAvatar, sessionToken string, conn *websocket.Conn) *PlayerActor {
	actorCtx, cancel := context.WithCancel(ctx)

	// Sanitize and validate player name
	sanitizedPlayerName := sanitizeString(playerName, MAX_PLAYER_NAME_LENGTH)
	if err := validatePlayerName(sanitizedPlayerName); err != nil {
		log.Printf("Warning: Invalid player name '%s': %v. Using fallback.", playerName, err)
		sanitizedPlayerName = fmt.Sprintf("Player_%s", playerID[:8]) // Fallback to safe name
	}

	// Allow an average of 2 messages per second,
	// with a burst capacity of 5 messages.
	chatLimiter := rate.NewLimiter(rate.Limit(2), 5)

	// Allow 10 actions per second, with a burst of 20.
	// This is generous for a human but stops a simple script.
	generalLimiter := rate.NewLimiter(10, 20)

	return &PlayerActor{
		playerID:      playerID,
		playerName:    sanitizedPlayerName,
		playerAvatar:  playerAvatar,
		sessionToken:  sessionToken,
		conn:          conn,
		send:          make(chan []byte, 256),
		state:         StateIdle,
		mailbox:       make(chan interface{}, 100),
		serverMailbox: make(chan interface{}, 100),
		shutdown:       make(chan struct{}),
		recentBatchIDs: make(map[string]time.Time),
		ctx:            actorCtx,
		cancel:         cancel,
		chatLimiter:    chatLimiter,
		generalLimiter: generalLimiter,
		actionSemaphore: nil, // Will be set via SetActionSemaphore
	}
}

// SetDependencies injects the required managers
func (pa *PlayerActor) SetDependencies(lifecycleManager interfaces.GameLifecycleManagerInterface, eventBus *events.EventBus, partyManager interfaces.PartyManagerInterface) {
	pa.lifecycleManager = lifecycleManager
	pa.eventBus = eventBus
	pa.partyManager = partyManager
}

// SetActionSemaphore injects the global action semaphore
func (pa *PlayerActor) SetActionSemaphore(sem *semaphore.Weighted) {
	pa.actionSemaphore = sem
}

// Start begins the PlayerActor's processing loops
func (pa *PlayerActor) Start() {
	log.Printf("[PlayerActor/%s] Starting", pa.playerID)

	// Start WebSocket read pump with panic protection
	helpers.GoSafe(pa.ctx, func(_ context.Context) { pa.readPump() })

	// Start WebSocket write pump with panic protection
	helpers.GoSafe(pa.ctx, func(_ context.Context) { pa.writePump() })

	// Start main processing loop with panic protection
	helpers.GoSafe(pa.ctx, func(_ context.Context) { pa.processLoop() })
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

// GetPlayerAvatar returns the player's avatar
func (pa *PlayerActor) GetPlayerAvatar() string {
	return pa.playerAvatar
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

	// Set maximum message size to prevent memory exhaustion attacks
	pa.conn.SetReadLimit(4096) // 4KB limit per message

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
	cleanupTicker := time.NewTicker(30 * time.Second)
	defer cleanupTicker.Stop()

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
			case <-cleanupTicker.C:
				pa.cleanupRecentBatchIDs()
		}
	}
}

// handleClientAction routes client actions based on current state
func (pa *PlayerActor) handleClientAction(action core.Action) {
	// Apply global admission control if semaphore is available
	if pa.actionSemaphore != nil {
		if err := pa.actionSemaphore.Acquire(context.Background(), 1); err != nil {
			// Server is at max capacity, drop the request
			pa.sendError("Server is busy, please try again.")
			return
		}
		defer pa.actionSemaphore.Release(1) // IMPORTANT: Release the semaphore when done
	}

	// Apply general rate limiting before any other processing
	if !pa.generalLimiter.Allow() {
		log.Printf("Player %s exceeded general rate limit. Disconnecting.", pa.playerID)
		pa.Stop() // Aggressive but effective response
		return
	}

	pa.stateMutex.RLock()
	currentState := pa.state
	pa.stateMutex.RUnlock()

	// Validate connection state before processing actions
	if !pa.isConnectionValid() {
		pa.sendError("Invalid connection state")
		return
	}

	log.Printf("[PlayerActor/%s] Handling action %s in state %s", pa.playerID, action.Type, currentState)

	// Handle party actions globally (not state-dependent)
	switch action.Type {
	case core.ActionCreateParty:
		pa.handleCreateParty(action)
		return
	case core.ActionInviteToParty:
		pa.handleInviteToParty(action)
		return
	case core.ActionJoinParty:
		pa.handleJoinParty(action)
		return
	case core.ActionLeaveParty:
		pa.handleLeaveParty(action)
		return
	}

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
	case core.ActionSyncLobbyState:
		pa.handleSyncLobbyState(action)
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
	helpers.GoSafe(pa.ctx, func(_ context.Context) {
		log.Printf("[PlayerActor/%s] Dispatching START_GAME for lobby %s", pa.playerID, pa.lobbyID)
		err := pa.lifecycleManager.StartGame(pa.lobbyID, pa.playerID)
		if err != nil {
			// If an error occurs (e.g., not enough players), send it back to the client.
			// This is safe to call from a goroutine as it sends to a channel.
			pa.sendError(fmt.Sprintf("Failed to start game: %v", err))
		}
	})
	log.Printf("[PlayerActor/%s] Dispatched START_GAME action for lobby %s", pa.playerID, pa.lobbyID)
}

// handleLobbyChat handles chat messages in lobby
func (pa *PlayerActor) handleLobbyChat(action core.Action) {
	// Apply rate limiting for lobby chat message batches
	if !pa.chatLimiter.Allow() {
		// The player is sending message batches too fast.
		// Send a private error message back to only this player.
		pa.sendRateLimitError("You are sending messages too quickly.")
		return // Drop the action.
	}

	// Handle bulk message validation and processing
	if err := pa.processBulkMessages(&action); err != nil {
		pa.sendError(fmt.Sprintf("Invalid message format: %v", err))
		return
	}

	// Extract processed messages
	messages, ok := action.Payload["messages"].([]string)
	if !ok {
		pa.sendError("Failed to process message batch")
		return
	}

	// TODO: Implement lobby chat broadcasting for bulk messages
	log.Printf("[PlayerActor/%s] Lobby chat batch with %d messages: %v", pa.playerID, len(messages), messages)
}

// handleSyncLobbyState forwards sync lobby state requests to the game
func (pa *PlayerActor) handleSyncLobbyState(action core.Action) {
	// Forward the sync action to the lifecycle manager to handle via the game actor
	if pa.lifecycleManager == nil {
		pa.sendError("Lifecycle manager not available")
		return
	}

	// Set the proper game ID for the action
	action.GameID = pa.lobbyID // Use lobbyID as gameID for lobby actions
	action.PlayerID = pa.playerID

	resultChan, err := pa.lifecycleManager.SendActionToGame(pa.lobbyID, action)
	if err != nil {
		log.Printf("[PlayerActor/%s] Failed to send sync lobby state action: %v", pa.playerID, err)
		pa.sendError(fmt.Sprintf("Failed to sync lobby state: %v", err))
		return
	}

	// Wait for the result
	select {
	case result := <-resultChan:
		if result.Error != nil {
			log.Printf("[PlayerActor/%s] Sync lobby state action rejected: %v", pa.playerID, result.Error)
			pa.sendError(fmt.Sprintf("Failed to sync lobby state: %v", result.Error))
		} else if len(result.Events) > 0 {
			if err := pa.lifecycleManager.BroadcastEventsToGame(pa.lobbyID, result.Events); err != nil {
				log.Printf("[PlayerActor/%s] Failed to broadcast lobby sync events: %v", pa.playerID, err)
			}
		}
	case <-time.After(2 * time.Second):
		log.Printf("[PlayerActor/%s] Sync lobby state action timed out", pa.playerID)
		pa.sendError("Lobby sync timed out. The server is busy.")
	case <-pa.ctx.Done():
		return
	}
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
	case "REACT_TO_MESSAGE":
		actionType = core.ActionReactToMessage
	case "UPDATE_STATUS":
		actionType = core.ActionSetSlackStatus
	}

	// Apply rate limiting for chat messages (now applies to batches)
	if actionType == core.ActionSendMessage {
		if batchID, ok := action.Payload["batch_id"].(string); ok && batchID != "" {
			pa.stateMutex.Lock() // Use the actor's mutex to protect the cache
			if _, exists := pa.recentBatchIDs[batchID]; exists {
				pa.stateMutex.Unlock()
				log.Printf("[PlayerActor/%s] Duplicate message batch dropped: %s", pa.playerID, batchID)
				return // Drop the duplicate action
			}
			// Cache the new batch ID
			pa.recentBatchIDs[batchID] = time.Now()
			pa.stateMutex.Unlock()
		}
		if !pa.chatLimiter.Allow() {
			// The player is sending message batches too fast.
			// Send a private error message back to only this player.
			pa.sendRateLimitError("You are sending messages too quickly.")
			return // Drop the action.
		}

		// Handle bulk message validation and processing
		if err := pa.processBulkMessages(&action); err != nil {
			pa.sendError(fmt.Sprintf("Invalid message format: %v", err))
			return
		}
	}

	// List of valid game actions that should be forwarded to the SessionManager
	validGameActions := map[core.ActionType]bool{
		core.ActionSendMessage:         true,
		core.ActionReactToMessage:      true,
		core.ActionSubmitVote:          true,
		core.ActionSubmitSkipVote:      true, // Allow skip vote actions
		core.ActionSubmitNightAction:   true,
		core.ActionMineTokens:          true,
		core.ActionSubmitPulseCheck:    true,
		core.ActionUseAbility:          true,
		core.ActionAbandonGame:         true,
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

	// Send action and get response channel
	resultChan, err := pa.lifecycleManager.SendActionToGame(pa.gameID, action)
	if err != nil {
		log.Printf("[PlayerActor/%s] Failed to send action to game: %v", pa.playerID, err)
		pa.sendError(fmt.Sprintf("Failed to process action: %v", err))
		return
	}

	// Wait for the result with a timeout to prevent the PlayerActor from hanging indefinitely
	select {
	case result := <-resultChan:
		if result.Error != nil {
			// The action was rejected by the GameActor. Propagate the error to the client.
			log.Printf("[PlayerActor/%s] Action rejected by game: %v", pa.playerID, result.Error)
			pa.sendError(fmt.Sprintf("Action rejected: %v", result.Error))
		} else {
			// Action was successful. Broadcast the resulting events to all players in the game.
			if len(result.Events) > 0 {
				if err := pa.lifecycleManager.BroadcastEventsToGame(pa.gameID, result.Events); err != nil {
					log.Printf("[PlayerActor/%s] Failed to broadcast events: %v", pa.playerID, err)
				}
			}
		}
	case <-time.After(2 * time.Second): // 2-second timeout
		log.Printf("[PlayerActor/%s] Action timed out after 2 seconds", pa.playerID)
		pa.sendError("Action timed out. The server is busy.")
	case <-pa.ctx.Done():
		// PlayerActor is shutting down, do nothing.
		return
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

func (pa *PlayerActor) cleanupRecentBatchIDs() {
	pa.stateMutex.Lock()
	defer pa.stateMutex.Unlock()

	cutoff := time.Now().Add(-BATCH_ID_CACHE_TTL)
	for id, timestamp := range pa.recentBatchIDs {
		if timestamp.Before(cutoff) {
			delete(pa.recentBatchIDs, id)
		}
	}
	log.Printf("[PlayerActor/%s] Cleaned up old batch IDs. Cache size: %d", pa.playerID, len(pa.recentBatchIDs))
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
		Type:      core.EventClientError,
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"error_code":    "GENERAL_ERROR",
			"message":       message,
			"retry_allowed": false,
		},
	}

	pa.sendEvent(event)
}

// sendRateLimitError sends a rate limit exceeded error message to the client
func (pa *PlayerActor) sendRateLimitError(message string) {
	log.Printf("Rate limit exceeded for player %s: %s", pa.playerID, message)
	event := core.Event{
		Type:      core.EventRateLimitExceeded,
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": message,
		},
	}

	pa.sendEvent(event)
}

// processBulkMessages validates and processes bulk message payloads
// FIX: This function now correctly handles both single and bulk message formats
// and converts them into a consistent structure for downstream processing.
func (pa *PlayerActor) processBulkMessages(action *core.Action) error {
	const MAX_BATCH_SIZE = 10 // Increased limit for resilience
	var messagesToProcess []map[string]interface{}

	if messagesPayload, ok := action.Payload["messages"]; ok {
		// Handle bulk messages format
		messagesInterface, ok := messagesPayload.([]interface{})
		if !ok {
			return fmt.Errorf("messages field must be an array of objects")
		}

		if len(messagesInterface) > MAX_BATCH_SIZE {
			return fmt.Errorf("batch size exceeds limit of %d", MAX_BATCH_SIZE)
		}

		for i, msg := range messagesInterface {
			msgData, ok := msg.(map[string]interface{})
			if !ok {
				return fmt.Errorf("message at index %d is not a valid object", i)
			}
			messagesToProcess = append(messagesToProcess, msgData)
		}

	} else if message, ok := action.Payload["message"].(string); ok {
		// Handle single message (legacy support) and convert to bulk format
		messagesToProcess = append(messagesToProcess, map[string]interface{}{
			"message":           message,
			"client_message_id": action.Payload["client_message_id"],
		})
	} else {
		return fmt.Errorf("payload must contain 'messages' or 'message' field")
	}

	if len(messagesToProcess) == 0 {
		return fmt.Errorf("no valid messages to process")
	}

	// Validate and sanitize each message
	var finalMessages []map[string]interface{}
	for _, msgData := range messagesToProcess {
		content, _ := msgData["message"].(string)
		if err := validateChatMessage(content); err != nil {
			return err
		}
		msgData["message"] = sanitizeString(content, MAX_CHAT_MESSAGE_LENGTH)
		finalMessages = append(finalMessages, msgData)
	}

	// Replace original payload with the sanitized, structured bulk format.
	// This ensures downstream logic only has to handle one format.
	action.Payload = map[string]interface{}{
		"messages":   finalMessages,
		"channel_id": action.Payload["channel_id"],
	}
	log.Printf("[PlayerActor/%s] Processing bulk message with %d messages", pa.playerID, len(finalMessages))

	return nil
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

// Party action handlers

// handleCreateParty creates a new party with this player as leader
func (pa *PlayerActor) handleCreateParty(action core.Action) {
	if pa.partyManager == nil {
		pa.sendError("Party manager not available")
		return
	}

	partyID, err := pa.partyManager.CreateParty(pa)
	if err != nil {
		pa.sendError(fmt.Sprintf("Failed to create party: %v", err))
		return
	}

	// Send success response
	event := core.Event{
		Type:      "PARTY_CREATED",
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"party_id": partyID,
			"leader":   pa.playerID,
		},
	}
	pa.sendEvent(event)
}

// handleInviteToParty invites another player to join the party
func (pa *PlayerActor) handleInviteToParty(action core.Action) {
	if pa.partyManager == nil {
		pa.sendError("Party manager not available")
		return
	}

	inviteeID, ok := action.Payload["invitee_id"].(string)
	if !ok || inviteeID == "" {
		pa.sendError("Missing invitee_id in party invite")
		return
	}

	// Send invite (the party manager will validate if player is in a party and is leader)
	invite, err := pa.partyManager.InviteToPartyByInviter(pa.playerID, inviteeID)
	if err != nil {
		pa.sendError(fmt.Sprintf("Failed to send party invite: %v", err))
		return
	}

	// Send confirmation to inviter
	confirmEvent := core.Event{
		Type:      "PARTY_INVITE_SENT",
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"invite_id":  invite.ID,
			"invitee_id": inviteeID,
		},
	}
	pa.sendEvent(confirmEvent)
}

// handleJoinParty allows player to join a party via invite
func (pa *PlayerActor) handleJoinParty(action core.Action) {
	if pa.partyManager == nil {
		pa.sendError("Party manager not available")
		return
	}

	inviteID, ok := action.Payload["invite_id"].(string)
	if !ok || inviteID == "" {
		pa.sendError("Missing invite_id in party join")
		return
	}

	err := pa.partyManager.JoinParty(pa.playerID, inviteID, pa)
	if err != nil {
		pa.sendError(fmt.Sprintf("Failed to join party: %v", err))
		return
	}

	// Send success response
	event := core.Event{
		Type:      "PARTY_JOINED",
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"invite_id": inviteID,
		},
	}
	pa.sendEvent(event)
}

// handleLeaveParty removes player from their current party
func (pa *PlayerActor) handleLeaveParty(action core.Action) {
	if pa.partyManager == nil {
		pa.sendError("Party manager not available")
		return
	}

	err := pa.partyManager.LeaveParty(pa.playerID)
	if err != nil {
		pa.sendError(fmt.Sprintf("Failed to leave party: %v", err))
		return
	}

	// Send confirmation to leaving player
	leftEvent := core.Event{
		Type:      "PARTY_LEFT",
		PlayerID:  pa.playerID,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{},
	}
	pa.sendEvent(leftEvent)
}

// Input validation and sanitization functions

// validatePlayerName validates a player name for safety and format
func validatePlayerName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("player name cannot be empty")
	}
	if len(name) > MAX_PLAYER_NAME_LENGTH {
		return fmt.Errorf("player name too long (max %d characters)", MAX_PLAYER_NAME_LENGTH)
	}
	if !utf8.ValidString(name) {
		return fmt.Errorf("player name contains invalid UTF-8")
	}
	if !playerNamePattern.MatchString(name) {
		return fmt.Errorf("player name contains invalid characters")
	}
	return nil
}

// validateChatMessage validates a chat message for safety and format
func validateChatMessage(message string) error {
	if len(message) == 0 {
		return fmt.Errorf("message cannot be empty")
	}
	if len(message) > MAX_CHAT_MESSAGE_LENGTH {
		return fmt.Errorf("message too long (max %d characters)", MAX_CHAT_MESSAGE_LENGTH)
	}
	if !utf8.ValidString(message) {
		return fmt.Errorf("message contains invalid UTF-8")
	}
	// Allow more characters in chat messages than player names
	if !safeCharPattern.MatchString(message) {
		return fmt.Errorf("message contains potentially unsafe characters")
	}
	return nil
}

// validateStatusMessage validates a status message for safety and format
func validateStatusMessage(status string) error {
	if len(status) > MAX_STATUS_MESSAGE_LENGTH {
		return fmt.Errorf("status message too long (max %d characters)", MAX_STATUS_MESSAGE_LENGTH)
	}
	if !utf8.ValidString(status) {
		return fmt.Errorf("status message contains invalid UTF-8")
	}
	if !safeCharPattern.MatchString(status) {
		return fmt.Errorf("status message contains potentially unsafe characters")
	}
	return nil
}

// sanitizeString removes potentially dangerous characters and limits length
func sanitizeString(input string, maxLength int) string {
	// Remove any null bytes or control characters
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != '\t' && r != '\n' && r != '\r' {
			return -1 // Remove control characters except tab, newline, carriage return
		}
		return r
	}, input)

	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)

	// Limit length
	if len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}

	return sanitized
}