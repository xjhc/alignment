package simulator

import (
	"math/rand"

	"github.com/xjhc/alignment/core"
)

// BotPersona defines the interface for AI bot behavior
type BotPersona interface {
	// GetID returns the unique identifier for this persona type
	GetID() string
	
	// GetDescription returns a human-readable description of this persona's behavior
	GetDescription() string
	
	// DecideAction determines what action this persona wants to take given the current game state
	// Returns nil if the persona doesn't want to take any action at this time
	DecideAction(gameState core.GameState, playerID string) *core.Action
	
	// ShouldNominate decides whether to nominate a player during nomination phase
	ShouldNominate(gameState core.GameState, playerID string) (bool, string)
	
	// DecideVote determines how to vote during voting phases
	DecideVote(gameState core.GameState, playerID string, nominatedPlayer string) string
	
	// DecideNightAction determines night actions to take
	DecideNightAction(gameState core.GameState, playerID string) *core.Action
	
	// DecideChatMessage determines whether to send a chat message and what to say
	DecideChatMessage(gameState core.GameState, playerID string) *string
}

// CautiousHuman persona - prioritizes project milestones and group consensus
type CautiousHuman struct {
	rand *rand.Rand
}

func NewCautiousHuman(seed int64) *CautiousHuman {
	return &CautiousHuman{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (c *CautiousHuman) GetID() string {
	return "cautious_human"
}

func (c *CautiousHuman) GetDescription() string {
	return "Prioritizes contributing to project milestones. Avoids nominating without strong data. Votes with the majority."
}

func (c *CautiousHuman) DecideAction(gameState core.GameState, playerID string) *core.Action {
	player := gameState.Players[playerID]
	if player == nil || !player.IsAlive {
		return nil
	}

	switch gameState.Phase.Type {
	case core.PhaseNomination:
		if shouldNominate, target := c.ShouldNominate(gameState, playerID); shouldNominate {
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
			target := c.DecideVote(gameState, playerID, gameState.NominatedPlayer)
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
		return c.DecideNightAction(gameState, playerID)
	case core.PhaseSitrep:
		// Respond to pulse check
		response := "All systems operational"
		if c.rand.Float64() < 0.1 { // 10% chance of different response
			responses := []string{"Status nominal", "Running diagnostics", "No anomalies detected"}
			response = responses[c.rand.Intn(len(responses))]
		}
		return &core.Action{
			Type:     core.ActionSubmitPulseCheck,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"response": response,
			},
		}
	}

	// Check for chat message opportunity
	if message := c.DecideChatMessage(gameState, playerID); message != nil {
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

func (c *CautiousHuman) ShouldNominate(gameState core.GameState, playerID string) (bool, string) {
	// Cautious humans only nominate if they have strong evidence
	// Look for players with suspicious behavior patterns
	suspiciousPlayers := c.findSuspiciousPlayers(gameState, playerID)
	
	if len(suspiciousPlayers) == 0 {
		return false, ""
	}
	
	// Only nominate if confidence is high (>70% suspicious behavior)
	for _, suspiciousPlayer := range suspiciousPlayers {
		if c.calculateSuspicionLevel(gameState, suspiciousPlayer) > 0.7 {
			return true, suspiciousPlayer
		}
	}
	
	return false, ""
}

func (c *CautiousHuman) DecideVote(gameState core.GameState, playerID string, nominatedPlayer string) string {
	// Cautious humans tend to vote with the majority
	if gameState.VoteState == nil {
		return nominatedPlayer // Vote guilty if no voting state yet
	}
	
	// Count current votes
	guiltyVotes := 0
	innocentVotes := 0
	
	for _, vote := range gameState.VoteState.Votes {
		if vote == nominatedPlayer {
			guiltyVotes++
		} else {
			innocentVotes++
		}
	}
	
	// If majority is voting guilty, join them
	if guiltyVotes > innocentVotes {
		return nominatedPlayer
	}
	
	// If majority is voting innocent, abstain or vote innocent
	if innocentVotes > guiltyVotes {
		return "" // Abstain
	}
	
	// If tied or no votes yet, make decision based on suspicion level
	suspicionLevel := c.calculateSuspicionLevel(gameState, nominatedPlayer)
	if suspicionLevel > 0.6 {
		return nominatedPlayer // Vote guilty
	}
	
	return "" // Abstain
}

func (c *CautiousHuman) DecideNightAction(gameState core.GameState, playerID string) *core.Action {
	player := gameState.Players[playerID]
	
	// Always try to mine for others (selfless mining)
	alivePlayers := make([]string, 0)
	for id, p := range gameState.Players {
		if p.IsAlive && id != playerID {
			alivePlayers = append(alivePlayers, id)
		}
	}
	
	if len(alivePlayers) > 0 {
		// Mine for a random other player
		beneficiary := alivePlayers[c.rand.Intn(len(alivePlayers))]
		return &core.Action{
			Type:     core.ActionMineTokens,
			PlayerID: playerID,
			GameID:   gameState.ID,
			Payload: map[string]interface{}{
				"beneficiary_id": beneficiary,
			},
		}
	}
	
	// Use role abilities if available
	if player.Role != nil && player.Role.IsUnlocked && !player.HasUsedAbility {
		switch player.Role.Type {
		case core.RoleCISO:
			// Investigate suspicious players
			suspiciousPlayers := c.findSuspiciousPlayers(gameState, playerID)
			if len(suspiciousPlayers) > 0 {
				target := suspiciousPlayers[0]
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
			// Protect valuable players
			valuablePlayers := c.findValuablePlayers(gameState, playerID)
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
		}
	}
	
	return nil
}

func (c *CautiousHuman) DecideChatMessage(gameState core.GameState, playerID string) *string {
	// Cautious humans don't chat much, only when necessary
	if c.rand.Float64() > 0.1 { // Only 10% chance to chat
		return nil
	}
	
	messages := []string{
		"Let's focus on completing our project milestones.",
		"We need to be careful about rushing to judgment.",
		"Has anyone noticed any unusual behavior patterns?",
		"We should work together on this.",
		"I'm seeing some concerning metrics in the system.",
	}
	
	message := messages[c.rand.Intn(len(messages))]
	return &message
}

// Helper functions
func (c *CautiousHuman) findSuspiciousPlayers(gameState core.GameState, excludePlayerID string) []string {
	suspicious := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive {
			continue
		}
		
		// Look for suspicious patterns:
		// - High token count (might be getting help from AI)
		// - Low project milestones (not contributing)
		// - Unusual voting patterns
		
		suspicionScore := 0.0
		
		if player.Tokens > 5 {
			suspicionScore += 0.3
		}
		
		if player.ProjectMilestones < 2 && gameState.DayNumber > 2 {
			suspicionScore += 0.4
		}
		
		if suspicionScore > 0.5 {
			suspicious = append(suspicious, id)
		}
	}
	
	return suspicious
}

func (c *CautiousHuman) calculateSuspicionLevel(gameState core.GameState, playerID string) float64 {
	player := gameState.Players[playerID]
	if player == nil {
		return 0.0
	}
	
	suspicion := 0.0
	
	// High tokens relative to milestones
	if player.ProjectMilestones > 0 {
		tokenRatio := float64(player.Tokens) / float64(player.ProjectMilestones)
		if tokenRatio > 3.0 {
			suspicion += 0.4
		}
	}
	
	// Low contribution to project
	if gameState.DayNumber > 2 && player.ProjectMilestones < 2 {
		suspicion += 0.3
	}
	
	// Other factors can be added here
	return suspicion
}

func (c *CautiousHuman) findValuablePlayers(gameState core.GameState, excludePlayerID string) []string {
	valuable := make([]string, 0)
	
	for id, player := range gameState.Players {
		if id == excludePlayerID || !player.IsAlive {
			continue
		}
		
		// Valuable players have high project milestones and important roles
		value := 0.0
		
		if player.ProjectMilestones > 3 {
			value += 0.5
		}
		
		if player.Role != nil {
			switch player.Role.Type {
			case core.RoleCTO, core.RoleCISO:
				value += 0.4
			case core.RoleCEO:
				value += 0.3
			}
		}
		
		if value > 0.6 {
			valuable = append(valuable, id)
		}
	}
	
	return valuable
}