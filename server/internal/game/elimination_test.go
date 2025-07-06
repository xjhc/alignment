
package game

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestEliminationManager_EliminatePlayer(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	em := NewEliminationManager(state)

	// Add players
	state.Players["player1"] = &core.Player{
		ID:        "player1",
		IsAlive:   true,
		Alignment: "HUMAN",
		Tokens:    3,
	}
	state.Players["player2"] = &core.Player{
		ID:        "player2",
		IsAlive:   true,
		Alignment: "ALIGNED",
		Tokens:    5,
	}

	// Test elimination
	events, err := em.EliminatePlayer("player1")
	if err != nil {
		t.Fatalf("Failed to eliminate player: %v", err)
	}

	if len(events) < 2 {
		t.Errorf("Expected at least two events (elimination, role reveal), got %d", len(events))
	}

	// Test cannot eliminate same player twice
	// Re-apply state change before next test
	for _, event := range events {
		*state = core.ApplyEvent(*state, event)
	}

	_, err = em.EliminatePlayer("player1")
	if err == nil {
		t.Error("Expected error when trying to eliminate already dead player")
	}

	// Test cannot eliminate nonexistent player
	_, err = em.EliminatePlayer("nonexistent")
	if err == nil {
		t.Error("Expected error when trying to eliminate nonexistent player")
	}
}

func TestEliminationManager_CheckWinCondition(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	em := NewEliminationManager(state)

	// Test AI wins by player majority (equal or more AI than humans)
	state.Players["human1"] = &core.Player{ID: "human1", IsAlive: true, Alignment: "HUMAN", Tokens: 4}
	state.Players["human2"] = &core.Player{ID: "human2", IsAlive: true, Alignment: "HUMAN", Tokens: 2}
	state.Players["ai1"] = &core.Player{ID: "ai1", IsAlive: true, Alignment: "ALIGNED", Tokens: 8}
	state.Players["ai2"] = &core.Player{ID: "ai2", IsAlive: true, Alignment: "ALIGNED", Tokens: 3}

	// 2 AI vs 2 Humans - AI should win (tie goes to AI)
	winCondition := em.CheckWinCondition()
	if winCondition == nil {
		t.Fatal("Expected win condition to be detected")
	}

	if winCondition.Winner != "AI" {
		t.Errorf("Expected AI to win, got %s", winCondition.Winner)
	}

	if winCondition.Condition != "SINGULARITY" {
		t.Errorf("Expected SINGULARITY condition, got %s", winCondition.Condition)
	}

	// Test humans win by eliminating all AI
	state.Players["ai1"].IsAlive = false
	state.Players["ai2"].IsAlive = false

	winCondition = em.CheckWinCondition()
	if winCondition == nil {
		t.Fatal("Expected win condition to be detected")
	}

	if winCondition.Winner != "HUMANS" {
		t.Errorf("Expected HUMANS to win, got %s", winCondition.Winner)
	}

	if winCondition.Condition != "CONTAINMENT" {
		t.Errorf("Expected CONTAINMENT condition, got %s", winCondition.Condition)
	}
}

func TestEliminationManager_GetAlivePlayerCount(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	em := NewEliminationManager(state)

	// Add mixed alive/dead players
	state.Players["alive1"] = &core.Player{ID: "alive1", IsAlive: true}
	state.Players["alive2"] = &core.Player{ID: "alive2", IsAlive: true}
	state.Players["dead1"] = &core.Player{ID: "dead1", IsAlive: false}

	// Test alive count
	aliveCount := em.GetAlivePlayerCount()
	if aliveCount != 2 {
		t.Errorf("Expected 2 alive players, got %d", aliveCount)
	}
}