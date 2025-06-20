package core

import (
	"testing"
	"time"
)

func TestApplyEvent_PlayerJoined(t *testing.T) {
	gameState := NewGameState("test-game")
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
	gameState := NewGameState("test-game")
	
	// Add two players
	gameState.Players["player-1"] = &Player{
		ID:     "player-1",
		Name:   "Alice",
		Tokens: 3,
		IsAlive: true,
	}
	gameState.Players["player-2"] = &Player{
		ID:     "player-2",
		Name:   "Bob",
		Tokens: 2,
		IsAlive: true,
	}

	event := Event{
		ID:        "event-1",
		Type:      EventVoteCast,
		GameID:    "test-game",
		PlayerID:  "player-1",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id":  "player-2",
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
	gameState := NewGameState("test-game")
	gameState.Players["player-1"] = &Player{
		ID:     "player-1",
		Name:   "Alice",
		Tokens: 5,
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
	gameState := NewGameState("test-game")
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
			"role_type":  "CISO",
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
	gameState := NewGameState("test-game")
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
}

func TestApplyEvent_NightActionSubmitted(t *testing.T) {
	gameState := NewGameState("test-game")
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
	gameState := NewGameState("test-game")
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
	gameState := NewGameState("test-game")
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
			"shock_type":      "MESSAGE_CORRUPTION",
			"description":     "Messages may be corrupted",
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
	gameState := NewGameState("test-game")
	gameState.Players["player-1"] = &Player{
		ID:     "player-1",
		Name:   "Alice",
		Tokens: 5,
		IsAlive: true,
		HasUsedAbility: true,
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
			"results": map[string]interface{}{
				"player-1": map[string]interface{}{
					"token_change":   float64(2),
					"status_message": "Mining successful",
				},
				"player-2": map[string]interface{}{
					"alignment":  "ALIGNED",
					"ai_equity":  float64(0),
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
	gameState := NewGameState("test-game")
	gameState.DayNumber = 1

	event := Event{
		ID:        "event-1",
		Type:      EventPhaseChanged,
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"phase_type": "SITREP",
			"duration":   float64(30), // 30 seconds
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
	gameState := NewGameState("test-game")

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
			gameState := NewGameState("test-game")
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