package lifecycle

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/mocks"
	"github.com/xjhc/alignment/server/internal/store"
	"github.com/xjhc/alignment/server/internal/test_helpers"
)

// MockPlayerActor for testing (since each test file has its own)
type MockPlayerActor struct {
	PlayerID     string
	PlayerName   string
	PlayerAvatar string
	CurrentState interfaces.PlayerState
	Messages     chan interface{}
}

func (m *MockPlayerActor) GetPlayerID() string              { return m.PlayerID }
func (m *MockPlayerActor) GetPlayerName() string            { return m.PlayerName }
func (m *MockPlayerActor) GetPlayerAvatar() string          { return m.PlayerAvatar }
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

func (m *MockPlayerActor) Stop() {
	// Mock implementation for testing
}

// MockActiveUserTracker for testing
type MockActiveUserTracker struct{}

func (m *MockActiveUserTracker) RemoveUserSessionByGameAndPlayer(gameID, playerID string) {
	// Mock implementation for testing
}

// Helper to create a test player actor
func createTestPlayerActor(id, name string) *MockPlayerActor {
	return &MockPlayerActor{
		PlayerID:     id,
		PlayerName:   name,
		CurrentState: interfaces.StateIdle,
		Messages:     make(chan interface{}, 10),
	}
}

// Helper to set up a GameLifecycleManager for testing
func setupTestManager(t *testing.T) (*GameLifecycleManager, *mocks.MockDataStore, *mocks.MockSupervisor, *events.EventBus) {
	ctx := context.Background()
	datastore := &mocks.MockDataStore{}
	supervisor := &mocks.MockSupervisor{}
	broadcaster := &mocks.MockBroadcaster{}
	eventBus := events.NewEventBus()

	// Mock PostgresStore and ActiveUserTracker for tests
	mockPostgresStore := (*store.PostgresStore)(nil) // Nil pointer for tests
	mockActiveUserTracker := &MockActiveUserTracker{}
	
	manager := NewGameLifecycleManager(ctx, datastore, broadcaster, supervisor, eventBus, mockPostgresStore, mockActiveUserTracker)

	t.Cleanup(func() {
		manager.Stop()
		// Don't call eventBus.Close() here as the manager might already close it
	})

	return manager, datastore, supervisor, eventBus
}

func TestGameLifecycleManager_CreateLobbyViaHTTP(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	hostName := "TestHost"
	lobbyName := "Test Lobby"
	avatar := "test-avatar"

	lobbyID, hostPlayerID, sessionToken, err := manager.CreateLobbyViaHTTP("user1", hostName, lobbyName, avatar, false, core.GameSettings{})

	assert.NoError(t, err)
	assert.NotEmpty(t, lobbyID)
	assert.NotEmpty(t, hostPlayerID)
	assert.NotEmpty(t, sessionToken)

	// Verify lobby was created
	lobby, err := manager.GetLobbyByID(lobbyID)
	assert.NoError(t, err)
	assert.Equal(t, lobbyName, lobby.Name)
	assert.Equal(t, hostPlayerID, lobby.HostPlayerID)
	assert.Equal(t, "WAITING_FOR_HOST", lobby.Status)

	// Verify token exists and is valid
	token, err := manager.ValidateSessionToken(sessionToken)
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestGameLifecycleManager_JoinLobbyWithActor(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Create a lobby first
	hostName := "TestHost"
	lobbyName := "Test Lobby"
	lobbyID, hostPlayerID, _, err := manager.CreateLobbyViaHTTP("user1", hostName, lobbyName, "avatar", false)
	require.NoError(t, err)

	// Create host actor and join
	hostActor := createTestPlayerActor(hostPlayerID, hostName)
	err = manager.JoinLobbyWithActor(lobbyID, hostActor)
	assert.NoError(t, err)

	// Verify lobby status changed
	lobby, err := manager.GetLobbyByID(lobbyID)
	assert.NoError(t, err)
	assert.Equal(t, "WAITING", lobby.Status)

	// Add another player
	player2 := createTestPlayerActor("player2", "Player2")
	err = manager.JoinLobbyWithActor(lobbyID, player2)
	assert.NoError(t, err)

	// Verify both players are in lobby
	playerActors := lobby.GetPlayerActors()
	assert.Len(t, playerActors, 2)
	assert.Contains(t, playerActors, hostPlayerID)
	assert.Contains(t, playerActors, "player2")
}

func TestGameLifecycleManager_JoinNonExistentLobby(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	player := createTestPlayerActor("player1", "Player1")
	err := manager.JoinLobbyWithActor("non-existent-lobby", player)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby not found")
}

func TestGameLifecycleManager_StartGame_NotHost(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Create lobby with host
	lobbyID, hostPlayerID, _, err := manager.CreateLobbyViaHTTP("user1", "Host", "Test Lobby", "avatar", false, core.GameSettings{})
	require.NoError(t, err)

	hostActor := createTestPlayerActor(hostPlayerID, "Host")
	err = manager.JoinLobbyWithActor(lobbyID, hostActor)
	require.NoError(t, err)

	// Add another player  
	player2 := createTestPlayerActor("player2", "Player2")
	err = manager.JoinLobbyWithActor(lobbyID, player2)
	require.NoError(t, err)

	// Try to start game as non-host
	err = manager.StartGame(lobbyID, "player2")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "only the host can start the game")
}

func TestGameLifecycleManager_StartGame_NotEnoughPlayers(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Create lobby with only host
	lobbyID, hostPlayerID, _, err := manager.CreateLobbyViaHTTP("user1", "Host", "Test Lobby", "avatar", false, core.GameSettings{})
	require.NoError(t, err)

	hostActor := createTestPlayerActor(hostPlayerID, "Host")
	err = manager.JoinLobbyWithActor(lobbyID, hostActor)
	require.NoError(t, err)

	// Try to start game with only one player (minimum is 2)
	err = manager.StartGame(lobbyID, hostPlayerID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not enough players")
}

func TestGameLifecycleManager_ValidateSessionToken(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Test with invalid token
	_, err := manager.ValidateSessionToken("invalid-token")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")

	// Create a valid token
	_, _, sessionToken, err := manager.CreateLobbyViaHTTP("user1", "Host", "Test Lobby", "avatar", false, core.GameSettings{})
	require.NoError(t, err)

	// Test with valid token
	token, err := manager.ValidateSessionToken(sessionToken)
	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestGameLifecycleManager_ValidateSession(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Create lobby and get token
	lobbyID, hostPlayerID, sessionToken, err := manager.CreateLobbyViaHTTP("user1", "Host", "Test Lobby", "avatar", false, core.GameSettings{})
	require.NoError(t, err)

	// Test valid session
	isValid := manager.ValidateSession(lobbyID, hostPlayerID, sessionToken)
	assert.True(t, isValid)

	// Test invalid session - wrong player ID
	isValid = manager.ValidateSession(lobbyID, "wrong-player", sessionToken)
	assert.False(t, isValid)

	// Test invalid session - wrong token
	isValid = manager.ValidateSession(lobbyID, hostPlayerID, "wrong-token")
	assert.False(t, isValid)
}

func TestGameLifecycleManager_GetPlayerInfo(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Create lobby
	lobbyID, hostPlayerID, _, err := manager.CreateLobbyViaHTTP("user1", "TestHost", "Test Lobby", "test-avatar", false, core.GameSettings{})
	require.NoError(t, err)

	// Get player info
	name, avatar, err := manager.GetPlayerInfo(lobbyID, hostPlayerID)
	assert.NoError(t, err)
	assert.Equal(t, "TestHost", name)
	assert.Equal(t, "test-avatar", avatar)

	// Test non-existent player
	_, _, err = manager.GetPlayerInfo(lobbyID, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "player info not found")
}

func TestGameLifecycleManager_GetLobbyList(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Initially empty
	lobbies := manager.GetLobbyList()
	assert.Len(t, lobbies, 0)

	// Create some lobbies
	_, _, _, err := manager.CreateLobbyViaHTTP("user1", "Host1", "Lobby1", "avatar1", false, core.GameSettings{})
	require.NoError(t, err)
	_, _, _, err = manager.CreateLobbyViaHTTP("user2", "Host2", "Lobby2", "avatar2", false, core.GameSettings{})
	require.NoError(t, err)

	// Check lobby list
	lobbies = manager.GetLobbyList()
	assert.Len(t, lobbies, 2)

	// Verify lobby structure
	for _, lobby := range lobbies {
		lobbyMap, ok := lobby.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, lobbyMap, "id")
		assert.Contains(t, lobbyMap, "name")
		assert.Contains(t, lobbyMap, "player_count")
		assert.Contains(t, lobbyMap, "max_players")
		assert.Contains(t, lobbyMap, "can_join")
		assert.Contains(t, lobbyMap, "status")
	}
}

func TestGameLifecycleManager_JoinLobby(t *testing.T) {
	manager, _, _, _ := setupTestManager(t)

	// Create a lobby
	lobbyID, _, _, err := manager.CreateLobbyViaHTTP("user1", "Host", "Test Lobby", "avatar", false, core.GameSettings{})
	require.NoError(t, err)

	// Join the lobby
	playerID, sessionToken, err := manager.JoinLobby(lobbyID, "user2", "Joiner", "joiner-avatar")
	assert.NoError(t, err)
	assert.NotEmpty(t, playerID)
	assert.NotEmpty(t, sessionToken)

	// Verify token is valid
	isValid := manager.ValidateSession(lobbyID, playerID, sessionToken)
	assert.True(t, isValid)

	// Try to join non-existent lobby
	_, _, err = manager.JoinLobby("non-existent", "user3", "Player", "avatar")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lobby not found")
}

func TestGameLifecycleManager_EventHandling(t *testing.T) {
	manager, _, _, eventBus := setupTestManager(t)

	// Create a lobby with players
	lobbyID, hostPlayerID, _, err := manager.CreateLobbyViaHTTP("user1", "Host", "Test Lobby", "avatar", false, core.GameSettings{})
	require.NoError(t, err)

	hostActor := createTestPlayerActor(hostPlayerID, "Host")
	err = manager.JoinLobbyWithActor(lobbyID, hostActor)
	require.NoError(t, err)

	player2 := createTestPlayerActor("player2", "Player2")
	err = manager.JoinLobbyWithActor(lobbyID, player2)
	require.NoError(t, err)

	// Verify lobby has 2 players
	lobby, _ := manager.GetLobbyByID(lobbyID)
	assert.Len(t, lobby.GetPlayerActors(), 2)

	// Set up a WaitGroup to wait for event processing
	var wg sync.WaitGroup
	wg.Add(1)

	// Monitor for lobby state changes using a goroutine
	go func() {
		defer wg.Done()
		// Poll for the expected state change with a timeout
		for i := 0; i < 50; i++ { // Poll up to 50 times
			lobby, _ := manager.GetLobbyByID(lobbyID)
			if len(lobby.GetPlayerActors()) == 1 {
				return // Expected state reached
			}
			time.Sleep(10 * time.Millisecond) // Brief pause between polls
		}
	}()

	// Publish player disconnected event
	eventBus.Publish(events.PlayerDisconnectedEvent{
		PlayerID: "player2",
		LobbyID:  lobbyID,
	})

	// Wait deterministically for the event to be processed
	test_helpers.WaitWithTimeout(&wg, 1*time.Second, t)

	// Verify player was removed from lobby
	lobby, _ = manager.GetLobbyByID(lobbyID)
	players := lobby.GetPlayerActors()
	assert.Len(t, players, 1)
	assert.Contains(t, players, hostPlayerID)
	assert.NotContains(t, players, "player2")
}

func TestGameLifecycleManager_RehydrateGamesFromStore_Success(t *testing.T) {
	manager, mockDataStore, mockSupervisor, _ := setupTestManager(t)

	// Set up mock data for active games
	gameIDs := []string{"game-1", "game-2", "game-3"}
	mockDataStore.ListActiveGamesResults = []mocks.ListActiveGamesResult{
		{GameIDs: gameIDs, Error: nil},
	}

	// Set up mock snapshots for each game
	for i, gameID := range gameIDs {
		gameState := test_helpers.CreateTestGameState(gameID, i+1)
		mockDataStore.GetLatestSnapshotResults = append(mockDataStore.GetLatestSnapshotResults, mocks.GetLatestSnapshotResult{
			State: gameState,
			Error: nil,
		})
		
		// Mock successful game actor creation
		mockGameActor := mocks.NewMockGameActor(gameID)
		mockSupervisor.CreateGameWithPlayersResults = append(mockSupervisor.CreateGameWithPlayersResults, mocks.CreateGameWithPlayersResult{
			Actor: mockGameActor,
			Error: nil,
		})
	}

	// Execute rehydration
	err := manager.RehydrateGamesFromStore()

	// Verify success
	assert.NoError(t, err)
	
	// Verify all calls were made
	assert.Len(t, mockDataStore.ListActiveGamesCalls, 1)
	assert.Len(t, mockDataStore.GetLatestSnapshotCalls, 3)
	assert.Len(t, mockSupervisor.CreateGameWithPlayersCalls, 3)
	
	// Verify games are tracked in manager
	manager.mutex.RLock()
	assert.Len(t, manager.gameActors, 3)
	for _, gameID := range gameIDs {
		assert.Contains(t, manager.gameActors, gameID)
	}
	manager.mutex.RUnlock()
}

func TestGameLifecycleManager_RehydrateGamesFromStore_EmptyGameList(t *testing.T) {
	manager, mockDataStore, _, _ := setupTestManager(t)

	// No active games
	mockDataStore.ListActiveGamesResults = []mocks.ListActiveGamesResult{
		{GameIDs: []string{}, Error: nil},
	}

	err := manager.RehydrateGamesFromStore()

	assert.NoError(t, err)
	assert.Len(t, mockDataStore.ListActiveGamesCalls, 1)

	// Verify no games tracked
	manager.mutex.RLock()
	assert.Len(t, manager.gameActors, 0)
	manager.mutex.RUnlock()
}