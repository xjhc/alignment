package game

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// PlayerInteractionHandler handles player-specific interactions
type PlayerInteractionHandler struct{}

// NewPlayerInteractionHandler creates a new PlayerInteractionHandler
func NewPlayerInteractionHandler() *PlayerInteractionHandler {
	return &PlayerInteractionHandler{}
}

// Handle processes player interaction actions
func (h *PlayerInteractionHandler) Handle(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	switch action.Type {
	case core.ActionSubmitPulseCheck:
		return h.handlePulseCheckSubmission(state, action, acker)
	case core.ActionSetSlackStatus:
		return h.handleStatusUpdate(state, action, acker)
	case core.ActionSubmitExitInterview:
		return h.handleExitInterview(state, action, acker)
	case core.ActionSubmitWhistleblowerVote:
		return h.handleWhistleblowerVote(state, action, acker)
	case core.ActionTriggerExtensionVoting:
		return h.handleExtensionVotingTrigger(state, action, acker)
	default:
		return nil, fmt.Errorf("unsupported player interaction action type: %s", action.Type)
	}
}

// handlePulseCheckSubmission processes pulse check responses
func (h *PlayerInteractionHandler) handlePulseCheckSubmission(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate player and phase
	player := state.Players[action.PlayerID]
	if player == nil || !player.IsAlive {
		return nil, fmt.Errorf("invalid or dead player submitting pulse check")
	}

	if state.Phase.Type != core.PhasePulseCheck {
		return nil, fmt.Errorf("pulse check submissions only allowed during PULSE_CHECK phase")
	}

	// Extract response from action payload
	response, ok := action.Payload["response"].(string)
	if !ok || response == "" {
		return nil, fmt.Errorf("invalid pulse check response")
	}

	// Validate response length (reasonable limits for free-form text)
	if len(response) > 200 {
		return nil, fmt.Errorf("pulse check response too long (max 200 characters)")
	}

	// Generate a single, lean update event containing only the new submission's data.
	// The core ApplyEvent function will be responsible for all state calculations.
	question := h.generatePulseCheckQuestion()
	updateEvent := core.Event{
		ID:        fmt.Sprintf("pulse_update_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPulseCheckUpdated,
		GameID:    acker.GetGameID(),
		PlayerID:  "", // Empty PlayerID makes this a public event visible to all players
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message_id":  fmt.Sprintf("pulse_check_day_%d", state.DayNumber), // Deterministic ID for the UI component
			"question":    question,
			"player_id":   action.PlayerID, // The new submission's data
			"player_name": player.Name,
			"response":    response,
		},
	}

	return []core.Event{updateEvent}, nil
}

// handleStatusUpdate processes status message updates
func (h *PlayerInteractionHandler) handleStatusUpdate(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate player exists and is alive
	player := state.Players[action.PlayerID]
	if player == nil || !player.IsAlive {
		return nil, fmt.Errorf("invalid or dead player updating status")
	}

	// Extract status message from payload
	statusMessage, ok := action.Payload["status_message"].(string)
	if !ok || statusMessage == "" {
		return nil, fmt.Errorf("invalid or missing status_message")
	}

	// Validate status message length
	if len(statusMessage) > 100 {
		return nil, fmt.Errorf("status message too long (max 100 characters)")
	}

	// Generate status changed event
	event := core.Event{
		ID:        fmt.Sprintf("status_changed_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventSlackStatusChanged,
		GameID:    acker.GetGameID(),
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"status": statusMessage,
		},
	}

	return []core.Event{event}, nil
}

// handleExitInterview handles exit interview submissions
func (h *PlayerInteractionHandler) handleExitInterview(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate player exists and is eliminated (not alive)
	player := state.Players[action.PlayerID]
	if player == nil {
		return nil, fmt.Errorf("invalid player submitting exit interview")
	}

	// Allow exit interview only for eliminated players
	if player.IsAlive {
		return nil, fmt.Errorf("only eliminated players can submit exit interviews")
	}

	// Extract parting shot from payload
	partingShot, ok := action.Payload["parting_shot"].(string)
	if !ok || partingShot == "" {
		return nil, fmt.Errorf("invalid or missing parting_shot")
	}

	// Validate parting shot length
	if len(partingShot) > 50 {
		return nil, fmt.Errorf("parting shot too long (max 50 characters)")
	}

	// Generate parting shot set event
	event := core.Event{
		ID:        fmt.Sprintf("parting_shot_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPartingShotSet,
		GameID:    acker.GetGameID(),
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"parting_shot": partingShot,
		},
	}

	return []core.Event{event}, nil
}

// handleWhistleblowerVote processes whistleblower votes from deactivated players
func (h *PlayerInteractionHandler) handleWhistleblowerVote(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate player exists but is not alive (deactivated)
	player := state.Players[action.PlayerID]
	if player == nil {
		return nil, fmt.Errorf("invalid player submitting whistleblower vote")
	}

	if player.IsAlive {
		return nil, fmt.Errorf("only deactivated players can submit whistleblower votes")
	}

	// Extract target from payload
	targetID, ok := action.Payload["target_player_id"].(string)
	if !ok || targetID == "" {
		return nil, fmt.Errorf("invalid or missing target_player_id")
	}

	// Validate target exists and is alive
	target := state.Players[targetID]
	if target == nil {
		return nil, fmt.Errorf("target player %s not found", targetID)
	}

	if !target.IsAlive {
		return nil, fmt.Errorf("cannot vote for eliminated player %s", targetID)
	}

	// Create whistleblower vote event
	event := core.Event{
		ID:        fmt.Sprintf("whistleblower_vote_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventWhistleblowerVoteCast,
		GameID:    acker.GetGameID(),
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"voter_id":   action.PlayerID,
			"voter_name": player.Name,
			"target_id":  targetID,
			"target_name": target.Name,
		},
	}

	return []core.Event{event}, nil
}

// handleExtensionVotingTrigger processes extension voting triggers
func (h *PlayerInteractionHandler) handleExtensionVotingTrigger(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate player exists and is alive
	player := state.Players[action.PlayerID]
	if player == nil || !player.IsAlive {
		return nil, fmt.Errorf("invalid or dead player triggering extension voting")
	}

	// Validate that extension voting is allowed in current phase
	if state.Phase.Type != core.PhaseDiscussion && state.Phase.Type != core.PhaseNomination {
		return nil, fmt.Errorf("extension voting only allowed during DISCUSSION or NOMINATION phases")
	}

	// Check if extension voting is already in progress
	if state.VoteState != nil && state.VoteState.Type == core.VoteExtension {
		return nil, fmt.Errorf("extension voting already in progress")
	}

	// Create extension voting triggered event
	event := core.Event{
		ID:        fmt.Sprintf("extension_voting_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventExtensionVotingTriggered,
		GameID:    acker.GetGameID(),
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"triggered_by": action.PlayerID,
			"trigger_name": player.Name,
			"phase":        string(state.Phase.Type),
		},
	}

	return []core.Event{event}, nil
}

// generatePulseCheckQuestion generates a pulse check question
func (h *PlayerInteractionHandler) generatePulseCheckQuestion() string {
	questions := []string{
		"How are you feeling about the current game state?",
		"What's your read on the group dynamics?",
		"Are there any concerns you'd like to share?",
		"How confident are you in your current strategy?",
		"What's your assessment of the threat level?",
	}
	
	// For now, return a simple question. In a full implementation,
	// this might use randomization or be based on game state
	return questions[0]
}