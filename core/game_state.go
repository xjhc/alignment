package core

import (
	"fmt"
	"time"
)

// GameState represents the complete state of a game
type GameState struct {
	ID              string                           `json:"id"`
	Phase           Phase                            `json:"phase"`
	DayNumber       int                              `json:"day_number"`
	Players         map[string]*Player               `json:"players"`
	CreatedAt       time.Time                        `json:"created_at"`
	UpdatedAt       time.Time                        `json:"updated_at"`
	Settings        GameSettings                     `json:"settings"`
	CrisisEvent     *CrisisEvent                     `json:"crisis_event,omitempty"`
	ChatMessages    []ChatMessage                    `json:"chat_messages"`
	VoteState       *VoteState                       `json:"vote_state,omitempty"`
	NominatedPlayer string                           `json:"nominated_player,omitempty"`
	WinCondition    *WinCondition                    `json:"win_condition,omitempty"`
	NightActions    map[string]*SubmittedNightAction `json:"night_actions,omitempty"`

	// Game-wide modifiers
	CorporateMandate *CorporateMandate `json:"corporate_mandate,omitempty"`

	// Daily tracking
	PulseCheckResponses map[string]string `json:"pulse_check_responses,omitempty"`

	// Temporary fields for night resolution (cleared each night)
	BlockedPlayersTonight   map[string]bool `json:"-"` // Not serialized
	ProtectedPlayersTonight map[string]bool `json:"-"` // Not serialized
}

// NewGameState creates a new game state
func NewGameState(id string) *GameState {
	now := time.Now()
	return &GameState{
		ID:           id,
		Phase:        Phase{Type: PhaseLobby, StartTime: now, Duration: 0},
		DayNumber:    0,
		Players:      make(map[string]*Player),
		CreatedAt:    now,
		UpdatedAt:    now,
		ChatMessages: make([]ChatMessage, 0),
		NightActions: make(map[string]*SubmittedNightAction),
		Settings: GameSettings{
			MaxPlayers:         10,
			MinPlayers:         2,
			SitrepDuration:     15 * time.Second,
			PulseCheckDuration: 30 * time.Second,
			DiscussionDuration: 2 * time.Minute,
			ExtensionDuration:  15 * time.Second,
			NominationDuration: 30 * time.Second,
			TrialDuration:      30 * time.Second,
			VerdictDuration:    30 * time.Second,
			NightDuration:      30 * time.Second,
			StartingTokens:     1,
			VotingThreshold:    0.5,
		},
	}
}

// ApplyEvent applies an event to the game state and returns a new state
func ApplyEvent(currentState GameState, event Event) GameState {
	newState := currentState
	newState.UpdatedAt = event.Timestamp

	switch event.Type {
	// Game lifecycle events
	case EventGameStarted:
		newState.applyGameStarted(event)
	case EventGameEnded:
		newState.applyGameEnded(event)
	case EventPhaseChanged:
		newState.applyPhaseChanged(event)
	case EventDayStarted:
		newState.applyDayStarted(event)
	case EventNightStarted:
		newState.applyNightStarted(event)

	// Player events
	case EventPlayerJoined:
		newState.applyPlayerJoined(event)
	case EventPlayerLeft:
		newState.applyPlayerLeft(event)
	case EventPlayerEliminated:
		newState.applyPlayerEliminated(event)
	case EventPlayerAligned:
		newState.applyPlayerAligned(event)
	case EventPlayerShocked:
		newState.applyPlayerShocked(event)
	case EventPlayerStatusChanged:
		newState.applyPlayerStatusChanged(event)
	case EventPlayerReconnected:
		newState.applyPlayerReconnected(event)
	case EventPlayerDisconnected:
		newState.applyPlayerDisconnected(event)

	// Role and ability events
	case EventRoleAssigned:
		newState.applyRoleAssigned(event)
	case EventRoleAbilityUnlocked:
		newState.applyRoleAbilityUnlocked(event)
	case EventProjectMilestone:
		newState.applyProjectMilestone(event)

	// Voting events
	case EventVoteCast:
		newState.applyVoteCast(event)
	case EventVoteStarted:
		newState.applyVoteStarted(event)
	case EventVoteCompleted:
		newState.applyVoteCompleted(event)
	case EventPlayerNominated:
		newState.applyPlayerNominated(event)

	// Token and mining events
	case EventTokensAwarded:
		newState.applyTokensAwarded(event)
	case EventTokensLost:
		newState.applyTokensLost(event)
	case EventMiningSuccessful:
		newState.applyMiningSuccessful(event)
	case EventMiningFailed:
		newState.applyMiningFailed(event)
	case EventMiningPoolUpdated:
		newState.applyMiningPoolUpdated(event)
	case EventTokensDistributed:
		newState.applyTokensDistributed(event)

	// Night action events
	case EventNightActionSubmitted:
		newState.applyNightActionSubmitted(event)
	case EventNightActionsResolved:
		newState.applyNightActionsResolved(event)
	case EventPlayerBlocked:
		newState.applyPlayerBlocked(event)
	case EventPlayerProtected:
		newState.applyPlayerProtected(event)
	case EventPlayerInvestigated:
		newState.applyPlayerInvestigated(event)

	// AI and conversion events
	case EventAIConversionAttempt:
		newState.applyAIConversionAttempt(event)
	case EventAIConversionSuccess:
		newState.applyAIConversionSuccess(event)
	case EventAIConversionFailed:
		newState.applyAIConversionFailed(event)

	// Communication events
	case EventChatMessage:
		newState.applyChatMessage(event)
	case EventSystemMessage:
		newState.applySystemMessage(event)
	case EventPrivateNotification:
		newState.applyPrivateNotification(event)

	// Crisis and pulse check events
	case EventCrisisTriggered:
		newState.applyCrisisTriggered(event)
	case EventPulseCheckStarted:
		newState.applyPulseCheckStarted(event)
	case EventPulseCheckSubmitted:
		newState.applyPulseCheckSubmitted(event)
	case EventPulseCheckRevealed:
		newState.applyPulseCheckRevealed(event)

	// Win condition events
	case EventVictoryCondition:
		newState.applyVictoryCondition(event)

	// Role ability events
	case EventRunAudit:
		newState.applyRunAudit(event)
	case EventOverclockServers:
		newState.applyOverclockServers(event)
	case EventIsolateNode:
		newState.applyIsolateNode(event)
	case EventPerformanceReview:
		newState.applyPerformanceReview(event)
	case EventReallocateBudget:
		newState.applyReallocateBudget(event)
	case EventPivot:
		newState.applyPivot(event)
	case EventDeployHotfix:
		newState.applyDeployHotfix(event)

	// Status events
	case EventSlackStatusChanged:
		newState.applySlackStatusChanged(event)
	case EventPartingShotSet:
		newState.applyPartingShotSet(event)

	// KPI events
	case EventKPIProgress:
		newState.applyKPIProgress(event)
	case EventKPICompleted:
		newState.applyKPICompleted(event)

	// System shock events
	case EventSystemShockApplied:
		newState.applySystemShockApplied(event)

	// AI equity events
	case EventAIEquityChanged:
		newState.applyAIEquityChanged(event)

	// Corporate mandate events
	case EventMandateActivated:
		newState.applyMandateActivated(event)
	case EventMandateEffect:
		newState.applyMandateEffect(event)

	// System shock effect events
	case EventShockEffectTriggered:
		newState.applyShockEffectTriggered(event)

	// Equity threshold events
	case EventEquityThreshold:
		newState.applyEquityThreshold(event)

	default:
		// Unknown event type - ignore
	}

	return newState
}

func (gs *GameState) applyGameStarted(event Event) {
	gs.Phase = Phase{
		Type:      PhaseSitrep,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.SitrepDuration,
	}
	gs.DayNumber = 1
}

func (gs *GameState) applyPlayerJoined(event Event) {
	playerID := event.PlayerID
	name, _ := event.Payload["name"].(string)
	jobTitle, _ := event.Payload["job_title"].(string)

	gs.Players[playerID] = &Player{
		ID:                playerID,
		Name:              name,
		JobTitle:          jobTitle,
		ControlType:       "HUMAN", // Default control type
		IsAlive:           true,
		Tokens:            gs.Settings.StartingTokens,
		ProjectMilestones: 0,
		StatusMessage:     "",
		JoinedAt:          event.Timestamp,
		Alignment:         "HUMAN", // Default alignment
	}
}

func (gs *GameState) applyPlayerLeft(event Event) {
	if player, exists := gs.Players[event.PlayerID]; exists {
		player.IsAlive = false
	}
}

func (gs *GameState) applyPhaseChanged(event Event) {
	newPhaseType, _ := event.Payload["phase_type"].(string)
	duration, _ := event.Payload["duration"].(float64)

	gs.Phase = Phase{
		Type:      PhaseType(newPhaseType),
		StartTime: event.Timestamp,
		Duration:  time.Duration(duration) * time.Second,
	}

	// Increment day number when transitioning to SITREP
	if PhaseType(newPhaseType) == PhaseSitrep {
		gs.DayNumber++
	}
}

func (gs *GameState) applyVoteCast(event Event) {
	playerID := event.PlayerID
	targetID, _ := event.Payload["target_id"].(string)
	voteType, _ := event.Payload["vote_type"].(string)

	// Initialize vote state if needed
	if gs.VoteState == nil {
		gs.VoteState = &VoteState{
			Type:         VoteType(voteType),
			Votes:        make(map[string]string),
			TokenWeights: make(map[string]int),
			Results:      make(map[string]int),
			IsComplete:   false,
		}
	}

	// Record the vote
	gs.VoteState.Votes[playerID] = targetID

	// Update token weights
	if player, exists := gs.Players[playerID]; exists {
		gs.VoteState.TokenWeights[playerID] = player.Tokens
	}

	// Recalculate results
	gs.VoteState.Results = make(map[string]int)
	for voterID, candidateID := range gs.VoteState.Votes {
		if tokens, exists := gs.VoteState.TokenWeights[voterID]; exists {
			gs.VoteState.Results[candidateID] += tokens
		}
	}
}

func (gs *GameState) applyTokensAwarded(event Event) {
	playerID := event.PlayerID
	amount, _ := event.Payload["amount"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		player.Tokens += int(amount)
	}
}

func (gs *GameState) applyMiningSuccessful(event Event) {
	playerID := event.PlayerID

	// Handle both int and float64 amount values
	var amount int
	if amountInt, ok := event.Payload["amount"].(int); ok {
		amount = amountInt
	} else if amountFloat, ok := event.Payload["amount"].(float64); ok {
		amount = int(amountFloat)
	} else {
		amount = 1 // Default amount
	}

	if player, exists := gs.Players[playerID]; exists {
		player.Tokens += amount
	}
}

func (gs *GameState) applyPlayerEliminated(event Event) {
	playerID := event.PlayerID
	roleType, _ := event.Payload["role_type"].(string)
	alignment, _ := event.Payload["alignment"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.IsAlive = false
		// Reveal role and alignment on elimination
		if player.Role == nil {
			player.Role = &Role{}
		}
		player.Role.Type = RoleType(roleType)
		player.Alignment = alignment
	}
}

func (gs *GameState) applyChatMessage(event Event) {
	message := ChatMessage{
		ID:         event.ID,
		PlayerID:   event.PlayerID,
		PlayerName: "",
		Message:    "",
		Timestamp:  event.Timestamp,
		IsSystem:   false,
	}

	if playerName, ok := event.Payload["player_name"].(string); ok {
		message.PlayerName = playerName
	}
	if messageText, ok := event.Payload["message"].(string); ok {
		message.Message = messageText
	}
	if isSystem, ok := event.Payload["is_system"].(bool); ok {
		message.IsSystem = isSystem
	}

	gs.ChatMessages = append(gs.ChatMessages, message)
}

func (gs *GameState) applyPlayerAligned(event Event) {
	playerID := event.PlayerID

	if player, exists := gs.Players[playerID]; exists {
		player.Alignment = "ALIGNED"
		// Reset any shock effects
		player.StatusMessage = ""
	}
}

func (gs *GameState) applyPlayerShocked(event Event) {
	playerID := event.PlayerID
	shockMessage, _ := event.Payload["shock_message"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.StatusMessage = shockMessage
		// System shock indicates failed conversion (proves humanity)
	}
}

func (gs *GameState) applyCrisisTriggered(event Event) {
	crisisType, _ := event.Payload["crisis_type"].(string)
	title, _ := event.Payload["title"].(string)
	description, _ := event.Payload["description"].(string)
	effects, _ := event.Payload["effects"].(map[string]interface{})

	gs.CrisisEvent = &CrisisEvent{
		Type:        crisisType,
		Title:       title,
		Description: description,
		Effects:     effects,
	}
}

func (gs *GameState) applyVictoryCondition(event Event) {
	winner, _ := event.Payload["winner"].(string)
	condition, _ := event.Payload["condition"].(string)
	description, _ := event.Payload["description"].(string)

	gs.WinCondition = &WinCondition{
		Winner:      winner,
		Condition:   condition,
		Description: description,
	}

	// End the game
	gs.Phase = Phase{
		Type:      PhaseGameOver,
		StartTime: event.Timestamp,
		Duration:  0,
	}
}

// Additional event handlers for complete game functionality

func (gs *GameState) applyGameEnded(event Event) {
	gs.Phase = Phase{
		Type:      PhaseGameOver,
		StartTime: event.Timestamp,
		Duration:  0,
	}
}

func (gs *GameState) applyDayStarted(event Event) {
	dayNumber, _ := event.Payload["day_number"].(float64)
	gs.DayNumber = int(dayNumber)

	gs.Phase = Phase{
		Type:      PhaseSitrep,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.SitrepDuration,
	}
}

func (gs *GameState) applyNightStarted(event Event) {
	gs.Phase = Phase{
		Type:      PhaseNight,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.NightDuration,
	}
}

func (gs *GameState) applyPlayerStatusChanged(event Event) {
	playerID := event.PlayerID
	newStatus, _ := event.Payload["status"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.StatusMessage = newStatus
	}
}

func (gs *GameState) applyPlayerReconnected(event Event) {
	// Player reconnection doesn't change game state directly
	// but could be used for analytics or notifications
}

func (gs *GameState) applyPlayerDisconnected(event Event) {
	// Player disconnection doesn't change game state directly
	// but could be used for analytics or notifications
}

func (gs *GameState) applyRoleAssigned(event Event) {
	playerID := event.PlayerID
	roleType, _ := event.Payload["role_type"].(string)
	roleName, _ := event.Payload["role_name"].(string)
	roleDescription, _ := event.Payload["role_description"].(string)
	kpiType, _ := event.Payload["kpi_type"].(string)
	kpiDescription, _ := event.Payload["kpi_description"].(string)
	alignment, _ := event.Payload["alignment"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.Role = &Role{
			Type:        RoleType(roleType),
			Name:        roleName,
			Description: roleDescription,
			IsUnlocked:  false,
		}

		if kpiType != "" {
			player.PersonalKPI = &PersonalKPI{
				Type:        KPIType(kpiType),
				Description: kpiDescription,
				Progress:    0,
				Target:      1, // Default target
				IsCompleted: false,
			}
		}

		player.Alignment = alignment
	}
}

func (gs *GameState) applyRoleAbilityUnlocked(event Event) {
	playerID := event.PlayerID
	abilityName, _ := event.Payload["ability_name"].(string)
	abilityDescription, _ := event.Payload["ability_description"].(string)

	if player, exists := gs.Players[playerID]; exists {
		if player.Role != nil {
			player.Role.IsUnlocked = true
			player.Role.Ability = &Ability{
				Name:        abilityName,
				Description: abilityDescription,
				IsReady:     true,
			}
		}
	}
}

func (gs *GameState) applyProjectMilestone(event Event) {
	playerID := event.PlayerID
	milestone, _ := event.Payload["milestone"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		player.ProjectMilestones = int(milestone)

		// Unlock role ability at 3 milestones
		if player.ProjectMilestones >= 3 && player.Role != nil && !player.Role.IsUnlocked {
			player.Role.IsUnlocked = true
			if player.Role.Ability != nil {
				player.Role.Ability.IsReady = true
			}
		}
	}
}

func (gs *GameState) applyVoteStarted(event Event) {
	voteType, _ := event.Payload["vote_type"].(string)

	gs.VoteState = &VoteState{
		Type:         VoteType(voteType),
		Votes:        make(map[string]string),
		TokenWeights: make(map[string]int),
		Results:      make(map[string]int),
		IsComplete:   false,
	}
}

func (gs *GameState) applyVoteCompleted(event Event) {
	if gs.VoteState != nil {
		gs.VoteState.IsComplete = true
	}
}

func (gs *GameState) applyPlayerNominated(event Event) {
	nominatedPlayerID, _ := event.Payload["nominated_player"].(string)
	gs.NominatedPlayer = nominatedPlayerID
}

func (gs *GameState) applyTokensLost(event Event) {
	playerID := event.PlayerID
	amount, _ := event.Payload["amount"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		player.Tokens -= int(amount)
		if player.Tokens < 0 {
			player.Tokens = 0
		}
	}
}

func (gs *GameState) applyMiningFailed(event Event) {
	playerID := event.PlayerID
	reason, _ := event.Payload["reason"].(string)

	if player, exists := gs.Players[playerID]; exists {
		if reason != "" {
			player.StatusMessage = "Mining failed: " + reason
		} else {
			player.StatusMessage = "Mining attempt failed"
		}
	}
}

func (gs *GameState) applyMiningPoolUpdated(event Event) {
	// Update mining pool difficulty or rewards
	newDifficulty, hasDifficulty := event.Payload["difficulty"].(float64)
	newBaseReward, hasReward := event.Payload["base_reward"].(float64)

	// Store mining pool state in crisis event effects for now
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{Effects: make(map[string]interface{})}
	}
	if gs.CrisisEvent.Effects == nil {
		gs.CrisisEvent.Effects = make(map[string]interface{})
	}

	if hasDifficulty {
		gs.CrisisEvent.Effects["mining_difficulty"] = newDifficulty
	}
	if hasReward {
		gs.CrisisEvent.Effects["mining_base_reward"] = newBaseReward
	}
}

func (gs *GameState) applyTokensDistributed(event Event) {
	// Handle bulk token distribution (e.g., from mining pool)
	distribution, ok := event.Payload["distribution"].(map[string]interface{})
	if !ok {
		return
	}

	for playerID, amountInterface := range distribution {
		if amount, ok := amountInterface.(float64); ok {
			if player, exists := gs.Players[playerID]; exists {
				player.Tokens += int(amount)
			}
		}
	}
}

func (gs *GameState) applyNightActionSubmitted(event Event) {
	playerID := event.PlayerID
	actionType, _ := event.Payload["action_type"].(string)
	targetID, _ := event.Payload["target_id"].(string)
	timestamp := event.Timestamp

	// Store the submitted night action
	if gs.NightActions == nil {
		gs.NightActions = make(map[string]*SubmittedNightAction)
	}

	gs.NightActions[playerID] = &SubmittedNightAction{
		PlayerID:  playerID,
		Type:      actionType,
		TargetID:  targetID,
		Payload:   event.Payload,
		Timestamp: timestamp,
	}

	// Update player's last action for reference
	if player, exists := gs.Players[playerID]; exists {
		player.LastNightAction = &NightAction{
			Type:     NightActionType(actionType),
			TargetID: targetID,
		}
	}
}

func (gs *GameState) applyNightActionsResolved(event Event) {
	// Process night action results
	results, ok := event.Payload["results"].(map[string]interface{})
	if !ok {
		return
	}

	// Update each player based on night action results
	for playerID, resultInterface := range results {
		if result, ok := resultInterface.(map[string]interface{}); ok {
			if player, exists := gs.Players[playerID]; exists {
				// Update tokens from mining or other actions
				if tokenChange, exists := result["token_change"]; exists {
					if change, ok := tokenChange.(float64); ok {
						player.Tokens += int(change)
						if player.Tokens < 0 {
							player.Tokens = 0
						}
					}
				}

				// Update status messages
				if status, exists := result["status_message"]; exists {
					if msg, ok := status.(string); ok {
						player.StatusMessage = msg
					}
				}

				// Update alignment changes from conversions
				if alignment, exists := result["alignment"]; exists {
					if align, ok := alignment.(string); ok {
						player.Alignment = align
					}
				}

				// Update AI equity
				if aiEquity, exists := result["ai_equity"]; exists {
					if equity, ok := aiEquity.(float64); ok {
						player.AIEquity = int(equity)
					}
				}

				// Reset night action tracking
				player.LastNightAction = nil
				player.HasUsedAbility = false
			}
		}
	}

	// Clear night action submissions
	gs.NightActions = make(map[string]*SubmittedNightAction)

	// Clear temporary night tracking
	gs.BlockedPlayersTonight = make(map[string]bool)
	gs.ProtectedPlayersTonight = make(map[string]bool)
}

func (gs *GameState) applyPlayerBlocked(event Event) {
	playerID := event.PlayerID
	blockedBy, _ := event.Payload["blocked_by"].(string)

	if player, exists := gs.Players[playerID]; exists {
		if blockedBy != "" {
			player.StatusMessage = "Action blocked by " + blockedBy
		} else {
			player.StatusMessage = "Action blocked"
		}
	}

	// Track blocked players for night resolution
	if gs.BlockedPlayersTonight == nil {
		gs.BlockedPlayersTonight = make(map[string]bool)
	}
	gs.BlockedPlayersTonight[playerID] = true
}

func (gs *GameState) applyPlayerProtected(event Event) {
	playerID := event.PlayerID
	protectedBy, _ := event.Payload["protected_by"].(string)

	if player, exists := gs.Players[playerID]; exists {
		if protectedBy != "" {
			player.StatusMessage = "Protected by " + protectedBy
		} else {
			player.StatusMessage = "Protected"
		}
	}

	// Track protected players for night resolution
	if gs.ProtectedPlayersTonight == nil {
		gs.ProtectedPlayersTonight = make(map[string]bool)
	}
	gs.ProtectedPlayersTonight[playerID] = true
}

func (gs *GameState) applyPlayerInvestigated(event Event) {
	// Investigation results are private to the investigator
	// Store the investigation for audit trails but don't modify visible state
	investigatorID := event.PlayerID
	_, _ = event.Payload["target_id"].(string)
	_, _ = event.Payload["result"].(string)

	// Investigations don't change public game state
	// Results are delivered privately to the investigator
	// We could store investigation history for admin/debug purposes
	if investigator, exists := gs.Players[investigatorID]; exists {
		// Mark ability as used
		investigator.HasUsedAbility = true
	}

	// The investigation result (alignment, role, etc.) is sent privately
	// and doesn't affect the global game state
}

func (gs *GameState) applyAIConversionAttempt(event Event) {
	targetID, _ := event.Payload["target_id"].(string)
	aiEquity, _ := event.Payload["ai_equity"].(float64)

	if player, exists := gs.Players[targetID]; exists {
		player.AIEquity = int(aiEquity)
	}
}

func (gs *GameState) applyAIConversionSuccess(event Event) {
	targetID := event.PlayerID

	if player, exists := gs.Players[targetID]; exists {
		player.Alignment = "ALIGNED"
		player.StatusMessage = "Conversion successful"
		player.AIEquity = 0 // Reset after successful conversion
	}
}

func (gs *GameState) applyAIConversionFailed(event Event) {
	targetID := event.PlayerID
	shockMessage, _ := event.Payload["shock_message"].(string)

	if player, exists := gs.Players[targetID]; exists {
		player.StatusMessage = shockMessage
		player.AIEquity = 0 // Reset after failed conversion
	}
}

func (gs *GameState) applySystemMessage(event Event) {
	message := ChatMessage{
		ID:         event.ID,
		PlayerID:   "SYSTEM",
		PlayerName: "Loebmate",
		Message:    "",
		Timestamp:  event.Timestamp,
		IsSystem:   true,
	}

	if messageText, ok := event.Payload["message"].(string); ok {
		message.Message = messageText
	}

	gs.ChatMessages = append(gs.ChatMessages, message)
}

func (gs *GameState) applyPrivateNotification(event Event) {
	// Private notifications don't affect global game state
	// They are delivered to specific players only
}

func (gs *GameState) applyPulseCheckStarted(event Event) {
	question, _ := event.Payload["question"].(string)

	// Store pulse check question in crisis event or separate field
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{
			Effects: make(map[string]interface{}),
		}
	}
	if gs.CrisisEvent.Effects == nil {
		gs.CrisisEvent.Effects = make(map[string]interface{})
	}
	gs.CrisisEvent.Effects["pulse_check_question"] = question
}

func (gs *GameState) applyPulseCheckSubmitted(event Event) {
	playerID := event.PlayerID
	response, _ := event.Payload["response"].(string)

	// Store pulse check responses (could be in a separate field)
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{Effects: make(map[string]interface{})}
	}
	if gs.CrisisEvent.Effects["pulse_responses"] == nil {
		gs.CrisisEvent.Effects["pulse_responses"] = make(map[string]interface{})
	}

	responses := gs.CrisisEvent.Effects["pulse_responses"].(map[string]interface{})
	responses[playerID] = response
}

func (gs *GameState) applyPulseCheckRevealed(event Event) {
	// Pulse check revelation triggers transition to discussion phase
	// The responses are already stored from submissions
}

// Role ability event handlers
func (gs *GameState) applyRunAudit(event Event) {
	// CISO audit ability - reveals alignment of target
	auditorID := event.PlayerID
	_, _ = event.Payload["target_id"].(string)
	_, _ = event.Payload["result"].(string)

	if auditor, exists := gs.Players[auditorID]; exists {
		auditor.HasUsedAbility = true
		auditor.StatusMessage = "Audit completed"
	}

	// Audit results are privately delivered to the CISO
	// Public game state doesn't change
}

func (gs *GameState) applyOverclockServers(event Event) {
	// CTO overclock ability - awards extra tokens to target
	ctoID := event.PlayerID
	targetID, _ := event.Payload["target_id"].(string)
	tokensAwarded, _ := event.Payload["tokens_awarded"].(float64)

	if cto, exists := gs.Players[ctoID]; exists {
		cto.HasUsedAbility = true
		cto.StatusMessage = "Servers overclocked"
	}

	if target, exists := gs.Players[targetID]; exists {
		target.Tokens += int(tokensAwarded)
		target.StatusMessage = "Received bonus tokens"
	}
}

func (gs *GameState) applyIsolateNode(event Event) {
	// COO isolate ability - blocks target's night action
	cooID := event.PlayerID
	targetID, _ := event.Payload["target_id"].(string)

	if coo, exists := gs.Players[cooID]; exists {
		coo.HasUsedAbility = true
		coo.StatusMessage = "Node isolated"
	}

	if target, exists := gs.Players[targetID]; exists {
		target.StatusMessage = "Connection isolated"
	}

	// Track blocked players for night resolution
	if gs.BlockedPlayersTonight == nil {
		gs.BlockedPlayersTonight = make(map[string]bool)
	}
	gs.BlockedPlayersTonight[targetID] = true
}

func (gs *GameState) applyPerformanceReview(event Event) {
	// CEO performance review - forces target to perform specific action
	ceoID := event.PlayerID
	targetID, _ := event.Payload["target_id"].(string)
	forcedAction, _ := event.Payload["forced_action"].(string)

	if ceo, exists := gs.Players[ceoID]; exists {
		ceo.HasUsedAbility = true
		ceo.StatusMessage = "Performance review completed"
	}

	if target, exists := gs.Players[targetID]; exists {
		target.StatusMessage = "Under performance review - " + forcedAction
	}

	// The forced action is handled by the night resolution system
}

func (gs *GameState) applyReallocateBudget(event Event) {
	// CFO budget reallocation - moves tokens between players
	cfoID := event.PlayerID
	fromPlayerID, _ := event.Payload["from_player"].(string)
	toPlayerID, _ := event.Payload["to_player"].(string)
	amount, _ := event.Payload["amount"].(float64)

	if cfo, exists := gs.Players[cfoID]; exists {
		cfo.HasUsedAbility = true
		cfo.StatusMessage = "Budget reallocated"
	}

	if fromPlayer, exists := gs.Players[fromPlayerID]; exists {
		fromPlayer.Tokens -= int(amount)
		if fromPlayer.Tokens < 0 {
			fromPlayer.Tokens = 0
		}
		fromPlayer.StatusMessage = "Budget reduced"
	}

	if toPlayer, exists := gs.Players[toPlayerID]; exists {
		toPlayer.Tokens += int(amount)
		toPlayer.StatusMessage = "Budget increased"
	}
}

func (gs *GameState) applyPivot(event Event) {
	// VP Platforms pivot - selects next day's crisis
	vpID := event.PlayerID
	selectedCrisis, _ := event.Payload["selected_crisis"].(string)

	if vp, exists := gs.Players[vpID]; exists {
		vp.HasUsedAbility = true
		vp.StatusMessage = "Strategy pivoted"
	}

	// Store the selected crisis for tomorrow's SITREP
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{Effects: make(map[string]interface{})}
	}
	if gs.CrisisEvent.Effects == nil {
		gs.CrisisEvent.Effects = make(map[string]interface{})
	}
	gs.CrisisEvent.Effects["next_crisis"] = selectedCrisis
}

func (gs *GameState) applyDeployHotfix(event Event) {
	// Ethics VP hotfix - redacts part of tomorrow's SITREP
	ethicsID := event.PlayerID
	redactionTarget, _ := event.Payload["redaction_target"].(string)

	if ethics, exists := gs.Players[ethicsID]; exists {
		ethics.HasUsedAbility = true
		ethics.StatusMessage = "Hotfix deployed"
	}

	// Store the redaction target for tomorrow's SITREP
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{Effects: make(map[string]interface{})}
	}
	if gs.CrisisEvent.Effects == nil {
		gs.CrisisEvent.Effects = make(map[string]interface{})
	}
	gs.CrisisEvent.Effects["sitrep_redaction"] = redactionTarget
}

// Status event handlers
func (gs *GameState) applySlackStatusChanged(event Event) {
	playerID := event.PlayerID
	status, _ := event.Payload["status"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.SlackStatus = status
	}
}

func (gs *GameState) applyPartingShotSet(event Event) {
	playerID := event.PlayerID
	partingShot, _ := event.Payload["parting_shot"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.PartingShot = partingShot
	}
}

// KPI event handlers
func (gs *GameState) applyKPIProgress(event Event) {
	playerID := event.PlayerID
	progress, _ := event.Payload["progress"].(float64)

	if player, exists := gs.Players[playerID]; exists && player.PersonalKPI != nil {
		player.PersonalKPI.Progress = int(progress)
	}
}

func (gs *GameState) applyKPICompleted(event Event) {
	playerID := event.PlayerID

	if player, exists := gs.Players[playerID]; exists && player.PersonalKPI != nil {
		player.PersonalKPI.IsCompleted = true
	}
}

// System shock event handlers
func (gs *GameState) applySystemShockApplied(event Event) {
	playerID := event.PlayerID
	shockType, _ := event.Payload["shock_type"].(string)
	description, _ := event.Payload["description"].(string)
	durationHours, _ := event.Payload["duration_hours"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		shock := SystemShock{
			Type:        ShockType(shockType),
			Description: description,
			ExpiresAt:   time.Now().Add(time.Duration(durationHours) * time.Hour),
			IsActive:    true,
		}

		if player.SystemShocks == nil {
			player.SystemShocks = make([]SystemShock, 0)
		}
		player.SystemShocks = append(player.SystemShocks, shock)
	}
}

// AI equity event handlers
func (gs *GameState) applyAIEquityChanged(event Event) {
	playerID := event.PlayerID
	change, _ := event.Payload["ai_equity_change"].(float64)
	newEquity, _ := event.Payload["new_ai_equity"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		if change != 0 {
			player.AIEquity += int(change)
		} else if newEquity != 0 {
			player.AIEquity = int(newEquity)
		}
	}
}

// Corporate mandate event handlers
func (gs *GameState) applyMandateActivated(event Event) {
	mandateType, _ := event.Payload["mandate_type"].(string)
	name, _ := event.Payload["name"].(string)
	description, _ := event.Payload["description"].(string)
	effects, _ := event.Payload["effects"].(map[string]interface{})

	gs.CorporateMandate = &CorporateMandate{
		Type:        MandateType(mandateType),
		Name:        name,
		Description: description,
		Effects:     effects,
		IsActive:    true,
	}
}

func (gs *GameState) applyMandateEffect(event Event) {
	if gs.CorporateMandate == nil {
		return
	}

	// Apply ongoing effects of the mandate
	effects, _ := event.Payload["effects"].(map[string]interface{})
	if effects != nil {
		for key, value := range effects {
			gs.CorporateMandate.Effects[key] = value
		}
	}
}

// System shock effect event handlers
func (gs *GameState) applyShockEffectTriggered(event Event) {
	playerID := event.PlayerID
	effectType, _ := event.Payload["effect_type"].(string)
	description, _ := event.Payload["description"].(string)

	if player, exists := gs.Players[playerID]; exists {
		// Apply specific shock effects
		switch effectType {
		case "message_corruption":
			player.StatusMessage = "Message corruption active"
		case "action_lock":
			player.StatusMessage = "Action lock in effect"
		case "forced_silence":
			player.StatusMessage = "Communication restricted"
		default:
			player.StatusMessage = description
		}
	}
}

// Equity threshold event handlers
func (gs *GameState) applyEquityThreshold(event Event) {
	playerID := event.PlayerID
	threshold, _ := event.Payload["threshold"].(int)
	action, _ := event.Payload["action"].(string)

	if player, exists := gs.Players[playerID]; exists {
		// Track equity threshold events for AI conversion logic
		player.StatusMessage = fmt.Sprintf("AI Equity threshold %d reached - %s", threshold, action)
	}
}