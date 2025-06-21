package comms

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/lobby"
)

// WebSocketManager handles WebSocket connections and message routing
type WebSocketManager struct {
	clients        map[string]*Client
	register       chan *Client
	unregister     chan *Client
	actionHandler  ActionHandler
	tokenValidator TokenValidator
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	GameID string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *WebSocketManager
}

// ActionHandler processes game actions from clients
type ActionHandler interface {
	HandleAction(action core.Action) error
}

// TokenValidator validates sessions and provides lobby access
type TokenValidator interface {
	ValidateSession(gameId, playerId, sessionToken string) bool
	GetLobbyActor(gameID string) (*lobby.LobbyActor, bool)
}

// Message represents a WebSocket message
type Message struct {
	Type    string                 `json:"type"`
	GameID  string                 `json:"game_id,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking
		return true
	},
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager(actionHandler ActionHandler, tokenValidator TokenValidator) *WebSocketManager {
	return &WebSocketManager{
		clients:        make(map[string]*Client),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		actionHandler:  actionHandler,
		tokenValidator: tokenValidator,
	}
}

// SetTokenValidator sets the token validator (for dependency injection)
func (wsm *WebSocketManager) SetTokenValidator(tokenValidator TokenValidator) {
	wsm.tokenValidator = tokenValidator
}

// Start begins the WebSocket manager's processing loop
func (wsm *WebSocketManager) Start() {
	go wsm.run()
}

// HandleWebSocket handles WebSocket connection upgrades
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

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		ID:     playerID,
		GameID: gameID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    wsm,
	}

	wsm.register <- client

	go client.writePump()
	go client.readPump()
}

// BroadcastToGame sends a message to all clients in a specific game
func (wsm *WebSocketManager) BroadcastToGame(gameID string, event core.Event) error {
	message := Message{
		Type:    string(event.Type),
		GameID:  gameID,
		Payload: event.Payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Send to all clients in the game
	for _, client := range wsm.clients {
		if client.GameID == gameID {
			select {
			case client.Send <- data:
			default:
				// Client buffer full, disconnect
				wsm.unregister <- client
			}
		}
	}

	return nil
}

// SendToPlayer sends a message to a specific player
func (wsm *WebSocketManager) SendToPlayer(gameID, playerID string, event core.Event) error {
	message := Message{
		Type:    string(event.Type),
		GameID:  gameID,
		Payload: event.Payload,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// Find the client for this player
	for _, client := range wsm.clients {
		if client.ID == playerID && client.GameID == gameID {
			select {
			case client.Send <- data:
				return nil
			default:
				// Client buffer full, disconnect
				wsm.unregister <- client
				return ErrClientDisconnected
			}
		}
	}

	return ErrPlayerNotFound
}

// run is the main processing loop for the WebSocketManager
func (wsm *WebSocketManager) run() {
	for {
		select {
		case client := <-wsm.register:
			wsm.clients[client.ID] = client
			log.Printf("Client %s connected to game %s", client.ID, client.GameID)

			// The player is already "in" the lobby on the backend.
			// Just tell the LobbyActor to send the current state to the new connection.
			if lobbyActor, exists := wsm.tokenValidator.GetLobbyActor(client.GameID); exists {
				lobbyActor.SendCurrentStateToPlayer(client.ID)
			} else {
				// TODO: Handle reconnecting to a game in progress
				log.Printf("No lobby actor found for game %s during connection", client.GameID)
			}

		case client := <-wsm.unregister:
			if _, ok := wsm.clients[client.ID]; ok {
				delete(wsm.clients, client.ID)
				close(client.Send)
				log.Printf("Client %s disconnected from game %s", client.ID, client.GameID)

				if wsm.actionHandler != nil && client.GameID != "" {
					leaveAction := core.Action{Type: core.ActionLeaveGame, PlayerID: client.ID, GameID: client.GameID, Timestamp: time.Now()}
					if err := wsm.actionHandler.HandleAction(leaveAction); err != nil {
						log.Printf("Failed to handle leave action for client %s: %v", client.ID, err)
					}
				}
			}
		}
	}
}

// readPump handles incoming messages from the client
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageData, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle ping messages separately
		if string(messageData) == `{"type":"ping"}` {
			// This is a heartbeat, just reset the deadline
			c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			continue
		}

		var message Message
		if err := json.Unmarshal(messageData, &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Extract game_id from payload if present
		var gameID string
		if message.Payload != nil {
			if id, ok := message.Payload["game_id"].(string); ok {
				gameID = id
			}
		}

		// If game_id wasn't in the payload, use the client's associated GameID
		// This handles subsequent messages after joining
		if gameID == "" {
			gameID = c.GameID
		}

		// Convert message to action and handle
		action := core.Action{
			Type:      core.ActionType(message.Type),
			PlayerID:  c.ID,
			GameID:    gameID,
			Timestamp: time.Now(),
			Payload:   message.Payload,
		}

		// Update client's game ID if joining a game
		if action.Type == core.ActionJoinGame {
			c.GameID = action.GameID
		}

		// Handle the action
		if err := c.Hub.actionHandler.HandleAction(action); err != nil {
			log.Printf("Failed to handle action: %v", err)
		}
	}
}

// writePump handles outgoing messages to the client
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateClientID generates a simple client ID
func generateClientID() string {
	return fmt.Sprintf("client_%d", time.Now().UnixNano())
}

// Custom errors
var (
	ErrClientDisconnected = fmt.Errorf("client disconnected")
	ErrPlayerNotFound     = fmt.Errorf("player not found")
)
