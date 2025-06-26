package core

// EventTypeValues exports all EventType constants as an array for validation and generation
var EventTypeValues = []EventType{
	// Game lifecycle events
	EventGameCreated,
	EventGameStarted,
	EventGameEnded,
	EventPhaseChanged,

	// Player events
	EventPlayerJoined,
	EventPlayerLeft,
	EventPlayerEliminated,
	EventPlayerRoleRevealed,
	EventPlayerAligned,
	EventPlayerShocked,

	// Voting events
	EventVoteStarted,
	EventVoteCast,
	EventVoteTallyUpdated,
	EventVoteCompleted,
	EventPlayerNominated,

	// Token and Mining events
	EventTokensAwarded,
	EventTokensSpent,
	EventMiningAttempted,
	EventMiningSuccessful,
	EventMiningFailed,

	// Night Action events
	EventNightActionsResolved,
	EventPlayerBlocked,
	EventPlayerProtected,
	EventPlayerInvestigated,

	// AI and Conversion events
	EventAIConversionAttempt,
	EventAIConversionSuccess,
	EventAIConversionFailed,
	EventAIRevealed,

	// Communication events
	EventChatMessage,
	EventSystemMessage,
	EventPrivateNotification,

	// Crisis and Special events
	EventCrisisTriggered,
	EventPulseCheckStarted,
	EventPulseCheckSubmitted,
	EventPulseCheckRevealed,
	EventRoleAbilityUnlocked,
	EventProjectMilestone,
	EventRoleAssigned,

	// Mining and Economy events
	EventMiningPoolUpdated,
	EventTokensDistributed,
	EventTokensLost,

	// Day/Night transition events
	EventDayStarted,
	EventNightStarted,
	EventNightActionSubmitted,
	EventAllPlayersReady,

	// Status and State events
	EventPlayerStatusChanged,
	EventGameStateSnapshot,
	EventGameStateUpdate,
	EventLobbyStateUpdate,
	EventClientIdentified,
	EventChatHistorySnapshot,
	EventPlayerReconnected,
	EventPlayerDisconnected,
	EventSyncComplete,

	// Win Condition events
	EventVictoryCondition,

	// Role Ability events
	EventRunAudit,
	EventOverclockServers,
	EventIsolateNode,
	EventPerformanceReview,
	EventReallocateBudget,
	EventPivot,
	EventDeployHotfix,

	// Player Status events
	EventSlackStatusChanged,
	EventPartingShotSet,

	// Personal KPI events
	EventKPIAssigned,
	EventKPIProgress,
	EventKPICompleted,

	// Corporate Mandate events
	EventMandateActivated,
	EventMandateEffect,

	// System Shock events
	EventSystemShockApplied,
	EventShockEffectTriggered,

	// AI Equity events
	EventAIEquityChanged,
	EventEquityThreshold,
}

// ActionTypeValues exports all ActionType constants as an array for validation and generation
var ActionTypeValues = []ActionType{
	// Lobby actions
	ActionCreateGame,
	ActionJoinGame,
	ActionLeaveGame,
	ActionStartGame,

	// Communication actions
	ActionSendMessage,
	ActionSubmitPulseCheck,

	// Voting actions
	ActionSubmitVote,
	ActionExtendDiscussion,

	// Night actions
	ActionSubmitNightAction,
	ActionMineTokens,
	ActionUseAbility,
	ActionAttemptConversion,
	ActionProjectMilestones,

	// Role-specific abilities
	ActionRunAudit,
	ActionOverclockServers,
	ActionIsolateNode,
	ActionPerformanceReview,
	ActionReallocateBudget,
	ActionPivot,
	ActionDeployHotfix,

	// Status actions
	ActionSetSlackStatus,

	// Meta actions
	ActionReconnect,
}

// PhaseTypeValues exports all PhaseType constants as an array for validation and generation
var PhaseTypeValues = []PhaseType{
	PhaseLobby,
	PhaseSitrep,
	PhasePulseCheck,
	PhaseDiscussion,
	PhaseExtension,
	PhaseNomination,
	PhaseTrial,
	PhaseVerdict,
	PhaseNight,
	PhaseGameOver,
}

// RoleTypeValues exports all RoleType constants as an array for validation and generation
var RoleTypeValues = []RoleType{
	RoleCISO,
	RoleCEO,
	RoleCTO,
	RoleCOO,
	RoleCFO,
	RoleEthics,
	RolePlatforms,
	RoleIntern,
}

// KPITypeValues exports all KPIType constants as an array for validation and generation
var KPITypeValues = []KPIType{
	KPICapitalist,
	KPIGuardian,
	KPIInquisitor,
	KPISuccessionPlanner,
	KPIScapegoat,
}

// ShockTypeValues exports all ShockType constants as an array for validation and generation
var ShockTypeValues = []ShockType{
	ShockMessageCorruption,
	ShockActionLock,
	ShockForcedSilence,
}

// MandateTypeValues exports all MandateType constants as an array for validation and generation
var MandateTypeValues = []MandateType{
	MandateAggressiveGrowth,
	MandateTransparency,
	MandateSecurityLockdown,
}

// NightActionTypeValues exports all NightActionType constants as an array for validation and generation
var NightActionTypeValues = []NightActionType{
	ActionMine,
	ActionConvert,
	ActionBlock,
	ActionInvestigate,
	ActionProtect,
}

// VoteTypeValues exports all VoteType constants as an array for validation and generation
var VoteTypeValues = []VoteType{
	VoteExtension,
	VoteNomination,
	VoteVerdict,
}