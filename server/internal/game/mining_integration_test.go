package game

import (
	"fmt"
	"testing"

	"github.com/xjhc/alignment/core"
)

// TestMiningManager_CorporateMandateEffects tests how corporate mandates affect mining
func TestMiningManager_CorporateMandateEffects(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Add test players
	for i := 0; i < 6; i++ {
		gameState.Players[fmt.Sprintf("player%d", i)] = &core.Player{
			ID:        fmt.Sprintf("player%d", i),
			IsAlive:   true,
			Alignment: "HUMAN",
			StatusMessage: "",
		}
	}

	// Set up corporate mandate that affects milestone requirements
	gameState.CorporateMandate = &core.CorporateMandate{
		Type:        core.MandateAggressiveGrowth,
		Name:        "Aggressive Growth Initiative",
		Description: "Reduced requirements for abilities",
		IsActive:    true,
		Effects: map[string]interface{}{
			"milestones_for_abilities": 2, // Reduced from default 3
		},
	}

	mm := NewMiningManager(gameState)

	// Test liquidity pool calculation with corporate mandate
	pool := mm.calculateLiquidityPool()
	expectedPool := 3 // 6 humans / 2 = 3 slots
	if pool != expectedPool {
		t.Errorf("Expected liquidity pool of %d, got %d", expectedPool, pool)
	}

	// Test mining requests
	requests := []MiningRequest{
		{MinerID: "player0", TargetID: "player1"},
		{MinerID: "player1", TargetID: "player2"},
		{MinerID: "player2", TargetID: "player3"},
		{MinerID: "player3", TargetID: "player4"}, // This should fail due to pool limit
	}

	result := mm.ResolveMining(requests)

	// Should have 3 successful mines (pool limit)
	if len(result.SuccessfulMines) != 3 {
		t.Errorf("Expected 3 successful mines, got %d", len(result.SuccessfulMines))
	}

	// Should have 1 failed mine
	if result.FailedMineCount != 1 {
		t.Errorf("Expected 1 failed mine, got %d", result.FailedMineCount)
	}
}

// TestMiningManager_PrioritySystem tests the mining priority system
func TestMiningManager_PrioritySystem(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Add test players with different mining history
	gameState.Players["priority_player"] = &core.Player{
		ID:            "priority_player",
		IsAlive:       true,
		Alignment:     "HUMAN",
		StatusMessage: "Mining failed - no slots available", // Has priority
	}
	gameState.Players["normal_player"] = &core.Player{
		ID:            "normal_player",
		IsAlive:       true,
		Alignment:     "HUMAN",
		StatusMessage: "",
	}
	gameState.Players["target1"] = &core.Player{
		ID:        "target1",
		IsAlive:   true,
		Alignment: "HUMAN",
	}
	gameState.Players["target2"] = &core.Player{
		ID:        "target2",
		IsAlive:   true,
		Alignment: "HUMAN",
	}

	// Add more players to reach minimum pool size
	gameState.Players["human1"] = &core.Player{ID: "human1", IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human2"] = &core.Player{ID: "human2", IsAlive: true, Alignment: "HUMAN"}

	mm := NewMiningManager(gameState)

	// Create requests where priority matters (only 1 slot available)
	requests := []MiningRequest{
		{MinerID: "normal_player", TargetID: "target1"},
		{MinerID: "priority_player", TargetID: "target2"}, // Should win due to priority
	}

	// Set up crisis to reduce pool to 1 slot
	gameState.CrisisEvent = &core.CrisisEvent{
		Effects: map[string]interface{}{
			"reduced_mining_pool": true,
		},
	}

	result := mm.ResolveMining(requests)

	// Should have 1 successful mine
	if len(result.SuccessfulMines) != 1 {
		t.Errorf("Expected 1 successful mine, got %d", len(result.SuccessfulMines))
	}

	// Priority player should win
	if result.SuccessfulMines["priority_player"] != "target2" {
		t.Error("Expected priority player to win mining slot")
	}

	// Normal player should fail (1 total request - 1 successful = 1 failed)
	if result.FailedMineCount != 1 {
		t.Error("Expected 1 failed mining request")
	}
}

// TestMiningManager_SelfMiningPrevention tests that players cannot mine for themselves
func TestMiningManager_SelfMiningPrevention(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Add test players
	gameState.Players["player1"] = &core.Player{
		ID:        "player1",
		IsAlive:   true,
		Alignment: "HUMAN",
	}
	gameState.Players["player2"] = &core.Player{
		ID:        "player2",
		IsAlive:   true,
		Alignment: "HUMAN",
	}

	mm := NewMiningManager(gameState)

	// Test self-mining validation
	err := mm.ValidateMiningRequest("player1", "player1")
	if err == nil {
		t.Error("Expected error for self-mining request")
	}
	if err.Error() != "players cannot mine for themselves" {
		t.Errorf("Expected self-mining error, got: %s", err.Error())
	}

	// Test valid mining validation
	err = mm.ValidateMiningRequest("player1", "player2")
	if err != nil {
		t.Errorf("Expected valid mining request to pass, got error: %s", err.Error())
	}
}

// TestMiningManager_DeadPlayerValidation tests mining validation with dead players
func TestMiningManager_DeadPlayerValidation(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Add test players
	gameState.Players["alive_player"] = &core.Player{
		ID:        "alive_player",
		IsAlive:   true,
		Alignment: "HUMAN",
	}
	gameState.Players["dead_player"] = &core.Player{
		ID:        "dead_player",
		IsAlive:   false,
		Alignment: "HUMAN",
	}

	mm := NewMiningManager(gameState)

	// Test dead player mining
	err := mm.ValidateMiningRequest("dead_player", "alive_player")
	if err == nil {
		t.Error("Expected error for dead player mining")
	}

	// Test mining for dead player
	err = mm.ValidateMiningRequest("alive_player", "dead_player")
	if err == nil {
		t.Error("Expected error for mining for dead player")
	}
}

// TestMiningManager_CrisisEffects tests various crisis event effects on mining
func TestMiningManager_CrisisEffects(t *testing.T) {
	gameState := core.NewGameState("test-game")

	// Add test players
	for i := 0; i < 8; i++ {
		gameState.Players[fmt.Sprintf("player%d", i)] = &core.Player{
			ID:        fmt.Sprintf("player%d", i),
			IsAlive:   true,
			Alignment: "HUMAN",
		}
	}

	mm := NewMiningManager(gameState)

	// Test normal liquidity pool (no crisis)
	normalPool := mm.calculateLiquidityPool()
	expectedNormal := 4 // 8 humans / 2 = 4 slots
	if normalPool != expectedNormal {
		t.Errorf("Expected normal pool of %d, got %d", expectedNormal, normalPool)
	}

	// Test reduced mining pool crisis
	gameState.CrisisEvent = &core.CrisisEvent{
		Effects: map[string]interface{}{
			"reduced_mining_pool": true,
		},
	}

	reducedPool := mm.calculateLiquidityPool()
	expectedReduced := 2 // 4 / 2 = 2 slots
	if reducedPool != expectedReduced {
		t.Errorf("Expected reduced pool of %d, got %d", expectedReduced, reducedPool)
	}

	// Test other crisis effects don't affect pool
	gameState.CrisisEvent = &core.CrisisEvent{
		Effects: map[string]interface{}{
			"some_other_effect": true,
		},
	}

	unaffectedPool := mm.calculateLiquidityPool()
	if unaffectedPool != expectedNormal {
		t.Errorf("Expected unaffected pool of %d, got %d", expectedNormal, unaffectedPool)
	}
}