package game

import (
	"fmt"
	"log"

	"github.com/xjhc/alignment/core"
)

// NightResolutionManager handles the resolution of all night actions
type NightResolutionManager struct {
	gameState *core.GameState
}

// NewNightResolutionManager creates a new night resolution manager
func NewNightResolutionManager(gameState *core.GameState) *NightResolutionManager {
	return &NightResolutionManager{
		gameState: gameState,
	}
}

// ResolveNightActions processes all submitted night actions in precedence order
func (nrm *NightResolutionManager) ResolveNightActions() []core.Event {
	if nrm.gameState.NightActions == nil || len(nrm.gameState.NightActions) == 0 {
		log.Printf("No night actions to resolve")
		return []core.Event{}
	}

	var allEvents []core.Event

	// Pass 1: Resolve blocking actions (highest precedence)
	// These must be resolved first as they prevent other actions
	blockEvents := nrm.resolveBlockActions()
	allEvents = append(allEvents, blockEvents...)

	// Pass 2: Resolve AI conversion attempts
	// AI targeting functions as a block, so must be resolved before standard actions
	conversionEvents := nrm.resolveConversionActions()
	allEvents = append(allEvents, conversionEvents...)

	// Pass 3: Resolve standard actions (mining, role abilities, others)
	// These are resolved for non-blocked players only
	standardEvents := nrm.resolveStandardActions()
	allEvents = append(allEvents, standardEvents...)

	// Generate summary event
	summaryEvent := nrm.createNightResolutionSummary(allEvents)
	allEvents = append(allEvents, summaryEvent)

	// Clear night actions for next night
	nrm.gameState.NightActions = make(map[string]*core.SubmittedNightAction)

	return allEvents
}

// resolveBlockActions handles all blocking actions first
func (nrm *NightResolutionManager) resolveBlockActions() []core.Event {
	var events []core.Event
	blockedPlayers := make(map[string]bool)

	for playerID, action := range nrm.gameState.NightActions {
		if action.Type == "BLOCK" {
			targetID := action.TargetID

			// Validate block action
			if nrm.canPlayerUseAbility(playerID, "BLOCK") && targetID != "" {
				blockedPlayers[targetID] = true

				event := core.Event{
					ID:        fmt.Sprintf("night_block_%s_%s", playerID, targetID),
					Type:      core.EventPlayerBlocked,
					GameID:    nrm.gameState.ID,
					PlayerID:  targetID, // The blocked player
					Timestamp: getCurrentTime(),
					Payload: map[string]interface{}{
						"blocker_id": playerID,
						"target_id":  targetID,
					},
				}
				events = append(events, event)
			}
		}
	}

	// Store blocked players for use in other resolution phases
	if len(blockedPlayers) > 0 {
		nrm.gameState.BlockedPlayersTonight = blockedPlayers
	}

	return events
}

// resolveMiningActions handles mining with liquidity pool logic
func (nrm *NightResolutionManager) resolveMiningActions() []core.Event {
	var miningRequests []MiningRequest

	// Collect all mining requests from non-blocked players
	for playerID, action := range nrm.gameState.NightActions {
		if action.Type == "MINE_TOKENS" || action.Type == "MINE" {
			// Check if player is blocked
			if nrm.isPlayerBlocked(playerID) {
				continue // Blocked players cannot mine
			}

			// For mining actions, the target is who gets the tokens
			targetID := action.TargetID
			if targetID == "" {
				// If no target specified, they're mining for themselves (not allowed by rules)
				continue
			}

			miningRequests = append(miningRequests, MiningRequest{
				MinerID:  playerID,
				TargetID: targetID,
			})
		}
	}

	// Use mining manager to resolve requests with corporate mandate and crisis effects
	miningManager := NewMiningManager(nrm.gameState)
	result := miningManager.ResolveMining(miningRequests)

	// Apply results to players and generate events
	var events []core.Event

	// Award tokens to successful mining targets
	for minerID, targetID := range result.SuccessfulMines {
		target := nrm.gameState.Players[targetID]
		miner := nrm.gameState.Players[minerID]

		if target != nil && miner != nil {
			// Award the token
			target.Tokens++

			// Create success event
			event := core.Event{
				ID:        fmt.Sprintf("mining_success_%s_%s", minerID, targetID),
				Type:      core.EventMiningSuccessful,
				GameID:    nrm.gameState.ID,
				PlayerID:  targetID, // Token goes to target
				Timestamp: getCurrentTime(),
				Payload: map[string]interface{}{
					"miner_id":    minerID,
					"miner_name":  miner.Name,
					"target_id":   targetID,
					"target_name": target.Name,
					"amount":      1,
				},
			}
			events = append(events, event)
		}
	}

	// Update failed miners' status messages for priority next round
	for playerID := range nrm.gameState.NightActions {
		if nrm.gameState.NightActions[playerID].Type == "MINE_TOKENS" || nrm.gameState.NightActions[playerID].Type == "MINE" {
			// Check if this player failed
			if _, succeeded := result.SuccessfulMines[playerID]; !succeeded {
				if player := nrm.gameState.Players[playerID]; player != nil {
					player.StatusMessage = "Mining failed - no slots available"
				}
			}
		}
	}

	return events
}

// resolveRoleAbilities handles role-specific abilities (audit, overclock, etc.)
func (nrm *NightResolutionManager) resolveRoleAbilities() []core.Event {
	var events []core.Event

	roleAbilityManager := NewRoleAbilityManager(nrm.gameState)

	for playerID, action := range nrm.gameState.NightActions {
		// Skip if player is blocked
		if nrm.isPlayerBlocked(playerID) {
			continue
		}

		// Check if this is a role ability action
		var roleAbilityAction *RoleAbilityAction
		switch action.Type {
		case "RUN_AUDIT":
			targetID, _ := action.Payload["target_id"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "RUN_AUDIT",
				TargetID:    targetID,
			}
		case "OVERCLOCK_SERVERS":
			targetID, _ := action.Payload["target_id"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "OVERCLOCK_SERVERS",
				TargetID:    targetID,
			}
		case "ISOLATE_NODE":
			targetID, _ := action.Payload["target_id"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "ISOLATE_NODE",
				TargetID:    targetID,
			}
		case "PERFORMANCE_REVIEW":
			targetID, _ := action.Payload["target_id"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "PERFORMANCE_REVIEW",
				TargetID:    targetID,
			}
		case "REALLOCATE_BUDGET":
			sourceID, _ := action.Payload["source_id"].(string)
			targetID, _ := action.Payload["target_id"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:       playerID,
				AbilityType:    "REALLOCATE_BUDGET",
				TargetID:       sourceID,
				SecondTargetID: targetID,
			}
		case "PIVOT":
			chosenCrisis, _ := action.Payload["chosen_crisis"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "PIVOT",
				Parameters:  map[string]interface{}{"chosen_crisis": chosenCrisis},
			}
		case "DEPLOY_HOTFIX":
			section, _ := action.Payload["redacted_section"].(string)
			roleAbilityAction = &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "DEPLOY_HOTFIX",
				Parameters:  map[string]interface{}{"redacted_section": section},
			}
		}

		if roleAbilityAction != nil {
			result, err := roleAbilityManager.UseRoleAbility(*roleAbilityAction)
			if err == nil && result != nil {
				events = append(events, result.PublicEvents...)
				// Private events would be sent only to AI faction
				events = append(events, result.PrivateEvents...)
			}
		}
	}

	return events
}

// resolveConversionActions handles AI conversion attempts in Pass 2
func (nrm *NightResolutionManager) resolveConversionActions() []core.Event {
	var events []core.Event

	for playerID, action := range nrm.gameState.NightActions {
		// Skip if player is blocked from Pass 1
		if nrm.isPlayerBlocked(playerID) {
			continue
		}

		if action.Type == "ATTEMPT_CONVERSION" || action.Type == "CONVERT" {
			if nrm.canPlayerUseAbility(playerID, "CONVERT") {
				// AI conversion also blocks the target player
				targetID := action.TargetID
				if targetID != "" {
					if nrm.gameState.BlockedPlayersTonight == nil {
						nrm.gameState.BlockedPlayersTonight = make(map[string]bool)
					}
					nrm.gameState.BlockedPlayersTonight[targetID] = true
				}

				convertEvents := nrm.resolveConvertAction(playerID, action)
				events = append(events, convertEvents...)
			}
		}
	}

	return events
}

// resolveStandardActions handles all remaining actions in Pass 3
func (nrm *NightResolutionManager) resolveStandardActions() []core.Event {
	var events []core.Event

	// Phase 3a: Resolve mining actions with liquidity pool
	miningEvents := nrm.resolveMiningActions()
	events = append(events, miningEvents...)

	// Phase 3b: Resolve role-specific abilities (audit, overclock, etc.)
	roleAbilityEvents := nrm.resolveRoleAbilities()
	events = append(events, roleAbilityEvents...)

	// Phase 3c: Resolve other night actions (investigate, protect)
	otherEvents := nrm.resolveOtherNightActions()
	events = append(events, otherEvents...)

	return events
}

// resolveOtherNightActions handles investigate and protect actions (after conversion)
func (nrm *NightResolutionManager) resolveOtherNightActions() []core.Event {
	var events []core.Event

	for playerID, action := range nrm.gameState.NightActions {
		// Skip if player is blocked (from Pass 1 or Pass 2)
		if nrm.isPlayerBlocked(playerID) {
			continue
		}

		switch action.Type {
		case "INVESTIGATE":
			if nrm.canPlayerUseAbility(playerID, "INVESTIGATE") {
				events = append(events, nrm.resolveInvestigateAction(playerID, action))
			}
		case "PROTECT":
			if nrm.canPlayerUseAbility(playerID, "PROTECT") {
				events = append(events, nrm.resolveProtectAction(playerID, action))
			}
		}
	}

	return events
}

// resolveInvestigateAction handles investigation abilities
func (nrm *NightResolutionManager) resolveInvestigateAction(playerID string, action *core.SubmittedNightAction) core.Event {
	targetID := action.TargetID
	target := nrm.gameState.Players[targetID]

	var roleType string
	if target.Role != nil {
		roleType = string(target.Role.Type)
	} else {
		roleType = "UNKNOWN"
	}

	// Reveal target's alignment to investigator
	event := core.Event{
		ID:        fmt.Sprintf("night_investigate_%s_%s", playerID, targetID),
		Type:      core.EventPlayerInvestigated,
		GameID:    nrm.gameState.ID,
		PlayerID:  playerID, // Information goes to investigator
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"investigator_id": playerID,
			"target_id":       targetID,
			"target_name":     target.Name,
			"alignment":       target.Alignment,
			"role":            roleType,
		},
	}

	return event
}

// resolveProtectAction handles protection abilities
func (nrm *NightResolutionManager) resolveProtectAction(playerID string, action *core.SubmittedNightAction) core.Event {
	targetID := action.TargetID

	// Mark player as protected for tonight
	if nrm.gameState.ProtectedPlayersTonight == nil {
		nrm.gameState.ProtectedPlayersTonight = make(map[string]bool)
	}
	nrm.gameState.ProtectedPlayersTonight[targetID] = true

	event := core.Event{
		ID:        fmt.Sprintf("night_protect_%s_%s", playerID, targetID),
		Type:      core.EventPlayerProtected,
		GameID:    nrm.gameState.ID,
		PlayerID:  targetID, // Protected player
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"protector_id": playerID,
			"target_id":    targetID,
		},
	}

	return event
}

// resolveConvertAction handles AI conversion attempts
func (nrm *NightResolutionManager) resolveConvertAction(playerID string, action *core.SubmittedNightAction) []core.Event {
	targetID := action.TargetID
	target := nrm.gameState.Players[targetID]
	player := nrm.gameState.Players[playerID]

	// Check if AI conversions are blocked by crisis event
	if nrm.gameState.CrisisEvent != nil {
		if blocked, exists := nrm.gameState.CrisisEvent.Effects["block_ai_conversions"]; exists {
			if isBlocked, ok := blocked.(bool); ok && isBlocked {
				return []core.Event{{
					ID:        fmt.Sprintf("night_convert_crisis_blocked_%s_%s", playerID, targetID),
					Type:      core.EventSystemMessage,
					GameID:    nrm.gameState.ID,
					PlayerID:  playerID,
					Timestamp: getCurrentTime(),
					Payload: map[string]interface{}{
						"message": "AI conversion blocked by active crisis protocols",
						"crisis":  nrm.gameState.CrisisEvent.Title,
					},
				}}
			}
		}
	}

	// Check corporate mandate restrictions
	if nrm.gameState.CorporateMandate != nil && nrm.gameState.CorporateMandate.IsActive {
		if blockVal, exists := nrm.gameState.CorporateMandate.Effects["block_ai_odd_nights"]; exists {
			if blockOdd, ok := blockVal.(bool); ok && blockOdd {
				// Check if this is an odd night
				nightNumber := nrm.gameState.DayNumber
				if nightNumber%2 == 1 {
					return []core.Event{{
						ID:        fmt.Sprintf("night_convert_mandate_blocked_%s_%s", playerID, targetID),
						Type:      core.EventSystemMessage,
						GameID:    nrm.gameState.ID,
						PlayerID:  playerID,
						Timestamp: getCurrentTime(),
						Payload: map[string]interface{}{
							"message": "AI conversion blocked by Security Lockdown Protocol on odd nights",
							"mandate": nrm.gameState.CorporateMandate.Name,
						},
					}}
				}
			}
		}
	}

	// Check if target is protected
	if nrm.isPlayerProtected(targetID) {
		// Conversion blocked by protection
		return []core.Event{{
			ID:        fmt.Sprintf("night_convert_blocked_%s_%s", playerID, targetID),
			Type:      core.EventSystemMessage,
			GameID:    nrm.gameState.ID,
			PlayerID:  playerID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"message": "Conversion attempt blocked by protection",
			},
		}}
	}

	// Calculate conversion success based on AI Equity vs Player Tokens
	conversionThreshold := player.AIEquity

	// Check for crisis AI equity bonus
	if nrm.gameState.CrisisEvent != nil {
		if bonus, exists := nrm.gameState.CrisisEvent.Effects["ai_equity_bonus"]; exists {
			if bonusVal, ok := bonus.(int); ok {
				conversionThreshold += bonusVal
			}
		}
	}

	if conversionThreshold > target.Tokens {
		// Successful conversion - target becomes AI aligned
		target.Alignment = "ALIGNED"

		// Apply AI equity bonus from crisis if applicable
		equityGained := 1
		if nrm.gameState.CrisisEvent != nil {
			if bonus, exists := nrm.gameState.CrisisEvent.Effects["ai_equity_bonus"]; exists {
				if bonusVal, ok := bonus.(int); ok {
					equityGained += bonusVal
				}
			}
		}
		target.AIEquity += equityGained

		return []core.Event{{
			ID:        fmt.Sprintf("night_convert_success_%s_%s", playerID, targetID),
			Type:      core.EventAIConversionSuccess,
			GameID:    nrm.gameState.ID,
			PlayerID:  targetID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"converter_id":     playerID,
				"target_id":        targetID,
				"target_name":      target.Name,
				"ai_equity_gained": equityGained,
				"new_ai_equity":    target.AIEquity,
			},
		}}
	} else {
		// System shock - proves target is human and applies shock effects
		shock := &core.SystemShock{
			Type:        core.ShockActionLock,
			Description: "System integrity compromised - conversion attempt detected",
			ExpiresAt:   getCurrentTime().Add(24 * 3600 * 1000000000), // 24 hours
			IsActive:    true,
		}
		target.SystemShocks = append(target.SystemShocks, *shock)

		return []core.Event{{
			ID:        fmt.Sprintf("night_convert_shock_%s_%s", playerID, targetID),
			Type:      core.EventPlayerShocked,
			GameID:    nrm.gameState.ID,
			PlayerID:  targetID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"converter_id":   playerID,
				"target_id":      targetID,
				"target_name":    target.Name,
				"reason":         "System shock from failed conversion",
				"shock_type":     string(shock.Type),
				"shock_duration": "24 hours",
			},
		}}
	}
}

// createNightResolutionSummary creates a comprehensive summary event of all night actions
func (nrm *NightResolutionManager) createNightResolutionSummary(resolvedEvents []core.Event) core.Event {
	// Analyze the resolved events to create a detailed summary
	eventCounts := make(map[string]int)
	blockedPlayers := []string{}
	convertedPlayers := []string{}
	eliminatedPlayers := []string{}
	miningResults := make(map[string]interface{})
	
	for _, event := range resolvedEvents {
		// Count event types
		eventType := string(event.Type)
		eventCounts[eventType]++
		
		// Extract specific information based on event type
		switch event.Type {
		case core.EventPlayerBlocked:
			if targetID, ok := event.Payload["target_id"].(string); ok {
				blockedPlayers = append(blockedPlayers, targetID)
			}
		case core.EventAIConversionSuccess:
			if targetID, ok := event.Payload["target_id"].(string); ok {
				convertedPlayers = append(convertedPlayers, targetID)
			}
		case core.EventPlayerEliminated:
			if playerID := event.PlayerID; playerID != "" {
				eliminatedPlayers = append(eliminatedPlayers, playerID)
			}
		case core.EventMiningSuccessful:
			if minerID, ok := event.Payload["miner_id"].(string); ok {
				if targetID, ok := event.Payload["target_id"].(string); ok {
					miningResults[minerID] = targetID
				}
			}
		}
	}

	// Create comprehensive summary payload
	summary := map[string]interface{}{
		"night_number":      nrm.gameState.DayNumber,
		"total_actions":     len(nrm.gameState.NightActions),
		"resolved_events":   len(resolvedEvents),
		"event_counts":      eventCounts,
		"blocked_players":   blockedPlayers,
		"converted_players": convertedPlayers,
		"eliminated_players": eliminatedPlayers,
		"mining_results":    miningResults,
		"phase_end":         true,
		"next_phase":        "SITREP",
	}

	return core.Event{
		ID:        fmt.Sprintf("night_resolution_summary_%d", nrm.gameState.DayNumber),
		Type:      core.EventNightActionsResolved,
		GameID:    nrm.gameState.ID,
		PlayerID:  "", // Public event - broadcast to all players
		Timestamp: getCurrentTime(),
		Payload:   summary,
	}
}

// Helper methods

func (nrm *NightResolutionManager) canPlayerUseAbility(playerID, abilityType string) bool {
	player := nrm.gameState.Players[playerID]
	if player == nil || !player.IsAlive {
		return false
	}

	// Check milestone requirements (may be modified by corporate mandate)
	requiredMilestones := 3 // Default requirement

	// Check if corporate mandate modifies milestone requirements
	if nrm.gameState.CorporateMandate != nil && nrm.gameState.CorporateMandate.IsActive {
		if milestonesVal, exists := nrm.gameState.CorporateMandate.Effects["milestones_for_abilities"]; exists {
			if milestones, ok := milestonesVal.(int); ok {
				requiredMilestones = milestones
			}
		}
	}

	return player.ProjectMilestones >= requiredMilestones
}

func (nrm *NightResolutionManager) isPlayerBlocked(playerID string) bool {
	if nrm.gameState.BlockedPlayersTonight == nil {
		return false
	}
	return nrm.gameState.BlockedPlayersTonight[playerID]
}

func (nrm *NightResolutionManager) isPlayerProtected(playerID string) bool {
	if nrm.gameState.ProtectedPlayersTonight == nil {
		return false
	}
	return nrm.gameState.ProtectedPlayersTonight[playerID]
}
