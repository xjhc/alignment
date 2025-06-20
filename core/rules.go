package core

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"time"
)

// CanPlayerAffordAbility checks if a player can afford to use their role ability
func CanPlayerAffordAbility(player Player, ability Ability) bool {
	if player.Role == nil || !player.Role.IsUnlocked {
		return false
	}
	
	if player.HasUsedAbility {
		return false
	}
	
	return ability.IsReady
}

// CanPlayerVote checks if a player is eligible to vote
func CanPlayerVote(player Player, phase PhaseType) bool {
	if !player.IsAlive {
		return false
	}
	
	// Check if player is silenced by system shock
	for _, shock := range player.SystemShocks {
		if shock.IsActive && shock.Type == ShockForcedSilence && time.Now().Before(shock.ExpiresAt) {
			return false
		}
	}
	
	// Check phase allows voting
	return phase == PhaseNomination || phase == PhaseVerdict || phase == PhaseExtension
}

// CanPlayerSendMessage checks if a player can send chat messages
func CanPlayerSendMessage(player Player) bool {
	if !player.IsAlive {
		return false
	}
	
	// Check if player is silenced by system shock
	for _, shock := range player.SystemShocks {
		if shock.IsActive && shock.Type == ShockForcedSilence && time.Now().Before(shock.ExpiresAt) {
			return false
		}
	}
	
	return true
}

// CanPlayerUseNightAction checks if a player can submit night actions
func CanPlayerUseNightAction(player Player, actionType NightActionType) bool {
	if !player.IsAlive {
		return false
	}
	
	// Check if player is blocked by system shock
	for _, shock := range player.SystemShocks {
		if shock.IsActive && shock.Type == ShockActionLock && time.Now().Before(shock.ExpiresAt) {
			return false
		}
	}
	
	// AI players can attempt conversion
	if player.Alignment == "ALIGNED" && actionType == ActionConvert {
		return true
	}
	
	// All players can mine tokens
	if actionType == ActionMine {
		return true
	}
	
	// Role-specific abilities
	if player.Role != nil && player.Role.IsUnlocked && !player.HasUsedAbility {
		switch player.Role.Type {
		case RoleCISO:
			return actionType == ActionInvestigate || actionType == ActionBlock
		case RoleCEO:
			return actionType == ActionProtect
		}
	}
	
	return false
}

// IsGamePhaseOver checks if the current game phase should end
func IsGamePhaseOver(gameState GameState, currentTime time.Time) bool {
	phaseEndTime := gameState.Phase.StartTime.Add(gameState.Phase.Duration)
	return currentTime.After(phaseEndTime)
}

// GetVoteWinner determines the winner of a vote based on results
func GetVoteWinner(voteState VoteState, threshold float64) (string, bool) {
	if voteState.Results == nil || len(voteState.Results) == 0 {
		return "", false
	}
	
	totalTokens := 0
	for _, tokens := range voteState.TokenWeights {
		totalTokens += tokens
	}
	
	requiredTokens := int(float64(totalTokens) * threshold)
	
	for playerID, votes := range voteState.Results {
		if votes >= requiredTokens {
			return playerID, true
		}
	}
	
	return "", false
}

// CalculateMiningSuccess determines if mining attempt succeeds
// Uses deterministic pseudo-random based on player ID and game state for testability
func CalculateMiningSuccess(player Player, difficulty float64, gameState GameState) bool {
	// Base success rate of 60%
	baseRate := 0.6
	
	// Bonus for having tokens (more resources = better equipment)
	tokenBonus := float64(player.Tokens) * 0.05
	if tokenBonus > 0.3 { // Cap at 30% bonus
		tokenBonus = 0.3
	}
	
	// Project milestone bonus
	milestoneBonus := float64(player.ProjectMilestones) * 0.1
	if milestoneBonus > 0.3 { // Cap at 30% bonus
		milestoneBonus = 0.3
	}
	
	successRate := baseRate + tokenBonus + milestoneBonus - difficulty
	if successRate < 0.1 { // Minimum 10% chance
		successRate = 0.1
	}
	if successRate > 0.9 { // Maximum 90% chance
		successRate = 0.9
	}
	
	// Deterministic pseudo-random based on player ID hash and day number
	// This ensures reproducible results for testing while maintaining randomness
	hash := hashPlayerAction(player.ID, gameState.DayNumber, "MINE")
	random := float64(hash%10000) / 10000.0 // 0.0 to 0.9999
	
	return random < successRate
}

// IsPlayerEligibleForRole checks if a player can be assigned a specific role
func IsPlayerEligibleForRole(player Player, roleType RoleType) bool {
	// Players can only have one role
	if player.Role != nil {
		return false
	}
	
	// All players are eligible for basic roles
	return true
}

// CalculateAIConversionSuccess determines if AI conversion attempt succeeds
// Uses deterministic pseudo-random based on target player and game state
func CalculateAIConversionSuccess(target Player, aiEquity int, gameState GameState) bool {
	// Base conversion rate increases with AI equity
	baseRate := float64(aiEquity) / 100.0 // 1% per equity point
	
	// Resistance based on player's role
	resistance := 0.0
	if target.Role != nil {
		switch target.Role.Type {
		case RoleCISO:
			resistance = 0.3 // CISO has high resistance
		case RoleEthics:
			resistance = 0.25 // Ethics VP has high resistance
		case RoleCEO:
			resistance = 0.2 // CEO has moderate resistance
		default:
			resistance = 0.1 // Other roles have low resistance
		}
	}
	
	// Additional resistance if player has high tokens (more resources for defense)
	if target.Tokens >= 5 {
		resistance += 0.1
	}
	
	successRate := baseRate - resistance
	if successRate < 0.05 { // Minimum 5% chance
		successRate = 0.05
	}
	if successRate > 0.8 { // Maximum 80% chance
		successRate = 0.8
	}
	
	// Deterministic pseudo-random based on target player ID and day number
	hash := hashPlayerAction(target.ID, gameState.DayNumber, "CONVERSION")
	random := float64(hash%10000) / 10000.0
	
	return random < successRate
}

// CheckWinCondition determines if any win condition has been met
func CheckWinCondition(gameState GameState) *WinCondition {
	aliveHumans := 0
	aliveAI := 0
	totalAlive := 0
	
	for _, player := range gameState.Players {
		if player.IsAlive {
			totalAlive++
			if player.Alignment == "ALIGNED" {
				aliveAI++
			} else {
				aliveHumans++
			}
		}
	}
	
	// Check for special Personal KPI win conditions first
	// Succession Planner KPI: Game ends with exactly 2 humans alive
	for _, player := range gameState.Players {
		if player.PersonalKPI != nil && player.PersonalKPI.Type == KPISuccessionPlanner {
			if aliveHumans == 2 && player.IsAlive && player.Alignment == "HUMAN" {
				return &WinCondition{
					Winner:      "HUMANS",
					Condition:   "SUCCESSION_PLANNER",
					Description: fmt.Sprintf("%s achieved succession plan with exactly 2 humans remaining", player.Name),
				}
			}
		}
	}
	
	// AI wins by majority (singularity)
	if aliveAI >= aliveHumans && aliveAI > 0 {
		return &WinCondition{
			Winner:      "AI",
			Condition:   "SINGULARITY",
			Description: "AI has achieved majority control",
		}
	}
	
	// Humans win by eliminating all AI (containment)
	if aliveAI == 0 && aliveHumans > 0 {
		return &WinCondition{
			Winner:      "HUMANS",
			Condition:   "CONTAINMENT",
			Description: "All AI threats have been contained",
		}
	}
	
	// Check if game has gone on too long (day limit)
	if gameState.DayNumber >= 7 { // Game ends after 7 days
		if aliveHumans > aliveAI {
			return &WinCondition{
				Winner:      "HUMANS",
				Condition:   "CONTAINMENT",
				Description: "Humans maintained control through time limit",
			}
		} else {
			return &WinCondition{
				Winner:      "AI",
				Condition:   "SINGULARITY",
				Description: "AI survived to time limit",
			}
		}
	}
	
	return nil // No win condition met
}

// IsValidNightActionTarget checks if a target is valid for a night action
func IsValidNightActionTarget(actor Player, target Player, actionType NightActionType) bool {
	// Can't target yourself for most actions
	if actor.ID == target.ID && actionType != ActionMine {
		return false
	}
	
	// Can't target dead players
	if !target.IsAlive {
		return false
	}
	
	// AI can only convert humans
	if actionType == ActionConvert && target.Alignment == "ALIGNED" {
		return false
	}
	
	return true
}

// CalculateTokenReward determines token rewards for various actions
func CalculateTokenReward(actionType EventType, player Player, gameState GameState) int {
	// Get base mining reward from game state (may be modified by crisis events)
	baseReward := 1
	if gameState.CrisisEvent != nil && gameState.CrisisEvent.Effects != nil {
		if miningReward, ok := gameState.CrisisEvent.Effects["mining_base_reward"].(float64); ok {
			baseReward = int(miningReward)
		}
	}
	
	switch actionType {
	case EventMiningSuccessful:
		// Base mining reward with milestone bonus
		milestoneBonus := player.ProjectMilestones / 3 // +1 token per 3 milestones
		return baseReward + milestoneBonus
		
	case EventProjectMilestone:
		// Reward for completing project milestones
		return 1
		
	case EventKPICompleted:
		// Reward for completing personal KPI
		if player.PersonalKPI != nil {
			switch player.PersonalKPI.Type {
			case KPICapitalist, KPIGuardian, KPIInquisitor:
				return 3
			case KPISuccessionPlanner:
				return 5 // Higher reward for difficult objective
			case KPIScapegoat:
				return 4 // Posthumous reward
			}
		}
		return 3
		
	default:
		return 0
	}
}

// IsMessageCorrupted checks if a message should be corrupted by system shock
// Uses deterministic pseudo-random based on player ID and current time
func IsMessageCorrupted(player Player, messageContent string) bool {
	for _, shock := range player.SystemShocks {
		if shock.IsActive && shock.Type == ShockMessageCorruption && time.Now().Before(shock.ExpiresAt) {
			// 25% chance of corruption
			// Use message content hash for deterministic corruption
			hash := hashStringWithID(messageContent, player.ID)
			probability := float64(hash%10000) / 10000.0
			return probability < 0.25
		}
	}
	return false
}

// hashPlayerAction creates a deterministic hash for player actions
// Used for reproducible "randomness" in testing while maintaining game balance
func hashPlayerAction(playerID string, dayNumber int, action string) uint32 {
	data := fmt.Sprintf("%s:%d:%s", playerID, dayNumber, action)
	hash := sha256.Sum256([]byte(data))
	return binary.BigEndian.Uint32(hash[:4])
}

// hashStringWithID creates a deterministic hash for string content with player ID
func hashStringWithID(content string, playerID string) uint32 {
	data := fmt.Sprintf("%s:%s", content, playerID)
	hash := sha256.Sum256([]byte(data))
	return binary.BigEndian.Uint32(hash[:4])
}

// CheckScapegoatKPI checks if a player achieved the Scapegoat KPI by being eliminated unanimously
func CheckScapegoatKPI(eliminatedPlayer Player, voteState VoteState) bool {
	if eliminatedPlayer.PersonalKPI == nil || eliminatedPlayer.PersonalKPI.Type != KPIScapegoat {
		return false
	}
	
	// Check if all votes were against this player (unanimous)
	totalVotes := len(voteState.Votes)
	votesAgainst := 0
	
	for _, targetID := range voteState.Votes {
		if targetID == eliminatedPlayer.ID {
			votesAgainst++
		}
	}
	
	// Must be unanimous (all votes against) and at least 3 voters
	return votesAgainst == totalVotes && totalVotes >= 3
}