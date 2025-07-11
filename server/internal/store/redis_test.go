
package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjhc/alignment/core"
)

func setupTestRedis(t *testing.T) *RedisDataStore {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default for local testing
	}

	datastore, err := NewRedisDataStore(redisAddr, "", 0)
	if err != nil {
		t.Skipf("Skipping Redis test - cannot connect to Redis at %s: %v", redisAddr, err)
	}

	// Cleanup function to run after the test
	t.Cleanup(func() {
		datastore.client.FlushDB(datastore.ctx)
		datastore.Close()
	})

	return datastore
}

func TestRedisDataStore_Connection(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	assert.NotNil(t, rds)
}

func TestRedisDataStore_AppendAndLoadEvents(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-game-1"

	event1 := core.Event{
		ID:        "event-1",
		Type:      "TEST_EVENT_1",
		GameID:    gameID,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"data": "value1"},
	}

	event2 := core.Event{
		ID:        "event-2",
		Type:      "TEST_EVENT_2",
		GameID:    gameID,
		Timestamp: time.Now().UTC().Add(1 * time.Second),
		Payload:   map[string]interface{}{"data": "value2"},
	}

	// Append events
	err := rds.AppendEvent(context.Background(), gameID, event1)
	assert.NoError(t, err)
	err = rds.AppendEvent(context.Background(), gameID, event2)
	assert.NoError(t, err)

	// Load events
	loadedEvents, err := rds.LoadEvents(context.Background(), gameID, 0)
	assert.NoError(t, err)
	assert.Len(t, loadedEvents, 2, "Should load 2 events")

	// Verify events are loaded correctly
	assert.Equal(t, event1.ID, loadedEvents[0].ID)
	assert.Equal(t, event1.Type, loadedEvents[0].Type)
	assert.Equal(t, event1.Payload["data"], loadedEvents[0].Payload["data"])

	assert.Equal(t, event2.ID, loadedEvents[1].ID)
	assert.Equal(t, event2.Type, loadedEvents[1].Type)
	assert.Equal(t, event2.Payload["data"], loadedEvents[1].Payload["data"])
}

func TestRedisDataStore_CreateAndLoadSnapshot(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-game-2"

	// Create a test game state
	gameState := core.GameState{
		ID:         gameID,
		DayNumber:  1,
		Phase:      core.Phase{Type: core.PhaseSitrep},
		Players:    make(map[string]*core.Player),
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	// Add a test player
	gameState.Players["player1"] = &core.Player{
		ID:       "player1",
		Name:     "Test Player",
		IsAlive:  true,
		Tokens:   10,
		JoinedAt: time.Now().UTC(),
	}

	// Save snapshot
	err := rds.CreateSnapshot(context.Background(), gameID, gameState)
	assert.NoError(t, err)

	// Load snapshot
	loadedState, err := rds.LoadSnapshot(context.Background(), gameID)
	require.NoError(t, err)
	require.NotNil(t, loadedState)

	// Verify basic state fields
	assert.Equal(t, gameState.ID, loadedState.ID)
	assert.Equal(t, gameState.DayNumber, loadedState.DayNumber)
	assert.Equal(t, gameState.Phase.Type, loadedState.Phase.Type)
	assert.Len(t, loadedState.Players, 1)

	// Verify player data
	player := loadedState.Players["player1"]
	require.NotNil(t, player)
	assert.Equal(t, "player1", player.ID)
	assert.Equal(t, "Test Player", player.Name)
	assert.True(t, player.IsAlive)
	assert.Equal(t, 10, player.Tokens)
}

func TestRedisDataStore_GetEventCount(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-game-3"

	// Initially should have 0 events
	count, err := rds.GetEventCount(context.Background(), gameID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Add an event
	event := core.Event{
		ID:        "event-count-test",
		Type:      "TEST_EVENT",
		GameID:    gameID,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"test": true},
	}

	err = rds.AppendEvent(context.Background(), gameID, event)
	assert.NoError(t, err)

	// Should now have 1 event
	count, err = rds.GetEventCount(context.Background(), gameID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRedisDataStore_DeleteGame(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-game-4"

	// Create some data
	event := core.Event{
		ID:        "delete-test",
		Type:      "TEST_EVENT",
		GameID:    gameID,
		Timestamp: time.Now().UTC(),
		Payload:   map[string]interface{}{"test": true},
	}

	gameState := core.GameState{
		ID:        gameID,
		DayNumber: 1,
		Phase:     core.Phase{Type: core.PhaseLobby},
		Players:   make(map[string]*core.Player),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	err := rds.AppendEvent(context.Background(), gameID, event)
	assert.NoError(t, err)
	err = rds.CreateSnapshot(context.Background(), gameID, gameState)
	assert.NoError(t, err)

	// Verify data exists
	count, err := rds.GetEventCount(context.Background(), gameID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = rds.LoadSnapshot(context.Background(), gameID)
	assert.NoError(t, err)

	// Delete the game
	err = rds.DeleteGame(context.Background(), gameID)
	assert.NoError(t, err)

	// Verify data is gone
	count, err = rds.GetEventCount(context.Background(), gameID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, err = rds.LoadSnapshot(context.Background(), gameID)
	assert.Error(t, err)
}

func TestRedisDataStore_GetGameMetadata(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-game-5"

	gameState := core.GameState{
		ID:        gameID,
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseDiscussion},
		Players:   make(map[string]*core.Player),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Add two players
	gameState.Players["p1"] = &core.Player{ID: "p1", Name: "Player 1"}
	gameState.Players["p2"] = &core.Player{ID: "p2", Name: "Player 2"}

	// Create snapshot (which also saves metadata)
	err := rds.CreateSnapshot(context.Background(), gameID, gameState)
	assert.NoError(t, err)

	// Get metadata
	metadata, err := rds.GetGameMetadata(context.Background(), gameID)
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	// Verify metadata fields
	assert.Equal(t, "2", metadata["snapshot_day"])
	assert.Equal(t, string(core.PhaseDiscussion), metadata["phase"])
	assert.Equal(t, "2", metadata["player_count"])
}

func TestRedisDataStore_SnapshotChecksumValidation(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-checksum-game"

	// Create a game state with events
	gameState := core.GameState{
		ID:        gameID,
		DayNumber: 1,
		Phase:     core.Phase{Type: core.PhaseSitrep},
		Players:   make(map[string]*core.Player),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Add some test events to create event history
	events := []core.Event{
		{
			ID:        "event-1",
			Type:      core.EventGameStarted,
			GameID:    gameID,
			Timestamp: time.Now().UTC(),
			Payload:   map[string]interface{}{"test": "data1"},
		},
		{
			ID:        "event-2", 
			Type:      core.EventPlayerJoined,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now().UTC().Add(1 * time.Second),
			Payload:   map[string]interface{}{"name": "Test Player"},
		},
	}

	// Append events to create history
	for _, event := range events {
		err := rds.AppendEvent(context.Background(), gameID, event)
		require.NoError(t, err)
	}

	// Add player to game state
	gameState.Players["player1"] = &core.Player{
		ID:       "player1",
		Name:     "Test Player",
		IsAlive:  true,
		Tokens:   10,
		JoinedAt: time.Now().UTC(),
	}

	// Save valid snapshot
	err := rds.CreateSnapshot(context.Background(), gameID, gameState)
	require.NoError(t, err)

	// Verify valid snapshot loads correctly
	loadedState, err := rds.LoadSnapshot(context.Background(), gameID)
	require.NoError(t, err)
	assert.Equal(t, gameState.ID, loadedState.ID)
	assert.NotEmpty(t, loadedState.Checksum)

	// Test checksum validation by corrupting the snapshot manually
	snapshotKey := "game:" + gameID + ":snapshot"
	corruptedData := `{"id":"` + gameID + `","checksum":"invalid-checksum","day_number":1,"players":{}}`
	err = rds.client.Set(rds.ctx, snapshotKey, corruptedData, time.Hour).Err()
	require.NoError(t, err)

	// Attempt to load corrupted snapshot - should trigger recovery
	recoveredState, err := rds.LoadSnapshot(context.Background(), gameID)
	require.NoError(t, err)
	assert.Equal(t, gameID, recoveredState.ID)
	assert.NotEmpty(t, recoveredState.Checksum)
	
	// Verify recovery reconstructed the state correctly from events
	assert.Len(t, recoveredState.Players, 1)
	assert.Contains(t, recoveredState.Players, "player1")
}

func TestRedisDataStore_RecoverFromEventHistory_Success(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "test-recovery-game"

	// Create a sequence of events that build up a game state
	events := []core.Event{
		{
			ID:        "init-event",
			Type:      core.EventGameStarted,
			GameID:    gameID,
			Timestamp: time.Now().UTC(),
			Payload:   map[string]interface{}{},
		},
		{
			ID:        "player-join-1",
			Type:      core.EventPlayerJoined,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now().UTC().Add(1 * time.Second),
			Payload:   map[string]interface{}{"name": "Alice"},
		},
		{
			ID:        "player-join-2",
			Type:      core.EventPlayerJoined,
			GameID:    gameID,
			PlayerID:  "player2",
			Timestamp: time.Now().UTC().Add(2 * time.Second),
			Payload:   map[string]interface{}{"name": "Bob"},
		},
		{
			ID:        "phase-change",
			Type:      core.EventPhaseChanged,
			GameID:    gameID,
			Timestamp: time.Now().UTC().Add(3 * time.Second),
			Payload: map[string]interface{}{
				"phase_type": "DISCUSSION",
				"duration":   120.0,
			},
		},
	}

	// Append all events
	for _, event := range events {
		err := rds.AppendEvent(context.Background(), gameID, event)
		require.NoError(t, err)
	}

	// Recover state from event history
	recoveredState, err := rds.RecoverFromEventHistory(context.Background(), gameID)
	require.NoError(t, err)
	require.NotNil(t, recoveredState)

	// Verify recovered state has correct basic properties
	assert.Equal(t, gameID, recoveredState.ID)
	assert.Equal(t, core.PhaseDiscussion, recoveredState.Phase.Type)
	assert.Equal(t, 2, recoveredState.DayNumber) // Should increment from EventGameStarted
	assert.NotEmpty(t, recoveredState.Checksum)

	// Verify players were correctly reconstructed
	assert.Len(t, recoveredState.Players, 2)
	assert.Contains(t, recoveredState.Players, "player1")
	assert.Contains(t, recoveredState.Players, "player2")
	assert.Equal(t, "Alice", recoveredState.Players["player1"].Name)
	assert.Equal(t, "Bob", recoveredState.Players["player2"].Name)
}

func TestRedisDataStore_RecoverFromEventHistory_NoEvents(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "empty-game"

	// Try to recover from non-existent event history
	_, err := rds.RecoverFromEventHistory(context.Background(), gameID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no events found to reconstruct game state")
}

func TestRedisDataStore_ListActiveGames(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)

	// Initially should have no active games
	games, err := rds.ListActiveGames(context.Background())
	require.NoError(t, err)
	assert.Len(t, games, 0)

	// Create snapshots for multiple games
	gameIDs := []string{"active-game-1", "active-game-2", "active-game-3"}
	for i, gameID := range gameIDs {
		gameState := core.GameState{
			ID:        gameID,
			DayNumber: i + 1,
			Phase:     core.Phase{Type: core.PhaseSitrep},
			Players:   make(map[string]*core.Player),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		err := rds.CreateSnapshot(context.Background(), gameID, gameState)
		require.NoError(t, err)
	}

	// Should now list all active games
	games, err = rds.ListActiveGames(context.Background())
	require.NoError(t, err)
	assert.Len(t, games, 3)
	
	// Verify all game IDs are present
	for _, expectedID := range gameIDs {
		assert.Contains(t, games, expectedID)
	}
}

func TestRedisDataStore_SnapshotCorruption_CompleteRecovery(t *testing.T) {
	if os.Getenv("CI") == "" && os.Getenv("REDIS_ADDR") == "" {
		t.Skip("Skipping Redis test locally; set REDIS_ADDR to run.")
	}

	rds := setupTestRedis(t)
	gameID := "corruption-recovery-test"

	// Create a complex event sequence representing a full game flow
	events := []core.Event{
		{
			ID:        "game-init",
			Type:      core.EventGameStarted,
			GameID:    gameID,
			Timestamp: time.Now().UTC(),
			Payload:   map[string]interface{}{},
		},
		{
			ID:        "role-assign-1",
			Type:      core.EventRoleAssigned,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now().UTC().Add(1 * time.Second),
			Payload: map[string]interface{}{
				"role_type":        "CEO",
				"role_name":        "Chief Executive Officer",
				"role_description": "Corporate leader",
				"alignment":        "HUMAN",
				"persona_name":     "Executive Alice",
				"job_title":        "CEO",
			},
		},
		{
			ID:        "tokens-award-1",
			Type:      core.EventTokensAwarded,
			GameID:    gameID,
			PlayerID:  "player1",
			Timestamp: time.Now().UTC().Add(2 * time.Second),
			Payload:   map[string]interface{}{"amount": 5.0},
		},
		{
			ID:        "phase-transition",
			Type:      core.EventPhaseChanged,
			GameID:    gameID,
			Timestamp: time.Now().UTC().Add(3 * time.Second),
			Payload: map[string]interface{}{
				"phase_type": "NIGHT",
				"duration":   30.0,
			},
		},
	}

	// Append all events to create comprehensive history
	for _, event := range events {
		err := rds.AppendEvent(context.Background(), gameID, event)
		require.NoError(t, err)
	}

	// Create a valid snapshot representing the final state
	finalGameState := core.GameState{
		ID:        gameID,
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players:   make(map[string]*core.Player),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	finalGameState.Players["player1"] = &core.Player{
		ID:        "player1",
		Name:      "Executive Alice",
		JobTitle:  "CEO",
		IsAlive:   true,
		Tokens:    6, // 1 starting + 5 awarded
		Alignment: "HUMAN",
		Role: &core.Role{
			Type:        "CEO",
			Name:        "Chief Executive Officer",
			Description: "Corporate leader",
		},
	}

	err := rds.CreateSnapshot(context.Background(), gameID, finalGameState)
	require.NoError(t, err)

	// Verify snapshot loads correctly initially
	loadedState, err := rds.LoadSnapshot(context.Background(), gameID)
	require.NoError(t, err)
	assert.Equal(t, 6, loadedState.Players["player1"].Tokens)

	// Manually corrupt the snapshot with completely invalid JSON
	snapshotKey := "game:" + gameID + ":snapshot"
	corruptedJSON := `{invalid json structure with missing quotes and malformed data}`
	err = rds.client.Set(rds.ctx, snapshotKey, corruptedJSON, time.Hour).Err()
	require.NoError(t, err)

	// Attempt to load - should trigger full recovery from events
	recoveredState, err := rds.LoadSnapshot(context.Background(), gameID)
	require.NoError(t, err)

	// Verify complete recovery with all event-driven state changes
	assert.Equal(t, gameID, recoveredState.ID)
	assert.Equal(t, core.PhaseNight, recoveredState.Phase.Type)
	assert.Equal(t, 2, recoveredState.DayNumber)
	
	// Verify player state was completely reconstructed from events
	require.Contains(t, recoveredState.Players, "player1")
	player := recoveredState.Players["player1"]
	assert.Equal(t, "Executive Alice", player.Name)
	assert.Equal(t, "CEO", player.JobTitle)
	assert.Equal(t, "HUMAN", player.Alignment)
	assert.Equal(t, 6, player.Tokens) // Should reflect the token award event
	require.NotNil(t, player.Role)
	assert.Equal(t, core.RoleType("CEO"), player.Role.Type)
	assert.Equal(t, "Chief Executive Officer", player.Role.Name)
	
	// Verify the recovered state has a valid checksum
	assert.NotEmpty(t, recoveredState.Checksum)
	isValid, err := recoveredState.ValidateChecksum()
	require.NoError(t, err)
	assert.True(t, isValid)
}

func TestRedisDataStore_ContextCancellation(t *testing.T) {
	rds := setupTestRedis(t)
	defer rds.Close()

	gameID := "test-game-context-cancel"
	event := core.Event{
		ID:        "event-1",
		Type:      core.EventGameStarted,
		GameID:    gameID,
		PlayerID:  "player1",
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"test": "data"},
	}

	// Test with canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context immediately

	// AppendEvent should fail with context canceled error
	err := rds.AppendEvent(ctx, gameID, event)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// GetEvents should fail with context canceled error
	_, err = rds.GetEvents(ctx, gameID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// ListActiveGames should fail with context canceled error
	_, err = rds.ListActiveGames(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// CreateSnapshot should fail with context canceled error
	gameState := core.NewGameState(gameID, time.Now())
	err = rds.CreateSnapshot(ctx, gameID, *gameState)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")

	// GetLatestSnapshot should fail with context canceled error
	_, err = rds.GetLatestSnapshot(ctx, gameID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}