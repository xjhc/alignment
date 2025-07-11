package comms

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/mocks"
)

// MockTokenValidator implements interfaces.TokenValidator for testing
type MockTokenValidator struct {
	sessions map[string]bool // sessionToken -> valid
	players  map[string]playerInfo // playerID -> playerInfo
}

type playerInfo struct {
	name   string
	avatar string
	err    error
}

func newMockTokenValidator() *MockTokenValidator {
	return &MockTokenValidator{
		sessions: make(map[string]bool),
		players:  make(map[string]playerInfo),
	}
}

func (m *MockTokenValidator) ValidateSession(gameID, playerID, sessionToken string) bool {
	if valid, exists := m.sessions[sessionToken]; exists {
		return valid
	}
	return false
}

func (m *MockTokenValidator) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	if info, exists := m.players[playerID]; exists {
		return info.name, info.avatar, info.err
	}
	return "", "", nil
}

// TestWebSocketManager_SessionExpiredFlow tests the session expiration flow
func TestWebSocketManager_SessionExpiredFlow(t *testing.T) {
	// Create test components
	ctx := context.Background()
	eventBus := events.NewEventBus()
	mockTokenValidator := newMockTokenValidator()
	mockLifecycleManager := &mocks.MockGameLifecycleManager{}

	// Setup mock data - invalid session
	mockTokenValidator.sessions["expired-token"] = false
	mockTokenValidator.players["test-player"] = playerInfo{
		name:   "TestPlayer",
		avatar: "avatar1",
		err:    nil,
	}

	// Create WebSocket manager
	wsm := NewWebSocketManager(ctx, mockTokenValidator)
	wsm.SetDependencies(mockLifecycleManager, eventBus)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(wsm.HandleWebSocket))
	defer server.Close()

	// Create WebSocket URL with expired session token
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.RawQuery = "gameId=test-game&playerId=test-player&sessionToken=expired-token"

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Set read deadline to prevent hanging
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read the SESSION_EXPIRED event
	var event core.Event
	err = conn.ReadJSON(&event)
	if err != nil {
		t.Fatalf("Failed to read SESSION_EXPIRED event: %v", err)
	}

	// Verify the event
	if event.Type != "SESSION_EXPIRED" {
		t.Errorf("Expected event type SESSION_EXPIRED, got %s", event.Type)
	}
	if event.GameID != "test-game" {
		t.Errorf("Expected GameID test-game, got %s", event.GameID)
	}
	if event.PlayerID != "test-player" {
		t.Errorf("Expected PlayerID test-player, got %s", event.PlayerID)
	}
	if !strings.Contains(event.ID, "session_expired_") {
		t.Errorf("Expected event ID to contain 'session_expired_', got %s", event.ID)
	}
	
	// Verify payload
	if event.Payload == nil {
		t.Error("Expected payload to be present")
		return
	}
	if event.Payload["reason"] != "session_invalid" {
		t.Errorf("Expected reason 'session_invalid', got %v", event.Payload["reason"])
	}
	if event.Payload["message"] != "Your session has expired. Please log in again." {
		t.Errorf("Expected message 'Your session has expired. Please log in again.', got %v", event.Payload["message"])
	}

	// Verify connection is closed after the event
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Error("Expected connection to be closed after SESSION_EXPIRED event")
	}
	if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
		t.Errorf("Expected WebSocket close error, got %v", err)
	}
}

// TestWebSocketManager_ValidSessionFlow tests the normal flow with valid session
func TestWebSocketManager_ValidSessionFlow(t *testing.T) {
	// Create test components
	ctx := context.Background()
	eventBus := events.NewEventBus()
	mockTokenValidator := newMockTokenValidator()
	mockLifecycleManager := &mocks.MockGameLifecycleManager{}

	// Setup mock data - valid session
	mockTokenValidator.sessions["valid-token"] = true
	mockTokenValidator.players["test-player"] = playerInfo{
		name:   "TestPlayer",
		avatar: "avatar1",
		err:    nil,
	}

	// Create WebSocket manager
	wsm := NewWebSocketManager(ctx, mockTokenValidator)
	wsm.SetDependencies(mockLifecycleManager, eventBus)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(wsm.HandleWebSocket))
	defer server.Close()

	// Create WebSocket URL with valid session token
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.RawQuery = "gameId=test-game&playerId=test-player&sessionToken=valid-token"

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Set read deadline to prevent hanging
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// Wait a bit for connection to establish
	time.Sleep(100 * time.Millisecond)

	// The connection should remain open (no SESSION_EXPIRED event)
	// Try to read with a short timeout - should timeout, not get SESSION_EXPIRED
	_, _, err = conn.ReadMessage()
	if err != nil {
		// Should be a timeout error, not a close error
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "deadline") {
			t.Errorf("Expected timeout error, got %v", err)
		}
	}
}

// TestWebSocketManager_MissingParameters tests the flow with missing required parameters
func TestWebSocketManager_MissingParameters(t *testing.T) {
	// Create test components
	ctx := context.Background()
	eventBus := events.NewEventBus()
	mockTokenValidator := newMockTokenValidator()
	mockLifecycleManager := &mocks.MockGameLifecycleManager{}

	// Create WebSocket manager
	wsm := NewWebSocketManager(ctx, mockTokenValidator)
	wsm.SetDependencies(mockLifecycleManager, eventBus)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(wsm.HandleWebSocket))
	defer server.Close()

	// Create WebSocket URL with missing parameters
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	u.RawQuery = "gameId=test-game&playerId=test-player" // Missing sessionToken

	// Attempt to connect to WebSocket
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	
	// Should fail with HTTP 400 Bad Request
	if err == nil {
		t.Error("Expected connection to fail with missing parameters")
		if conn != nil {
			conn.Close()
		}
		return
	}
	
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 Bad Request, got %d", resp.StatusCode)
	}
	
	if conn != nil {
		conn.Close()
	}
}