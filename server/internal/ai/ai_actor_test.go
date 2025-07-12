package ai

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestAIActor_Creation(t *testing.T) {
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}
	defer actor.Stop()

	if actor.GetPlayerID() != "ai-1" {
		t.Errorf("Expected player ID ai-1, got %s", actor.GetPlayerID())
	}

	if actor.GetStatus() != "active" {
		t.Errorf("Expected status active, got %s", actor.GetStatus())
	}
}

func TestAIActor_StartStop(t *testing.T) {
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}

	// Start the actor
	actor.Start()

	// Verify it's running
	if actor.GetStatus() != "active" {
		t.Error("Actor should be active after start")
	}

	// Stop the actor
	actor.Stop()

	// Give it a moment to shut down
	time.Sleep(100 * time.Millisecond)

	// Verify it's stopped
	status := actor.GetStatus()
	if status == "active" {
		t.Error("Actor should not be active after stop")
	}
}

func TestAIActor_TriggerAction(t *testing.T) {
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}
	defer actor.Stop()

	actor.Start()

	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"ai-1": {
				ID:                "ai-1",
				IsAlive:           true,
				Alignment:         "AI",
				ControlType:       "AI",
				ProjectMilestones: 1,
			},
			"human-1": {
				ID:          "human-1",
				IsAlive:     true,
				Alignment:   "HUMAN",
				ControlType: "HUMAN",
				Tokens:      3,
			},
		},
	}

	// Trigger a night action
	resultChan := actor.TriggerAction(TriggerNightStart, gameState, nil)

	// Wait for result with timeout
	select {
	case action := <-resultChan:
		// AI may return nil if no good action is available
		if action != nil {
			if action.PlayerID != "ai-1" {
				t.Errorf("Expected action from ai-1, got %s", action.PlayerID)
			}
			if action.GameID != "test-game" {
				t.Errorf("Expected game ID test-game, got %s", action.GameID)
			}
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for AI action")
	}
}

func TestAIActor_PhaseBasedTriggers(t *testing.T) {
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}
	defer actor.Stop()

	actor.Start()

	tests := []struct {
		name      string
		trigger   AITrigger
		phase     core.PhaseType
		expectAction bool
	}{
		{
			name:      "Night phase may produce action",
			trigger:   TriggerNightStart,
			phase:     core.PhaseNight,
			expectAction: false, // May or may not produce action depending on game state
		},
		{
			name:      "Nomination phase should produce vote",
			trigger:   TriggerVotingStart,
			phase:     core.PhaseNomination,
			expectAction: true,
		},
		{
			name:      "Chat trigger may not produce action",
			trigger:   TriggerRandomChat,
			phase:     core.PhaseDiscussion,
			expectAction: false, // Depends on AI decision
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := &core.GameState{
				ID:        "test-game",
				DayNumber: 2,
				Phase:     core.Phase{Type: tt.phase},
				Players: map[string]*core.Player{
					"ai-1": {
						ID:        "ai-1",
						IsAlive:   true,
						Alignment: "AI",
						Tokens:    3,
					},
					"human-1": {
						ID:        "human-1",
						IsAlive:   true,
						Alignment: "HUMAN",
						Tokens:    4,
					},
				},
			}

			resultChan := actor.TriggerAction(tt.trigger, gameState, nil)

			select {
			case action := <-resultChan:
				if tt.expectAction && action == nil {
					t.Error("Expected action but got nil")
				}
				if !tt.expectAction && action != nil {
					t.Logf("Got unexpected action: %s (this may be normal)", action.Type)
				}
			case <-time.After(2 * time.Second):
				if tt.expectAction {
					t.Error("Timeout waiting for expected action")
				}
			}
		})
	}
}

func TestAIActor_ConcurrentTriggers(t *testing.T) {
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}
	defer actor.Stop()

	actor.Start()

	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"ai-1": {
				ID:        "ai-1",
				IsAlive:   true,
				Alignment: "AI",
			},
		},
	}

	// Send multiple triggers concurrently
	numTriggers := 5
	resultChans := make([]chan *core.Action, numTriggers)

	for i := 0; i < numTriggers; i++ {
		resultChans[i] = actor.TriggerAction(TriggerNightStart, gameState, nil)
	}

	// Collect results
	responses := 0
	timeout := time.After(5 * time.Second)

	for i := 0; i < numTriggers; i++ {
		select {
		case action := <-resultChans[i]:
			if action != nil {
				responses++
			}
		case <-timeout:
			t.Logf("Timeout on trigger %d", i)
		}
	}

	if responses == 0 {
		t.Error("Expected at least one response from concurrent triggers")
	}

	t.Logf("Got %d responses from %d triggers", responses, numTriggers)
}

func TestAIActor_PanicRecovery(t *testing.T) {
	// This test verifies that the AI actor handles panics gracefully
	// We can't easily induce a panic in the current implementation,
	// but we can test the timeout mechanism
	
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}
	defer actor.Stop()

	actor.Start()

	// Set a very short timeout to test timeout handling
	actor.responseTimeout = 1 * time.Millisecond

	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"ai-1": {
				ID:        "ai-1",
				IsAlive:   true,
				Alignment: "AI",
			},
		},
	}

	resultChan := actor.TriggerAction(TriggerNightStart, gameState, nil)

	// Should get a response (likely nil due to timeout) without panic
	select {
	case action := <-resultChan:
		// This is expected - either an action or nil due to timeout
		t.Logf("Got action: %v", action)
	case <-time.After(1 * time.Second):
		t.Error("Should have received response within timeout")
	}

	// Actor should still be responsive after timeout
	if actor.GetStatus() != "active" {
		t.Error("Actor should still be active after timeout")
	}
}

func TestAIActor_UpdateGameState(t *testing.T) {
	persona := GetShadowPersona()
	actor, err := NewAIActor("ai-1", "test-game", persona)
	if err != nil {
		t.Fatalf("Failed to create AIActor: %v", err)
	}
	defer actor.Stop()

	gameState1 := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
	}

	gameState2 := &core.GameState{
		ID:        "test-game",
		DayNumber: 2,
	}

	// Update game state (should not error)
	actor.UpdateGameState(gameState1)
	actor.UpdateGameState(gameState2)

	// This test mainly verifies the method doesn't panic
	// More sophisticated state tracking could be added later
}

func TestAITriggerTypes(t *testing.T) {
	// Test that all trigger types are defined
	triggers := []AITrigger{
		TriggerPhaseChange,
		TriggerPlayerSpoke,
		TriggerVotingStart,
		TriggerNightStart,
		TriggerRandomChat,
		TriggerGameEnd,
	}

	for _, trigger := range triggers {
		if string(trigger) == "" {
			t.Errorf("Trigger %v should have non-empty string value", trigger)
		}
	}
}

// Benchmark tests
func BenchmarkAIActor_TriggerAction(b *testing.B) {
	persona := GetShadowPersona()
	actor, _ := NewAIActor("ai-1", "test-game", persona)
	defer actor.Stop()

	actor.Start()

	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"ai-1": {
				ID:        "ai-1",
				IsAlive:   true,
				Alignment: "AI",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resultChan := actor.TriggerAction(TriggerNightStart, gameState, nil)
		// Wait for result to complete the benchmark
		<-resultChan
	}
}