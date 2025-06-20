package comms

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/gorilla/websocket"
)

// WebSocketManager handles WebSocket connections and message routing
type WebSocketManager struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	// Message handler
	actionHandler ActionHandler
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
func NewWebSocketManager(actionHandler ActionHandler) *WebSocketManager {
	return &WebSocketManager{
		clients:       make(map[string]*Client),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan []byte),
		actionHandler: actionHandler,
	}
}

// Start begins the WebSocket manager's processing loop
func (wsm *WebSocketManager) Start() {
	go wsm.run()
}

// HandleWebSocket handles WebSocket connection upgrades
func (wsm *WebSocketManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Extract client ID from query params or generate one
	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		clientID = generateClientID()
	}

	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  wsm,
	}

	wsm.register <- client

	// Start client goroutines
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

// run handles client registration/unregistration and broadcasting
func (wsm *WebSocketManager) run() {
	for {
		select {
		case client := <-wsm.register:
			wsm.clients[client.ID] = client
			log.Printf("Client %s connected", client.ID)

		case client := <-wsm.unregister:
			if _, ok := wsm.clients[client.ID]; ok {
				delete(wsm.clients, client.ID)
				close(client.Send)
				log.Printf("Client %s disconnected", client.ID)
			}

		case message := <-wsm.broadcast:
			for _, client := range wsm.clients {
				select {
				case client.Send <- message:
				default:
					wsm.unregister <- client
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

		var message Message
		if err := json.Unmarshal(messageData, &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// Convert message to action and handle
		action := core.Action{
			Type:      core.ActionType(message.Type),
			PlayerID:  c.ID,
			GameID:    message.GameID,
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
