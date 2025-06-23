package mocks

import (
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockBroadcaster is a mock implementation of the Broadcaster interface
type MockBroadcaster struct {
	BroadcastToGameCalls []BroadcastToGameCall
	SendToPlayerCalls    []SendToPlayerCall

	BroadcastToGameResults []error
	SendToPlayerResults    []error
}

type BroadcastToGameCall struct {
	GameID string
	Event  core.Event
}

type SendToPlayerCall struct {
	GameID   string
	PlayerID string
	Event    core.Event
}

// Ensure MockBroadcaster implements the interface at compile time
var _ interfaces.Broadcaster = (*MockBroadcaster)(nil)

func (m *MockBroadcaster) BroadcastToGame(gameID string, event core.Event) error {
	m.BroadcastToGameCalls = append(m.BroadcastToGameCalls, BroadcastToGameCall{
		GameID: gameID,
		Event:  event,
	})

	if len(m.BroadcastToGameResults) > 0 {
		result := m.BroadcastToGameResults[0]
		if len(m.BroadcastToGameResults) > 1 {
			m.BroadcastToGameResults = m.BroadcastToGameResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockBroadcaster) SendToPlayer(gameID, playerID string, event core.Event) error {
	m.SendToPlayerCalls = append(m.SendToPlayerCalls, SendToPlayerCall{
		GameID:   gameID,
		PlayerID: playerID,
		Event:    event,
	})

	if len(m.SendToPlayerResults) > 0 {
		result := m.SendToPlayerResults[0]
		if len(m.SendToPlayerResults) > 1 {
			m.SendToPlayerResults = m.SendToPlayerResults[1:]
		}
		return result
	}

	return nil
}