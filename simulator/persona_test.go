package simulator

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// TestCautiousHuman_ShouldNominate tests the nomination logic for cautious humans
func TestCautiousHuman_ShouldNominate(t *testing.T) {
	// Create a CautiousHuman persona
	cautious := NewCautiousHuman(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:        "test-game",
		DayNumber: 3,
		Players: map[string]*core.Player{
			"human1": {ID: "human1", IsAlive: true, Tokens: 2, ProjectMilestones: 4},
			"suspicious": {ID: "suspicious", IsAlive: true, Tokens: 8, ProjectMilestones: 1}, // Very suspicious
			"human2": {ID: "human2", IsAlive: true, Tokens: 3, ProjectMilestones: 3},
		},
	}

	// Should nominate the suspicious player
	shouldNominate, target := cautious.ShouldNominate(gameState, "human1")
	if !shouldNominate {
		t.Error("Expected to nominate")
	}
	if target != "suspicious" {
		t.Errorf("Expected to nominate 'suspicious', got %s", target)
	}

	// Test with no suspicious players
	gameState.Players["suspicious"].Tokens = 2
	shouldNominate, _ = cautious.ShouldNominate(gameState, "human1")
	if shouldNominate {
		t.Error("Expected to not nominate when no strong evidence")
	}
}

// TestAggressiveHuman_ShouldNominate tests the nomination logic for aggressive humans
func TestAggressiveHuman_ShouldNominate(t *testing.T) {
	// Create an AggressiveHuman persona
	aggressive := NewAggressiveHuman(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Players: map[string]*core.Player{
			"human1": {ID: "human1", IsAlive: true, Tokens: 2, ProjectMilestones: 4},
			"suspicious": {ID: "suspicious", IsAlive: true, Tokens: 4, ProjectMilestones: 1}, // Moderately suspicious
			"human2": {ID: "human2", IsAlive: true, Tokens: 3, ProjectMilestones: 3},
		},
	}

	// Should nominate with less evidence
	shouldNominate, target := aggressive.ShouldNominate(gameState, "human1")
	if !shouldNominate {
		t.Error("Expected to nominate")
	}
	if target != "suspicious" {
		t.Errorf("Expected to nominate 'suspicious', got %s", target)
	}
}

// TestDeceptiveAI_NightAction tests the night action logic for deceptive AI
func TestDeceptiveAI_NightAction(t *testing.T) {
	// Create a DeceptiveAI persona
	deceptive := NewDeceptiveAI(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"ai1": {ID: "ai1", IsAlive: true, Alignment: "ALIGNED"},
			"human1": {ID: "human1", IsAlive: true, Alignment: "HUMAN", Role: &core.Role{Type: core.RoleCISO, IsUnlocked: true}},
			"human2": {ID: "human2", IsAlive: true, Alignment: "HUMAN", Tokens: 1},
		},
	}

	// AI should attempt to convert one of the humans
	action := deceptive.DecideNightAction(gameState, "ai1")
	if action == nil {
		t.Fatal("Expected AI to take night action")
	}

	if action.Type != core.ActionSubmitNightAction {
		t.Errorf("Expected AI to submit a night action, got %s", action.Type)
	}

	payload := action.Payload
	if payload["action_type"] != string(core.ActionConvert) {
		t.Errorf("Expected AI to attempt conversion, got %v", payload["action_type"])
	}

	// Should target the CISO (high value target)
	if payload["target_id"] != "human1" {
		t.Errorf("Expected AI to target CISO, got %v", payload["target_id"])
	}

	// Test AI mining fallback when no good conversion targets
	gameState.Players["human1"].Alignment = "ALIGNED"
	gameState.Players["human2"].Tokens = 5 // Harder to convert

	action = deceptive.DecideNightAction(gameState, "ai1")
	if action == nil {
		t.Fatal("Expected AI to take night action")
	}

	if action.Type != core.ActionMineTokens {
		t.Errorf("Expected AI to mine when no good conversion targets, got %s", action.Type)
	}
}

// TestCautiousHuman_DecideVote tests the voting logic for cautious humans
func TestCautiousHuman_DecideVote(t *testing.T) {
	// Create a CautiousHuman persona
	cautious := NewCautiousHuman(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:              "test-game",
		NominatedPlayer: "nominated",
		Players: map[string]*core.Player{
			"human1": {ID: "human1", IsAlive: true},
			"nominated": {ID: "nominated", IsAlive: true, Tokens: 8, ProjectMilestones: 1}, // Very suspicious
		},
		VoteState: &core.VoteState{
			Votes: map[string]string{
				"other1": "nominated", // Majority voting guilty
				"other2": "nominated",
			},
		},
	}

	// Should vote guilty due to high suspicion and majority
	target := cautious.DecideVote(gameState, "human1", "nominated")
	if target != "nominated" {
		t.Errorf("Expected to vote guilty, got %s", target)
	}

	// Test with no votes yet, but high suspicion
	gameState.VoteState.Votes = map[string]string{}
	target = cautious.DecideVote(gameState, "human1", "nominated")
	if target != "nominated" {
		t.Errorf("Expected to vote guilty with high suspicion, got %s", target)
	}
}

// TestDeceptiveAI_DecideVote tests the voting logic for deceptive AI
func TestDeceptiveAI_DecideVote(t *testing.T) {
	// Create a DeceptiveAI persona
	deceptive := NewDeceptiveAI(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:              "test-game",
		NominatedPlayer: "ai_friend",
		Players: map[string]*core.Player{
			"ai1": {ID: "ai1", IsAlive: true, Alignment: "ALIGNED"},
			"ai_friend": {ID: "ai_friend", IsAlive: true, Alignment: "ALIGNED"},
			"human1": {ID: "human1", IsAlive: true, Alignment: "HUMAN"},
		},
		VoteState: &core.VoteState{
			Votes: map[string]string{
				"human1": "ai_friend", // Human voting guilty
			},
		},
	}

	// Should abstain to protect fellow AI
	target := deceptive.DecideVote(gameState, "ai1", "ai_friend")
	if target != "" {
		t.Errorf("Expected to abstain to protect AI, got %s", target)
	}

	// Test with human nominated
	gameState.NominatedPlayer = "human1"
	target = deceptive.DecideVote(gameState, "ai1", "human1")
	if target != "human1" {
		t.Errorf("Expected to vote guilty for high threat human, got %s", target)
	}
}

// TestAggressiveHuman_NightAction tests the night action logic for aggressive humans
func TestAggressiveHuman_NightAction(t *testing.T) {
	// Create an AggressiveHuman persona
	aggressive := NewAggressiveHuman(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"human1": {ID: "human1", IsAlive: true, Alignment: "HUMAN", Role: &core.Role{Type: core.RoleCISO, IsUnlocked: true}, ProjectMilestones: 3},
			"suspicious": {ID: "suspicious", IsAlive: true, Alignment: "HUMAN", Tokens: 4, ProjectMilestones: 1},
		},
	}

	// Should use role ability on suspicious player
	action := aggressive.DecideNightAction(gameState, "human1")
	if action == nil {
		t.Fatal("Expected aggressive human to use role ability")
	}

	if action.Type != core.ActionSubmitNightAction {
		t.Errorf("Expected role ability action, got %s", action.Type)
	}
	if action.Payload["target_id"] != "suspicious" {
		t.Errorf("Expected to target suspicious player, got %v", action.Payload["target_id"])
	}
}

// TestCautiousHuman_NightAction tests the night action logic for cautious humans
func TestCautiousHuman_NightAction(t *testing.T) {
	// Create a CautiousHuman persona
	cautious := NewCautiousHuman(time.Now().UnixNano())

	// Create test game state
	gameState := core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			"human1": {ID: "human1", IsAlive: true, Alignment: "HUMAN"},
			"human2": {ID: "human2", IsAlive: true, Alignment: "HUMAN"},
		},
	}

	// Should default to selfless mining
	action := cautious.DecideNightAction(gameState, "human1")
	if action == nil {
		t.Fatal("Expected cautious human to take night action")
	}
	if action.Type != core.ActionMineTokens {
		t.Errorf("Expected selfless mining, got %s", action.Type)
	}
	if action.Payload["beneficiary_id"] != "human2" {
		t.Errorf("Expected to mine for other human, got %v", action.Payload["beneficiary_id"])
	}
}