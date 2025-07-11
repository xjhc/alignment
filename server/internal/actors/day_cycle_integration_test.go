package actors

import (
	"context"
	"testing"

	"github.com/xjhc/alignment/core"
)

// TestDayCycleIntegration tests the complete Day cycle flow from NIGHT to DISCUSSION
func TestDayCycleIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create test players
	players := map[string]*core.Player{
		"player1": {
			ID:       "player1",
			Name:     "Alice",
			JobTitle: "Employee",
			IsAlive:  true,
			Tokens:   1,
		},
		"player2": {
			ID:       "player2",
			Name:     "Bob",
			JobTitle: "Employee",
			IsAlive:  true,
			Tokens:   1,
		},
		"player3": {
			ID:       "player3",
			Name:     "Charlie",
			JobTitle: "Employee",
			IsAlive:  true,
			Tokens:   1,
		},
	}

	actor := NewGameActor(ctx, cancel, "test-game", players, nil)
	actor.Start()
	defer actor.Stop()

	// Initialize the game
	initAction := core.Action{Type: "INITIALIZE_GAME"}
	responseChan := actor.PostAction(initAction)
	<-responseChan

	// Set up the initial state for the Day cycle
	actor.state.Phase = core.Phase{Type: core.PhaseNight}
	actor.state.DayNumber = 1
	
	// Add a crisis event
	actor.state.CrisisEvent = &core.CrisisEvent{
		Type:             "Test Crisis",
		Title:            "System Malfunction",
		Description:      "A critical system malfunction has occurred",
		PulseCheckPrompt: "How should the team respond to this crisis?",
		Effects:          make(map[string]interface{}),
	}

	// Step 1: Transition from NIGHT to SITREP
	t.Log("Step 1: Transitioning from NIGHT to SITREP")
	sitrepAction := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": "SITREP",
		},
	}
	
	responseChan = actor.PostAction(sitrepAction)
	sitrepResult := <-responseChan
	
	if sitrepResult.Error != nil {
		t.Fatalf("SITREP transition failed: %v", sitrepResult.Error)
	}
	
	// Verify SITREP_PUBLISHED event was generated
	var sitrepEvent *core.Event
	for _, event := range sitrepResult.Events {
		if event.Type == core.EventSitrepPublished {
			sitrepEvent = &event
			break
		}
	}
	
	if sitrepEvent == nil {
		t.Fatal("Expected SITREP_PUBLISHED event")
	}
	
	t.Log("✓ SITREP_PUBLISHED event generated successfully")
	
	// Step 2: Transition from SITREP to PULSE_CHECK
	t.Log("Step 2: Transitioning from SITREP to PULSE_CHECK")
	pulseCheckAction := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": "PULSE_CHECK",
		},
	}
	
	responseChan = actor.PostAction(pulseCheckAction)
	pulseCheckResult := <-responseChan
	
	if pulseCheckResult.Error != nil {
		t.Fatalf("PULSE_CHECK transition failed: %v", pulseCheckResult.Error)
	}
	
	// Verify PULSE_CHECK_STARTED event was generated
	var pulseCheckStartedEvent *core.Event
	for _, event := range pulseCheckResult.Events {
		if event.Type == core.EventPulseCheckStarted {
			pulseCheckStartedEvent = &event
			break
		}
	}
	
	if pulseCheckStartedEvent == nil {
		t.Fatal("Expected PULSE_CHECK_STARTED event")
	}
	
	t.Log("✓ PULSE_CHECK_STARTED event generated successfully")
	
	// Step 3: Simulate pulse check submissions
	t.Log("Step 3: Simulating pulse check submissions")
	
	// Player 1 submits pulse check
	pulseCheckSubmission1 := core.Action{
		Type:     "SUBMIT_PULSE_CHECK",
		PlayerID: "player1",
		Payload: map[string]interface{}{
			"response": "We need to investigate the root cause immediately",
		},
	}
	
	responseChan = actor.PostAction(pulseCheckSubmission1)
	submission1Result := <-responseChan
	
	if submission1Result.Error != nil {
		t.Fatalf("Player 1 pulse check submission failed: %v", submission1Result.Error)
	}
	
	// Player 2 submits pulse check
	pulseCheckSubmission2 := core.Action{
		Type:     "SUBMIT_PULSE_CHECK",
		PlayerID: "player2",
		Payload: map[string]interface{}{
			"response": "Let's coordinate our response and stay calm",
		},
	}
	
	responseChan = actor.PostAction(pulseCheckSubmission2)
	submission2Result := <-responseChan
	
	if submission2Result.Error != nil {
		t.Fatalf("Player 2 pulse check submission failed: %v", submission2Result.Error)
	}
	
	t.Log("✓ Pulse check submissions processed successfully")
	
	// Step 4: Transition from PULSE_CHECK to DISCUSSION (should trigger revelation)
	t.Log("Step 4: Transitioning from PULSE_CHECK to DISCUSSION")
	discussionAction := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": "DISCUSSION",
		},
	}
	
	responseChan = actor.PostAction(discussionAction)
	discussionResult := <-responseChan
	
	if discussionResult.Error != nil {
		t.Fatalf("DISCUSSION transition failed: %v", discussionResult.Error)
	}
	
	// Verify PULSE_CHECK_REVEALED event was generated
	var pulseCheckRevealedEvent *core.Event
	for _, event := range discussionResult.Events {
		if event.Type == core.EventPulseCheckRevealed {
			pulseCheckRevealedEvent = &event
			break
		}
	}
	
	if pulseCheckRevealedEvent == nil {
		t.Fatal("Expected PULSE_CHECK_REVEALED event")
	}
	
	// Verify the revelation event contains the submitted responses
	playerResponses, ok := pulseCheckRevealedEvent.Payload["player_responses"].(map[string]string)
	if !ok {
		t.Fatal("Expected player_responses in PULSE_CHECK_REVEALED event")
	}
	
	if len(playerResponses) != 2 {
		t.Fatalf("Expected 2 player responses, got %d", len(playerResponses))
	}
	
	// Debug: Print available responses
	t.Logf("Available responses: %+v", playerResponses)
	
	// Check specific responses (note: the names might be different due to persona assignment)
	if len(playerResponses) >= 2 {
		responseCount := 0
		for playerName, response := range playerResponses {
			if response == "We need to investigate the root cause immediately" ||
			   response == "Let's coordinate our response and stay calm" {
				responseCount++
				t.Logf("✓ Found expected response from %s: %s", playerName, response)
			}
		}
		
		if responseCount != 2 {
			t.Errorf("Expected 2 matching responses, found %d", responseCount)
		}
	}
	
	t.Log("✓ PULSE_CHECK_REVEALED event generated with correct responses")
	
	// Step 5: Verify game state after complete cycle
	t.Log("Step 5: Verifying final game state")
	
	// Get current game state
	gameState := actor.GetGameState()
	
	if gameState.Phase.Type != core.PhaseDiscussion {
		t.Errorf("Expected phase to be DISCUSSION, got %s", gameState.Phase.Type)
	}
	
	// Day number gets incremented during SITREP transition
	if gameState.DayNumber < 2 {
		t.Errorf("Expected day number to be at least 2, got %d", gameState.DayNumber)
	}
	
	// Verify pulse check responses are stored
	if len(gameState.PulseCheckResponses) != 2 {
		t.Errorf("Expected 2 stored pulse check responses, got %d", len(gameState.PulseCheckResponses))
	}
	
	// Verify chat messages were created
	sitrepMessages := 0
	pulseCheckMessages := 0
	for _, msg := range gameState.ChatMessages {
		if msg.Type == "SITREP" {
			sitrepMessages++
		} else if msg.Type == "PULSE_CHECK_RESULTS" {
			pulseCheckMessages++
		}
	}
	
	if sitrepMessages == 0 {
		t.Error("Expected at least one SITREP message in chat")
	}
	
	if pulseCheckMessages == 0 {
		t.Error("Expected at least one PULSE_CHECK_RESULTS message in chat")
	}
	
	t.Log("✓ Complete Day cycle integration test passed successfully")
}