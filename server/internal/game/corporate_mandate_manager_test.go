package game

import (
	"testing"

	"github.com/xjhc/alignment/core"
)

func TestCorporateMandateManager_AssignRandomMandate(t *testing.T) {
	gameState := &core.GameState{
		ID:       "test-game",
		Settings: core.GameSettings{StartingTokens: 2},
		Players:  make(map[string]*core.Player),
	}

	manager := NewCorporateMandateManager(gameState)

	// Test mandate assignment
	mandate := manager.AssignRandomMandate()

	if mandate == nil {
		t.Fatal("Expected mandate to be assigned, got nil")
	}

	if !mandate.IsActive {
		t.Error("Expected mandate to be active")
	}

	if mandate.Name == "" {
		t.Error("Expected mandate to have a name")
	}

	if mandate.Description == "" {
		t.Error("Expected mandate to have a description")
	}

	// Verify the mandate is stored in game state
	if gameState.CorporateMandate == nil {
		t.Error("Expected mandate to be stored in game state")
	}

	if gameState.CorporateMandate.Type != mandate.Type {
		t.Error("Expected stored mandate type to match returned mandate type")
	}
}

func TestCorporateMandateManager_SpecificMandateEffects(t *testing.T) {
	gameState := &core.GameState{
		ID:       "test-game",
		Settings: core.GameSettings{StartingTokens: 2},
		Players: map[string]*core.Player{
			"player1": {ID: "player1", IsAlive: true, Tokens: 2},
		},
		DayNumber: 1,
	}

	manager := NewCorporateMandateManager(gameState)

	tests := []struct {
		name         string
		mandateType  core.MandateType
		expectEffect string
	}{
		{
			name:         "Aggressive Growth",
			mandateType:  core.MandateAggressiveGrowth,
			expectEffect: "starting_tokens_modifier",
		},
		{
			name:         "Total Transparency",
			mandateType:  core.MandateTransparency,
			expectEffect: "public_voting_only",
		},
		{
			name:         "Security Lockdown",
			mandateType:  core.MandateSecurityLockdown,
			expectEffect: "milestones_for_abilities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mandate := manager.ActivateMandate(tt.mandateType)

			if mandate == nil {
				t.Fatal("Expected mandate to be assigned")
			}

			if _, exists := mandate.Effects[tt.expectEffect]; !exists {
				t.Errorf("Expected mandate to have effect %s", tt.expectEffect)
			}
		})
	}
}

func TestCorporateMandateManager_AggressiveGrowthEffects(t *testing.T) {
	gameState := &core.GameState{
		ID:       "test-game",
		Settings: core.GameSettings{StartingTokens: 2},
		Players: map[string]*core.Player{
			"player1": {ID: "player1", IsAlive: true, Tokens: 2},
		},
		DayNumber: 1,
	}

	manager := NewCorporateMandateManager(gameState)
	initialTokens := gameState.Settings.StartingTokens

	mandate := manager.ActivateMandate(core.MandateAggressiveGrowth)
	
	if mandate == nil {
		t.Fatal("Expected mandate to be assigned")
	}

	// Check that starting tokens were increased
	if gameState.Settings.StartingTokens != initialTokens+1 {
		t.Errorf("Expected starting tokens to increase by 1, got %d", gameState.Settings.StartingTokens)
	}

	// Check that existing players got the bonus (day 1 only)
	player := gameState.Players["player1"]
	expectedTokens := 3 // 2 initial + 1 bonus
	if player.Tokens != expectedTokens {
		t.Errorf("Expected player tokens to be %d, got %d", expectedTokens, player.Tokens)
	}

	// Check mining restrictions
	modifier, slotsReduced := manager.CheckMiningRestrictions()
	if modifier != 0.75 {
		t.Errorf("Expected mining modifier to be 0.75, got %f", modifier)
	}
	if !slotsReduced {
		t.Error("Expected mining slots to be reduced")
	}
}

func TestCorporateMandateManager_TransparencyEffects(t *testing.T) {
	gameState := &core.GameState{
		ID:       "test-game",
		Settings: core.GameSettings{},
		Players:  make(map[string]*core.Player),
	}

	manager := NewCorporateMandateManager(gameState)
	manager.ActivateMandate(core.MandateTransparency)

	// Check communication restrictions
	publicOnly, noDirectMessages := manager.CheckCommunicationRestrictions()
	if !publicOnly {
		t.Error("Expected public voting only to be true")
	}
	if !noDirectMessages {
		t.Error("Expected no direct messages to be true")
	}
}

func TestCorporateMandateManager_SecurityLockdownEffects(t *testing.T) {
	gameState := &core.GameState{
		ID:        "test-game",
		Settings:  core.GameSettings{},
		Players:   make(map[string]*core.Player),
		DayNumber: 3, // Odd night
	}

	manager := NewCorporateMandateManager(gameState)
	manager.ActivateMandate(core.MandateSecurityLockdown)

	// Check milestone requirement
	requirement := manager.GetMilestoneRequirement()
	if requirement != 4 {
		t.Errorf("Expected milestone requirement to be 4, got %d", requirement)
	}

	// Check AI conversion restriction on odd nights
	allowed := manager.IsAIConversionAllowed()
	if allowed {
		t.Error("Expected AI conversion to be blocked on odd nights")
	}

	// Test even night
	gameState.DayNumber = 4
	allowed = manager.IsAIConversionAllowed()
	if !allowed {
		t.Error("Expected AI conversion to be allowed on even nights")
	}
}

func TestCorporateMandateManager_RoleAbilityRequirements(t *testing.T) {
	gameState := &core.GameState{
		ID:       "test-game",
		Settings: core.GameSettings{},
		Players:  make(map[string]*core.Player),
	}

	manager := NewCorporateMandateManager(gameState)

	player := &core.Player{
		ID:                "player1",
		ProjectMilestones: 3,
	}

	// Test without mandate (default requirement: 3)
	canUse, reason := manager.CheckRoleAbilityRequirements(player)
	if !canUse {
		t.Errorf("Expected player to be able to use ability without mandate, got reason: %s", reason)
	}

	// Test with Security Lockdown (requirement: 4)
	manager.ActivateMandate(core.MandateSecurityLockdown)
	canUse, reason = manager.CheckRoleAbilityRequirements(player)
	if canUse {
		t.Error("Expected player to be unable to use ability with Security Lockdown mandate")
	}
	if reason == "" {
		t.Error("Expected reason to be provided when ability is blocked")
	}

	// Test with enough milestones
	player.ProjectMilestones = 4
	canUse, reason = manager.CheckRoleAbilityRequirements(player)
	if !canUse {
		t.Errorf("Expected player to be able to use ability with sufficient milestones, got reason: %s", reason)
	}
}

func TestCorporateMandateManager_MiningPoolModification(t *testing.T) {
	gameState := &core.GameState{
		ID:       "test-game",
		Settings: core.GameSettings{},
		Players:  make(map[string]*core.Player),
	}

	manager := NewCorporateMandateManager(gameState)
	baseMiningSlots := 5

	// Test without mandate
	slots := manager.ApplyMandateToMiningPool(baseMiningSlots)
	if slots != baseMiningSlots {
		t.Errorf("Expected slots to remain unchanged without mandate, got %d", slots)
	}

	// Test with Aggressive Growth mandate
	manager.ActivateMandate(core.MandateAggressiveGrowth)
	slots = manager.ApplyMandateToMiningPool(baseMiningSlots)
	expected := baseMiningSlots - 1 // Reduced by 1
	if slots != expected {
		t.Errorf("Expected slots to be reduced to %d, got %d", expected, slots)
	}
}