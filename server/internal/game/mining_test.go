package game

import (
	"fmt"
	"testing"
)

func TestMiningManager_ValidateMiningRequest(t *testing.T) {
	gameState := NewGameState("test-game")
	gameState.Phase.Type = PhaseNight

	// Add test players
	gameState.Players["alice"] = &Player{
		ID:      "alice",
		Name:    "Alice",
		IsAlive: true,
		Tokens:  2,
	}
	gameState.Players["bob"] = &Player{
		ID:      "bob",
		Name:    "Bob",
		IsAlive: true,
		Tokens:  1,
	}
	gameState.Players["charlie"] = &Player{
		ID:      "charlie",
		Name:    "Charlie",
		IsAlive: false,
		Tokens:  0,
	}

	miningManager := NewMiningManager(gameState)

	tests := []struct {
		name        string
		minerID     string
		targetID    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid mining request",
			minerID:     "alice",
			targetID:    "bob",
			expectError: false,
		},
		{
			name:        "Cannot mine for self",
			minerID:     "alice",
			targetID:    "alice",
			expectError: true,
			errorMsg:    "cannot mine for yourself - mining must be selfless",
		},
		{
			name:        "Dead player cannot mine",
			minerID:     "charlie",
			targetID:    "alice",
			expectError: true,
			errorMsg:    "dead players cannot mine",
		},
		{
			name:        "Cannot mine for dead player",
			minerID:     "alice",
			targetID:    "charlie",
			expectError: true,
			errorMsg:    "cannot mine for dead players",
		},
		{
			name:        "Nonexistent miner",
			minerID:     "nonexistent",
			targetID:    "alice",
			expectError: true,
			errorMsg:    "miner player not found",
		},
		{
			name:        "Nonexistent target",
			minerID:     "alice",
			targetID:    "nonexistent",
			expectError: true,
			errorMsg:    "target player not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := miningManager.ValidateMiningRequest(tt.minerID, tt.targetID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}

	// Test wrong phase
	gameState.Phase.Type = PhaseDiscussion
	err := miningManager.ValidateMiningRequest("alice", "bob")
	if err == nil || err.Error() != "mining actions can only be submitted during night phase" {
		t.Errorf("Expected phase error, got: %v", err)
	}
}

func TestMiningManager_CalculateLiquidityPool(t *testing.T) {
	tests := []struct {
		name           string
		livingHumans   int
		expectedSlots  int
		crisisModifier int
	}{
		{
			name:          "4 living humans = 2 slots",
			livingHumans:  4,
			expectedSlots: 2,
		},
		{
			name:          "5 living humans = 2 slots",
			livingHumans:  5,
			expectedSlots: 2,
		},
		{
			name:          "6 living humans = 3 slots",
			livingHumans:  6,
			expectedSlots: 3,
		},
		{
			name:          "1 living human = 1 slot (minimum)",
			livingHumans:  1,
			expectedSlots: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := NewGameState("test-game")

			// Add the specified number of living humans
			for i := 0; i < tt.livingHumans; i++ {
				gameState.Players[fmt.Sprintf("human%d", i)] = &Player{
					IsAlive:   true,
					Alignment: "HUMAN",
				}
			}

			// Add some other players to ensure the calculation is correct
			gameState.Players["ai1"] = &Player{IsAlive: true, Alignment: "ALIGNED"}
			gameState.Players["dead1"] = &Player{IsAlive: false, Alignment: "HUMAN"}

			miningManager := NewMiningManager(gameState)
			slots := miningManager.calculateLiquidityPool()
			if slots != tt.expectedSlots {
				t.Errorf("Expected %d slots, got %d", tt.expectedSlots, slots)
			}
		})
	}

	// Test with crisis event modifier
	t.Run("crisis modifier", func(t *testing.T) {
		gameState := NewGameState("test-game")

		// Add 4 living humans
		for i := 0; i < 4; i++ {
			gameState.Players[fmt.Sprintf("human%d", i)] = &Player{
				IsAlive:   true,
				Alignment: "HUMAN",
			}
		}

		gameState.CrisisEvent = &CrisisEvent{
			Effects: map[string]interface{}{
				"mining_slots_modifier": 1, // +1 slot
			},
		}

		miningManager := NewMiningManager(gameState)
		slots := miningManager.calculateLiquidityPool()
		expected := 3 // 4 humans / 2 + 1 modifier
		if slots != expected {
			t.Errorf("Expected %d slots with crisis modifier, got %d", expected, slots)
		}
	})
}

func TestMiningManager_ResolveMining(t *testing.T) {
	gameState := NewGameState("test-game")

	// Add test players
	gameState.Players["alice"] = &Player{
		ID:            "alice",
		IsAlive:       true,
		Tokens:        3,
		StatusMessage: "",
	}
	gameState.Players["bob"] = &Player{
		ID:            "bob",
		IsAlive:       true,
		Tokens:        1,
		StatusMessage: "",
	}
	gameState.Players["charlie"] = &Player{
		ID:            "charlie",
		IsAlive:       true,
		Tokens:        2,
		StatusMessage: "Mining failed - no slots available", // Had previous failure
	}
	gameState.Players["dave"] = &Player{
		ID:            "dave",
		IsAlive:       true,
		Tokens:        1,
		StatusMessage: "",
	}

	// Add humans for liquidity pool calculation
	gameState.Players["human1"] = &Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human2"] = &Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human3"] = &Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human4"] = &Player{IsAlive: true, Alignment: "HUMAN"}

	miningManager := NewMiningManager(gameState)

	tests := []struct {
		name            string
		requests        []MiningRequest
		expectedSuccess int
		expectedFailed  int
		expectedWinners []string
	}{
		{
			name: "All requests succeed when slots available",
			requests: []MiningRequest{
				{MinerID: "alice", TargetID: "bob"},
			},
			expectedSuccess: 1,
			expectedFailed:  0,
			expectedWinners: []string{"alice"},
		},
		{
			name: "Priority system - failed mining history wins",
			requests: []MiningRequest{
				{MinerID: "alice", TargetID: "bob"},    // 3 tokens, no history
				{MinerID: "charlie", TargetID: "dave"}, // 2 tokens, failed history
				{MinerID: "bob", TargetID: "alice"},    // 1 token, no history
			},
			expectedSuccess: 2, // 2 slots available
			expectedFailed:  1,
			expectedWinners: []string{"charlie", "bob"}, // charlie (failed history), bob (fewest tokens)
		},
		{
			name: "Invalid requests filtered out",
			requests: []MiningRequest{
				{MinerID: "alice", TargetID: "alice"}, // Self-mining, invalid
				{MinerID: "bob", TargetID: "charlie"}, // Valid
			},
			expectedSuccess: 1,
			expectedFailed:  0,
			expectedWinners: []string{"bob"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := miningManager.ResolveMining(tt.requests)

			if len(result.SuccessfulMines) != tt.expectedSuccess {
				t.Errorf("Expected %d successful mines, got %d", tt.expectedSuccess, len(result.SuccessfulMines))
			}

			if result.FailedMineCount != tt.expectedFailed {
				t.Errorf("Expected %d failed mines, got %d", tt.expectedFailed, result.FailedMineCount)
			}

			// Check that expected winners are included
			for _, winner := range tt.expectedWinners {
				if _, exists := result.SuccessfulMines[winner]; !exists {
					t.Errorf("Expected %s to be successful miner", winner)
				}
			}
		})
	}
}

func TestMiningManager_UpdatePlayerTokens(t *testing.T) {
	gameState := NewGameState("test-game")

	// Add test players
	gameState.Players["alice"] = &Player{
		ID:      "alice",
		IsAlive: true,
		Tokens:  2,
	}
	gameState.Players["bob"] = &Player{
		ID:      "bob",
		IsAlive: true,
		Tokens:  1,
	}

	miningManager := NewMiningManager(gameState)

	result := &MiningResult{
		SuccessfulMines: map[string]string{
			"alice": "bob", // Alice mined for Bob
		},
		FailedMineCount: 0,
		TotalRequests:   1,
		AvailableSlots:  2,
	}

	events := miningManager.UpdatePlayerTokens(result)

	// Should generate mining successful event
	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Type != EventMiningSuccessful {
		t.Errorf("Expected EventMiningSuccessful, got %s", event.Type)
	}

	if event.PlayerID != "bob" {
		t.Errorf("Expected token to go to bob, got %s", event.PlayerID)
	}

	if event.Payload["miner_id"] != "alice" {
		t.Errorf("Expected miner_id to be alice, got %v", event.Payload["miner_id"])
	}

	if event.Payload["amount"] != 1 {
		t.Errorf("Expected amount to be 1, got %v", event.Payload["amount"])
	}
}
