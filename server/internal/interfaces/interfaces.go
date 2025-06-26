package interfaces

import "github.com/xjhc/alignment/core"

// PlayerActorInterface defines the interface for PlayerActor
type PlayerActorInterface interface {
	GetPlayerID() string
	GetPlayerName() string
	GetPlayerAvatar() string
	GetSessionToken() string
	GetState() PlayerState
	TransitionToLobby(lobbyID string) error
	TransitionToGame(gameID string) error
	TransitionToIdle() error
	SendServerMessage(message interface{})
}

// PlayerState represents the current state of a player in the system
type PlayerState int

const (
	StateIdle PlayerState = iota
	StateInLobby
	StateInGame
)

func (ps PlayerState) String() string {
	switch ps {
	case StateIdle:
		return "Idle"
	case StateInLobby:
		return "InLobby"
	case StateInGame:
		return "InGame"
	default:
		return "Unknown"
	}
}

// GameActorInterface defines the interface for GameActor
type GameActorInterface interface {
	GetGameID() string
	PostAction(action core.Action) chan ProcessActionResult
	GetGameState() *core.GameState
	CreatePlayerStateUpdateEvent(playerID string) core.Event
	Stop()
}

// ProcessActionResult contains the result of processing an action
type ProcessActionResult struct {
	Events []core.Event
	Error  error
}

// LobbyManagerInterface defines the interface for lobby management
type LobbyManagerInterface interface {
	JoinLobby(gameID, playerName, playerAvatar string) (string, string, error)
	JoinLobbyWithActor(lobbyID string, player PlayerActorInterface) error
	LeaveLobby(lobbyID string, playerID string) error
	StartGame(lobbyID string, hostPlayerID string) error
	CreateLobby(hostPlayer PlayerActorInterface, lobbyName string) (string, error)
	GetLobbyList() []interface{}
	GetLobby(lobbyID string) (interface{}, bool)
	ValidateSession(gameID, playerID, sessionToken string) bool
	GetPlayerInfo(gameID, playerID string) (string, string, error)
}

// SessionManagerInterface defines the interface for session management
type SessionManagerInterface interface {
	JoinGame(gameID string, player PlayerActorInterface) error
	LeaveGame(gameID string, playerID string) error
	SendActionToGame(gameID string, action core.Action) error
	CreateGameFromLobby(lobbyID string, playerActors map[string]PlayerActorInterface) error
}

// GameLifecycleManagerInterface unifies lobby and session management
type GameLifecycleManagerInterface interface {
	// Lobby management
	CreateLobbyViaHTTP(hostPlayerName, lobbyName, playerAvatar string) (string, string, string, error)
	JoinLobby(lobbyID, playerName, playerAvatar string) (string, string, error)
	JoinLobbyWithActor(lobbyID string, playerActor PlayerActorInterface) error
	StartGame(lobbyID string, hostPlayerID string) error
	ValidateSessionToken(token string) (interface{}, error)
	
	// Game session management
	SendActionToGame(gameID string, action core.Action) error
	
	// Utility
	Stop()
}

// SupervisorInterface manages GameActors
type SupervisorInterface interface {
	CreateGameWithPlayers(gameID string, players map[string]*core.Player) (GameActorInterface, error)
	GetActor(gameID string) (GameActorInterface, bool)
	RemoveGame(gameID string)
}

// DataStore interface for persistence
type DataStore interface {
	AppendEvent(gameID string, event core.Event) error
	GetEvents(gameID string) ([]core.Event, error)
	GetEventsSince(gameID string, timestamp string) ([]core.Event, error)
	LoadEvents(gameID string, afterSequence int) ([]core.Event, error)
	CreateSnapshot(gameID string, state core.GameState) error
	GetLatestSnapshot(gameID string) (*core.GameState, error)
	Close() error
}

// Broadcaster interface for sending events
type Broadcaster interface {
	BroadcastToGame(gameID string, event core.Event) error
	SendToPlayer(gameID, playerID string, event core.Event) error
}

type GameStateSnapshot struct {
	GameID    string
	GameState interface{}
}

type TransitionToGame struct {
	GameID    string
	GameState interface{}
}
