
package game

import (
	"testing"

	"github.com/xjhc/alignment/core"
)

func TestWhistleblowerManager_StartVoting(t *testing.T) {
	// Create test game state
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Players: map[string]*core.Player{
			"alive1": {ID: "alive1", IsAlive: true},
			"dead1":  {ID: "dead1", IsAlive: false},
		},
	}

	manager := NewWhistleblowerManager(gameState)

	// Should start voting if there's a dead player
	event := manager.StartWhistleblowerVoting()
	if event == nil {
		t.Fatal("Expected voting to start")
	}

	if event.Type != core.EventWhistleblowerVotingStarted {
		t.Errorf("Expected event type WHISTLEBLOWER_VOTING_STARTED, got %s", event.Type)
	}

	// Should not start voting if no dead players
	gameState.Players["dead1"].IsAlive = true
	event = manager.StartWhistleblowerVoting()
	if event != nil {
		t.Error("Expected voting to not start with no dead players")
	}
}

func TestWhistleblowerManager_SubmitVote(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"dead1": {ID: "dead1", IsAlive: false},
			"alive1": {ID: "alive1", IsAlive: true},
		},
		WhistleblowerVoting: &core.WhistleblowerVoting{
			IsActive: true,
			CrisisOptions: []core.CrisisEventOption{
				{Type: "CRISIS_A", Title: "Crisis A"},
			},
		},
	}

	manager := NewWhistleblowerManager(gameState)

	// Valid vote from dead player
	event, err := manager.SubmitVote("dead1", "CRISIS_A")
	if err != nil {
		t.Fatalf("Expected valid vote to succeed, got error: %v", err)
	}

	if event.Type != core.EventWhistleblowerVoteCast {
		t.Errorf("Expected event type WHISTLEBLOWER_VOTE_CAST, got %s", event.Type)
	}

	// Invalid vote from alive player
	_, err = manager.SubmitVote("alive1", "CRISIS_A")
	if err == nil {
		t.Error("Expected error when alive player tries to vote")
	}

	// Invalid crisis option
	_, err = manager.SubmitVote("dead1", "INVALID_CRISIS")
	if err == nil {
		t.Error("Expected error for invalid crisis option")
	}
}

func TestWhistleblowerManager_CompleteVoting(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"dead1": {ID: "dead1", Name: "Dead 1", IsAlive: false},
			"dead2": {ID: "dead2", Name: "Dead 2", IsAlive: false},
		},
		WhistleblowerVoting: &core.WhistleblowerVoting{
			IsActive: true,
			CrisisOptions: []core.CrisisEventOption{
				{Type: "CRISIS_A"},
				{Type: "CRISIS_B"},
			},
			Votes: map[string]string{
				"dead1": "CRISIS_A",
				"dead2": "CRISIS_A",
			},
			VoteResults: map[string]int{
				"CRISIS_A": 2,
			},
		},
	}

	manager := NewWhistleblowerManager(gameState)

	// Should complete since all dead players have voted
	event := manager.CheckVotingComplete()
	if event == nil {
		t.Fatal("Expected voting to complete")
	}

	if event.Type != core.EventWhistleblowerVotingCompleted {
		t.Errorf("Expected event type WHISTLEBLOWER_VOTING_COMPLETED, got %s", event.Type)
	}

	// Check winning crisis
	winningCrisis, _ := event.Payload["selected_crisis"].(string)
	if winningCrisis != "CRISIS_A" {
		t.Errorf("Expected CRISIS_A to win, got %s", winningCrisis)
	}
}

func TestWhistleblowerManager_TieBreaker(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"dead1": {ID: "dead1", IsAlive: false},
			"dead2": {ID: "dead2", IsAlive: false},
		},
		WhistleblowerVoting: &core.WhistleblowerVoting{
			IsActive: true,
			CrisisOptions: []core.CrisisEventOption{
				{Type: "CRISIS_A"},
				{Type: "CRISIS_B"},
			},
			Votes: map[string]string{
				"dead1": "CRISIS_A",
				"dead2": "CRISIS_B",
			},
			VoteResults: map[string]int{
				"CRISIS_A": 1,
				"CRISIS_B": 1,
			},
		},
	}

	manager := NewWhistleblowerManager(gameState)

	// Force completion
	event := manager.ForceCompleteVoting()
	if event == nil {
		t.Fatal("Expected voting to complete")
	}

	// Check winning crisis (will be random, but should be one of the options)
	winningCrisis, _ := event.Payload["selected_crisis"].(string)
	if winningCrisis != "CRISIS_A" && winningCrisis != "CRISIS_B" {
		t.Errorf("Expected winner to be CRISIS_A or CRISIS_B, got %s", winningCrisis)
	}
}

func TestWhistleblowerManager_NoVotes(t *testing.T) {
	gameState := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"dead1": {ID: "dead1", IsAlive: false},
		},
		WhistleblowerVoting: &core.WhistleblowerVoting{
			IsActive: true,
			CrisisOptions: []core.CrisisEventOption{
				{Type: "CRISIS_A"},
				{Type: "CRISIS_B"},
			},
			Votes:       map[string]string{},
			VoteResults: map[string]int{},
		},
	}

	manager := NewWhistleblowerManager(gameState)

	// Force completion
	event := manager.ForceCompleteVoting()
	if event == nil {
		t.Fatal("Expected voting to complete")
	}

	// Should select a random crisis if no votes
	winningCrisis, _ := event.Payload["selected_crisis"].(string)
	if winningCrisis != "CRISIS_A" && winningCrisis != "CRISIS_B" {
		t.Errorf("Expected random winner from options, got %s", winningCrisis)
	}
}

func TestWhistleblowerManager_ClearVoting(t *testing.T) {
	gameState := &core.GameState{
		WhistleblowerVoting: &core.WhistleblowerVoting{IsActive: true},
	}
	manager := NewWhistleblowerManager(gameState)
	manager.ClearVoting()

	if gameState.WhistleblowerVoting != nil {
		t.Error("Expected whistleblower voting state to be cleared")
	}
}