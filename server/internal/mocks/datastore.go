package mocks

import (
	"context"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockDataStore is a mock implementation of the DataStore interface
type MockDataStore struct {
	AppendEventCalls          []AppendEventCall
	GetEventsCalls           []GetEventsCall
	GetEventsSinceCalls      []GetEventsSinceCall
	LoadEventsCalls          []LoadEventsCall
	CreateSnapshotCalls      []CreateSnapshotCall
	GetLatestSnapshotCalls   []GetLatestSnapshotCall
	ListActiveGamesCalls     []ListActiveGamesCall
	CloseCalls               []CloseCall

	AppendEventResults       []error
	GetEventsResults         []GetEventsResult
	GetEventsSinceResults    []GetEventsSinceResult
	LoadEventsResults        []LoadEventsResult
	CreateSnapshotResults    []error
	GetLatestSnapshotResults []GetLatestSnapshotResult
	ListActiveGamesResults   []ListActiveGamesResult
	CloseResults             []error
}

type AppendEventCall struct {
	GameID string
	Event  core.Event
}

type GetEventsCall struct {
	GameID string
}

type GetEventsResult struct {
	Events []core.Event
	Error  error
}

type GetEventsSinceCall struct {
	GameID    string
	Timestamp string
}

type GetEventsSinceResult struct {
	Events []core.Event
	Error  error
}

type LoadEventsCall struct {
	GameID        string
	AfterSequence int
}

type LoadEventsResult struct {
	Events []core.Event
	Error  error
}

type CreateSnapshotCall struct {
	GameID string
	State  core.GameState
}

type GetLatestSnapshotCall struct {
	GameID string
}

type GetLatestSnapshotResult struct {
	State *core.GameState
	Error error
}

type ListActiveGamesCall struct{}

type ListActiveGamesResult struct {
	GameIDs []string
	Error   error
}

type CloseCall struct{}

// Ensure MockDataStore implements the interface at compile time
var _ interfaces.DataStore = (*MockDataStore)(nil)

func (m *MockDataStore) AppendEvent(ctx context.Context, gameID string, event core.Event) error {
	m.AppendEventCalls = append(m.AppendEventCalls, AppendEventCall{
		GameID: gameID,
		Event:  event,
	})

	if len(m.AppendEventResults) > 0 {
		result := m.AppendEventResults[0]
		if len(m.AppendEventResults) > 1 {
			m.AppendEventResults = m.AppendEventResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockDataStore) GetEvents(ctx context.Context, gameID string) ([]core.Event, error) {
	m.GetEventsCalls = append(m.GetEventsCalls, GetEventsCall{
		GameID: gameID,
	})

	if len(m.GetEventsResults) > 0 {
		result := m.GetEventsResults[0]
		if len(m.GetEventsResults) > 1 {
			m.GetEventsResults = m.GetEventsResults[1:]
		}
		return result.Events, result.Error
	}

	return nil, nil
}

func (m *MockDataStore) GetEventsSince(ctx context.Context, gameID string, timestamp string) ([]core.Event, error) {
	m.GetEventsSinceCalls = append(m.GetEventsSinceCalls, GetEventsSinceCall{
		GameID:    gameID,
		Timestamp: timestamp,
	})

	if len(m.GetEventsSinceResults) > 0 {
		result := m.GetEventsSinceResults[0]
		if len(m.GetEventsSinceResults) > 1 {
			m.GetEventsSinceResults = m.GetEventsSinceResults[1:]
		}
		return result.Events, result.Error
	}

	return nil, nil
}

func (m *MockDataStore) LoadEvents(ctx context.Context, gameID string, afterSequence int) ([]core.Event, error) {
	m.LoadEventsCalls = append(m.LoadEventsCalls, LoadEventsCall{
		GameID:        gameID,
		AfterSequence: afterSequence,
	})

	if len(m.LoadEventsResults) > 0 {
		result := m.LoadEventsResults[0]
		if len(m.LoadEventsResults) > 1 {
			m.LoadEventsResults = m.LoadEventsResults[1:]
		}
		return result.Events, result.Error
	}

	return nil, nil
}

func (m *MockDataStore) CreateSnapshot(ctx context.Context, gameID string, state core.GameState) error {
	m.CreateSnapshotCalls = append(m.CreateSnapshotCalls, CreateSnapshotCall{
		GameID: gameID,
		State:  state,
	})

	if len(m.CreateSnapshotResults) > 0 {
		result := m.CreateSnapshotResults[0]
		if len(m.CreateSnapshotResults) > 1 {
			m.CreateSnapshotResults = m.CreateSnapshotResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockDataStore) GetLatestSnapshot(ctx context.Context, gameID string) (*core.GameState, error) {
	m.GetLatestSnapshotCalls = append(m.GetLatestSnapshotCalls, GetLatestSnapshotCall{
		GameID: gameID,
	})

	if len(m.GetLatestSnapshotResults) > 0 {
		result := m.GetLatestSnapshotResults[0]
		if len(m.GetLatestSnapshotResults) > 1 {
			m.GetLatestSnapshotResults = m.GetLatestSnapshotResults[1:]
		}
		return result.State, result.Error
	}

	return nil, nil
}

func (m *MockDataStore) ListActiveGames(ctx context.Context) ([]string, error) {
	m.ListActiveGamesCalls = append(m.ListActiveGamesCalls, ListActiveGamesCall{})

	if len(m.ListActiveGamesResults) > 0 {
		result := m.ListActiveGamesResults[0]
		if len(m.ListActiveGamesResults) > 1 {
			m.ListActiveGamesResults = m.ListActiveGamesResults[1:]
		}
		return result.GameIDs, result.Error
	}

	return nil, nil
}

func (m *MockDataStore) Close() error {
	m.CloseCalls = append(m.CloseCalls, CloseCall{})

	if len(m.CloseResults) > 0 {
		result := m.CloseResults[0]
		if len(m.CloseResults) > 1 {
			m.CloseResults = m.CloseResults[1:]
		}
		return result
	}

	return nil
}