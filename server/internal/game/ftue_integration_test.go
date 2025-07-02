package game

import (
	"testing"

	"github.com/xjhc/alignment/core"
)

// TestFTUE_CompleteUserJourney demonstrates the full First Time User Experience
func TestFTUE_CompleteUserJourney(t *testing.T) {
	// Create a game state with a new player
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"new_user": {
				ID:          "new_user",
				ControlType: "HUMAN",
				IsAlive:     true,
				// No SeenHints or DisableLoebmateHints - completely new user
			},
			"veteran_user": {
				ID:                   "veteran_user",
				ControlType:          "HUMAN",
				IsAlive:              true,
				SeenHints:            map[string]bool{"NOMINATION": true, "DISCUSSION": true},
				DisableLoebmateHints: false,
			},
			"disabled_user": {
				ID:                   "disabled_user",
				ControlType:          "HUMAN",
				IsAlive:              true,
				DisableLoebmateHints: true,
			},
		},
		Phase: core.Phase{
			Type: core.PhaseNomination,
		},
	}

	hm := NewHintManager(gameState)

	t.Run("New user gets hints on first encounter", func(t *testing.T) {
		// New user encounters nomination phase for the first time
		hints := hm.CheckForHints(core.PhaseNomination)
		
		// Should get exactly 1 hint for the new user only
		if len(hints) != 1 {
			t.Errorf("Expected 1 hint for new user, got %d", len(hints))
		}
		
		// Verify the hint is for the new user
		if hints[0].PlayerID != "new_user" {
			t.Errorf("Expected hint for new_user, got %s", hints[0].PlayerID)
		}
		
		// Verify hint content and metadata
		if hints[0].Type != core.EventPrivateNotification {
			t.Errorf("Expected EventPrivateNotification, got %s", hints[0].Type)
		}
		
		if hints[0].Payload["type"] != "LOEBMATE_HINT" {
			t.Errorf("Expected LOEBMATE_HINT type, got %v", hints[0].Payload["type"])
		}
		
		if hints[0].Payload["phase"] != "NOMINATION" {
			t.Errorf("Expected NOMINATION phase, got %v", hints[0].Payload["phase"])
		}
		
		// Verify hint was marked as seen
		if !gameState.Players["new_user"].SeenHints["NOMINATION"] {
			t.Error("Expected hint to be marked as seen after sending")
		}
	})

	t.Run("Progressive disclosure - no repeated hints", func(t *testing.T) {
		// Same phase again - should not get hints
		hints := hm.CheckForHints(core.PhaseNomination)
		
		if len(hints) != 0 {
			t.Errorf("Expected 0 hints for repeated phase, got %d", len(hints))
		}
	})

	t.Run("New phase triggers new hint", func(t *testing.T) {
		// Move to discussion phase
		hints := hm.CheckForHints(core.PhaseDiscussion)
		
		// Should get hint for discussion phase
		if len(hints) != 1 {
			t.Errorf("Expected 1 hint for new phase, got %d", len(hints))
		}
		
		if hints[0].Payload["phase"] != "DISCUSSION" {
			t.Errorf("Expected DISCUSSION phase hint, got %v", hints[0].Payload["phase"])
		}
	})

	t.Run("Veteran user with some seen hints gets selective hints", func(t *testing.T) {
		// Veteran has seen NOMINATION and DISCUSSION, but not VERDICT
		hints := hm.CheckForHints(core.PhaseVerdict)
		
		// Should get hint for verdict since they haven't seen it
		if len(hints) != 2 { // new_user + veteran_user for VERDICT
			t.Errorf("Expected 2 hints for VERDICT phase, got %d", len(hints))
		}
		
		// Verify one is for veteran
		foundVeteranHint := false
		for _, hint := range hints {
			if hint.PlayerID == "veteran_user" {
				foundVeteranHint = true
				break
			}
		}
		if !foundVeteranHint {
			t.Error("Expected veteran user to receive hint for unseen phase")
		}
		
		// Try a phase veteran has seen
		hints = hm.CheckForHints(core.PhaseNomination)
		
		// Should only get hint for new_user (who we already tested above has it marked as seen)
		// Actually, new_user now has it marked as seen, so should be 0 hints
		if len(hints) != 0 {
			t.Errorf("Expected 0 hints for phase both users have seen, got %d", len(hints))
		}
	})

	t.Run("Disabled user gets no hints", func(t *testing.T) {
		// Test with a phase disabled_user hasn't seen
		hints := hm.CheckForHints(core.PhaseNight)
		
		// Should get hints for new_user and veteran_user, but not disabled_user
		expectedHints := 2 // new_user + veteran_user
		if len(hints) != expectedHints {
			t.Errorf("Expected %d hints (excluding disabled user), got %d", expectedHints, len(hints))
		}
		
		// Verify disabled user is not in the hints
		for _, hint := range hints {
			if hint.PlayerID == "disabled_user" {
				t.Error("Disabled user should not receive hints")
			}
		}
	})

	t.Run("AI and dead players get no hints", func(t *testing.T) {
		// Add AI and dead players
		gameState.Players["ai_player"] = &core.Player{
			ID:          "ai_player",
			ControlType: "AI",
			IsAlive:     true,
		}
		gameState.Players["dead_player"] = &core.Player{
			ID:          "dead_player",
			ControlType: "HUMAN",
			IsAlive:     false,
		}
		
		hints := hm.CheckForHints(core.PhaseSitrep)
		
		// Should only get hints for living humans with hints enabled
		for _, hint := range hints {
			player := gameState.Players[hint.PlayerID]
			if player.ControlType != "HUMAN" {
				t.Errorf("Non-human player %s received hint", hint.PlayerID)
			}
			if !player.IsAlive {
				t.Errorf("Dead player %s received hint", hint.PlayerID)
			}
			if player.DisableLoebmateHints {
				t.Errorf("Player with disabled hints %s received hint", hint.PlayerID)
			}
		}
	})
}

// TestFTUE_HelpCommand demonstrates the /help command functionality
func TestFTUE_HelpCommand(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"user1": {
				ID:          "user1",
				ControlType: "HUMAN",
				IsAlive:     true,
			},
		},
		Phase: core.Phase{
			Type: core.PhaseNomination,
		},
	}

	hm := NewHintManager(gameState)

	// Test getting hint for current phase (this is the method we have in HintManager)
	hintText := hm.GetPhaseHint(core.PhaseNomination)

	// Verify hint text contains expected elements
	if hintText == "" {
		t.Error("Expected non-empty hint text for nomination phase")
	}

	// Test hints for different phases
	phases := []core.PhaseType{
		core.PhaseSitrep,
		core.PhasePulseCheck,
		core.PhaseDiscussion,
		core.PhaseNomination,
		core.PhaseVerdict,
		core.PhaseNight,
	}

	for _, phase := range phases {
		hint := hm.GetPhaseHint(phase)
		if hint == "" {
			t.Errorf("Expected hint text for phase %s", phase)
		}
	}
}