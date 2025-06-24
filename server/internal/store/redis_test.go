package store

import (
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
	err := rds.AppendEvent(gameID, event1)
	assert.NoError(t, err)
	err = rds.AppendEvent(gameID, event2)
	assert.NoError(t, err)

	// Load events
	loadedEvents, err := rds.LoadEvents(gameID, 0)
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
	err := rds.CreateSnapshot(gameID, gameState)
	assert.NoError(t, err)

	// Load snapshot
	loadedState, err := rds.LoadSnapshot(gameID)
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
	count, err := rds.GetEventCount(gameID)
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

	err = rds.AppendEvent(gameID, event)
	assert.NoError(t, err)

	// Should now have 1 event
	count, err = rds.GetEventCount(gameID)
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

	err := rds.AppendEvent(gameID, event)
	assert.NoError(t, err)
	err = rds.CreateSnapshot(gameID, gameState)
	assert.NoError(t, err)

	// Verify data exists
	count, err := rds.GetEventCount(gameID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	_, err = rds.LoadSnapshot(gameID)
	assert.NoError(t, err)

	// Delete the game
	err = rds.DeleteGame(gameID)
	assert.NoError(t, err)

	// Verify data is gone
	count, err = rds.GetEventCount(gameID)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	_, err = rds.LoadSnapshot(gameID)
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
	err := rds.CreateSnapshot(gameID, gameState)
	assert.NoError(t, err)

	// Get metadata
	metadata, err := rds.GetGameMetadata(gameID)
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	// Verify metadata fields
	assert.Equal(t, "2", metadata["snapshot_day"])
	assert.Equal(t, string(core.PhaseDiscussion), metadata["phase"])
	assert.Equal(t, "2", metadata["player_count"])
}