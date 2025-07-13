package game

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjhc/alignment/core"
)

// MockActionAcker implements ActionAcker for testing
type MockActionAcker struct {
	GameID string
}

func (m *MockActionAcker) GetGameID() string {
	return m.GameID
}

func (m *MockActionAcker) ValidateActionPayloadSize(action core.Action) error {
	// Mock implementation
	return nil
}

func (m *MockActionAcker) GetPlayer(id string) (*core.Player, bool) {
	// Mock implementation
	if id == "player-a" {
		return &core.Player{ID: id, Name: "Alice", IsAlive: true}, true
	}
	if id == "player-b" {
		return &core.Player{ID: id, Name: "Bob", IsAlive: true}, true
	}
	return nil, false
}

func (m *MockActionAcker) GetRandom() *rand.Rand {
	// Mock implementation - not used in this test
	return rand.New(rand.NewSource(1))
}

func (m *MockActionAcker) BroadcastToSpectators(events []core.Event) {
	// Mock implementation
}

func TestChatHandler_HandleReaction_CorrectPlayerAttribution(t *testing.T) {
	// Arrange
	state := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"player-a": {ID: "player-a", Name: "Alice", IsAlive: true},
			"player-b": {ID: "player-b", Name: "Bob", IsAlive: true},
		},
		Phase: core.Phase{Type: core.PhaseDiscussion},
	}
	
	handler := NewChatHandler(nil) // No spectator broadcaster needed for this test
	acker := &MockActionAcker{GameID: "test-game"}

	// Player B reacts to a message from Player A
	reactionAction := core.Action{
		PlayerID: "player-b", // Bob is reacting
		Type:     core.ActionReactToMessage,
		Payload: map[string]interface{}{
			"message_id":  "msg-from-alice",
			"emoji":       "👍",
			"channel_id": "#war-room",
		},
	}

	// Act
	events, err := handler.Handle(state, reactionAction, acker)

	// Assert
	require.NoError(t, err)
	require.Len(t, events, 1)

	reactionEvent := events[0]
	assert.Equal(t, core.EventMessageReaction, reactionEvent.Type)
	
	// CRITICAL ASSERTION: The event's PlayerID must be the reactor's ID (Bob)
	assert.Equal(t, "player-b", reactionEvent.PlayerID)

	// Verify the payload structure contains correct player attribution
	payload, ok := reactionEvent.Payload.(core.MessageReactionPayload)
	require.True(t, ok, "Payload should be of type MessageReactionPayload")
	
	assert.Equal(t, "msg-from-alice", payload.MessageID)
	assert.Equal(t, "👍", payload.Emoji)
	assert.Equal(t, "player-b", payload.PlayerID)
	assert.Equal(t, "Bob", payload.PlayerName)
}

func TestChatHandler_HandleReaction_ValidatesPlayer(t *testing.T) {
	// Arrange
	state := &core.GameState{
		ID: "test-game",
		Players: map[string]*core.Player{
			"player-a": {ID: "player-a", Name: "Alice", IsAlive: true},
		},
		Phase: core.Phase{Type: core.PhaseDiscussion},
	}
	
	handler := NewChatHandler(nil)
	acker := &MockActionAcker{GameID: "test-game"}

	// Non-existent player tries to react
	reactionAction := core.Action{
		PlayerID: "non-existent-player",
		Type:     core.ActionReactToMessage,
		Payload: map[string]interface{}{
			"message_id": "msg-1",
			"emoji":      "👍",
			"channel_id": "#war-room",
		},
	}

	// Act
	events, err := handler.Handle(state, reactionAction, acker)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in game")
	assert.Nil(t, events)
}