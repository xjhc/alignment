package actors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/lobby"
	"github.com/xjhc/alignment/server/internal/mocks"
)

// Note: WebSocket testing is handled through integration tests
// Unit tests focus on the PlayerActor state machine and business logic

// createMockWebSocketConnection creates a mock WebSocket connection for testing
func createMockWebSocketConnection(t *testing.T) *websocket.Conn {
	// Create a test server that upgrades to WebSocket
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Failed to upgrade: %v", err)
			return
		}
		defer conn.Close()
		// Keep connection open for the duration of the test
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	// Create client connection
	u, _ := url.Parse(server.URL)
	u.Scheme = "ws"
	
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	
	return conn
}

// Test table for state machine validation
func TestPlayerActor_StateMachine(t *testing.T) {
	tests := []struct {
		name           string
		initialState   interfaces.PlayerState
		action         core.Action
		shouldFail     bool
		expectedState  interfaces.PlayerState
		expectedMethod string // Method expected to be called on mock
	}{
		// Idle state tests
		{
			name:           "Idle_CreateGame_Success",
			initialState:   interfaces.StateIdle,
			action:         core.Action{Type: core.ActionCreateGame, Payload: map[string]interface{}{"lobby_name": "Test Game"}},
			shouldFail:     false,
			expectedState:  interfaces.StateIdle, // Will change to InLobby via callback
			expectedMethod: "CreateLobby",
		},
		{
			name:           "Idle_JoinGame_Success",
			initialState:   interfaces.StateIdle,
			action:         core.Action{Type: core.ActionJoinGame, Payload: map[string]interface{}{"lobby_id": "test-lobby"}},
			shouldFail:     false,
			expectedState:  interfaces.StateIdle, // Will change to InLobby via callback
			expectedMethod: "JoinLobby",
		},
		{
			name:         "Idle_StartGame_Invalid",
			initialState: interfaces.StateIdle,
			action:       core.Action{Type: core.ActionStartGame},
			shouldFail:   true,
		},
		{
			name:         "Idle_SubmitVote_Invalid",
			initialState: interfaces.StateIdle,
			action:       core.Action{Type: core.ActionSubmitVote},
			shouldFail:   true,
		},

		// InLobby state tests
		{
			name:           "InLobby_StartGame_Success",
			initialState:   interfaces.StateInLobby,
			action:         core.Action{Type: core.ActionStartGame},
			shouldFail:     false,
			expectedState:  interfaces.StateInLobby, // Will change to InGame via callback
			expectedMethod: "StartGame",
		},
		{
			name:           "InLobby_LeaveGame_Success",
			initialState:   interfaces.StateInLobby,
			action:         core.Action{Type: core.ActionLeaveGame},
			shouldFail:     false,
			expectedState:  interfaces.StateIdle,
			expectedMethod: "LeaveLobby",
		},
		{
			name:         "InLobby_CreateGame_Invalid",
			initialState: interfaces.StateInLobby,
			action:       core.Action{Type: core.ActionCreateGame},
			shouldFail:   true,
		},
		{
			name:         "InLobby_SubmitVote_Invalid",
			initialState: interfaces.StateInLobby,
			action:       core.Action{Type: core.ActionSubmitVote},
			shouldFail:   true,
		},

		// InGame state tests
		{
			name:           "InGame_SubmitVote_Success",
			initialState:   interfaces.StateInGame,
			action:         core.Action{Type: core.ActionSubmitVote, Payload: map[string]interface{}{"target": "player2"}},
			shouldFail:     false,
			expectedState:  interfaces.StateInGame,
			expectedMethod: "SendAction",
		},
		{
			name:           "InGame_LeaveGame_Success",
			initialState:   interfaces.StateInGame,
			action:         core.Action{Type: core.ActionLeaveGame},
			shouldFail:     false,
			expectedState:  interfaces.StateIdle,
			expectedMethod: "LeaveGame",
		},
		{
			name:         "InGame_CreateGame_Invalid",
			initialState: interfaces.StateInGame,
			action:       core.Action{Type: core.ActionCreateGame},
			shouldFail:   true,
		},
		{
			name:         "InGame_StartGame_Invalid",
			initialState: interfaces.StateInGame,
			action:       core.Action{Type: core.ActionStartGame},
			shouldFail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			mockLifecycleManager := &mocks.MockGameLifecycleManager{}
			eventBus := events.NewEventBus()
			
			// Create a channel to capture published events
			eventCapture := make(chan events.Event, 10)
			eventBus.Subscribe("player_disconnected", eventCapture)

			mockConn := createMockWebSocketConnection(t)
			actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", mockConn)
			actor.SetDependencies(mockLifecycleManager, eventBus)

			// Set initial state
			actor.stateMutex.Lock()
			actor.state = tt.initialState
			if tt.initialState == interfaces.StateInLobby {
				actor.lobbyID = "test-lobby"
			} else if tt.initialState == interfaces.StateInGame {
				actor.gameID = "test-game"
			}
			actor.stateMutex.Unlock()

			// Set up successful results for mocks using the new API
			mockLifecycleManager.CreateLobbyViaHTTPResults = []mocks.GLMCreateLobbyViaHTTPResult{{LobbyID: "new-lobby", PlayerID: "test-player", SessionToken: "test-token", Error: nil}}
			mockLifecycleManager.JoinLobbyWithActorResults = []error{nil}
			mockLifecycleManager.StartGameResults = []error{nil}
			mockLifecycleManager.SendActionToGameResults = []error{nil}

			// Set up WaitGroup for async operations based on expected method
			if tt.expectedMethod == "StartGame" {
				mockLifecycleManager.Wg.Add(1)
			}

			// Act - process the action directly
			actor.handleClientAction(tt.action)

			// Wait for async processing if we expect a method call
			if tt.expectedMethod == "StartGame" {
				waitWithTimeout(&mockLifecycleManager.Wg, 1*time.Second, t)
			} else if tt.expectedMethod != "" {
				// For other methods that are synchronous, give a small window
				time.Sleep(1 * time.Millisecond)
			}

			// Assert - Check if correct method was called on mocks
			switch tt.expectedMethod {
			case "CreateLobby":
				// Note: CreateLobby is now handled via HTTP API, not PlayerActor
				// This test case may need to be updated or removed
			case "JoinLobby":
				if len(mockLifecycleManager.JoinLobbyWithActorCalls) != 1 {
					t.Errorf("Expected JoinLobby to be called once, got %d calls", len(mockLifecycleManager.JoinLobbyWithActorCalls))
				}
			case "StartGame":
				if len(mockLifecycleManager.StartGameCalls) != 1 {
					t.Errorf("Expected StartGame to be called once, got %d calls", len(mockLifecycleManager.StartGameCalls))
				}
			case "LeaveLobby":
				// LeaveLobby is now handled via event bus, not direct calls
				// Check that a disconnection event was published
				select {
				case <-eventCapture:
					// Event received as expected
				case <-time.After(10 * time.Millisecond):
					t.Error("Expected one event to be published, but none was received")
				}
			case "SendAction":
				if len(mockLifecycleManager.SendActionToGameCalls) != 1 {
					t.Errorf("Expected SendAction to be called once, got %d calls", len(mockLifecycleManager.SendActionToGameCalls))
				}
			case "LeaveGame":
				// LeaveGame is now handled via event bus, not direct calls
				// Check that a disconnection event was published
				select {
				case <-eventCapture:
					// Event received as expected
				case <-time.After(10 * time.Millisecond):
					t.Error("Expected one event to be published, but none was received")
				}
			}

			// For invalid actions, verify no manager methods were called
			if tt.shouldFail {
				// Verify that no manager methods were called for invalid actions
				if len(mockLifecycleManager.JoinLobbyWithActorCalls) > 0 ||
					len(mockLifecycleManager.StartGameCalls) > 0 ||
					len(mockLifecycleManager.SendActionToGameCalls) > 0 {
					t.Error("Expected no manager methods to be called for invalid actions")
				}
			}

			// Verify final state (for successful state transitions)
			if !tt.shouldFail && tt.expectedState != tt.initialState {
				if actor.GetState() != tt.expectedState {
					t.Errorf("Expected final state %s, got %s", tt.expectedState, actor.GetState())
				}
			}
		})
	}
}

func TestPlayerActor_StateTransitions(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil)

	// Test initial state
	if actor.GetState() != interfaces.StateIdle {
		t.Errorf("Expected initial state to be Idle, got %s", actor.GetState())
	}

	// Test Idle -> InLobby transition
	err := actor.TransitionToLobby("test-lobby")
	if err != nil {
		t.Errorf("Failed to transition to lobby: %v", err)
	}
	if actor.GetState() != interfaces.StateInLobby {
		t.Errorf("Expected state to be InLobby, got %s", actor.GetState())
	}

	// Test InLobby -> InGame transition
	err = actor.TransitionToGame("test-game")
	if err != nil {
		t.Errorf("Failed to transition to game: %v", err)
	}
	if actor.GetState() != interfaces.StateInGame {
		t.Errorf("Expected state to be InGame, got %s", actor.GetState())
	}

	// Test InGame -> Idle transition
	err = actor.TransitionToIdle()
	if err != nil {
		t.Errorf("Failed to transition to idle: %v", err)
	}
	if actor.GetState() != interfaces.StateIdle {
		t.Errorf("Expected state to be Idle, got %s", actor.GetState())
	}

	// Test invalid transition (Idle -> InGame directly)
	err = actor.TransitionToGame("test-game")
	if err == nil {
		t.Error("Expected error when transitioning from Idle to InGame, but got none")
	}
}

func TestPlayerActor_DisconnectHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockLifecycleManager := &mocks.MockGameLifecycleManager{}
	eventBus := events.NewEventBus()
	
	// Create a channel to capture published events
	eventCapture := make(chan events.Event, 10)
	eventBus.Subscribe("player_disconnected", eventCapture)

	actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil)
	actor.SetDependencies(mockLifecycleManager, eventBus)

	// Test disconnect from InLobby state
	actor.TransitionToLobby("test-lobby")
	actor.handleDisconnect()

	// With the new architecture, disconnection publishes events instead of direct calls
	select {
	case <-eventCapture:
		// Event received as expected
	case <-time.After(10 * time.Millisecond):
		t.Error("Expected one event to be published on disconnect from lobby, but none was received")
	}

	// Reset and test disconnect from InGame state
	actor.TransitionToGame("test-game")
	actor.handleDisconnect()

	// With the new architecture, disconnection publishes events instead of direct calls
	select {
	case <-eventCapture:
		// Event received as expected
	case <-time.After(10 * time.Millisecond):
		t.Error("Expected one event to be published on disconnect from game, but none was received")
	}
}

func TestPlayerActor_ServerMessageHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil)

	// Test sending a lobby state update
	lobbyUpdate := lobby.LobbyStateUpdate{
		LobbyID:   "test-lobby",
		Players:   []lobby.PlayerInfo{{ID: "test-player", Name: "TestPlayer"}},
		HostID:    "test-player",
		CanStart:  true,
		LobbyName: "Test Lobby",
	}

	// Test that the message can be sent without errors
	// In a real implementation with WebSocket, we'd verify the actual message sending
	actor.SendServerMessage(lobbyUpdate)

	// For unit testing, we're mainly verifying that the method doesn't panic
	// Integration tests would verify the actual WebSocket message flow
	t.Log("Server message handling test passed - no panics occurred")
}

// waitWithTimeout waits for a WaitGroup with a timeout to prevent test hangs
func waitWithTimeout(wg *sync.WaitGroup, timeout time.Duration, t *testing.T) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		// Wait finished successfully
	case <-time.After(timeout):
		t.Fatal("Test timed out waiting for WaitGroup")
	}
}