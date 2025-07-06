
package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// WhistleblowerManager handles the whistleblower protocol for deactivated players
type WhistleblowerManager struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewWhistleblowerManager creates a new whistleblower manager
func NewWhistleblowerManager(gameState *core.GameState) *WhistleblowerManager {
	return &WhistleblowerManager{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// StartWhistleblowerVoting initiates voting for deactivated players to select the next day's crisis
func (wm *WhistleblowerManager) StartWhistleblowerVoting() *core.Event {
	// Only start if there are deactivated players
	deactivatedPlayers := wm.getDeactivatedPlayers()
	if len(deactivatedPlayers) == 0 {
		return nil
	}

	// Don't start if voting is already active
	if wm.gameState.WhistleblowerVoting != nil && wm.gameState.WhistleblowerVoting.IsActive {
		return nil
	}

	// Get 3 random crisis options
	crisisOptions := wm.selectRandomCrisisOptions(3)

	event := &core.Event{
		ID:        fmt.Sprintf("whistleblower_voting_started_%d_%d", wm.gameState.DayNumber, time.Now().UnixNano()),
		Type:      core.EventWhistleblowerVotingStarted,
		GameID:    wm.gameState.ID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"crisis_options": crisisOptions,
		},
	}

	return event
}

// SubmitVote processes a whistleblower vote from a deactivated player
func (wm *WhistleblowerManager) SubmitVote(playerID, crisisType string) (*core.Event, error) {
	// Verify player is deactivated
	player, exists := wm.gameState.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player not found")
	}

	if player.IsAlive {
		return nil, fmt.Errorf("only deactivated players can vote in whistleblower protocol")
	}

	// Verify voting is active
	if wm.gameState.WhistleblowerVoting == nil || !wm.gameState.WhistleblowerVoting.IsActive {
		return nil, fmt.Errorf("whistleblower voting is not currently active")
	}

	// Verify crisis type is valid option
	validOption := false
	for _, option := range wm.gameState.WhistleblowerVoting.CrisisOptions {
		if option.Type == crisisType {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, fmt.Errorf("invalid crisis option")
	}

	event := &core.Event{
		ID:        fmt.Sprintf("whistleblower_vote_cast_%s_%d", playerID, time.Now().UnixNano()),
		Type:      core.EventWhistleblowerVoteCast,
		GameID:    wm.gameState.ID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_id":     playerID,
			"player_name":   player.Name,
			"crisis_choice": crisisType,
		},
	}

	return event, nil
}

// CheckVotingComplete checks if all deactivated players have voted and completes voting if so
func (wm *WhistleblowerManager) CheckVotingComplete() *core.Event {
	if wm.gameState.WhistleblowerVoting == nil || !wm.gameState.WhistleblowerVoting.IsActive {
		return nil
	}

	deactivatedPlayers := wm.getDeactivatedPlayers()
	votedCount := len(wm.gameState.WhistleblowerVoting.Votes)

	// Complete voting if all deactivated players have voted
	if votedCount >= len(deactivatedPlayers) {
		return wm.completeVoting()
	}

	return nil
}

// ForceCompleteVoting completes the voting regardless of participation
func (wm *WhistleblowerManager) ForceCompleteVoting() *core.Event {
	if wm.gameState.WhistleblowerVoting == nil || !wm.gameState.WhistleblowerVoting.IsActive {
		return nil
	}

	return wm.completeVoting()
}

// GetSelectedCrisis returns the crisis chosen by the whistleblower vote, if available
func (wm *WhistleblowerManager) GetSelectedCrisis() string {
	if wm.gameState.WhistleblowerVoting == nil || !wm.gameState.WhistleblowerVoting.IsComplete {
		return ""
	}
	return wm.gameState.WhistleblowerVoting.SelectedCrisis
}

// ClearVoting clears the whistleblower voting state for the next night
func (wm *WhistleblowerManager) ClearVoting() {
	wm.gameState.WhistleblowerVoting = nil
}

// completeVoting finalizes the voting and determines the winning crisis
func (wm *WhistleblowerManager) completeVoting() *core.Event {
	// Find the crisis with the most votes
	selectedCrisis := wm.findWinningCrisis()

	event := &core.Event{
		ID:        fmt.Sprintf("whistleblower_voting_completed_%d_%d", wm.gameState.DayNumber, time.Now().UnixNano()),
		Type:      core.EventWhistleblowerVotingCompleted,
		GameID:    wm.gameState.ID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"selected_crisis": selectedCrisis,
			"vote_results":    wm.gameState.WhistleblowerVoting.VoteResults,
		},
	}

	return event
}

// findWinningCrisis determines which crisis received the most votes
func (wm *WhistleblowerManager) findWinningCrisis() string {
	if wm.gameState.WhistleblowerVoting == nil {
		return ""
	}

	maxVotes := 0
	var winners []string

	// Find crisis(es) with the most votes
	for crisisType, votes := range wm.gameState.WhistleblowerVoting.VoteResults {
		if votes > maxVotes {
			maxVotes = votes
			winners = []string{crisisType}
		} else if votes == maxVotes && votes > 0 {
			winners = append(winners, crisisType)
		}
	}

	// Handle ties by random selection
	if len(winners) == 0 {
		// No votes cast, pick random option
		if len(wm.gameState.WhistleblowerVoting.CrisisOptions) > 0 {
			return wm.gameState.WhistleblowerVoting.CrisisOptions[wm.rng.Intn(len(wm.gameState.WhistleblowerVoting.CrisisOptions))].Type
		}
		return ""
	} else if len(winners) == 1 {
		return winners[0]
	} else {
		// Tie-breaker
		return winners[wm.rng.Intn(len(winners))]
	}
}

// getDeactivatedPlayers returns all players who are not alive
func (wm *WhistleblowerManager) getDeactivatedPlayers() []*core.Player {
	var deactivated []*core.Player
	for _, player := range wm.gameState.Players {
		if !player.IsAlive {
			deactivated = append(deactivated, player)
		}
	}
	return deactivated
}

// selectRandomCrisisOptions selects N random crisis options for voting
func (wm *WhistleblowerManager) selectRandomCrisisOptions(count int) []core.CrisisEventOption {
	// Get all available crisis types from the crisis manager
	crisisManager := NewCrisisEventManager(wm.gameState)
	allCrises := crisisManager.GetAllCrisisEvents()

	// Convert to options format
	var allOptions []core.CrisisEventOption
	for _, crisis := range allCrises {
		allOptions = append(allOptions, core.CrisisEventOption{
			Type:        string(crisis.Type),
			Title:       crisis.Title,
			Description: crisis.Description,
		})
	}

	// Randomly select the requested number of options
	if len(allOptions) <= count {
		return allOptions
	}

	// Shuffle and take first N
	wm.rng.Shuffle(len(allOptions), func(i, j int) {
		allOptions[i], allOptions[j] = allOptions[j], allOptions[i]
	})

	return allOptions[:count]
}