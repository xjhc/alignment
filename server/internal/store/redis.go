package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/redis/go-redis/v9"
)

// RedisDataStore implements DataStore interface using Redis
type RedisDataStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisDataStore creates a new Redis data store
func NewRedisDataStore(addr, password string, db int) (*RedisDataStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("Connected to Redis at %s", addr)

	return &RedisDataStore{
		client: client,
		ctx:    ctx,
	}, nil
}

// AppendEvent appends an event to the game's Redis Stream (WAL)
func (rds *RedisDataStore) AppendEvent(gameID string, event core.Event) error {
	streamKey := fmt.Sprintf("game:%s:events", gameID)

	// Serialize event payload
	payloadJSON, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event payload: %w", err)
	}

	// Prepare stream fields
	fields := map[string]interface{}{
		"event_id":  event.ID,
		"type":      string(event.Type),
		"game_id":   event.GameID,
		"player_id": event.PlayerID,
		"timestamp": event.Timestamp.Unix(),
		"payload":   string(payloadJSON),
	}

	// Add to stream
	_, err = rds.client.XAdd(rds.ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: fields,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to append event to stream: %w", err)
	}

	// Set TTL on the stream (events expire after 7 days)
	rds.client.Expire(rds.ctx, streamKey, 7*24*time.Hour)

	return nil
}

// SaveSnapshot saves a complete game state snapshot
func (rds *RedisDataStore) SaveSnapshot(gameID string, state *core.GameState) error {
	snapshotKey := fmt.Sprintf("game:%s:snapshot", gameID)

	// Serialize game state
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal game state: %w", err)
	}

	// Save snapshot
	err = rds.client.Set(rds.ctx, snapshotKey, stateJSON, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save snapshot: %w", err)
	}

	// Update metadata
	metaKey := fmt.Sprintf("game:%s:meta", gameID)
	metadata := map[string]interface{}{
		"last_snapshot": time.Now().Unix(),
		"snapshot_day":  state.DayNumber,
		"phase":         string(state.Phase.Type),
		"player_count":  len(state.Players),
		"created_at":    state.CreatedAt.Unix(),
		"updated_at":    state.UpdatedAt.Unix(),
	}

	err = rds.client.HMSet(rds.ctx, metaKey, metadata).Err()
	if err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	rds.client.Expire(rds.ctx, metaKey, 7*24*time.Hour)

	log.Printf("Saved snapshot for game %s", gameID)
	return nil
}

// LoadEvents loads events from Redis Stream after a specific sequence
func (rds *RedisDataStore) LoadEvents(gameID string, afterSequence int) ([]core.Event, error) {
	streamKey := fmt.Sprintf("game:%s:events", gameID)

	// Determine start position
	start := "0"
	if afterSequence > 0 {
		start = fmt.Sprintf("%d", afterSequence+1)
	}

	// Read from stream
	streams, err := rds.client.XRead(rds.ctx, &redis.XReadArgs{
		Streams: []string{streamKey, start},
		Count:   1000, // Max events to read at once
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []core.Event{}, nil // No events found
		}
		return nil, fmt.Errorf("failed to read events from stream: %w", err)
	}

	var events []core.Event

	for _, stream := range streams {
		for _, message := range stream.Messages {
			event, err := rds.parseEventFromMessage(message)
			if err != nil {
				log.Printf("Failed to parse event %s: %v", message.ID, err)
				continue
			}
			events = append(events, event)
		}
	}

	return events, nil
}

// LoadSnapshot loads the latest game state snapshot
func (rds *RedisDataStore) LoadSnapshot(gameID string) (*core.GameState, error) {
	snapshotKey := fmt.Sprintf("game:%s:snapshot", gameID)

	// Get snapshot data
	stateJSON, err := rds.client.Get(rds.ctx, snapshotKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("no snapshot found for game %s", gameID)
		}
		return nil, fmt.Errorf("failed to load snapshot: %w", err)
	}

	// Deserialize game state
	var state core.GameState
	err = json.Unmarshal([]byte(stateJSON), &state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal game state: %w", err)
	}

	log.Printf("Loaded snapshot for game %s", gameID)
	return &state, nil
}

// GetGameMetadata retrieves game metadata
func (rds *RedisDataStore) GetGameMetadata(gameID string) (map[string]string, error) {
	metaKey := fmt.Sprintf("game:%s:meta", gameID)

	metadata, err := rds.client.HGetAll(rds.ctx, metaKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get game metadata: %w", err)
	}

	return metadata, nil
}

// DeleteGame removes all game data from Redis
func (rds *RedisDataStore) DeleteGame(gameID string) error {
	keys := []string{
		fmt.Sprintf("game:%s:events", gameID),
		fmt.Sprintf("game:%s:snapshot", gameID),
		fmt.Sprintf("game:%s:meta", gameID),
	}

	err := rds.client.Del(rds.ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete game data: %w", err)
	}

	log.Printf("Deleted all data for game %s", gameID)
	return nil
}

// ListActiveGames returns IDs of all games with recent activity
func (rds *RedisDataStore) ListActiveGames() ([]string, error) {
	pattern := "game:*:meta"

	keys, err := rds.client.Keys(rds.ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list games: %w", err)
	}

	var gameIDs []string
	for _, key := range keys {
		// Extract game ID from key format "game:{id}:meta"
		if len(key) > 12 { // len("game:") + len(":meta") = 10
			gameID := key[5 : len(key)-5] // Remove "game:" prefix and ":meta" suffix
			gameIDs = append(gameIDs, gameID)
		}
	}

	return gameIDs, nil
}

// GetEventCount returns the number of events in a game's stream
func (rds *RedisDataStore) GetEventCount(gameID string) (int64, error) {
	streamKey := fmt.Sprintf("game:%s:events", gameID)

	count, err := rds.client.XLen(rds.ctx, streamKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get event count: %w", err)
	}

	return count, nil
}

// Close closes the Redis connection
func (rds *RedisDataStore) Close() error {
	return rds.client.Close()
}

// parseEventFromMessage converts a Redis stream message to a game Event
func (rds *RedisDataStore) parseEventFromMessage(message redis.XMessage) (core.Event, error) {
	var event core.Event

	// Extract fields from message
	eventID, ok := message.Values["event_id"].(string)
	if !ok {
		return event, fmt.Errorf("missing event_id")
	}

	eventType, ok := message.Values["type"].(string)
	if !ok {
		return event, fmt.Errorf("missing type")
	}

	gameID, ok := message.Values["game_id"].(string)
	if !ok {
		return event, fmt.Errorf("missing game_id")
	}

	playerID, _ := message.Values["player_id"].(string) // Optional field

	timestampStr, ok := message.Values["timestamp"].(string)
	if !ok {
		return event, fmt.Errorf("missing timestamp")
	}

	payloadStr, ok := message.Values["payload"].(string)
	if !ok {
		return event, fmt.Errorf("missing payload")
	}

	// Parse timestamp
	timestampUnix, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return event, fmt.Errorf("invalid timestamp: %w", err)
	}

	// Parse payload
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(payloadStr), &payload)
	if err != nil {
		return event, fmt.Errorf("invalid payload JSON: %w", err)
	}

	// Construct event
	event = core.Event{
		ID:        eventID,
		Type:      core.EventType(eventType),
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Unix(timestampUnix, 0),
		Payload:   payload,
	}

	return event, nil
}

// GetGameStats returns statistics about the game
func (rds *RedisDataStore) GetGameStats(gameID string) (map[string]interface{}, error) {
	metadata, err := rds.GetGameMetadata(gameID)
	if err != nil {
		return nil, err
	}

	eventCount, err := rds.GetEventCount(gameID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"event_count": eventCount,
		"metadata":    metadata,
	}

	return stats, nil
}
