package game

import (
	"context"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockGameActor for testing SessionManager
type MockGameActor struct {
	GameID string
	PostActionResult chan interfaces.ProcessActionResult
}

func (m *MockGameActor) GetGameID() string {
	return m.GameID
}

func (m *MockGameActor) PostAction(action core.Action) chan interfaces.ProcessActionResult {
	return m.PostActionResult
}

func (m *MockGameActor) GetGameState() *core.GameState {
	return &core.GameState{ID: m.GameID}
}

func (m *MockGameActor) CreatePlayerStateUpdateEvent(playerID string) core.Event {
	return core.Event{Type: "GAME_STATE_UPDATE", PlayerID: playerID, Payload: map[string]interface{}{
		"game_state": "snapshot_for_" + playerID,
	}}
}

func (m *MockGameActor) Stop() {}


// TestSessionManager_CreateGameFromLobby tests the atomic game creation from a lobby
func TestSessionManager_CreateGameFromLobby(t *testing.T) {
	ctx := context.Background()

	// Mocks
	mockDatastore := &MockDataStore{}
	mockBroadcaster := &MockBroadcaster{}
	mockSupervisor := &MockSupervisor{}

	sessionManager := NewSessionManager(ctx, mockDatastore, mockBroadcaster, mockSupervisor)

	lobbyID := "test-lobby"
	playerActors := map[string]interfaces.PlayerActorInterface{
		"player1": &MockPlayerActor{PlayerID: "player1"},
		"player2": &MockPlayerActor{PlayerID: "player2"},
	}

	// Set up supervisor mock to return a game actor
	mockGameActor := &MockGameActor{
		GameID: lobbyID,
		PostActionResult: make(chan interfaces.ProcessActionResult, 1),
	}

	// Simulate GameActor returning state updates after initialization
	go func() {
		updateEvents := []core.Event{
			{Type: "GAME_STATE_UPDATE", PlayerID: "player1", Payload: map[string]interface{}{"game_state": "snapshot1"}},
			{Type: "GAME_STATE_UPDATE", PlayerID: "player2", Payload: map[string]interface{}{"game_state": "snapshot2"}},
		}
		mockGameActor.PostActionResult <- interfaces.ProcessActionResult{Events: updateEvents}
	}()

	mockSupervisor.CreateGameResult = mockGameActor

	err := sessionManager.CreateGameFromLobby(lobbyID, playerActors)

	if err != nil {
		t.Fatalf("CreateGameFromLobby failed: %v", err)
	}

	// Verify that a game actor was created
	if mockSupervisor.CreateGameCalls != 1 {
		t.Errorf("Expected CreateGameWithPlayers to be called once, got %d", mockSupervisor.CreateGameCalls)
	}

	// Verify the game session was created
	if len(sessionManager.gameSessions[lobbyID]) != 2 {
		t.Errorf("Expected 2 players in the game session, got %d", len(sessionManager.gameSessions[lobbyID]))
	}

	// Verify that each player actor received a transition message
	for playerID, playerActor := range playerActors {
		mockActor, ok := playerActor.(*MockPlayerActor)
		if !ok {
			t.Fatalf("Failed to cast player actor to mock")
		}

		select {
		case msg := <-mockActor.Messages:
			// The first message is the state update
			updateMsg, ok := msg.(core.Event)
			if !ok {
				t.Fatalf("Expected first message to be core.Event, got %T", msg)
			}
			if updateMsg.Type != "GAME_STATE_UPDATE" {
				t.Errorf("Expected GAME_STATE_UPDATE, got %s", updateMsg.Type)
			}

			// The second message is the GAME_STARTED event
			select {
			case msg = <-mockActor.Messages:
				startMsg, ok := msg.(core.Event)
				if !ok {
					t.Fatalf("Expected second message to be core.Event, got %T", msg)
				}
				if startMsg.Type != core.EventGameStarted {
					t.Errorf("Expected GAME_STARTED, got %s", startMsg.Type)
				}
			case <-time.After(100*time.Millisecond):
				t.Errorf("Player %s did not receive GAME_STARTED event", playerID)
			}

		case <-time.After(100 * time.Millisecond):
			t.Errorf("Player %s did not receive any messages", playerID)
		}
	}
}

// MockPlayerActor for testing SessionManager
type MockPlayerActor struct {
	PlayerID      string
	PlayerName    string
	SessionToken  string
	CurrentState  interfaces.PlayerState
	Messages      chan interface{}
}

func (m *MockPlayerActor) GetPlayerID() string     { return m.PlayerID }
func (m *MockPlayerActor) GetPlayerName() string   { return m.PlayerName }
func (m *MockPlayerActor) GetSessionToken() string { return m.SessionToken }
func (m *MockPlayerActor) GetState() interfaces.PlayerState { return m.CurrentState }

func (m *MockPlayerActor) TransitionToLobby(lobbyID string) error {
	m.CurrentState = interfaces.StateInLobby
	return nil
}

func (m *MockPlayerActor) TransitionToGame(gameID string) error {
	m.CurrentState = interfaces.StateInGame
	return nil
}

func (m *MockPlayerActor) TransitionToIdle() error {
	m.CurrentState = interfaces.StateIdle
	return nil
}

func (m *MockPlayerActor) SendServerMessage(message interface{}) {
	if m.Messages == nil {
		m.Messages = make(chan interface{}, 10)
	}
	m.Messages <- message
}

// MockBroadcaster for testing
type MockBroadcaster struct{}
func (m *MockBroadcaster) BroadcastToGame(gameID string, event core.Event) error { return nil }
func (m *MockBroadcaster) SendToPlayer(gameID, playerID string, event core.Event) error { return nil }

// MockSupervisor for testing
type MockSupervisor struct {
	CreateGameCalls int
	CreateGameResult interfaces.GameActorInterface
}

func (m *MockSupervisor) CreateGameWithPlayers(gameID string, players map[string]*core.Player) (interfaces.GameActorInterface, error) {
	m.CreateGameCalls++
	return m.CreateGameResult, nil
}

func (m *MockSupervisor) GetActor(gameID string) (interfaces.GameActorInterface, bool) { return nil, false }
func (m *MockSupervisor) RemoveGame(gameID string) {}

// MockDataStore for testing
type MockDataStore struct{}
func (m *MockDataStore) AppendEvent(gameID string, event core.Event) error { return nil }
func (m *MockDataStore) LoadEvents(gameID string, afterSequence int) ([]core.Event, error) { return nil, nil }
func (m *MockDataStore) CreateSnapshot(gameID string, state core.GameState) error { return nil }
func (m *MockDataStore) GetLatestSnapshot(gameID string) (*core.GameState, error) { return nil, nil }
func (m *MockDataStore) Close() error { return nil }
func (m *MockDataStore) GetEventsSince(gameID string, timestamp string) ([]core.Event, error) { return nil, nil }