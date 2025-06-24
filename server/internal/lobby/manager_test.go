package lobby

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/mocks"
)

// MockPlayerActor for testing LobbyManager
type MockPlayerActor struct {
	PlayerID     string
	PlayerName   string
	CurrentState interfaces.PlayerState
	Messages     chan interface{}
}

func (m *MockPlayerActor) GetPlayerID() string              { return m.PlayerID }
func (m *MockPlayerActor) GetPlayerName() string            { return m.PlayerName }
func (m *MockPlayerActor) GetSessionToken() string          { return "test-token" }
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

// MockSessionManager for testing LobbyManager
type MockSessionManager struct {
	CreateGameCalls int
	CreateGameError error
}

func (m *MockSessionManager) CreateGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	m.CreateGameCalls++
	return m.CreateGameError
}

func (m *MockSessionManager) JoinGame(gameID string, player interfaces.PlayerActorInterface) error {
	return nil
}
func (m *MockSessionManager) LeaveGame(gameID string, playerID string) error           { return nil }
func (m *MockSessionManager) SendActionToGame(gameID string, action interface{}) error { return nil }

// TestLobbyManager_StartGame tests the full game start flow from the lobby
func TestLobbyManager_StartGame(t *testing.T) {
	mockSessionManager := &mocks.MockSessionManager{}
	lobbyManager := NewLobbyManager(mockSessionManager)

	// Create a lobby and add players
	hostActor := &MockPlayerActor{PlayerID: "host", PlayerName: "Host"}
	player2 := &MockPlayerActor{PlayerID: "player2", PlayerName: "Player 2"}

	lobbyID := "test-lobby"
	lobby := NewLobby(lobbyID, "Test Lobby", hostActor.GetPlayerID(), hostActor)
	lobby.AddPlayer(player2) // Add enough players to meet min players

	lobbyManager.lobbies[lobbyID] = lobby

	// Test starting the game
	err := lobbyManager.StartGame(lobbyID, "host")
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	// Verify that CreateGameFromLobby was called
	if len(mockSessionManager.CreateGameFromLobbyCalls) != 1 {
		t.Errorf("Expected CreateGameFromLobby to be called once, got %d", len(mockSessionManager.CreateGameFromLobbyCalls))
	}

	// Verify that the lobby was removed after starting
	if _, exists := lobbyManager.lobbies[lobbyID]; exists {
		t.Error("Expected lobby to be removed after game start")
	}
}

// TestLobbyManager_StartGame_NotEnoughPlayers tests starting with insufficient players
func TestLobbyManager_StartGame_NotEnoughPlayers(t *testing.T) {
	mockSessionManager := &mocks.MockSessionManager{}
	lobbyManager := NewLobbyManager(mockSessionManager)

	hostActor := &MockPlayerActor{PlayerID: "host", PlayerName: "Host"}
	lobbyID := "test-lobby"

	lobbyManager.lobbies[lobbyID] = NewLobby(lobbyID, "Test Lobby", hostActor.GetPlayerID(), hostActor)

	// Attempt to start with only one player
	err := lobbyManager.StartGame(lobbyID, "host")
	if err == nil {
		t.Error("Expected error when starting game with not enough players")
	}
}

// TestLobbyManager_StartGame_NotHost tests starting by a non-host player
func TestLobbyManager_StartGame_NotHost(t *testing.T) {
	mockSessionManager := &mocks.MockSessionManager{}
	lobbyManager := NewLobbyManager(mockSessionManager)

	hostActor := &MockPlayerActor{PlayerID: "host", PlayerName: "Host"}
	player2 := &MockPlayerActor{PlayerID: "player2", PlayerName: "Player 2"}

	lobbyID := "test-lobby"
	lobby := NewLobby(lobbyID, "Test Lobby", hostActor.GetPlayerID(), hostActor)
	lobby.AddPlayer(player2)

	lobbyManager.lobbies[lobbyID] = lobby

	// Attempt to start game by non-host
	err := lobbyManager.StartGame(lobbyID, "player2")
	if err == nil {
		t.Error("Expected error when non-host tries to start the game")
	}
}

// TestLobby_AddAndRemovePlayer tests adding and removing players from a lobby
func TestLobby_AddAndRemovePlayer(t *testing.T) {
	hostActor := &MockPlayerActor{PlayerID: "host", PlayerName: "Host"}
	lobby := NewLobby("test-lobby", "Test Lobby", hostActor.GetPlayerID(), hostActor)

	// Add a player
	player2 := &MockPlayerActor{PlayerID: "player2", PlayerName: "Player 2"}
	err := lobby.AddPlayer(player2)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	if len(lobby.Players) != 2 {
		t.Errorf("Expected 2 players in lobby, got %d", len(lobby.Players))
	}

	// Remove the player
	lobby.RemovePlayer("player2")

	if len(lobby.Players) != 1 {
		t.Errorf("Expected 1 player in lobby after removal, got %d", len(lobby.Players))
	}
}

// TestLobby_LobbyFull tests joining a full lobby
func TestLobby_LobbyFull(t *testing.T) {
	hostActor := &MockPlayerActor{PlayerID: "host", PlayerName: "Host"}
	lobby := NewLobby("test-lobby", "Test Lobby", hostActor.GetPlayerID(), hostActor)
	lobby.MaxPlayers = 1

	// Attempt to join full lobby
	player2 := &MockPlayerActor{PlayerID: "player2", PlayerName: "Player 2"}
	err := lobby.AddPlayer(player2)
	if err != ErrLobbyFull {
		t.Errorf("Expected ErrLobbyFull, got %v", err)
	}
}

// TestLobby_StateUpdateBroadcast tests that state updates are broadcast correctly
func TestLobby_StateUpdateBroadcast(t *testing.T) {
	hostActor := &MockPlayerActor{PlayerID: "host", PlayerName: "Host"}
	lobby := NewLobby("test-lobby", "Test Lobby", hostActor.GetPlayerID(), hostActor)

	// Add player and check for broadcast
	player2 := &MockPlayerActor{PlayerID: "player2", PlayerName: "Player 2"}
	err := lobby.AddPlayer(player2)
	if err != nil {
		t.Fatalf("Failed to add player: %v", err)
	}

	// Host should receive update
	select {
	case msg := <-hostActor.Messages:
		event, ok := msg.(core.Event)
		assert.True(t, ok, "Expected message to be of type core.Event, got %T", msg)
		assert.Equal(t, core.EventType("LOBBY_STATE_UPDATE"), event.Type)
		
		// Check the event payload
		players, ok := event.Payload["players"].([]PlayerInfo)
		assert.True(t, ok, "Expected payload to contain a slice of PlayerInfo")
		assert.Len(t, players, 2, "Expected 2 players in the state update")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Host did not receive state update")
	}

	// New player should also receive update
	select {
	case msg := <-player2.Messages:
		event, ok := msg.(core.Event)
		assert.True(t, ok, "Expected message to be of type core.Event, got %T", msg)
		assert.Equal(t, core.EventType("LOBBY_STATE_UPDATE"), event.Type)
		
		players, ok := event.Payload["players"].([]PlayerInfo)
		assert.True(t, ok, "Expected payload to contain a slice of PlayerInfo")
		assert.Len(t, players, 2, "Expected 2 players in the state update")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("New player did not receive state update")
	}

	// Remove player and check for broadcast
	lobby.RemovePlayer("player2")

	// Host should receive update
	select {
	case msg := <-hostActor.Messages:
		event, ok := msg.(core.Event)
		assert.True(t, ok, "Expected message to be of type core.Event, got %T", msg)
		assert.Equal(t, core.EventType("LOBBY_STATE_UPDATE"), event.Type)
		
		players, ok := event.Payload["players"].([]PlayerInfo)
		assert.True(t, ok, "Expected payload to contain a slice of PlayerInfo")
		assert.Len(t, players, 1, "Expected 1 player after removal")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Host did not receive state update after removal")
	}
}
