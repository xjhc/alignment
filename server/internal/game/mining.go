package game

import (
	"fmt"
	"log"
	"sort"

	"github.com/xjhc/alignment/core"
)

// MiningManager handles the liquidity pool system for token mining
type MiningManager struct {
	gameState *core.GameState
}

// NewMiningManager creates a new mining manager
func NewMiningManager(gameState *core.GameState) *MiningManager {
	return &MiningManager{
		gameState: gameState,
	}
}

// MiningRequest represents a player's mining action
type MiningRequest struct {
	MinerID  string `json:"miner_id"`
	TargetID string `json:"target_id"`
}

// MiningResult contains the results of mining resolution
type MiningResult struct {
	SuccessfulMines map[string]string `json:"successful_mines"` // miner_id -> target_id
	FailedMineCount int               `json:"failed_mine_count"`
	TotalRequests   int               `json:"total_requests"`
	AvailableSlots  int               `json:"available_slots"`
}

// ResolveMining processes all mining requests and determines outcomes
func (mm *MiningManager) ResolveMining(requests []MiningRequest) *MiningResult {
	// Filter valid requests (selfless mining check)
	validRequests := mm.filterValidRequests(requests)

	// Calculate available mining slots
	availableSlots := mm.calculateLiquidityPool()

	// If we have slots for everyone, all succeed
	if len(validRequests) <= availableSlots {
		result := &MiningResult{
			SuccessfulMines: make(map[string]string),
			FailedMineCount: 0,
			TotalRequests:   len(validRequests),
			AvailableSlots:  availableSlots,
		}

		for _, req := range validRequests {
			result.SuccessfulMines[req.MinerID] = req.TargetID
		}

		return result
	}

	// Apply priority system to determine winners
	prioritizedRequests := mm.applyPrioritySystem(validRequests)

	// Select successful miners based on priority
	successfulMines := make(map[string]string)
	for i := 0; i < availableSlots && i < len(prioritizedRequests); i++ {
		req := prioritizedRequests[i]
		successfulMines[req.MinerID] = req.TargetID
	}

	return &MiningResult{
		SuccessfulMines: successfulMines,
		FailedMineCount: len(validRequests) - len(successfulMines),
		TotalRequests:   len(validRequests),
		AvailableSlots:  availableSlots,
	}
}

// filterValidRequests removes invalid mining requests
func (mm *MiningManager) filterValidRequests(requests []MiningRequest) []MiningRequest {
	var validRequests []MiningRequest

	for _, req := range requests {
		// Check selfless mining rule
		if req.MinerID == req.TargetID {
			continue // Cannot mine for yourself
		}

		// Check if both players exist and are alive
		miner, minerExists := mm.gameState.Players[req.MinerID]
		target, targetExists := mm.gameState.Players[req.TargetID]

		if !minerExists || !targetExists {
			continue // Invalid player IDs
		}

		if !miner.IsAlive || !target.IsAlive {
			continue // Dead players cannot mine or be mined for
		}

		validRequests = append(validRequests, req)
	}

	return validRequests
}

// calculateLiquidityPool determines how many mining slots are available
func (mm *MiningManager) calculateLiquidityPool() int {
	// Count living humans
	livingHumans := 0
	for _, player := range mm.gameState.Players {
		if player.IsAlive && player.Alignment == "HUMAN" {
			livingHumans++
		}
	}

	// Base calculation: floor(living_humans / 2)
	baseSlots := livingHumans / 2

	// Apply crisis event modifiers if any
	if mm.gameState.CrisisEvent != nil && mm.gameState.CrisisEvent.Effects != nil {
		// Check for reduced mining pool
		if reduced, exists := mm.gameState.CrisisEvent.Effects["reduced_mining_pool"]; exists {
			if isReduced, ok := reduced.(bool); ok && isReduced {
				baseSlots = baseSlots / 2 // 50% reduction
			}
		}

		// Check for mining slots modifier
		if modifier, exists := mm.gameState.CrisisEvent.Effects["mining_slots_modifier"]; exists {
			if modValue, ok := modifier.(int); ok {
				baseSlots += modValue
			}
		}
	}

	// Apply corporate mandate modifiers
	if mm.gameState.CorporateMandate != nil && mm.gameState.CorporateMandate.IsActive {
		// Check for reduced mining slots from mandate
		if reducedVal, exists := mm.gameState.CorporateMandate.Effects["reduced_mining_slots"]; exists {
			if reduced, ok := reducedVal.(bool); ok && reduced {
				baseSlots-- // Aggressive Growth Quarter reduces by 1
			}
		}
	}

	// Apply LIAISON Protocol bonus if active
	if mm.gameState.Settings.CustomSettings != nil {
		if active, exists := mm.gameState.Settings.CustomSettings["liaison_protocol_active"]; exists {
			if isActive, ok := active.(bool); ok && isActive {
				baseSlots += 2 // +2 mining slots from LIAISON Protocol
				log.Printf("[MiningManager] LIAISON Protocol bonus applied: +2 mining slots")
			}
		}
	}

	// Ensure minimum of 1 slot if there are living players
	if baseSlots < 1 && livingHumans > 0 {
		baseSlots = 1
	}

	return baseSlots
}

// applyPrioritySystem sorts mining requests by priority rules
func (mm *MiningManager) applyPrioritySystem(requests []MiningRequest) []MiningRequest {
	// Create a copy to avoid modifying the original
	prioritized := make([]MiningRequest, len(requests))
	copy(prioritized, requests)

	// Sort by priority:
	// 1. Players who failed to mine on previous night (higher priority)
	// 2. Players with fewer tokens (higher priority)
	// 3. Random tiebreaker (using player ID for deterministic results)
	sort.Slice(prioritized, func(i, j int) bool {
		reqA, reqB := prioritized[i], prioritized[j]

		// Get player data
		playerA := mm.gameState.Players[reqA.MinerID]
		playerB := mm.gameState.Players[reqB.MinerID]

		// Priority 1: Failed mining attempts from previous night
		failedA := mm.hasFailedMiningHistory(reqA.MinerID)
		failedB := mm.hasFailedMiningHistory(reqB.MinerID)

		if failedA && !failedB {
			return true // A has higher priority
		}
		if !failedA && failedB {
			return false // B has higher priority
		}

		// Priority 2: Fewer tokens (higher priority)
		if playerA.Tokens != playerB.Tokens {
			return playerA.Tokens < playerB.Tokens
		}

		// Priority 3: Deterministic tiebreaker using player ID
		return reqA.MinerID < reqB.MinerID
	})

	return prioritized
}

// hasFailedMiningHistory checks if a player failed to mine on the previous night
func (mm *MiningManager) hasFailedMiningHistory(playerID string) bool {
	// TODO: This would need to be tracked in player state or game history
	// For now, we'll use a simple heuristic based on player data
	player := mm.gameState.Players[playerID]
	if player == nil {
		return false
	}

	// If player has StatusMessage indicating mining failure, prioritize them
	return player.StatusMessage == "Mining failed - no slots available"
}

// UpdatePlayerTokens applies the mining results to player token counts
func (mm *MiningManager) UpdatePlayerTokens(result *MiningResult) []core.Event {
	var events []core.Event

	// Award tokens to successful mining targets
	for minerID, targetID := range result.SuccessfulMines {
		event := core.Event{
			ID:        fmt.Sprintf("mining_success_%s_%s", minerID, targetID),
			Type:      core.EventMiningSuccessful,
			GameID:    mm.gameState.ID,
			PlayerID:  targetID, // Token goes to target
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"miner_id":  minerID,
				"target_id": targetID,
				"amount":    1,
			},
		}
		events = append(events, event)
	}

	// For now, we don't generate individual failed mining events
	// They would be included in the night resolution summary

	return events
}

// ValidateMiningRequest checks if a mining request is valid
func (mm *MiningManager) ValidateMiningRequest(minerID, targetID string) error {
	// Check selfless mining rule
	if minerID == targetID {
		return fmt.Errorf("cannot mine for yourself - mining must be selfless")
	}

	// Check if miner exists and is alive
	miner, exists := mm.gameState.Players[minerID]
	if !exists {
		return fmt.Errorf("miner player not found")
	}
	if !miner.IsAlive {
		return fmt.Errorf("dead players cannot mine")
	}

	// Check if target exists and is alive
	target, exists := mm.gameState.Players[targetID]
	if !exists {
		return fmt.Errorf("target player not found")
	}
	if !target.IsAlive {
		return fmt.Errorf("cannot mine for dead players")
	}

	// Check if it's night phase
	if mm.gameState.Phase.Type != core.PhaseNight {
		return fmt.Errorf("mining actions can only be submitted during night phase")
	}

	return nil
}

// HandleMineAction processes a single mining action and returns events
func (mm *MiningManager) HandleMineAction(action core.Action) ([]core.Event, error) {
	targetID, _ := action.Payload["target_id"].(string)

	// Validate the mining request
	if err := mm.ValidateMiningRequest(action.PlayerID, targetID); err != nil {
		// Create a failed mining event
		event := core.Event{
			ID:        fmt.Sprintf("mining_failed_%s_%d", action.PlayerID, getCurrentTime().UnixNano()),
			Type:      core.EventMiningFailed,
			GameID:    mm.gameState.ID,
			PlayerID:  action.PlayerID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"target_id": targetID,
				"reason":    err.Error(),
			},
		}
		return []core.Event{event}, nil
	}

	// For single mining actions, we create a successful mining event
	// (The actual mining resolution happens at the end of night phase)
	event := core.Event{
		ID:        fmt.Sprintf("mining_attempted_%s_%s_%d", action.PlayerID, targetID, getCurrentTime().UnixNano()),
		Type:      core.EventMiningAttempted,
		GameID:    mm.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id": targetID,
			"miner_id":  action.PlayerID,
		},
	}

	return []core.Event{event}, nil
}
