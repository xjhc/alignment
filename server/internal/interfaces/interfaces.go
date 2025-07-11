package interfaces

import (
	"context"
	"time"
	"github.com/xjhc/alignment/core"
)

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
	Stop() // Add Stop method for proper cleanup
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
	CreateLobbyViaHTTP(userID, hostPlayerName, lobbyName, playerAvatar string, isPrivate bool) (string, string, string, error)
	JoinLobby(lobbyID, userID, playerName, playerAvatar string) (string, string, error)
	JoinLobbyWithActor(lobbyID string, playerActor PlayerActorInterface) error
	StartGame(lobbyID string, hostPlayerID string) error
	ValidateSessionToken(token string) (interface{}, error)
	GetLobbyList() []interface{}
	
	// Game session management
	SendActionToGame(gameID string, action core.Action) (chan ProcessActionResult, error)
	BroadcastEventsToGame(gameID string, events []core.Event) error
	GetGameActor(gameID string) (GameActorInterface, bool)
	ReconnectPlayerToGame(gameID string, playerActor PlayerActorInterface) error
	
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
	AppendEvent(ctx context.Context, gameID string, event core.Event) error
	GetEvents(ctx context.Context, gameID string) ([]core.Event, error)
	GetEventsSince(ctx context.Context, gameID string, timestamp string) ([]core.Event, error)
	LoadEvents(ctx context.Context, gameID string, afterSequence int) ([]core.Event, error)
	CreateSnapshot(ctx context.Context, gameID string, state core.GameState) error
	GetLatestSnapshot(ctx context.Context, gameID string) (*core.GameState, error)
	ListActiveGames(ctx context.Context) ([]string, error)
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

// PartyInvite represents an invitation to join a party
type PartyInvite struct {
	ID         string    `json:"id"`
	PartyID    string    `json:"partyId"`
	InviterID  string    `json:"inviterId"`
	InviteeID  string    `json:"inviteeId"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

// PartyManagerInterface defines the interface for party management
type PartyManagerInterface interface {
	CreateParty(leader PlayerActorInterface) (string, error)
	InviteToPartyByInviter(inviterID, inviteeID string) (*PartyInvite, error)
	JoinParty(playerID, inviteID string, playerActor PlayerActorInterface) error
	LeaveParty(playerID string) error
	GetPendingInvites(playerID string) []*PartyInvite
	IsInParty(playerID string) bool
}

// ForceLogoutInterface defines the interface for forcing player logout
type ForceLogoutInterface interface {
	ForceLogoutPlayer(playerID string, reason string)
}
