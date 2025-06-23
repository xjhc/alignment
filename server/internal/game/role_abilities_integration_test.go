package game

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// TestRoleAbilityManager_CorporateMandateIntegration tests how corporate mandates affect role abilities
func TestRoleAbilityManager_CorporateMandateIntegration(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Create player with insufficient milestones normally
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 2, // Normally insufficient (need 3)
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: true,
		},
		Alignment: "HUMAN",
	}

	gameState.Players["target"] = &core.Player{
		ID:        "target",
		IsAlive:   true,
		Alignment: "ALIGNED",
	}

	// Set up corporate mandate that reduces milestone requirements
	gameState.CorporateMandate = &core.CorporateMandate{
		Type:        core.MandateAggressiveGrowth,
		Name:        "Growth Initiative",
		Description: "Reduced ability requirements",
		IsActive:    true,
		Effects: map[string]interface{}{
			"milestones_for_abilities": 2, // Reduced from 3
		},
	}

	ram := NewRoleAbilityManager(gameState)

	// Test that player can now use ability due to reduced requirements
	canUse, reason := ram.CanUseAbility("player")
	if !canUse {
		t.Errorf("Expected player to be able to use ability with corporate mandate, got: %s", reason)
	}

	// Test using the ability
	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "RUN_AUDIT",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use ability with corporate mandate: %v", err)
	}

	if len(result.PublicEvents) == 0 {
		t.Error("Expected public events from audit ability")
	}

	if len(result.PrivateEvents) == 0 {
		t.Error("Expected private events from audit ability")
	}
}

// TestRoleAbilityManager_SystemShockRecovery tests system shock expiration
func TestRoleAbilityManager_SystemShockRecovery(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Create player with expired system shock
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: true,
		},
		SystemShocks: []core.SystemShock{
			{
				Type:      core.ShockActionLock,
				IsActive:  true,
				ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
			},
		},
	}

	gameState.Players["target"] = &core.Player{
		ID:        "target",
		IsAlive:   true,
		Alignment: "ALIGNED",
	}

	ram := NewRoleAbilityManager(gameState)

	// Player should be able to use ability despite having shock (it's expired)
	canUse, reason := ram.CanUseAbility("player")
	if !canUse {
		t.Errorf("Expected player to be able to use ability with expired shock, got: %s", reason)
	}

	// Test using the ability
	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "RUN_AUDIT",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use ability with expired shock: %v", err)
	}

	if len(result.PublicEvents) == 0 {
		t.Error("Expected public events from audit ability")
	}
}

// TestRoleAbilityManager_AlignedVsHumanAbilities tests how alignment affects role abilities
func TestRoleAbilityManager_AlignedVsHumanAbilities(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Create human CISO
	gameState.Players["human_ciso"] = &core.Player{
		ID:                "human_ciso",
		IsAlive:           true,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}

	// Create aligned CISO
	gameState.Players["aligned_ciso"] = &core.Player{
		ID:                "aligned_ciso",
		IsAlive:           true,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}

	// Create human and aligned targets
	gameState.Players["human_target"] = &core.Player{
		ID:        "human_target",
		IsAlive:   true,
		Alignment: "HUMAN",
	}

	gameState.Players["aligned_target"] = &core.Player{
		ID:        "aligned_target",
		IsAlive:   true,
		Alignment: "ALIGNED",
	}

	ram := NewRoleAbilityManager(gameState)

	// Test human CISO isolating human target (should work normally)
	humanAction := RoleAbilityAction{
		PlayerID:    "human_ciso",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "human_target",
	}

	result, err := ram.UseRoleAbility(humanAction)
	if err != nil {
		t.Fatalf("Human CISO should be able to isolate human target: %v", err)
	}

	// Target should be blocked
	if !gameState.BlockedPlayersTonight["human_target"] {
		t.Error("Expected human target to be blocked by human CISO")
	}

	// Test aligned CISO isolating aligned target (should fizzle)
	alignedAction := RoleAbilityAction{
		PlayerID:    "aligned_ciso",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "aligned_target",
	}

	result, err = ram.UseRoleAbility(alignedAction)
	if err != nil {
		t.Fatalf("Aligned CISO should be able to attempt isolation: %v", err)
	}

	// Target should NOT be blocked (fizzle case)
	if gameState.BlockedPlayersTonight["aligned_target"] {
		t.Error("Expected aligned target to NOT be blocked by aligned CISO (fizzle)")
	}

	// Should have private event about fizzle
	if len(result.PrivateEvents) == 0 {
		t.Error("Expected private event about ability fizzle")
	}
}

// TestRoleAbilityManager_CTOOverclockDifferentAlignments tests CTO overclock with different alignments
func TestRoleAbilityManager_CTOOverclockDifferentAlignments(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Create human CTO
	gameState.Players["human_cto"] = &core.Player{
		ID:                "human_cto",
		IsAlive:           true,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
		Tokens:            2,
		Role: &core.Role{
			Type:       core.RoleCTO,
			IsUnlocked: true,
		},
	}

	// Create aligned CTO
	gameState.Players["aligned_cto"] = &core.Player{
		ID:                "aligned_cto",
		IsAlive:           true,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		Tokens:            2,
		Role: &core.Role{
			Type:       core.RoleCTO,
			IsUnlocked: true,
		},
	}

	// Create target player
	gameState.Players["target"] = &core.Player{
		ID:       "target",
		IsAlive:  true,
		Tokens:   1,
		AIEquity: 0,
	}

	ram := NewRoleAbilityManager(gameState)

	// Test human CTO overclocking (should not give AI equity)
	humanAction := RoleAbilityAction{
		PlayerID:    "human_cto",
		AbilityType: "OVERCLOCK_SERVERS",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(humanAction)
	if err != nil {
		t.Fatalf("Human CTO should be able to overclock: %v", err)
	}

	// Both should get tokens
	if gameState.Players["human_cto"].Tokens != 3 {
		t.Errorf("Expected human CTO to have 3 tokens, got %d", gameState.Players["human_cto"].Tokens)
	}

	if gameState.Players["target"].Tokens != 2 {
		t.Errorf("Expected target to have 2 tokens, got %d", gameState.Players["target"].Tokens)
	}

	// Target should NOT have AI equity from human CTO
	if gameState.Players["target"].AIEquity != 0 {
		t.Errorf("Expected target to have 0 AI equity from human CTO, got %d", gameState.Players["target"].AIEquity)
	}

	// Reset target
	gameState.Players["target"].Tokens = 1
	gameState.Players["target"].AIEquity = 0

	// Test aligned CTO overclocking (should give AI equity)
	alignedAction := RoleAbilityAction{
		PlayerID:    "aligned_cto",
		AbilityType: "OVERCLOCK_SERVERS",
		TargetID:    "target",
	}

	result, err = ram.UseRoleAbility(alignedAction)
	if err != nil {
		t.Fatalf("Aligned CTO should be able to overclock: %v", err)
	}

	// Both should get tokens
	if gameState.Players["aligned_cto"].Tokens != 3 {
		t.Errorf("Expected aligned CTO to have 3 tokens, got %d", gameState.Players["aligned_cto"].Tokens)
	}

	if gameState.Players["target"].Tokens != 2 {
		t.Errorf("Expected target to have 2 tokens after aligned overclock, got %d", gameState.Players["target"].Tokens)
	}

	// Target should have AI equity from aligned CTO
	if gameState.Players["target"].AIEquity != 2 {
		t.Errorf("Expected target to have 2 AI equity from aligned CTO, got %d", gameState.Players["target"].AIEquity)
	}

	// Should have private event about AI equity gain
	if len(result.PrivateEvents) == 0 {
		t.Error("Expected private event about AI equity gain")
	}
}

// TestRoleAbilityManager_CFOBudgetReallocation tests CFO budget reallocation edge cases
func TestRoleAbilityManager_CFOBudgetReallocation(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Create CFO
	gameState.Players["cfo"] = &core.Player{
		ID:                "cfo",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCFO,
			IsUnlocked: true,
		},
	}

	// Create source player with tokens
	gameState.Players["rich_player"] = &core.Player{
		ID:      "rich_player",
		IsAlive: true,
		Tokens:  5,
	}

	// Create target player without tokens
	gameState.Players["poor_player"] = &core.Player{
		ID:      "poor_player",
		IsAlive: true,
		Tokens:  0,
	}

	// Create player with insufficient tokens
	gameState.Players["broke_player"] = &core.Player{
		ID:      "broke_player",
		IsAlive: true,
		Tokens:  0,
	}

	ram := NewRoleAbilityManager(gameState)

	// Test normal reallocation
	normalAction := RoleAbilityAction{
		PlayerID:       "cfo",
		AbilityType:    "REALLOCATE_BUDGET",
		TargetID:       "rich_player",
		SecondTargetID: "poor_player",
	}

	_, err := ram.UseRoleAbility(normalAction)
	if err != nil {
		t.Fatalf("CFO should be able to reallocate budget: %v", err)
	}

	// Rich player should lose a token
	if gameState.Players["rich_player"].Tokens != 4 {
		t.Errorf("Expected rich player to have 4 tokens, got %d", gameState.Players["rich_player"].Tokens)
	}

	// Poor player should gain a token
	if gameState.Players["poor_player"].Tokens != 1 {
		t.Errorf("Expected poor player to have 1 token, got %d", gameState.Players["poor_player"].Tokens)
	}

	// Test reallocation from player with insufficient tokens
	insufficientAction := RoleAbilityAction{
		PlayerID:       "cfo",
		AbilityType:    "REALLOCATE_BUDGET",
		TargetID:       "broke_player",
		SecondTargetID: "poor_player",
	}

	_, err = ram.UseRoleAbility(insufficientAction)
	if err == nil {
		t.Error("Expected error when source player has insufficient tokens")
	}

	if err.Error() != "source player has insufficient tokens" {
		t.Errorf("Expected insufficient tokens error, got: %s", err.Error())
	}
}

// TestRoleAbilityManager_AbilityAlreadyUsed tests one-time use restriction
func TestRoleAbilityManager_AbilityAlreadyUsed(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Create player with ability
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 3,
		HasUsedAbility:    false,
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: true,
		},
	}

	gameState.Players["target"] = &core.Player{
		ID:        "target",
		IsAlive:   true,
		Alignment: "ALIGNED",
	}

	ram := NewRoleAbilityManager(gameState)

	// First use should work
	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "RUN_AUDIT",
		TargetID:    "target",
	}

	_, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("First ability use should succeed: %v", err)
	}

	// Player should be marked as having used ability
	if !gameState.Players["player"].HasUsedAbility {
		t.Error("Expected player to be marked as having used ability")
	}

	// Second use should fail
	_, err = ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for second ability use")
	}

	if err.Error() != "player has already used their ability" {
		t.Errorf("Expected already used error, got: %s", err.Error())
	}
}
