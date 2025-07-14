package core

import (
	"encoding/json"
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestApplyEvent_PlayerJoined(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	now := time.Now()

	event := Event{
		ID:        "event-1",
		Type:      EventPlayerJoined,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			"name":      "Alice",
			"job_title": "Software Engineer",
		},
	}

	newState := ApplyEvent(*gameState, event)

	// Verify player was added
	if len(newState.Players) != 1 {
		t.Errorf("Expected 1 player, got %d", len(newState.Players))
	}

	player := newState.Players["player-1"]
	if player == nil {
		t.Fatal("Player not found")
	}

	if player.Name != "Alice" {
		t.Errorf("Expected name 'Alice', got '%s'", player.Name)
	}
	if player.JobTitle != "Software Engineer" {
		t.Errorf("Expected job title 'Software Engineer', got '%s'", player.JobTitle)
	}
	if !player.IsAlive {
		t.Error("Expected player to be alive")
	}
	if player.Tokens != gameState.Settings.StartingTokens {
		t.Errorf("Expected %d starting tokens, got %d", gameState.Settings.StartingTokens, player.Tokens)
	}
	if player.Alignment != "HUMAN" {
		t.Errorf("Expected alignment 'HUMAN', got '%s'", player.Alignment)
	}
}

func TestApplyEvent_VoteCast(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())

	// Add two players
	gameState.Players["player-1"] = &Player{
		ID:      "player-1",
		Name:    "Alice",
		Tokens:  3,
		IsAlive: true,
	}
	gameState.Players["player-2"] = &Player{
		ID:      "player-2",
		Name:    "Bob",
		Tokens:  2,
		IsAlive: true,
	}

	event := Event{
		ID:        "event-1",
		Type:      EventVoteCast,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": "player-2",
			"vote_type": "NOMINATION",
		},
	}

	newState := ApplyEvent(*gameState, event)

	// Verify vote state was created and updated
	if newState.VoteState == nil {
		t.Fatal("VoteState should not be nil")
	}

	if newState.VoteState.Type != VoteNomination {
		t.Errorf("Expected vote type NOMINATION, got %s", newState.VoteState.Type)
	}

	if newState.VoteState.Votes["player-1"] != "player-2" {
		t.Errorf("Expected player-1 to vote for player-2")
	}

	if newState.VoteState.TokenWeights["player-1"] != 3 {
		t.Errorf("Expected token weight 3, got %d", newState.VoteState.TokenWeights["player-1"])
	}

	if newState.VoteState.Results["player-2"] != 3 {
		t.Errorf("Expected 3 votes for player-2, got %d", newState.VoteState.Results["player-2"])
	}
}

func TestApplyEvent_TokensAwarded(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:      "player-1",
		Name:    "Alice",
		Tokens:  5,
		IsAlive: true,
	}

	event := Event{
		ID:        "event-1",
		Type:      EventTokensAwarded,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"amount": float64(3),
		},
	}

	newState := ApplyEvent(*gameState, event)

	player := newState.Players["player-1"]
	if player.Tokens != 8 {
		t.Errorf("Expected 8 tokens (5+3), got %d", player.Tokens)
	}
}

func TestApplyEvent_PlayerEliminated(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:        "player-1",
		Name:      "Alice",
		IsAlive:   true,
		Alignment: "HUMAN",
	}

	event := Event{
		ID:        "event-1",
		Type:      EventPlayerEliminated,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"role_type": "CISO",
			"alignment": "HUMAN",
		},
	}

	newState := ApplyEvent(*gameState, event)

	player := newState.Players["player-1"]
	if player.IsAlive {
		t.Error("Expected player to be eliminated (not alive)")
	}
	if player.Role == nil || player.Role.Type != RoleCISO {
		t.Error("Expected role to be revealed as CISO")
	}
	if player.Alignment != "HUMAN" {
		t.Errorf("Expected alignment 'HUMAN', got '%s'", player.Alignment)
	}
}

func TestApplyEvent_RoleAssigned(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:      "player-1",
		Name:    "Alice",
		IsAlive: true,
	}

	event := Event{
		ID:        "event-1",
		Type:      EventRoleAssigned,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"role_type":        "CISO",
			"role_name":        "Chief Information Security Officer",
			"role_description": "Protects the company from cyber threats",
			"kpi_type":         "GUARDIAN",
			"kpi_description":  "Keep the CISO alive until Day 4",
			"alignment":        "HUMAN",
			"ability": map[string]interface{}{
				"name":        "Isolate Node",
				"description": "Block a player from taking any actions tonight. If you are aligned and target another aligned player, the action appears to work but doesn't actually block them.",
				"isReady":     false,
			},
		},
	}

	newState := ApplyEvent(*gameState, event)

	player := newState.Players["player-1"]
	if player.Role == nil {
		t.Fatal("Role should not be nil")
	}

	if player.Role.Type != RoleCISO {
		t.Errorf("Expected role type CISO, got %s", player.Role.Type)
	}
	if player.Role.Name != "Chief Information Security Officer" {
		t.Errorf("Expected role name 'Chief Information Security Officer', got '%s'", player.Role.Name)
	}
	if player.PersonalKPI == nil {
		t.Fatal("PersonalKPI should not be nil")
	}
	if player.PersonalKPI.Type != KPIGuardian {
		t.Errorf("Expected KPI type GUARDIAN, got %s", player.PersonalKPI.Type)
	}
	if player.Alignment != "HUMAN" {
		t.Errorf("Expected alignment 'HUMAN', got '%s'", player.Alignment)
	}
	
	// Test that ability data is properly extracted and assigned
	if player.Role.Ability == nil {
		t.Fatal("Role ability should not be nil")
	}
	if player.Role.Ability.Name != "Isolate Node" {
		t.Errorf("Expected ability name 'Isolate Node', got '%s'", player.Role.Ability.Name)
	}
	if player.Role.Ability.Description != "Block a player from taking any actions tonight. If you are aligned and target another aligned player, the action appears to work but doesn't actually block them." {
		t.Errorf("Expected specific ability description, got '%s'", player.Role.Ability.Description)
	}
	if player.Role.Ability.IsReady != false {
		t.Errorf("Expected ability IsReady to be false, got %v", player.Role.Ability.IsReady)
	}
}

func TestApplyEvent_NightActionSubmitted(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:      "player-1",
		Name:    "Alice",
		IsAlive: true,
	}

	now := time.Now()
	event := Event{
		ID:        "event-1",
		Type:      EventNightActionSubmitted,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			"action_type": "MINE",
			"target_id":   "player-2",
		},
	}

	newState := ApplyEvent(*gameState, event)

	// Check submitted night action was stored
	if len(newState.NightActions) != 1 {
		t.Errorf("Expected 1 night action, got %d", len(newState.NightActions))
	}

	action := newState.NightActions["player-1"]
	if action == nil {
		t.Fatal("Night action should not be nil")
	}

	if action.Type != "MINE" {
		t.Errorf("Expected action type 'MINE', got '%s'", action.Type)
	}
	if action.TargetID != "player-2" {
		t.Errorf("Expected target 'player-2', got '%s'", action.TargetID)
	}

	// Check player's last action was updated
	player := newState.Players["player-1"]
	if player.LastNightAction == nil {
		t.Fatal("LastNightAction should not be nil")
	}
	if player.LastNightAction.Type != ActionMine {
		t.Errorf("Expected last action type MINE, got %s", player.LastNightAction.Type)
	}
}

func TestApplyEvent_AIConversionSuccess(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:        "player-1",
		Name:      "Alice",
		IsAlive:   true,
		Alignment: "HUMAN",
		AIEquity:  50,
	}

	event := Event{
		ID:        "event-1",
		Type:      EventAIConversionSuccess,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{},
	}

	newState := ApplyEvent(*gameState, event)

	player := newState.Players["player-1"]
	if player.Alignment != "ALIGNED" {
		t.Errorf("Expected alignment 'ALIGNED', got '%s'", player.Alignment)
	}
	if player.AIEquity != 0 {
		t.Errorf("Expected AI equity to be reset to 0, got %d", player.AIEquity)
	}
	if player.StatusMessage != "Conversion successful" {
		t.Errorf("Expected status message 'Conversion successful', got '%s'", player.StatusMessage)
	}
}

func TestApplyEvent_SystemShockApplied(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:      "player-1",
		Name:    "Alice",
		IsAlive: true,
	}

	event := Event{
		ID:        "event-1",
		Type:      EventSystemShockApplied,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"shock_type":     "MESSAGE_CORRUPTION",
			"description":    "Messages may be corrupted",
			"duration_hours": float64(24),
		},
	}

	newState := ApplyEvent(*gameState, event)

	player := newState.Players["player-1"]
	if len(player.SystemShocks) != 1 {
		t.Errorf("Expected 1 system shock, got %d", len(player.SystemShocks))
	}

	shock := player.SystemShocks[0]
	if shock.Type != ShockMessageCorruption {
		t.Errorf("Expected shock type MESSAGE_CORRUPTION, got %s", shock.Type)
	}
	if !shock.IsActive {
		t.Error("Expected shock to be active")
	}
	if shock.Description != "Messages may be corrupted" {
		t.Errorf("Expected description 'Messages may be corrupted', got '%s'", shock.Description)
	}
}

func TestApplyEvent_NightActionsResolved(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:              "player-1",
		Name:            "Alice",
		Tokens:          5,
		IsAlive:         true,
		HasUsedAbility:  true,
		LastNightAction: &NightAction{Type: ActionMine},
	}
	gameState.Players["player-2"] = &Player{
		ID:        "player-2",
		Name:      "Bob",
		IsAlive:   true,
		Alignment: "HUMAN",
		AIEquity:  25,
	}

	// Set up some night actions
	gameState.NightActions = map[string]*SubmittedNightAction{
		"player-1": {PlayerID: "player-1", Type: "MINE"},
	}

	event := Event{
		ID:        "event-1",
		Type:      EventNightActionsResolved,
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_state_changes": map[string]interface{}{
				"player-1": map[string]interface{}{
					"tokens_gained":  float64(2),
					"status_message": "Mining successful",
				},
				"player-2": map[string]interface{}{
					"alignment":      "ALIGNED",
					"ai_equity":      float64(0),
					"status_message": "Converted to AI",
				},
			},
		},
	}

	newState := ApplyEvent(*gameState, event)

	// Check player-1 results
	player1 := newState.Players["player-1"]
	if player1.Tokens != 7 { // 5 + 2
		t.Errorf("Expected player-1 to have 7 tokens, got %d", player1.Tokens)
	}
	if player1.StatusMessage != "Mining successful" {
		t.Errorf("Expected status 'Mining successful', got '%s'", player1.StatusMessage)
	}
	if player1.LastNightAction != nil {
		t.Error("Expected LastNightAction to be cleared")
	}
	if player1.HasUsedAbility {
		t.Error("Expected HasUsedAbility to be reset")
	}

	// Check player-2 results
	player2 := newState.Players["player-2"]
	if player2.Alignment != "ALIGNED" {
		t.Errorf("Expected player-2 alignment 'ALIGNED', got '%s'", player2.Alignment)
	}
	if player2.AIEquity != 0 {
		t.Errorf("Expected player-2 AI equity 0, got %d", player2.AIEquity)
	}

	// Check night actions were cleared
	if len(newState.NightActions) != 0 {
		t.Errorf("Expected night actions to be cleared, got %d", len(newState.NightActions))
	}
}

func TestApplyEvent_PhaseTransition(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.DayNumber = 1

	event := Event{
		ID:        "event-1",
		Type:      EventPhaseChanged,
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"phase_type": "SITREP",
			"duration":   float64(30 * time.Second), // 30 seconds as nanoseconds
		},
	}

	newState := ApplyEvent(*gameState, event)

	if newState.Phase.Type != PhaseSitrep {
		t.Errorf("Expected phase SITREP, got %s", newState.Phase.Type)
	}
	if newState.Phase.Duration != 30*time.Second {
		t.Errorf("Expected duration 30s, got %v", newState.Phase.Duration)
	}

	// Day number should increment when transitioning to SITREP
	if newState.DayNumber != 2 {
		t.Errorf("Expected day number 2, got %d", newState.DayNumber)
	}
}

func TestApplyEvent_VictoryCondition(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())

	event := Event{
		ID:        "event-1",
		Type:      EventVictoryCondition,
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"winner":      "HUMANS",
			"condition":   "CONTAINMENT",
			"description": "All AI threats eliminated",
		},
	}

	newState := ApplyEvent(*gameState, event)

	if newState.WinCondition == nil {
		t.Fatal("WinCondition should not be nil")
	}

	if newState.WinCondition.Winner != "HUMANS" {
		t.Errorf("Expected winner 'HUMANS', got '%s'", newState.WinCondition.Winner)
	}
	if newState.WinCondition.Condition != "CONTAINMENT" {
		t.Errorf("Expected condition 'CONTAINMENT', got '%s'", newState.WinCondition.Condition)
	}

	// Game should end
	if newState.Phase.Type != PhaseGameOver {
		t.Errorf("Expected phase GAME_OVER, got %s", newState.Phase.Type)
	}
}

func TestApplyEvent_ProjectMilestone(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:                "player-1",
		Name:              "Alice",
		IsAlive:           true,
		ProjectMilestones: 2,
		Role: &Role{
			Type:       RoleCISO,
			Name:       "Chief Information Security Officer",
			IsUnlocked: false,
		},
	}

	event := Event{
		ID:        "event-1",
		Type:      EventProjectMilestone,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"milestone": float64(3),
		},
	}

	newState := ApplyEvent(*gameState, event)

	player := newState.Players["player-1"]
	if player.ProjectMilestones != 3 {
		t.Errorf("Expected 3 project milestones, got %d", player.ProjectMilestones)
	}

	// Role should be unlocked when reaching 3 milestones
	if player.Role == nil || !player.Role.IsUnlocked {
		t.Error("Expected role to be unlocked at 3 milestones")
	}

	// Test milestone without role unlock (when already unlocked)
	gameState.Players["player-2"] = &Player{
		ID:                "player-2",
		Name:              "Bob",
		IsAlive:           true,
		ProjectMilestones: 1,
		Role: &Role{
			Type:       RoleCTO,
			Name:       "Chief Technology Officer",
			IsUnlocked: true, // Already unlocked
		},
	}

	event2 := Event{
		ID:        "event-2",
		Type:      EventProjectMilestone,
		GameID:    "test-game",
		PlayerID:  "player-2",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"milestone": float64(2),
		},
	}

	newState2 := ApplyEvent(newState, event2)
	player2 := newState2.Players["player-2"]
	if player2.ProjectMilestones != 2 {
		t.Errorf("Expected 2 project milestones for player-2, got %d", player2.ProjectMilestones)
	}
	if !player2.Role.IsUnlocked {
		t.Error("Expected role to remain unlocked for player-2")
	}
}

// Test table-driven approach for role abilities
func TestApplyEvent_RoleAbilities(t *testing.T) {
	testCases := []struct {
		name         string
		eventType    EventType
		playerRole   RoleType
		expectedUsed bool
		expectedMsg  string
	}{
		{
			name:         "CISO Audit",
			eventType:    EventRunAudit,
			playerRole:   RoleCISO,
			expectedUsed: true,
			expectedMsg:  "Audit completed",
		},
		{
			name:         "CTO Overclock",
			eventType:    EventOverclockServers,
			playerRole:   RoleCTO,
			expectedUsed: true,
			expectedMsg:  "Servers overclocked",
		},
		{
			name:         "COO Isolate",
			eventType:    EventIsolateNode,
			playerRole:   RoleCOO,
			expectedUsed: true,
			expectedMsg:  "Node isolated",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gameState := NewGameState("test-game", time.Now())
			gameState.Players["player-1"] = &Player{
				ID:      "player-1",
				Name:    "Alice",
				IsAlive: true,
				Role: &Role{
					Type:       tc.playerRole,
					IsUnlocked: true,
				},
				HasUsedAbility: false,
			}

			event := Event{
				ID:        "event-1",
				Type:      tc.eventType,
				GameID:    "test-game",
				PlayerID:  "player-1",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"target_id": "player-2",
				},
			}

			newState := ApplyEvent(*gameState, event)

			player := newState.Players["player-1"]
			if player.HasUsedAbility != tc.expectedUsed {
				t.Errorf("Expected HasUsedAbility %v, got %v", tc.expectedUsed, player.HasUsedAbility)
			}
			if player.StatusMessage != tc.expectedMsg {
				t.Errorf("Expected status message '%s', got '%s'", tc.expectedMsg, player.StatusMessage)
			}
		})
	}
}

func TestApplyEvent_ChatMessage_NestedPayload(t *testing.T) {
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{
		ID:      "player-1",
		Name:    "Kelly",
		IsAlive: true,
	}

	now := time.Now()
	event := Event{
		ID:        "event-1",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			"sender_id":   "player-1",
			"sender_name": "Kelly",
			"message":     "Hello, this is a test message",
			"isSystem":    false,
			"channel_id":  "#war-room",
			"id":          "msg-1752251515631258344",
			"timestamp":   "2025-07-11T09:31:55.631258598-07:00",
		},
	}

	newState := ApplyEvent(*gameState, event)

	// Check that chat message was added
	if len(newState.ChatMessages) != 1 {
		t.Errorf("Expected 1 chat message, got %d", len(newState.ChatMessages))
	}

	chatMsg := newState.ChatMessages[0]
	if chatMsg.ID != "msg-1752251515631258344" {
		t.Errorf("Expected message ID 'msg-1752251515631258344', got '%s'", chatMsg.ID)
	}
	if chatMsg.PlayerID != "player-1" {
		t.Errorf("Expected player ID 'player-1', got '%s'", chatMsg.PlayerID)
	}
	if chatMsg.PlayerName != "Kelly" {
		t.Errorf("Expected player name 'Kelly', got '%s'", chatMsg.PlayerName)
	}
	if chatMsg.Message != "Hello, this is a test message" {
		t.Errorf("Expected message 'Hello, this is a test message', got '%s'", chatMsg.Message)
	}
	if chatMsg.ChannelID != "#war-room" {
		t.Errorf("Expected channel ID '#war-room', got '%s'", chatMsg.ChannelID)
	}
	if chatMsg.IsSystem != false {
		t.Errorf("Expected IsSystem false, got %v", chatMsg.IsSystem)
	}
}

func TestApplyEvent_ChatMessage_InvalidPayload_HandledGracefully(t *testing.T) {
	// Setup: Create an initial state and a CHAT_MESSAGE event with an invalid (nested) payload
	// This tests that our new strict typing gracefully handles malformed events
	gameState := NewGameState("test-game", time.Now())
	gameState.Players["player-1"] = &Player{ID: "player-1", Name: "Kelly", IsAlive: true}

	now := time.Now()
	event := Event{
		ID:        "event-1",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			// This is the old nested structure that should no longer work
			"channel_id":         "#war-room",
			"client_message_id":  "1752251515357_q50ckoe8m",
			"day_number":         float64(1),
			"phase":              "SITREP",
			"message": map[string]interface{}{
				"id":         "msg-1752251515631258344",
				"playerID":   "guest:924472f5-652f-4cb0-bb18-881e79dc503c",
				"playerName": "Kelly",
				"message":    "4",
				"isSystem":   false,
				"channelID":  "#war-room",
				"timestamp":  now.Format(time.RFC3339Nano),
			},
		},
	}

	// Act: Apply the event - this should fail gracefully with our new strict typing
	newState := ApplyEvent(*gameState, event)

	// Assert: Invalid payload should be rejected, no chat message added
	assert.Len(t, newState.ChatMessages, 0, "Invalid payload should be rejected - no defensive parsing")
}

func TestApplyMessageReaction_Immutability(t *testing.T) {
	// Arrange: Create a game state with a chat message
	gameState := NewGameState("test-game", time.Now())
	now := time.Now()

	// First add a chat message
	chatEvent := Event{
		ID:        "msg-event-1",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			"sender_id":   "player-1",
			"sender_name": "Alice",
			"message":     "Hello world!",
			"channel_id":  "#war-room",
			"id":          "msg-1",
		},
	}
	gameStateWithMessage := ApplyEvent(*gameState, chatEvent)

	// Store reference to original ChatMessages slice
	originalSlice := gameStateWithMessage.ChatMessages
	originalSlicePtr := &gameStateWithMessage.ChatMessages[0]

	// Create a reaction event
	reactionEvent := Event{
		ID:        "reaction-event-1",
		Type:      EventMessageReaction,
		GameID:    "test-game",
		PlayerID:  "player-2",
		Timestamp: now.Add(time.Second),
		Payload: map[string]interface{}{
			"message_id":   "msg-1",
			"emoji":        "👍",
			"player_id":    "player-2",
			"player_name":  "Bob",
		},
	}

	// Act: Apply the reaction event
	newState := ApplyEvent(gameStateWithMessage, reactionEvent)

	// Assert: The ChatMessages slice should have a different memory address (immutability)
	newSlice := newState.ChatMessages
	newSlicePtr := &newState.ChatMessages[0]

	// Verify immutability: different slice references
	assert.True(t, &originalSlice[0] != &newSlice[0], "ChatMessages slice should have different underlying array")
	assert.NotSame(t, originalSlicePtr, newSlicePtr, "ChatMessage objects should be new instances")

	// Verify content correctness: the reaction was added
	assert.Len(t, newState.ChatMessages, 1, "Should still have 1 message")
	assert.Equal(t, "msg-1", newState.ChatMessages[0].ID, "Message ID should be preserved")
	assert.Len(t, newState.ChatMessages[0].Reactions, 1, "Should have 1 reaction")
	
	reaction := newState.ChatMessages[0].Reactions[0]
	assert.Equal(t, "👍", reaction.Emoji, "Reaction emoji should be correct")
	assert.Equal(t, "player-2", reaction.PlayerID, "Reaction player ID should be correct")
	assert.Equal(t, "Bob", reaction.PlayerName, "Reaction player name should be correct")

	// Verify original state is unchanged (true immutability)
	assert.Len(t, gameStateWithMessage.ChatMessages[0].Reactions, 0, "Original state should be unchanged")
}

func TestApplyMessageReaction_ToggleReaction(t *testing.T) {
	// Arrange: Create a game state with a chat message that already has a reaction
	gameState := NewGameState("test-game", time.Now())
	now := time.Now()

	// First add a chat message
	chatEvent := Event{
		ID:        "msg-event-1",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			"sender_id":   "player-1",
			"sender_name": "Alice",
			"message":     "Hello world!",
			"channel_id":  "#war-room",
			"id":          "msg-1",
		},
	}
	gameStateWithMessage := ApplyEvent(*gameState, chatEvent)

	// Add first reaction
	reactionEvent1 := Event{
		ID:        "reaction-event-1",
		Type:      EventMessageReaction,
		GameID:    "test-game",
		PlayerID:  "player-2",
		Timestamp: now.Add(time.Second),
		Payload: map[string]interface{}{
			"message_id":   "msg-1",
			"emoji":        "👍",
			"player_id":    "player-2",
			"player_name":  "Bob",
		},
	}
	gameStateWithReaction := ApplyEvent(gameStateWithMessage, reactionEvent1)

	// Act: Toggle the same reaction (should remove it)
	reactionEvent2 := Event{
		ID:        "reaction-event-2",
		Type:      EventMessageReaction,
		GameID:    "test-game",
		PlayerID:  "player-2",  // Same player
		Timestamp: now.Add(2 * time.Second),
		Payload: map[string]interface{}{
			"message_id":   "msg-1",
			"emoji":        "👍",  // Same emoji
			"player_id":    "player-2",
			"player_name":  "Bob",
		},
	}
	finalState := ApplyEvent(gameStateWithReaction, reactionEvent2)

	// Assert: Reaction should be removed (toggled off)
	assert.Len(t, finalState.ChatMessages, 1, "Should still have 1 message")
	assert.Len(t, finalState.ChatMessages[0].Reactions, 0, "Reaction should be removed when toggled")

	// Verify immutability: different slice references
	assert.True(t, &gameStateWithReaction.ChatMessages[0] != &finalState.ChatMessages[0], "ChatMessages slice should have different underlying array")
}

func TestApplyMessageReaction_CorrectPlayerAttribution(t *testing.T) {
	// Arrange: Create a game state with two players and a chat message from player A
	gameState := NewGameState("test-game", time.Now())
	now := time.Now()

	// Add two players
	gameState.Players["player-a"] = &Player{ID: "player-a", Name: "Alice", IsAlive: true}
	gameState.Players["player-b"] = &Player{ID: "player-b", Name: "Bob", IsAlive: true}

	// Add a chat message from Alice
	chatEvent := Event{
		ID:        "msg-event-1",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-a",
		Timestamp: now,
		Payload: map[string]interface{}{
			"sender_id":   "player-a",
			"sender_name": "Alice",
			"message":     "Hello world!",
			"channel_id":  "#war-room",
			"id":          "msg-1",
		},
	}
	gameStateWithMessage := ApplyEvent(*gameState, chatEvent)

	// Act: Bob reacts to Alice's message using new typed payload structure
	reactionEvent := Event{
		ID:        "reaction-event-1",
		Type:      EventMessageReaction,
		GameID:    "test-game",
		PlayerID:  "player-b", // Bob is reacting
		Timestamp: now.Add(1 * time.Second),
		Payload: map[string]interface{}{
			"message_id":  "msg-1",
			"emoji":       "👍",
			"player_id":   "player-b", // Explicit player ID in payload
			"player_name": "Bob",      // Bob's name in payload
		},
	}
	finalState := ApplyEvent(gameStateWithMessage, reactionEvent)

	// Assert: Reaction should be attributed to Bob, not Alice
	assert.Len(t, finalState.ChatMessages, 1, "Should have 1 message")
	assert.Len(t, finalState.ChatMessages[0].Reactions, 1, "Should have 1 reaction")

	reaction := finalState.ChatMessages[0].Reactions[0]
	assert.Equal(t, "👍", reaction.Emoji, "Should have correct emoji")
	assert.Equal(t, "player-b", reaction.PlayerID, "Should be attributed to Bob (player-b), not Alice")
	assert.Equal(t, "Bob", reaction.PlayerName, "Should have Bob's name, not Alice's")
}

func TestApplyMessageReaction_FullStateIntegrityCheck(t *testing.T) {
	// Simulate the exact scenario from the bug report
	// This test verifies that reactions persist correctly in the game state
	gameState := NewGameState("test-game", time.Now())
	now := time.Now()

	// Step 1: Add a chat message
	chatEvent := Event{
		ID:        "chat-msg-12345",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-a",
		Timestamp: now,
		Payload: map[string]interface{}{
			"sender_id":   "player-a",
			"sender_name": "Alice",
			"message":     "Hello everyone!",
			"channel_id":  "#war-room",
			"id":          "chat-msg-12345",
		},
	}
	stateWithMessage := ApplyEvent(*gameState, chatEvent)

	// Verify message was added correctly
	assert.Len(t, stateWithMessage.ChatMessages, 1, "Should have 1 message")
	assert.Equal(t, "chat-msg-12345", stateWithMessage.ChatMessages[0].ID, "Message ID should be correct")
	assert.Len(t, stateWithMessage.ChatMessages[0].Reactions, 0, "Should start with no reactions")

	// Step 2: Add a reaction using the exact payload structure from the bug report
	reactionEvent := Event{
		ID:        "reaction-event-67890",
		Type:      EventMessageReaction,
		GameID:    "test-game",
		PlayerID:  "player-b",
		Timestamp: now.Add(time.Second),
		Payload: map[string]interface{}{
			"message_id":  "chat-msg-12345",
			"emoji":       "👍",
			"player_id":   "player-b",
			"player_name": "Bob",
		},
	}

	// Step 3: Apply the reaction event
	finalState := ApplyEvent(stateWithMessage, reactionEvent)

	// Step 4: Comprehensive assertions
	assert.Len(t, finalState.ChatMessages, 1, "Should still have 1 message")
	
	message := finalState.ChatMessages[0]
	assert.Equal(t, "chat-msg-12345", message.ID, "Message ID should be preserved")
	
	// THE CRITICAL ASSERTION: Reactions should be present in the state
	assert.NotNil(t, message.Reactions, "Reactions array should not be nil")
	assert.Len(t, message.Reactions, 1, "Should have exactly 1 reaction")
	
	reaction := message.Reactions[0]
	assert.Equal(t, "👍", reaction.Emoji, "Reaction emoji should be correct")
	assert.Equal(t, "player-b", reaction.PlayerID, "Reaction player ID should be correct")
	assert.Equal(t, "Bob", reaction.PlayerName, "Reaction player name should be correct")
	assert.False(t, reaction.Timestamp.IsZero(), "Reaction timestamp should be set")

	// Step 5: Verify immutability - original state should be unchanged
	assert.Len(t, stateWithMessage.ChatMessages[0].Reactions, 0, "Original state should remain unchanged")

	// Step 6: Serialize to JSON to verify the data structure would be correctly transmitted
	jsonData, err := json.Marshal(finalState)
	assert.NoError(t, err, "Should be able to marshal final state to JSON")
	
	// Verify the JSON contains the reaction data
	assert.Contains(t, string(jsonData), `"reactions":[{"emoji":"👍"`, "JSON should contain the reaction")
}

func TestApplyChatMessage_ReactionsInitializedAsEmptySlice(t *testing.T) {
	// This test verifies the fix for the omitempty JSON issue
	// The Reactions field must be initialized as an empty slice, not nil
	gameState := NewGameState("test-game", time.Now())
	now := time.Now()

	// Create a chat message event
	chatEvent := Event{
		ID:        "chat-msg-test",
		Type:      EventChatMessage,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: now,
		Payload: map[string]interface{}{
			"sender_id":   "player-1",
			"sender_name": "Alice",
			"message":     "Hello world!",
			"channel_id":  "#war-room",
			"id":          "chat-msg-test",
		},
	}

	// Apply the chat message event
	newState := ApplyEvent(*gameState, chatEvent)

	// Critical assertions for the omitempty fix
	assert.Len(t, newState.ChatMessages, 1, "Should have 1 chat message")
	
	message := newState.ChatMessages[0]
	assert.NotNil(t, message.Reactions, "Reactions should not be nil")
	assert.Equal(t, 0, len(message.Reactions), "Reactions should be an empty slice")
	
	// The most important test: JSON serialization should include the reactions field
	jsonData, err := json.Marshal(newState)
	assert.NoError(t, err, "Should be able to marshal state to JSON")
	
	// Parse the JSON to verify the reactions field exists
	var parsedState map[string]interface{}
	err = json.Unmarshal(jsonData, &parsedState)
	assert.NoError(t, err, "Should be able to unmarshal JSON")
	
	chatMessages, exists := parsedState["chat_messages"].([]interface{})
	assert.True(t, exists, "chat_messages should exist in JSON")
	assert.Len(t, chatMessages, 1, "Should have 1 message in JSON")
	
	messageObj, ok := chatMessages[0].(map[string]interface{})
	assert.True(t, ok, "Message should be a JSON object")
	
	// THE KEY ASSERTION: reactions field should exist in JSON (not omitted due to omitempty)
	reactions, reactionsExists := messageObj["reactions"]
	assert.True(t, reactionsExists, "reactions field should exist in JSON (not omitted due to omitempty)")
	assert.NotNil(t, reactions, "reactions should not be null in JSON")
	
	reactionsArray, ok := reactions.([]interface{})
	assert.True(t, ok, "reactions should be an array in JSON")
	assert.Len(t, reactionsArray, 0, "reactions array should be empty but present")
}