package game

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// VoteHandler handles vote-related actions
type VoteHandler struct {
	votingManager interface{
		HandleVoteAction(action core.Action) ([]core.Event, error)
		GetWinner() (string, int, bool)
		IsVoteComplete() bool
		CompleteVote()
		ClearVote()
	}
}

// NewVoteHandler creates a new VoteHandler
func NewVoteHandler(votingManager interface{
	HandleVoteAction(action core.Action) ([]core.Event, error)
	GetWinner() (string, int, bool)
	IsVoteComplete() bool
	CompleteVote()
	ClearVote()
}) *VoteHandler {
	return &VoteHandler{
		votingManager: votingManager,
	}
}

// Handle processes vote actions
func (h *VoteHandler) Handle(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	switch action.Type {
	case core.ActionSubmitVote:
		return h.handleVoteAction(state, action, acker)
	case core.ActionSubmitSkipVote:
		return h.handleSkipVoteAction(state, action, acker)
	default:
		return nil, fmt.Errorf("unsupported vote action type: %s", action.Type)
	}
}

// handleVoteAction processes regular vote actions
func (h *VoteHandler) handleVoteAction(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Use the voting manager to handle the vote
	events, err := h.votingManager.HandleVoteAction(action)
	if err != nil {
		return nil, err
	}

	// Check if we need to process vote completion and phase transitions
	additionalEvents := h.processVoteCompletion(state, acker)
	events = append(events, additionalEvents...)

	return events, nil
}

// handleSkipVoteAction processes skip vote actions
func (h *VoteHandler) handleSkipVoteAction(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Validate player exists and is alive
	player, exists := state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot vote to skip")
	}

	// Use core validation to process the skip vote action
	events, err := core.ProcessPlayerAction(*state, action, time.Now())
	if err != nil {
		return nil, err
	}

	// Apply the skip vote event to calculate the new state
	newState := *state
	for _, event := range events {
		newState = core.ApplyEvent(newState, event)
	}

	// Check if all living human players have voted to skip
	livingHumans := 0
	for _, p := range newState.Players {
		if p.IsAlive && p.ControlType == "HUMAN" {
			livingHumans++
		}
	}

	skipVotes := len(newState.SkipVotes)
	if skipVotes >= livingHumans && livingHumans > 0 {
		// All living humans have voted to skip - trigger immediate phase transition
		nextPhase := GetNextPhase(newState.Phase.Type)
		if nextPhase != core.PhaseGameOver {
			phaseDuration := GetPhaseDuration(nextPhase, state.Settings)
			transitionPayload := core.PhaseChangedPayload{
				PhaseType: string(nextPhase),
				Duration:  phaseDuration.Seconds(),
			}
			transitionEvent := core.NewEventWithTypedPayload(
				fmt.Sprintf("phase_transition_%s_%d", action.GameID, time.Now().UnixNano()),
				core.EventPhaseChanged,
				acker.GetGameID(),
				"",
				time.Now(),
				transitionPayload,
			)
			events = append(events, transitionEvent)
			
			// Add skip vote reset event to clear frontend state for the new phase
			// Calculate required votes for the new phase
			livingHumans := 0
			for _, player := range newState.Players {
				if player.IsAlive && player.ControlType == "HUMAN" {
					livingHumans++
				}
			}
			
			skipVotePayload := core.SkipVoteUpdatedPayload{
				CurrentVotes:  0,
				RequiredVotes: livingHumans,
				Voters:        []string{},
				HasVoted:      false,
				PlayerName:    "",
			}
			skipVoteResetEvent := core.NewEventWithTypedPayload(
				fmt.Sprintf("skip_vote_reset_%s_%d", action.GameID, time.Now().UnixNano()),
				core.EventSkipVoteUpdated,
				acker.GetGameID(),
				"",
				time.Now(),
				skipVotePayload,
			)
			events = append(events, skipVoteResetEvent)
		}
	}

	return events, nil
}

// processVoteCompletion handles vote completion logic
func (h *VoteHandler) processVoteCompletion(state *core.GameState, acker ActionAcker) []core.Event {
	if state.VoteState == nil {
		return []core.Event{}
	}

	var events []core.Event

	// Check if all alive players have voted
	if h.votingManager.IsVoteComplete() {
		// Mark vote as complete
		h.votingManager.CompleteVote()

		// Create detailed vote completion payload for UI components
		voteCompletePayload := core.VoteCompletedPayload{
			VoteType: string(state.VoteState.Type),
		}
		voteCompleteEvent := core.NewEventWithTypedPayload(
			fmt.Sprintf("vote_completed_%s_%d", state.VoteState.Type, time.Now().UnixNano()),
			core.EventVoteCompleted,
			acker.GetGameID(),
			"",
			time.Now(),
			voteCompletePayload,
		)
		events = append(events, voteCompleteEvent)

		// Create chat message for vote results display
		voteResultChatEvent := h.createVoteResultChatMessage(state, acker)
		events = append(events, voteResultChatEvent)

		// Handle extension vote results
		if state.VoteState.Type == core.VoteExtension {
			extensionEvents := h.handleExtensionVoteResults(state, acker)
			events = append(events, extensionEvents...)
		}
	}

	return events
}

// createVoteBreakdown creates a detailed breakdown of who voted for whom
func (h *VoteHandler) createVoteBreakdown(state *core.GameState) []map[string]interface{} {
	var breakdown []map[string]interface{}

	for voterID, targetID := range state.VoteState.Votes {
		voter := state.Players[voterID]
		target := state.Players[targetID]

		entry := map[string]interface{}{
			"voter_id":   voterID,
			"voter_name": voter.Name,
			"target_id":  targetID,
		}

		if target != nil {
			entry["target_name"] = target.Name
		} else {
			entry["target_name"] = "Unknown"
		}

		breakdown = append(breakdown, entry)
	}

	return breakdown
}

// createWinnerInfo creates winner information
func (h *VoteHandler) createWinnerInfo(state *core.GameState) map[string]interface{} {
	winner, votes, hasTie := h.votingManager.GetWinner()
	
	winnerInfo := map[string]interface{}{
		"winner_id":   winner,
		"vote_count":  votes,
		"has_tie":     hasTie,
	}

	if winner != "" {
		if player := state.Players[winner]; player != nil {
			winnerInfo["winner_name"] = player.Name
		}
	}

	return winnerInfo
}

// calculateTotalTokenWeight calculates total token weight for votes
func (h *VoteHandler) calculateTotalTokenWeight(state *core.GameState) int {
	totalWeight := 0
	for voterID := range state.VoteState.Votes {
		if player := state.Players[voterID]; player != nil {
			totalWeight += player.Tokens
		}
	}
	return totalWeight
}

// getWinnerID gets the winner ID from voting manager
func (h *VoteHandler) getWinnerID(state *core.GameState) string {
	winner, _, _ := h.votingManager.GetWinner()
	return winner
}

// getWinnerName gets the winner name from game state
func (h *VoteHandler) getWinnerName(state *core.GameState) string {
	winner, _, _ := h.votingManager.GetWinner()
	if winner != "" {
		if player := state.Players[winner]; player != nil {
			return player.Name
		}
	}
	return ""
}

// getWinnerVoteCount gets the winner vote count
func (h *VoteHandler) getWinnerVoteCount(state *core.GameState) int {
	_, votes, _ := h.votingManager.GetWinner()
	return votes
}

// createVoteResultChatMessage creates a chat message for vote results
func (h *VoteHandler) createVoteResultChatMessage(state *core.GameState, acker ActionAcker) core.Event {
	winner, votes, hasTie := h.votingManager.GetWinner()
	
	var message string
	if hasTie {
		message = "Vote ended in a tie. No player eliminated."
	} else if winner != "" {
		if player := state.Players[winner]; player != nil {
			message = fmt.Sprintf("Vote completed. %s received %d votes.", player.Name, votes)
		} else {
			message = fmt.Sprintf("Vote completed. Player %s received %d votes.", winner, votes)
		}
	} else {
		message = "Vote completed with no clear winner."
	}

	chatPayload := core.ChatMessagePayload{
		SenderID:   "system",
		SenderName: "System",
		Message:    message,
		IsSystem:   true,
		ChannelID:  "#war-room",
	}
	return core.NewEventWithTypedPayload(
		fmt.Sprintf("vote_result_chat_%d", time.Now().UnixNano()),
		core.EventChatMessage,
		acker.GetGameID(),
		"",
		time.Now(),
		chatPayload,
	)
}

// handleExtensionVoteResults handles extension vote results
func (h *VoteHandler) handleExtensionVoteResults(state *core.GameState, acker ActionAcker) []core.Event {
	winner, votes, hasTie := h.votingManager.GetWinner()
	
	var events []core.Event
	
	// Create a result message based on the vote outcome
	var resultMessage string
	var nextAction string
	
	if hasTie {
		// In case of tie, default to moving to nomination (equivalent to "NOMINATE" winning)
		resultMessage = "Extension vote ended in a tie. Moving to nomination phase."
		nextAction = "NOMINATE"
	} else if winner == "EXTEND" {
		resultMessage = fmt.Sprintf("Extension vote passed with %d votes. Discussion extended by 1 minute.", votes)
		nextAction = "EXTEND"
	} else if winner == "NOMINATE" {
		resultMessage = fmt.Sprintf("Vote to move to nomination passed with %d votes. Proceeding to nomination phase.", votes)
		nextAction = "NOMINATE"
	} else {
		// Fallback case - shouldn't happen with proper validation
		resultMessage = "Extension vote completed with unclear results. Moving to nomination phase."
		nextAction = "NOMINATE"
	}
	
	// Create chat message about the extension vote result
	chatPayload := core.ChatMessagePayload{
		SenderID:   "system",
		SenderName: "System",
		Message:    resultMessage,
		IsSystem:   true,
		ChannelID:  "#war-room",
	}
	chatEvent := core.NewEventWithTypedPayload(
		fmt.Sprintf("extension_vote_result_%d", time.Now().UnixNano()),
		core.EventChatMessage,
		acker.GetGameID(),
		"",
		time.Now(),
		chatPayload,
	)
	events = append(events, chatEvent)
	
	// Handle phase transition based on vote result
	if nextAction == "EXTEND" {
		// Extend the discussion phase by 1 minute
		// This will be handled by the phase manager via a scheduled timer extension
		extensionPayload := core.PhaseChangedPayload{
			PhaseType: string(core.PhaseDiscussion),
			Duration:  state.Settings.ExtensionDuration.Seconds(),
		}
		extensionEvent := core.NewEventWithTypedPayload(
			fmt.Sprintf("discussion_extended_%d", time.Now().UnixNano()),
			core.EventPhaseChanged,
			acker.GetGameID(),
			"",
			time.Now(),
			extensionPayload,
		)
		events = append(events, extensionEvent)
	} else {
		// Move to nomination phase
		nominationPayload := core.PhaseChangedPayload{
			PhaseType: string(core.PhaseNomination),
			Duration:  state.Settings.NominationDuration.Seconds(),
		}
		nominationEvent := core.NewEventWithTypedPayload(
			fmt.Sprintf("phase_transition_nomination_%d", time.Now().UnixNano()),
			core.EventPhaseChanged,
			acker.GetGameID(),
			"",
			time.Now(),
			nominationPayload,
		)
		events = append(events, nominationEvent)
	}
	
	return events
}