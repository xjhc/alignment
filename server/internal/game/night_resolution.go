package game

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// NightResolutionManager handles the resolution of all night actions
type NightResolutionManager struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewNightResolutionManager creates a new night resolution manager
func NewNightResolutionManager(gameState *core.GameState) *NightResolutionManager {
	return &NightResolutionManager{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ResolveNightActions processes all submitted night actions in precedence order
func (nrm *NightResolutionManager) ResolveNightActions() []core.Event {
	// Default to mining for players who didn't submit an action
	for playerID, player := range nrm.gameState.Players {
		if player.IsAlive {
			if _, submitted := nrm.gameState.NightActions[playerID]; !submitted {
				// Find a random living target that is not the player themselves
				var possibleTargets []string
				for otherPlayerID, otherPlayer := range nrm.gameState.Players {
					if otherPlayer.IsAlive && otherPlayerID != playerID {
						possibleTargets = append(possibleTargets, otherPlayerID)
					}
				}

				if len(possibleTargets) > 0 {
					targetID := possibleTargets[nrm.rng.Intn(len(possibleTargets))]
					if nrm.gameState.NightActions == nil {
						nrm.gameState.NightActions = make(map[string]*core.SubmittedNightAction)
					}
					nrm.gameState.NightActions[playerID] = &core.SubmittedNightAction{
						PlayerID:  playerID,
						Type:      "MINE",
						TargetID:  targetID,
						Timestamp: time.Now(),
					}
					log.Printf("Player %s defaulted to mining for %s", player.Name, targetID)
				}
			}
		}
	}

	if nrm.gameState.NightActions == nil || len(nrm.gameState.NightActions) == 0 {
		log.Printf("No night actions to resolve")
		return []core.Event{}
	}

	// Process night actions internally to determine results without emitting granular events
	results := nrm.processAllNightActions()

	// Generate single authoritative night resolution event
	summaryEvent := nrm.createNightResolutionSummary(results)

	// Clear night actions for next night
	nrm.gameState.NightActions = make(map[string]*core.SubmittedNightAction)

	return []core.Event{summaryEvent}
}

// NightActionResults holds the structured results of all night actions
type NightActionResults struct {
	BlockedPlayers      []map[string]interface{} `json:"blocked_players"`
	ConvertedPlayers    []map[string]interface{} `json:"converted_players"`
	ShockedPlayers      []map[string]interface{} `json:"shocked_players"`
	EliminatedPlayers   []map[string]interface{} `json:"eliminated_players"`
	MiningResults       []map[string]interface{} `json:"mining_results"`
	RoleAbilityResults  []map[string]interface{} `json:"role_ability_results"`
	MilestoneResults    []map[string]interface{} `json:"milestone_results"`
	FailedActions       []map[string]interface{} `json:"failed_actions"`
	PlayerStateChanges  map[string]map[string]interface{} `json:"player_state_changes"`
}

// processAllNightActions handles all night actions and returns structured results
func (nrm *NightResolutionManager) processAllNightActions() *NightActionResults {
	results := &NightActionResults{
		BlockedPlayers:      []map[string]interface{}{},
		ConvertedPlayers:    []map[string]interface{}{},
		ShockedPlayers:      []map[string]interface{}{},
		EliminatedPlayers:   []map[string]interface{}{},
		MiningResults:       []map[string]interface{}{},
		RoleAbilityResults:  []map[string]interface{}{},
		MilestoneResults:    []map[string]interface{}{},
		FailedActions:       []map[string]interface{}{},
		PlayerStateChanges:  make(map[string]map[string]interface{}),
	}

	// Pass 1: Process blocking actions (highest precedence)
	nrm.processBlockActions(results)

	// Pass 2: Process AI conversion attempts
	nrm.processConversionActions(results)

	// Pass 3: Process Intern actions (BOOTCAMP and SHADOW)
	nrm.processInternActions(results)

	// Pass 4: Process standard actions (mining, role abilities, others)
	nrm.processStandardActions(results)

	return results
}

// processBlockActions handles all blocking actions and updates results
func (nrm *NightResolutionManager) processBlockActions(results *NightActionResults) {
	for playerID, action := range nrm.gameState.NightActions {
		if action.Type == "BLOCK" || action.Type == "ISOLATE_NODE" {
			targetID := action.TargetID

			// Validate block action
			if nrm.canPlayerUseAbility(playerID, action.Type) && targetID != "" {
				// Block the target player
				nrm.blockPlayer(targetID)

				// Get player names for structured results
				blocker := nrm.gameState.Players[playerID]
				target := nrm.gameState.Players[targetID]

				if blocker != nil && target != nil {
					results.BlockedPlayers = append(results.BlockedPlayers, map[string]interface{}{
						"player_id":     targetID,
						"player_name":   target.Name,
						"blocker_id":    playerID,
						"blocker_name":  blocker.Name,
						"block_type":    action.Type,
					})

					// Track state change for target
					nrm.addPlayerStateChange(results, targetID, "status_message", fmt.Sprintf("Action blocked by %s", blocker.Name))
				}
			}
		}
	}
}

// processConversionActions handles AI conversion attempts
func (nrm *NightResolutionManager) processConversionActions(results *NightActionResults) {
	for playerID, action := range nrm.gameState.NightActions {
		if action.Type == "CONVERT" {
			targetID := action.TargetID

			// Skip blocked players
			if nrm.isPlayerBlocked(playerID) {
				continue
			}

			// Validate conversion action
			if nrm.canPlayerUseAbility(playerID, "CONVERT") && targetID != "" {
				player := nrm.gameState.Players[playerID]
				target := nrm.gameState.Players[targetID]

				if player != nil && target != nil && player.Alignment == "ALIGNED" {
					// Calculate conversion success - use AI equity from player
					success := core.CalculateAIConversionSuccess(*target, player.AIEquity, *nrm.gameState)

					if success {
						// Successful conversion
						results.ConvertedPlayers = append(results.ConvertedPlayers, map[string]interface{}{
							"player_id":        targetID,
							"player_name":      target.Name,
							"converter_id":     playerID,
							"converter_name":   player.Name,
							"previous_equity":  target.AIEquity,
							"new_equity":       0, // Reset after successful conversion
						})

						// Track state changes
						nrm.addPlayerStateChange(results, targetID, "alignment", "ALIGNED")
						nrm.addPlayerStateChange(results, targetID, "ai_equity", 0)
						nrm.addPlayerStateChange(results, targetID, "status_message", "Conversion successful")
					} else {
						// Failed conversion - system shock
						shockMessage := fmt.Sprintf("System shock: Failed AI conversion by %s", player.Name)

						results.ShockedPlayers = append(results.ShockedPlayers, map[string]interface{}{
							"player_id":      targetID,
							"player_name":    target.Name,
							"shock_type":     "CONVERSION_FAILURE",
							"shock_duration": 24,
							"reason":         "Failed AI conversion attempt",
							"converter_id":   playerID,
							"converter_name": player.Name,
						})

						// Track state changes
						nrm.addPlayerStateChange(results, targetID, "ai_equity", 0)
						nrm.addPlayerStateChange(results, targetID, "status_message", shockMessage)
					}
				}
			}
		}
	}
}

// processInternActions handles BOOTCAMP and SHADOW actions for the Intern role
func (nrm *NightResolutionManager) processInternActions(results *NightActionResults) {
	for playerID, action := range nrm.gameState.NightActions {
		// Skip blocked players
		if nrm.isPlayerBlocked(playerID) {
			continue
		}

		player := nrm.gameState.Players[playerID]
		if player == nil || player.Role == nil || player.Role.Type != core.RoleIntern {
			continue
		}

		switch action.Type {
		case "BOOTCAMP":
			nrm.processBootcampAction(playerID, action, results)
		case "SHADOW":
			nrm.processShadowAction(playerID, action, results)
		}
	}
}

// processStandardActions handles mining, role abilities, and other standard actions
func (nrm *NightResolutionManager) processStandardActions(results *NightActionResults) {
	for playerID, action := range nrm.gameState.NightActions {
		// Skip blocked players for standard actions
		if nrm.isPlayerBlocked(playerID) {
			continue
		}

		switch action.Type {
		case "MINE":
			nrm.processMiningAction(playerID, action, results)
		case "PROJECT_MILESTONE":
			nrm.processProjectMilestoneAction(playerID, action, results)
		default:
			// Handle role abilities
			if nrm.isRoleAbility(action.Type) {
				nrm.processRoleAbilityAction(playerID, action, results)
			}
		}
	}
}

// processBootcampAction handles BOOTCAMP action - grants the Intern 1 Bootcamp Point
func (nrm *NightResolutionManager) processBootcampAction(playerID string, action *core.SubmittedNightAction, results *NightActionResults) {
	player := nrm.gameState.Players[playerID]
	if player == nil {
		return
	}

	// Increment Bootcamp Points
	player.BootcampPoints++

	results.RoleAbilityResults = append(results.RoleAbilityResults, map[string]interface{}{
		"player_id":          playerID,
		"player_name":        player.Name,
		"ability_type":       "BOOTCAMP",
		"message":            fmt.Sprintf("%s completed bootcamp training and gained 1 Bootcamp Point", player.Name),
		"bootcamp_points":    player.BootcampPoints,
	})

	// Track state change
	nrm.addPlayerStateChange(results, playerID, "bootcamp_points", player.BootcampPoints)
}

// processShadowAction handles SHADOW action - copies another player's ability
func (nrm *NightResolutionManager) processShadowAction(playerID string, action *core.SubmittedNightAction, results *NightActionResults) {
	player := nrm.gameState.Players[playerID]
	if player == nil {
		return
	}

	// Check if Intern has enough Bootcamp Points
	if player.BootcampPoints < 1 {
		results.FailedActions = append(results.FailedActions, map[string]interface{}{
			"player_id":   playerID,
			"player_name": player.Name,
			"action_type": "SHADOW",
			"reason":      "Insufficient Bootcamp Points",
		})
		return
	}

	// Get shadow target (player whose ability to copy)
	shadowTargetID := action.TargetID
	shadowTarget := nrm.gameState.Players[shadowTargetID]
	if shadowTarget == nil || shadowTarget.Role == nil || !shadowTarget.Role.IsUnlocked {
		results.FailedActions = append(results.FailedActions, map[string]interface{}{
			"player_id":   playerID,
			"player_name": player.Name,
			"action_type": "SHADOW",
			"reason":      "Target does not have an unlocked ability to shadow",
		})
		return
	}

	// Check project milestone requirement (3+ milestones)
	if shadowTarget.ProjectMilestones < 3 {
		results.FailedActions = append(results.FailedActions, map[string]interface{}{
			"player_id":   playerID,
			"player_name": player.Name,
			"action_type": "SHADOW",
			"reason":      "Target's ability is not unlocked (needs 3+ project milestones)",
		})
		return
	}

	// Get final target from payload
	var finalTargetID string
	if payload, ok := action.Payload["shadow_target_id"].(string); ok {
		finalTargetID = payload
	}

	// Consume Bootcamp Point
	player.BootcampPoints--

	// Determine the ability type based on the shadow target's role
	var abilityType string
	switch shadowTarget.Role.Type {
	case core.RoleCISO:
		abilityType = "ISOLATE_NODE"
	case core.RoleCTO:
		abilityType = "OVERCLOCK_SERVERS"
	case core.RoleEthics:
		abilityType = "RUN_AUDIT"
	case core.RoleCEO:
		abilityType = "PERFORMANCE_REVIEW"
	case core.RoleCFO:
		abilityType = "REALLOCATE_BUDGET"
	case core.RoleCOO:
		abilityType = "PIVOT"
	case core.RolePlatforms:
		abilityType = "DEPLOY_HOTFIX"
	default:
		results.FailedActions = append(results.FailedActions, map[string]interface{}{
			"player_id":   playerID,
			"player_name": player.Name,
			"action_type": "SHADOW",
			"reason":      "Unknown or unshadowable role ability",
		})
		return
	}

	// Create a RoleAbilityAction with the Intern's ID but using the copied ability
	// This ensures the Intern's alignment determines the effect version
	roleAbilityManager := NewRoleAbilityManager(nrm.gameState)
	roleAbilityAction := &RoleAbilityAction{
		PlayerID:    playerID, // CRITICAL: Use Intern's ID for alignment determination
		AbilityType: abilityType,
		TargetID:    finalTargetID,
	}

	// Handle special cases that need additional parameters
	if abilityType == "REALLOCATE_BUDGET" {
		if secondTarget, ok := action.Payload["second_target_id"].(string); ok {
			roleAbilityAction.SecondTargetID = secondTarget
		}
	} else if abilityType == "PIVOT" {
		if chosenCrisis, ok := action.Payload["chosen_crisis"].(string); ok {
			roleAbilityAction.Parameters = map[string]interface{}{"chosen_crisis": chosenCrisis}
		}
	} else if abilityType == "DEPLOY_HOTFIX" {
		if section, ok := action.Payload["redacted_section"].(string); ok {
			roleAbilityAction.Parameters = map[string]interface{}{"redacted_section": section}
		}
	}

	// Execute the copied ability using the RoleAbilityManager
	// The RoleAbilityManager will use the Intern's alignment to determine the effect
	result, err := roleAbilityManager.UseRoleAbility(*roleAbilityAction)
	if err != nil {
		results.FailedActions = append(results.FailedActions, map[string]interface{}{
			"player_id":   playerID,
			"player_name": player.Name,
			"action_type": "SHADOW",
			"reason":      fmt.Sprintf("Failed to execute copied ability: %v", err),
		})
		return
	}

	// Add the shadow action result
	results.RoleAbilityResults = append(results.RoleAbilityResults, map[string]interface{}{
		"player_id":          playerID,
		"player_name":        player.Name,
		"ability_type":       fmt.Sprintf("SHADOW_%s", abilityType),
		"shadowed_player":    shadowTarget.Name,
		"target_id":          finalTargetID,
		"message":            fmt.Sprintf("%s used %s (shadowed from %s)", player.Name, abilityType, shadowTarget.Name),
		"bootcamp_points":    player.BootcampPoints,
	})

	// Track state change for Bootcamp Points
	nrm.addPlayerStateChange(results, playerID, "bootcamp_points", player.BootcampPoints)

	// Process the ability result events
	if result != nil {
		// Note: The events from the ability execution will be handled by the RoleAbilityManager
		// and will appear to come from the Intern in public logs
	}
}

// Helper functions for processing night actions

// addPlayerStateChange tracks a state change for a player
func (nrm *NightResolutionManager) addPlayerStateChange(results *NightActionResults, playerID, key string, value interface{}) {
	if results.PlayerStateChanges[playerID] == nil {
		results.PlayerStateChanges[playerID] = make(map[string]interface{})
	}
	results.PlayerStateChanges[playerID][key] = value
}

// blockPlayer marks a player as blocked for the night
func (nrm *NightResolutionManager) blockPlayer(playerID string) {
	if nrm.gameState.BlockedPlayersTonight == nil {
		nrm.gameState.BlockedPlayersTonight = make(map[string]bool)
	}
	nrm.gameState.BlockedPlayersTonight[playerID] = true
}

// isPlayerBlocked checks if a player is blocked for the night
func (nrm *NightResolutionManager) isPlayerBlocked(playerID string) bool {
	if nrm.gameState.BlockedPlayersTonight == nil {
		return false
	}
	return nrm.gameState.BlockedPlayersTonight[playerID]
}

// isRoleAbility checks if an action type is a role ability
func (nrm *NightResolutionManager) isRoleAbility(actionType string) bool {
	roleAbilities := []string{
		"RUN_AUDIT", "OVERCLOCK_SERVERS", "ISOLATE_NODE",
		"PERFORMANCE_REVIEW", "REALLOCATE_BUDGET", "PIVOT", "DEPLOY_HOTFIX",
	}
	for _, ability := range roleAbilities {
		if actionType == ability {
			return true
		}
	}
	return false
}

// processMiningAction handles mining actions and updates results
func (nrm *NightResolutionManager) processMiningAction(playerID string, action *core.SubmittedNightAction, results *NightActionResults) {
	targetID := action.TargetID
	player := nrm.gameState.Players[playerID]
	target := nrm.gameState.Players[targetID]

	if player != nil && target != nil && targetID != playerID {
		// Calculate mining success
		success := core.CalculateMiningSuccess(*player, 0.2, *nrm.gameState) // Standard difficulty

		if success {
			// Successful mining
			reward := core.CalculateTokenReward(core.EventMiningSuccessful, *player, *nrm.gameState)

			results.MiningResults = append(results.MiningResults, map[string]interface{}{
				"miner_id":     playerID,
				"miner_name":   player.Name,
				"target_id":    targetID,
				"target_name":  target.Name,
				"tokens_mined": reward,
				"success":      true,
			})

			// Track state change for target (gains tokens)
			nrm.addPlayerStateChange(results, targetID, "tokens_gained", reward)
		} else {
			// Failed mining
			results.FailedActions = append(results.FailedActions, map[string]interface{}{
				"player_id":   playerID,
				"player_name": player.Name,
				"action_type": "MINE",
				"target_id":   targetID,
				"reason":      "Mining attempt failed",
			})
		}
	}
}

// processProjectMilestoneAction handles project milestone advancement
func (nrm *NightResolutionManager) processProjectMilestoneAction(playerID string, action *core.SubmittedNightAction, results *NightActionResults) {
	player := nrm.gameState.Players[playerID]
	if player == nil {
		return
	}

	// Advance milestone
	newMilestones := player.ProjectMilestones + 1
	roleUnlocked := newMilestones >= 3 && (player.Role == nil || !player.Role.IsUnlocked)

	results.MilestoneResults = append(results.MilestoneResults, map[string]interface{}{
		"player_id":        playerID,
		"player_name":      player.Name,
		"milestones_count": newMilestones,
		"role_unlocked":    roleUnlocked,
		"message":          fmt.Sprintf("%s advanced to %d project milestones", player.Name, newMilestones),
	})

	// Track state changes
	nrm.addPlayerStateChange(results, playerID, "project_milestones", newMilestones)
	if roleUnlocked {
		nrm.addPlayerStateChange(results, playerID, "role_unlocked", true)
	}
}

// processRoleAbilityAction handles role ability usage
func (nrm *NightResolutionManager) processRoleAbilityAction(playerID string, action *core.SubmittedNightAction, results *NightActionResults) {
	player := nrm.gameState.Players[playerID]
	if player == nil || player.Role == nil || !player.Role.IsUnlocked || player.HasUsedAbility {
		return
	}

	// Process the specific role ability
	results.RoleAbilityResults = append(results.RoleAbilityResults, map[string]interface{}{
		"player_id":    playerID,
		"player_name":  player.Name,
		"ability_type": action.Type,
		"target_id":    action.TargetID,
		"message":      fmt.Sprintf("%s used %s", player.Name, action.Type),
	})

	// Track that ability was used
	nrm.addPlayerStateChange(results, playerID, "has_used_ability", true)
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
					ID:        fmt.Sprintf("ai_conversion_blocked_crisis_%s_%s", playerID, targetID),
					Type:      core.EventAIConversionBlocked,
					GameID:    nrm.gameState.ID,
					PlayerID:  "", // Public event
					Timestamp: getCurrentTime(),
					Payload: map[string]interface{}{
						"converter_id":     playerID,
						"target_id":        targetID,
						"blocking_reason":  "CRISIS",
						"blocking_source":  nrm.gameState.CrisisEvent.Title,
						"blocking_details": map[string]interface{}{
							"crisis_type": nrm.gameState.CrisisEvent.Type,
						},
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
						ID:        fmt.Sprintf("ai_conversion_blocked_mandate_%s_%s", playerID, targetID),
						Type:      core.EventAIConversionBlocked,
						GameID:    nrm.gameState.ID,
						PlayerID:  "", // Public event
						Timestamp: getCurrentTime(),
						Payload: map[string]interface{}{
							"converter_id":     playerID,
							"target_id":        targetID,
							"blocking_reason":  "MANDATE",
							"blocking_source":  nrm.gameState.CorporateMandate.Name,
							"blocking_details": map[string]interface{}{
								"mandate_type": nrm.gameState.CorporateMandate.Type,
								"night_number": nightNumber,
								"odd_night_restriction": true,
							},
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
			ID:        fmt.Sprintf("ai_conversion_blocked_protection_%s_%s", playerID, targetID),
			Type:      core.EventAIConversionBlocked,
			GameID:    nrm.gameState.ID,
			PlayerID:  "", // Public event
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"converter_id":     playerID,
				"target_id":        targetID,
				"blocking_reason":  "PROTECTION",
				"blocking_source":  "Player Protection",
				"blocking_details": map[string]interface{}{
					"protection_type": "night_protection",
				},
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

// createNightResolutionSummary creates a comprehensive summary event from structured results
func (nrm *NightResolutionManager) createNightResolutionSummary(results *NightActionResults) core.Event {
	// Create comprehensive summary payload with structured information
	summary := map[string]interface{}{
		"night_number":         nrm.gameState.DayNumber,
		"total_actions":        len(nrm.gameState.NightActions),
		"blocked_players":      results.BlockedPlayers,
		"converted_players":    results.ConvertedPlayers,
		"shocked_players":      results.ShockedPlayers,
		"eliminated_players":   results.EliminatedPlayers,
		"mining_results":       results.MiningResults,
		"role_ability_results": results.RoleAbilityResults,
		"milestone_results":    results.MilestoneResults,
		"failed_actions":       results.FailedActions,
		"player_state_changes": results.PlayerStateChanges,
		"phase_end":            true,
		"next_phase":           "SITREP",
		"summary_message":      nrm.createHumanReadableSummary(results),
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
func (nrm *NightResolutionManager) createHumanReadableSummary(results *NightActionResults) string {
	summary := fmt.Sprintf("Night %d Summary:\n", nrm.gameState.DayNumber)

	if len(results.BlockedPlayers) > 0 {
		summary += "• Players blocked from actions: "
		for i, p := range results.BlockedPlayers {
			if i > 0 {
				summary += ", "
			}
			summary += p["player_name"].(string)
		}
		summary += "\n"
	}

	if len(results.ConvertedPlayers) > 0 {
		summary += "• Players converted by AI: "
		for i, p := range results.ConvertedPlayers {
			if i > 0 {
				summary += ", "
			}
			summary += p["player_name"].(string)
		}
		summary += "\n"
	}

	if len(results.ShockedPlayers) > 0 {
		summary += "• Players experienced system shock: "
		for i, p := range results.ShockedPlayers {
			if i > 0 {
				summary += ", "
			}
			summary += p["player_name"].(string)
		}
		summary += "\n"
	}

	if len(results.MiningResults) > 0 {
		summary += fmt.Sprintf("• %d successful mining operations completed\n", len(results.MiningResults))
	}

	if len(results.RoleAbilityResults) > 0 {
		summary += fmt.Sprintf("• %d role abilities were used\n", len(results.RoleAbilityResults))
	}

	if len(results.MilestoneResults) > 0 {
		summary += "• Project milestone advancement: "
		rolesUnlocked := 0
		for i, m := range results.MilestoneResults {
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