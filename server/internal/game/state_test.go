package game

import (
	"testing"
	"time"
)

func TestNewGameState(t *testing.T) {
	gameID := "test-game-123"
	state := NewGameState(gameID)
	
	if state.ID != gameID {
		t.Errorf("Expected game ID %s, got %s", gameID, state.ID)
	}
	
	if state.Phase.Type != PhaseSetup {
		t.Errorf("Expected phase %s, got %s", PhaseSetup, state.Phase.Type)
	}
	
	if len(state.Players) != 0 {
		t.Errorf("Expected empty players map, got %d players", len(state.Players))
	}
	
	if state.Turn != 0 {
		t.Errorf("Expected turn 0, got %d", state.Turn)
	}
}

func TestGameStateApplyEvent_PlayerJoined(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"
	playerName := "TestPlayer"
	
	event := Event{
		ID:        "event-1",
		Type:      EventPlayerJoined,
		GameID:    "test-game",
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name": playerName,
		},
	}
	
	err := state.ApplyEvent(event)
	if err != nil {
		t.Fatalf("Failed to apply event: %v", err)
	}
	
	player, exists := state.Players[playerID]
	if !exists {
		t.Fatal("Player was not added to game state")
	}
	
	if player.Name != playerName {
		t.Errorf("Expected player name %s, got %s", playerName, player.Name)
	}
	
	if player.Tokens != 1 {
		t.Errorf("Expected player to start with 1 token, got %d", player.Tokens)
	}
	
	if !player.IsActive {
		t.Error("Expected player to be active")
	}
}

func TestGameStateApplyEvent_TokensAwarded(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"
	
	// First add the player
	joinEvent := Event{
		Type:      EventPlayerJoined,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"name": "TestPlayer"},
	}
	state.ApplyEvent(joinEvent)
	
	// Award tokens
	awardEvent := Event{
		Type:      EventTokensAwarded,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"amount": float64(5)},
	}
	
	err := state.ApplyEvent(awardEvent)
	if err != nil {
		t.Fatalf("Failed to apply tokens awarded event: %v", err)
	}
	
	player := state.Players[playerID]
	expectedTokens := 1 + 5 // Starting tokens + awarded tokens
	if player.Tokens != expectedTokens {
		t.Errorf("Expected player to have %d tokens, got %d", expectedTokens, player.Tokens)
	}
}

func TestGameStateApplyEvent_PhaseChanged(t *testing.T) {
	state := NewGameState("test-game")
	
	event := Event{
		Type:      EventPhaseChanged,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"phase_type": string(PhaseDay),
			"duration":   float64(120), // 2 minutes
		},
	}
	
	err := state.ApplyEvent(event)
	if err != nil {
		t.Fatalf("Failed to apply phase change event: %v", err)
	}
	
	if state.Phase.Type != PhaseDay {
		t.Errorf("Expected phase %s, got %s", PhaseDay, state.Phase.Type)
	}
	
	expectedDuration := 2 * time.Minute
	if state.Phase.Duration != expectedDuration {
		t.Errorf("Expected phase duration %v, got %v", expectedDuration, state.Phase.Duration)
	}
	
	if state.Turn != 1 {
		t.Errorf("Expected turn to increment to 1, got %d", state.Turn)
	}
}