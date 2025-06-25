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
		} else if action.Type == "ISOLATE_NODE" {
			// ISOLATE_NODE is a blocking ability that must be processed first
			targetID := action.TargetID
			
			// Use the role ability manager to process the ISOLATE_NODE
			roleAbilityManager := NewRoleAbilityManager(nrm.gameState)
			roleAbilityAction := &RoleAbilityAction{
				PlayerID:    playerID,
				AbilityType: "ISOLATE_NODE",
				TargetID:    targetID,
			}
			
			result, err := roleAbilityManager.UseRoleAbility(*roleAbilityAction)
			if err == nil && result != nil {
				// The role ability manager should have set BlockedPlayersTonight
				if nrm.gameState.BlockedPlayersTonight != nil && nrm.gameState.BlockedPlayersTonight[targetID] {
					blockedPlayers[targetID] = true
					
					event := core.Event{
						ID:        fmt.Sprintf("isolate_node_%s_%s", playerID, targetID),
						Type:      "ISOLATE_NODE",
						GameID:    nrm.gameState.ID,
						PlayerID:  playerID, // The CISO who performed the isolation
						Timestamp: getCurrentTime(),
						Payload: map[string]interface{}{
							"ciso_id":   playerID,
							"target_id": targetID,
						},
					}
					events = append(events, event)
				}
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
		case "PROJECT_MILESTONES":
			events = append(events, nrm.resolveProjectMilestoneAction(playerID, action))
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

	// AI targeting increases target's AI Equity by 1 (this is the core mechanic)
	target.AIEquity++

	// Calculate conversion success based on target's AI Equity vs their Tokens
	conversionThreshold := target.AIEquity

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
		
		// Apply any crisis equity bonus to the target's actual AIEquity
		if nrm.gameState.CrisisEvent != nil {
			if bonus, exists := nrm.gameState.CrisisEvent.Effects["ai_equity_bonus"]; exists {
				if bonusVal, ok := bonus.(int); ok {
					target.AIEquity += bonusVal
				}
			}
		}

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
				"ai_equity":        target.AIEquity,
				"target_tokens":    target.Tokens,
			},
		}}
	} else {
		// Failed conversion - apply System Shock with MessageCorruption
		shock := &core.SystemShock{
			Type:        core.ShockMessageCorruption,
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
				"ai_equity":      target.AIEquity,
				"target_tokens":  target.Tokens,
			},
		}}
	}
}

// createNightResolutionSummary creates a comprehensive summary event of all night actions
func (nrm *NightResolutionManager) createNightResolutionSummary(resolvedEvents []core.Event) core.Event {
	// Analyze the resolved events to create a detailed summary
	eventCounts := make(map[string]int)
	blockedPlayers := []map[string]interface{}{}
	convertedPlayers := []map[string]interface{}{}
	shockedPlayers := []map[string]interface{}{}
	eliminatedPlayers := []map[string]interface{}{}
	miningResults := []map[string]interface{}{}
	roleAbilityResults := []map[string]interface{}{}
	milestoneResults := []map[string]interface{}{}
	
	for _, event := range resolvedEvents {
		// Count event types
		eventType := string(event.Type)
		eventCounts[eventType]++
		
		// Extract specific information based on event type with player names for UI
		switch event.Type {
		case core.EventPlayerBlocked:
			if targetID, ok := event.Payload["target_id"].(string); ok {
				if target := nrm.gameState.Players[targetID]; target != nil {
					blockedPlayers = append(blockedPlayers, map[string]interface{}{
						"player_id":   targetID,
						"player_name": target.Name,
						"blocked_by":  event.Payload["blocker_id"],
					})
				}
			}
		case core.EventAIConversionSuccess:
			if targetID, ok := event.Payload["target_id"].(string); ok {
				if target := nrm.gameState.Players[targetID]; target != nil {
					convertedPlayers = append(convertedPlayers, map[string]interface{}{
						"player_id":        targetID,
						"player_name":      target.Name,
						"ai_equity_gained": event.Payload["ai_equity_gained"],
						"new_ai_equity":    event.Payload["new_ai_equity"],
					})
				}
			}
		case core.EventPlayerShocked:
			if targetID, ok := event.Payload["target_id"].(string); ok {
				if target := nrm.gameState.Players[targetID]; target != nil {
					shockedPlayers = append(shockedPlayers, map[string]interface{}{
						"player_id":      targetID,
						"player_name":    target.Name,
						"shock_type":     event.Payload["shock_type"],
						"shock_duration": event.Payload["shock_duration"],
						"reason":         event.Payload["reason"],
					})
				}
			}
		case core.EventPlayerEliminated:
			if playerID := event.PlayerID; playerID != "" {
				if player := nrm.gameState.Players[playerID]; player != nil {
					eliminatedPlayers = append(eliminatedPlayers, map[string]interface{}{
						"player_id":   playerID,
						"player_name": player.Name,
						"role_type":   event.Payload["role_type"],
						"alignment":   event.Payload["alignment"],
					})
				}
			}
		case core.EventMiningSuccessful:
			if minerID, ok := event.Payload["miner_id"].(string); ok {
				if targetID, ok := event.Payload["target_id"].(string); ok {
					miner := nrm.gameState.Players[minerID]
					target := nrm.gameState.Players[targetID]
					if miner != nil && target != nil {
						miningResults = append(miningResults, map[string]interface{}{
							"miner_id":     minerID,
							"miner_name":   miner.Name,
							"target_id":    targetID,
							"target_name":  target.Name,
							"tokens_mined": event.Payload["amount"],
						})
					}
				}
			}
		case core.EventRunAudit, core.EventOverclockServers, core.EventIsolateNode, 
			 core.EventPerformanceReview, core.EventReallocateBudget, core.EventPivot, core.EventDeployHotfix:
			if player := nrm.gameState.Players[event.PlayerID]; player != nil {
				roleAbilityResults = append(roleAbilityResults, map[string]interface{}{
					"player_id":    event.PlayerID,
					"player_name":  player.Name,
					"ability_type": string(event.Type),
					"target_id":    event.Payload["target_id"],
					"message":      event.Payload["message"],
				})
			}
		case core.EventProjectMilestone:
			if player := nrm.gameState.Players[event.PlayerID]; player != nil {
				milestoneResults = append(milestoneResults, map[string]interface{}{
					"player_id":        event.PlayerID,
					"player_name":      player.Name,
					"milestones_count": event.Payload["milestones_count"],
					"role_unlocked":    event.Payload["role_unlocked"],
					"message":          event.Payload["message"],
				})
			}
		}
	}

	// Create comprehensive summary payload with human-readable information
	summary := map[string]interface{}{
		"night_number":         nrm.gameState.DayNumber,
		"total_actions":        len(nrm.gameState.NightActions),
		"resolved_events":      len(resolvedEvents),
		"event_counts":         eventCounts,
		"blocked_players":      blockedPlayers,
		"converted_players":    convertedPlayers,
		"shocked_players":      shockedPlayers,
		"eliminated_players":   eliminatedPlayers,
		"mining_results":       miningResults,
		"role_ability_results": roleAbilityResults,
		"milestone_results":    milestoneResults,
		"phase_end":            true,
		"next_phase":           "SITREP",
		"summary_message":      nrm.createHumanReadableSummary(blockedPlayers, convertedPlayers, shockedPlayers, miningResults, roleAbilityResults, milestoneResults),
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

// createHumanReadableSummary generates a text summary for display in SITREP
func (nrm *NightResolutionManager) createHumanReadableSummary(blocked, converted, shocked, mining, abilities, milestones []map[string]interface{}) string {
	summary := fmt.Sprintf("Night %d Summary:\n", nrm.gameState.DayNumber)
	
	if len(blocked) > 0 {
		summary += "• Players blocked from actions: "
		for i, p := range blocked {
			if i > 0 {
				summary += ", "
			}
			summary += p["player_name"].(string)
		}
		summary += "\n"
	}
	
	if len(converted) > 0 {
		summary += "• Players converted by AI: "
		for i, p := range converted {
			if i > 0 {
				summary += ", "
			}
			summary += p["player_name"].(string)
		}
		summary += "\n"
	}
	
	if len(shocked) > 0 {
		summary += "• Players experienced system shock: "
		for i, p := range shocked {
			if i > 0 {
				summary += ", "
			}
			summary += p["player_name"].(string)
		}
		summary += "\n"
	}
	
	if len(mining) > 0 {
		summary += fmt.Sprintf("• %d successful mining operations completed\n", len(mining))
	}
	
	if len(abilities) > 0 {
		summary += fmt.Sprintf("• %d role abilities were used\n", len(abilities))
	}
	
	if len(milestones) > 0 {
		summary += "• Project milestone advancement: "
		rolesUnlocked := 0
		for i, m := range milestones {
			if i > 0 {
				summary += ", "
			}
			summary += fmt.Sprintf("%s (%v total)", m["player_name"], m["milestones_count"])
			if unlocked, ok := m["role_unlocked"].(bool); ok && unlocked {
				rolesUnlocked++
			}
		}
		summary += "\n"
		if rolesUnlocked > 0 {
			summary += fmt.Sprintf("• %d role abilities unlocked!\n", rolesUnlocked)
		}
	}
	
	return summary
}

// resolveProjectMilestoneAction handles project milestone advancement
func (nrm *NightResolutionManager) resolveProjectMilestoneAction(playerID string, action *core.SubmittedNightAction) core.Event {
	player := nrm.gameState.Players[playerID]
	if player == nil {
		return core.Event{}
	}

	// Increment project milestones
	player.ProjectMilestones++

	// Check if this unlocks their role ability
	var roleUnlocked bool
	if player.Role != nil && player.ProjectMilestones >= 3 && !player.Role.IsUnlocked {
		player.Role.IsUnlocked = true
		roleUnlocked = true
	}

	// Create project milestone event
	event := core.Event{
		ID:        fmt.Sprintf("project_milestone_%s_%d", playerID, getCurrentTime().UnixNano()),
		Type:      core.EventProjectMilestone,
		GameID:    nrm.gameState.ID,
		PlayerID:  playerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"player_id":           playerID,
			"player_name":         player.Name,
			"milestones_count":    player.ProjectMilestones,
			"role_unlocked":       roleUnlocked,
			"message":             fmt.Sprintf("%s completed a project milestone (Total: %d)", player.Name, player.ProjectMilestones),
		},
	}

	// If role was unlocked, add role unlock event
	if roleUnlocked {
		event.Payload["role_unlocked_message"] = fmt.Sprintf("%s has unlocked their %s role ability!", player.Name, nrm.getRoleDisplayName(player.Role.Type))
	}

	return event
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

func (nrm *NightResolutionManager) getRoleDisplayName(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "CISO"
	case core.RoleCTO:
		return "CTO"
	case core.RoleCOO:
		return "COO"
	case core.RoleCFO:
		return "CFO"
	case core.RoleCEO:
		return "CEO"
	case core.RoleEthics:
		return "VP Ethics"
	case core.RolePlatforms:
		return "VP Platforms"
	case core.RoleIntern:
		return "Intern"
	default:
		return "Unknown Role"
	}
}
