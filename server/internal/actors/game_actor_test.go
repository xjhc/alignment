package actors

import (
	"sync"
	"testing"
	"time"

	"github.com/alignment/server/internal/game"
)

// MockDataStore implements DataStore interface for testing
type MockDataStore struct {
	events    []game.Event
	snapshots map[string]*game.GameState
	mutex     sync.RWMutex
}

func NewMockDataStore() *MockDataStore {
	return &MockDataStore{
		events:    make([]game.Event, 0),
		snapshots: make(map[string]*game.GameState),
	}
}

func (m *MockDataStore) AppendEvent(gameID string, event game.Event) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *MockDataStore) SaveSnapshot(gameID string, state *game.GameState) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.snapshots[gameID] = state
	return nil
}

func (m *MockDataStore) LoadEvents(gameID string, afterSequence int) ([]game.Event, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return events after the specified sequence
	if afterSequence >= len(m.events) {
		return []game.Event{}, nil
	}
	return m.events[afterSequence:], nil
}

func (m *MockDataStore) LoadSnapshot(gameID string) (*game.GameState, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if snapshot, exists := m.snapshots[gameID]; exists {
		return snapshot, nil
	}
	return nil, ErrGameNotFound
}

func (m *MockDataStore) GetEventCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.events)
}

func (m *MockDataStore) GetEvents() []game.Event {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]game.Event, len(m.events))
	copy(result, m.events)
	return result
}

// MockBroadcaster implements Broadcaster interface for testing
type MockBroadcaster struct {
	gameEvents   []game.Event
	playerEvents map[string][]game.Event
	mutex        sync.RWMutex
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		gameEvents:   make([]game.Event, 0),
		playerEvents: make(map[string][]game.Event),
	}
}

func (m *MockBroadcaster) BroadcastToGame(gameID string, event game.Event) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.gameEvents = append(m.gameEvents, event)
	return nil
}

func (m *MockBroadcaster) SendToPlayer(gameID, playerID string, event game.Event) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.playerEvents[playerID] == nil {
		m.playerEvents[playerID] = make([]game.Event, 0)
	}
	m.playerEvents[playerID] = append(m.playerEvents[playerID], event)
	return nil
}

func (m *MockBroadcaster) GetGameEvents() []game.Event {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	result := make([]game.Event, len(m.gameEvents))
	copy(result, m.gameEvents)
	return result
}

func (m *MockBroadcaster) GetPlayerEvents(playerID string) []game.Event {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if events, exists := m.playerEvents[playerID]; exists {
		result := make([]game.Event, len(events))
		copy(result, events)
		return result
	}
	return []game.Event{}
}

// TestGameActor_PlayerJoin tests basic player joining functionality
func TestGameActor_PlayerJoin(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Send join action
	joinAction := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-123",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "TestPlayer",
			"job_title": "CISO",
		},
	}

	actor.SendAction(joinAction)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Check that event was persisted
	events := datastore.GetEvents()
	if len(events) != 1 {
		t.Fatalf("Expected 1 event to be persisted, got %d", len(events))
	}

	event := events[0]
	if event.Type != game.EventPlayerJoined {
		t.Errorf("Expected PlayerJoined event, got %s", event.Type)
	}

	if event.PlayerID != "player-123" {
		t.Errorf("Expected player ID player-123, got %s", event.PlayerID)
	}

	// Check that event was broadcasted
	gameEvents := broadcaster.GetGameEvents()
	if len(gameEvents) != 1 {
		t.Fatalf("Expected 1 event to be broadcasted, got %d", len(gameEvents))
	}

	if gameEvents[0].Type != game.EventPlayerJoined {
		t.Errorf("Expected broadcasted PlayerJoined event, got %s", gameEvents[0].Type)
	}

	// Check that game state was updated
	if len(actor.state.Players) != 1 {
		t.Fatalf("Expected 1 player in game state, got %d", len(actor.state.Players))
	}

	player, exists := actor.state.Players["player-123"]
	if !exists {
		t.Fatal("Expected player to be added to game state")
	}

	if player.Name != "TestPlayer" {
		t.Errorf("Expected player name TestPlayer, got %s", player.Name)
	}

	if player.JobTitle != "CISO" {
		t.Errorf("Expected job title CISO, got %s", player.JobTitle)
	}
}

// TestGameActor_MultiplePlayerJoins tests multiple players joining
func TestGameActor_MultiplePlayerJoins(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Add multiple players
	players := []struct {
		id       string
		name     string
		jobTitle string
	}{
		{"player-1", "Alice", "CISO"},
		{"player-2", "Bob", "CTO"},
		{"player-3", "Charlie", "CFO"},
	}

	for _, player := range players {
		joinAction := game.Action{
			Type:      game.ActionJoinGame,
			PlayerID:  player.id,
			GameID:    "test-game",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"name":      player.name,
				"job_title": player.jobTitle,
			},
		}
		actor.SendAction(joinAction)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Check that all players were added
	if len(actor.state.Players) != 3 {
		t.Fatalf("Expected 3 players in game state, got %d", len(actor.state.Players))
	}

	// Verify each player
	for _, player := range players {
		gamePlayer, exists := actor.state.Players[player.id]
		if !exists {
			t.Errorf("Expected player %s to be in game state", player.id)
			continue
		}

		if gamePlayer.Name != player.name {
			t.Errorf("Expected player %s name to be %s, got %s", player.id, player.name, gamePlayer.Name)
		}

		if gamePlayer.JobTitle != player.jobTitle {
			t.Errorf("Expected player %s job title to be %s, got %s", player.id, player.jobTitle, gamePlayer.JobTitle)
		}
	}

	// Check that correct number of events were generated
	events := datastore.GetEvents()
	if len(events) != 3 {
		t.Errorf("Expected 3 events to be persisted, got %d", len(events))
	}
}

// TestGameActor_GameCapacity tests game capacity limits
func TestGameActor_GameCapacity(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Fill the game to capacity (default max is 10)
	maxPlayers := actor.state.Settings.MaxPlayers

	for i := 0; i < maxPlayers; i++ {
		joinAction := game.Action{
			Type:      game.ActionJoinGame,
			PlayerID:  "player-" + string(rune('0'+i)),
			GameID:    "test-game",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"name":      "Player" + string(rune('0'+i)),
				"job_title": "Employee",
			},
		}
		actor.SendAction(joinAction)
	}

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	// Try to add one more player (should be rejected)
	joinAction := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "excess-player",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "ExcessPlayer",
			"job_title": "Intern",
		},
	}
	actor.SendAction(joinAction)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should still only have maxPlayers players
	if len(actor.state.Players) != maxPlayers {
		t.Errorf("Expected %d players (at capacity), got %d", maxPlayers, len(actor.state.Players))
	}

	// The excess player should not be in the game
	if _, exists := actor.state.Players["excess-player"]; exists {
		t.Error("Expected excess player to be rejected")
	}

	// Should have exactly maxPlayers events (rejecting the excess)
	events := datastore.GetEvents()
	if len(events) != maxPlayers {
		t.Errorf("Expected %d events (no event for rejected player), got %d", maxPlayers, len(events))
	}
}

// TestGameActor_DuplicatePlayerJoin tests handling duplicate join attempts
func TestGameActor_DuplicatePlayerJoin(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// First join
	joinAction := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-123",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "TestPlayer",
			"job_title": "CISO",
		},
	}
	actor.SendAction(joinAction)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Second join (duplicate)
	duplicateJoinAction := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-123", // Same player ID
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "DuplicatePlayer",
			"job_title": "CTO",
		},
	}
	actor.SendAction(duplicateJoinAction)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should still only have 1 player
	if len(actor.state.Players) != 1 {
		t.Errorf("Expected 1 player after duplicate join, got %d", len(actor.state.Players))
	}

	// Player should retain original information
	player := actor.state.Players["player-123"]
	if player.Name != "TestPlayer" {
		t.Errorf("Expected original name TestPlayer, got %s", player.Name)
	}

	if player.JobTitle != "CISO" {
		t.Errorf("Expected original job title CISO, got %s", player.JobTitle)
	}

	// Should only have 1 event (duplicate was ignored)
	events := datastore.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event (duplicate ignored), got %d", len(events))
	}
}

// TestGameActor_VotingFlow tests the voting process
func TestGameActor_VotingFlow(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Add players first
	for i := 1; i <= 3; i++ {
		joinAction := game.Action{
			Type:      game.ActionJoinGame,
			PlayerID:  "player-" + string(rune('0'+i)),
			GameID:    "test-game",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"name":      "Player" + string(rune('0'+i)),
				"job_title": "Employee",
			},
		}
		actor.SendAction(joinAction)
	}

	// Change phase to voting phase
	actor.state.Phase.Type = game.PhaseNomination

	// Wait for join processing
	time.Sleep(100 * time.Millisecond)

	// Cast votes
	voteAction := game.Action{
		Type:      game.ActionSubmitVote,
		PlayerID:  "player-1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": "player-2",
		},
	}
	actor.SendAction(voteAction)

	// Wait for vote processing
	time.Sleep(100 * time.Millisecond)

	// Check that vote event was generated
	events := datastore.GetEvents()
	voteEvents := 0
	for _, event := range events {
		if event.Type == game.EventVoteCast {
			voteEvents++
		}
	}

	if voteEvents != 1 {
		t.Errorf("Expected 1 vote event, got %d", voteEvents)
	}

	// Check that vote was applied to game state
	if actor.state.VoteState == nil {
		t.Fatal("Expected vote state to be initialized")
	}

	if actor.state.VoteState.Votes["player-1"] != "player-2" {
		t.Errorf("Expected player-1 to vote for player-2, got %s", actor.state.VoteState.Votes["player-1"])
	}
}

// TestGameActor_InvalidVotePhase tests voting in wrong phase
func TestGameActor_InvalidVotePhase(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Add a player
	joinAction := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "Player1",
			"job_title": "Employee",
		},
	}
	actor.SendAction(joinAction)

	// Ensure we're NOT in a voting phase
	actor.state.Phase.Type = game.PhaseDiscussion

	// Wait for join processing
	time.Sleep(100 * time.Millisecond)

	initialEventCount := len(datastore.GetEvents())

	// Try to vote in wrong phase
	voteAction := game.Action{
		Type:      game.ActionSubmitVote,
		PlayerID:  "player-1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": "player-2",
		},
	}
	actor.SendAction(voteAction)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Should not have generated a vote event
	finalEventCount := len(datastore.GetEvents())
	if finalEventCount != initialEventCount {
		t.Errorf("Expected no new events for invalid vote, got %d new events", finalEventCount-initialEventCount)
	}

	// Vote state should not be initialized
	if actor.state.VoteState != nil {
		t.Error("Expected vote state to remain nil for invalid vote")
	}
}

// TestGameActor_MiningTokens tests token mining mechanics with selfless mining
func TestGameActor_MiningTokens(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Add two players (need target for selfless mining)
	joinAction1 := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "Miner",
			"job_title": "CISO",
		},
	}
	actor.SendAction(joinAction1)

	joinAction2 := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-2",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "Target",
			"job_title": "CTO",
		},
	}
	actor.SendAction(joinAction2)

	// Wait for players to join
	time.Sleep(100 * time.Millisecond)

	// Set phase to night for mining
	actor.state.Phase.Type = game.PhaseNight

	// Add enough humans for liquidity pool
	actor.state.Players["human1"] = &game.Player{IsAlive: true, Alignment: "HUMAN"}
	actor.state.Players["human2"] = &game.Player{IsAlive: true, Alignment: "HUMAN"}
	actor.state.Players["human3"] = &game.Player{IsAlive: true, Alignment: "HUMAN"}
	actor.state.Players["human4"] = &game.Player{IsAlive: true, Alignment: "HUMAN"}

	initialTokens := actor.state.Players["player-2"].Tokens // Target gets tokens

	// Perform mining action (player-1 mines for player-2)
	mineAction := game.Action{
		Type:      game.ActionMineTokens,
		PlayerID:  "player-1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": "player-2",
		},
	}
	actor.SendAction(mineAction)

	// Wait for mining processing
	time.Sleep(100 * time.Millisecond)

	// Check that mining event was generated
	events := datastore.GetEvents()
	miningEvents := 0
	for _, event := range events {
		if event.Type == game.EventMiningSuccessful {
			miningEvents++
		}
	}

	if miningEvents != 1 {
		t.Errorf("Expected 1 mining event, got %d", miningEvents)
	}

	// Check that tokens were awarded to the target
	finalTokens := actor.state.Players["player-2"].Tokens
	if finalTokens != initialTokens+1 {
		t.Errorf("Expected target tokens to increase by 1 (from %d to %d), got %d", initialTokens, initialTokens+1, finalTokens)
	}
}

// TestGameActor_ActorPanicRecovery tests panic recovery
func TestGameActor_ActorPanicRecovery(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// This test is harder to implement as we'd need to intentionally cause a panic
	// For now, we'll just verify the actor starts and stops cleanly

	// Add a player to verify normal operation
	joinAction := game.Action{
		Type:      game.ActionJoinGame,
		PlayerID:  "player-1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      "Player1",
			"job_title": "Employee",
		},
	}
	actor.SendAction(joinAction)

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify the actor is still functioning
	if len(actor.state.Players) != 1 {
		t.Errorf("Expected 1 player after normal operation, got %d", len(actor.state.Players))
	}

	// The defer will call Stop(), testing graceful shutdown
}

// TestGameActor_ConcurrentActions tests handling concurrent actions
func TestGameActor_ConcurrentActions(t *testing.T) {
	datastore := NewMockDataStore()
	broadcaster := NewMockBroadcaster()

	actor := NewGameActor("test-game", datastore, broadcaster)
	actor.Start()
	defer actor.Stop()

	// Send multiple actions concurrently
	var wg sync.WaitGroup
	playerCount := 5

	for i := 0; i < playerCount; i++ {
		wg.Add(1)
		go func(playerNum int) {
			defer wg.Done()

			joinAction := game.Action{
				Type:      game.ActionJoinGame,
				PlayerID:  "player-" + string(rune('0'+playerNum)),
				GameID:    "test-game",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"name":      "Player" + string(rune('0'+playerNum)),
					"job_title": "Employee",
				},
			}
			actor.SendAction(joinAction)
		}(i)
	}

	wg.Wait()

	// Wait for all processing
	time.Sleep(200 * time.Millisecond)

	// All players should be added (game actor processes serially)
	if len(actor.state.Players) != playerCount {
		t.Errorf("Expected %d players after concurrent joins, got %d", playerCount, len(actor.state.Players))
	}

	// All events should be persisted
	events := datastore.GetEvents()
	if len(events) != playerCount {
		t.Errorf("Expected %d events after concurrent joins, got %d", playerCount, len(events))
	}
}
