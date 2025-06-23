package actors

import (
	"context"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/lobby"
)

// MockLobbyManager implements LobbyManagerInterface for testing
type MockLobbyManager struct {
	createLobbyCalls  []CreateLobbyCall
	joinLobbyCalls    []JoinLobbyCall
	leaveLobbyCalls   []LeaveLobbyCall
	startGameCalls    []StartGameCall
	createLobbyResult CreateLobbyResult
	joinLobbyResult   error
	leaveLobbyResult  error
	startGameResult   error
}

type CreateLobbyCall struct {
	Player    interfaces.PlayerActorInterface
	LobbyName string
}

type CreateLobbyResult struct {
	LobbyID string
	Error   error
}

type JoinLobbyCall struct {
	LobbyID string
	Player  interfaces.PlayerActorInterface
}

type LeaveLobbyCall struct {
	LobbyID  string
	PlayerID string
}

type StartGameCall struct {
	LobbyID      string
	HostPlayerID string
}

func (m *MockLobbyManager) CreateLobby(player interfaces.PlayerActorInterface, lobbyName string) (string, error) {
	m.createLobbyCalls = append(m.createLobbyCalls, CreateLobbyCall{
		Player:    player,
		LobbyName: lobbyName,
	})
	return m.createLobbyResult.LobbyID, m.createLobbyResult.Error
}

func (m *MockLobbyManager) JoinLobbyWithActor(lobbyID string, player interfaces.PlayerActorInterface) error {
	m.joinLobbyCalls = append(m.joinLobbyCalls, JoinLobbyCall{
		LobbyID: lobbyID,
		Player:  player,
	})
	return m.joinLobbyResult
}

func (m *MockLobbyManager) LeaveLobby(lobbyID string, playerID string) error {
	m.leaveLobbyCalls = append(m.leaveLobbyCalls, LeaveLobbyCall{
		LobbyID:  lobbyID,
		PlayerID: playerID,
	})
	return m.leaveLobbyResult
}

func (m *MockLobbyManager) StartGame(lobbyID string, hostPlayerID string) error {
	m.startGameCalls = append(m.startGameCalls, StartGameCall{
		LobbyID:      lobbyID,
		HostPlayerID: hostPlayerID,
	})
	return m.startGameResult
}

// Unused methods for interface compliance
func (m *MockLobbyManager) JoinLobby(gameID, playerName, playerAvatar string) (string, string, error) {
	return "", "", nil
}
func (m *MockLobbyManager) GetLobbyList() []interface{}                                { return nil }
func (m *MockLobbyManager) GetLobby(lobbyID string) (interface{}, bool)                { return nil, false }
func (m *MockLobbyManager) ValidateSession(gameID, playerID, sessionToken string) bool { return false }
func (m *MockLobbyManager) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	return "", "", nil
}

// MockSessionManager implements SessionManagerInterface for testing
type MockSessionManager struct {
	joinGameCalls    []JoinGameCall
	leaveGameCalls   []LeaveGameCall
	sendActionCalls  []SendActionCall
	joinGameResult   error
	leaveGameResult  error
	sendActionResult error
}

type JoinGameCall struct {
	GameID string
	Player interfaces.PlayerActorInterface
}

type LeaveGameCall struct {
	GameID   string
	PlayerID string
}

type SendActionCall struct {
	GameID string
	Action core.Action
}

func (m *MockSessionManager) JoinGame(gameID string, player interfaces.PlayerActorInterface) error {
	m.joinGameCalls = append(m.joinGameCalls, JoinGameCall{
		GameID: gameID,
		Player: player,
	})
	return m.joinGameResult
}

func (m *MockSessionManager) LeaveGame(gameID string, playerID string) error {
	m.leaveGameCalls = append(m.leaveGameCalls, LeaveGameCall{
		GameID:   gameID,
		PlayerID: playerID,
	})
	return m.leaveGameResult
}

func (m *MockSessionManager) SendActionToGame(gameID string, action core.Action) error {
	m.sendActionCalls = append(m.sendActionCalls, SendActionCall{
		GameID: gameID,
		Action: action,
	})
	return m.sendActionResult
}

func (m *MockSessionManager) CreateGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	return nil
}

// Note: WebSocket testing is handled through integration tests
// Unit tests focus on the PlayerActor state machine and business logic

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

			mockLobby := &MockLobbyManager{}
			mockSession := &MockSessionManager{}

			actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil) // Use nil for WebSocket in unit tests
			actor.SetDependencies(mockLobby, mockSession)

			// Set initial state
			actor.stateMutex.Lock()
			actor.state = tt.initialState
			if tt.initialState == interfaces.StateInLobby {
				actor.lobbyID = "test-lobby"
			} else if tt.initialState == interfaces.StateInGame {
				actor.gameID = "test-game"
			}
			actor.stateMutex.Unlock()

			// Set up successful results for mocks
			mockLobby.createLobbyResult = CreateLobbyResult{LobbyID: "new-lobby", Error: nil}
			mockLobby.joinLobbyResult = nil
			mockLobby.leaveLobbyResult = nil
			mockLobby.startGameResult = nil
			mockSession.sendActionResult = nil
			mockSession.leaveGameResult = nil

			// Act - process the action directly
			actor.handleClientAction(tt.action)

			// Give time for async processing
			time.Sleep(10 * time.Millisecond)

			// Assert - Check if correct method was called on mocks
			switch tt.expectedMethod {
			case "CreateLobby":
				if len(mockLobby.createLobbyCalls) != 1 {
					t.Errorf("Expected CreateLobby to be called once, got %d calls", len(mockLobby.createLobbyCalls))
				}
			case "JoinLobby":
				if len(mockLobby.joinLobbyCalls) != 1 {
					t.Errorf("Expected JoinLobby to be called once, got %d calls", len(mockLobby.joinLobbyCalls))
				}
			case "StartGame":
				if len(mockLobby.startGameCalls) != 1 {
					t.Errorf("Expected StartGame to be called once, got %d calls", len(mockLobby.startGameCalls))
				}
			case "LeaveLobby":
				if len(mockLobby.leaveLobbyCalls) != 1 {
					t.Errorf("Expected LeaveLobby to be called once, got %d calls", len(mockLobby.leaveLobbyCalls))
				}
			case "SendAction":
				if len(mockSession.sendActionCalls) != 1 {
					t.Errorf("Expected SendAction to be called once, got %d calls", len(mockSession.sendActionCalls))
				}
			case "LeaveGame":
				if len(mockSession.leaveGameCalls) != 1 {
					t.Errorf("Expected LeaveGame to be called once, got %d calls", len(mockSession.leaveGameCalls))
				}
			}

			// For invalid actions, verify no manager methods were called
			if tt.shouldFail {
				// Verify that no manager methods were called for invalid actions
				if len(mockLobby.createLobbyCalls) > 0 || len(mockLobby.joinLobbyCalls) > 0 ||
					len(mockLobby.startGameCalls) > 0 || len(mockLobby.leaveLobbyCalls) > 0 ||
					len(mockSession.sendActionCalls) > 0 || len(mockSession.leaveGameCalls) > 0 {
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

	mockLobby := &MockLobbyManager{}
	mockSession := &MockSessionManager{}

	actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil)
	actor.SetDependencies(mockLobby, mockSession)

	// Test disconnect from InLobby state
	actor.TransitionToLobby("test-lobby")
	actor.handleDisconnect()

	if len(mockLobby.leaveLobbyCalls) != 1 {
		t.Errorf("Expected LeaveLobby to be called once on disconnect from lobby, got %d calls", len(mockLobby.leaveLobbyCalls))
	}
	if mockLobby.leaveLobbyCalls[0].LobbyID != "test-lobby" {
		t.Errorf("Expected lobby ID 'test-lobby', got '%s'", mockLobby.leaveLobbyCalls[0].LobbyID)
	}

	// Reset and test disconnect from InGame state
	mockSession.leaveGameCalls = nil
	actor.TransitionToGame("test-game")
	actor.handleDisconnect()

	if len(mockSession.leaveGameCalls) != 1 {
		t.Errorf("Expected LeaveGame to be called once on disconnect from game, got %d calls", len(mockSession.leaveGameCalls))
	}
	if mockSession.leaveGameCalls[0].GameID != "test-game" {
		t.Errorf("Expected game ID 'test-game', got '%s'", mockSession.leaveGameCalls[0].GameID)
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

// MockLobbyManager implements LobbyManagerInterface for testing
type MockLobbyManager struct {
	createLobbyCalls  []CreateLobbyCall
	joinLobbyCalls    []JoinLobbyCall
	leaveLobbyCalls   []LeaveLobbyCall
	startGameCalls    []StartGameCall
	createLobbyResult CreateLobbyResult
	joinLobbyResult   error
	leaveLobbyResult  error
	startGameResult   error
}

type CreateLobbyCall struct {
	Player    interfaces.PlayerActorInterface
	LobbyName string
}

type CreateLobbyResult struct {
	LobbyID string
	Error   error
}

type JoinLobbyCall struct {
	LobbyID string
	Player  interfaces.PlayerActorInterface
}

type LeaveLobbyCall struct {
	LobbyID  string
	PlayerID string
}

type StartGameCall struct {
	LobbyID      string
	HostPlayerID string
}

func (m *MockLobbyManager) CreateLobby(player interfaces.PlayerActorInterface, lobbyName string) (string, error) {
	m.createLobbyCalls = append(m.createLobbyCalls, CreateLobbyCall{
		Player:    player,
		LobbyName: lobbyName,
	})
	return m.createLobbyResult.LobbyID, m.createLobbyResult.Error
}

func (m *MockLobbyManager) JoinLobbyWithActor(lobbyID string, player interfaces.PlayerActorInterface) error {
	m.joinLobbyCalls = append(m.joinLobbyCalls, JoinLobbyCall{
		LobbyID: lobbyID,
		Player:  player,
	})
	return m.joinLobbyResult
}

func (m *MockLobbyManager) LeaveLobby(lobbyID string, playerID string) error {
	m.leaveLobbyCalls = append(m.leaveLobbyCalls, LeaveLobbyCall{
		LobbyID:  lobbyID,
		PlayerID: playerID,
	})
	return m.leaveLobbyResult
}

func (m *MockLobbyManager) StartGame(lobbyID string, hostPlayerID string) error {
	m.startGameCalls = append(m.startGameCalls, StartGameCall{
		LobbyID:      lobbyID,
		HostPlayerID: hostPlayerID,
	})
	return m.startGameResult
}

// Unused methods for interface compliance
func (m *MockLobbyManager) JoinLobby(gameID, playerName, playerAvatar string) (string, string, error) {
	return "", "", nil
}
func (m *MockLobbyManager) GetLobbyList() []interface{}                                { return nil }
func (m *MockLobbyManager) GetLobby(lobbyID string) (interface{}, bool)                { return nil, false }
func (m *MockLobbyManager) ValidateSession(gameID, playerID, sessionToken string) bool { return false }
func (m *MockLobbyManager) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	return "", "", nil
}

// MockSessionManager implements SessionManagerInterface for testing
type MockSessionManager struct {
	joinGameCalls    []JoinGameCall
	leaveGameCalls   []LeaveGameCall
	sendActionCalls  []SendActionCall
	joinGameResult   error
	leaveGameResult  error
	sendActionResult error
}

type JoinGameCall struct {
	GameID string
	Player interfaces.PlayerActorInterface
}

type LeaveGameCall struct {
	GameID   string
	PlayerID string
}

type SendActionCall struct {
	GameID string
	Action core.Action
}

func (m *MockSessionManager) JoinGame(gameID string, player interfaces.PlayerActorInterface) error {
	m.joinGameCalls = append(m.joinGameCalls, JoinGameCall{
		GameID: gameID,
		Player: player,
	})
	return m.joinGameResult
}

func (m *MockSessionManager) LeaveGame(gameID string, playerID string) error {
	m.leaveGameCalls = append(m.leaveGameCalls, LeaveGameCall{
		GameID:   gameID,
		PlayerID: playerID,
	})
	return m.leaveGameResult
}

func (m *MockSessionManager) SendActionToGame(gameID string, action core.Action) error {
	m.sendActionCalls = append(m.sendActionCalls, SendActionCall{
		GameID: gameID,
		Action: action,
	})
	return m.sendActionResult
}

func (m *MockSessionManager) CreateGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	return nil
}

// Note: WebSocket testing is handled through integration tests
// Unit tests focus on the PlayerActor state machine and business logic

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

			mockLobby := &MockLobbyManager{}
			mockSession := &MockSessionManager{}

			actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil) // Use nil for WebSocket in unit tests
			actor.SetDependencies(mockLobby, mockSession)

			// Set initial state
			actor.stateMutex.Lock()
			actor.state = tt.initialState
			if tt.initialState == interfaces.StateInLobby {
				actor.lobbyID = "test-lobby"
			} else if tt.initialState == interfaces.StateInGame {
				actor.gameID = "test-game"
			}
			actor.stateMutex.Unlock()

			// Set up successful results for mocks
			mockLobby.createLobbyResult = CreateLobbyResult{LobbyID: "new-lobby", Error: nil}
			mockLobby.joinLobbyResult = nil
			mockLobby.leaveLobbyResult = nil
			mockLobby.startGameResult = nil
			mockSession.sendActionResult = nil
			mockSession.leaveGameResult = nil

			// Act - process the action directly
			actor.handleClientAction(tt.action)

			// Give time for async processing
			time.Sleep(10 * time.Millisecond)

			// Assert - Check if correct method was called on mocks
			switch tt.expectedMethod {
			case "CreateLobby":
				if len(mockLobby.createLobbyCalls) != 1 {
					t.Errorf("Expected CreateLobby to be called once, got %d calls", len(mockLobby.createLobbyCalls))
				}
			case "JoinLobby":
				if len(mockLobby.joinLobbyCalls) != 1 {
					t.Errorf("Expected JoinLobby to be called once, got %d calls", len(mockLobby.joinLobbyCalls))
				}
			case "StartGame":
				if len(mockLobby.startGameCalls) != 1 {
					t.Errorf("Expected StartGame to be called once, got %d calls", len(mockLobby.startGameCalls))
				}
			case "LeaveLobby":
				if len(mockLobby.leaveLobbyCalls) != 1 {
					t.Errorf("Expected LeaveLobby to be called once, got %d calls", len(mockLobby.leaveLobbyCalls))
				}
			case "SendAction":
				if len(mockSession.sendActionCalls) != 1 {
					t.Errorf("Expected SendAction to be called once, got %d calls", len(mockSession.sendActionCalls))
				}
			case "LeaveGame":
				if len(mockSession.leaveGameCalls) != 1 {
					t.Errorf("Expected LeaveGame to be called once, got %d calls", len(mockSession.leaveGameCalls))
				}
			}

			// For invalid actions, verify no manager methods were called
			if tt.shouldFail {
				// Verify that no manager methods were called for invalid actions
				if len(mockLobby.createLobbyCalls) > 0 || len(mockLobby.joinLobbyCalls) > 0 ||
					len(mockLobby.startGameCalls) > 0 || len(mockLobby.leaveLobbyCalls) > 0 ||
					len(mockSession.sendActionCalls) > 0 || len(mockSession.leaveGameCalls) > 0 {
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

	mockLobby := &MockLobbyManager{}
	mockSession := &MockSessionManager{}

	actor := NewPlayerActor(ctx, "test-player", "TestPlayer", "test-token", nil)
	actor.SetDependencies(mockLobby, mockSession)

	// Test disconnect from InLobby state
	actor.TransitionToLobby("test-lobby")
	actor.handleDisconnect()

	if len(mockLobby.leaveLobbyCalls) != 1 {
		t.Errorf("Expected LeaveLobby to be called once on disconnect from lobby, got %d calls", len(mockLobby.leaveLobbyCalls))
	}
	if mockLobby.leaveLobbyCalls[0].LobbyID != "test-lobby" {
		t.Errorf("Expected lobby ID 'test-lobby', got '%s'", mockLobby.leaveLobbyCalls[0].LobbyID)
	}

	// Reset and test disconnect from InGame state
	mockSession.leaveGameCalls = nil
	actor.TransitionToGame("test-game")
	actor.handleDisconnect()

	if len(mockSession.leaveGameCalls) != 1 {
		t.Errorf("Expected LeaveGame to be called once on disconnect from game, got %d calls", len(mockSession.leaveGameCalls))
	}
	if mockSession.leaveGameCalls[0].GameID != "test-game" {
		t.Errorf("Expected game ID 'test-game', got '%s'", mockSession.leaveGameCalls[0].GameID)
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
