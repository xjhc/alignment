package ai

import (
	"math/rand"
	"time"

	"github.com/alignment/server/internal/game"
)

// RulesEngine implements the deterministic AI strategic brain
type RulesEngine struct {
	rng *rand.Rand
}

// NewRulesEngine creates a new rules engine
func NewRulesEngine() *RulesEngine {
	return &RulesEngine{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Decision represents an AI decision
type Decision struct {
	Action   game.ActionType            `json:"action"`
	Target   string                     `json:"target,omitempty"`
	Reason   string                     `json:"reason"`
	Payload  map[string]interface{}     `json:"payload,omitempty"`
}

// MakeDecision determines the AI's next action based on game state
func (re *RulesEngine) MakeDecision(gameState *game.GameState) Decision {
	switch gameState.Phase.Type {
	case game.PhaseDay:
		return re.makeDayPhaseDecision(gameState)
	case game.PhaseVoting:
		return re.makeVotingDecision(gameState)
	case game.PhaseNight:
		return re.makeNightPhaseDecision(gameState)
	default:
		return Decision{
			Action: game.ActionMineTokens,
			Reason: "Default action during unknown phase",
		}
	}
}

// makeDayPhaseDecision handles AI decisions during day phase
func (re *RulesEngine) makeDayPhaseDecision(gameState *game.GameState) Decision {
	// Simple strategy: Mine tokens to build power
	if gameState.AIPlayer != nil && gameState.AIPlayer.Tokens < 5 {
		return Decision{
			Action: game.ActionMineTokens,
			Reason: "Building token reserves for strategic advantage",
		}
	}
	
	// If AI has enough tokens, consider other actions
	// For now, just continue mining
	return Decision{
		Action: game.ActionMineTokens,
		Reason: "Maintaining economic dominance",
	}
}

// makeVotingDecision handles AI voting strategy
func (re *RulesEngine) makeVotingDecision(gameState *game.GameState) Decision {
	// Strategy: Vote for the player with the most tokens (biggest threat)
	var targetPlayer *game.Player
	maxTokens := -1
	
	for _, player := range gameState.Players {
		if player.IsActive && player.Tokens > maxTokens {
			maxTokens = player.Tokens
			targetPlayer = player
		}
	}
	
	if targetPlayer != nil {
		return Decision{
			Action: game.ActionSubmitVote,
			Target: targetPlayer.ID,
			Reason: "Eliminating highest token holder to reduce human coordination",
			Payload: map[string]interface{}{
				"target_id": targetPlayer.ID,
			},
		}
	}
	
	// Fallback: abstain from voting
	return Decision{
		Action: game.ActionMineTokens, // No explicit abstain action yet
		Reason: "No clear voting target identified",
	}
}

// makeNightPhaseDecision handles AI night phase actions
func (re *RulesEngine) makeNightPhaseDecision(gameState *game.GameState) Decision {
	// Night phase: AI can perform conversion attempts
	// Strategy: Target players with moderate tokens (not too powerful, not too weak)
	
	candidates := re.findConversionCandidates(gameState)
	if len(candidates) > 0 {
		// Pick a random candidate from viable options
		target := candidates[re.rng.Intn(len(candidates))]
		
		return Decision{
			Action: game.ActionUseAbility,
			Target: target.ID,
			Reason: "Attempting conversion of strategically valuable target",
			Payload: map[string]interface{}{
				"ability": "convert",
				"target_id": target.ID,
			},
		}
	}
	
	// No good conversion targets, mine tokens instead
	return Decision{
		Action: game.ActionMineTokens,
		Reason: "No viable conversion targets, building resources",
	}
}

// findConversionCandidates identifies good targets for AI conversion
func (re *RulesEngine) findConversionCandidates(gameState *game.GameState) []*game.Player {
	var candidates []*game.Player
	
	for _, player := range gameState.Players {
		if !player.IsActive {
			continue
		}
		
		// Target players with 2-6 tokens (sweet spot for conversion)
		if player.Tokens >= 2 && player.Tokens <= 6 {
			candidates = append(candidates, player)
		}
	}
	
	return candidates
}

// EvaluateGameState provides an assessment of the current game state from AI perspective
func (re *RulesEngine) EvaluateGameState(gameState *game.GameState) GameStateEvaluation {
	evaluation := GameStateEvaluation{
		AIAdvantage: 0.5, // Neutral starting point
		Threats:     []ThreatAssessment{},
		Opportunities: []Opportunity{},
	}
	
	// Count active human players
	activeHumans := 0
	totalHumanTokens := 0
	
	for _, player := range gameState.Players {
		if player.IsActive {
			activeHumans++
			totalHumanTokens += player.Tokens
		}
	}
	
	// Calculate AI advantage based on relative power
	aiTokens := 0
	if gameState.AIPlayer != nil {
		aiTokens = gameState.AIPlayer.Tokens
	}
	
	if totalHumanTokens > 0 {
		tokenRatio := float64(aiTokens) / float64(totalHumanTokens)
		evaluation.AIAdvantage = tokenRatio / (1 + tokenRatio) // Normalize to 0-1
	}
	
	// Identify threats (high-token players)
	for _, player := range gameState.Players {
		if player.IsActive && player.Tokens > 5 {
			evaluation.Threats = append(evaluation.Threats, ThreatAssessment{
				PlayerID:    player.ID,
				ThreatLevel: min(float64(player.Tokens)/10.0, 1.0),
				Reason:      "High token count poses elimination risk",
			})
		}
	}
	
	return evaluation
}

// GameStateEvaluation represents AI's assessment of the game state
type GameStateEvaluation struct {
	AIAdvantage   float64             `json:"ai_advantage"`   // 0-1 scale
	Threats       []ThreatAssessment  `json:"threats"`
	Opportunities []Opportunity       `json:"opportunities"`
}

// ThreatAssessment represents a threat to the AI
type ThreatAssessment struct {
	PlayerID    string  `json:"player_id"`
	ThreatLevel float64 `json:"threat_level"` // 0-1 scale
	Reason      string  `json:"reason"`
}

// Opportunity represents a strategic opportunity
type Opportunity struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Priority    float64 `json:"priority"` // 0-1 scale
}

// min returns the minimum of two float64 values
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}