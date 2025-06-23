package ai

import (
	"fmt"
	"math/rand"
	"sort"
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
	threats := re.calculateThreatAssessment(players)
	aiPlayer := re.findAIPlayer(players)

	if aiPlayer == nil {
		return Decision{
			Action: "MINE_TOKENS",
			Target: re.selectRandomTarget(players),
			Reason: "No AI player found - defaulting to mining",
		}
	}

	// Prioritize conversion if AI has sufficient equity
	if aiEquity, exists := aiPlayer["ai_equity"].(int); exists && aiEquity >= 2 {
		// Find best conversion target (low token count, high threat)
		bestTarget := re.selectConversionTarget(threats)
		if bestTarget != "" {
			return Decision{
				Action: "ATTEMPT_CONVERSION",
				Target: bestTarget,
				Reason: "Converting high-priority threat target",
				Payload: map[string]interface{}{
					"ai_equity": aiEquity,
				},
			}
		}
	}

	// If tokens are low, prioritize mining over other actions
	if tokens, exists := aiPlayer["tokens"].(int); exists && tokens < 2 {
		miningTarget := re.selectMiningTarget(players)
		return Decision{
			Action: "MINE_TOKENS",
			Target: miningTarget,
			Reason: "Building resource base - low token count",
			Payload: map[string]interface{}{
				"target_name": re.getPlayerName(players, miningTarget),
			},
		}
	}

	// Check if we should use role ability
	if milestones, exists := aiPlayer["project_milestones"].(int); exists && milestones >= 3 {
		if re.rng.Float64() < 0.3 { // 30% chance to use ability
			return Decision{
				Action: "USE_ABILITY",
				Target: re.selectAbilityTarget(threats),
				Reason: "Using role ability strategically",
			}
		}
	}

	// Default to mining for someone else (selfless mining rule)
	miningTarget := re.selectMiningTarget(players)
	return Decision{
		Action: "MINE_TOKENS",
		Target: miningTarget,
		Reason: "Default mining strategy",
		Payload: map[string]interface{}{
			"target_name": re.getPlayerName(players, miningTarget),
		},
	}
}

// makeDayDecision decides what to do during day phases
func (re *RulesEngine) makeDayDecision(players map[string]interface{}) Decision {
	threats := re.calculateThreatAssessment(players)

	// During voting phases, select target based on threat assessment
	if len(threats) > 0 {
		// Target highest threat non-AI player
		for _, threat := range threats {
			if playerData, exists := players[threat.PlayerID]; exists {
				if playerMap, ok := playerData.(map[string]interface{}); ok {
					// Avoid targeting other AI players
					if alignment, exists := playerMap["alignment"].(string); !exists || alignment != "ALIGNED" {
						return Decision{
							Action: "VOTE",
							Target: threat.PlayerID,
							Reason: fmt.Sprintf("Targeting high threat player: %.2f threat level", threat.ThreatLevel),
							Payload: map[string]interface{}{
								"threat_level": threat.ThreatLevel,
								"player_name":  re.getPlayerName(players, threat.PlayerID),
							},
						}
					}
				}
			}
		}
	}

	return Decision{
		Action: "SPEAK",
		Reason: "Participating in discussion to maintain cover",
		Payload: map[string]interface{}{
			"message": "I agree we need to be strategic about identifying the AI threat.",
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

// calculateThreatAssessment creates a prioritized threat assessment
func (re *RulesEngine) calculateThreatAssessment(players map[string]interface{}) []PlayerThreat {
	threats := make([]PlayerThreat, 0)

	for playerID, playerData := range players {
		if playerMap, ok := playerData.(map[string]interface{}); ok {
			if isAlive, exists := playerMap["is_alive"].(bool); exists && isAlive {
				// Skip AI aligned players
				if alignment, exists := playerMap["alignment"].(string); exists && alignment == "ALIGNED" {
					continue
				}

				threatLevel := re.calculatePlayerThreatLevel(playerMap)
				reasoning := re.generateThreatReasoning(playerMap)

				threats = append(threats, PlayerThreat{
					PlayerID:    playerID,
					ThreatLevel: threatLevel,
					Reasoning:   reasoning,
					IsRevealed:  false,
				})
			}
		}
	}

	// Sort by threat level (highest first)
	sort.Slice(threats, func(i, j int) bool {
		return threats[i].ThreatLevel > threats[j].ThreatLevel
	})

	return threats
}

// calculatePlayerThreatLevel calculates how threatening a player is
func (re *RulesEngine) calculatePlayerThreatLevel(playerMap map[string]interface{}) float64 {
	threatLevel := 0.0

	// High token count = high threat (they can defend against conversion)
	if tokens, exists := playerMap["tokens"].(int); exists {
		threatLevel += float64(tokens) * 0.3
	}

	// High milestones = high threat (they have abilities)
	if milestones, exists := playerMap["project_milestones"].(int); exists {
		threatLevel += float64(milestones) * 0.2
	}

	// Cap threat level at 1.0
	if threatLevel > 1.0 {
		threatLevel = 1.0
	}

	return threatLevel
}

// generateThreatReasoning creates reasoning for threat assessment
func (re *RulesEngine) generateThreatReasoning(playerMap map[string]interface{}) []string {
	reasoning := make([]string, 0)

	if tokens, exists := playerMap["tokens"].(int); exists && tokens >= 3 {
		reasoning = append(reasoning, "High token count - conversion resistant")
	}

	if milestones, exists := playerMap["project_milestones"].(int); exists && milestones >= 3 {
		reasoning = append(reasoning, "Has unlocked role abilities")
	}

	if len(reasoning) == 0 {
		reasoning = append(reasoning, "Standard threat assessment")
	}

	return reasoning
}

// findAIPlayer locates the AI-aligned player
func (re *RulesEngine) findAIPlayer(players map[string]interface{}) map[string]interface{} {
	for _, playerData := range players {
		if playerMap, ok := playerData.(map[string]interface{}); ok {
			if alignment, exists := playerMap["alignment"].(string); exists && alignment == "ALIGNED" {
				if isAlive, exists := playerMap["is_alive"].(bool); exists && isAlive {
					return playerMap
				}
			}
		}
	}
	return nil
}

// selectConversionTarget selects the best target for conversion
func (re *RulesEngine) selectConversionTarget(threats []PlayerThreat) string {
	// Target players with moderate threat but low token count
	for _, threat := range threats {
		if threat.ThreatLevel > 0.3 && threat.ThreatLevel < 0.8 {
			return threat.PlayerID
		}
	}

	// Fallback to highest threat if no moderate targets
	if len(threats) > 0 {
		return threats[0].PlayerID
	}

	return ""
}

// selectMiningTarget selects who to mine tokens for (cannot be self)
func (re *RulesEngine) selectMiningTarget(players map[string]interface{}) string {
	candidates := make([]string, 0)

	for playerID, playerData := range players {
		if playerMap, ok := playerData.(map[string]interface{}); ok {
			if isAlive, exists := playerMap["is_alive"].(bool); exists && isAlive {
				// Don't mine for AI players - mine for humans only
				if alignment, exists := playerMap["alignment"].(string); !exists || alignment != "ALIGNED" {
					candidates = append(candidates, playerID)
				}
			}
		}
	}

	// Mine for a random human player (selfless mining rule)
	if len(candidates) > 0 {
		return candidates[re.rng.Intn(len(candidates))]
	}

	// If no human targets, don't mine (shouldn't happen in normal game)
	return ""
}

// selectAbilityTarget selects target for role ability use
func (re *RulesEngine) selectAbilityTarget(threats []PlayerThreat) string {
	// Target highest threat for abilities
	if len(threats) > 0 {
		return threats[0].PlayerID
	}
	return ""
}

// getPlayerName retrieves player name from the players map
func (re *RulesEngine) getPlayerName(players map[string]interface{}, playerID string) string {
	if playerData, exists := players[playerID]; exists {
		if playerMap, ok := playerData.(map[string]interface{}); ok {
			if name, exists := playerMap["name"].(string); exists {
				return name
			}
		}
	}
	return "Unknown Player"
}
