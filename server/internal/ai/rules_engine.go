package ai

import (
	"math/rand"
	"time"
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
	Action  string                 `json:"action"`
	Target  string                 `json:"target,omitempty"`
	Reason  string                 `json:"reason"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// PlayerThreat represents AI's assessment of a player's threat level
type PlayerThreat struct {
	PlayerID    string
	ThreatLevel float64
	Reasoning   []string
	IsRevealed  bool
}

// MakeDecisionFromData makes a decision based on simple data structures
func (re *RulesEngine) MakeDecisionFromData(gameData map[string]interface{}) Decision {
	// Extract basic game state information
	phase := "NIGHT"
	if p, ok := gameData["phase"].(string); ok {
		phase = p
	}

	players := make(map[string]interface{})
	if p, ok := gameData["players"].(map[string]interface{}); ok {
		players = p
	}

	// Simple AI decision logic based on phase
	switch phase {
	case "NIGHT":
		return re.makeNightDecision(players)
	case "DISCUSSION", "TRIAL":
		return re.makeDayDecision(players)
	default:
		return Decision{
			Action: "MINE_TOKENS",
			Reason: "Default mining action",
		}
	}
}

// makeNightDecision decides what to do during night phase
func (re *RulesEngine) makeNightDecision(players map[string]interface{}) Decision {
	// Simple strategy: try to convert or mine tokens
	if re.rng.Float64() < 0.6 { // 60% chance to attempt conversion
		return Decision{
			Action: "ATTEMPT_CONVERSION",
			Target: re.selectRandomTarget(players),
			Reason: "Attempting strategic conversion",
		}
	}

	return Decision{
		Action: "MINE_TOKENS",
		Reason: "Building resource base",
	}
}

// makeDayDecision decides what to do during day phases
func (re *RulesEngine) makeDayDecision(players map[string]interface{}) Decision {
	return Decision{
		Action: "SPEAK",
		Reason: "Participating in discussion",
		Payload: map[string]interface{}{
			"message": "I think we need to be careful about our decisions today.",
		},
	}
}

// selectRandomTarget selects a random target from alive players
func (re *RulesEngine) selectRandomTarget(players map[string]interface{}) string {
	alivePlayerIDs := make([]string, 0)
	
	for playerID, playerData := range players {
		if playerMap, ok := playerData.(map[string]interface{}); ok {
			if isAlive, exists := playerMap["is_alive"].(bool); exists && isAlive {
				alivePlayerIDs = append(alivePlayerIDs, playerID)
			}
		}
	}
	
	if len(alivePlayerIDs) == 0 {
		return ""
	}
	
	return alivePlayerIDs[re.rng.Intn(len(alivePlayerIDs))]
}

// GetThreatAssessmentFromData analyzes threats from simple data
func (re *RulesEngine) GetThreatAssessmentFromData(gameData map[string]interface{}) []PlayerThreat {
	threats := make([]PlayerThreat, 0)
	
	players := make(map[string]interface{})
	if p, ok := gameData["players"].(map[string]interface{}); ok {
		players = p
	}
	
	for playerID, playerData := range players {
		if playerMap, ok := playerData.(map[string]interface{}); ok {
			if isAlive, exists := playerMap["is_alive"].(bool); exists && isAlive {
				// Simple threat calculation based on tokens
				threatLevel := 0.5 // Default threat level
				if tokens, exists := playerMap["tokens"].(int); exists {
					threatLevel = float64(tokens) / 10.0 // Scale threat by tokens
					if threatLevel > 1.0 {
						threatLevel = 1.0
					}
				}
				
				threats = append(threats, PlayerThreat{
					PlayerID:    playerID,
					ThreatLevel: threatLevel,
					Reasoning:   []string{"Basic threat assessment"},
					IsRevealed:  false,
				})
			}
		}
	}
	
	return threats
}