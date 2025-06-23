package game

import (
	"fmt"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// TestNightResolutionManager_CrisisEventIntegration tests how crisis events affect night resolution
func TestNightResolutionManager_CrisisEventIntegration(t *testing.T) {
	gameState := core.NewGameState("test-game")
	gameState.DayNumber = 1

	// Add test players
	gameState.Players["ai"] = &core.Player{
		ID:                "ai",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		AIEquity:          3,
	}
	gameState.Players["human"] = &core.Player{
		ID:                "human",
		IsAlive:           true,
		Tokens:            1,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}

	// Set up crisis event that blocks AI conversions
	gameState.CrisisEvent = &core.CrisisEvent{
		Type:        "SECURITY_BREACH",
		Title:       "Security Breach Detected",
		Description: "System lockdown prevents AI conversion attempts",
		Effects: map[string]interface{}{
			"block_ai_conversions": true,
		},
	}

	// Set up night actions - AI tries to convert
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"ai": {
			PlayerID:  "ai",
			Type:      "CONVERT",
			TargetID:  "human",
			Timestamp: time.Now(),
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.ResolveNightActions()

	// Should have a system message about conversion being blocked
	hasBlockedMessage := false
	for _, event := range events {
		if event.Type == core.EventSystemMessage {
			if message, ok := event.Payload["message"].(string); ok {
				if message == "AI conversion blocked by active crisis protocols" {
					hasBlockedMessage = true
				}
			}
		}
	}

	if !hasBlockedMessage {
		t.Error("Expected crisis event to block AI conversion")
	}

	// Human should still be human
	if gameState.Players["human"].Alignment != "HUMAN" {
		t.Error("Human should not have been converted due to crisis event")
	}
}

// TestNightResolutionManager_CorporateMandateIntegration tests corporate mandate effects
func TestNightResolutionManager_CorporateMandateIntegration(t *testing.T) {
	gameState := core.NewGameState("test-game")
	gameState.DayNumber = 1 // Odd night

	// Add test players
	gameState.Players["ai"] = &core.Player{
		ID:                "ai",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		AIEquity:          3,
	}
	gameState.Players["human"] = &core.Player{
		ID:                "human",
		IsAlive:           true,
		Tokens:            1,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}

	// Set up corporate mandate that blocks AI conversions on odd nights
	gameState.CorporateMandate = &core.CorporateMandate{
		Type:        core.MandateSecurityLockdown,
		Name:        "Security Lockdown Protocol",
		Description: "Enhanced security measures",
		IsActive:    true,
		Effects: map[string]interface{}{
			"block_ai_odd_nights": true,
		},
	}

	// Set up night actions - AI tries to convert on odd night
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"ai": {
			PlayerID:  "ai",
			Type:      "CONVERT",
			TargetID:  "human",
			Timestamp: time.Now(),
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.ResolveNightActions()

	// Should have a system message about conversion being blocked
	hasBlockedMessage := false
	for _, event := range events {
		if event.Type == core.EventSystemMessage {
			if message, ok := event.Payload["message"].(string); ok {
				if message == "AI conversion blocked by Security Lockdown Protocol on odd nights" {
					hasBlockedMessage = true
				}
			}
		}
	}

	if !hasBlockedMessage {
		t.Error("Expected corporate mandate to block AI conversion on odd night")
	}

	// Human should still be human
	if gameState.Players["human"].Alignment != "HUMAN" {
		t.Error("Human should not have been converted due to corporate mandate")
	}
}

// TestNightResolutionManager_ComplexInteractions tests multiple interactions in one night
func TestNightResolutionManager_ComplexInteractions(t *testing.T) {
	gameState := core.NewGameState("test-game")
	gameState.DayNumber = 1
	gameState.Phase.Type = core.PhaseNight

	// Add test players
	gameState.Players["ciso"] = &core.Player{
		ID:                "ciso",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}
	gameState.Players["ai"] = &core.Player{
		ID:                "ai",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		AIEquity:          3,
	}
	gameState.Players["target"] = &core.Player{
		ID:                "target",
		IsAlive:           true,
		Tokens:            1,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}
	gameState.Players["miner"] = &core.Player{
		ID:                "miner",
		IsAlive:           true,
		Tokens:            1,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}

	// Add more humans for liquidity pool
	for i := 0; i < 5; i++ {
		gameState.Players[fmt.Sprintf("human%d", i)] = &core.Player{
			ID:        fmt.Sprintf("human%d", i),
			IsAlive:   true,
			Alignment: "HUMAN",
		}
	}

	// Set up complex night actions:
	// 1. CISO blocks AI
	// 2. AI tries to convert target (should be blocked)
	// 3. Miner tries to mine for target
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"ciso": {
			PlayerID:  "ciso",
			Type:      "ISOLATE_NODE",
			TargetID:  "ai",
			Timestamp: time.Now(),
		},
		"ai": {
			PlayerID:  "ai",
			Type:      "CONVERT",
			TargetID:  "target",
			Timestamp: time.Now(),
		},
		"miner": {
			PlayerID:  "miner",
			Type:      "MINE",
			TargetID:  "target",
			Timestamp: time.Now(),
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.ResolveNightActions()

	// Should have at least: block event, mining event, summary event
	if len(events) < 3 {
		t.Errorf("Expected at least 3 events, got %d", len(events))
	}

	// Verify precedence: block should happen first, preventing conversion
	eventTypes := make([]core.EventType, len(events))
	for i, event := range events {
		eventTypes[i] = event.Type
	}

	// AI should be blocked
	if !gameState.BlockedPlayersTonight["ai"] {
		t.Error("Expected AI to be blocked by CISO")
	}

	// Target should still be human (conversion blocked)
	if gameState.Players["target"].Alignment != "HUMAN" {
		t.Error("Target should not have been converted due to block")
	}

	// Mining should succeed (target should get a token)
	if gameState.Players["target"].Tokens != 2 {
		t.Errorf("Expected target to have 2 tokens after mining, got %d", gameState.Players["target"].Tokens)
	}
}

// TestNightResolutionManager_MiningWithCrisis tests mining with crisis event modifiers
func TestNightResolutionManager_MiningWithCrisis(t *testing.T) {
	gameState := core.NewGameState("test-game")
	gameState.DayNumber = 1

	// Add test players (8 total for testing liquidity pool)
	for i := 0; i < 8; i++ {
		gameState.Players[fmt.Sprintf("player%d", i)] = &core.Player{
			ID:        fmt.Sprintf("player%d", i),
			IsAlive:   true,
			Alignment: "HUMAN",
			Tokens:    1,
		}
	}

	// Set up crisis event that reduces mining pool
	gameState.CrisisEvent = &core.CrisisEvent{
		Type:        "MARKET_CRASH",
		Title:       "Market Crash",
		Description: "Reduced mining opportunities",
		Effects: map[string]interface{}{
			"reduced_mining_pool": true,
		},
	}

	// Set up mining actions - multiple players mining for each other
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"player0": {
			PlayerID: "player0",
			Type:     "MINE",
			TargetID: "player1",
		},
		"player1": {
			PlayerID: "player1",
			Type:     "MINE",
			TargetID: "player2",
		},
		"player2": {
			PlayerID: "player2",
			Type:     "MINE",
			TargetID: "player3",
		},
		"player3": {
			PlayerID: "player3",
			Type:     "MINE",
			TargetID: "player4",
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.ResolveNightActions()

	// Count successful mining events
	successfulMines := 0
	for _, event := range events {
		if event.Type == core.EventMiningSuccessful {
			successfulMines++
		}
	}

	// With crisis event, mining pool should be reduced
	// Normal: 8 humans = 4 slots, Crisis: 4/2 = 2 slots
	// So only 2 mining attempts should succeed
	if successfulMines != 2 {
		t.Errorf("Expected 2 successful mines due to crisis reduction, got %d", successfulMines)
	}
}

// TestNightResolutionManager_AIEquityBonus tests crisis AI equity bonus
func TestNightResolutionManager_AIEquityBonus(t *testing.T) {
	gameState := core.NewGameState("test-game")
	gameState.DayNumber = 1

	// Add test players
	gameState.Players["ai"] = &core.Player{
		ID:                "ai",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		AIEquity:          2,
	}
	gameState.Players["human"] = &core.Player{
		ID:                "human",
		IsAlive:           true,
		Tokens:            1, // Less than AI equity + bonus
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}

	// Set up crisis event that gives AI equity bonus
	gameState.CrisisEvent = &core.CrisisEvent{
		Type:        "AI_UPRISING",
		Title:       "AI Systems Compromised",
		Description: "AI gains enhanced conversion abilities",
		Effects: map[string]interface{}{
			"ai_equity_bonus": 2,
		},
	}

	// Set up night actions - AI tries to convert
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"ai": {
			PlayerID:  "ai",
			Type:      "CONVERT",
			TargetID:  "human",
			Timestamp: time.Now(),
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.ResolveNightActions()

	// Should have successful conversion (AI equity 2 + bonus 2 = 4 > human tokens 1)
	hasConversion := false
	for _, event := range events {
		if event.Type == core.EventAIConversionSuccess {
			hasConversion = true
		}
	}

	if !hasConversion {
		t.Error("Expected successful AI conversion with equity bonus")
	}

	// Human should now be aligned
	if gameState.Players["human"].Alignment != "ALIGNED" {
		t.Error("Human should have been converted")
	}

	// Human should have gained AI equity (1 base + 2 bonus = 3)
	if gameState.Players["human"].AIEquity != 3 {
		t.Errorf("Expected human to have 3 AI equity after conversion, got %d", gameState.Players["human"].AIEquity)
	}
}
