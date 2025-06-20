package game

import (
	"time"
)

// Event represents a game event that changes state
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	GameID    string                 `json:"game_id"`
	PlayerID  string                 `json:"player_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// EventType represents different types of game events
type EventType string

const (
	// Game lifecycle events
	EventGameCreated  EventType = "GAME_CREATED"
	EventGameStarted  EventType = "GAME_STARTED"
	EventGameEnded    EventType = "GAME_ENDED"
	EventPhaseChanged EventType = "PHASE_CHANGED"

	// Player events
	EventPlayerJoined       EventType = "PLAYER_JOINED"
	EventPlayerLeft         EventType = "PLAYER_LEFT"
	EventPlayerEliminated   EventType = "PLAYER_ELIMINATED"
	EventPlayerRoleRevealed EventType = "PLAYER_ROLE_REVEALED"
	EventPlayerAligned      EventType = "PLAYER_ALIGNED"
	EventPlayerShocked      EventType = "PLAYER_SHOCKED"

	// Voting events
	EventVoteStarted      EventType = "VOTE_STARTED"
	EventVoteCast         EventType = "VOTE_CAST"
	EventVoteTallyUpdated EventType = "VOTE_TALLY_UPDATED"
	EventVoteCompleted    EventType = "VOTE_COMPLETED"
	EventPlayerNominated  EventType = "PLAYER_NOMINATED"

	// Token and Mining events
	EventTokensAwarded    EventType = "TOKENS_AWARDED"
	EventTokensSpent      EventType = "TOKENS_SPENT"
	EventMiningAttempted  EventType = "MINING_ATTEMPTED"
	EventMiningSuccessful EventType = "MINING_SUCCESSFUL"
	EventMiningFailed     EventType = "MINING_FAILED"

	// Night Action events
	EventNightActionsResolved EventType = "NIGHT_ACTIONS_RESOLVED"
	EventPlayerBlocked        EventType = "PLAYER_BLOCKED"
	EventPlayerProtected      EventType = "PLAYER_PROTECTED"
	EventPlayerInvestigated   EventType = "PLAYER_INVESTIGATED"

	// AI and Conversion events
	EventAIConversionAttempt EventType = "AI_CONVERSION_ATTEMPT"
	EventAIConversionSuccess EventType = "AI_CONVERSION_SUCCESS"
	EventAIConversionFailed  EventType = "AI_CONVERSION_FAILED"
	EventAIRevealed          EventType = "AI_REVEALED"

	// Communication events
	EventChatMessage         EventType = "CHAT_MESSAGE"
	EventSystemMessage       EventType = "SYSTEM_MESSAGE"
	EventPrivateNotification EventType = "PRIVATE_NOTIFICATION"

	// Crisis and Special events
	EventCrisisTriggered     EventType = "CRISIS_TRIGGERED"
	EventPulseCheckStarted   EventType = "PULSE_CHECK_STARTED"
	EventPulseCheckSubmitted EventType = "PULSE_CHECK_SUBMITTED"
	EventPulseCheckRevealed  EventType = "PULSE_CHECK_REVEALED"
	EventRoleAbilityUnlocked EventType = "ROLE_ABILITY_UNLOCKED"
	EventProjectMilestone    EventType = "PROJECT_MILESTONE"
	EventRoleAssigned        EventType = "ROLE_ASSIGNED"

	// Mining and Economy events
	EventMiningPoolUpdated EventType = "MINING_POOL_UPDATED"
	EventTokensDistributed EventType = "TOKENS_DISTRIBUTED"
	EventTokensLost        EventType = "TOKENS_LOST"

	// Day/Night transition events
	EventDayStarted           EventType = "DAY_STARTED"
	EventNightStarted         EventType = "NIGHT_STARTED"
	EventNightActionSubmitted EventType = "NIGHT_ACTION_SUBMITTED"
	EventAllPlayersReady      EventType = "ALL_PLAYERS_READY"

	// Status and State events
	EventPlayerStatusChanged EventType = "PLAYER_STATUS_CHANGED"
	EventGameStateSnapshot   EventType = "GAME_STATE_SNAPSHOT"
	EventPlayerReconnected   EventType = "PLAYER_RECONNECTED"
	EventPlayerDisconnected  EventType = "PLAYER_DISCONNECTED"

	// Win Condition events
	EventVictoryCondition EventType = "VICTORY_CONDITION"

	// Role Ability events
	EventRunAudit          EventType = "RUN_AUDIT"
	EventOverclockServers  EventType = "OVERCLOCK_SERVERS"
	EventIsolateNode       EventType = "ISOLATE_NODE"
	EventPerformanceReview EventType = "PERFORMANCE_REVIEW"
	EventReallocateBudget  EventType = "REALLOCATE_BUDGET"
	EventPivot             EventType = "PIVOT"
	EventDeployHotfix      EventType = "DEPLOY_HOTFIX"

	// Slack Status events
	EventSlackStatusChanged EventType = "SLACK_STATUS_CHANGED"
	EventPartingShotSet     EventType = "PARTING_SHOT_SET"

	// Personal KPI events
	EventKPIProgress  EventType = "KPI_PROGRESS"
	EventKPICompleted EventType = "KPI_COMPLETED"

	// Corporate Mandate events
	EventMandateActivated EventType = "MANDATE_ACTIVATED"
	EventMandateEffect    EventType = "MANDATE_EFFECT"

	// System Shock events
	EventSystemShockApplied   EventType = "SYSTEM_SHOCK_APPLIED"
	EventShockEffectTriggered EventType = "SHOCK_EFFECT_TRIGGERED"

	// AI Equity events
	EventAIEquityChanged EventType = "AI_EQUITY_CHANGED"
	EventEquityThreshold EventType = "EQUITY_THRESHOLD"
)

// Action represents a player action that can generate events
type Action struct {
	Type      ActionType             `json:"type"`
	PlayerID  string                 `json:"player_id"`
	GameID    string                 `json:"game_id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// ActionType represents different types of player actions
type ActionType string

const (
	// Lobby actions
	ActionJoinGame  ActionType = "JOIN_GAME"
	ActionLeaveGame ActionType = "LEAVE_GAME"
	ActionStartGame ActionType = "START_GAME"

	// Communication actions
	ActionSendMessage      ActionType = "SEND_MESSAGE"
	ActionSubmitPulseCheck ActionType = "SUBMIT_PULSE_CHECK"

	// Voting actions
	ActionSubmitVote       ActionType = "SUBMIT_VOTE"
	ActionExtendDiscussion ActionType = "EXTEND_DISCUSSION"

	// Night actions
	ActionSubmitNightAction ActionType = "SUBMIT_NIGHT_ACTION"
	ActionMineTokens        ActionType = "MINE_TOKENS"
	ActionUseAbility        ActionType = "USE_ABILITY"
	ActionAttemptConversion ActionType = "ATTEMPT_CONVERSION"
	ActionProjectMilestones ActionType = "PROJECT_MILESTONES"

	// Role-specific abilities
	ActionRunAudit          ActionType = "RUN_AUDIT"
	ActionOverclockServers  ActionType = "OVERCLOCK_SERVERS"
	ActionIsolateNode       ActionType = "ISOLATE_NODE"
	ActionPerformanceReview ActionType = "PERFORMANCE_REVIEW"
	ActionReallocateBudget  ActionType = "REALLOCATE_BUDGET"
	ActionPivot             ActionType = "PIVOT"
	ActionDeployHotfix      ActionType = "DEPLOY_HOTFIX"

	// Status actions
	ActionSetSlackStatus ActionType = "SET_SLACK_STATUS"

	// Meta actions
	ActionReconnect ActionType = "RECONNECT"
)

// ApplyEvent applies an event to the game state
func (gs *GameState) ApplyEvent(event Event) error {
	gs.UpdatedAt = event.Timestamp

	switch event.Type {
	// Game lifecycle events
	case EventGameStarted:
		return gs.applyGameStarted(event)
	case EventGameEnded:
		return gs.applyGameEnded(event)
	case EventPhaseChanged:
		return gs.applyPhaseChanged(event)
	case EventDayStarted:
		return gs.applyDayStarted(event)
	case EventNightStarted:
		return gs.applyNightStarted(event)

	// Player events
	case EventPlayerJoined:
		return gs.applyPlayerJoined(event)
	case EventPlayerLeft:
		return gs.applyPlayerLeft(event)
	case EventPlayerEliminated:
		return gs.applyPlayerEliminated(event)
	case EventPlayerAligned:
		return gs.applyPlayerAligned(event)
	case EventPlayerShocked:
		return gs.applyPlayerShocked(event)
	case EventPlayerStatusChanged:
		return gs.applyPlayerStatusChanged(event)
	case EventPlayerReconnected:
		return gs.applyPlayerReconnected(event)
	case EventPlayerDisconnected:
		return gs.applyPlayerDisconnected(event)

	// Role and ability events
	case EventRoleAssigned:
		return gs.applyRoleAssigned(event)
	case EventRoleAbilityUnlocked:
		return gs.applyRoleAbilityUnlocked(event)
	case EventProjectMilestone:
		return gs.applyProjectMilestone(event)

	// Voting events
	case EventVoteCast:
		return gs.applyVoteCast(event)
	case EventVoteStarted:
		return gs.applyVoteStarted(event)
	case EventVoteCompleted:
		return gs.applyVoteCompleted(event)
	case EventPlayerNominated:
		return gs.applyPlayerNominated(event)

	// Token and mining events
	case EventTokensAwarded:
		return gs.applyTokensAwarded(event)
	case EventTokensLost:
		return gs.applyTokensLost(event)
	case EventMiningSuccessful:
		return gs.applyMiningSuccessful(event)
	case EventMiningFailed:
		return gs.applyMiningFailed(event)
	case EventMiningPoolUpdated:
		return gs.applyMiningPoolUpdated(event)
	case EventTokensDistributed:
		return gs.applyTokensDistributed(event)

	// Night action events
	case EventNightActionSubmitted:
		return gs.applyNightActionSubmitted(event)
	case EventNightActionsResolved:
		return gs.applyNightActionsResolved(event)
	case EventPlayerBlocked:
		return gs.applyPlayerBlocked(event)
	case EventPlayerProtected:
		return gs.applyPlayerProtected(event)
	case EventPlayerInvestigated:
		return gs.applyPlayerInvestigated(event)

	// AI and conversion events
	case EventAIConversionAttempt:
		return gs.applyAIConversionAttempt(event)
	case EventAIConversionSuccess:
		return gs.applyAIConversionSuccess(event)
	case EventAIConversionFailed:
		return gs.applyAIConversionFailed(event)

	// Communication events
	case EventChatMessage:
		return gs.applyChatMessage(event)
	case EventSystemMessage:
		return gs.applySystemMessage(event)
	case EventPrivateNotification:
		return gs.applyPrivateNotification(event)

	// Crisis and pulse check events
	case EventCrisisTriggered:
		return gs.applyCrisisTriggered(event)
	case EventPulseCheckStarted:
		return gs.applyPulseCheckStarted(event)
	case EventPulseCheckSubmitted:
		return gs.applyPulseCheckSubmitted(event)
	case EventPulseCheckRevealed:
		return gs.applyPulseCheckRevealed(event)

	// Win condition events
	case EventVictoryCondition:
		return gs.applyVictoryCondition(event)

	// Role ability events
	case EventRunAudit:
		return gs.applyRunAudit(event)
	case EventOverclockServers:
		return gs.applyOverclockServers(event)
	case EventIsolateNode:
		return gs.applyIsolateNode(event)
	case EventPerformanceReview:
		return gs.applyPerformanceReview(event)
	case EventReallocateBudget:
		return gs.applyReallocateBudget(event)
	case EventPivot:
		return gs.applyPivot(event)
	case EventDeployHotfix:
		return gs.applyDeployHotfix(event)

	// Status events
	case EventSlackStatusChanged:
		return gs.applySlackStatusChanged(event)
	case EventPartingShotSet:
		return gs.applyPartingShotSet(event)

	// KPI events
	case EventKPIProgress:
		return gs.applyKPIProgress(event)
	case EventKPICompleted:
		return gs.applyKPICompleted(event)

	// System shock events
	case EventSystemShockApplied:
		return gs.applySystemShockApplied(event)

	// AI equity events
	case EventAIEquityChanged:
		return gs.applyAIEquityChanged(event)

	default:
		// Unknown event type - log but don't error
		return nil
	}
}

func (gs *GameState) applyGameStarted(event Event) error {
	gs.Phase = Phase{
		Type:      PhaseSitrep,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.SitrepDuration,
	}
	gs.DayNumber = 1
	return nil
}

func (gs *GameState) applyPlayerJoined(event Event) error {
	playerID := event.PlayerID
	name, _ := event.Payload["name"].(string)
	jobTitle, _ := event.Payload["job_title"].(string)

	gs.Players[playerID] = &Player{
		ID:                playerID,
		Name:              name,
		JobTitle:          jobTitle,
		IsAlive:           true,
		Tokens:            gs.Settings.StartingTokens,
		ProjectMilestones: 0,
		StatusMessage:     "",
		JoinedAt:          event.Timestamp,
		Alignment:         "HUMAN", // Default alignment
	}
	return nil
}

func (gs *GameState) applyPlayerLeft(event Event) error {
	if player, exists := gs.Players[event.PlayerID]; exists {
		player.IsAlive = false
	}
	return nil
}

func (gs *GameState) applyPhaseChanged(event Event) error {
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

	return nil
}

func (gs *GameState) applyVoteCast(event Event) error {
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

	return nil
}

func (gs *GameState) applyTokensAwarded(event Event) error {
	playerID := event.PlayerID
	amount, _ := event.Payload["amount"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		player.Tokens += int(amount)
	}
	return nil
}

func (gs *GameState) applyMiningSuccessful(event Event) error {
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
	return nil
}

func (gs *GameState) applyPlayerEliminated(event Event) error {
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
	return nil
}

func (gs *GameState) applyChatMessage(event Event) error {
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
	return nil
}

func (gs *GameState) applyPlayerAligned(event Event) error {
	playerID := event.PlayerID

	if player, exists := gs.Players[playerID]; exists {
		player.Alignment = "ALIGNED"
		// Reset any shock effects
		player.StatusMessage = ""
	}
	return nil
}

func (gs *GameState) applyPlayerShocked(event Event) error {
	playerID := event.PlayerID
	shockMessage, _ := event.Payload["shock_message"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.StatusMessage = shockMessage
		// System shock indicates failed conversion (proves humanity)
	}
	return nil
}

func (gs *GameState) applyCrisisTriggered(event Event) error {
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
	return nil
}

func (gs *GameState) applyVictoryCondition(event Event) error {
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

	return nil
}

// Additional event handlers for complete game functionality

func (gs *GameState) applyGameEnded(event Event) error {
	gs.Phase = Phase{
		Type:      PhaseGameOver,
		StartTime: event.Timestamp,
		Duration:  0,
	}
	return nil
}

func (gs *GameState) applyDayStarted(event Event) error {
	dayNumber, _ := event.Payload["day_number"].(float64)
	gs.DayNumber = int(dayNumber)

	gs.Phase = Phase{
		Type:      PhaseSitrep,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.SitrepDuration,
	}
	return nil
}

func (gs *GameState) applyNightStarted(event Event) error {
	gs.Phase = Phase{
		Type:      PhaseNight,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.NightDuration,
	}
	return nil
}

func (gs *GameState) applyPlayerStatusChanged(event Event) error {
	playerID := event.PlayerID
	newStatus, _ := event.Payload["status"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.StatusMessage = newStatus
	}
	return nil
}

func (gs *GameState) applyPlayerReconnected(event Event) error {
	// Player reconnection doesn't change game state directly
	// but could be used for analytics or notifications
	return nil
}

func (gs *GameState) applyPlayerDisconnected(event Event) error {
	// Player disconnection doesn't change game state directly
	// but could be used for analytics or notifications
	return nil
}

func (gs *GameState) applyRoleAssigned(event Event) error {
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
	return nil
}

func (gs *GameState) applyRoleAbilityUnlocked(event Event) error {
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
	return nil
}

func (gs *GameState) applyProjectMilestone(event Event) error {
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
	return nil
}

func (gs *GameState) applyVoteStarted(event Event) error {
	voteType, _ := event.Payload["vote_type"].(string)

	gs.VoteState = &VoteState{
		Type:         VoteType(voteType),
		Votes:        make(map[string]string),
		TokenWeights: make(map[string]int),
		Results:      make(map[string]int),
		IsComplete:   false,
	}
	return nil
}

func (gs *GameState) applyVoteCompleted(event Event) error {
	if gs.VoteState != nil {
		gs.VoteState.IsComplete = true
	}
	return nil
}

func (gs *GameState) applyPlayerNominated(event Event) error {
	nominatedPlayerID, _ := event.Payload["nominated_player"].(string)
	gs.NominatedPlayer = nominatedPlayerID
	return nil
}

func (gs *GameState) applyTokensLost(event Event) error {
	playerID := event.PlayerID
	amount, _ := event.Payload["amount"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		player.Tokens -= int(amount)
		if player.Tokens < 0 {
			player.Tokens = 0
		}
	}
	return nil
}

func (gs *GameState) applyMiningFailed(event Event) error {
	// Mining failure doesn't change state but can be tracked for analytics
	return nil
}

func (gs *GameState) applyMiningPoolUpdated(event Event) error {
	// Mining pool updates would be tracked in game settings or separate state
	// For now, just acknowledge the event
	return nil
}

func (gs *GameState) applyTokensDistributed(event Event) error {
	// Handle bulk token distribution (e.g., from mining pool)
	distribution, ok := event.Payload["distribution"].(map[string]interface{})
	if !ok {
		return nil
	}

	for playerID, amountInterface := range distribution {
		if amount, ok := amountInterface.(float64); ok {
			if player, exists := gs.Players[playerID]; exists {
				player.Tokens += int(amount)
			}
		}
	}
	return nil
}

func (gs *GameState) applyNightActionSubmitted(event Event) error {
	playerID := event.PlayerID
	actionType, _ := event.Payload["action_type"].(string)
	targetID, _ := event.Payload["target_id"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.LastNightAction = &NightAction{
			Type:     NightActionType(actionType),
			TargetID: targetID,
		}
	}
	return nil
}

func (gs *GameState) applyNightActionsResolved(event Event) error {
	// Night actions resolution would update various player states
	// This is a complex event that would be handled by the rules engine
	results, ok := event.Payload["results"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Process results for each player
	for playerID, resultInterface := range results {
		if result, ok := resultInterface.(map[string]interface{}); ok {
			if player, exists := gs.Players[playerID]; exists {
				// Update player based on night action results
				if newTokens, exists := result["tokens"]; exists {
					if tokens, ok := newTokens.(float64); ok {
						player.Tokens = int(tokens)
					}
				}
				if newStatus, exists := result["status"]; exists {
					if status, ok := newStatus.(string); ok {
						player.StatusMessage = status
					}
				}
			}
		}
	}
	return nil
}

func (gs *GameState) applyPlayerBlocked(event Event) error {
	playerID := event.PlayerID

	if player, exists := gs.Players[playerID]; exists {
		player.StatusMessage = "Action blocked"
	}
	return nil
}

func (gs *GameState) applyPlayerProtected(event Event) error {
	playerID := event.PlayerID

	if player, exists := gs.Players[playerID]; exists {
		player.StatusMessage = "Protected"
	}
	return nil
}

func (gs *GameState) applyPlayerInvestigated(event Event) error {
	// Investigation results are usually private to the investigator
	// The game state doesn't change, but results are delivered privately
	return nil
}

func (gs *GameState) applyAIConversionAttempt(event Event) error {
	targetID, _ := event.Payload["target_id"].(string)
	aiEquity, _ := event.Payload["ai_equity"].(float64)

	if player, exists := gs.Players[targetID]; exists {
		player.AIEquity = int(aiEquity)
	}
	return nil
}

func (gs *GameState) applyAIConversionSuccess(event Event) error {
	targetID := event.PlayerID

	if player, exists := gs.Players[targetID]; exists {
		player.Alignment = "ALIGNED"
		player.StatusMessage = "Conversion successful"
		player.AIEquity = 0 // Reset after successful conversion
	}
	return nil
}

func (gs *GameState) applyAIConversionFailed(event Event) error {
	targetID := event.PlayerID
	shockMessage, _ := event.Payload["shock_message"].(string)

	if player, exists := gs.Players[targetID]; exists {
		player.StatusMessage = shockMessage
		player.AIEquity = 0 // Reset after failed conversion
	}
	return nil
}

func (gs *GameState) applySystemMessage(event Event) error {
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
	return nil
}

func (gs *GameState) applyPrivateNotification(event Event) error {
	// Private notifications don't affect global game state
	// They are delivered to specific players only
	return nil
}

func (gs *GameState) applyPulseCheckStarted(event Event) error {
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
	return nil
}

func (gs *GameState) applyPulseCheckSubmitted(event Event) error {
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
	return nil
}

func (gs *GameState) applyPulseCheckRevealed(event Event) error {
	// Pulse check revelation triggers transition to discussion phase
	// The responses are already stored from submissions
	return nil
}

// Role ability event handlers
func (gs *GameState) applyRunAudit(event Event) error {
	// Audit results don't change game state directly, but may be tracked
	return nil
}

func (gs *GameState) applyOverclockServers(event Event) error {
	// Token awards are handled by the role ability manager
	return nil
}

func (gs *GameState) applyIsolateNode(event Event) error {
	// Blocking is handled by the role ability manager
	return nil
}

func (gs *GameState) applyPerformanceReview(event Event) error {
	// Forced actions are handled by the role ability manager
	return nil
}

func (gs *GameState) applyReallocateBudget(event Event) error {
	// Token transfers are handled by the role ability manager
	return nil
}

func (gs *GameState) applyPivot(event Event) error {
	// Crisis selection is handled by the role ability manager
	return nil
}

func (gs *GameState) applyDeployHotfix(event Event) error {
	// SITREP redaction is handled during SITREP generation
	return nil
}

// Status event handlers
func (gs *GameState) applySlackStatusChanged(event Event) error {
	playerID := event.PlayerID
	status, _ := event.Payload["status"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.SlackStatus = status
	}
	return nil
}

func (gs *GameState) applyPartingShotSet(event Event) error {
	playerID := event.PlayerID
	partingShot, _ := event.Payload["parting_shot"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.PartingShot = partingShot
	}
	return nil
}

// KPI event handlers
func (gs *GameState) applyKPIProgress(event Event) error {
	playerID := event.PlayerID
	progress, _ := event.Payload["progress"].(float64)

	if player, exists := gs.Players[playerID]; exists && player.PersonalKPI != nil {
		player.PersonalKPI.Progress = int(progress)
	}
	return nil
}

func (gs *GameState) applyKPICompleted(event Event) error {
	playerID := event.PlayerID

	if player, exists := gs.Players[playerID]; exists && player.PersonalKPI != nil {
		player.PersonalKPI.IsCompleted = true
	}
	return nil
}

// System shock event handlers
func (gs *GameState) applySystemShockApplied(event Event) error {
	playerID := event.PlayerID
	shockType, _ := event.Payload["shock_type"].(string)
	description, _ := event.Payload["description"].(string)
	durationHours, _ := event.Payload["duration_hours"].(float64)

	if player, exists := gs.Players[playerID]; exists {
		shock := SystemShock{
			Type:        ShockType(shockType),
			Description: description,
			ExpiresAt:   getCurrentTime().Add(time.Duration(durationHours) * time.Hour),
			IsActive:    true,
		}

		if player.SystemShocks == nil {
			player.SystemShocks = make([]SystemShock, 0)
		}
		player.SystemShocks = append(player.SystemShocks, shock)
	}
	return nil
}

// AI equity event handlers
func (gs *GameState) applyAIEquityChanged(event Event) error {
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
	return nil
}
