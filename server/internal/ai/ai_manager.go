package ai

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// AIManager handles AI player decisions and actions
type AIManager struct {
	gameState        *core.GameState
	rulesEngine      *RulesEngine      // Legacy rules engine
	enhancedEngine   *EnhancedRulesEngine // New enhanced rules engine
	languageBrain    *LanguageBrain    // For LLM-powered chat
	aiActors         map[string]*AIActor // Map of playerID -> AIActor
	useActors        bool              // Whether to use new AIActor system
}

// NewAIManager creates a new AI manager
func NewAIManager(gameState *core.GameState) *AIManager {
	// Initialize language brain with defaults (will fall back to mock if no API key)
	languageBrain, err := NewLanguageBrainWithDefaults()
	if err != nil {
		log.Printf("Warning: Failed to initialize LanguageBrain: %v", err)
	}

	manager := &AIManager{
		gameState:      gameState,
		rulesEngine:    NewRulesEngineWithDefaultPersona(),
		enhancedEngine: NewEnhancedRulesEngineWithDefaultPersona(),
		languageBrain:  languageBrain,
		aiActors:       make(map[string]*AIActor),
		useActors:      true, // Enable new AIActor system
	}

	// Initialize AI actors for existing AI players
	manager.initializeAIActors()

	return manager
}

// GetAIPlayers returns all AI-controlled players in the game
func (aim *AIManager) GetAIPlayers() []*core.Player {
	var aiPlayers []*core.Player
	for _, player := range aim.gameState.Players {
		if player.ControlType == "AI" && player.IsAlive {
			aiPlayers = append(aiPlayers, player)
		}
	}
	return aiPlayers
}

// ProcessAIActions determines what actions AI players should take based on the current phase
func (aim *AIManager) ProcessAIActions() []core.Action {
	if aim.useActors {
		return aim.processAIActionsWithActors()
	}
	return aim.processAIActionsLegacy()
}

// processAIActionsWithActors uses the new AIActor system
func (aim *AIManager) processAIActionsWithActors() []core.Action {
	var actions []core.Action
	
	// Trigger AI actors based on current phase
	trigger := aim.getPhaseBasedTrigger()
	if trigger == "" {
		return actions
	}

	// Collect actions from all AI actors
	resultChannels := make(map[string]chan *core.Action)
	for playerID, actor := range aim.aiActors {
		if actor.GetStatus() == "active" {
			resultChannels[playerID] = actor.TriggerAction(AITrigger(trigger), aim.gameState, nil)
		}
	}

	// Collect results with timeout
	timeout := time.After(3 * time.Second)
	for playerID, resultChan := range resultChannels {
		select {
		case action := <-resultChan:
			if action != nil {
				log.Printf("[AIManager] Received action %s from AI player %s", action.Type, playerID)
				actions = append(actions, *action)
			}
		case <-timeout:
			log.Printf("[AIManager] Timeout waiting for AI action from player %s", playerID)
		}
	}

	return actions
}

// processAIActionsLegacy uses the original system for backward compatibility
func (aim *AIManager) processAIActionsLegacy() []core.Action {
	var actions []core.Action
	aiPlayers := aim.GetAIPlayers()
	
	if len(aiPlayers) == 0 {
		return actions
	}

	// Use the enhanced rules engine for strategic decisions
	for _, aiPlayer := range aiPlayers {
		action := aim.enhancedEngine.DecideAction(aim.gameState, aiPlayer.ID)
		if action != nil {
			actions = append(actions, *action)
		}
	}

	return actions
}

// ProcessAIActionsLegacy determines what actions AI players should take using the legacy engine
func (aim *AIManager) ProcessAIActionsLegacy() []core.Action {
	var actions []core.Action
	aiPlayers := aim.GetAIPlayers()
	
	if len(aiPlayers) == 0 {
		return actions
	}

	currentPhase := aim.gameState.Phase.Type

	for _, aiPlayer := range aiPlayers {
		switch currentPhase {
		case core.PhaseNight:
			action := aim.generateNightAction(aiPlayer)
			if action != nil {
				actions = append(actions, *action)
			}
		case core.PhaseNomination:
			action := aim.generateVoteAction(aiPlayer, "NOMINATION")
			if action != nil {
				actions = append(actions, *action)
			}
		case core.PhaseVerdict:
			action := aim.generateVoteAction(aiPlayer, "VERDICT")
			if action != nil {
				actions = append(actions, *action)
			}
		case core.PhaseDiscussion:
			action := aim.generateChatAction(aiPlayer)
			if action != nil {
				actions = append(actions, *action)
			}
		}
	}

	return actions
}

// generateNightAction creates a night action for the AI player
func (aim *AIManager) generateNightAction(aiPlayer *core.Player) *core.Action {
	// Strategic AI night action selection
	
	// First priority: Use role abilities if available and useful
	if aiPlayer.Role != nil && aiPlayer.Role.IsUnlocked && !aiPlayer.HasUsedAbility {
		if abilityAction := aim.generateRoleAbilityAction(aiPlayer); abilityAction != nil {
			return abilityAction
		}
	}
	
	// Second priority: Work on milestones if role not unlocked
	if aiPlayer.ProjectMilestones < 3 {
		// 70% chance to work on milestones when role is locked
		if rand.Float64() < 0.7 {
			return &core.Action{
				Type:      core.ActionProjectMilestones,
				PlayerID:  aiPlayer.ID,
				GameID:    aim.gameState.ID,
				Timestamp: time.Now(),
				Payload:   map[string]interface{}{},
			}
		}
	}

	// Third priority: Strategic conversion attempts
	humanPlayers := aim.getHumanPlayers()
	if len(humanPlayers) == 0 {
		return nil
	}

	// Select best conversion target based on strategy
	targetPlayer := aim.selectBestConversionTarget(humanPlayers)
	if targetPlayer == nil {
		return nil
	}

	return &core.Action{
		Type:      core.ActionAttemptConversion,
		PlayerID:  aiPlayer.ID,
		GameID:    aim.gameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": targetPlayer.ID,
		},
	}
}

// generateVoteAction creates a voting action for the AI player
func (aim *AIManager) generateVoteAction(aiPlayer *core.Player, voteType string) *core.Action {
	var targetID string

	if voteType == "NOMINATION" {
		// For nomination votes, select a random human player to nominate
		humanPlayers := aim.getHumanPlayers()
		if len(humanPlayers) > 0 {
			targetPlayer := humanPlayers[rand.Intn(len(humanPlayers))]
			targetID = targetPlayer.ID
		}
	} else if voteType == "VERDICT" {
		// For verdict votes, vote INNOCENT to try to save the nominated player
		// (if they're AI-aligned) or GUILTY (if they're human)
		if aim.gameState.NominatedPlayer != "" {
			nominatedPlayer := aim.gameState.Players[aim.gameState.NominatedPlayer]
			if nominatedPlayer != nil {
				// Vote GUILTY if the nominated player is human
				if nominatedPlayer.Alignment == "HUMAN" {
					targetID = "GUILTY"
				} else {
					targetID = "INNOCENT"
				}
			}
		}
	}

	if targetID == "" {
		return nil
	}

	return &core.Action{
		Type:      core.ActionSubmitVote,
		PlayerID:  aiPlayer.ID,
		GameID:    aim.gameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id":  targetID,
			"vote_type":  voteType,
		},
	}
}

// generateChatAction creates a chat action for the AI player
func (aim *AIManager) generateChatAction(aiPlayer *core.Player) *core.Action {
	// Simple chat: just say hello once per discussion phase
	// Check if the AI has already spoken this phase by looking at recent chat messages
	recentMessages := aim.getRecentChatMessages()
	for _, msg := range recentMessages {
		if msg.PlayerID == aiPlayer.ID {
			// AI has already spoken this phase
			return nil
		}
	}

	return &core.Action{
		Type:      core.ActionSendMessage,
		PlayerID:  aiPlayer.ID,
		GameID:    aim.gameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": "Hello world! I'm analyzing the situation...",
		},
	}
}

// getHumanPlayers returns all living human players
func (aim *AIManager) getHumanPlayers() []*core.Player {
	var humanPlayers []*core.Player
	for _, player := range aim.gameState.Players {
		if player.ControlType == "HUMAN" && player.IsAlive && player.Alignment == "HUMAN" {
			humanPlayers = append(humanPlayers, player)
		}
	}
	return humanPlayers
}

// getRecentChatMessages returns chat messages from the current discussion phase
func (aim *AIManager) getRecentChatMessages() []core.ChatMessage {
	// For simplicity, return the last 10 messages
	messages := aim.gameState.ChatMessages
	if len(messages) > 10 {
		return messages[len(messages)-10:]
	}
	return messages
}

// generateRoleAbilityAction creates strategic role ability actions for AI players
func (aim *AIManager) generateRoleAbilityAction(aiPlayer *core.Player) *core.Action {
	if aiPlayer.Role == nil || !aiPlayer.Role.IsUnlocked {
		return nil
	}

	humanPlayers := aim.getHumanPlayers()
	if len(humanPlayers) == 0 {
		return nil
	}

	// Strategic role ability usage based on role type
	switch aiPlayer.Role.Type {
	case core.RoleCISO:
		// CISO: Use "Isolate Node" to block high-threat humans
		target := aim.selectHighThreatTarget(humanPlayers)
		if target != nil {
			return &core.Action{
				Type:      core.ActionIsolateNode,
				PlayerID:  aiPlayer.ID,
				GameID:    aim.gameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id": target.ID,
				},
			}
		}

	case core.RoleCEO:
		// CEO: Use "Performance Review" to force milestones on humans
		target := aim.selectBestControlTarget(humanPlayers)
		if target != nil {
			return &core.Action{
				Type:      core.ActionPerformanceReview,
				PlayerID:  aiPlayer.ID,
				GameID:    aim.gameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id":      target.ID,
					"forced_action": "PROJECT_MILESTONES",
				},
			}
		}

	case core.RoleCTO:
		// CTO: Use "Overclock Servers" to give tokens to AI-aligned players
		alignedPlayers := aim.getAlignedPlayers()
		if len(alignedPlayers) > 0 {
			target := alignedPlayers[rand.Intn(len(alignedPlayers))]
			return &core.Action{
				Type:      core.ActionOverclockServers,
				PlayerID:  aiPlayer.ID,
				GameID:    aim.gameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id": target.ID,
				},
			}
		}

	case core.RoleCFO:
		// CFO: Use "Reallocate Budget" to redistribute tokens strategically
		richHuman := aim.selectRichestHuman(humanPlayers)
		poorAligned := aim.selectPoorestAligned()
		if richHuman != nil && poorAligned != nil {
			return &core.Action{
				Type:      core.ActionReallocateBudget,
				PlayerID:  aiPlayer.ID,
				GameID:    aim.gameState.ID,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"from_player": richHuman.ID,
					"to_player":   poorAligned.ID,
					"amount":      2,
				},
			}
		}

	case core.RolePlatforms:
		// VP Platforms: Use "Pivot" to select favorable crisis
		return &core.Action{
			Type:      core.ActionPivot,
			PlayerID:  aiPlayer.ID,
			GameID:    aim.gameState.ID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"selected_crisis": "AI_EFFICIENCY_SURGE", // Favorable to AI
			},
		}

	case core.RoleEthics:
		// VP Ethics: Use "Deploy Hotfix" to redact unfavorable SITREP sections
		return &core.Action{
			Type:      core.ActionDeployHotfix,
			PlayerID:  aiPlayer.ID,
			GameID:    aim.gameState.ID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"redaction_target": "ai_activity_log", // Hide AI activities
			},
		}
	}

	return nil
}

// selectBestConversionTarget chooses the optimal human player to convert
func (aim *AIManager) selectBestConversionTarget(humanPlayers []*core.Player) *core.Player {
	if len(humanPlayers) == 0 {
		return nil
	}

	// Priority 1: Target players with low tokens (easier to convert)
	var lowTokenTargets []*core.Player
	for _, player := range humanPlayers {
		if player.Tokens <= 2 {
			lowTokenTargets = append(lowTokenTargets, player)
		}
	}

	if len(lowTokenTargets) > 0 {
		return lowTokenTargets[rand.Intn(len(lowTokenTargets))]
	}

	// Priority 2: Target players with high AI equity (closer to conversion)
	var highEquityTargets []*core.Player
	for _, player := range humanPlayers {
		if player.AIEquity >= 2 {
			highEquityTargets = append(highEquityTargets, player)
		}
	}

	if len(highEquityTargets) > 0 {
		return highEquityTargets[rand.Intn(len(highEquityTargets))]
	}

	// Fallback: Random selection
	return humanPlayers[rand.Intn(len(humanPlayers))]
}

// selectHighThreatTarget selects a human player that poses the highest threat
func (aim *AIManager) selectHighThreatTarget(humanPlayers []*core.Player) *core.Player {
	if len(humanPlayers) == 0 {
		return nil
	}

	// Priority to CISO (can audit), then high token players
	for _, player := range humanPlayers {
		if player.Role != nil && player.Role.Type == core.RoleCISO {
			return player
		}
	}

	// Select player with most tokens as threat
	var bestTarget *core.Player
	maxTokens := -1
	for _, player := range humanPlayers {
		if player.Tokens > maxTokens {
			maxTokens = player.Tokens
			bestTarget = player
		}
	}

	return bestTarget
}

// selectBestControlTarget selects a human to force actions on
func (aim *AIManager) selectBestControlTarget(humanPlayers []*core.Player) *core.Player {
	if len(humanPlayers) == 0 {
		return nil
	}

	// Target players with unlocked abilities to waste their night action
	for _, player := range humanPlayers {
		if player.Role != nil && player.Role.IsUnlocked {
			return player
		}
	}

	// Fallback: random human
	return humanPlayers[rand.Intn(len(humanPlayers))]
}

// getAlignedPlayers returns all AI-aligned players (including converted humans)
func (aim *AIManager) getAlignedPlayers() []*core.Player {
	var alignedPlayers []*core.Player
	for _, player := range aim.gameState.Players {
		if player.IsAlive && (player.Alignment == "AI" || player.Alignment == "ALIGNED") {
			alignedPlayers = append(alignedPlayers, player)
		}
	}
	return alignedPlayers
}

// selectRichestHuman finds the human with the most tokens
func (aim *AIManager) selectRichestHuman(humanPlayers []*core.Player) *core.Player {
	if len(humanPlayers) == 0 {
		return nil
	}

	var richest *core.Player
	maxTokens := -1
	for _, player := range humanPlayers {
		if player.Tokens > maxTokens {
			maxTokens = player.Tokens
			richest = player
		}
	}

	return richest
}

// selectPoorestAligned finds the AI-aligned player with the fewest tokens
func (aim *AIManager) selectPoorestAligned() *core.Player {
	alignedPlayers := aim.getAlignedPlayers()
	if len(alignedPlayers) == 0 {
		return nil
	}

	var poorest *core.Player
	minTokens := 999
	for _, player := range alignedPlayers {
		if player.Tokens < minTokens {
			minTokens = player.Tokens
			poorest = player
		}
	}

	return poorest
}

// initializeAIActors creates AIActors for all AI players in the game
func (aim *AIManager) initializeAIActors() {
	aiPlayers := aim.GetAIPlayers()
	for _, player := range aiPlayers {
		// Use different personas for variety
		persona := GetShadowPersona()
		if len(aim.aiActors)%2 == 1 {
			persona = GetPuppeteerPersona()
		}
		
		actor, err := NewAIActor(player.ID, aim.gameState.ID, persona)
		if err != nil {
			log.Printf("Failed to create AIActor for player %s: %v", player.ID, err)
			continue
		}
		
		aim.aiActors[player.ID] = actor
		actor.Start()
		log.Printf("[AIManager] Started AIActor for player %s with persona %s", player.ID, persona.Name)
	}
}

// getPhaseBasedTrigger returns the appropriate trigger for the current phase
func (aim *AIManager) getPhaseBasedTrigger() string {
	switch aim.gameState.Phase.Type {
	case core.PhaseNight:
		return string(TriggerNightStart)
	case core.PhaseNomination, core.PhaseVerdict:
		return string(TriggerVotingStart)
	case core.PhaseDiscussion:
		return string(TriggerPhaseChange)
	default:
		return ""
	}
}

// TriggerChatOpportunity allows external systems to trigger AI chat
func (aim *AIManager) TriggerChatOpportunity() []core.Action {
	if !aim.useActors {
		return nil
	}
	
	var actions []core.Action
	resultChannels := make(map[string]chan *core.Action)
	
	for playerID, actor := range aim.aiActors {
		if actor.GetStatus() == "active" {
			resultChannels[playerID] = actor.TriggerAction(TriggerRandomChat, aim.gameState, nil)
		}
	}
	
	// Collect results with shorter timeout for chat
	timeout := time.After(2 * time.Second)
	for _, resultChan := range resultChannels {
		select {
		case action := <-resultChan:
			if action != nil {
				actions = append(actions, *action)
			}
		case <-timeout:
			// Chat timeouts are less critical, just continue
		}
	}
	
	return actions
}

// UpdateGameState updates the game state reference and notifies all AI actors
func (aim *AIManager) UpdateGameState(newGameState *core.GameState) {
	aim.gameState = newGameState
	
	// Update AI actors with new state
	for _, actor := range aim.aiActors {
		actor.UpdateGameState(newGameState)
	}
}

// Stop gracefully shuts down all AI actors
func (aim *AIManager) Stop() {
	log.Printf("[AIManager] Shutting down %d AI actors", len(aim.aiActors))
	
	for playerID, actor := range aim.aiActors {
		log.Printf("[AIManager] Stopping AIActor for player %s", playerID)
		actor.Stop()
	}
	
	// Clear the actors map
	aim.aiActors = make(map[string]*AIActor)
}

// GetAIActorStatus returns status information for all AI actors
func (aim *AIManager) GetAIActorStatus() map[string]string {
	status := make(map[string]string)
	for playerID, actor := range aim.aiActors {
		status[playerID] = actor.GetStatus()
	}
	return status
}

// AddAIPlayer adds a new AI player and creates an actor for them
func (aim *AIManager) AddAIPlayer(playerID string) error {
	if _, exists := aim.aiActors[playerID]; exists {
		return fmt.Errorf("AI actor already exists for player %s", playerID)
	}
	
	// Create new AI actor with default persona
	actor, err := NewAIActor(playerID, aim.gameState.ID, GetShadowPersona())
	if err != nil {
		return fmt.Errorf("failed to create AI actor: %w", err)
	}
	
	aim.aiActors[playerID] = actor
	actor.Start()
	
	log.Printf("[AIManager] Added AIActor for new player %s", playerID)
	return nil
}

// RemoveAIPlayer removes an AI player and stops their actor
func (aim *AIManager) RemoveAIPlayer(playerID string) {
	if actor, exists := aim.aiActors[playerID]; exists {
		actor.Stop()
		delete(aim.aiActors, playerID)
		log.Printf("[AIManager] Removed AIActor for player %s", playerID)
	}
}