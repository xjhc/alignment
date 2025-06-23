package mocks

import (
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockSupervisor is a mock implementation of the SupervisorInterface
type MockSupervisor struct {
	CreateGameWithPlayersCalls []CreateGameWithPlayersCall
	GetActorCalls              []GetActorCall
	RemoveGameCalls            []RemoveGameCall

	CreateGameWithPlayersResults []CreateGameWithPlayersResult
	GetActorResults              []GetActorResult
	RemoveGameResults            []interface{} // No return value, but keep for consistency
}

type CreateGameWithPlayersCall struct {
	GameID  string
	Players map[string]*core.Player
}

type CreateGameWithPlayersResult struct {
	Actor interfaces.GameActorInterface
	Error error
}

type GetActorCall struct {
	GameID string
}

type GetActorResult struct {
	Actor interfaces.GameActorInterface
	Found bool
}

type RemoveGameCall struct {
	GameID string
}

// Ensure MockSupervisor implements the interface at compile time
var _ interfaces.SupervisorInterface = (*MockSupervisor)(nil)

func (m *MockSupervisor) CreateGameWithPlayers(gameID string, players map[string]*core.Player) (interfaces.GameActorInterface, error) {
	m.CreateGameWithPlayersCalls = append(m.CreateGameWithPlayersCalls, CreateGameWithPlayersCall{
		GameID:  gameID,
		Players: players,
	})

	if len(m.CreateGameWithPlayersResults) > 0 {
		result := m.CreateGameWithPlayersResults[0]
		if len(m.CreateGameWithPlayersResults) > 1 {
			m.CreateGameWithPlayersResults = m.CreateGameWithPlayersResults[1:]
		}
		return result.Actor, result.Error
	}

	return nil, nil
}

func (m *MockSupervisor) GetActor(gameID string) (interfaces.GameActorInterface, bool) {
	m.GetActorCalls = append(m.GetActorCalls, GetActorCall{
		GameID: gameID,
	})

	if len(m.GetActorResults) > 0 {
		result := m.GetActorResults[0]
		if len(m.GetActorResults) > 1 {
			m.GetActorResults = m.GetActorResults[1:]
		}
		return result.Actor, result.Found
	}

	return nil, false
}

func (m *MockSupervisor) RemoveGame(gameID string) {
	m.RemoveGameCalls = append(m.RemoveGameCalls, RemoveGameCall{
		GameID: gameID,
	})

	// No return value to handle
}