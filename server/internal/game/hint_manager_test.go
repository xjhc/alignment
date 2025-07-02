package game

import (
	"testing"

	"github.com/xjhc/alignment/core"
)

func TestHintManager_CheckForHints(t *testing.T) {
	// Create a test game state with human players
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"human1": {
				ID:          "human1",
				Name:        "Alice",
				IsAlive:     true,
				ControlType: "HUMAN",
			},
			"human2": {
				ID:          "human2",
				Name:        "Bob",
				IsAlive:     true,
				ControlType: "HUMAN",
			},
			"ai1": {
				ID:          "ai1",
				Name:        "Charlie",
				IsAlive:     true,
				ControlType: "AI",
			},
			"dead1": {
				ID:          "dead1",
				Name:        "Dave",
				IsAlive:     false,
				ControlType: "HUMAN",
			},
		},
		Phase: core.Phase{
			Type: core.PhaseNomination,
		},
	}

	hm := NewHintManager(gameState)

	// Test that hints are generated for nomination phase
	hints := hm.CheckForHints(core.PhaseNomination)

	// Should generate hints for 2 living human players
	if len(hints) != 2 {
		t.Errorf("Expected 2 hints, got %d", len(hints))
	}

	// Check that all hints are private notifications
	for _, hint := range hints {
		if hint.Type != core.EventPrivateNotification {
			t.Errorf("Expected EventPrivateNotification, got %s", hint.Type)
		}
		
		if hint.GameID != "test-game" {
			t.Errorf("Expected game ID 'test-game', got %s", hint.GameID)
		}
		
		// Should be targeted to a specific player
		if hint.PlayerID == "" {
			t.Errorf("Expected hint to have PlayerID, got empty string")
		}
		
		// Check payload structure
		if hint.Payload["type"] != "LOEBMATE_HINT" {
			t.Errorf("Expected type LOEBMATE_HINT, got %v", hint.Payload["type"])
		}
		
		if hint.Payload["phase"] != string(core.PhaseNomination) {
			t.Errorf("Expected phase %s, got %v", core.PhaseNomination, hint.Payload["phase"])
		}
		
		if hint.Payload["sender"] != "Loebmate" {
			t.Errorf("Expected sender Loebmate, got %v", hint.Payload["sender"])
		}
		
		// Should have a message
		message, ok := hint.Payload["message"].(string)
		if !ok || message == "" {
			t.Errorf("Expected non-empty message, got %v", hint.Payload["message"])
		}
	}
}

func TestHintManager_GetPhaseHint(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
	}
	hm := NewHintManager(gameState)

	testCases := []struct {
		phase    core.PhaseType
		expected bool // whether hint should be non-empty
	}{
		{core.PhaseSitrep, true},
		{core.PhasePulseCheck, true},
		{core.PhaseDiscussion, true},
		{core.PhaseNomination, true},
		{core.PhaseVerdict, true},
		{core.PhaseNight, true},
		{core.PhaseLobby, false}, // No hint for lobby phase in current implementation
		{core.PhaseGameOver, false}, // No hint for game over
	}

	for _, tc := range testCases {
		hint := hm.GetPhaseHint(tc.phase)
		isEmpty := hint == ""
		
		if tc.expected && isEmpty {
			t.Errorf("Expected non-empty hint for phase %s, got empty", tc.phase)
		}
		
		if !tc.expected && !isEmpty {
			t.Errorf("Expected empty hint for phase %s, got: %s", tc.phase, hint)
		}
	}
}

func TestHintManager_ShouldShowHint(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"human1": {
				ID:          "human1",
				ControlType: "HUMAN",
				IsAlive:     true,
			},
			"human_disabled": {
				ID:                   "human_disabled",
				ControlType:          "HUMAN",
				IsAlive:              true,
				DisableLoebmateHints: true,
			},
			"human_seen": {
				ID:          "human_seen",
				ControlType: "HUMAN",
				IsAlive:     true,
				SeenHints:   map[string]bool{"NOMINATION": true},
			},
			"ai1": {
				ID:          "ai1",
				ControlType: "AI",
				IsAlive:     true,
			},
			"dead1": {
				ID:          "dead1",
				ControlType: "HUMAN",
				IsAlive:     false,
			},
		},
	}

	hm := NewHintManager(gameState)

	// Living human should get hints
	if !hm.shouldShowHint("human1", core.PhaseNomination) {
		t.Error("Expected living human to receive hints")
	}

	// Human with disabled hints should not get hints
	if hm.shouldShowHint("human_disabled", core.PhaseNomination) {
		t.Error("Expected human with disabled hints to not receive hints")
	}

	// Human who has seen the hint should not get it again
	if hm.shouldShowHint("human_seen", core.PhaseNomination) {
		t.Error("Expected human who has seen hint to not receive it again")
	}

	// Human who has seen different hint should still get this one
	if !hm.shouldShowHint("human_seen", core.PhaseDiscussion) {
		t.Error("Expected human to receive hints for unseen phases")
	}

	// AI should not get hints
	if hm.shouldShowHint("ai1", core.PhaseNomination) {
		t.Error("Expected AI to not receive hints")
	}

	// Dead human should not get hints
	if hm.shouldShowHint("dead1", core.PhaseNomination) {
		t.Error("Expected dead human to not receive hints")
	}

	// Non-existent player should not get hints
	if hm.shouldShowHint("nonexistent", core.PhaseNomination) {
		t.Error("Expected non-existent player to not receive hints")
	}
}

func TestHintManager_MarkHintAsSeen(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"human1": {
				ID:          "human1",
				ControlType: "HUMAN",
				IsAlive:     true,
			},
		},
	}

	hm := NewHintManager(gameState)
	player := gameState.Players["human1"]

	// Initially should have no seen hints
	if player.SeenHints != nil && len(player.SeenHints) > 0 {
		t.Error("Expected player to start with no seen hints")
	}

	// Mark a hint as seen
	hm.markHintAsSeen("human1", core.PhaseNomination)

	// Check that the hint was marked as seen
	if player.SeenHints == nil {
		t.Error("Expected SeenHints to be initialized")
	}

	if !player.SeenHints["NOMINATION"] {
		t.Error("Expected NOMINATION hint to be marked as seen")
	}

	// Mark another hint as seen
	hm.markHintAsSeen("human1", core.PhaseDiscussion)

	// Check that both hints are now seen
	if !player.SeenHints["NOMINATION"] || !player.SeenHints["DISCUSSION"] {
		t.Error("Expected both NOMINATION and DISCUSSION hints to be marked as seen")
	}
}

func TestHintManager_ProgressiveDisclosure(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"new_player": {
				ID:          "new_player",
				ControlType: "HUMAN",
				IsAlive:     true,
			},
		},
	}

	hm := NewHintManager(gameState)

	// First time seeing nomination phase - should get hint
	hints := hm.CheckForHints(core.PhaseNomination)
	if len(hints) != 1 {
		t.Errorf("Expected 1 hint for new player, got %d", len(hints))
	}

	// Second time seeing nomination phase - should not get hint
	hints = hm.CheckForHints(core.PhaseNomination)
	if len(hints) != 0 {
		t.Errorf("Expected 0 hints for repeated phase, got %d", len(hints))
	}

	// Different phase - should get hint
	hints = hm.CheckForHints(core.PhaseDiscussion)
	if len(hints) != 1 {
		t.Errorf("Expected 1 hint for new phase, got %d", len(hints))
	}
}