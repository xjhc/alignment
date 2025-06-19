package game

import (
	"time"
)

// GameState represents the complete state of a game
type GameState struct {
	ID          string                 `json:"id"`
	Phase       Phase                  `json:"phase"`
	Turn        int                    `json:"turn"`
	Players     map[string]*Player     `json:"players"`
	AIPlayer    *AIPlayer              `json:"ai_player,omitempty"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Settings    GameSettings          `json:"settings"`
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
	PhaseSetup     PhaseType = "setup"
	PhaseDay       PhaseType = "day"
	PhaseVoting    PhaseType = "voting"
	PhaseNight     PhaseType = "night"
	PhaseGameOver  PhaseType = "game_over"
)

// Player represents a human player
type Player struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Tokens   int       `json:"tokens"`
	IsActive bool      `json:"is_active"`
	Role     *Role     `json:"role,omitempty"`
	JoinedAt time.Time `json:"joined_at"`
}

// AIPlayer represents the AI player
type AIPlayer struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Tokens   int    `json:"tokens"`
	IsActive bool   `json:"is_active"`
}

// Role represents a player's role and abilities
type Role struct {
	Type        RoleType `json:"type"`
	Description string   `json:"description"`
	IsUnlocked  bool     `json:"is_unlocked"`
}

// RoleType represents different player roles
type RoleType string

const (
	RoleEmployee    RoleType = "employee"
	RoleManager     RoleType = "manager"
	RoleExecutive   RoleType = "executive"
	RoleWhistleblower RoleType = "whistleblower"
)

// GameSettings contains game configuration
type GameSettings struct {
	MaxPlayers     int           `json:"max_players"`
	PhaseTimeout   time.Duration `json:"phase_timeout"`
	TokensToWin    int           `json:"tokens_to_win"`
	VotingThreshold float64      `json:"voting_threshold"`
}

// NewGameState creates a new game state
func NewGameState(id string) *GameState {
	now := time.Now()
	return &GameState{
		ID:        id,
		Phase:     Phase{Type: PhaseSetup, StartTime: now, Duration: time.Minute * 5},
		Turn:      0,
		Players:   make(map[string]*Player),
		CreatedAt: now,
		UpdatedAt: now,
		Settings: GameSettings{
			MaxPlayers:      8,
			PhaseTimeout:    time.Minute * 2,
			TokensToWin:     10,
			VotingThreshold: 0.5,
		},
	}
}