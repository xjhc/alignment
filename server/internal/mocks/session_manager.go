package mocks

import (
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockSessionManager is a mock implementation of the SessionManagerInterface
type MockSessionManager struct {
	JoinGameCalls             []JoinGameCall
	LeaveGameCalls            []LeaveGameCall
	SendActionToGameCalls     []SendActionToGameCall
	CreateGameFromLobbyCalls  []CreateGameFromLobbyCall

	JoinGameResults           []error
	LeaveGameResults          []error
	SendActionToGameResults   []error
	CreateGameFromLobbyResults []error
}

type JoinGameCall struct {
	GameID string
	Player interfaces.PlayerActorInterface
}

type LeaveGameCall struct {
	GameID   string
	PlayerID string
}

type SendActionToGameCall struct {
	GameID string
	Action core.Action
}

type CreateGameFromLobbyCall struct {
	LobbyID      string
	PlayerActors map[string]interfaces.PlayerActorInterface
}

// Ensure MockSessionManager implements the interface at compile time
var _ interfaces.SessionManagerInterface = (*MockSessionManager)(nil)

func (m *MockSessionManager) JoinGame(gameID string, player interfaces.PlayerActorInterface) error {
	m.JoinGameCalls = append(m.JoinGameCalls, JoinGameCall{
		GameID: gameID,
		Player: player,
	})

	if len(m.JoinGameResults) > 0 {
		result := m.JoinGameResults[0]
		if len(m.JoinGameResults) > 1 {
			m.JoinGameResults = m.JoinGameResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockSessionManager) LeaveGame(gameID string, playerID string) error {
	m.LeaveGameCalls = append(m.LeaveGameCalls, LeaveGameCall{
		GameID:   gameID,
		PlayerID: playerID,
	})

	if len(m.LeaveGameResults) > 0 {
		result := m.LeaveGameResults[0]
		if len(m.LeaveGameResults) > 1 {
			m.LeaveGameResults = m.LeaveGameResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockSessionManager) SendActionToGame(gameID string, action core.Action) error {
	m.SendActionToGameCalls = append(m.SendActionToGameCalls, SendActionToGameCall{
		GameID: gameID,
		Action: action,
	})

	if len(m.SendActionToGameResults) > 0 {
		result := m.SendActionToGameResults[0]
		if len(m.SendActionToGameResults) > 1 {
			m.SendActionToGameResults = m.SendActionToGameResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockSessionManager) CreateGameFromLobby(lobbyID string, playerActors map[string]interfaces.PlayerActorInterface) error {
	m.CreateGameFromLobbyCalls = append(m.CreateGameFromLobbyCalls, CreateGameFromLobbyCall{
		LobbyID:      lobbyID,
		PlayerActors: playerActors,
	})

	if len(m.CreateGameFromLobbyResults) > 0 {
		result := m.CreateGameFromLobbyResults[0]
		if len(m.CreateGameFromLobbyResults) > 1 {
			m.CreateGameFromLobbyResults = m.CreateGameFromLobbyResults[1:]
		}
		return result
	}

	return nil
}