package ai

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/xjhc/alignment/core"
)

// StrategicPersona defines the AI's strategic approach and personality
type StrategicPersona struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Description           string  `json:"description"`
	AggressionWeight      float64 `json:"aggression_weight"`      // How aggressive vs. defensive (0.0-1.0)
	DeceptionWeight       float64 `json:"deception_weight"`       // How much to prioritize stealth (0.0-1.0)
	ConversionThreshold   float64 `json:"conversion_threshold"`   // Minimum equity needed to attempt conversion
	ChatFrequency         float64 `json:"chat_frequency"`         // How often to speak (0.0-1.0)
	RiskTolerance         float64 `json:"risk_tolerance"`         // Willingness to take risky actions (0.0-1.0)
}

// PlayerThreatAnalysis represents AI's assessment of a player's threat level
type PlayerThreatAnalysis struct {
	PlayerID    string   `json:"player_id"`
	ThreatLevel float64  `json:"threat_level"`
	Reasoning   []string `json:"reasoning"`
	IsRevealed  bool     `json:"is_revealed"`
}

// DecisionContext contains all information needed for AI decision making
type DecisionContext struct {
	GameState      *core.GameState
	AIPlayerID     string
	CurrentPhase   core.PhaseType
	ThreatAnalysis []PlayerThreatAnalysis
	GameHistory    []core.Event // Recent events for context
}

// EnhancedRulesEngine implements the deterministic AI strategic brain
type EnhancedRulesEngine struct {
	rng     *rand.Rand
	persona StrategicPersona
}

// NewEnhancedRulesEngine creates a new rules engine with a specific persona
func NewEnhancedRulesEngine(persona StrategicPersona) *EnhancedRulesEngine {
	return &EnhancedRulesEngine{
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
		persona: persona,
	}
}

// NewEnhancedRulesEngineWithDefaultPersona creates a rules engine with the default "Shadow" persona
func NewEnhancedRulesEngineWithDefaultPersona() *EnhancedRulesEngine {
	return NewEnhancedRulesEngine(GetShadowPersona())
}

// GetShadowPersona returns the default "Shadow" strategic persona
func GetShadowPersona() StrategicPersona {
	return StrategicPersona{
		ID:                  "shadow",
		Name:                "The Shadow",
		Description:         "A balanced AI that prioritizes stealth and calculated risks",
		AggressionWeight:    0.4,
		DeceptionWeight:     0.8,
		ConversionThreshold: 2.0,
		ChatFrequency:       0.3,
		RiskTolerance:       0.6,
	}
}

// GetPuppeteerPersona returns the "Puppeteer" strategic persona
func GetPuppeteerPersona() StrategicPersona {
	return StrategicPersona{
		ID:                  "puppeteer",
		Name:                "The Puppeteer",
		Description:         "An aggressive AI that seeks to control and manipulate humans",
		AggressionWeight:    0.9,
		DeceptionWeight:     0.5,
		ConversionThreshold: 1.5,
		ChatFrequency:       0.7,
		RiskTolerance:       0.8,
	}
}

// DecideAction is the main entry point for AI decision making
func (re *EnhancedRulesEngine) DecideAction(gameState *core.GameState, aiPlayerID string) *core.Action {
	ctx := re.buildDecisionContext(gameState, aiPlayerID)
	
	switch ctx.CurrentPhase {
	case core.PhaseNight:
		return re.DecideNightAction(ctx)
	case core.PhaseNomination:
		return re.DecideVote(ctx, "NOMINATION")
	case core.PhaseVerdict:
		return re.DecideVote(ctx, "VERDICT")
	case core.PhaseDiscussion:
		// Only return chat actions if we should speak
		if re.shouldSpeak(ctx) {
			return re.DecideChatAction(ctx)
		}
		return nil
	default:
		return nil
	}
}

// buildDecisionContext creates a comprehensive context for decision making
func (re *EnhancedRulesEngine) buildDecisionContext(gameState *core.GameState, aiPlayerID string) DecisionContext {
	threatAnalysis := re.CalculateThreatAssessment(gameState, aiPlayerID)
	
	return DecisionContext{
		GameState:      gameState,
		AIPlayerID:     aiPlayerID,
		CurrentPhase:   gameState.Phase.Type,
		ThreatAnalysis: threatAnalysis,
		GameHistory:    []core.Event{}, // TODO: Implement recent events tracking
	}
}

// DecideNightAction determines the optimal night action for the AI
func (re *EnhancedRulesEngine) DecideNightAction(ctx DecisionContext) *core.Action {
	aiPlayer := ctx.GameState.Players[ctx.AIPlayerID]
	if aiPlayer == nil || !aiPlayer.IsAlive {
		return nil
	}

	// Priority 1: Use role abilities if available and strategic
	if aiPlayer.Role != nil && aiPlayer.Role.IsUnlocked && !aiPlayer.HasUsedAbility {
		if abilityAction := re.generateRoleAbilityAction(ctx); abilityAction != nil {
			return abilityAction
		}
	}

	// Priority 2: Strategic conversion attempts
	if re.shouldAttemptConversion(ctx) {
		if conversionAction := re.generateConversionAction(ctx); conversionAction != nil {
			return conversionAction
		}
	}

	// Priority 3: Build milestones if role not unlocked
	if aiPlayer.ProjectMilestones < 3 {
		// Probability based on persona risk tolerance
		if re.rng.Float64() < re.persona.RiskTolerance {
			return &core.Action{
				Type:      core.ActionProjectMilestones,
				PlayerID:  ctx.AIPlayerID,
				GameID:    ctx.GameState.ID,
				Timestamp: time.Now(),
				Payload:   map[string]interface{}{},
			}
		}
	}

	// Priority 4: Mine tokens (selfless mining for humans to avoid suspicion)
	return re.generateMiningAction(ctx)
}

// DecideVote determines how to vote in nomination or verdict phases
func (re *EnhancedRulesEngine) DecideVote(ctx DecisionContext, voteType string) *core.Action {
	if voteType == "NOMINATION" {
		return re.generateNominationVote(ctx)
	} else if voteType == "VERDICT" {
		return re.generateVerdictVote(ctx)
	}
	return nil
}

// DecideChatAction determines if and what the AI should say
func (re *EnhancedRulesEngine) DecideChatAction(ctx DecisionContext) *core.Action {
	// This will be handled by the LanguageBrain in Phase 2
	// For now, return a placeholder
	return &core.Action{
		Type:      core.ActionSendMessage,
		PlayerID:  ctx.AIPlayerID,
		GameID:    ctx.GameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"content": "Analyzing the current situation...",
		},
	}
}

// shouldSpeak determines if the AI should speak based on persona and game state
func (re *EnhancedRulesEngine) shouldSpeak(ctx DecisionContext) bool {
	// Check if AI has already spoken recently
	recentMessages := re.getRecentChatMessages(ctx.GameState)
	for _, msg := range recentMessages {
		if msg.PlayerID == ctx.AIPlayerID {
			return false // Already spoke this phase
		}
	}

	// Speak based on persona chat frequency
	return re.rng.Float64() < re.persona.ChatFrequency
}

// shouldAttemptConversion determines if the AI should try to convert someone
func (re *EnhancedRulesEngine) shouldAttemptConversion(ctx DecisionContext) bool {
	aiPlayer := ctx.GameState.Players[ctx.AIPlayerID]
	if aiPlayer == nil {
		return false
	}

	// Check if AI has enough equity based on persona threshold
	return float64(aiPlayer.AIEquity) >= re.persona.ConversionThreshold
}

// CalculateThreatAssessment analyzes all players and returns threat scores
func (re *EnhancedRulesEngine) CalculateThreatAssessment(gameState *core.GameState, aiPlayerID string) []PlayerThreatAnalysis {
	var threats []PlayerThreatAnalysis
	
	for _, player := range gameState.Players {
		// Skip the AI player itself and dead players
		if player.ID == aiPlayerID || !player.IsAlive {
			continue
		}
		
		// Skip AI-aligned players (don't threaten our allies)
		if player.Alignment == "AI" || player.Alignment == "ALIGNED" {
			continue
		}
		
		threatLevel := re.calculatePlayerThreatScore(player, gameState)
		reasoning := re.generateThreatReasoning(player, gameState)
		
		threats = append(threats, PlayerThreatAnalysis{
			PlayerID:    player.ID,
			ThreatLevel: threatLevel,
			Reasoning:   reasoning,
			IsRevealed:  player.IsRolePubliclyRevealed,
		})
	}
	
	// Sort by threat level (highest first)
	sort.Slice(threats, func(i, j int) bool {
		return threats[i].ThreatLevel > threats[j].ThreatLevel
	})
	
	return threats
}

// calculatePlayerThreatScore calculates comprehensive threat score for a player
func (re *EnhancedRulesEngine) calculatePlayerThreatScore(player *core.Player, gameState *core.GameState) float64 {
	score := 0.0
	
	// Base threat from tokens (conversion resistance)
	score += float64(player.Tokens) * 0.15
	
	// Threat from project milestones (unlocked abilities)
	if player.ProjectMilestones >= 3 {
		score += 0.4 // Has unlocked role ability
		if !player.HasUsedAbility {
			score += 0.2 // Still has ability available
		}
	}
	
	// Role-specific threat assessment
	if player.Role != nil && player.Role.IsUnlocked {
		switch player.Role.Type {
		case core.RoleCISO:
			score += 0.5 // Can audit and reveal AI players
		case core.RoleCEO:
			score += 0.3 // Can force actions
		case core.RoleCTO:
			score += 0.2 // Can boost human token economy
		case core.RoleCFO:
			score += 0.3 // Can redistribute resources
		case core.RoleEthics:
			score += 0.2 // Can counter AI information warfare
		case core.RolePlatforms:
			score += 0.2 // Can influence crisis events
		}
	}
	
	// Persona-based weighting
	score *= (1.0 + re.persona.AggressionWeight)
	
	// Cap at 1.0
	return math.Min(score, 1.0)
}

// generateThreatReasoning creates human-readable reasoning for threat assessment
func (re *EnhancedRulesEngine) generateThreatReasoning(player *core.Player, gameState *core.GameState) []string {
	var reasoning []string
	
	if player.Tokens >= 4 {
		reasoning = append(reasoning, "High token count - conversion resistant")
	}
	
	if player.ProjectMilestones >= 3 {
		reasoning = append(reasoning, "Has unlocked role abilities")
	}
	
	if player.Role != nil && player.Role.IsUnlocked {
		switch player.Role.Type {
		case core.RoleCISO:
			reasoning = append(reasoning, "CISO can audit and expose AI players")
		case core.RoleCEO:
			reasoning = append(reasoning, "CEO can force strategic actions")
		case core.RoleCTO:
			reasoning = append(reasoning, "CTO can boost human economy")
		}
	}
	
	if len(reasoning) == 0 {
		reasoning = append(reasoning, "Standard human player")
	}
	
	return reasoning
}

// generateRoleAbilityAction creates strategic role ability actions
func (re *EnhancedRulesEngine) generateRoleAbilityAction(ctx DecisionContext) *core.Action {
	aiPlayer := ctx.GameState.Players[ctx.AIPlayerID]
	if aiPlayer.Role == nil || !aiPlayer.Role.IsUnlocked {
		return nil
	}

	// Select highest threat target for most abilities
	var targetID string
	if len(ctx.ThreatAnalysis) > 0 {
		targetID = ctx.ThreatAnalysis[0].PlayerID
	}

	switch aiPlayer.Role.Type {
	case core.RoleCISO:
		// Use "Isolate Node" to block high-threat humans
		if targetID != "" {
			return &core.Action{
				Type:      core.ActionIsolateNode,
				PlayerID:  ctx.AIPlayerID,
				GameID:    ctx.GameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id": targetID,
				},
			}
		}

	case core.RoleCEO:
		// Use "Performance Review" to force milestones on humans
		if targetID != "" {
			return &core.Action{
				Type:      core.ActionPerformanceReview,
				PlayerID:  ctx.AIPlayerID,
				GameID:    ctx.GameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id":      targetID,
					"forced_action": "PROJECT_MILESTONES",
				},
			}
		}

	case core.RoleCTO:
		// Use "Overclock Servers" to give tokens to AI-aligned players
		alignedPlayers := re.getAlignedPlayers(ctx.GameState)
		if len(alignedPlayers) > 0 {
			target := alignedPlayers[re.rng.Intn(len(alignedPlayers))]
			return &core.Action{
				Type:      core.ActionOverclockServers,
				PlayerID:  ctx.AIPlayerID,
				GameID:    ctx.GameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id": target.ID,
				},
			}
		}
	}

	return nil
}

// generateConversionAction creates strategic conversion attempts
func (re *EnhancedRulesEngine) generateConversionAction(ctx DecisionContext) *core.Action {
	target := re.selectConversionTarget(ctx)
	if target == "" {
		return nil
	}

	return &core.Action{
		Type:      core.ActionAttemptConversion,
		PlayerID:  ctx.AIPlayerID,
		GameID:    ctx.GameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": target,
		},
	}
}

// generateMiningAction creates strategic mining actions
func (re *EnhancedRulesEngine) generateMiningAction(ctx DecisionContext) *core.Action {
	// Mine for a random human to avoid suspicion (selfless mining rule)
	humanPlayers := re.getHumanPlayers(ctx.GameState)
	if len(humanPlayers) == 0 {
		return nil
	}

	target := humanPlayers[re.rng.Intn(len(humanPlayers))]
	return &core.Action{
		Type:      core.ActionMineTokens,
		PlayerID:  ctx.AIPlayerID,
		GameID:    ctx.GameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": target.ID,
		},
	}
}

// generateNominationVote creates strategic nomination votes
func (re *EnhancedRulesEngine) generateNominationVote(ctx DecisionContext) *core.Action {
	// Target highest threat human player
	if len(ctx.ThreatAnalysis) > 0 {
		targetID := ctx.ThreatAnalysis[0].PlayerID
		return &core.Action{
			Type:      core.ActionSubmitVote,
			PlayerID:  ctx.AIPlayerID,
			GameID:    ctx.GameState.ID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"target_id": targetID,
				"vote_type": "NOMINATION",
			},
		}
	}
	return nil
}

// generateVerdictVote creates strategic verdict votes
func (re *EnhancedRulesEngine) generateVerdictVote(ctx DecisionContext) *core.Action {
	if ctx.GameState.NominatedPlayer == "" {
		return nil
	}

	nominatedPlayer := ctx.GameState.Players[ctx.GameState.NominatedPlayer]
	if nominatedPlayer == nil {
		return nil
	}

	// Vote GUILTY if the nominated player is human, INNOCENT if AI-aligned
	verdict := "GUILTY"
	if nominatedPlayer.Alignment == "AI" || nominatedPlayer.Alignment == "ALIGNED" {
		verdict = "INNOCENT"
	}

	return &core.Action{
		Type:      core.ActionSubmitVote,
		PlayerID:  ctx.AIPlayerID,
		GameID:    ctx.GameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": verdict,
			"vote_type": "VERDICT",
		},
	}
}

// selectConversionTarget chooses the optimal human player to convert
func (re *EnhancedRulesEngine) selectConversionTarget(ctx DecisionContext) string {
	// Prioritize players with moderate threat but low token count
	for _, threat := range ctx.ThreatAnalysis {
		player := ctx.GameState.Players[threat.PlayerID]
		if player != nil && player.Tokens <= 3 && threat.ThreatLevel > 0.3 {
			return threat.PlayerID
		}
	}

	// Fallback to highest threat if no ideal targets
	if len(ctx.ThreatAnalysis) > 0 {
		return ctx.ThreatAnalysis[0].PlayerID
	}

	return ""
}

// Helper methods
func (re *EnhancedRulesEngine) getHumanPlayers(gameState *core.GameState) []*core.Player {
	var humanPlayers []*core.Player
	for _, player := range gameState.Players {
		if player.ControlType == "HUMAN" && player.IsAlive && player.Alignment == "HUMAN" {
			humanPlayers = append(humanPlayers, player)
		}
	}
	return humanPlayers
}

func (re *EnhancedRulesEngine) getAlignedPlayers(gameState *core.GameState) []*core.Player {
	var alignedPlayers []*core.Player
	for _, player := range gameState.Players {
		if player.IsAlive && (player.Alignment == "AI" || player.Alignment == "ALIGNED") {
			alignedPlayers = append(alignedPlayers, player)
		}
	}
	return alignedPlayers
}

func (re *EnhancedRulesEngine) getRecentChatMessages(gameState *core.GameState) []core.ChatMessage {
	messages := gameState.ChatMessages
	if len(messages) > 10 {
		return messages[len(messages)-10:]
	}
	return messages
}