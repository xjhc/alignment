package test_helpers

import (
	"sync"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// WaitWithTimeout waits for a WaitGroup to complete or times out
// This replaces time.Sleep() for deterministic test synchronization as mandated by ADR-008 Pattern V
func WaitWithTimeout(wg *sync.WaitGroup, timeout time.Duration, t *testing.T) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	
	select {
	case <-c:
		// Completed successfully
	case <-time.After(timeout):
		t.Fatalf("Test timed out waiting for WaitGroup after %v", timeout)
	}
}

// WaitWithTimeoutNoFatal is like WaitWithTimeout but returns a boolean instead of calling t.Fatalf
func WaitWithTimeoutNoFatal(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	
	select {
	case <-c:
		return true // Completed successfully
	case <-time.After(timeout):
		return false // Timed out
	}
}

// CreateTestGameState creates a test game state with basic setup for testing
func CreateTestGameState(gameID string, dayNumber int) *core.GameState {
	gameState := core.NewGameState(gameID, time.Now().UTC())
	gameState.DayNumber = dayNumber
	gameState.Phase = core.Phase{
		Type:      core.PhaseSitrep,
		StartTime: time.Now().UTC(),
		Duration:  15 * time.Second,
	}
	
	// Add a test player
	gameState.Players["test-player-1"] = &core.Player{
		ID:                "test-player-1",
		Name:              "Test Player",
		JobTitle:          "Test Role",
		ControlType:       "HUMAN",
		Status:            core.PlayerStatusAlive,
		IsAlive:           true,
		ConnectionStatus:  "CONNECTED",
		Tokens:            1,
		ProjectMilestones: 0,
		StatusMessage:     "",
		JoinedAt:          time.Now().UTC(),
		Alignment:         "HUMAN",
	}
	
	return gameState
}