package actors

import (
	"context"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// Helper function to create a test GameActor with initialized players
func createTestGameActor(t *testing.T) *GameActor {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create test players
	players := map[string]*core.Player{
		"player1": {
			ID:       "player1",
			Name:     "Alice",
			JobTitle: "Employee",
			IsAlive:  true,
		},
		"player2": {
			ID:       "player2",
			Name:     "Bob",
			JobTitle: "Employee",
			IsAlive:  true,
		},
		"player3": {
			ID:       "player3",
			Name:     "Charlie",
			JobTitle: "Employee",
			IsAlive:  true,
		},
	}

	actor := NewGameActor(ctx, cancel, "test-game", players)

	actor.Start()
	t.Cleanup(actor.Stop)

	// Initialize the game via an action
	initAction := core.Action{Type: "INITIALIZE_GAME"}
	responseChan := actor.PostAction(initAction)
	<-responseChan // Wait for initialization to complete

	return actor
}

func TestGameActor_ProcessAction_LeaveGame(t *testing.T) {
	actor := createTestGameActor(t)

	// Test valid leave game action
	leaveAction := core.Action{
		Type:      core.ActionLeaveGame,
		PlayerID:  "player1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	responseChan := actor.PostAction(leaveAction)
	result := <-responseChan
	if result.Error != nil {
		t.Errorf("Expected no error for valid leave action, got: %v", result.Error)
	}

	if len(result.Events) != 1 { // Only the PLAYER_LEFT event
		t.Errorf("Expected 1 PLAYER_LEFT event, got %d", len(result.Events))
	}
	
	// Verify the event type is correct
	if result.Events[0].Type != core.EventPlayerLeft {
		t.Errorf("Expected event type to be PLAYER_LEFT, got %s", result.Events[0].Type)
	}
}

func TestGameActor_ProcessAction_InvalidPlayer(t *testing.T) {
	actor := createTestGameActor(t)

	// Test action from non-existent player
	invalidAction := core.Action{
		Type:      core.ActionLeaveGame,
		PlayerID:  "non-existent-player",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	responseChan := actor.PostAction(invalidAction)
	result := <-responseChan
	if result.Error == nil {
		t.Error("Expected error for action from non-existent player")
	}
}

func TestGameActor_ProcessAction_PhaseTransition(t *testing.T) {
	actor := createTestGameActor(t)

	// Test phase transition action
	phaseAction := core.Action{
		Type:      core.ActionType("PHASE_TRANSITION"),
		PlayerID:  "SYSTEM",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"next_phase": string(core.PhaseDiscussion),
		},
	}

	responseChan := actor.PostAction(phaseAction)
	result := <-responseChan
	if result.Error != nil {
		t.Errorf("Expected no error for valid phase transition, got: %v", result.Error)
	}

	// Verify game state was updated
	gameState := actor.GetGameState()
	if gameState.Phase.Type != core.PhaseDiscussion {
		t.Errorf("Expected game phase to be %s, got %s", core.PhaseDiscussion, gameState.Phase.Type)
	}
}

func TestGameActor_GetGameState(t *testing.T) {
	actor := createTestGameActor(t)

	gameState := actor.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should not be nil")
	}

	if gameState.ID != "test-game" {
		t.Errorf("Expected game ID 'test-game', got '%s'", gameState.ID)
	}

	if len(gameState.Players) != 3 {
		t.Errorf("Expected 3 players, got %d", len(gameState.Players))
	}

	// Verify players were initialized correctly
	for _, player := range gameState.Players {
		if !player.IsAlive {
			t.Errorf("Player %s should be alive initially", player.ID)
		}
		if player.Role == nil {
			t.Errorf("Player %s should have a role assigned", player.ID)
		}
		if player.Alignment == "" {
			t.Errorf("Player %s should have an alignment assigned", player.ID)
		}
	}

	if gameState.DayNumber != 1 {
		t.Errorf("Expected day number 1, got %d", gameState.DayNumber)
	}
}

func TestGameActor_CreatePlayerStateUpdateEvent(t *testing.T) {
	actor := createTestGameActor(t)

	snapshot := actor.CreatePlayerStateUpdateEvent("player1")

	if snapshot.Type != "GAME_STATE_UPDATE" {
		t.Errorf("Expected event type GAME_STATE_UPDATE, got %s", snapshot.Type)
	}

	if snapshot.PlayerID != "player1" {
		t.Errorf("Expected player ID 'player1', got '%s'", snapshot.PlayerID)
	}

	payload := snapshot.Payload
	gameState, ok := payload["game_state"].(*core.GameState)
	if !ok {
		t.Fatal("Expected 'game_state' in payload of type *core.GameState")
	}

	// Player1 should have full information
	if player1, exists := gameState.Players["player1"]; !exists || player1.Role == nil {
		t.Error("Player1 should exist and have role info in their own snapshot")
	}
	// Player2 should have stripped information
	if player2, exists := gameState.Players["player2"]; !exists || player2.Role != nil {
		t.Error("Player2 should exist but not have role info in Player1's snapshot")
	}
}
