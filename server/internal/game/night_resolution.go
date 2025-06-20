package game

import (
	"fmt"
	"log"
)

// NightResolutionManager handles the resolution of all night actions
type NightResolutionManager struct {
	gameState *GameState
}

// NewNightResolutionManager creates a new night resolution manager
func NewNightResolutionManager(gameState *GameState) *NightResolutionManager {
	return &NightResolutionManager{
		gameState: gameState,
	}
}

// ResolveNightActions processes all submitted night actions in precedence order
func (nrm *NightResolutionManager) ResolveNightActions() []Event {
	if nrm.gameState.NightActions == nil || len(nrm.gameState.NightActions) == 0 {
		log.Printf("No night actions to resolve")
		return []Event{}
	}

	var allEvents []Event

	// Phase 1: Resolve blocking actions (highest precedence)
	blockEvents := nrm.resolveBlockActions()
	allEvents = append(allEvents, blockEvents...)

	// Phase 2: Resolve mining actions with liquidity pool
	miningEvents := nrm.resolveMiningActions()
	allEvents = append(allEvents, miningEvents...)

	// Phase 3: Resolve role-specific abilities (audit, overclock, etc.)
	roleAbilityEvents := nrm.resolveRoleAbilities()
	allEvents = append(allEvents, roleAbilityEvents...)

	// Phase 4: Resolve other night actions (investigate, protect, convert)
	nightActionEvents := nrm.resolveOtherNightActions()
	allEvents = append(allEvents, nightActionEvents...)

	// Generate summary event
	summaryEvent := nrm.createNightResolutionSummary(allEvents)
	allEvents = append(allEvents, summaryEvent)

	// Clear night actions for next night
	nrm.gameState.NightActions = make(map[string]*SubmittedNightAction)

	return allEvents
}

// resolveBlockActions handles all blocking actions first
func (nrm *NightResolutionManager) resolveBlockActions() []Event {
	var events []Event
	blockedPlayers := make(map[string]bool)

	for playerID, action := range nrm.gameState.NightActions {
		if action.Type == "BLOCK" {
			targetID := action.TargetID

			// Validate block action
			if nrm.canPlayerUseAbility(playerID, "BLOCK") && targetID != "" {
				blockedPlayers[targetID] = true

				event := Event{
					ID:        fmt.Sprintf("night_block_%s_%s", playerID, targetID),
					Type:      EventPlayerBlocked,
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
func (nrm *NightResolutionManager) resolveMiningActions() []Event {
	var miningRequests []MiningRequest

	// Collect all mining requests from non-blocked players
	for playerID, action := range nrm.gameState.NightActions {
		if action.Type == "MINE" {
			// Check if player is blocked
			if nrm.isPlayerBlocked(playerID) {
				continue // Blocked players cannot mine
			}

			if action.TargetID != "" {
				miningRequests = append(miningRequests, MiningRequest{
					MinerID:  playerID,
					TargetID: action.TargetID,
				})
			}
		}
	}

	// Use mining manager to resolve requests
	miningManager := NewMiningManager(nrm.gameState)
	result := miningManager.ResolveMining(miningRequests)

	// Generate events for mining results
	return miningManager.UpdatePlayerTokens(result)
}

// resolveRoleAbilities handles role-specific abilities (audit, overclock, etc.)
func (nrm *NightResolutionManager) resolveRoleAbilities() []Event {
	var events []Event

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

// resolveOtherNightActions handles investigate, protect, and convert actions
func (nrm *NightResolutionManager) resolveOtherNightActions() []Event {
	var events []Event

	for playerID, action := range nrm.gameState.NightActions {
		// Skip if player is blocked
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
		case "CONVERT":
			if nrm.canPlayerUseAbility(playerID, "CONVERT") {
				events = append(events, nrm.resolveConvertAction(playerID, action)...)
			}
		}
	}

	return events
}

// resolveInvestigateAction handles investigation abilities
func (nrm *NightResolutionManager) resolveInvestigateAction(playerID string, action *SubmittedNightAction) Event {
	targetID := action.TargetID
	target := nrm.gameState.Players[targetID]

	var roleType string
	if target.Role != nil {
		roleType = string(target.Role.Type)
	} else {
		roleType = "UNKNOWN"
	}

	// Reveal target's alignment to investigator
	event := Event{
		ID:        fmt.Sprintf("night_investigate_%s_%s", playerID, targetID),
		Type:      EventPlayerInvestigated,
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
func (nrm *NightResolutionManager) resolveProtectAction(playerID string, action *SubmittedNightAction) Event {
	targetID := action.TargetID

	// Mark player as protected for tonight
	if nrm.gameState.ProtectedPlayersTonight == nil {
		nrm.gameState.ProtectedPlayersTonight = make(map[string]bool)
	}
	nrm.gameState.ProtectedPlayersTonight[targetID] = true

	event := Event{
		ID:        fmt.Sprintf("night_protect_%s_%s", playerID, targetID),
		Type:      EventPlayerProtected,
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
func (nrm *NightResolutionManager) resolveConvertAction(playerID string, action *SubmittedNightAction) []Event {
	targetID := action.TargetID
	target := nrm.gameState.Players[targetID]

	// Check if target is protected
	if nrm.isPlayerProtected(targetID) {
		// Conversion blocked by protection
		return []Event{{
			ID:        fmt.Sprintf("night_convert_blocked_%s_%s", playerID, targetID),
			Type:      EventSystemMessage,
			GameID:    nrm.gameState.ID,
			PlayerID:  playerID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"message": "Conversion attempt blocked by protection",
			},
		}}
	}

	// Calculate conversion success based on AI Equity vs Player Tokens
	player := nrm.gameState.Players[playerID]
	if player.AIEquity > target.Tokens {
		// Successful conversion
		return []Event{{
			ID:        fmt.Sprintf("night_convert_success_%s_%s", playerID, targetID),
			Type:      EventAIConversionSuccess,
			GameID:    nrm.gameState.ID,
			PlayerID:  targetID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"converter_id": playerID,
				"target_id":    targetID,
			},
		}}
	} else {
		// System shock - proves target is human
		return []Event{{
			ID:        fmt.Sprintf("night_convert_shock_%s_%s", playerID, targetID),
			Type:      EventPlayerShocked,
			GameID:    nrm.gameState.ID,
			PlayerID:  targetID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"converter_id": playerID,
				"target_id":    targetID,
				"reason":       "System shock from failed conversion",
			},
		}}
	}
}

// createNightResolutionSummary creates a summary event of all night actions
func (nrm *NightResolutionManager) createNightResolutionSummary(resolvedEvents []Event) Event {
	// Count different types of actions
	summary := map[string]interface{}{
		"total_actions":   len(nrm.gameState.NightActions),
		"resolved_events": len(resolvedEvents),
		"phase_end":       true,
	}

	return Event{
		ID:        fmt.Sprintf("night_resolution_summary_%d", nrm.gameState.DayNumber),
		Type:      EventNightActionsResolved,
		GameID:    nrm.gameState.ID,
		PlayerID:  "",
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

	// Check if player has required milestones (simplified)
	return player.ProjectMilestones >= 3
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
