
package core

import (
	"encoding/json"
	"time"
)

// Event represents a game event that changes state
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	GameID    string                 `json:"gameId"`
	PlayerID  string                 `json:"playerId,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// NewEventWithTypedPayload creates a new event with a strongly-typed payload
// This helper ensures consistent event creation and eliminates defensive parsing
func NewEventWithTypedPayload(id string, eventType EventType, gameID, playerID string, timestamp time.Time, payload interface{}) Event {
	// Convert the typed payload to a map[string]interface{} for backwards compatibility
	// During the transition period, this allows both old and new code to work together
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		// Fallback to empty payload if marshaling fails
		return Event{
			ID:        id,
			Type:      eventType,
			GameID:    gameID,
			PlayerID:  playerID,
			Timestamp: timestamp,
			Payload:   make(map[string]interface{}),
		}
	}

	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payloadMap); err != nil {
		// Fallback to empty payload if unmarshaling fails
		payloadMap = make(map[string]interface{})
	}

	return Event{
		ID:        id,
		Type:      eventType,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: timestamp,
		Payload:   payloadMap,
	}
}

// EventType represents different types of game events
type EventType string

const (
	// Game lifecycle events
	EventGameCreated              EventType = "GAME_CREATED"
	EventGameStarted              EventType = "GAME_STARTED"
	EventGameStartCountdownStart  EventType = "GAME_START_COUNTDOWN_INITIATED"
	EventGameStartCountdownUpdate EventType = "GAME_START_COUNTDOWN_UPDATE"
	EventGameStartCountdownCancel EventType = "GAME_START_COUNTDOWN_CANCELLED"
	EventGameEnded                EventType = "GAME_ENDED"
	EventPhaseChanged             EventType = "PHASE_CHANGED"

	// Player events
	EventPlayerJoined        EventType = "PLAYER_JOINED"
	EventPlayerLeft          EventType = "PLAYER_LEFT"
	EventPlayerEliminated    EventType = "PLAYER_ELIMINATED"
	EventPlayerAbandoned     EventType = "PLAYER_ABANDONED"
	EventPlayerRoleRevealed  EventType = "PLAYER_ROLE_REVEALED"
	EventPlayerAligned       EventType = "PLAYER_ALIGNED"
	EventAlignmentChanged    EventType = "ALIGNMENT_CHANGED"
	EventPlayerShocked       EventType = "PLAYER_SHOCKED"
	EventHostTransferred     EventType = "HOST_TRANSFERRED"

	// Voting events
	EventVoteStarted              EventType = "VOTE_STARTED"
	EventVoteCast                 EventType = "VOTE_CAST"
	EventVoteTallyUpdated         EventType = "VOTE_TALLY_UPDATED"
	EventVoteCompleted            EventType = "VOTE_COMPLETED"
	EventPlayerNominated          EventType = "PLAYER_NOMINATED"
	EventExtensionVotingTriggered EventType = "EXTENSION_VOTING_TRIGGERED"

	// Token and Mining events
	EventTokensAwarded    EventType = "TOKENS_AWARDED"
	EventTokensSpent      EventType = "TOKENS_SPENT"
	EventMiningAttempted  EventType = "MINING_ATTEMPTED"
	// Deprecated: see ADR-006. Use EventNightActionsResolved with comprehensive payload instead of individual mining events
	EventMiningSuccessful EventType = "MINING_SUCCESSFUL"
	// Deprecated: see ADR-006. Use EventNightActionsResolved with comprehensive payload instead of individual mining events
	EventMiningFailed     EventType = "MINING_FAILED"

	// Night Action events
	EventNightActionsResolved EventType = "NIGHT_ACTIONS_RESOLVED"
	// Deprecated: see ADR-006. Use EventNightActionsResolved with comprehensive payload instead of individual action events
	EventPlayerBlocked        EventType = "PLAYER_BLOCKED"
	// Deprecated: see ADR-006. Use EventNightActionsResolved with comprehensive payload instead of individual action events
	EventPlayerProtected      EventType = "PLAYER_PROTECTED"
	// Deprecated: see ADR-006. Use EventNightActionsResolved with comprehensive payload instead of individual action events
	EventPlayerInvestigated   EventType = "PLAYER_INVESTIGATED"

	// AI and Conversion events
	EventAIConversionAttempt EventType = "AI_CONVERSION_ATTEMPT"
	EventAIConversionSuccess EventType = "AI_CONVERSION_SUCCESS"
	EventAIConversionFailed  EventType = "AI_CONVERSION_FAILED"
	EventAIRevealed          EventType = "AI_REVEALED"

	// Communication events
	EventChatMessage         EventType = "CHAT_MESSAGE"
	EventMessageReaction     EventType = "MESSAGE_REACTION"
	EventSystemMessage       EventType = "SYSTEM_MESSAGE" // DEPRECATED: Use specific semantic events
	EventPrivateNotification EventType = "PRIVATE_NOTIFICATION"
	EventIncitingIncident    EventType = "INCITING_INCIDENT"
	EventLoebmateMessage     EventType = "LOEBMATE_MESSAGE"

	// Specific semantic events replacing SYSTEM_MESSAGE
	EventClientError              EventType = "CLIENT_ERROR"
	EventSitrepPublished          EventType = "SITREP_PUBLISHED"
	EventLiaisonProtocolActivated EventType = "LIAISON_PROTOCOL_ACTIVATED"
	EventLiaisonIntelRevealed     EventType = "LIAISON_INTEL_REVEALED"
	EventAIConversionBlocked      EventType = "AI_CONVERSION_BLOCKED"
	EventGameRuleModified         EventType = "GAME_RULE_MODIFIED"

	// Crisis and Special events
	EventCrisisTriggered     EventType = "CRISIS_TRIGGERED"
	EventPulseCheckStarted   EventType = "PULSE_CHECK_STARTED"
	// EventPulseCheckSubmitted is deprecated in favor of EventPulseCheckUpdated
	EventPulseCheckUpdated   EventType = "PULSE_CHECK_UPDATED"
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
	EventGameStateUpdate     EventType = "GAME_STATE_UPDATE"
	EventLobbyStateUpdate    EventType = "LOBBY_STATE_UPDATE"
	EventClientIdentified    EventType = "CLIENT_IDENTIFIED"
	EventChatHistorySnapshot EventType = "CHAT_HISTORY_SNAPSHOT"
	EventPlayerReconnected         EventType = "PLAYER_RECONNECTED"
	EventPlayerDisconnected        EventType = "PLAYER_DISCONNECTED"
	EventPlayerConnectionStatusChanged EventType = "PLAYER_CONNECTION_STATUS_CHANGED"
	EventSyncComplete        EventType = "SYNC_COMPLETE"
	EventRateLimitExceeded   EventType = "RATE_LIMIT_EXCEEDED"
	EventSessionExpired      EventType = "SESSION_EXPIRED"
	EventForceLogout         EventType = "FORCE_LOGOUT"

	// Phase skipping events
	EventSkipVoteUpdated EventType = "SKIP_VOTE_UPDATED"

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

	// Player Status events
	EventSlackStatusChanged EventType = "SLACK_STATUS_CHANGED"
	EventPartingShotSet     EventType = "PARTING_SHOT_SET"
	EventWhisperSent        EventType = "WHISPER_SENT"

	// Personal KPI events
	EventKPIAssigned  EventType = "KPI_ASSIGNED"
	EventKPIProgress  EventType = "KPI_PROGRESS"
	EventKPICompleted EventType = "KPI_COMPLETED"

	// System Shock events
	EventSystemShockApplied   EventType = "SYSTEM_SHOCK_APPLIED"
	EventShockEffectTriggered EventType = "SHOCK_EFFECT_TRIGGERED"

	// AI Equity events
	EventAIEquityChanged EventType = "AI_EQUITY_CHANGED"
	EventEquityThreshold EventType = "EQUITY_THRESHOLD"

	// Whistleblower Protocol events
	EventWhistleblowerVotingStarted   EventType = "WHISTLEBLOWER_VOTING_STARTED"
	EventWhistleblowerVoteCast        EventType = "WHISTLEBLOWER_VOTE_CAST"
	EventWhistleblowerVotingCompleted EventType = "WHISTLEBLOWER_VOTING_COMPLETED"

	// Corporate Mandate events
	EventMandateActivated EventType = "MANDATE_ACTIVATED"
	EventMandateEffect    EventType = "MANDATE_EFFECT"

	// Spectator events
	EventSpectatorStateSnapshot EventType = "SPECTATOR_STATE_SNAPSHOT"
	EventSpectatorJoined        EventType = "SPECTATOR_JOINED"
	EventSpectatorLeft          EventType = "SPECTATOR_LEFT"
	EventSpectatorChatMessage   EventType = "SPECTATOR_CHAT_MESSAGE"
)

// Action represents a player action that can generate events
type Action struct {
	Type      ActionType             `json:"type"`
	PlayerID  string                 `json:"playerId"`
	GameID    string                 `json:"gameId"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// ActionType represents different types of player actions
type ActionType string

const (
	// Lobby actions
	ActionCreateGame ActionType = "CREATE_GAME"
	ActionJoinGame   ActionType = "JOIN_GAME"
	ActionLeaveGame  ActionType = "LEAVE_GAME"
	ActionStartGame  ActionType = "START_GAME"

	// Party actions
	ActionCreateParty   ActionType = "CREATE_PARTY"
	ActionInviteToParty ActionType = "INVITE_TO_PARTY"
	ActionJoinParty     ActionType = "JOIN_PARTY"
	ActionLeaveParty    ActionType = "LEAVE_PARTY"

	// Communication actions
	ActionSendMessage      ActionType = "SEND_MESSAGE"
	ActionReactToMessage   ActionType = "REACT_TO_MESSAGE"
	ActionSubmitPulseCheck ActionType = "SUBMIT_PULSE_CHECK"

	// Voting actions
	ActionSubmitVote             ActionType = "SUBMIT_VOTE"
	ActionExtendDiscussion       ActionType = "EXTEND_DISCUSSION"
	ActionSubmitSkipVote         ActionType = "SUBMIT_SKIP_VOTE"
	ActionTriggerExtensionVoting ActionType = "TRIGGER_EXTENSION_VOTING"

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
	ActionSetSlackStatus      ActionType = "SET_SLACK_STATUS"
	ActionSubmitExitInterview ActionType = "SUBMIT_EXIT_INTERVIEW"
	ActionWhisper             ActionType = "WHISPER"

	// Meta actions
	ActionReconnect      ActionType = "RECONNECT"
	ActionAbandonGame    ActionType = "ABANDON_GAME"
	ActionSyncLobbyState ActionType = "SYNC_LOBBY_STATE"
	
	// Internal server actions (for GameLifecycleManager -> GameActor communication)
	ActionSetPlayerConnectionStatus ActionType = "SET_PLAYER_CONNECTION_STATUS"
	ActionAbandonPlayer            ActionType = "ABANDON_PLAYER"
	ActionAssignCorporateMandate   ActionType = "ASSIGN_CORPORATE_MANDATE"

	// Whistleblower Protocol actions
	ActionSubmitWhistleblowerVote ActionType = "SUBMIT_WHISTLEBLOWER_VOTE"

	// Spectator actions
	ActionPostSpectatorMessage ActionType = "POST_SPECTATOR_MESSAGE"
)

// Phase represents the current game phase
type Phase struct {
	Type      PhaseType     `json:"type"`
	StartTime time.Time     `json:"startTime"`
	Duration  time.Duration `json:"duration"`
}

// PhaseType represents different phases of the game
type PhaseType string

const (
	PhaseLobby      PhaseType = "LOBBY"
	PhaseSitrep     PhaseType = "SITREP"
	PhasePulseCheck PhaseType = "PULSE_CHECK"
	PhaseDiscussion PhaseType = "DISCUSSION"
	PhaseExtension  PhaseType = "EXTENSION"
	PhaseNomination PhaseType = "NOMINATION"
	PhaseTrial      PhaseType = "TRIAL"
	PhaseVerdict    PhaseType = "VERDICT"
	PhaseNight      PhaseType = "NIGHT"
	PhaseGameOver   PhaseType = "GAME_OVER"
)

// Player represents a human or AI player
type Player struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	JobTitle          string       `json:"jobTitle"`
	ControlType       string       `json:"controlType"` // "HUMAN" or "AI"
	Status            PlayerStatus `json:"status"`
	IsAlive           bool         `json:"isAlive"`
	ConnectionStatus  string       `json:"connectionStatus"` // "CONNECTED", "DISCONNECTED"
	Tokens            int          `json:"tokens"`
	ProjectMilestones int          `json:"projectMilestones"`
	StatusMessage     string       `json:"statusMessage"`
	JoinedAt          time.Time    `json:"joinedAt"`

	// Private fields (only visible to the player themselves)
	Alignment              string       `json:"alignment,omitempty"` // "HUMAN" or "AI" or "ALIGNED"
	Role                   *Role        `json:"role,omitempty"`
	PersonalKPI            *PersonalKPI `json:"personalKPI,omitempty"`
	AIEquity               int          `json:"aiEquity,omitempty"` // For alignment conversion
	HasUsedAbility         bool         `json:"hasUsedAbility,omitempty"`
	LastNightAction        *NightAction `json:"lastNightAction,omitempty"`
	HasSubmittedPulseCheck bool         `json:"hasSubmittedPulseCheck,omitempty"`
	LobbyHandle            string       `json:"lobbyHandle,omitempty"` // Original lobby identity for post-game reveal
	BootcampPoints         int          `json:"bootcampPoints,omitempty"` // Intern role resource for shadowing abilities
	WhisperUsedDay         int          `json:"whisperUsedDay,omitempty"` // Day number when whisper was last used

	// FTUE and Assistance Settings
	SeenHints            map[string]bool `json:"seenHints,omitempty"`             // Phase hints the player has seen (key: phase name)
	DisableLoebmateHints bool            `json:"disableLoebmateHints,omitempty"`  // Whether to disable Loebmate assistance

	// Public status and effects
	SlackStatus            string        `json:"slackStatus,omitempty"`
	PartingShot            string        `json:"partingShot,omitempty"`
	SystemShocks           []SystemShock `json:"systemShocks,omitempty"`
	IsRolePubliclyRevealed bool          `json:"isRolePubliclyRevealed"`
}

// PlayerStatus represents the status of a player in the game
type PlayerStatus string

const (
	PlayerStatusAlive        PlayerStatus = "ALIVE"
	PlayerStatusEliminated   PlayerStatus = "ELIMINATED"
	PlayerStatusAbandoned    PlayerStatus = "ABANDONED"
	PlayerStatusDisconnected PlayerStatus = "DISCONNECTED"
)

// Role represents a player's role and abilities
type Role struct {
	Type        RoleType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsUnlocked  bool     `json:"isUnlocked"`
	Ability     *Ability `json:"ability,omitempty"`
}

// RoleType represents different player roles
type RoleType string

const (
	RoleCISO      RoleType = "CISO"      // Chief Information Security Officer
	RoleCEO       RoleType = "CEO"       // Chief Executive Officer
	RoleCTO       RoleType = "CTO"       // Chief Technology Officer
	RoleCOO       RoleType = "COO"       // Chief Operating Officer
	RoleCFO       RoleType = "CFO"       // Chief Financial Officer
	RoleEthics    RoleType = "ETHICS"    // VP, Ethics & Alignment
	RolePlatforms RoleType = "PLATFORMS" // VP, Platforms
	RoleIntern    RoleType = "INTERN"    // Intern
)

// Ability represents a role's special ability
type Ability struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsReady     bool   `json:"isReady"`
}

// PersonalKPI represents a player's secret objective
type PersonalKPI struct {
	Type        KPIType `json:"type"`
	Description string  `json:"description"`
	Progress    int     `json:"progress"`
	Target      int     `json:"target"`
	IsCompleted bool    `json:"isCompleted"`
	Reward      string  `json:"reward"`
}

// KPIType represents different types of personal objectives
type KPIType string

const (
	KPICapitalist        KPIType = "CAPITALIST"         // End with most tokens
	KPIGuardian          KPIType = "GUARDIAN"           // Keep CISO alive to Day 4
	KPIInquisitor        KPIType = "INQUISITOR"         // Vote correctly 3 times
	KPISuccessionPlanner KPIType = "SUCCESSION_PLANNER" // End with exactly 2 humans
	KPIScapegoat         KPIType = "SCAPEGOAT"          // Get eliminated unanimously
)

// SystemShock represents temporary effects from failed AI conversion
type SystemShock struct {
	Type        ShockType `json:"type"`
	Description string    `json:"description"`
	ExpiresAt   time.Time `json:"expiresAt"`
	IsActive    bool      `json:"isActive"`
}

// ShockType represents different shock effects
type ShockType string

const (
	ShockMessageCorruption ShockType = "MESSAGE_CORRUPTION" // 25% chance messages become "lol"
	ShockActionLock        ShockType = "ACTION_LOCK"        // Cannot use role ability
	ShockForcedSilence     ShockType = "FORCED_SILENCE"     // Cannot speak during day
)

// CorporateMandate represents game-wide rule modifiers
type CorporateMandate struct {
	Type        MandateType            `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effects     map[string]interface{} `json:"effects"`
	IsActive    bool                   `json:"isActive"`
}

// MandateType represents different corporate mandates
type MandateType string

const (
	MandateAggressiveGrowth MandateType = "AGGRESSIVE_GROWTH"
	MandateTransparency     MandateType = "TOTAL_TRANSPARENCY"
	MandateSecurityLockdown MandateType = "SECURITY_LOCKDOWN"
)

// NightAction represents an action taken during night phase
type NightAction struct {
	Type           NightActionType `json:"type"`
	TargetID       string          `json:"targetId,omitempty"`
	ShadowTargetID string          `json:"shadowTargetId,omitempty"` // For SHADOW action: the final target of the copied ability
}

// NightActionType represents types of night actions
type NightActionType string

const (
	ActionMine        NightActionType = "MINE"
	ActionConvert     NightActionType = "CONVERT"
	ActionBlock       NightActionType = "BLOCK"
	ActionInvestigate NightActionType = "INVESTIGATE"
	ActionProtect     NightActionType = "PROTECT"
	ActionBootcamp    NightActionType = "BOOTCAMP"
	ActionShadow      NightActionType = "SHADOW"
)

// ChatMessage represents a chat message
type ChatMessage struct {
	ID              string                 `json:"id"`
	ClientMessageID string                 `json:"clientMessageID,omitempty"` // Client-generated ID for confirmation
	PlayerID        string                 `json:"playerID"`
	PlayerName      string                 `json:"playerName"`
	Message         string                 `json:"message"`
	Timestamp       time.Time              `json:"timestamp"`
	IsSystem        bool                   `json:"isSystem"`
	Type            string                 `json:"type,omitempty"`       // Message type for system messages (e.g., "PULSE_CHECK", "SITREP")
	ChannelID       string                 `json:"channelID"`            // "#war-room" or "#aligned"
	ReactToID       string                 `json:"reactToID,omitempty"`  // ID of message being reacted to
	Reactions       []EmojiReaction        `json:"reactions"`            // Emoji reactions on this message
	Metadata        map[string]interface{} `json:"metadata,omitempty"`   // Additional data for system messages
}

// EmojiReaction represents an emoji reaction to a message
type EmojiReaction struct {
	Emoji      string    `json:"emoji"`       // The emoji unicode or name
	PlayerID   string    `json:"player_id"`   // Player who reacted
	PlayerName string    `json:"player_name"` // Player name for quick display
	Timestamp  time.Time `json:"timestamp"`   // When the reaction was added
}

// VoteState represents the current voting state
type VoteState struct {
	Type         VoteType          `json:"type"`
	Votes        map[string]string `json:"votes"`        // PlayerID -> TargetID
	TokenWeights map[string]int    `json:"tokenWeights"` // PlayerID -> Token count
	Results      map[string]int    `json:"results"`      // TargetID -> Total tokens
	IsComplete   bool              `json:"isComplete"`
}

// VoteType represents different types of votes
type VoteType string

const (
	VoteExtension  VoteType = "EXTENSION"
	VoteNomination VoteType = "NOMINATION"
	VoteVerdict    VoteType = "VERDICT"
)

// WinCondition represents a game victory condition
type WinCondition struct {
	Winner      string `json:"winner"`    // "HUMANS" or "AI"
	Condition   string `json:"condition"` // "CONTAINMENT" or "SINGULARITY"
	Description string `json:"description"`
}

// GameSettings contains game configuration
type GameSettings struct {
	MaxPlayers               int                    `json:"maxPlayers"`
	MinPlayers               int                    `json:"minPlayers"`
	SitrepDuration           time.Duration          `json:"sitrepDuration"`
	PulseCheckDuration       time.Duration          `json:"pulseCheckDuration"`
	DiscussionDuration       time.Duration          `json:"discussionDuration"`
	ExtensionDuration        time.Duration          `json:"extensionDuration"`
	NominationDuration       time.Duration          `json:"nominationDuration"`
	TrialDuration            time.Duration          `json:"trialDuration"`
	VerdictDuration          time.Duration          `json:"verdictDuration"`
	NightDuration            time.Duration          `json:"nightDuration"`
	StartingTokens           int                    `json:"startingTokens"`
	VotingThreshold          float64                `json:"votingThreshold"`
	InitialAlignedHumanCount int                    `json:"initialAlignedHumanCount"`
	PlayAsAI                 bool                   `json:"playAsAI"`
	CustomSettings           map[string]interface{} `json:"customSettings,omitempty"`
}

// SubmittedNightAction represents an action submitted during the night phase
type SubmittedNightAction struct {
	PlayerID  string                 `json:"playerID"`
	Type      string                 `json:"type"` // "MINE", "BLOCK", "INVESTIGATE", etc.
	TargetID  string                 `json:"targetID"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// NightActionResolutionPayload represents the comprehensive payload for NIGHT_ACTIONS_RESOLVED event
type NightActionResolutionPayload struct {
	Summary              string                                `json:"summary"`
	PlayerStateChanges   map[string]PlayerStateChanges        `json:"player_state_changes"`
	ActionResults        map[string]ActionResult              `json:"action_results"`
	BlockedPlayers       []string                             `json:"blocked_players"`
	ConversionAttempts   []ConversionAttempt                  `json:"conversion_attempts"`
	RoleAbilityUsages    []RoleAbilityUsage                   `json:"role_ability_usages"`
	MiningResults        MiningResults                        `json:"mining_results"`
	PublicAnnouncements  []string                             `json:"public_announcements"`
	PrivateNotifications map[string][]PrivateNotification     `json:"private_notifications"`
}

// PlayerStateChanges represents all changes to a player's state during night resolution
type PlayerStateChanges struct {
	TokensGained        int                    `json:"tokens_gained,omitempty"`
	TokensLost          int                    `json:"tokens_lost,omitempty"`
	StatusMessage       string                 `json:"status_message,omitempty"`
	Alignment           string                 `json:"alignment,omitempty"`
	AIEquity            int                    `json:"ai_equity,omitempty"`
	ProjectMilestones   int                    `json:"project_milestones,omitempty"`
	HasUsedAbility      bool                   `json:"has_used_ability,omitempty"`
	RoleUnlocked        bool                   `json:"role_unlocked,omitempty"`
	SystemShocks        []SystemShock          `json:"system_shocks,omitempty"`
	WasBlocked          bool                   `json:"was_blocked,omitempty"`
	WasTargeted         bool                   `json:"was_targeted,omitempty"`
	ActionCancelled     bool                   `json:"action_cancelled,omitempty"`
	CustomEffects       map[string]interface{} `json:"custom_effects,omitempty"`
}

// ActionResult represents the outcome of a specific night action
type ActionResult struct {
	PlayerID    string                 `json:"player_id"`
	ActionType  string                 `json:"action_type"`
	TargetID    string                 `json:"target_id,omitempty"`
	Success     bool                   `json:"success"`
	BlockedBy   string                 `json:"blocked_by,omitempty"`
	FailReason  string                 `json:"fail_reason,omitempty"`
	Effects     map[string]interface{} `json:"effects,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// ConversionAttempt represents an AI conversion attempt and its outcome
type ConversionAttempt struct {
	AIID           string `json:"ai_id"`
	TargetID       string `json:"target_id"`
	AIEquityBefore int    `json:"ai_equity_before"`
	AIEquityAfter  int    `json:"ai_equity_after"`
	TargetTokens   int    `json:"target_tokens"`
	Success        bool   `json:"success"`
	SystemShock    string `json:"system_shock,omitempty"`
	WasBlocked     bool   `json:"was_blocked,omitempty"`
	BlockedBy      string `json:"blocked_by,omitempty"`
}

// RoleAbilityUsage represents a role ability being used during the night
type RoleAbilityUsage struct {
	PlayerID     string                 `json:"player_id"`
	RoleType     string                 `json:"role_type"`
	AbilityName  string                 `json:"ability_name"`
	TargetID     string                 `json:"target_id,omitempty"`
	Success      bool                   `json:"success"`
	PublicEffect string                 `json:"public_effect,omitempty"`
	WasBlocked   bool                   `json:"was_blocked,omitempty"`
	BlockedBy    string                 `json:"blocked_by,omitempty"`
	Effects      map[string]interface{} `json:"effects,omitempty"`
}

// MiningResults represents the aggregated mining results for the night
type MiningResults struct {
	TotalAttempts    int                    `json:"total_attempts"`
	SuccessfulSlots  int                    `json:"successful_slots"`
	AvailableSlots   int                    `json:"available_slots"`
	LiquidityPool    int                    `json:"liquidity_pool"`
	SuccessfulMiners []MiningAttempt        `json:"successful_miners"`
	FailedMiners     []MiningAttempt        `json:"failed_miners"`
	PriorityRules    map[string]interface{} `json:"priority_rules"`
}

// MiningAttempt represents a single mining attempt
type MiningAttempt struct {
	PlayerID        string `json:"player_id"`
	BeneficiaryID   string `json:"beneficiary_id"`
	TokensAwarded   int    `json:"tokens_awarded"`
	Priority        int    `json:"priority"`
	FailureReason   string `json:"failure_reason,omitempty"`
	WasBlocked      bool   `json:"was_blocked,omitempty"`
	BlockedBy       string `json:"blocked_by,omitempty"`
}

// PrivateNotification represents a private message delivered to specific players
type PrivateNotification struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title,omitempty"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Channel     string                 `json:"channel,omitempty"`
	Urgent      bool                   `json:"urgent,omitempty"`
}

// WhistleblowerVote represents a vote by a deactivated player on crisis options
type WhistleblowerVote struct {
	PlayerID     string    `json:"playerID"`
	PlayerName   string    `json:"playerName"`
	CrisisChoice string    `json:"crisisChoice"` // Crisis type they voted for
	Timestamp    time.Time `json:"timestamp"`
}

// WhistleblowerVoting represents the current whistleblower voting state
type WhistleblowerVoting struct {
	IsActive      bool              `json:"isActive"`
	CrisisOptions []CrisisEventOption `json:"crisisOptions"` // 3 options to choose from
	Votes         map[string]string `json:"votes"`           // PlayerID -> CrisisType
	VoteResults   map[string]int    `json:"voteResults"`     // CrisisType -> Vote count
	SelectedCrisis string           `json:"selectedCrisis"`  // Winning crisis type
	IsComplete    bool              `json:"isComplete"`
}

// CrisisEventOption represents a crisis option for whistleblower voting
type CrisisEventOption struct {
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CrisisEvent represents an active crisis event affecting the game
type CrisisEvent struct {
	Type             string                 `json:"type"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	PulseCheckPrompt string                 `json:"pulseCheckPrompt,omitempty"`
	Effects          map[string]interface{} `json:"effects"`
	Duration         int                    `json:"duration,omitempty"`
	TriggeredAt      time.Time              `json:"triggeredAt,omitempty"`
}

// SitrepSection represents a section of the daily report
type SitrepSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"` // "standard", "classified", "redacted"
}

// DailySitrep represents the complete daily situation report
type DailySitrep struct {
	DayNumber  int             `json:"day_number"`
	Date       time.Time       `json:"date"`
	Sections   []SitrepSection `json:"sections"`
	AlertLevel string          `json:"alert_level"`
	Summary    string          `json:"summary"`
	FooterNote string          `json:"footer_note"`
}

// Spectator represents a spectator observing the game
type Spectator struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	JoinedAt time.Time `json:"joined_at"`
}

// PublicGameState represents a filtered view of the game state for spectators
type PublicGameState struct {
	GameID       string              `json:"game_id"`
	Phase        PhaseType           `json:"phase"`
	DayNumber    int                 `json:"day_number"`
	Players      []PublicPlayerInfo  `json:"players"`
	TokenCounts  map[string]int      `json:"token_counts"`
	PhaseEndTime time.Time           `json:"phase_end_time"`
	CrisisEvent  *CrisisEvent        `json:"crisis_event,omitempty"`
	ChatHistory  []ChatMessage       `json:"chat_history,omitempty"`
}

// PublicPlayerInfo represents public information about a player for spectators
type PublicPlayerInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	JobTitle      string `json:"job_title"`
	IsActive      bool   `json:"is_active"`
	StatusMessage string `json:"status_message"`
	TokenCount    int    `json:"token_count"`
	// Note: No role, alignment, or KPI information for spectators
}

// Event Payload Types - Strongly typed payloads for all events
// This ensures consistent event structures and eliminates defensive parsing

// Game Lifecycle Event Payloads
type GameCreatedPayload struct {
	GameID   string    `json:"game_id"`
	HostID   string    `json:"host_id"`
	Settings GameSettings `json:"settings"`
}

type GameStartedPayload struct {
	GameID string `json:"game_id"`
}

type GameEndedPayload struct {
	GameID      string       `json:"game_id"`
	Winner      string       `json:"winner"`
	Condition   string       `json:"condition"`
	Description string       `json:"description"`
}

type PhaseChangedPayload struct {
	PhaseType string  `json:"phase_type"`
	Duration  float64 `json:"duration"`
}

type DayStartedPayload struct {
	DayNumber int `json:"day_number"`
}

type NightStartedPayload struct {
	DayNumber int `json:"day_number"`
}

// Player Event Payloads
type PlayerJoinedPayload struct {
	Name     string `json:"name"`
	JobTitle string `json:"job_title"`
}

type PlayerLeftPayload struct {
	PlayerName string `json:"player_name"`
}

type PlayerEliminatedPayload struct {
	RoleType  string `json:"role_type"`
	Alignment string `json:"alignment"`
}

type PlayerAbandonedPayload struct {
	RevealedRole string `json:"revealed_role"`
	PlayerName   string `json:"player_name"`
}

type PlayerConnectionStatusChangedPayload struct {
	PlayerID         string `json:"player_id"`
	ConnectionStatus string `json:"connection_status"`
}

type PlayerRoleRevealedPayload struct {
	PlayerID string `json:"player_id"`
}

type PlayerAlignedPayload struct {
	Alignment string `json:"alignment"`
}

type PlayerShockedPayload struct {
	ShockMessage string `json:"shock_message"`
}

type PlayerStatusChangedPayload struct {
	Status string `json:"status"`
}

// Voting Event Payloads
type VoteStartedPayload struct {
	VoteType string `json:"vote_type"`
}

type VoteCastPayload struct {
	TargetID    string `json:"target_id"`
	VoteType    string `json:"vote_type"`
	TokenWeight int    `json:"token_weight"`
}

type VoteTallyUpdatedPayload struct {
	VoteType      string                 `json:"vote_type"`
	Results       map[string]interface{} `json:"results"`
	TokenWeights  map[string]interface{} `json:"token_weights"`
	IsComplete    bool                   `json:"is_complete"`
	VoterID       string                 `json:"voter_id"`
	TargetID      string                 `json:"target_id"`
	PublicVoting  bool                   `json:"public_voting,omitempty"`
	VoterChoices  map[string]interface{} `json:"voter_choices,omitempty"`
}

type VoteCompletedPayload struct {
	VoteType string `json:"vote_type"`
}

type PlayerNominatedPayload struct {
	NominatedPlayer string `json:"nominated_player"`
}

// Token and Mining Event Payloads
type TokensAwardedPayload struct {
	Amount float64 `json:"amount"`
}

type TokensLostPayload struct {
	Amount float64 `json:"amount"`
}

type MiningSuccessfulPayload struct {
	Amount int `json:"amount"`
}

type MiningFailedPayload struct {
	Reason string `json:"reason"`
}

type MiningPoolUpdatedPayload struct {
	Difficulty  float64 `json:"difficulty,omitempty"`
	BaseReward  float64 `json:"base_reward,omitempty"`
}

type TokensDistributedPayload struct {
	Distribution map[string]interface{} `json:"distribution"`
}

// Night Action Event Payloads
type NightActionSubmittedPayload struct {
	ActionType string `json:"action_type"`
	TargetID   string `json:"target_id"`
}

// AI Conversion Event Payloads
type AIConversionAttemptPayload struct {
	TargetID  string  `json:"target_id"`
	AIEquity  float64 `json:"ai_equity"`
}

type AIConversionSuccessPayload struct {
	TargetID string `json:"target_id"`
}

type AIConversionFailedPayload struct {
	ShockMessage string `json:"shock_message"`
}

// Communication Event Payloads
type ChatMessagePayload struct {
	SenderID    string `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	Message     string `json:"message"`
	IsSystem    bool   `json:"isSystem,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	ID          string `json:"id,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type MessageReactionPayload struct {
	MessageID  string `json:"message_id"`
	Emoji      string `json:"emoji"`
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name,omitempty"`
}

type SystemMessagePayload struct {
	Message string `json:"message"`
}

// Crisis and Pulse Check Event Payloads
type CrisisTriggeredPayload struct {
	CrisisType       string                 `json:"crisis_type"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	PulseCheckPrompt string                 `json:"pulse_check_prompt"`
	Effects          map[string]interface{} `json:"effects"`
}

type PulseCheckStartedPayload struct {
	Question string `json:"question"`
}

type PulseCheckUpdatedPayload struct {
	MessageID  string `json:"message_id"`
	Question   string `json:"question"`
	PlayerID   string `json:"player_id"`
	Response   string `json:"response"`
	PlayerName string `json:"player_name"`
}

type PulseCheckRevealedPayload struct {
	PlayerResponses map[string]interface{} `json:"player_responses"`
	TotalResponses  float64                `json:"total_responses"`
	Summary         string                 `json:"summary"`
}

// Role and Ability Event Payloads
type RoleAssignedPayload struct {
	RoleType        string `json:"role_type"`
	RoleName        string `json:"role_name"`
	RoleDescription string `json:"role_description"`
	KPIType         string `json:"kpi_type"`
	KPIDescription  string `json:"kpi_description"`
	Alignment       string `json:"alignment"`
	PersonaName     string `json:"persona_name"`
	JobTitle        string `json:"job_title"`
	LobbyHandle     string `json:"lobby_handle"`
}

type RoleAbilityUnlockedPayload struct {
	AbilityName        string `json:"ability_name"`
	AbilityDescription string `json:"ability_description"`
}

type ProjectMilestonePayload struct {
	Milestone float64 `json:"milestone"`
}

// Specific Role Ability Event Payloads
type RunAuditPayload struct {
	TargetID string `json:"target_id"`
	Result   string `json:"result"`
}

type OverclockServersPayload struct {
	TargetID       string  `json:"target_id"`
	TokensAwarded  float64 `json:"tokens_awarded"`
}

type IsolateNodePayload struct {
	TargetID string `json:"target_id"`
}

type PerformanceReviewPayload struct {
	TargetID     string `json:"target_id"`
	ForcedAction string `json:"forced_action"`
}

type ReallocateBudgetPayload struct {
	FromPlayer string  `json:"from_player"`
	ToPlayer   string  `json:"to_player"`
	Amount     float64 `json:"amount"`
}

type PivotPayload struct {
	SelectedCrisis string `json:"selected_crisis"`
}

type DeployHotfixPayload struct {
	RedactionTarget string `json:"redaction_target"`
}

// Player Blocking and Protection Event Payloads
type PlayerBlockedPayload struct {
	BlockedBy string `json:"blocked_by,omitempty"`
}

type PlayerProtectedPayload struct {
	ProtectedBy string `json:"protected_by,omitempty"`
}

type PlayerInvestigatedPayload struct {
	TargetID string `json:"target_id"`
	Result   string `json:"result"`
}

// Status Event Payloads
type SlackStatusChangedPayload struct {
	Status     string `json:"status"`
	PlayerName string `json:"player_name"`
}

type PartingShotSetPayload struct {
	PartingShot string `json:"parting_shot"`
	PlayerName  string `json:"player_name"`
}

type WhisperSentPayload struct {
	TargetID   string `json:"target_id"`
	DayNumber  int    `json:"day_number"`
}

// KPI Event Payloads
type KPIAssignedPayload struct {
	KPIType     string  `json:"kpi_type"`
	Description string  `json:"description"`
	Target      float64 `json:"target"`
	Reward      string  `json:"reward"`
}

type KPIProgressPayload struct {
	Progress float64 `json:"progress"`
}

// System Shock Event Payloads
type SystemShockAppliedPayload struct {
	ShockType     string  `json:"shock_type"`
	Description   string  `json:"description"`
	DurationHours float64 `json:"duration_hours"`
}

type ShockEffectTriggeredPayload struct {
	EffectType  string `json:"effect_type"`
	Description string `json:"description"`
}

// AI Equity Event Payloads
type AIEquityChangedPayload struct {
	AIEquityChange int `json:"ai_equity_change,omitempty"`
	NewAIEquity    int `json:"new_ai_equity,omitempty"`
}

type EquityThresholdPayload struct {
	Threshold int    `json:"threshold"`
	Action    string `json:"action"`
}

// Corporate Mandate Event Payloads
type MandateActivatedPayload struct {
	MandateType string                 `json:"mandate_type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effects     map[string]interface{} `json:"effects"`
}

type MandateEffectPayload struct {
	Effects map[string]interface{} `json:"effects"`
}

// Phase Skipping Event Payloads
type SkipVoteUpdatedPayload struct {
	CurrentVotes    int      `json:"current_votes"`
	RequiredVotes   int      `json:"required_votes"`
	Voters          []string `json:"voters"`
	HasVoted        bool     `json:"has_voted"`
	PlayerName      string   `json:"player_name"`
	VotingPlayerID  string   `json:"voting_player_id"`
}

// Whistleblower Protocol Event Payloads
type WhistleblowerVotingStartedPayload struct {
	CrisisOptions []interface{} `json:"crisis_options"`
}

type WhistleblowerVoteCastPayload struct {
	CrisisType string `json:"crisis_type"`
}

type WhistleblowerVotingCompletedPayload struct {
	SelectedCrisis string `json:"selected_crisis"`
}

// Semantic Event Payloads
type SitrepPublishedPayload struct {
	DailySitrep map[string]interface{} `json:"daily_sitrep"`
}

type LiaisonProtocolActivatedPayload struct {
	AIPercentage      float64 `json:"ai_percentage"`
	MiningBonusSlots  float64 `json:"mining_bonus_slots"`
}

type GameRuleModifiedPayload struct {
	RuleCategory     string `json:"rule_category"`
	ModificationType string `json:"modification_type"`
	Source           string `json:"source"`
}

// Game State Event Payloads
type GameStateUpdatePayload struct {
	GameState interface{} `json:"game_state"`
}

type LobbyStateUpdatePayload struct {
	LobbyID    string      `json:"lobby_id"`
	Players    interface{} `json:"players"`
	HostID     string      `json:"host_id"`
	CanStart   bool        `json:"can_start"`
	Name       string      `json:"name"`
	MaxPlayers int         `json:"max_players"`
}

type VictoryConditionPayload struct {
	Winner      string `json:"winner"`
	Condition   string `json:"condition"`
	Description string `json:"description"`
}