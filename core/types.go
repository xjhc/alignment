package core

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

	// Player Status events
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

// Phase represents the current game phase
type Phase struct {
	Type      PhaseType     `json:"type"`
	StartTime time.Time     `json:"start_time"`
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

// Player represents a human player
type Player struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	JobTitle          string    `json:"job_title"`
	IsAlive           bool      `json:"is_alive"`
	Tokens            int       `json:"tokens"`
	ProjectMilestones int       `json:"project_milestones"`
	StatusMessage     string    `json:"status_message"`
	JoinedAt          time.Time `json:"joined_at"`

	// Private fields (only visible to the player themselves)
	Alignment       string       `json:"alignment,omitempty"` // "HUMAN" or "ALIGNED"
	Role            *Role        `json:"role,omitempty"`
	PersonalKPI     *PersonalKPI `json:"personal_kpi,omitempty"`
	AIEquity        int          `json:"ai_equity,omitempty"` // For alignment conversion
	HasUsedAbility  bool         `json:"has_used_ability,omitempty"`
	LastNightAction *NightAction `json:"last_night_action,omitempty"`

	// Public status and effects
	SlackStatus  string        `json:"slack_status,omitempty"`
	PartingShot  string        `json:"parting_shot,omitempty"`
	SystemShocks []SystemShock `json:"system_shocks,omitempty"`
}

// Role represents a player's role and abilities
type Role struct {
	Type        RoleType `json:"type"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsUnlocked  bool     `json:"is_unlocked"`
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
	IsReady     bool   `json:"is_ready"`
}

// PersonalKPI represents a player's secret objective
type PersonalKPI struct {
	Type        KPIType `json:"type"`
	Description string  `json:"description"`
	Progress    int     `json:"progress"`
	Target      int     `json:"target"`
	IsCompleted bool    `json:"is_completed"`
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
	ExpiresAt   time.Time `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
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
	IsActive    bool                   `json:"is_active"`
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
	Type     NightActionType `json:"type"`
	TargetID string          `json:"target_id,omitempty"`
}

// NightActionType represents types of night actions
type NightActionType string

const (
	ActionMine        NightActionType = "MINE"
	ActionConvert     NightActionType = "CONVERT"
	ActionBlock       NightActionType = "BLOCK"
	ActionInvestigate NightActionType = "INVESTIGATE"
	ActionProtect     NightActionType = "PROTECT"
)

// CrisisEvent represents a daily crisis that affects game rules
type CrisisEvent struct {
	Type        string                 `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Effects     map[string]interface{} `json:"effects"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID         string    `json:"id"`
	PlayerID   string    `json:"player_id"`
	PlayerName string    `json:"player_name"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	IsSystem   bool      `json:"is_system"`
}

// VoteState represents the current voting state
type VoteState struct {
	Type         VoteType          `json:"type"`
	Votes        map[string]string `json:"votes"`         // PlayerID -> TargetID
	TokenWeights map[string]int    `json:"token_weights"` // PlayerID -> Token count
	Results      map[string]int    `json:"results"`       // TargetID -> Total tokens
	IsComplete   bool              `json:"is_complete"`
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
	MaxPlayers         int           `json:"max_players"`
	MinPlayers         int           `json:"min_players"`
	SitrepDuration     time.Duration `json:"sitrep_duration"`
	PulseCheckDuration time.Duration `json:"pulse_check_duration"`
	DiscussionDuration time.Duration `json:"discussion_duration"`
	ExtensionDuration  time.Duration `json:"extension_duration"`
	NominationDuration time.Duration `json:"nomination_duration"`
	TrialDuration      time.Duration `json:"trial_duration"`
	VerdictDuration    time.Duration `json:"verdict_duration"`
	NightDuration      time.Duration `json:"night_duration"`
	StartingTokens     int           `json:"starting_tokens"`
	VotingThreshold    float64       `json:"voting_threshold"`
}

// SubmittedNightAction represents an action submitted during the night phase
type SubmittedNightAction struct {
	PlayerID  string                 `json:"player_id"`
	Type      string                 `json:"type"` // "MINE", "BLOCK", "INVESTIGATE", etc.
	TargetID  string                 `json:"target_id"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}