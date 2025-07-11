package mocks

import (
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockGameActor is a mock implementation of the GameActorInterface
type MockGameActor struct {
	gameID string
}

// NewMockGameActor creates a new MockGameActor with the given gameID
func NewMockGameActor(gameID string) *MockGameActor {
	return &MockGameActor{gameID: gameID}
}

// Ensure MockGameActor implements the interface at compile time
var _ interfaces.GameActorInterface = (*MockGameActor)(nil)

func (m *MockGameActor) GetGameID() string {
	return m.gameID
}

func (m *MockGameActor) PostAction(action core.Action) chan interfaces.ProcessActionResult {
	resultChan := make(chan interfaces.ProcessActionResult, 1)
	resultChan <- interfaces.ProcessActionResult{
		Events: []core.Event{},
		Error:  nil,
	}
	return resultChan
}

func (m *MockGameActor) GetGameState() *core.GameState {
	return &core.GameState{
		ID: m.gameID,
	}
}

func (m *MockGameActor) CreatePlayerStateUpdateEvent(playerID string) core.Event {
	return core.Event{
		ID:      "test-event",
		Type:    "PLAYER_STATE_UPDATE",
		GameID:  m.gameID,
		PlayerID: playerID,
	}
}

func (m *MockGameActor) Stop() {
	// Mock implementation - no-op
}