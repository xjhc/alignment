package actors

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/xjhc/alignment/core"
)

func TestHandlePhaseTransition_SkipTrialWhenNoNominations(t *testing.T) {
	// Setup: Create a GameActor in NOMINATION phase with no votes
	ga := createTestGameActor(3) // 3 players
	ga.state.Phase = core.Phase{
		Type:      core.PhaseNomination,
		StartTime: time.Now(),
	}
	ga.state.VoteState = &core.VoteState{
		Type:        "NOMINATION",
		Votes:       make(map[string]string), // Empty votes
		TokenWeights: make(map[string]int),
		Results:     make(map[string]int),
		IsComplete:  true,
	}

	// Create phase transition action that would normally go to TRIAL
	action := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": string(core.PhaseTrial),
		},
	}

	// Execute the phase transition
	events, err := ga.handlePhaseTransition(action)

	// Assertions
	assert.NoError(t, err, "Phase transition should not error")
	assert.Len(t, events, 2, "Should generate 2 events: system message + phase change")

	// Check that a system message was generated
	systemMessage := events[0]
	assert.Equal(t, core.EventChatMessage, systemMessage.Type)
	assert.Contains(t, systemMessage.Payload["message"], "No consensus was reached")
	assert.Equal(t, "Loebmate", systemMessage.Payload["player_name"])
	assert.Equal(t, true, systemMessage.Payload["is_system"])

	// Check that the phase changed to NIGHT, not TRIAL
	phaseChangeEvent := events[1]
	assert.Equal(t, core.EventPhaseChanged, phaseChangeEvent.Type)
	assert.Equal(t, string(core.PhaseNight), phaseChangeEvent.Payload["phase_type"])
	assert.Equal(t, string(core.PhaseNomination), phaseChangeEvent.Payload["previous_phase"])
}

func TestHandlePhaseTransition_ProceedToTrialWhenPlayerNominated(t *testing.T) {
	// Setup: Create a GameActor in NOMINATION phase with a clear winner
	ga := createTestGameActor(3) // 3 players
	ga.state.Phase = core.Phase{
		Type:      core.PhaseNomination,
		StartTime: time.Now(),
	}

	// Set up vote state with a clear winner
	playerIDs := make([]string, 0, len(ga.state.Players))
	for id := range ga.state.Players {
		playerIDs = append(playerIDs, id)
	}
	winner := playerIDs[0]
	voter1 := playerIDs[1]
	voter2 := playerIDs[2]

	ga.state.VoteState = &core.VoteState{
		Type: "NOMINATION",
		Votes: map[string]string{
			voter1: winner, // Player 1 votes for winner
			voter2: winner, // Player 2 votes for winner
		},
		TokenWeights: map[string]int{
			voter1: 1,
			voter2: 1,
		},
		Results: map[string]int{
			winner: 2, // Winner has 2 votes
		},
		IsComplete: true,
	}

	// Create phase transition action that would go to TRIAL
	action := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": string(core.PhaseTrial),
		},
	}

	// Execute the phase transition
	events, err := ga.handlePhaseTransition(action)

	// Assertions
	assert.NoError(t, err, "Phase transition should not error")
	assert.Len(t, events, 1, "Should generate 1 event: phase change (no skip)")

	// Check that the phase proceeded to TRIAL normally
	phaseChangeEvent := events[0]
	assert.Equal(t, core.EventPhaseChanged, phaseChangeEvent.Type)
	assert.Equal(t, string(core.PhaseTrial), phaseChangeEvent.Payload["phase_type"])
	assert.Equal(t, string(core.PhaseNomination), phaseChangeEvent.Payload["previous_phase"])
}

func TestHandlePhaseTransition_ProceedToTrialWhenTieVote(t *testing.T) {
	// Setup: Create a GameActor in NOMINATION phase with a tie
	ga := createTestGameActor(4) // 4 players
	ga.state.Phase = core.Phase{
		Type:      core.PhaseNomination,
		StartTime: time.Now(),
	}

	// Set up vote state with a tie
	playerIDs := make([]string, 0, len(ga.state.Players))
	for id := range ga.state.Players {
		playerIDs = append(playerIDs, id)
	}

	ga.state.VoteState = &core.VoteState{
		Type: "NOMINATION",
		Votes: map[string]string{
			playerIDs[2]: playerIDs[0], // Player 2 votes for Player 0
			playerIDs[3]: playerIDs[1], // Player 3 votes for Player 1
		},
		TokenWeights: map[string]int{
			playerIDs[2]: 1,
			playerIDs[3]: 1,
		},
		Results: map[string]int{
			playerIDs[0]: 1, // Tie: 1 vote each
			playerIDs[1]: 1,
		},
		IsComplete: true,
	}

	// Create phase transition action that would go to TRIAL
	action := core.Action{
		Type: "PHASE_TRANSITION",
		Payload: map[string]interface{}{
			"next_phase": string(core.PhaseTrial),
		},
	}

	// Execute the phase transition
	events, err := ga.handlePhaseTransition(action)

	// Assertions
	assert.NoError(t, err, "Phase transition should not error")
	assert.Len(t, events, 2, "Should generate 2 events: system message + phase change")

	// Check that a system message was generated (tie = no nomination)
	systemMessage := events[0]
	assert.Equal(t, core.EventChatMessage, systemMessage.Type)
	assert.Contains(t, systemMessage.Payload["message"], "No consensus was reached")

	// Check that the phase changed to NIGHT (skip trial due to tie)
	phaseChangeEvent := events[1]
	assert.Equal(t, core.EventPhaseChanged, phaseChangeEvent.Type)
	assert.Equal(t, string(core.PhaseNight), phaseChangeEvent.Payload["phase_type"])
}

// Helper function to create a test GameActor with specified number of players
func createTestGameActor(numPlayers int) *GameActor {
	settings := core.GameSettings{
		NightDuration: 60 * time.Second,
		// Add other required durations
		SitrepDuration:     30 * time.Second,
		PulseCheckDuration: 30 * time.Second,
		DiscussionDuration: 60 * time.Second,
		NominationDuration: 30 * time.Second,
		TrialDuration:      30 * time.Second,
		VerdictDuration:    30 * time.Second,
		ExtensionDuration:  60 * time.Second,
	}

	players := make(map[string]*core.Player)
	for i := 0; i < numPlayers; i++ {
		playerID := string(rune('A' + i)) // Players A, B, C, etc.
		players[playerID] = &core.Player{
			ID:           playerID,
			Name:         "Player " + playerID,
			IsAlive:      true,
			ControlType:  "HUMAN",
			Alignment:    core.AlignmentHuman,
			Tokens:       1,
		}
	}

	gameState := &core.GameState{
		ID:       "test-game",
		Players:  players,
		Settings: settings,
		Phase: core.Phase{
			Type:      core.PhaseNomination,
			StartTime: time.Now(),
		},
		DayNumber: 1,
	}

	return &GameActor{
		gameID: "test-game",
		state:  gameState,
		votingManager: &MockVotingManager{
			state: gameState,
		},
	}
}

// Mock voting manager for testing
type MockVotingManager struct {
	state *core.GameState
}

func (m *MockVotingManager) GetWinner() (string, int, bool) {
	if m.state.VoteState == nil || len(m.state.VoteState.Results) == 0 {
		return "", 0, false // No votes = no winner, no tie
	}

	maxVotes := 0
	winner := ""
	tieCount := 0

	for playerID, votes := range m.state.VoteState.Results {
		if votes > maxVotes {
			maxVotes = votes
			winner = playerID
			tieCount = 1
		} else if votes == maxVotes && votes > 0 {
			tieCount++
		}
	}

	return winner, maxVotes, tieCount > 1
}

func (m *MockVotingManager) ClearVote() {
	m.state.VoteState = nil
}

func (m *MockVotingManager) HandleVoteAction(action core.Action) ([]core.Event, error) {
	return []core.Event{}, nil
}

func (m *MockVotingManager) IsVoteComplete() bool {
	return m.state.VoteState != nil && m.state.VoteState.IsComplete
}