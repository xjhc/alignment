package game

import (
	"fmt"

	"github.com/xjhc/alignment/core"
)

// VotingManager handles voting logic and calculations
type VotingManager struct {
	gameState *core.GameState
}

// NewVotingManager creates a new voting manager
func NewVotingManager(gameState *core.GameState) *VotingManager {
	return &VotingManager{
		gameState: gameState,
	}
}

// StartVote initializes a new voting session
func (vm *VotingManager) StartVote(voteType core.VoteType) *core.VoteState {
	voteState := &core.VoteState{
		Type:         voteType,
		Votes:        make(map[string]string),
		TokenWeights: make(map[string]int),
		Results:      make(map[string]int),
		IsComplete:   false,
	}

	vm.gameState.VoteState = voteState
	return voteState
}

// CastVote records a player's vote
func (vm *VotingManager) CastVote(playerID, targetID string) error {
	if vm.gameState.VoteState == nil {
		return fmt.Errorf("no active vote session")
	}

	player, exists := vm.gameState.Players[playerID]
	if !exists {
		return fmt.Errorf("player %s not found", playerID)
	}

	if !player.IsAlive {
		return fmt.Errorf("dead players cannot vote")
	}

	// Record the vote
	vm.gameState.VoteState.Votes[playerID] = targetID
	vm.gameState.VoteState.TokenWeights[playerID] = player.Tokens

	// Recalculate results
	vm.calculateVoteResults()

	return nil
}

// calculateVoteResults tallies all votes by token weight
func (vm *VotingManager) calculateVoteResults() {
	if vm.gameState.VoteState == nil {
		return
	}

	results := make(map[string]int)

	for voterID, candidateID := range vm.gameState.VoteState.Votes {
		if tokens, exists := vm.gameState.VoteState.TokenWeights[voterID]; exists {
			results[candidateID] += tokens
		}
	}

	vm.gameState.VoteState.Results = results
}

// GetVoteResults returns current vote tallies
func (vm *VotingManager) GetVoteResults() map[string]int {
	if vm.gameState.VoteState == nil {
		return make(map[string]int)
	}

	return vm.gameState.VoteState.Results
}

// GetWinner returns the candidate with the most votes (tokens)
func (vm *VotingManager) GetWinner() (string, int, bool) {
	if vm.gameState.VoteState == nil {
		return "", 0, false
	}

	var winner string
	var maxVotes int
	var tie bool

	winnerCount := 0
	for candidateID, votes := range vm.gameState.VoteState.Results {
		if votes > maxVotes {
			winner = candidateID
			maxVotes = votes
			winnerCount = 1
			tie = false
		} else if votes == maxVotes && votes > 0 {
			winnerCount++
			tie = true
		}
	}

	if tie && winnerCount > 1 {
		return "", maxVotes, true
	}

	return winner, maxVotes, false
}

// IsVoteComplete checks if all alive players have voted
func (vm *VotingManager) IsVoteComplete() bool {
	if vm.gameState.VoteState == nil {
		return false
	}

	alivePlayers := 0
	for _, player := range vm.gameState.Players {
		if player.IsAlive {
			alivePlayers++
		}
	}

	return len(vm.gameState.VoteState.Votes) >= alivePlayers
}

// CompleteVote finalizes the voting session
func (vm *VotingManager) CompleteVote() {
	if vm.gameState.VoteState != nil {
		vm.gameState.VoteState.IsComplete = true
	}
}

// ClearVote removes the current vote state
func (vm *VotingManager) ClearVote() {
	vm.gameState.VoteState = nil
}

// EliminationManager handles player elimination logic
type EliminationManager struct {
	gameState *core.GameState
}

// NewEliminationManager creates a new elimination manager
func NewEliminationManager(gameState *core.GameState) *EliminationManager {
	return &EliminationManager{
		gameState: gameState,
	}
}

// EliminatePlayer removes a player from the game
func (em *EliminationManager) EliminatePlayer(playerID string) (*core.Player, error) {
	player, exists := em.gameState.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found", playerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("player %s is already eliminated", playerID)
	}

	// Mark player as eliminated
	player.IsAlive = false

	// Return the eliminated player for role/alignment reveal
	return player, nil
}

// CheckWinCondition evaluates if either faction has won
func (em *EliminationManager) CheckWinCondition() *core.WinCondition {
	aliveHumans := 0
	aliveAligned := 0
	totalTokens := 0
	alignedTokens := 0

	for _, player := range em.gameState.Players {
		if player.IsAlive {
			totalTokens += player.Tokens

			if player.Alignment == "ALIGNED" {
				aliveAligned++
				alignedTokens += player.Tokens
			} else if player.Alignment == "HUMAN" {
				aliveHumans++
			}
		}
	}

	// Check elimination conditions first (higher priority)

	// AI wins if no humans remain
	if aliveHumans == 0 && aliveAligned > 0 {
		return &core.WinCondition{
			Winner:      "AI",
			Condition:   "ELIMINATION",
			Description: "All human players have been eliminated or converted",
		}
	}

	// Humans win if no aligned players remain (Containment)
	if aliveAligned == 0 && aliveHumans > 0 {
		return &core.WinCondition{
			Winner:      "HUMANS",
			Condition:   "CONTAINMENT",
			Description: "All AI-aligned players have been eliminated",
		}
	}

	// AI wins if they control 51% of tokens (Singularity)
	if totalTokens > 0 && float64(alignedTokens)/float64(totalTokens) > 0.51 {
		return &core.WinCondition{
			Winner:      "AI",
			Condition:   "SINGULARITY",
			Description: "AI has achieved token majority control",
		}
	}

	// Game continues
	return nil
}

// GetAlivePlayerCount returns the number of living players
func (em *EliminationManager) GetAlivePlayerCount() int {
	count := 0
	for _, player := range em.gameState.Players {
		if player.IsAlive {
			count++
		}
	}
	return count
}

// GetAlivePlayers returns all living players
func (em *EliminationManager) GetAlivePlayers() map[string]*core.Player {
	alive := make(map[string]*core.Player)
	for id, player := range em.gameState.Players {
		if player.IsAlive {
			alive[id] = player
		}
	}
	return alive
}

// IsPlayerAlive checks if a player is still alive
func (em *EliminationManager) IsPlayerAlive(playerID string) bool {
	if player, exists := em.gameState.Players[playerID]; exists {
		return player.IsAlive
	}
	return false
}

// VoteValidator provides voting validation logic
type VoteValidator struct {
	gameState *core.GameState
}

// NewVoteValidator creates a new vote validator
func NewVoteValidator(gameState *core.GameState) *VoteValidator {
	return &VoteValidator{
		gameState: gameState,
	}
}

// CanPlayerVote checks if a player is eligible to vote
func (vv *VoteValidator) CanPlayerVote(playerID string) error {
	player, exists := vv.gameState.Players[playerID]
	if !exists {
		return fmt.Errorf("player %s not found", playerID)
	}

	if !player.IsAlive {
		return fmt.Errorf("eliminated players cannot vote")
	}

	return nil
}

// CanPlayerBeVoted checks if a player can be voted for
func (vv *VoteValidator) CanPlayerBeVoted(targetID string, voteType core.VoteType) error {
	target, exists := vv.gameState.Players[targetID]
	if !exists {
		return fmt.Errorf("target player %s not found", targetID)
	}

	if !target.IsAlive && voteType != core.VoteExtension {
		return fmt.Errorf("cannot vote for eliminated player")
	}

	return nil
}

// IsValidVotePhase checks if voting is allowed in current phase
func (vv *VoteValidator) IsValidVotePhase(voteType core.VoteType) error {
	switch voteType {
	case core.VoteExtension:
		if vv.gameState.Phase.Type != core.PhaseExtension {
			return fmt.Errorf("extension votes only allowed during extension phase")
		}
	case core.VoteNomination:
		if vv.gameState.Phase.Type != core.PhaseNomination {
			return fmt.Errorf("nomination votes only allowed during nomination phase")
		}
	case core.VoteVerdict:
		if vv.gameState.Phase.Type != core.PhaseVerdict {
			return fmt.Errorf("verdict votes only allowed during verdict phase")
		}
	default:
		return fmt.Errorf("unknown vote type: %s", voteType)
	}

	return nil
}

// HandleVoteAction processes a vote action and returns events
func (vm *VotingManager) HandleVoteAction(action core.Action) ([]core.Event, error) {
	targetID, _ := action.Payload["target_id"].(string)
	
	// Create validator to check if vote is valid
	validator := NewVoteValidator(vm.gameState)
	
	// Determine vote type based on current phase
	var voteType core.VoteType
	switch vm.gameState.Phase.Type {
	case core.PhaseNomination:
		voteType = core.VoteNomination
	case core.PhaseVerdict:
		voteType = core.VoteVerdict
	case core.PhaseExtension:
		voteType = core.VoteExtension
	default:
		return nil, fmt.Errorf("voting not allowed in phase %s", vm.gameState.Phase.Type)
	}
	
	// Validate the vote
	if err := validator.IsValidVotePhase(voteType); err != nil {
		return nil, err
	}
	
	if err := validator.CanPlayerVote(action.PlayerID); err != nil {
		return nil, err
	}
	
	if targetID != "" {
		if err := validator.CanPlayerBeVoted(targetID, voteType); err != nil {
			return nil, err
		}
	}
	
	// Create the vote event
	event := core.Event{
		ID:        fmt.Sprintf("vote_%s_%s_%d", action.PlayerID, targetID, getCurrentTime().UnixNano()),
		Type:      core.EventVoteCast,
		GameID:    vm.gameState.ID,
		PlayerID:  action.PlayerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"target_id": targetID,
			"vote_type": string(voteType),
		},
	}
	
	return []core.Event{event}, nil
}
