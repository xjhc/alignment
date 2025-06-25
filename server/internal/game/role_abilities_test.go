package game

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestRoleAbilityManager_UseRunAudit(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Create VP Ethics with unlocked ability
	gameState.Players["auditor"] = &core.Player{
		ID:                "auditor",
		Name:              "VP Ethics",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: true,
		},
	}

	// Create target player
	gameState.Players["target"] = &core.Player{
		ID:        "target",
		Name:      "Target",
		IsAlive:   true,
		Alignment: "ALIGNED", // Secret alignment
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "auditor",
		AbilityType: "RUN_AUDIT",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use audit ability: %v", err)
	}

	// Should have both public and private events
	if len(result.PublicEvents) != 1 {
		t.Errorf("Expected 1 public event, got %d", len(result.PublicEvents))
	}

	if len(result.PrivateEvents) != 1 {
		t.Errorf("Expected 1 private event, got %d", len(result.PrivateEvents))
	}

	// Public event should always show "not corrupt"
	publicEvent := result.PublicEvents[0]
	if publicEvent.Type != core.EventRunAudit {
		t.Errorf("Expected EventRunAudit, got %s", publicEvent.Type)
	}

	if publicEvent.Payload["result"] != "not_corrupt" {
		t.Errorf("Expected result 'not_corrupt', got %v", publicEvent.Payload["result"])
	}

	// Private event should reveal true alignment
	privateEvent := result.PrivateEvents[0]
	if privateEvent.Payload["true_alignment"] != "ALIGNED" {
		t.Errorf("Expected true_alignment 'ALIGNED', got %v", privateEvent.Payload["true_alignment"])
	}

	// Player should be marked as having used ability
	if !gameState.Players["auditor"].HasUsedAbility {
		t.Error("Expected player to be marked as having used ability")
	}
}

func TestRoleAbilityManager_UseOverclockServers(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Create CTO with unlocked ability
	gameState.Players["cto"] = &core.Player{
		ID:                "cto",
		Name:              "CTO",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED", // AI-aligned CTO
		Role: &core.Role{
			Type:       core.RoleCTO,
			IsUnlocked: true,
		},
	}

	// Create target player
	gameState.Players["target"] = &core.Player{
		ID:       "target",
		Name:     "Target",
		IsAlive:  true,
		Tokens:   1,
		AIEquity: 0,
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "cto",
		AbilityType: "OVERCLOCK_SERVERS",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use overclock ability: %v", err)
	}

	// Both players should have gained a token
	if gameState.Players["cto"].Tokens != 3 {
		t.Errorf("Expected CTO to have 3 tokens, got %d", gameState.Players["cto"].Tokens)
	}

	if gameState.Players["target"].Tokens != 2 {
		t.Errorf("Expected target to have 2 tokens, got %d", gameState.Players["target"].Tokens)
	}

	// Target should have gained AI equity (since CTO is aligned)
	if gameState.Players["target"].AIEquity != 2 {
		t.Errorf("Expected target to have 2 AI equity, got %d", gameState.Players["target"].AIEquity)
	}

	// Should have private event for AI equity change
	if len(result.PrivateEvents) != 1 {
		t.Errorf("Expected 1 private event, got %d", len(result.PrivateEvents))
	}
}

func TestRoleAbilityManager_UseIsolateNode(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Create CISO with unlocked ability
	gameState.Players["ciso"] = &core.Player{
		ID:                "ciso",
		Name:              "CISO",
		IsAlive:           true,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}

	// Create target player
	gameState.Players["target"] = &core.Player{
		ID:        "target",
		Name:      "Target",
		IsAlive:   true,
		Alignment: "HUMAN",
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "ciso",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use isolate ability: %v", err)
	}

	// Target should be blocked
	if !gameState.BlockedPlayersTonight["target"] {
		t.Error("Expected target to be blocked")
	}

	// Should have public event
	if len(result.PublicEvents) != 1 {
		t.Errorf("Expected 1 public event, got %d", len(result.PublicEvents))
	}
}

func TestRoleAbilityManager_UseIsolateNode_AlignedCISO(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Create ALIGNED CISO with unlocked ability
	gameState.Players["ciso"] = &core.Player{
		ID:                "ciso",
		Name:              "CISO",
		IsAlive:           true,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}

	// Create ALIGNED target player
	gameState.Players["target"] = &core.Player{
		ID:        "target",
		Name:      "Target",
		IsAlive:   true,
		Alignment: "ALIGNED",
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "ciso",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "target",
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use isolate ability: %v", err)
	}

	// Target should NOT be actually blocked (fizzle case)
	if gameState.BlockedPlayersTonight != nil && gameState.BlockedPlayersTonight["target"] {
		t.Error("Expected aligned target to NOT be blocked when CISO is aligned")
	}

	// Should have both public and private events
	if len(result.PublicEvents) != 1 {
		t.Errorf("Expected 1 public event, got %d", len(result.PublicEvents))
	}

	if len(result.PrivateEvents) != 1 {
		t.Errorf("Expected 1 private event (fizzle), got %d", len(result.PrivateEvents))
	}
}

func TestRoleAbilityManager_UseReallocateBudget(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Create CFO with unlocked ability
	gameState.Players["cfo"] = &core.Player{
		ID:                "cfo",
		Name:              "CFO",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCFO,
			IsUnlocked: true,
		},
	}

	// Create source and target players
	gameState.Players["rich_player"] = &core.Player{
		ID:      "rich_player",
		Name:    "Rich Player",
		IsAlive: true,
		Tokens:  5,
	}

	gameState.Players["poor_player"] = &core.Player{
		ID:      "poor_player",
		Name:    "Poor Player",
		IsAlive: true,
		Tokens:  0,
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:       "cfo",
		AbilityType:    "REALLOCATE_BUDGET",
		TargetID:       "rich_player", // Source (loses token)
		SecondTargetID: "poor_player", // Target (gains token)
	}

	result, err := ram.UseRoleAbility(action)
	if err != nil {
		t.Fatalf("Failed to use reallocate ability: %v", err)
	}

	// Source should lose a token
	if gameState.Players["rich_player"].Tokens != 4 {
		t.Errorf("Expected rich player to have 4 tokens, got %d", gameState.Players["rich_player"].Tokens)
	}

	// Target should gain a token
	if gameState.Players["poor_player"].Tokens != 1 {
		t.Errorf("Expected poor player to have 1 token, got %d", gameState.Players["poor_player"].Tokens)
	}

	// Should have public event
	if len(result.PublicEvents) != 1 {
		t.Errorf("Expected 1 public event, got %d", len(result.PublicEvents))
	}
}

func TestRoleAbilityManager_CanUseAbility(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Player with unlocked ability
	gameState.Players["ready"] = &core.Player{
		ID:                "ready",
		IsAlive:           true,
		ProjectMilestones: 3,
		HasUsedAbility:    false,
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: true,
		},
	}

	// Player with locked ability
	gameState.Players["locked"] = &core.Player{
		ID:                "locked",
		IsAlive:           true,
		ProjectMilestones: 2, // Not enough milestones
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: false,
		},
	}

	// Player with system shock
	gameState.Players["shocked"] = &core.Player{
		ID:                "shocked",
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
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
		},
	}

	ram := NewRoleAbilityManager(gameState)

	// Test ready player
	canUse, reason := ram.CanUseAbility("ready")
	if !canUse {
		t.Errorf("Expected ready player to be able to use ability, got: %s", reason)
	}

	// Test locked player
	canUse, reason = ram.CanUseAbility("locked")
	if canUse {
		t.Error("Expected locked player to NOT be able to use ability")
	}
	if reason != "role ability not unlocked (need 3 project milestones)" {
		t.Errorf("Expected milestone reason, got: %s", reason)
	}

	// Test shocked player
	canUse, reason = ram.CanUseAbility("shocked")
	if canUse {
		t.Error("Expected shocked player to NOT be able to use ability")
	}
	if reason != "system shock prevents ability use" {
		t.Errorf("Expected shock reason, got: %s", reason)
	}
}

func TestRoleAbilityManager_SystemShockPrevention(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Player with action lock shock
	gameState.Players["shocked"] = &core.Player{
		ID:                "shocked",
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
				ExpiresAt: time.Now().Add(1 * time.Hour),
			},
		},
	}

	gameState.Players["target"] = &core.Player{
		ID:      "target",
		IsAlive: true,
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "shocked",
		AbilityType: "RUN_AUDIT",
		TargetID:    "target",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected system shock to prevent ability use")
	}

	if err.Error() != "system shock prevents ability use" {
		t.Errorf("Expected shock error, got: %v", err)
	}
}

// Additional Edge Case Tests

func TestRoleAbilityManager_InvalidAbilityType(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "INVALID_ABILITY",
		TargetID:    "target",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for invalid ability type")
	}
}

func TestRoleAbilityManager_NonExistentPlayer(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "nonexistent",
		AbilityType: "RUN_AUDIT",
		TargetID:    "target",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for nonexistent player")
	}
}

func TestRoleAbilityManager_NonExistentTarget(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleEthics,
			IsUnlocked: true,
		},
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "RUN_AUDIT",
		TargetID:    "nonexistent",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for nonexistent target")
	}
}

func TestRoleAbilityManager_DeadPlayerAttempt(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["dead"] = &core.Player{
		ID:                "dead",
		IsAlive:           false, // Dead player
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}
	gameState.Players["target"] = &core.Player{
		ID:      "target",
		IsAlive: true,
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "dead",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "target",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for dead player trying to use ability")
	}
}

func TestRoleAbilityManager_TargetDeadPlayer(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}
	gameState.Players["dead"] = &core.Player{
		ID:      "dead",
		IsAlive: false, // Dead target
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "dead",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for targeting dead player")
	}
}

func TestRoleAbilityManager_CFOInsufficientTokensForReallocation(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["cfo"] = &core.Player{
		ID:                "cfo",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCFO,
			IsUnlocked: true,
		},
	}
	gameState.Players["source"] = &core.Player{
		ID:      "source",
		IsAlive: true,
		Tokens:  0, // No tokens to transfer
	}
	gameState.Players["dest"] = &core.Player{
		ID:      "dest",
		IsAlive: true,
		Tokens:  5,
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:       "cfo",
		AbilityType:    "REALLOCATE_BUDGET",
		TargetID:       "source",      // Source player
		SecondTargetID: "dest",        // Destination player
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for CFO trying to reallocate when source has no tokens")
	}
	
	if err.Error() != "source player has no tokens to reallocate" {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestRoleAbilityManager_CEOSelfTarget(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["ceo"] = &core.Player{
		ID:                "ceo",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCEO,
			IsUnlocked: true,
		},
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "ceo",
		AbilityType: "PERFORMANCE_REVIEW",
		TargetID:    "ceo", // Self-target
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for CEO targeting themselves")
	}
}

func TestRoleAbilityManager_AbilityAlreadyUsedEdgeCase(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.Players["player"] = &core.Player{
		ID:                "player",
		IsAlive:           true,
		ProjectMilestones: 3,
		HasUsedAbility:    true, // Already used ability
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}
	gameState.Players["target"] = &core.Player{
		ID:      "target",
		IsAlive: true,
	}

	ram := NewRoleAbilityManager(gameState)

	action := RoleAbilityAction{
		PlayerID:    "player",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "target",
	}

	_, err := ram.UseRoleAbility(action)
	if err == nil {
		t.Error("Expected error for player who already used ability")
	}
}

func TestRoleAbilityManager_ConcurrentAbilityUsage(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	
	// Two players with different abilities
	gameState.Players["ciso"] = &core.Player{
		ID:                "ciso",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCISO,
			IsUnlocked: true,
		},
	}
	gameState.Players["cto"] = &core.Player{
		ID:                "cto",
		IsAlive:           true,
		ProjectMilestones: 3,
		Role: &core.Role{
			Type:       core.RoleCTO,
			IsUnlocked: true,
		},
	}
	gameState.Players["target"] = &core.Player{
		ID:      "target",
		IsAlive: true,
		Tokens:  5,
	}

	ram := NewRoleAbilityManager(gameState)

	// CISO blocks target
	action1 := RoleAbilityAction{
		PlayerID:    "ciso",
		AbilityType: "ISOLATE_NODE",
		TargetID:    "target",
	}

	result1, err1 := ram.UseRoleAbility(action1)
	if err1 != nil {
		t.Fatalf("CISO ability failed: %v", err1)
	}

	// CTO tries to enhance target (should work in same resolution)
	action2 := RoleAbilityAction{
		PlayerID:    "cto",
		AbilityType: "OVERCLOCK_SERVERS",
		TargetID:    "target",
	}

	result2, err2 := ram.UseRoleAbility(action2)
	if err2 != nil {
		t.Fatalf("CTO ability failed: %v", err2)
	}

	// Both abilities should succeed
	if len(result1.PublicEvents) == 0 {
		t.Error("CISO ability should generate public events")
	}
	if len(result2.PublicEvents) == 0 {
		t.Error("CTO ability should generate public events")
	}
}
