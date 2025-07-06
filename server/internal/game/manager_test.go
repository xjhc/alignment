package game

import (
	"context"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/mocks"
	"github.com/xjhc/alignment/server/internal/lifecycle"
	"github.com/xjhc/alignment/server/internal/events"
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

// TestGameLifecycleManager_CreateGameFromLobby tests the atomic game creation from a lobby
func TestGameLifecycleManager_CreateGameFromLobby(t *testing.T) {
	ctx := context.Background()

	// Mocks
	mockDatastore := &mocks.MockDataStore{}
	mockBroadcaster := &mocks.MockBroadcaster{}
	mockSupervisor := &mocks.MockSupervisor{}
	eventBus := events.NewEventBus()

	lifecycleManager := lifecycle.NewGameLifecycleManager(ctx, mockDatastore, mockBroadcaster, mockSupervisor, eventBus, nil, nil)

	playerActors := map[string]interfaces.PlayerActorInterface{
		"player1": NewMockPlayerActor("player1", "Player 1"),
		"player2": NewMockPlayerActor("player2", "Player 2"),
	}

	// Create lobby via HTTP first (this is the new flow)
	lobbyID, _, _, err := lifecycleManager.CreateLobbyViaHTTP("player1", "Host Player", "Test Game", "👤", false)
	if err != nil {
		t.Fatalf("CreateLobbyViaHTTP failed: %v", err)
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
	
	// Set up the supervisor to return the game actor when GetActor is called
	mockSupervisor.GetActorResults = []mocks.GetActorResult{
		{Actor: mockGameActor, Found: true},
	}

	// Join the host to the lobby
	err = lifecycleManager.JoinLobbyWithActor(lobbyID, playerActors["player1"])
	if err != nil {
		t.Fatalf("JoinLobbyWithActor failed for host: %v", err)
	}

	// Join second player to the lobby  
	err = lifecycleManager.JoinLobbyWithActor(lobbyID, playerActors["player2"])
	if err != nil {
		t.Fatalf("JoinLobbyWithActor failed for player2: %v", err)
	}

	// Start the game (this triggers the internal CreateGameFromLobby logic)
	err = lifecycleManager.StartGame(lobbyID, "player1")
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Wait a bit for the countdown to complete and game to start
	time.Sleep(4 * time.Second)

	// Verify that a game actor was created
	if len(mockSupervisor.CreateGameWithPlayersCalls) != 1 {
		t.Errorf("Expected CreateGameWithPlayers to be called once, got %d", len(mockSupervisor.CreateGameWithPlayersCalls))
	}

	// Verify the game actor exists
	_, exists := lifecycleManager.GetGameActor(lobbyID)
	if !exists {
		t.Errorf("Expected game actor to exist for game %s", lobbyID)
	}

	// Verify that each player actor received game events
	for playerID, playerActor := range playerActors {
		mockActor, ok := playerActor.(*MockPlayerActor)
		if !ok {
			t.Fatalf("Failed to cast player actor to mock")
		}

		// Consume all messages and look for the expected events
		var receivedGameStarted, receivedStateUpdate bool
		timeout := time.After(500 * time.Millisecond)
		
		for !receivedGameStarted || !receivedStateUpdate {
			select {
			case msg := <-mockActor.Messages:
				event, ok := msg.(core.Event)
				if !ok {
					continue // Skip non-event messages
				}
				
				switch event.Type {
				case core.EventGameStarted:
					receivedGameStarted = true
					if event.GameID != lobbyID {
						t.Errorf("Player %s: Expected GameID %s, got %s", playerID, lobbyID, event.GameID)
					}
				case "GAME_STATE_UPDATE":
					receivedStateUpdate = true
				}
				
			case <-timeout:
				if !receivedGameStarted {
					t.Errorf("Player %s did not receive GAME_STARTED event", playerID)
				}
				if !receivedStateUpdate {
					t.Errorf("Player %s did not receive GAME_STATE_UPDATE event", playerID)
				}
				break
			}
		}
	}
}

// MockPlayerActor for testing GameLifecycleManager
type MockPlayerActor struct {
	PlayerID     string
	PlayerName   string
	PlayerAvatar string
	SessionToken string
	CurrentState interfaces.PlayerState
	Messages     chan interface{}
}

func NewMockPlayerActor(playerID, playerName string) *MockPlayerActor {
	return &MockPlayerActor{
		PlayerID:     playerID,
		PlayerName:   playerName,
		PlayerAvatar: "👤",
		SessionToken: "test-token",
		CurrentState: interfaces.StateIdle,
		Messages:     make(chan interface{}, 10),
	}
}

func (m *MockPlayerActor) GetPlayerID() string              { return m.PlayerID }
func (m *MockPlayerActor) GetPlayerName() string            { return m.PlayerName }
func (m *MockPlayerActor) GetPlayerAvatar() string          { return m.PlayerAvatar }
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

func (m *MockPlayerActor) Stop() {
	// Mock implementation for testing
}

