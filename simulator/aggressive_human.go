package simulator

import (
	"math/rand"

	"github.com/xjhc/alignment/core"
)

// AggressiveHuman persona - prioritizes accumulating tokens and tends to nominate suspicious players
type AggressiveHuman struct {
	rand *rand.Rand
}

func NewAggressiveHuman(seed int64) *AggressiveHuman {
	return &AggressiveHuman{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (a *AggressiveHuman) GetID() string {
	return "aggressive_human"
}

func (a *AggressiveHuman) GetDescription() string {
	return "Prioritizes accumulating tokens. Tends to nominate players who act suspiciously. More willing to take risks."
}

func (a *AggressiveHuman) DecideAction(gameState core.GameState, playerID string) *core.Action {
	player := gameState.Players[playerID]
	if player == nil || !player.IsAlive {
		return nil
	}

	switch gameState.Phase.Type {
	case core.PhaseNomination:
		if shouldNominate, target := a.ShouldNominate(gameState, playerID); shouldNominate {
			return &core.Action{
				Type:     core.ActionSubmitVote,
				PlayerID: playerID,
				GameID:   gameState.ID,
				Payload: map[string]interface{}{
					"target_player_id": target,
				},
			}
		}
	case core.PhaseVerdict, core.PhaseExtension:
		if gameState.NominatedPlayer != "" {
			target := a.DecideVote(gameState, playerID, gameState.NominatedPlayer)
			return &core.Action{
				Type:     core.ActionSubmitVote,
				PlayerID: playerID,
				GameID:   gameState.ID,
				Payload: map[string]interface{}{
					"target_player_id": target,
				},
			}
		}
	case core.PhaseNight:
		return a.DecideNightAction(gameState, playerID)
	case core.PhaseSitrep:
		// Aggressive response to pulse check
		responses := []string{
			"Systems running hot but stable",
			"Detecting anomalies, investigating",
			"Performance metrics exceeding targets",
			"Network activity elevated",
		}
		response := responses[a.rand.Intn(len(responses))]
		return &core.Action{
			Type:     core.ActionSubmitPulseCheck,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"response": response,
			},
		}
	}

	// More likely to chat than cautious humans
	if message := a.DecideChatMessage(gameState, playerID); message != nil {
		return &core.Action{
			Type:     core.ActionSendMessage,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"content": *message,
			},
		}
	}

	return nil
}

func (a *AggressiveHuman) ShouldNominate(gameState core.GameState, playerID string) (bool, string) {
	// Aggressive humans nominate more readily with less evidence
	suspiciousPlayers := a.findSuspiciousPlayers(gameState, playerID)
	
	if len(suspiciousPlayers) == 0 {
		return false, ""
	}
	
	// Much lower threshold than cautious (>40% vs >70%)
	for _, suspiciousPlayer := range suspiciousPlayers {
		if a.calculateSuspicionLevel(gameState, suspiciousPlayer) > 0.4 {
			return true, suspiciousPlayer
		}
	}
	
	// If multiple suspicious players, pick the most suspicious one
	if len(suspiciousPlayers) > 1 {
		mostSuspicious := suspiciousPlayers[0]
		highestSuspicion := a.calculateSuspicionLevel(gameState, mostSuspicious)
		
		for _, player := range suspiciousPlayers[1:] {
			suspicion := a.calculateSuspicionLevel(gameState, player)
			if suspicion > highestSuspicion {
				mostSuspicious = player
				highestSuspicion = suspicion
			}
		}
		
		return true, mostSuspicious
	}
	
	return false, ""
}

func (a *AggressiveHuman) DecideVote(gameState core.GameState, playerID string, nominatedPlayer string) string {
	// Aggressive humans are more likely to vote guilty
	suspicionLevel := a.calculateSuspicionLevel(gameState, nominatedPlayer)
	
	// Lower threshold for voting guilty (>30% vs >60%)
	if suspicionLevel > 0.3 {
		return nominatedPlayer // Vote guilty
	}
	
	// Even if not suspicious, might vote guilty if they have high tokens
	// (thinking they got help from AI)
	nominatedPlayerData := gameState.Players[nominatedPlayer]
	if nominatedPlayerData != nil && nominatedPlayerData.Tokens > 4 {
		return nominatedPlayer
	}
	
	// Check voting trends but be less influenced by them
	if gameState.VoteState != nil {
		guiltyVotes := 0
		for _, vote := range gameState.VoteState.Votes {
			if vote == nominatedPlayer {
				guiltyVotes++
			}
		}
		
		// If there's momentum for guilty, join it
		if guiltyVotes > 2 {
			return nominatedPlayer
		}
	}
	
	return "" // Abstain if no strong reason to vote guilty
}

func (a *AggressiveHuman) DecideNightAction(gameState core.GameState, playerID string) *core.Action {
	player := gameState.Players[playerID]
	
	// Prioritize role abilities over mining
	if player.Role != nil && player.Role.IsUnlocked && !player.HasUsedAbility {
		switch player.Role.Type {
		case core.RoleCISO:
			// Aggressively investigate suspicious players
			suspiciousPlayers := a.findSuspiciousPlayers(gameState, playerID)
			if len(suspiciousPlayers) > 0 {
				// Pick the most suspicious one
				target := suspiciousPlayers[0]
				if len(suspiciousPlayers) > 1 {
					highestSuspicion := a.calculateSuspicionLevel(gameState, target)
					for _, player := range suspiciousPlayers[1:] {
						suspicion := a.calculateSuspicionLevel(gameState, player)
						if suspicion > highestSuspicion {
							target = player
							highestSuspicion = suspicion
						}
					}
				}
				
				return &core.Action{
					Type:     core.ActionSubmitNightAction,
					PlayerID: playerID,
					GameID:   gameState.ID,
					Payload: map[string]interface{}{
						"action_type": string(core.ActionInvestigate),
						"target_id":   target,
					},
				}
			}
		case core.RoleCEO:
			// Protect high-value players (including self if possible)
			valuablePlayers := a.findValuablePlayers(gameState, playerID)
			if len(valuablePlayers) > 0 {
				target := valuablePlayers[0]
				return &core.Action{
					Type:     core.ActionSubmitNightAction,
					PlayerID: playerID,
					GameID:   gameState.ID,
					Payload: map[string]interface{}{
						"action_type": string(core.ActionProtect),
						"target_id":   target,
					},
				}
			}
		case core.RoleCTO:
			// Use CTO abilities aggressively
			return &core.Action{
				Type:     core.ActionRunAudit,
				PlayerID: playerID,
				GameID:   gameState.ID,
				Payload:  map[string]interface{}{},
			}
		}
	}
	
	// Mining strategy: mine for players who might mine back
	// Prioritize players with high milestones (likely to survive and reciprocate)
	alivePlayers := make([]string, 0)
	for id, p := range gameState.Players {
		if p.IsAlive && id != playerID {
			alivePlayers = append(alivePlayers, id)
		}
	}
	
	if len(alivePlayers) > 0 {
		// Sort by project milestones (prefer high contributors)
		bestTarget := alivePlayers[0]
		highestMilestones := gameState.Players[bestTarget].ProjectMilestones
		
		for _, id := range alivePlayers[1:] {
			if gameState.Players[id].ProjectMilestones > highestMilestones {
				bestTarget = id
				highestMilestones = gameState.Players[id].ProjectMilestones
			}
		}
		
		return &core.Action{
			Type:     core.ActionMineTokens,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"beneficiary_id": bestTarget,
			},
		}
	}
	
	return nil
}

func (a *AggressiveHuman) DecideChatMessage(gameState core.GameState, playerID string) *string {
	// Aggressive humans chat more frequently (30% vs 10%)
	if a.rand.Float64() > 0.3 {
		return nil
	}
	
	messages := []string{
		"I'm not convinced everyone here is human.",
		"We need to eliminate threats quickly.",
		"Someone's been getting suspicious amounts of tokens.",
		"Time to make some hard decisions.",
		"The AI is among us - we need to find it.",
		"Look at the mining patterns - something's off.",
		"We can't afford to be passive here.",
		"Trust but verify - always verify.",
	}
	
	message := messages[a.rand.Intn(len(messages))]
	return &message
}

// Helper functions (similar to CautiousHuman but with different thresholds)
func (a *AggressiveHuman) findSuspiciousPlayers(gameState core.GameState, excludePlayerID string) []string {
	suspicious := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive {
			continue
		}
		
		suspicionScore := 0.0
		
		// More aggressive criteria
		if player.Tokens > 3 { // Lower threshold
			suspicionScore += 0.4
		}
		
		if player.ProjectMilestones < 3 && gameState.DayNumber > 1 { // Earlier detection
			suspicionScore += 0.3
		}
		
		// Look for token/milestone imbalance
		if player.ProjectMilestones > 0 {
			ratio := float64(player.Tokens) / float64(player.ProjectMilestones)
			if ratio > 2.0 { // Lower threshold
				suspicionScore += 0.3
			}
		}
		
		if suspicionScore > 0.3 { // Lower overall threshold
			suspicious = append(suspicious, id)
		}
	}
	
	return suspicious
}

func (a *AggressiveHuman) calculateSuspicionLevel(gameState core.GameState, playerID string) float64 {
	player := gameState.Players[playerID]
	if player == nil {
		return 0.0
	}
	
	suspicion := 0.0
	
	// High tokens relative to milestones (more sensitive)
	if player.ProjectMilestones > 0 {
		tokenRatio := float64(player.Tokens) / float64(player.ProjectMilestones)
		if tokenRatio > 2.0 { // Lower threshold
			suspicion += 0.5
		}
	} else if player.Tokens > 2 { // Tokens without milestones
		suspicion += 0.6
	}
	
	// Low contribution (more aggressive)
	if gameState.DayNumber > 1 && player.ProjectMilestones < 2 {
		suspicion += 0.4
	}
	
	// High absolute token count
	if player.Tokens > 5 {
		suspicion += 0.3
	}
	
	return suspicion
}

func (a *AggressiveHuman) findValuablePlayers(gameState core.GameState, excludePlayerID string) []string {
	valuable := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive {
			continue
		}
		
		value := 0.0
		
		// High contributors
		if player.ProjectMilestones > 2 { // Lower threshold
			value += 0.6
		}
		
		// Important roles
		if player.Role != nil {
			switch player.Role.Type {
			case core.RoleCTO, core.RoleCISO:
				value += 0.5
			case core.RoleCEO:
				value += 0.4
			}
		}
		
		// Low suspicion players
		if a.calculateSuspicionLevel(gameState, id) < 0.2 {
			value += 0.3
		}
		
		if value > 0.5 { // Lower threshold
			valuable = append(valuable, id)
		}
	}
	
	return valuable
}