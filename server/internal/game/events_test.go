package game

import (
	"testing"
	"time"
)

// TestEventApplication_PlayerLifecycle tests the complete player join/leave/elimination cycle
func TestEventApplication_PlayerLifecycle(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"
	playerName := "TestPlayer"
	jobTitle := "CISO"

	// Test player joining
	joinEvent := Event{
		ID:        "event-1",
		Type:      EventPlayerJoined,
		GameID:    "test-game",
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      playerName,
			"job_title": jobTitle,
		},
	}

	err := state.ApplyEvent(joinEvent)
	if err != nil {
		t.Fatalf("Failed to apply player joined event: %v", err)
	}

	// Verify player was added correctly
	player, exists := state.Players[playerID]
	if !exists {
		t.Fatal("Player was not added to game state")
	}

	if player.Name != playerName {
		t.Errorf("Expected player name %s, got %s", playerName, player.Name)
	}

	if player.JobTitle != jobTitle {
		t.Errorf("Expected job title %s, got %s", jobTitle, player.JobTitle)
	}

	if !player.IsAlive {
		t.Error("Expected player to be alive")
	}

	if player.Alignment != "HUMAN" {
		t.Errorf("Expected player alignment HUMAN, got %s", player.Alignment)
	}

	if player.Tokens != state.Settings.StartingTokens {
		t.Errorf("Expected player to start with %d tokens, got %d", state.Settings.StartingTokens, player.Tokens)
	}

	// Test player elimination
	eliminationEvent := Event{
		ID:        "event-2",
		Type:      EventPlayerEliminated,
		GameID:    "test-game",
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"role_type": "CISO",
			"alignment": "HUMAN",
		},
	}

	err = state.ApplyEvent(eliminationEvent)
	if err != nil {
		t.Fatalf("Failed to apply player eliminated event: %v", err)
	}

	// Verify player was eliminated correctly
	if player.IsAlive {
		t.Error("Expected player to be eliminated")
	}

	if player.Role == nil || player.Role.Type != RoleCISO {
		t.Error("Expected player role to be revealed on elimination")
	}
}

// TestEventApplication_RoleAssignment tests role and ability mechanics
func TestEventApplication_RoleAssignment(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"

	// Add player first
	joinEvent := Event{
		Type:     EventPlayerJoined,
		PlayerID: playerID,
		Payload:  map[string]interface{}{"name": "TestPlayer", "job_title": "CISO"},
	}
	state.ApplyEvent(joinEvent)

	// Test role assignment
	roleEvent := Event{
		ID:        "event-2",
		Type:      EventRoleAssigned,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"role_type":        "CISO",
			"role_name":        "Chief Security Officer",
			"role_description": "Protects company assets",
			"kpi_type":         "GUARDIAN",
			"kpi_description":  "Survive until Day 4",
			"alignment":        "HUMAN",
		},
	}

	err := state.ApplyEvent(roleEvent)
	if err != nil {
		t.Fatalf("Failed to apply role assigned event: %v", err)
	}

	player := state.Players[playerID]
	if player.Role == nil {
		t.Fatal("Expected player to have a role")
	}

	if player.Role.Type != RoleCISO {
		t.Errorf("Expected role type CISO, got %s", player.Role.Type)
	}

	if player.PersonalKPI == nil || player.PersonalKPI.Description != "Survive until Day 4" {
		if player.PersonalKPI == nil {
			t.Error("Expected personal KPI to be set")
		} else {
			t.Errorf("Expected personal KPI description 'Survive until Day 4', got %s", player.PersonalKPI.Description)
		}
	}

	// Test milestone progression
	milestoneEvent := Event{
		Type:     EventProjectMilestone,
		PlayerID: playerID,
		Payload:  map[string]interface{}{"milestone": float64(3)},
	}

	err = state.ApplyEvent(milestoneEvent)
	if err != nil {
		t.Fatalf("Failed to apply milestone event: %v", err)
	}

	if player.ProjectMilestones != 3 {
		t.Errorf("Expected 3 milestones, got %d", player.ProjectMilestones)
	}

	if !player.Role.IsUnlocked {
		t.Error("Expected role ability to be unlocked at 3 milestones")
	}

	// Test ability unlock
	abilityEvent := Event{
		Type:     EventRoleAbilityUnlocked,
		PlayerID: playerID,
		Payload: map[string]interface{}{
			"ability_name":        "Isolate Node",
			"ability_description": "Block target player actions",
		},
	}

	err = state.ApplyEvent(abilityEvent)
	if err != nil {
		t.Fatalf("Failed to apply ability unlock event: %v", err)
	}

	if player.Role.Ability == nil {
		t.Fatal("Expected player to have an ability")
	}

	if player.Role.Ability.Name != "Isolate Node" {
		t.Errorf("Expected ability name 'Isolate Node', got %s", player.Role.Ability.Name)
	}
}

// TestEventApplication_VotingSystem tests the complete voting mechanics
func TestEventApplication_VotingSystem(t *testing.T) {
	state := NewGameState("test-game")

	// Add multiple players
	players := []string{"player1", "player2", "player3"}
	for i, playerID := range players {
		joinEvent := Event{
			Type:     EventPlayerJoined,
			PlayerID: playerID,
			Payload: map[string]interface{}{
				"name":      "Player" + string(rune('1'+i)),
				"job_title": "Employee",
			},
		}
		state.ApplyEvent(joinEvent)

		// Give different token amounts
		tokenEvent := Event{
			Type:     EventTokensAwarded,
			PlayerID: playerID,
			Payload:  map[string]interface{}{"amount": float64(i + 1)},
		}
		state.ApplyEvent(tokenEvent)
	}

	// Start a nomination vote
	voteStartEvent := Event{
		Type:    EventVoteStarted,
		Payload: map[string]interface{}{"vote_type": "NOMINATION"},
	}

	err := state.ApplyEvent(voteStartEvent)
	if err != nil {
		t.Fatalf("Failed to start vote: %v", err)
	}

	if state.VoteState == nil {
		t.Fatal("Expected vote state to be initialized")
	}

	if state.VoteState.Type != VoteNomination {
		t.Errorf("Expected vote type NOMINATION, got %s", state.VoteState.Type)
	}

	// Cast votes
	voteEvent1 := Event{
		Type:     EventVoteCast,
		PlayerID: "player1",
		Payload: map[string]interface{}{
			"target_id": "player2",
			"vote_type": "NOMINATION",
		},
	}

	voteEvent2 := Event{
		Type:     EventVoteCast,
		PlayerID: "player2",
		Payload: map[string]interface{}{
			"target_id": "player3",
			"vote_type": "NOMINATION",
		},
	}

	voteEvent3 := Event{
		Type:     EventVoteCast,
		PlayerID: "player3",
		Payload: map[string]interface{}{
			"target_id": "player2",
			"vote_type": "NOMINATION",
		},
	}

	// Apply votes
	state.ApplyEvent(voteEvent1)
	state.ApplyEvent(voteEvent2)
	state.ApplyEvent(voteEvent3)

	// Check vote results (player1=2 tokens, player2=3 tokens, player3=4 tokens)
	// player2 should have: 2 tokens (from player1) + 4 tokens (from player3) = 6 tokens
	// player3 should have: 3 tokens (from player2) = 3 tokens
	expectedVotesForPlayer2 := 2 + 4 // player1 + player3 tokens
	expectedVotesForPlayer3 := 3     // player2 tokens

	if state.VoteState.Results["player2"] != expectedVotesForPlayer2 {
		t.Errorf("Expected player2 to have %d votes, got %d", expectedVotesForPlayer2, state.VoteState.Results["player2"])
	}

	if state.VoteState.Results["player3"] != expectedVotesForPlayer3 {
		t.Errorf("Expected player3 to have %d votes, got %d", expectedVotesForPlayer3, state.VoteState.Results["player3"])
	}

	// Complete vote
	voteCompleteEvent := Event{
		Type: EventVoteCompleted,
	}

	err = state.ApplyEvent(voteCompleteEvent)
	if err != nil {
		t.Fatalf("Failed to complete vote: %v", err)
	}

	if !state.VoteState.IsComplete {
		t.Error("Expected vote to be marked as complete")
	}
}

// TestEventApplication_AIConversion tests AI conversion mechanics
func TestEventApplication_AIConversion(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"

	// Add player
	joinEvent := Event{
		Type:     EventPlayerJoined,
		PlayerID: playerID,
		Payload:  map[string]interface{}{"name": "TestPlayer", "job_title": "Employee"},
	}
	state.ApplyEvent(joinEvent)

	// Test conversion attempt
	conversionAttemptEvent := Event{
		Type:     EventAIConversionAttempt,
		PlayerID: playerID,
		Payload: map[string]interface{}{
			"target_id": playerID,
			"ai_equity": float64(2),
		},
	}

	err := state.ApplyEvent(conversionAttemptEvent)
	if err != nil {
		t.Fatalf("Failed to apply conversion attempt: %v", err)
	}

	player := state.Players[playerID]
	if player.AIEquity != 2 {
		t.Errorf("Expected AI equity to be 2, got %d", player.AIEquity)
	}

	// Test successful conversion
	conversionSuccessEvent := Event{
		Type:     EventAIConversionSuccess,
		PlayerID: playerID,
	}

	err = state.ApplyEvent(conversionSuccessEvent)
	if err != nil {
		t.Fatalf("Failed to apply conversion success: %v", err)
	}

	if player.Alignment != "ALIGNED" {
		t.Errorf("Expected player alignment to be ALIGNED, got %s", player.Alignment)
	}

	if player.AIEquity != 0 {
		t.Errorf("Expected AI equity to be reset to 0, got %d", player.AIEquity)
	}
}

// TestEventApplication_PhaseTransitions tests game phase management
func TestEventApplication_PhaseTransitions(t *testing.T) {
	state := NewGameState("test-game")

	// Test game start
	gameStartEvent := Event{
		Type:      EventGameStarted,
		Timestamp: time.Now(),
	}

	err := state.ApplyEvent(gameStartEvent)
	if err != nil {
		t.Fatalf("Failed to apply game start event: %v", err)
	}

	if state.Phase.Type != PhaseSitrep {
		t.Errorf("Expected phase to be SITREP, got %s", state.Phase.Type)
	}

	if state.DayNumber != 1 {
		t.Errorf("Expected day number to be 1, got %d", state.DayNumber)
	}

	// Test day start
	dayStartEvent := Event{
		Type:      EventDayStarted,
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{"day_number": float64(2)},
	}

	err = state.ApplyEvent(dayStartEvent)
	if err != nil {
		t.Fatalf("Failed to apply day start event: %v", err)
	}

	if state.DayNumber != 2 {
		t.Errorf("Expected day number to be 2, got %d", state.DayNumber)
	}

	// Test night start
	nightStartEvent := Event{
		Type:      EventNightStarted,
		Timestamp: time.Now(),
	}

	err = state.ApplyEvent(nightStartEvent)
	if err != nil {
		t.Fatalf("Failed to apply night start event: %v", err)
	}

	if state.Phase.Type != PhaseNight {
		t.Errorf("Expected phase to be NIGHT, got %s", state.Phase.Type)
	}

	if state.Phase.Duration != state.Settings.NightDuration {
		t.Errorf("Expected night duration %v, got %v", state.Settings.NightDuration, state.Phase.Duration)
	}
}

// TestEventApplication_NightActions tests night action mechanics
func TestEventApplication_NightActions(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"
	targetID := "player-456"

	// Add players
	for _, id := range []string{playerID, targetID} {
		joinEvent := Event{
			Type:     EventPlayerJoined,
			PlayerID: id,
			Payload:  map[string]interface{}{"name": "Player", "job_title": "Employee"},
		}
		state.ApplyEvent(joinEvent)
	}

	// Test night action submission
	nightActionEvent := Event{
		Type:     EventNightActionSubmitted,
		PlayerID: playerID,
		Payload: map[string]interface{}{
			"action_type": "MINE",
			"target_id":   targetID,
		},
	}

	err := state.ApplyEvent(nightActionEvent)
	if err != nil {
		t.Fatalf("Failed to apply night action event: %v", err)
	}

	player := state.Players[playerID]
	if player.LastNightAction == nil {
		t.Fatal("Expected player to have a last night action")
	}

	if player.LastNightAction.Type != ActionMine {
		t.Errorf("Expected action type MINE, got %s", player.LastNightAction.Type)
	}

	if player.LastNightAction.TargetID != targetID {
		t.Errorf("Expected target ID %s, got %s", targetID, player.LastNightAction.TargetID)
	}

	// Test night action resolution
	resolutionEvent := Event{
		Type: EventNightActionsResolved,
		Payload: map[string]interface{}{
			"results": map[string]interface{}{
				playerID: map[string]interface{}{
					"tokens": float64(5),
					"status": "Mining successful",
				},
				targetID: map[string]interface{}{
					"tokens": float64(3),
					"status": "Mined by another player",
				},
			},
		},
	}

	err = state.ApplyEvent(resolutionEvent)
	if err != nil {
		t.Fatalf("Failed to apply night resolution event: %v", err)
	}

	if player.Tokens != 5 {
		t.Errorf("Expected player to have 5 tokens, got %d", player.Tokens)
	}

	if player.StatusMessage != "Mining successful" {
		t.Errorf("Expected status 'Mining successful', got %s", player.StatusMessage)
	}

	target := state.Players[targetID]
	if target.Tokens != 3 {
		t.Errorf("Expected target to have 3 tokens, got %d", target.Tokens)
	}
}

// TestEventApplication_ChatAndCommunication tests messaging system
func TestEventApplication_ChatAndCommunication(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"

	// Add player
	joinEvent := Event{
		Type:     EventPlayerJoined,
		PlayerID: playerID,
		Payload:  map[string]interface{}{"name": "TestPlayer", "job_title": "Employee"},
	}
	state.ApplyEvent(joinEvent)

	// Test chat message
	chatEvent := Event{
		ID:        "msg-1",
		Type:      EventChatMessage,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_name": "TestPlayer",
			"message":     "Hello everyone!",
			"is_system":   false,
		},
	}

	err := state.ApplyEvent(chatEvent)
	if err != nil {
		t.Fatalf("Failed to apply chat message event: %v", err)
	}

	if len(state.ChatMessages) != 1 {
		t.Fatalf("Expected 1 chat message, got %d", len(state.ChatMessages))
	}

	message := state.ChatMessages[0]
	if message.PlayerID != playerID {
		t.Errorf("Expected message from %s, got %s", playerID, message.PlayerID)
	}

	if message.Message != "Hello everyone!" {
		t.Errorf("Expected message 'Hello everyone!', got %s", message.Message)
	}

	if message.IsSystem {
		t.Error("Expected message to not be system message")
	}

	// Test system message
	systemEvent := Event{
		ID:        "sys-1",
		Type:      EventSystemMessage,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": "Day 1 has begun!",
		},
	}

	err = state.ApplyEvent(systemEvent)
	if err != nil {
		t.Fatalf("Failed to apply system message event: %v", err)
	}

	if len(state.ChatMessages) != 2 {
		t.Fatalf("Expected 2 chat messages, got %d", len(state.ChatMessages))
	}

	systemMessage := state.ChatMessages[1]
	if systemMessage.PlayerID != "SYSTEM" {
		t.Errorf("Expected system message from SYSTEM, got %s", systemMessage.PlayerID)
	}

	if !systemMessage.IsSystem {
		t.Error("Expected message to be system message")
	}
}

// TestEventApplication_WinConditions tests victory detection
func TestEventApplication_WinConditions(t *testing.T) {
	state := NewGameState("test-game")

	// Test victory condition
	victoryEvent := Event{
		Type:      EventVictoryCondition,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"winner":      "HUMANS",
			"condition":   "CONTAINMENT",
			"description": "All AI players have been eliminated",
		},
	}

	err := state.ApplyEvent(victoryEvent)
	if err != nil {
		t.Fatalf("Failed to apply victory condition event: %v", err)
	}

	if state.WinCondition == nil {
		t.Fatal("Expected win condition to be set")
	}

	if state.WinCondition.Winner != "HUMANS" {
		t.Errorf("Expected winner HUMANS, got %s", state.WinCondition.Winner)
	}

	if state.WinCondition.Condition != "CONTAINMENT" {
		t.Errorf("Expected condition CONTAINMENT, got %s", state.WinCondition.Condition)
	}

	if state.Phase.Type != PhaseGameOver {
		t.Errorf("Expected phase to be GAME_OVER, got %s", state.Phase.Type)
	}
}

// TestEventApplication_PulseCheck tests pulse check mechanics
func TestEventApplication_PulseCheck(t *testing.T) {
	state := NewGameState("test-game")
	playerID := "player-123"

	// Add player
	joinEvent := Event{
		Type:     EventPlayerJoined,
		PlayerID: playerID,
		Payload:  map[string]interface{}{"name": "TestPlayer", "job_title": "Employee"},
	}
	state.ApplyEvent(joinEvent)

	// Test pulse check start
	pulseStartEvent := Event{
		Type: EventPulseCheckStarted,
		Payload: map[string]interface{}{
			"question": "What's your biggest concern today?",
		},
	}

	err := state.ApplyEvent(pulseStartEvent)
	if err != nil {
		t.Fatalf("Failed to apply pulse check start event: %v", err)
	}

	if state.CrisisEvent == nil {
		t.Fatal("Expected crisis event to be initialized")
	}

	question := state.CrisisEvent.Effects["pulse_check_question"]
	if question != "What's your biggest concern today?" {
		t.Errorf("Expected pulse check question to be stored, got %v", question)
	}

	// Test pulse check submission
	pulseSubmitEvent := Event{
		Type:     EventPulseCheckSubmitted,
		PlayerID: playerID,
		Payload: map[string]interface{}{
			"response": "Trust but verify",
		},
	}

	err = state.ApplyEvent(pulseSubmitEvent)
	if err != nil {
		t.Fatalf("Failed to apply pulse check submission event: %v", err)
	}

	responses := state.CrisisEvent.Effects["pulse_responses"].(map[string]interface{})
	if responses[playerID] != "Trust but verify" {
		t.Errorf("Expected pulse response to be stored, got %v", responses[playerID])
	}
}
