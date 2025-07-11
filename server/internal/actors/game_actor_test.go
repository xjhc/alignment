package actors

import (
	"context"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// Helper function to create a test GameActor with initialized players
func createTestGameActor(t *testing.T) *GameActor {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Create test players
	players := map[string]*core.Player{
		"player1": {
			ID:       "player1",
			Name:     "Alice",
			JobTitle: "Employee",
			IsAlive:  true,
		},
		"player2": {
			ID:       "player2",
			Name:     "Bob",
			JobTitle: "Employee",
			IsAlive:  true,
		},
		"player3": {
			ID:       "player3",
			Name:     "Charlie",
			JobTitle: "Employee",
			IsAlive:  true,
		},
	}

	actor := NewGameActor(ctx, cancel, "test-game", players, nil) // nil PostgreSQL store for tests

	actor.Start()
	t.Cleanup(actor.Stop)

	// Initialize the game via an action
	initAction := core.Action{Type: "INITIALIZE_GAME"}
	responseChan := actor.PostAction(initAction)
	<-responseChan // Wait for initialization to complete

	return actor
}

// TestSitrepPhaseTransition tests the SITREP phase transition generates correct events
func TestSitrepPhaseTransition(t *testing.T) {
	actor := createTestGameActor(t)
	
	// Set up game state for SITREP phase transition
	actor.state.Phase = core.Phase{Type: core.PhaseNight}
	actor.state.DayNumber = 1
	
	// Create a phase transition action to SITREP
	action := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": "SITREP",
		},
	}
	
	// Process the action
	responseChan := actor.PostAction(action)
	result := <-responseChan
	
	// Verify the action was processed successfully
	if result.Error != nil {
		t.Fatalf("Expected no error, got: %v", result.Error)
	}
	
	// Verify events were generated
	if len(result.Events) == 0 {
		t.Fatal("Expected events to be generated for SITREP phase transition")
	}
	
	// Check for SITREP_PUBLISHED event
	var sitrepEvent *core.Event
	for _, event := range result.Events {
		if event.Type == core.EventSitrepPublished {
			sitrepEvent = &event
			break
		}
	}
	
	if sitrepEvent == nil {
		t.Fatal("Expected SITREP_PUBLISHED event to be generated")
	}
	
	// Verify the event has the correct structure
	if sitrepEvent.GameID != "test-game" {
		t.Errorf("Expected GameID 'test-game', got %v", sitrepEvent.GameID)
	}
	
	// Verify the payload contains daily_sitrep
	if _, exists := sitrepEvent.Payload["daily_sitrep"]; !exists {
		t.Error("Expected daily_sitrep in event payload")
	}
}

// TestPulseCheckPhaseTransition tests the PULSE_CHECK phase transition generates correct events
func TestPulseCheckPhaseTransition(t *testing.T) {
	actor := createTestGameActor(t)
	
	// Set up game state for PULSE_CHECK phase transition
	actor.state.Phase = core.Phase{Type: core.PhaseSitrep}
	actor.state.DayNumber = 2
	
	// Create a crisis event to provide pulse check question
	actor.state.CrisisEvent = &core.CrisisEvent{
		Type:             "Test Crisis",
		Title:            "Test Crisis Title",
		Description:      "Test crisis description",
		PulseCheckPrompt: "What is your response to this crisis?",
	}
	
	// Create a phase transition action to PULSE_CHECK
	action := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": "PULSE_CHECK",
		},
	}
	
	// Process the action
	responseChan := actor.PostAction(action)
	result := <-responseChan
	
	// Verify the action was processed successfully
	if result.Error != nil {
		t.Fatalf("Expected no error, got: %v", result.Error)
	}
	
	// Verify events were generated
	if len(result.Events) == 0 {
		t.Fatal("Expected events to be generated for PULSE_CHECK phase transition")
	}
	
	// Check for PULSE_CHECK_STARTED event
	var pulseCheckEvent *core.Event
	for _, event := range result.Events {
		if event.Type == core.EventPulseCheckStarted {
			pulseCheckEvent = &event
			break
		}
	}
	
	if pulseCheckEvent == nil {
		t.Fatal("Expected PULSE_CHECK_STARTED event to be generated")
	}
	
	// Verify the event has the correct structure
	if pulseCheckEvent.GameID != "test-game" {
		t.Errorf("Expected GameID 'test-game', got %v", pulseCheckEvent.GameID)
	}
	
	// Verify the payload contains question and day_number
	if _, exists := pulseCheckEvent.Payload["question"]; !exists {
		t.Error("Expected question in event payload")
	}
	if _, exists := pulseCheckEvent.Payload["day_number"]; !exists {
		t.Error("Expected day_number in event payload")
	}
}

// TestPulseCheckRevelation tests the pulse check revelation event generation
func TestPulseCheckRevelation(t *testing.T) {
	actor := createTestGameActor(t)
	
	// Set up game state with pulse check responses
	actor.state.Phase = core.Phase{Type: core.PhasePulseCheck}
	actor.state.DayNumber = 2
	actor.state.PulseCheckResponses = map[string]string{
		"player1": "This is my response to the crisis",
		"player2": "I think we should take immediate action",
	}
	
	// Create a phase transition action to DISCUSSION (which should trigger pulse check revelation)
	action := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": "DISCUSSION",
		},
	}
	
	// Process the action
	responseChan := actor.PostAction(action)
	result := <-responseChan
	
	// Verify the action was processed successfully
	if result.Error != nil {
		t.Fatalf("Expected no error, got: %v", result.Error)
	}
	
	// Check for PULSE_CHECK_REVEALED event
	var revelationEvent *core.Event
	for _, event := range result.Events {
		if event.Type == core.EventPulseCheckRevealed {
			revelationEvent = &event
			break
		}
	}
	
	if revelationEvent == nil {
		t.Fatal("Expected PULSE_CHECK_REVEALED event to be generated")
	}
	
	// Verify the event has the correct structure
	if revelationEvent.GameID != "test-game" {
		t.Errorf("Expected GameID 'test-game', got %v", revelationEvent.GameID)
	}
	
	// Verify the payload contains player responses
	if _, exists := revelationEvent.Payload["player_responses"]; !exists {
		t.Error("Expected player_responses in event payload")
	}
	if _, exists := revelationEvent.Payload["total_responses"]; !exists {
		t.Error("Expected total_responses in event payload")
	}
}

func TestGameActor_ProcessAction_LeaveGame(t *testing.T) {
	actor := createTestGameActor(t)

	// Test valid leave game action
	leaveAction := core.Action{
		Type:      core.ActionLeaveGame,
		PlayerID:  "player1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	responseChan := actor.PostAction(leaveAction)
	result := <-responseChan
	if result.Error != nil {
		t.Errorf("Expected no error for valid leave action, got: %v", result.Error)
	}

	if len(result.Events) != 1 { // Only the PLAYER_LEFT event
		t.Errorf("Expected 1 PLAYER_LEFT event, got %d", len(result.Events))
	}
	
	// Verify the event type is correct
	if result.Events[0].Type != core.EventPlayerLeft {
		t.Errorf("Expected event type to be PLAYER_LEFT, got %s", result.Events[0].Type)
	}
}

func TestGameActor_ProcessAction_InvalidPlayer(t *testing.T) {
	actor := createTestGameActor(t)

	// Test action from non-existent player
	invalidAction := core.Action{
		Type:      core.ActionLeaveGame,
		PlayerID:  "non-existent-player",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	responseChan := actor.PostAction(invalidAction)
	result := <-responseChan
	if result.Error == nil {
		t.Error("Expected error for action from non-existent player")
	}
}

func TestGameActor_ProcessAction_PhaseTransition(t *testing.T) {
	actor := createTestGameActor(t)

	// Test phase transition action
	phaseAction := core.Action{
		Type:      core.ActionType("PHASE_TRANSITION"),
		PlayerID:  "SYSTEM",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"next_phase": string(core.PhaseDiscussion),
		},
	}

	responseChan := actor.PostAction(phaseAction)
	result := <-responseChan
	if result.Error != nil {
		t.Errorf("Expected no error for valid phase transition, got: %v", result.Error)
	}

	// Verify game state was updated
	gameState := actor.GetGameState()
	if gameState.Phase.Type != core.PhaseDiscussion {
		t.Errorf("Expected game phase to be %s, got %s", core.PhaseDiscussion, gameState.Phase.Type)
	}
}

func TestGameActor_GetGameState(t *testing.T) {
	actor := createTestGameActor(t)

	gameState := actor.GetGameState()
	if gameState == nil {
		t.Fatal("Game state should not be nil")
	}

	if gameState.ID != "test-game" {
		t.Errorf("Expected game ID 'test-game', got '%s'", gameState.ID)
	}

	if len(gameState.Players) != 3 {
		t.Errorf("Expected 3 players, got %d", len(gameState.Players))
	}

	// Verify players were initialized correctly
	for _, player := range gameState.Players {
		if !player.IsAlive {
			t.Errorf("Player %s should be alive initially", player.ID)
		}
		if player.Role == nil {
			t.Errorf("Player %s should have a role assigned", player.ID)
		}
		if player.Alignment == "" {
			t.Errorf("Player %s should have an alignment assigned", player.ID)
		}
	}

	if gameState.DayNumber != 1 {
		t.Errorf("Expected day number 1, got %d", gameState.DayNumber)
	}
}

func TestGameActor_CreatePlayerStateUpdateEvent(t *testing.T) {
	actor := createTestGameActor(t)

	snapshot := actor.CreatePlayerStateUpdateEvent("player1")

	if snapshot.Type != "GAME_STATE_UPDATE" {
		t.Errorf("Expected event type GAME_STATE_UPDATE, got %s", snapshot.Type)
	}

	if snapshot.PlayerID != "player1" {
		t.Errorf("Expected player ID 'player1', got '%s'", snapshot.PlayerID)
	}

	payload := snapshot.Payload
	gameState, ok := payload["game_state"].(*core.GameState)
	if !ok {
		t.Fatal("Expected 'game_state' in payload of type *core.GameState")
	}

	// Player1 should have full information
	if player1, exists := gameState.Players["player1"]; !exists || player1.Role == nil {
		t.Error("Player1 should exist and have role info in their own snapshot")
	}
	// Player2 should have stripped information
	if player2, exists := gameState.Players["player2"]; !exists || player2.Role != nil {
		t.Error("Player2 should exist but not have role info in Player1's snapshot")
	}
}

func TestGameActor_ProcessAction_WhistleblowerVote(t *testing.T) {
	actor := createTestGameActor(t)
	
	// Set up a deactivated player (eliminate one player)
	gameState := actor.GetGameState()
	gameState.Players["player1"].IsAlive = false
	
	// Set up whistleblower voting state manually for testing
	gameState.WhistleblowerVoting = &core.WhistleblowerVoting{
		IsActive: true,
		CrisisOptions: []core.CrisisEventOption{
			{Type: "SYSTEM_SHOCK", Title: "System Shock", Description: "Test crisis"},
			{Type: "DATA_BREACH", Title: "Data Breach", Description: "Test crisis"},
			{Type: "INSIDER_THREAT", Title: "Insider Threat", Description: "Test crisis"},
		},
		Votes:       make(map[string]string),
		VoteResults: make(map[string]int),
	}

	// Test valid whistleblower vote action
	whistleblowerAction := core.Action{
		Type:      core.ActionSubmitWhistleblowerVote,
		PlayerID:  "player1", // Deactivated player
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"crisis_choice": "SYSTEM_SHOCK",
		},
	}

	responseChan := actor.PostAction(whistleblowerAction)
	result := <-responseChan
	
	if result.Error != nil {
		t.Errorf("Expected no error for valid whistleblower vote, got: %v", result.Error)
	}

	if len(result.Events) == 0 {
		t.Error("Expected at least one event for whistleblower vote")
	}

	// Verify the first event is a whistleblower vote cast event
	if result.Events[0].Type != core.EventWhistleblowerVoteCast {
		t.Errorf("Expected event type to be WHISTLEBLOWER_VOTE_CAST, got %s", result.Events[0].Type)
	}

	// Test invalid player (alive player trying to vote)
	invalidAction := core.Action{
		Type:      core.ActionSubmitWhistleblowerVote,
		PlayerID:  "player2", // Alive player
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"crisis_choice": "SYSTEM_SHOCK",
		},
	}

	responseChan2 := actor.PostAction(invalidAction)
	result2 := <-responseChan2
	
	if result2.Error == nil {
		t.Error("Expected error for alive player trying to submit whistleblower vote")
	}

	// Test missing crisis_choice
	missingPayloadAction := core.Action{
		Type:      core.ActionSubmitWhistleblowerVote,
		PlayerID:  "player1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{}, // Missing crisis_choice
	}

	responseChan3 := actor.PostAction(missingPayloadAction)
	result3 := <-responseChan3
	
	if result3.Error == nil {
		t.Error("Expected error for missing crisis_choice in payload")
	}
}

// TestGameActor_PostAction_Backpressure tests that the mailbox backpressure mechanism works correctly
func TestGameActor_PostAction_Backpressure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create test players
	players := map[string]*core.Player{
		"player1": {
			ID:       "player1",
			Name:     "Alice",
			JobTitle: "Employee",
			IsAlive:  true,
		},
	}

	// Create actor but don't start it (so the mailbox won't be processed)
	actor := NewGameActor(ctx, cancel, "test-game", players, nil)

	// Fill the mailbox (capacity is 100)
	// We need to send 101 messages to trigger backpressure
	testAction := core.Action{
		Type:      core.ActionLeaveGame,
		PlayerID:  "player1",
		GameID:    "test-game",
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	// Fill the mailbox to capacity
	for i := 0; i < 100; i++ {
		responseChan := actor.PostAction(testAction)
		// Don't wait for response since the actor isn't started
		select {
		case result := <-responseChan:
			if result.Error != nil {
				t.Fatalf("Unexpected error on message %d: %v", i, result.Error)
			}
		case <-time.After(10 * time.Millisecond):
			// This is expected for the unsent messages since the actor isn't processing
		}
	}

	// Now the 101st message should trigger backpressure
	responseChan := actor.PostAction(testAction)
	
	// This should return immediately with an error (backpressure)
	select {
	case result := <-responseChan:
		if result.Error == nil {
			t.Error("Expected error due to backpressure, but got no error")
		}
		expectedErrMsg := "server is busy, action for game test-game was dropped"
		if result.Error.Error() != expectedErrMsg {
			t.Errorf("Expected error message '%s', got '%s'", expectedErrMsg, result.Error.Error())
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("PostAction should have returned immediately due to backpressure")
	}
}
