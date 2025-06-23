package game

import (
	"context"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/mocks"
)

// MockGameActor for testing SessionManager
type MockGameActor struct {
	GameID           string
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
	mockDatastore := &mocks.MockDataStore{}
	mockBroadcaster := &mocks.MockBroadcaster{}
	mockSupervisor := &mocks.MockSupervisor{}

	sessionManager := NewSessionManager(ctx, mockDatastore, mockBroadcaster, mockSupervisor)

	lobbyID := "test-lobby"
	playerActors := map[string]interfaces.PlayerActorInterface{
		"player1": &MockPlayerActor{PlayerID: "player1"},
		"player2": &MockPlayerActor{PlayerID: "player2"},
	}

	// Set up supervisor mock to return a game actor
	mockGameActor := &MockGameActor{
		GameID:           lobbyID,
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

	mockSupervisor.CreateGameWithPlayersResults = []mocks.CreateGameWithPlayersResult{
		{Actor: mockGameActor, Error: nil},
	}

	err := sessionManager.CreateGameFromLobby(lobbyID, playerActors)

	if err != nil {
		t.Fatalf("CreateGameFromLobby failed: %v", err)
	}

	// Verify that a game actor was created
	if len(mockSupervisor.CreateGameWithPlayersCalls) != 1 {
		t.Errorf("Expected CreateGameWithPlayers to be called once, got %d", len(mockSupervisor.CreateGameWithPlayersCalls))
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
			// The first message should be TransitionToGame
			transitionMsg, ok := msg.(interfaces.TransitionToGame)
			if !ok {
				t.Fatalf("Expected first message to be TransitionToGame, got %T", msg)
			}
			if transitionMsg.GameID != lobbyID {
				t.Errorf("Expected GameID %s, got %s", lobbyID, transitionMsg.GameID)
			}

			// The second message should be the state update event
			select {
			case msg = <-mockActor.Messages:
				updateMsg, ok := msg.(core.Event)
				if !ok {
					t.Fatalf("Expected second message to be core.Event, got %T", msg)
				}
				if updateMsg.Type != "GAME_STATE_UPDATE" {
					t.Errorf("Expected GAME_STATE_UPDATE, got %s", updateMsg.Type)
				}
			case <-time.After(100 * time.Millisecond):
				t.Errorf("Player %s did not receive state update event", playerID)
			}

		case <-time.After(100 * time.Millisecond):
			t.Errorf("Player %s did not receive any messages", playerID)
		}
	}
}

// MockPlayerActor for testing SessionManager
type MockPlayerActor struct {
	PlayerID     string
	PlayerName   string
	SessionToken string
	CurrentState interfaces.PlayerState
	Messages     chan interface{}
}

func (m *MockPlayerActor) GetPlayerID() string              { return m.PlayerID }
func (m *MockPlayerActor) GetPlayerName() string            { return m.PlayerName }
func (m *MockPlayerActor) GetSessionToken() string          { return m.SessionToken }
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

