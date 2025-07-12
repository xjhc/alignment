package ai

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestEnhancedRulesEngine_CalculateThreatAssessment(t *testing.T) {
	engine := NewEnhancedRulesEngineWithDefaultPersona()
	aiPlayerID := "ai-1"

	// Create test game state
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseNight},
		Players: map[string]*core.Player{
			aiPlayerID: {
				ID:        aiPlayerID,
				Name:      "AI Player",
				IsAlive:   true,
				Alignment: "AI",
				Tokens:    3,
			},
			"human-1": {
				ID:                "human-1",
				Name:              "High Threat Human",
				IsAlive:           true,
				Alignment:         "HUMAN",
				Tokens:            5,
				ProjectMilestones: 3,
				Role: &core.Role{
					Type:       core.RoleCISO,
					IsUnlocked: true,
				},
			},
			"human-2": {
				ID:                "human-2",
				Name:              "Low Threat Human",
				IsAlive:           true,
				Alignment:         "HUMAN",
				Tokens:            1,
				ProjectMilestones: 1,
			},
			"dead-human": {
				ID:        "dead-human",
				Name:      "Dead Human",
				IsAlive:   false,
				Alignment: "HUMAN",
				Tokens:    10,
			},
		},
	}

	threats := engine.CalculateThreatAssessment(gameState, aiPlayerID)

	// Should only include alive human players
	if len(threats) != 2 {
		t.Errorf("Expected 2 threats, got %d", len(threats))
	}

	// Should be sorted by threat level (highest first)
	if threats[0].PlayerID != "human-1" {
		t.Errorf("Expected human-1 to be highest threat, got %s", threats[0].PlayerID)
	}

	// High threat player should have higher score than low threat
	if threats[0].ThreatLevel <= threats[1].ThreatLevel {
		t.Errorf("Expected human-1 threat level (%.2f) > human-2 (%.2f)", 
			threats[0].ThreatLevel, threats[1].ThreatLevel)
	}

	// High threat should have CISO-related reasoning
	hasCSIOReasoning := false
	for _, reason := range threats[0].Reasoning {
		if reason == "CISO can audit and expose AI players" {
			hasCSIOReasoning = true
			break
		}
	}
	if !hasCSIOReasoning {
		t.Error("Expected CISO-related reasoning for high threat player")
	}
}

func TestEnhancedRulesEngine_DecideNightAction(t *testing.T) {
	tests := []struct {
		name           string
		aiPlayer       *core.Player
		humanPlayers   []*core.Player
		expectedAction core.ActionType
	}{
		{
			name: "AI with role ability should use it",
			aiPlayer: &core.Player{
				ID:                "ai-1",
				IsAlive:           true,
				Alignment:         "AI",
				ProjectMilestones: 3,
				HasUsedAbility:    false,
				Role: &core.Role{
					Type:       core.RoleCISO,
					IsUnlocked: true,
				},
			},
			humanPlayers: []*core.Player{
				{ID: "human-1", IsAlive: true, Alignment: "HUMAN", ControlType: "HUMAN", Tokens: 3},
			},
			expectedAction: core.ActionIsolateNode,
		},
		{
			name: "AI without unlocked role should consider milestones",
			aiPlayer: &core.Player{
				ID:                "ai-1",
				IsAlive:           true,
				Alignment:         "AI",
				ProjectMilestones: 1,
				AIEquity:          1, // Not enough for conversion
			},
			humanPlayers: []*core.Player{
				{ID: "human-1", IsAlive: true, Alignment: "HUMAN", ControlType: "HUMAN", Tokens: 3},
			},
			expectedAction: core.ActionProjectMilestones, // AI chooses to work on milestones
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewEnhancedRulesEngineWithDefaultPersona()
			
			players := map[string]*core.Player{
				tt.aiPlayer.ID: tt.aiPlayer,
			}
			for i, human := range tt.humanPlayers {
				playerID := fmt.Sprintf("human-%d", i+1)
				human.ID = playerID
				players[playerID] = human
			}

			gameState := &core.GameState{
				ID:        "test-game",
				DayNumber: 1,
				Phase:     core.Phase{Type: core.PhaseNight},
				Players:   players,
			}

			action := engine.DecideAction(gameState, tt.aiPlayer.ID)

			if action == nil {
				t.Fatal("Expected action, got nil")
			}

			if action.Type != tt.expectedAction {
				t.Errorf("Expected action type %s, got %s", tt.expectedAction, action.Type)
			}

			if action.PlayerID != tt.aiPlayer.ID {
				t.Errorf("Expected player ID %s, got %s", tt.aiPlayer.ID, action.PlayerID)
			}
		})
	}
}

func TestEnhancedRulesEngine_DecideVote(t *testing.T) {
	engine := NewEnhancedRulesEngineWithDefaultPersona()
	aiPlayerID := "ai-1"

	gameState := &core.GameState{
		ID:              "test-game",
		DayNumber:       2,
		Phase:           core.Phase{Type: core.PhaseNomination},
		NominatedPlayer: "human-1",
		Players: map[string]*core.Player{
			aiPlayerID: {
				ID:        aiPlayerID,
				Name:      "AI Player",
				IsAlive:   true,
				Alignment: "AI",
			},
			"human-1": {
				ID:        "human-1",
				Name:      "Human 1",
				IsAlive:   true,
				Alignment: "HUMAN",
				Tokens:    5,
			},
			"human-2": {
				ID:        "human-2",
				Name:      "Human 2", 
				IsAlive:   true,
				Alignment: "HUMAN",
				Tokens:    2,
			},
		},
	}

	// Test nomination vote
	action := engine.DecideAction(gameState, aiPlayerID)
	if action == nil {
		t.Fatal("Expected nomination action, got nil")
	}

	if action.Type != core.ActionSubmitVote {
		t.Errorf("Expected ActionSubmitVote, got %s", action.Type)
	}

	// Should target a human player
	targetID := action.Payload["target_id"].(string)
	if targetID != "human-1" && targetID != "human-2" {
		t.Errorf("Expected to target a human player, got %s", targetID)
	}

	// Test verdict vote
	gameState.Phase.Type = core.PhaseVerdict
	action = engine.DecideAction(gameState, aiPlayerID)
	if action == nil {
		t.Fatal("Expected verdict action, got nil")
	}

	// Should vote GUILTY for human player
	verdict := action.Payload["target_id"].(string)
	if verdict != "GUILTY" {
		t.Errorf("Expected GUILTY verdict for human player, got %s", verdict)
	}
}

func TestStrategicPersonas(t *testing.T) {
	shadow := GetShadowPersona()
	puppeteer := GetPuppeteerPersona()

	// Test shadow persona characteristics
	if shadow.ID != "shadow" {
		t.Errorf("Expected shadow ID, got %s", shadow.ID)
	}
	if shadow.AggressionWeight >= puppeteer.AggressionWeight {
		t.Error("Shadow should be less aggressive than Puppeteer")
	}
	if shadow.DeceptionWeight <= puppeteer.DeceptionWeight {
		t.Error("Shadow should be more deceptive than Puppeteer")
	}

	// Test puppeteer persona characteristics
	if puppeteer.ID != "puppeteer" {
		t.Errorf("Expected puppeteer ID, got %s", puppeteer.ID)
	}
	if puppeteer.ChatFrequency <= shadow.ChatFrequency {
		t.Error("Puppeteer should be more chatty than Shadow")
	}
}

func TestPlayerThreatAnalysis(t *testing.T) {
	engine := NewEnhancedRulesEngineWithDefaultPersona()

	// Test CISO player (high threat)
	cisoPlayer := &core.Player{
		ID:                "ciso",
		IsAlive:           true,
		Tokens:            4,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}

	gameState := &core.GameState{Players: map[string]*core.Player{}}
	threatScore := engine.calculatePlayerThreatScore(cisoPlayer, gameState)

	if threatScore < 0.8 {
		t.Errorf("CISO should have high threat score, got %.2f", threatScore)
	}

	reasoning := engine.generateThreatReasoning(cisoPlayer, gameState)
	hasCISOThreat := false
	for _, reason := range reasoning {
		if reason == "CISO can audit and expose AI players" {
			hasCISOThreat = true
			break
		}
	}
	if !hasCISOThreat {
		t.Error("CISO reasoning should mention audit capability")
	}
}

// Benchmark tests for performance
func BenchmarkThreatAssessment(b *testing.B) {
	engine := NewEnhancedRulesEngineWithDefaultPersona()
	gameState := createLargeTestGameState(50) // 50 players

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.CalculateThreatAssessment(gameState, "ai-1")
	}
}

func BenchmarkDecideAction(b *testing.B) {
	engine := NewEnhancedRulesEngineWithDefaultPersona()
	gameState := createLargeTestGameState(20) // 20 players

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.DecideAction(gameState, "ai-1")
	}
}

// Helper function to create test game states
func createLargeTestGameState(playerCount int) *core.GameState {
	players := make(map[string]*core.Player)
	
	// Add AI player
	players["ai-1"] = &core.Player{
		ID:        "ai-1",
		IsAlive:   true,
		Alignment: "AI",
		Tokens:    3,
	}

	// Add human players
	for i := 1; i < playerCount; i++ {
		playerID := fmt.Sprintf("human-%d", i)
		players[playerID] = &core.Player{
			ID:                playerID,
			IsAlive:           true,
			Alignment:         "HUMAN",
			Tokens:            rand.Intn(6) + 1,
			ProjectMilestones: rand.Intn(4),
		}
	}

	return &core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseNight, StartTime: time.Now()},
		Players:   players,
	}
}