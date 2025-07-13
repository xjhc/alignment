package test_helpers

import (
	"context"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/lifecycle"
	"github.com/xjhc/alignment/server/internal/mocks"
)

// CreateTestSessionManager creates a session manager with mock dependencies for testing
func CreateTestSessionManager() (*lifecycle.GameLifecycleManager, *mocks.MockDataStore, *mocks.MockSupervisor, *mocks.MockBroadcaster) {
	ctx := context.Background()
	
	mockDatastore := &mocks.MockDataStore{}
	mockSupervisor := &mocks.MockSupervisor{}
	mockBroadcaster := &mocks.MockBroadcaster{}
	eventBus := events.NewEventBus()
	
	// Create a mock active user tracker
	mockActiveUserTracker := &MockActiveUserTracker{}
	
	// Initialize GameLifecycleManager
	glm := lifecycle.NewGameLifecycleManager(ctx, mockDatastore, mockBroadcaster, mockSupervisor, eventBus, nil, mockActiveUserTracker)
	
	return glm, mockDatastore, mockSupervisor, mockBroadcaster
}

// MockActiveUserTracker is a mock implementation for testing
type MockActiveUserTracker struct{}

func (m *MockActiveUserTracker) RemoveUserSessionByGameAndPlayer(gameID, playerID string) {
	// No-op for testing
}