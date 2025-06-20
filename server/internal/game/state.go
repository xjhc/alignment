package game

import (
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
			MinPlayers:         6,
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

// getCurrentTime returns the current time (helper function)
func getCurrentTime() time.Time {
	return time.Now()
}
