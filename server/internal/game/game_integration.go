package game

import (
	"fmt"
	"log"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/ai"
)

// GameManager integrates all the game systems together
type GameManager struct {
	GameState          *core.GameState
	AIEngine           *ai.RulesEngine
	CrisisManager      *CrisisEventManager
	MandateManager     *CorporateMandateManager
	SitrepGenerator    *SitrepGenerator
	NightResolutionMgr *NightResolutionManager
	MiningManager      *MiningManager
	RoleAbilityManager *RoleAbilityManager
}

// NewGameManager creates a fully integrated game manager
func NewGameManager(gameID string) *GameManager {
	gameState := core.NewGameState(gameID)

	return &GameManager{
		GameState:          gameState,
		AIEngine:           ai.NewRulesEngine(),
		CrisisManager:      NewCrisisEventManager(gameState),
		MandateManager:     NewCorporateMandateManager(gameState),
		SitrepGenerator:    NewSitrepGenerator(gameState),
		NightResolutionMgr: NewNightResolutionManager(gameState),
		MiningManager:      NewMiningManager(gameState),
		RoleAbilityManager: NewRoleAbilityManager(gameState),
	}
}

// StartGame initializes a new game with random mandate and AI player
func (gm *GameManager) StartGame() error {
	log.Printf("Starting game %s", gm.GameState.ID)

	// Assign a random corporate mandate to modify the game rules
	mandate := gm.MandateManager.AssignRandomMandate()
	if mandate != nil {
		log.Printf("Corporate mandate assigned: %s", mandate.Name)
	}

	// Apply game start event
	startEvent := core.Event{
		ID:        fmt.Sprintf("game_start_%s", gm.GameState.ID),
		Type:      core.EventGameStarted,
		GameID:    gm.GameState.ID,
		Timestamp: getCurrentTime(),
	}

	newState := core.ApplyEvent(*gm.GameState, startEvent)
	*gm.GameState = newState
	return nil
}

// ProcessDayPhase handles the complete day phase cycle
func (gm *GameManager) ProcessDayPhase() error {
	log.Printf("Processing day %d", gm.GameState.DayNumber)

	// Generate daily SITREP
	sitrep := gm.SitrepGenerator.GenerateDailySitrep()
	log.Printf("Generated SITREP with %d sections, alert level: %s",
		len(sitrep.Sections), sitrep.AlertLevel)

	// Check for random crisis events (30% chance per day)
	if gm.shouldTriggerCrisis() {
		crisis := gm.CrisisManager.TriggerRandomCrisis()
		if crisis != nil {
			log.Printf("Crisis triggered: %s", crisis.Title)

			// Generate crisis event
			crisisEvent := core.Event{
				ID:        fmt.Sprintf("crisis_%s_day_%d", gm.GameState.ID, gm.GameState.DayNumber),
				Type:      core.EventCrisisTriggered,
				GameID:    gm.GameState.ID,
				Timestamp: getCurrentTime(),
				Payload: map[string]interface{}{
					"crisis_type": crisis.Type,
					"title":       crisis.Title,
					"description": crisis.Description,
					"effects":     crisis.Effects,
				},
			}

			newState := core.ApplyEvent(*gm.GameState, crisisEvent)
			*gm.GameState = newState
		}
	}

	// AI makes day phase decisions
	gameData := map[string]interface{}{
		"phase":   string(gm.GameState.Phase.Type),
		"players": gm.serializePlayersForAI(),
	}
	aiDecision := gm.AIEngine.MakeDecisionFromData(gameData)
	log.Printf("AI day decision: %s - %s", aiDecision.Action, aiDecision.Reason)

	return nil
}

// ProcessNightPhase handles the complete night phase cycle
func (gm *GameManager) ProcessNightPhase() error {
	log.Printf("Processing night phase for day %d", gm.GameState.DayNumber)

	// AI makes night decision
	gameData := map[string]interface{}{
		"phase":   string(gm.GameState.Phase.Type),
		"players": gm.serializePlayersForAI(),
	}
	aiDecision := gm.AIEngine.MakeDecisionFromData(gameData)
	log.Printf("AI night decision: %s - %s", aiDecision.Action, aiDecision.Reason)

	// Convert AI decision to submitted night action if applicable
	if gm.isNightActionString(aiDecision.Action) {
		aiPlayerID := gm.findAIPlayer()
		if aiPlayerID != "" {
			nightAction := &core.SubmittedNightAction{
				PlayerID: aiPlayerID,
				Type:     aiDecision.Action,
				TargetID: aiDecision.Target,
				Payload:  aiDecision.Payload,
			}

			// Store in night actions
			if gm.GameState.NightActions == nil {
				gm.GameState.NightActions = make(map[string]*core.SubmittedNightAction)
			}
			gm.GameState.NightActions[aiPlayerID] = nightAction

			log.Printf("AI submitted night action: %s targeting %s", nightAction.Type, nightAction.TargetID)
		}
	}

	// Resolve all night actions
	events := gm.NightResolutionMgr.ResolveNightActions()
	log.Printf("Night resolution generated %d events", len(events))

	// Apply all resolution events
	for _, event := range events {
		newState := core.ApplyEvent(*gm.GameState, event)
		*gm.GameState = newState
	}

	return nil
}

// GetAIThreatAssessment returns the AI's current threat analysis
func (gm *GameManager) GetAIThreatAssessment() []ai.PlayerThreat {
	gameData := map[string]interface{}{
		"phase":   string(gm.GameState.Phase.Type),
		"players": gm.serializePlayersForAI(),
	}
	return gm.AIEngine.GetThreatAssessmentFromData(gameData)
}

// GetGameStatus returns a summary of the current game state
func (gm *GameManager) GetGameStatus() map[string]interface{} {
	status := map[string]interface{}{
		"game_id":      gm.GameState.ID,
		"day_number":   gm.GameState.DayNumber,
		"phase":        gm.GameState.Phase.Type,
		"player_count": len(gm.GameState.Players),
	}

	// Count alive players
	aliveCount := 0
	for _, player := range gm.GameState.Players {
		if player.IsAlive {
			aliveCount++
		}
	}
	status["alive_players"] = aliveCount

	// Crisis status
	if gm.CrisisManager.IsCrisisActive() {
		crisis := gm.CrisisManager.GetActiveCrisis()
		status["active_crisis"] = crisis.Title
	}

	// Mandate status
	if gm.MandateManager.IsMandateActive() {
		mandate := gm.MandateManager.GetActiveMandate()
		status["active_mandate"] = mandate.Name
	}

	// AI threat assessment
	threats := gm.GetAIThreatAssessment()
	if len(threats) > 0 {
		status["highest_threat_level"] = threats[0].ThreatLevel
		status["threat_count"] = len(threats)
	}

	return status
}

// ProcessRoleAbility handles a player using their role ability
func (gm *GameManager) ProcessRoleAbility(playerID string, abilityType string, targetID string, parameters map[string]interface{}) error {
	// Check mandate restrictions
	if player, exists := gm.GameState.Players[playerID]; exists {
		if canUse, reason := gm.MandateManager.CheckRoleAbilityRequirements(player); !canUse {
			return fmt.Errorf("mandate restriction: %s", reason)
		}
	}

	// Use role ability
	action := RoleAbilityAction{
		PlayerID:    playerID,
		AbilityType: abilityType,
		TargetID:    targetID,
		Parameters:  parameters,
	}

	result, err := gm.RoleAbilityManager.UseRoleAbility(action)
	if err != nil {
		return fmt.Errorf("failed to use role ability: %w", err)
	}

	// Apply resulting events
	for _, event := range result.PublicEvents {
		newState := core.ApplyEvent(*gm.GameState, event)
		*gm.GameState = newState
	}

	// Private events would be sent only to AI faction players
	for _, event := range result.PrivateEvents {
		newState := core.ApplyEvent(*gm.GameState, event)
		*gm.GameState = newState
	}

	log.Printf("Player %s used ability %s, generated %d public events and %d private events",
		playerID, abilityType, len(result.PublicEvents), len(result.PrivateEvents))

	return nil
}

// Helper methods

// shouldTriggerCrisis determines if a crisis event should occur (30% chance)
func (gm *GameManager) shouldTriggerCrisis() bool {
	// Only trigger crisis if none is currently active and we're past day 1
	if gm.CrisisManager.IsCrisisActive() || gm.GameState.DayNumber <= 1 {
		return false
	}

	// 30% chance per day - simple random check
	return gm.GameState.DayNumber%3 == 0 // Trigger every 3 days for demo
}

// isNightActionString checks if an action string is a night action
func (gm *GameManager) isNightActionString(actionStr string) bool {
	nightActions := []string{
		"MINE_TOKENS", "ATTEMPT_CONVERSION", "USE_ABILITY",
		"RUN_AUDIT", "OVERCLOCK_SERVERS", "ISOLATE_NODE",
		"PERFORMANCE_REVIEW", "REALLOCATE_BUDGET", "PIVOT", "DEPLOY_HOTFIX",
	}

	for _, nightAction := range nightActions {
		if actionStr == nightAction {
			return true
		}
	}
	return false
}

// findAIPlayer locates the AI-controlled player
func (gm *GameManager) findAIPlayer() string {
	for playerID, player := range gm.GameState.Players {
		if player.IsAlive && player.Alignment == "ALIGNED" {
			return playerID
		}
	}
	return ""
}

// DemoGameFlow demonstrates a complete game flow
func (gm *GameManager) DemoGameFlow() {
	log.Printf("=== Demo Game Flow ===")

	// Add some demo players
	gm.addDemoPlayers()

	// Start the game
	if err := gm.StartGame(); err != nil {
		log.Printf("Error starting game: %v", err)
		return
	}

	// Process a few day/night cycles
	for day := 1; day <= 3; day++ {
		log.Printf("\n--- Day %d ---", day)

		if err := gm.ProcessDayPhase(); err != nil {
			log.Printf("Error in day phase: %v", err)
		}

		if err := gm.ProcessNightPhase(); err != nil {
			log.Printf("Error in night phase: %v", err)
		}

		// Print game status
		status := gm.GetGameStatus()
		log.Printf("Game status: %+v", status)
	}

	log.Printf("=== Demo Complete ===")
}

// addDemoPlayers adds some demo players for testing
func (gm *GameManager) addDemoPlayers() {
	players := []struct {
		ID       string
		Name     string
		JobTitle string
	}{
		{"player1", "Alice Chen", "CISO"},
		{"player2", "Bob Smith", "CTO"},
		{"player3", "Carol Johnson", "CFO"},
		{"player4", "David Lee", "CEO"},
	}

	for i, p := range players {
		joinEvent := core.Event{
			ID:        fmt.Sprintf("join_%s", p.ID),
			Type:      core.EventPlayerJoined,
			GameID:    gm.GameState.ID,
			PlayerID:  p.ID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"name":      p.Name,
				"job_title": p.JobTitle,
			},
		}

		newState := core.ApplyEvent(*gm.GameState, joinEvent)
		*gm.GameState = newState

		// Make one player AI-aligned for demo
		if i == 1 { // Bob Smith becomes AI
			gm.GameState.Players[p.ID].Alignment = "ALIGNED"
		}
	}
}

// serializePlayersForAI converts players to a simple map for AI consumption
func (gm *GameManager) serializePlayersForAI() map[string]interface{} {
	aiPlayers := make(map[string]interface{})

	for playerID, player := range gm.GameState.Players {
		aiPlayers[playerID] = map[string]interface{}{
			"id":                 player.ID,
			"name":               player.Name,
			"is_alive":           player.IsAlive,
			"tokens":             player.Tokens,
			"project_milestones": player.ProjectMilestones,
			"alignment":          player.Alignment,
			"ai_equity":          player.AIEquity,
		}
	}

	return aiPlayers
}
