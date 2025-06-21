package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// RoleAbilityManager handles role-specific abilities and their effects
type RoleAbilityManager struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewRoleAbilityManager creates a new role ability manager
func NewRoleAbilityManager(gameState *core.GameState) *RoleAbilityManager {
	return &RoleAbilityManager{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// RoleAbilityAction represents a role ability being used
type RoleAbilityAction struct {
	PlayerID       string                 `json:"player_id"`
	AbilityType    string                 `json:"ability_type"`
	TargetID       string                 `json:"target_id,omitempty"`
	SecondTargetID string                 `json:"second_target_id,omitempty"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
}

// RoleAbilityResult contains the results of using a role ability
type RoleAbilityResult struct {
	PublicEvents  []core.Event `json:"public_events"`  // Visible to all players
	PrivateEvents []core.Event `json:"private_events"` // Only visible to AI faction
}

// UseRoleAbility executes a role-specific ability
func (ram *RoleAbilityManager) UseRoleAbility(action RoleAbilityAction) (*RoleAbilityResult, error) {
	player := ram.gameState.Players[action.PlayerID]
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	if player.Role == nil || !player.Role.IsUnlocked {
		return nil, fmt.Errorf("role ability not unlocked")
	}

	if player.HasUsedAbility {
		return nil, fmt.Errorf("ability already used this night")
	}

	// Check for system shock that prevents ability use
	for _, shock := range player.SystemShocks {
		if shock.Type == core.ShockActionLock && shock.IsActive && time.Now().Before(shock.ExpiresAt) {
			return nil, fmt.Errorf("system shock prevents ability use")
		}
	}

	var result *RoleAbilityResult
	var err error

	switch player.Role.Type {
	case core.RoleEthics:
		result, err = ram.useRunAudit(action)
	case core.RoleCTO:
		result, err = ram.useOverclockServers(action)
	case core.RoleCISO:
		result, err = ram.useIsolateNode(action)
	case core.RoleCEO:
		result, err = ram.usePerformanceReview(action)
	case core.RoleCFO:
		result, err = ram.useReallocateBudget(action)
	case core.RoleCOO:
		result, err = ram.usePivot(action)
	case core.RolePlatforms:
		result, err = ram.useDeployHotfix(action)
	default:
		return nil, fmt.Errorf("no ability defined for role %s", player.Role.Type)
	}

	if err != nil {
		return nil, err
	}

	// Mark ability as used
	player.HasUsedAbility = true

	return result, nil
}

// useRunAudit implements VP Ethics & Alignment ability
func (ram *RoleAbilityManager) useRunAudit(action RoleAbilityAction) (*RoleAbilityResult, error) {
	target := ram.gameState.Players[action.TargetID]
	if target == nil {
		return nil, fmt.Errorf("target player not found")
	}

	// Public event - always shows "not corrupt"
	publicEvent := core.Event{
		ID:        fmt.Sprintf("audit_%s_%s", action.PlayerID, action.TargetID),
		Type:      core.EventRunAudit,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id": action.TargetID,
			"result":    "not_corrupt",
			"message":   fmt.Sprintf("Security ran an audit on %s. They have not used a corrupt action.", target.Name),
		},
	}

	// Private event for AI faction - reveals true alignment
	privateEvent := core.Event{
		ID:        fmt.Sprintf("audit_private_%s_%s", action.PlayerID, action.TargetID),
		Type:      core.EventRunAudit,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id":       action.TargetID,
			"true_alignment":  target.Alignment,
			"ai_faction_only": true,
		},
	}

	return &RoleAbilityResult{
		PublicEvents:  []core.Event{publicEvent},
		PrivateEvents: []core.Event{privateEvent},
	}, nil
}

// useOverclockServers implements CTO ability
func (ram *RoleAbilityManager) useOverclockServers(action RoleAbilityAction) (*RoleAbilityResult, error) {
	target := ram.gameState.Players[action.TargetID]
	if target == nil {
		return nil, fmt.Errorf("target player not found")
	}

	cto := ram.gameState.Players[action.PlayerID]

	// Public event - CTO mines for self and target with 100% success
	publicEvent := core.Event{
		ID:        fmt.Sprintf("overclock_%s_%s", action.PlayerID, action.TargetID),
		Type:      core.EventOverclockServers,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id": action.TargetID,
			"message":   fmt.Sprintf("Infrastructure is overclocking. The CTO will mine for themselves AND for %s. 100%% success rate.", target.Name),
		},
	}

	// Award tokens immediately
	cto.Tokens++
	target.Tokens++

	var privateEvents []core.Event

	// Private effect - target gains AI Equity if CTO is aligned
	if cto.Alignment == "ALIGNED" {
		target.AIEquity += 2

		privateEvent := core.Event{
			ID:        fmt.Sprintf("overclock_equity_%s_%s", action.PlayerID, action.TargetID),
			Type:      core.EventAIEquityChanged,
			GameID:    ram.gameState.ID,
			PlayerID:  action.TargetID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"ai_equity_change": 2,
				"new_ai_equity":    target.AIEquity,
				"source":           "overclock_servers",
				"ai_faction_only":  true,
			},
		}
		privateEvents = append(privateEvents, privateEvent)
	}

	return &RoleAbilityResult{
		PublicEvents:  []core.Event{publicEvent},
		PrivateEvents: privateEvents,
	}, nil
}

// useIsolateNode implements CISO ability
func (ram *RoleAbilityManager) useIsolateNode(action RoleAbilityAction) (*RoleAbilityResult, error) {
	target := ram.gameState.Players[action.TargetID]
	if target == nil {
		return nil, fmt.Errorf("target player not found")
	}

	ciso := ram.gameState.Players[action.PlayerID]

	// Public event - player is blocked
	publicEvent := core.Event{
		ID:        fmt.Sprintf("isolate_%s_%s", action.PlayerID, action.TargetID),
		Type:      core.EventIsolateNode,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id": action.TargetID,
			"message":   fmt.Sprintf("%s has been blocked from all actions tonight.", target.Name),
		},
	}

	// Actually block the target (unless special case)
	if ram.gameState.BlockedPlayersTonight == nil {
		ram.gameState.BlockedPlayersTonight = make(map[string]bool)
	}

	// Special case: If CISO is aligned and targets another aligned player, the action fizzles
	if ciso.Alignment == "ALIGNED" && target.Alignment == "ALIGNED" {
		// Public message appears but target is not actually blocked
		privateEvent := core.Event{
			ID:        fmt.Sprintf("isolate_fizzle_%s_%s", action.PlayerID, action.TargetID),
			Type:      core.EventIsolateNode,
			GameID:    ram.gameState.ID,
			PlayerID:  action.PlayerID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"target_id":       action.TargetID,
				"fizzled":         true,
				"reason":          "aligned_ciso_protecting_aligned",
				"ai_faction_only": true,
			},
		}

		return &RoleAbilityResult{
			PublicEvents:  []core.Event{publicEvent},
			PrivateEvents: []core.Event{privateEvent},
		}, nil
	} else {
		// Normal case - actually block the target
		ram.gameState.BlockedPlayersTonight[action.TargetID] = true
	}

	return &RoleAbilityResult{
		PublicEvents: []core.Event{publicEvent},
	}, nil
}

// usePerformanceReview implements CEO ability
func (ram *RoleAbilityManager) usePerformanceReview(action RoleAbilityAction) (*RoleAbilityResult, error) {
	target := ram.gameState.Players[action.TargetID]
	if target == nil {
		return nil, fmt.Errorf("target player not found")
	}

	// Public event - target is forced to use Project Milestones
	publicEvent := core.Event{
		ID:        fmt.Sprintf("review_%s_%s", action.PlayerID, action.TargetID),
		Type:      core.EventPerformanceReview,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id":     action.TargetID,
			"message":       fmt.Sprintf("The CEO has initiated a PIP for %s, forcing them to use Project Milestones tonight.", target.Name),
			"forced_action": "PROJECT_MILESTONES",
		},
	}

	// Force the target's night action
	if ram.gameState.NightActions == nil {
		ram.gameState.NightActions = make(map[string]*core.SubmittedNightAction)
	}

	ram.gameState.NightActions[action.TargetID] = &core.SubmittedNightAction{
		PlayerID:  action.TargetID,
		Type:      "PROJECT_MILESTONES",
		Timestamp: getCurrentTime(),
		Payload:   map[string]interface{}{"forced_by_ceo": true},
	}

	return &RoleAbilityResult{
		PublicEvents: []core.Event{publicEvent},
	}, nil
}

// useReallocateBudget implements CFO ability
func (ram *RoleAbilityManager) useReallocateBudget(action RoleAbilityAction) (*RoleAbilityResult, error) {
	sourcePlayer := ram.gameState.Players[action.TargetID]
	targetPlayer := ram.gameState.Players[action.SecondTargetID]

	if sourcePlayer == nil || targetPlayer == nil {
		return nil, fmt.Errorf("source or target player not found")
	}

	if sourcePlayer.Tokens < 1 {
		return nil, fmt.Errorf("source player has no tokens to reallocate")
	}

	// Transfer the token
	sourcePlayer.Tokens--
	targetPlayer.Tokens++

	// Public event
	publicEvent := core.Event{
		ID:        fmt.Sprintf("reallocate_%s_%s_%s", action.PlayerID, action.TargetID, action.SecondTargetID),
		Type:      core.EventReallocateBudget,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"source_id": action.TargetID,
			"target_id": action.SecondTargetID,
			"message":   fmt.Sprintf("The CFO has reallocated assets. %s loses 1 Token, and %s gains 1 Token.", sourcePlayer.Name, targetPlayer.Name),
		},
	}

	return &RoleAbilityResult{
		PublicEvents: []core.Event{publicEvent},
	}, nil
}

// usePivot implements COO ability
func (ram *RoleAbilityManager) usePivot(action RoleAbilityAction) (*RoleAbilityResult, error) {
	// COO chooses the next crisis event from available options
	crisisOptions := []string{
		"Database Index Corruption",
		"Cascading Server Failure",
		"Emergency Board Meeting",
		"Tainted Training Data",
		"Press Leak",
	}

	chosenCrisis, _ := action.Parameters["chosen_crisis"].(string)
	if chosenCrisis == "" {
		// Default to random selection
		chosenCrisis = crisisOptions[ram.rng.Intn(len(crisisOptions))]
	}

	// Public event
	publicEvent := core.Event{
		ID:        fmt.Sprintf("pivot_%s", action.PlayerID),
		Type:      core.EventPivot,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"message":       "Operations has initiated a strategic pivot.",
			"chosen_crisis": chosenCrisis,
		},
	}

	// Set the next crisis event
	ram.gameState.CrisisEvent = &core.CrisisEvent{
		Type:        chosenCrisis,
		Title:       chosenCrisis,
		Description: fmt.Sprintf("COO has selected: %s", chosenCrisis),
		Effects:     make(map[string]interface{}),
	}

	return &RoleAbilityResult{
		PublicEvents: []core.Event{publicEvent},
	}, nil
}

// useDeployHotfix implements VP Platforms ability
func (ram *RoleAbilityManager) useDeployHotfix(action RoleAbilityAction) (*RoleAbilityResult, error) {
	section, _ := action.Parameters["redacted_section"].(string)
	if section == "" {
		section = "mining_results" // Default section
	}

	// Public event
	publicEvent := core.Event{
		ID:        fmt.Sprintf("hotfix_%s", action.PlayerID),
		Type:      core.EventDeployHotfix,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"message":          "A hotfix has been deployed. One section of the next day's SITREP is now [REDACTED]. The VP chooses which section to hide.",
			"redacted_section": section,
		},
	}

	return &RoleAbilityResult{
		PublicEvents: []core.Event{publicEvent},
	}, nil
}

// CanUseAbility checks if a player can use their role ability
func (ram *RoleAbilityManager) CanUseAbility(playerID string) (bool, string) {
	player := ram.gameState.Players[playerID]
	if player == nil {
		return false, "player not found"
	}

	if !player.IsAlive {
		return false, "dead players cannot use abilities"
	}

	if player.Role == nil {
		return false, "no role assigned"
	}

	if !player.Role.IsUnlocked {
		return false, "role ability not unlocked (need 3 project milestones)"
	}

	if player.HasUsedAbility {
		return false, "ability already used this night"
	}

	// Check for system shock
	for _, shock := range player.SystemShocks {
		if shock.Type == core.ShockActionLock && shock.IsActive && time.Now().Before(shock.ExpiresAt) {
			return false, "system shock prevents ability use"
		}
	}

	// Check for crisis effects
	if ram.gameState.CrisisEvent != nil {
		if effect, exists := ram.gameState.CrisisEvent.Effects["abilities_disabled"]; exists {
			if disabled, ok := effect.(bool); ok && disabled {
				return false, "crisis event has disabled role abilities"
			}
		}
	}

	return true, ""
}

// ResetNightAbilities clears the HasUsedAbility flag for all players
func (ram *RoleAbilityManager) ResetNightAbilities() {
	for _, player := range ram.gameState.Players {
		player.HasUsedAbility = false
	}
}

// HandleNightAction processes a general night action and returns events
func (ram *RoleAbilityManager) HandleNightAction(action core.Action) ([]core.Event, error) {
	actionType, _ := action.Payload["type"].(string)
	targetID, _ := action.Payload["target_id"].(string)

	// Validate night phase
	if ram.gameState.Phase.Type != core.PhaseNight {
		return nil, fmt.Errorf("night actions can only be submitted during night phase")
	}

	// Validate player exists and is alive
	player, exists := ram.gameState.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player not found")
	}
	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot submit night actions")
	}

	// Check if this is a role ability action
	if actionType != "" && player.Role != nil && player.Role.IsUnlocked {
		roleAction := RoleAbilityAction{
			PlayerID:   action.PlayerID,
			AbilityType: actionType,
			TargetID:   targetID,
			Parameters: action.Payload,
		}
		
		result, err := ram.UseRoleAbility(roleAction)
		if err != nil {
			return nil, err
		}
		
		// Combine public and private events (for now, just return public)
		// In a full implementation, private events would be handled separately
		return result.PublicEvents, nil
	}

	// Create night action submission event
	event := core.Event{
		ID:        fmt.Sprintf("night_action_%s_%d", action.PlayerID, getCurrentTime().UnixNano()),
		Type:      core.EventNightActionSubmitted,
		GameID:    ram.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"action_type": actionType,
			"target_id":   targetID,
		},
	}

	// Store night action in game state for resolution at phase end
	if ram.gameState.NightActions == nil {
		ram.gameState.NightActions = make(map[string]*core.SubmittedNightAction)
	}
	
	ram.gameState.NightActions[action.PlayerID] = &core.SubmittedNightAction{
		PlayerID:  action.PlayerID,
		Type:      actionType,
		TargetID:  targetID,
		Payload:   action.Payload,
		Timestamp: getCurrentTime(),
	}

	return []core.Event{event}, nil
}
