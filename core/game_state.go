
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// GameState represents the complete state of a game
type GameState struct {
	Version             int                              `json:"version"`         // For handling state migrations
	Checksum            string                           `json:"checksum,omitempty"` // For verifying data integrity
	ID                  string                           `json:"id"`
	Phase               Phase                            `json:"phase"`
	DayNumber           int                              `json:"day_number"`
	Players             map[string]*Player               `json:"players"`
	CreatedAt           time.Time                        `json:"created_at"`
	UpdatedAt           time.Time                        `json:"updated_at"`
	Settings            GameSettings                     `json:"settings"`
	CrisisEvent         *CrisisEvent                     `json:"crisis_event,omitempty"`
	ChatMessages        []ChatMessage                    `json:"chat_messages"`
	VoteState           *VoteState                       `json:"vote_state,omitempty"`
	WhistleblowerVoting *WhistleblowerVoting             `json:"whistleblower_voting,omitempty"`
	NominatedPlayer     string                           `json:"nominated_player,omitempty"`
	WinCondition        *WinCondition                    `json:"win_condition,omitempty"`
	NightActions        map[string]*SubmittedNightAction `json:"night_actions,omitempty"`

	// Game-wide modifiers
	CorporateMandate *CorporateMandate `json:"corporate_mandate,omitempty"`

	// Daily tracking
	PulseCheckResponses map[string]string `json:"pulse_check_responses,omitempty"`

	// Phase skipping
	SkipVotes map[string]bool `json:"skip_votes,omitempty"` // PlayerID -> bool

	// Temporary fields for night resolution (cleared each night)
	BlockedPlayersTonight   map[string]bool `json:"-"` // Not serialized
	ProtectedPlayersTonight map[string]bool `json:"-"` // Not serialized
}

// NewGameState creates a new game state
func NewGameState(id string, currentTime time.Time) *GameState {
	return &GameState{
		Version:      1,
		ID:           id,
		Phase:        Phase{Type: PhaseLobby, StartTime: currentTime, Duration: 0},
		DayNumber:    0,
		Players:      make(map[string]*Player),
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
		ChatMessages: make([]ChatMessage, 0),
		NightActions: make(map[string]*SubmittedNightAction),
		Settings: GameSettings{
			MaxPlayers:               10,
			MinPlayers:               2,
			SitrepDuration:           15 * time.Second,
			PulseCheckDuration:       30 * time.Second,
			DiscussionDuration:       2 * time.Minute,
			ExtensionDuration:        15 * time.Second,
			NominationDuration:       30 * time.Second,
			TrialDuration:            30 * time.Second,
			VerdictDuration:          30 * time.Second,
			NightDuration:            30 * time.Second,
			StartingTokens:           1,
			VotingThreshold:          0.5,
			InitialAlignedHumanCount: 0,
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
	case EventPlayerAbandoned:
		newState.applyPlayerAbandoned(event)
	case EventPlayerConnectionStatusChanged:
		newState.applyPlayerConnectionStatusChanged(event)
	case EventPlayerRoleRevealed:
		newState.applyPlayerRoleRevealed(event)
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
	case EventVoteTallyUpdated:
		newState.applyVoteTallyUpdated(event)
	case EventVoteStarted:
		newState.applyVoteStarted(event)
	case EventVoteCompleted:
		newState.applyVoteCompleted(event)
	case EventPlayerNominated:
		newState.applyPlayerNominated(event)
	case EventExtensionVotingTriggered:
		newState.applyExtensionVotingTriggered(event)

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
	case EventMessageReaction:
		newState.applyMessageReaction(event)
	case EventSystemMessage:
		newState.applySystemMessage(event) // DEPRECATED: Use specific semantic events
	case EventPrivateNotification:
		newState.applyPrivateNotification(event)

	// Specific semantic events
	case EventClientError:
		newState.applyClientError(event)
	case EventSitrepPublished:
		newState.applySitrepPublished(event)
	case EventLiaisonProtocolActivated:
		newState.applyLiaisonProtocolActivated(event)
	case EventLiaisonIntelRevealed:
		newState.applyLiaisonIntelRevealed(event)
	case EventAIConversionBlocked:
		newState.applyAIConversionBlocked(event)
	case EventGameRuleModified:
		newState.applyGameRuleModified(event)

	// Crisis and pulse check events
	case EventCrisisTriggered:
		newState.applyCrisisTriggered(event)
	case EventPulseCheckStarted:
		newState.applyPulseCheckStarted(event)
	case EventPulseCheckUpdated:
		newState.applyPulseCheckUpdated(event)
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
	case EventKPIAssigned:
		newState.applyKPIAssigned(event)
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

	// Phase skipping events
	case EventSkipVoteUpdated:
		newState.applySkipVoteUpdated(event)

	// Whistleblower Protocol events
	case EventWhistleblowerVotingStarted:
		newState.applyWhistleblowerVotingStarted(event)
	case EventWhistleblowerVoteCast:
		newState.applyWhistleblowerVoteCast(event)
	case EventWhistleblowerVotingCompleted:
		newState.applyWhistleblowerVotingCompleted(event)

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
		Status:            PlayerStatusAlive,
		IsAlive:           true,
		ConnectionStatus:  "CONNECTED", // Default connection status
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

	// Clear skip votes at the start of each new phase
	gs.SkipVotes = make(map[string]bool)

	// Increment day number when transitioning to SITREP
	if PhaseType(newPhaseType) == PhaseSitrep {
		gs.DayNumber++

		// Reset pulse check submission flags for all players at the start of each new day
		for _, player := range gs.Players {
			player.HasSubmittedPulseCheck = false
		}
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

func (gs *GameState) applyVoteTallyUpdated(event Event) {
	voteType, _ := event.Payload["vote_type"].(string)
	results, _ := event.Payload["results"].(map[string]interface{})
	tokenWeights, _ := event.Payload["token_weights"].(map[string]interface{})
	isComplete, _ := event.Payload["is_complete"].(bool)
	voterID, _ := event.Payload["voter_id"].(string)
	targetID, _ := event.Payload["target_id"].(string)

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

	// Apply the authoritative vote tally from the event
	if results != nil {
		gs.VoteState.Results = make(map[string]int)
		for candidateID, voteCountInterface := range results {
			if voteCount, ok := voteCountInterface.(float64); ok {
				gs.VoteState.Results[candidateID] = int(voteCount)
			} else if voteCount, ok := voteCountInterface.(int); ok {
				gs.VoteState.Results[candidateID] = voteCount
			}
		}
	}

	// Apply token weights from the event
	if tokenWeights != nil {
		gs.VoteState.TokenWeights = make(map[string]int)
		for playerID, weightInterface := range tokenWeights {
			if weight, ok := weightInterface.(float64); ok {
				gs.VoteState.TokenWeights[playerID] = int(weight)
			} else if weight, ok := weightInterface.(int); ok {
				gs.VoteState.TokenWeights[playerID] = weight
			}
		}
	}

	// Record the individual vote that triggered this update
	if voterID != "" && targetID != "" {
		gs.VoteState.Votes[voterID] = targetID
	}

	// Apply public voting information if transparency mandate is active
	if publicVoting, exists := event.Payload["public_voting"].(bool); exists && publicVoting {
		if voterChoices, exists := event.Payload["voter_choices"].(map[string]interface{}); exists {
			// Replace with authoritative voter choices
			gs.VoteState.Votes = make(map[string]string)
			for voterID, choiceInterface := range voterChoices {
				if choice, ok := choiceInterface.(string); ok {
					gs.VoteState.Votes[voterID] = choice
				}
			}
		}
	}

	// Update completion status
	gs.VoteState.IsComplete = isComplete
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
		player.Status = PlayerStatusEliminated
		player.IsAlive = false
		// Reveal role and alignment on elimination
		if player.Role == nil {
			player.Role = &Role{}
		}
		player.Role.Type = RoleType(roleType)
		player.Alignment = alignment
	}
}

func (gs *GameState) applyPlayerAbandoned(event Event) {
	playerID := event.PlayerID
	revealedRole, _ := event.Payload["revealed_role"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.Status = PlayerStatusAbandoned
		player.IsAlive = false
		player.IsRolePubliclyRevealed = true
		player.StatusMessage = "ABANDONED"

		// Reveal role but NOT alignment (as per requirements)
		if player.Role == nil {
			player.Role = &Role{}
		}
		player.Role.Type = RoleType(revealedRole)

		// Clear tokens as they are forfeited
		player.Tokens = 0
	}
}

func (gs *GameState) applyPlayerConnectionStatusChanged(event Event) {
	playerID, _ := event.Payload["player_id"].(string)
	connectionStatus, _ := event.Payload["connection_status"].(string)
	
	if player, exists := gs.Players[playerID]; exists {
		player.ConnectionStatus = connectionStatus
		
		// Update status message to reflect connection state
		if connectionStatus == "DISCONNECTED" {
			player.StatusMessage = "DISCONNECTED"
		} else if connectionStatus == "CONNECTED" {
			player.StatusMessage = "" // Clear disconnected status
		}
	}
}

func (gs *GameState) applyPlayerRoleRevealed(event Event) {
	playerID := event.PlayerID
	if playerID == "" {
		// Try to get player ID from payload if not in event
		if pid, ok := event.Payload["player_id"].(string); ok {
			playerID = pid
		}
	}

	if player, exists := gs.Players[playerID]; exists {
		player.IsRolePubliclyRevealed = true
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

	// Handle backend payload format: sender_name, sender_id, message
	if senderName, ok := event.Payload["sender_name"].(string); ok {
		message.PlayerName = senderName
	} else if playerName, ok := event.Payload["player_name"].(string); ok {
		// Fallback for legacy format
		message.PlayerName = playerName
	}

	if senderID, ok := event.Payload["sender_id"].(string); ok && message.PlayerID == "" {
		// If event.PlayerID is empty, use sender_id from payload
		message.PlayerID = senderID
	}

	if messageText, ok := event.Payload["message"].(string); ok {
		message.Message = messageText
	}

	if isSystem, ok := event.Payload["is_system"].(bool); ok {
		message.IsSystem = isSystem
	}

	// Handle channel information
	if channelID, ok := event.Payload["channel_id"].(string); ok {
		message.ChannelID = channelID
	}

	gs.ChatMessages = append(gs.ChatMessages, message)
}

func (gs *GameState) applyMessageReaction(event Event) {
	messageID, ok := event.Payload["message_id"].(string)
	if !ok {
		return // Invalid payload
	}
	
	emoji, ok := event.Payload["emoji"].(string)
	if !ok {
		return // Invalid payload
	}
	
	playerID := event.PlayerID
	playerName := ""
	
	if name, ok := event.Payload["player_name"].(string); ok {
		playerName = name
	}
	
	// Find the chat message to add the reaction to
	for i := range gs.ChatMessages {
		if gs.ChatMessages[i].ID == messageID {
			// Initialize reactions if nil
			if gs.ChatMessages[i].Reactions == nil {
				gs.ChatMessages[i].Reactions = []EmojiReaction{}
			}
			
			// Check if this player already reacted with this emoji
			found := false
			for j := range gs.ChatMessages[i].Reactions {
				if gs.ChatMessages[i].Reactions[j].PlayerID == playerID && gs.ChatMessages[i].Reactions[j].Emoji == emoji {
					// Remove the reaction (toggle off)
					gs.ChatMessages[i].Reactions = append(gs.ChatMessages[i].Reactions[:j], gs.ChatMessages[i].Reactions[j+1:]...)
					found = true
					break
				}
			}
			
			// If not found, add the reaction
			if !found {
				reaction := EmojiReaction{
					Emoji:      emoji,
					PlayerID:   playerID,
					PlayerName: playerName,
					Timestamp:  event.Timestamp,
				}
				gs.ChatMessages[i].Reactions = append(gs.ChatMessages[i].Reactions, reaction)
			}
			break
		}
	}
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
	pulseCheckPrompt, _ := event.Payload["pulse_check_prompt"].(string)
	effects, _ := event.Payload["effects"].(map[string]interface{})

	gs.CrisisEvent = &CrisisEvent{
		Type:             crisisType,
		Title:            title,
		Description:      description,
		PulseCheckPrompt: pulseCheckPrompt,
		Effects:          effects,
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

	// New persona fields
	personaName, _ := event.Payload["persona_name"].(string)
	jobTitle, _ := event.Payload["job_title"].(string)
	lobbyHandle, _ := event.Payload["lobby_handle"].(string)

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

		// Update player identity with persona information
		player.Alignment = alignment
		if personaName != "" {
			player.Name = personaName
		}
		if jobTitle != "" {
			player.JobTitle = jobTitle
		}
		if lobbyHandle != "" {
			player.LobbyHandle = lobbyHandle
		}
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

func (gs *GameState) applyExtensionVotingTriggered(event Event) {
	// Start extension voting by creating a new vote state
	gs.VoteState = &VoteState{
		Type:         VoteExtension,
		Votes:        make(map[string]string),
		TokenWeights: make(map[string]int),
		Results:      make(map[string]int),
		IsComplete:   false,
	}
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
	// Apply player state changes from the structured payload
	if playerStateChanges, ok := event.Payload["player_state_changes"].(map[string]interface{}); ok {
		for playerID, changesInterface := range playerStateChanges {
			if changes, ok := changesInterface.(map[string]interface{}); ok {
				if player, exists := gs.Players[playerID]; exists {
					// Apply each state change to the player
					for key, value := range changes {
						switch key {
						case "tokens_gained":
							if tokens, ok := value.(float64); ok {
								player.Tokens += int(tokens)
							} else if tokens, ok := value.(int); ok {
								player.Tokens += tokens
							}
						case "status_message":
							if msg, ok := value.(string); ok {
								player.StatusMessage = msg
							}
						case "alignment":
							if align, ok := value.(string); ok {
								player.Alignment = align
							}
						case "ai_equity":
							if equity, ok := value.(float64); ok {
								player.AIEquity = int(equity)
							} else if equity, ok := value.(int); ok {
								player.AIEquity = equity
							}
						case "project_milestones":
							if milestones, ok := value.(float64); ok {
								player.ProjectMilestones = int(milestones)
							} else if milestones, ok := value.(int); ok {
								player.ProjectMilestones = milestones
							}
						case "has_used_ability":
							if used, ok := value.(bool); ok {
								player.HasUsedAbility = used
							}
						case "role_unlocked":
							if unlocked, ok := value.(bool); ok && unlocked {
								if player.Role != nil {
									player.Role.IsUnlocked = true
									if player.Role.Ability != nil {
										player.Role.Ability.IsReady = true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Reset night action tracking for all players
	for _, player := range gs.Players {
		player.LastNightAction = nil
		// HasUsedAbility is handled in the state changes section above
		// If not explicitly set in state changes, reset to false
		if playerStateChanges, ok := event.Payload["player_state_changes"].(map[string]interface{}); ok {
			if playerChanges, exists := playerStateChanges[player.ID]; exists {
				if changes, ok := playerChanges.(map[string]interface{}); ok {
					if _, hasAbilitySet := changes["has_used_ability"]; !hasAbilitySet {
						player.HasUsedAbility = false
					}
				} else {
					player.HasUsedAbility = false
				}
			} else {
				player.HasUsedAbility = false
			}
		} else {
			player.HasUsedAbility = false
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

// Specific semantic event handlers

func (gs *GameState) applyClientError(event Event) {
	// Client errors don't modify game state - they're for client handling only
	// Could log for analytics but no state changes needed
}

func (gs *GameState) applySitrepPublished(event Event) {
	// SITREP events contain crisis and day information but don't modify core game state
	// The crisis data is already applied via EventCrisisTriggered
	// This event is primarily for client rendering
}

func (gs *GameState) applyLiaisonProtocolActivated(event Event) {
	// Liaison protocol activation modifies game mechanics
	// Store activation state for mining bonuses and other effects
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{Effects: make(map[string]interface{})}
	}
	if gs.CrisisEvent.Effects == nil {
		gs.CrisisEvent.Effects = make(map[string]interface{})
	}

	aiPercentage, _ := event.Payload["ai_percentage"].(float64)
	bonusSlots, _ := event.Payload["mining_bonus_slots"].(float64)

	gs.CrisisEvent.Effects["liaison_protocol_active"] = true
	gs.CrisisEvent.Effects["liaison_ai_percentage"] = aiPercentage
	gs.CrisisEvent.Effects["liaison_mining_bonus"] = int(bonusSlots)
}

func (gs *GameState) applyLiaisonIntelRevealed(event Event) {
	// Intel reveals don't modify game state - they're informational
	// Could track for analytics but no state changes needed
}

func (gs *GameState) applyAIConversionBlocked(event Event) {
	// Conversion blocks are informational and don't modify state
	// The actual blocking logic is handled in the night resolution
}

func (gs *GameState) applyGameRuleModified(event Event) {
	// Rule modifications can affect various game mechanics
	// Store in crisis effects for reference during rule evaluation
	if gs.CrisisEvent == nil {
		gs.CrisisEvent = &CrisisEvent{Effects: make(map[string]interface{})}
	}
	if gs.CrisisEvent.Effects == nil {
		gs.CrisisEvent.Effects = make(map[string]interface{})
	}

	ruleCategory, _ := event.Payload["rule_category"].(string)
	modificationType, _ := event.Payload["modification_type"].(string)
	source, _ := event.Payload["source"].(string)

	ruleKey := fmt.Sprintf("rule_mod_%s_%s", ruleCategory, modificationType)
	gs.CrisisEvent.Effects[ruleKey] = source
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

	// Initialize pulse check responses for this day
	gs.PulseCheckResponses = make(map[string]string)
}

func (gs *GameState) applyPulseCheckUpdated(event Event) {
	// 1. Get new submission data from the event payload.
	messageID, _ := event.Payload["message_id"].(string)
	question, _ := event.Payload["question"].(string)
	playerID, _ := event.Payload["player_id"].(string)
	response, _ := event.Payload["response"].(string)
	playerName, _ := event.Payload["player_name"].(string)

	// 2. Update the internal state map first. This is the source of truth.
	if gs.PulseCheckResponses == nil {
		gs.PulseCheckResponses = make(map[string]string)
	}
	// Only update if this is a real player submission (not initial message creation)
	if playerID != "" && response != "" {
		gs.PulseCheckResponses[playerID] = response
		if p, ok := gs.Players[playerID]; ok {
			p.HasSubmittedPulseCheck = true
		}
	}

	// 3. Prepare the complete metadata for the UI component FROM THE NOW-UPDATED STATE.
	uiResponses := make(map[string]interface{})
	for pid, presp := range gs.PulseCheckResponses {
		var pName string
		// Optimization: if we have the name in the current event, use it. Otherwise, look it up.
		if pid == playerID {
			pName = playerName
		} else if p, ok := gs.Players[pid]; ok {
			pName = p.Name
		}
		if pName != "" {
			uiResponses[pName] = presp
		}
	}
	totalResponses := len(gs.PulseCheckResponses)

	// 4. Find and update the existing PulseCheck message, or create it if it doesn't exist.
	found := false
	for i := range gs.ChatMessages {
		if gs.ChatMessages[i].ID == messageID {
			// Update existing message
			gs.ChatMessages[i].Message = question
			gs.ChatMessages[i].Timestamp = event.Timestamp
			if gs.ChatMessages[i].Metadata == nil {
				gs.ChatMessages[i].Metadata = make(map[string]interface{})
			}
			gs.ChatMessages[i].Metadata["pulseCheckResponses"] = uiResponses
			gs.ChatMessages[i].Metadata["total_responses"] = totalResponses
			found = true
			break
		}
	}

	if !found {
		// Create new pulse check message if it doesn't exist
		message := ChatMessage{
			ID:         messageID,
			PlayerID:   "",
			PlayerName: "NEXUS",
			Message:    question,
			Timestamp:  event.Timestamp,
			IsSystem:   true,
			Type:       "PULSE_CHECK", // Set the type for the client renderer
			ChannelID:  "#war-room",   // Pulse checks always happen in the main channel
			Metadata: map[string]interface{}{
				"pulseCheckResponses": uiResponses,
				"question":            question,
				"total_responses":     totalResponses,
			},
		}
		gs.ChatMessages = append(gs.ChatMessages, message)
	}
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
		player.StatusMessage = status
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
func (gs *GameState) applyKPIAssigned(event Event) {
	playerID := event.PlayerID
	kpiType, _ := event.Payload["kpi_type"].(string)
	description, _ := event.Payload["description"].(string)
	target, _ := event.Payload["target"].(float64)
	reward, _ := event.Payload["reward"].(string)

	if player, exists := gs.Players[playerID]; exists {
		player.PersonalKPI = &PersonalKPI{
			Type:        KPIType(kpiType),
			Description: description,
			Progress:    0,
			Target:      int(target),
			IsCompleted: false,
			Reward:      reward,
		}
	}
}

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
			ExpiresAt:   event.Timestamp.Add(time.Duration(durationHours) * time.Hour),
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

// applySkipVoteUpdated handles skip vote events
func (gs *GameState) applySkipVoteUpdated(event Event) {
	playerID := event.PlayerID
	hasVoted, _ := event.Payload["has_voted"].(bool)

	if gs.SkipVotes == nil {
		gs.SkipVotes = make(map[string]bool)
	}

	if hasVoted {
		gs.SkipVotes[playerID] = true
	} else {
		delete(gs.SkipVotes, playerID)
	}
}

// Whistleblower Protocol event handlers

// applyWhistleblowerVotingStarted initializes whistleblower voting state
func (gs *GameState) applyWhistleblowerVotingStarted(event Event) {
	crisisOptions, _ := event.Payload["crisis_options"].([]interface{})

	// Convert crisis options from interface{} to CrisisEventOption structs
	var options []CrisisEventOption
	for _, optionInterface := range crisisOptions {
		if optionMap, ok := optionInterface.(map[string]interface{}); ok {
			option := CrisisEventOption{
				Type:        optionMap["type"].(string),
				Title:       optionMap["title"].(string),
				Description: optionMap["description"].(string),
			}
			options = append(options, option)
		}
	}

	gs.WhistleblowerVoting = &WhistleblowerVoting{
		IsActive:      true,
		CrisisOptions: options,
		Votes:         make(map[string]string),
		VoteResults:   make(map[string]int),
		IsComplete:    false,
	}
}

// applyWhistleblowerVoteCast records a whistleblower vote
func (gs *GameState) applyWhistleblowerVoteCast(event Event) {
	playerID := event.PlayerID
	crisisType, _ := event.Payload["crisis_type"].(string)

	// Initialize whistleblower voting if it doesn't exist
	if gs.WhistleblowerVoting == nil {
		gs.WhistleblowerVoting = &WhistleblowerVoting{
			IsActive:    true,
			Votes:       make(map[string]string),
			VoteResults: make(map[string]int),
			IsComplete:  false,
		}
	}

	// Record the vote
	gs.WhistleblowerVoting.Votes[playerID] = crisisType

	// Recalculate vote results
	gs.WhistleblowerVoting.VoteResults = make(map[string]int)
	for _, votedCrisis := range gs.WhistleblowerVoting.Votes {
		gs.WhistleblowerVoting.VoteResults[votedCrisis]++
	}
}

// applyWhistleblowerVotingCompleted finalizes whistleblower voting
func (gs *GameState) applyWhistleblowerVotingCompleted(event Event) {
	selectedCrisis, _ := event.Payload["selected_crisis"].(string)

	if gs.WhistleblowerVoting == nil {
		gs.WhistleblowerVoting = &WhistleblowerVoting{
			Votes:       make(map[string]string),
			VoteResults: make(map[string]int),
		}
	}

	gs.WhistleblowerVoting.SelectedCrisis = selectedCrisis
	gs.WhistleblowerVoting.IsComplete = true
	gs.WhistleblowerVoting.IsActive = false
}

// ProcessPlayerAction is the formal Action-to-Event translation layer
// This function takes the current state and a player's desired action,
// performs all necessary validation, and returns the list of Events that should result.
// It does NOT modify the state itself - that is done by ApplyEvent.
func ProcessPlayerAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	// Validate basic action requirements
	if err := validateActionBasics(gameState, action, currentTime); err != nil {
		return nil, err
	}

	// Route to specific action handlers
	switch action.Type {
	case ActionSubmitVote:
		return processVoteAction(gameState, action, currentTime)
	case ActionSubmitSkipVote:
		return processSkipVoteAction(gameState, action, currentTime)
	case ActionSubmitNightAction:
		return processNightAction(gameState, action, currentTime)
	case ActionMineTokens:
		return processMiningAction(gameState, action, currentTime)
	case ActionSendMessage:
		return processChatAction(gameState, action, currentTime)
	case ActionLeaveGame:
		return processLeaveGameAction(gameState, action, currentTime)
	case ActionAbandonGame:
		return processAbandonGameAction(gameState, action, currentTime)
	case ActionUseAbility:
		return processAbilityAction(gameState, action, currentTime)

	// Role-specific actions
	case ActionRunAudit, ActionOverclockServers, ActionIsolateNode, ActionPerformanceReview, ActionReallocateBudget, ActionPivot, ActionDeployHotfix:
		return processRoleAction(gameState, action, currentTime)

	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// validateActionBasics performs common validation for all actions
func validateActionBasics(gameState GameState, action Action, currentTime time.Time) error {
	// Check if player exists
	player, exists := gameState.Players[action.PlayerID]
	if !exists {
		return fmt.Errorf("player %s not found", action.PlayerID)
	}

	// Check if player is alive (most actions require this)
	if !player.IsAlive && action.Type != ActionLeaveGame {
		return fmt.Errorf("player %s is not alive", action.PlayerID)
	}

	// Validate game ID matches
	if action.GameID != gameState.ID {
		return fmt.Errorf("action game ID %s does not match current game %s", action.GameID, gameState.ID)
	}

	return nil
}

// processVoteAction handles voting actions
func processVoteAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Check if player can vote in current phase
	if !CanPlayerVote(*player, gameState.Phase.Type, currentTime) {
		return nil, fmt.Errorf("player %s cannot vote in phase %s", action.PlayerID, gameState.Phase.Type)
	}

	// Extract vote target from payload
	targetPlayerID, ok := action.Payload["target_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target_id in vote action")
	}

	// Validate target exists and is alive (unless it's a verdict vote)
	if action.Payload["vote_type"] != "VERDICT" {
		targetPlayer, exists := gameState.Players[targetPlayerID]
		if !exists {
			return nil, fmt.Errorf("vote target %s not found", targetPlayerID)
		}
		if !targetPlayer.IsAlive {
			return nil, fmt.Errorf("cannot vote for eliminated player %s", targetPlayerID)
		}
	}

	// Generate vote cast event
	events := []Event{
		{
			ID:        fmt.Sprintf("vote_%s_%s_%d", action.PlayerID, targetPlayerID, currentTime.UnixNano()),
			Type:      EventVoteCast,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload: map[string]interface{}{
				"target_id":    targetPlayerID,
				"vote_type":    action.Payload["vote_type"],
				"token_weight": player.Tokens,
			},
		},
	}

	return events, nil
}

// processMiningAction handles token mining actions
func processMiningAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Check if player can use night actions (mining is a night action)
	if !CanPlayerUseNightAction(*player, ActionMine, currentTime) {
		return nil, fmt.Errorf("player %s cannot mine tokens at this time", action.PlayerID)
	}

	// Extract beneficiary from payload
	beneficiaryID, ok := action.Payload["target_player_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid target_player_id in mining action")
	}

	// Validate beneficiary exists and is alive
	beneficiary, exists := gameState.Players[beneficiaryID]
	if !exists {
		return nil, fmt.Errorf("mining beneficiary %s not found", beneficiaryID)
	}
	if !beneficiary.IsAlive {
		return nil, fmt.Errorf("cannot mine for eliminated player %s", beneficiaryID)
	}

	// Generate mining attempt event
	events := []Event{
		{
			ID:        fmt.Sprintf("mining_%s_%s_%d", action.PlayerID, beneficiaryID, currentTime.UnixNano()),
			Type:      EventMiningAttempted,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload: map[string]interface{}{
				"target_player_id": beneficiaryID,
				"difficulty":     0.2, // Standard mining difficulty
			},
		},
	}

	return events, nil
}

// processChatAction handles chat message actions
func processChatAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Check if player can send messages
	if !CanPlayerSendMessage(*player, currentTime) {
		return nil, fmt.Errorf("player %s cannot send messages at this time", action.PlayerID)
	}

	// Extract message content
	messages, ok := action.Payload["messages"].([]string)
	if !ok || len(messages) == 0 {
		return nil, fmt.Errorf("missing or empty messages content")
	}

	var events []Event
	for _, content := range messages {
		if content == "" {
			continue // Skip empty messages
		}

		// Check for message corruption
		isCorrupted := IsMessageCorrupted(*player, content, currentTime)

		// Generate chat message event
		events = append(events, Event{
			ID:        fmt.Sprintf("chat_%s_%d", action.PlayerID, currentTime.UnixNano()),
			Type:      EventChatMessage,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload: map[string]interface{}{
				"sender_id":   action.PlayerID,
				"sender_name": player.Name,
				"message":     content,
				"corrupted":   isCorrupted,
				"channel_id":  action.Payload["channel_id"],
			},
		})
	}

	return events, nil
}

// processNightAction handles night action submissions
func processNightAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Extract night action details
	actionType, ok := action.Payload["action_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing action_type in night action")
	}

	nightActionType := NightActionType(actionType)

	// Check if player can use this night action
	if !CanPlayerUseNightAction(*player, nightActionType, currentTime) {
		return nil, fmt.Errorf("player %s cannot use night action %s", action.PlayerID, actionType)
	}

	// Validate target if provided
	if targetID, hasTarget := action.Payload["target_id"].(string); hasTarget {
		if targetPlayer, exists := gameState.Players[targetID]; !exists || !targetPlayer.IsAlive {
			return nil, fmt.Errorf("invalid night action target %s", targetID)
		}
		if !IsValidNightActionTarget(*player, *gameState.Players[targetID], nightActionType) {
			return nil, fmt.Errorf("invalid target for night action %s", actionType)
		}
	}

	// Generate night action submitted event
	events := []Event{
		{
			ID:        fmt.Sprintf("night_action_%s_%s_%d", action.PlayerID, actionType, currentTime.UnixNano()),
			Type:      EventNightActionSubmitted,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload:   action.Payload, // Pass through the action payload
		},
	}

	return events, nil
}

// processLeaveGameAction handles player leaving the game
func processLeaveGameAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	// Generate player left event
	events := []Event{
		{
			ID:        fmt.Sprintf("leave_%s_%d", action.PlayerID, currentTime.UnixNano()),
			Type:      EventPlayerLeft,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload:   map[string]interface{}{},
		},
	}

	return events, nil
}

// processAbandonGameAction handles player abandoning the game with consequences
func processAbandonGameAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Check if game is in progress (can't abandon from lobby)
	if gameState.Phase.Type == PhaseLobby || gameState.Phase.Type == PhaseGameOver {
		return nil, fmt.Errorf("cannot abandon game in phase %s", gameState.Phase.Type)
	}

	// Extract the player's role for revelation
	var revealedRole string
	if player.Role != nil {
		revealedRole = string(player.Role.Type)
	} else {
		revealedRole = "UNKNOWN"
	}

	// Generate player abandoned event
	events := []Event{
		{
			ID:        fmt.Sprintf("abandon_%s_%d", action.PlayerID, currentTime.UnixNano()),
			Type:      EventPlayerAbandoned,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload: map[string]interface{}{
				"revealed_role": revealedRole,
				"player_name":   player.Name,
			},
		},
	}

	return events, nil
}

// processAbilityAction handles role ability usage
func processAbilityAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Check if player has a role and can use abilities
	if player.Role == nil || !player.Role.IsUnlocked || player.HasUsedAbility {
		return nil, fmt.Errorf("player %s cannot use role abilities", action.PlayerID)
	}

	// Extract ability details
	abilityType, ok := action.Payload["ability_type"].(string)
	if !ok {
		return nil, fmt.Errorf("missing ability_type in ability action")
	}

	// Generate ability used event
	events := []Event{
		{
			ID:        fmt.Sprintf("ability_%s_%s_%d", action.PlayerID, abilityType, currentTime.UnixNano()),
			Type:      EventRoleAbilityUnlocked,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload:   action.Payload,
		},
	}

	return events, nil
}

// processRoleAction handles role-specific actions
func processRoleAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	player := gameState.Players[action.PlayerID]

	// Validate player has appropriate role
	if player.Role == nil || !player.Role.IsUnlocked {
		return nil, fmt.Errorf("player %s does not have an unlocked role", action.PlayerID)
	}

	// Map action types to role types and generate appropriate events
	var eventType EventType
	switch action.Type {
	case ActionRunAudit:
		if player.Role.Type != RoleCTO {
			return nil, fmt.Errorf("only CTO can run audits")
		}
		eventType = EventRunAudit
	case ActionOverclockServers:
		if player.Role.Type != RoleCTO {
			return nil, fmt.Errorf("only CTO can overclock servers")
		}
		eventType = EventOverclockServers
	case ActionIsolateNode:
		if player.Role.Type != RoleCTO {
			return nil, fmt.Errorf("only CTO can isolate nodes")
		}
		eventType = EventIsolateNode
	default:
		return nil, fmt.Errorf("unsupported role action: %s", action.Type)
	}

	// Generate role action event
	events := []Event{
		{
			ID:        fmt.Sprintf("role_%s_%s_%d", action.PlayerID, action.Type, currentTime.UnixNano()),
			Type:      eventType,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload:   action.Payload,
		},
	}

	return events, nil
}

// processSkipVoteAction handles skip vote actions
func processSkipVoteAction(gameState GameState, action Action, currentTime time.Time) ([]Event, error) {
	// Check if phase allows skipping (exclude TRIAL and GAME_OVER phases)
	if gameState.Phase.Type == PhaseTrial || gameState.Phase.Type == PhaseGameOver || gameState.Phase.Type == PhaseLobby {
		return nil, fmt.Errorf("cannot skip phase %s", gameState.Phase.Type)
	}

	// Check if player already voted to skip
	if gameState.SkipVotes != nil && gameState.SkipVotes[action.PlayerID] {
		return nil, fmt.Errorf("player %s has already voted to skip", action.PlayerID)
	}

	// Generate skip vote updated event
	events := []Event{
		{
			ID:        fmt.Sprintf("skip_vote_%s_%d", action.PlayerID, currentTime.UnixNano()),
			Type:      EventSkipVoteUpdated,
			PlayerID:  action.PlayerID,
			GameID:    gameState.ID,
			Timestamp: currentTime,
			Payload: map[string]interface{}{
				"has_voted": true,
			},
		},
	}

	return events, nil
}

// CalculateChecksum computes a SHA256 hash of the game state data for integrity verification
func (gs *GameState) CalculateChecksum() (string, error) {
	// Create a copy of the state without the checksum field to avoid circular references
	stateForHashing := *gs
	stateForHashing.Checksum = ""

	// Serialize the state to JSON for consistent hashing
	data, err := json.Marshal(stateForHashing)
	if err != nil {
		return "", fmt.Errorf("failed to marshal game state for checksum: %v", err)
	}

	// Calculate SHA256 hash
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// UpdateChecksum recalculates and updates the checksum field
func (gs *GameState) UpdateChecksum() error {
	checksum, err := gs.CalculateChecksum()
	if err != nil {
		return err
	}
	gs.Checksum = checksum
	return nil
}

// ValidateChecksum verifies that the stored checksum matches the calculated checksum
func (gs *GameState) ValidateChecksum() (bool, error) {
	if gs.Checksum == "" {
		// No checksum to validate - this is allowed for backwards compatibility
		return true, nil
	}

	expectedChecksum, err := gs.CalculateChecksum()
	if err != nil {
		return false, fmt.Errorf("failed to calculate checksum for validation: %v", err)
	}

	return gs.Checksum == expectedChecksum, nil
}