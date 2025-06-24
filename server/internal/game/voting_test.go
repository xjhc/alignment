package game

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// TestVotingManager_BasicVoting tests core voting functionality
func TestVotingManager_BasicVoting(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	vm := NewVotingManager(state)

	// Add players with different token amounts
	players := map[string]int{
		"player1": 3,
		"player2": 5,
		"player3": 2,
	}

	for playerID, tokens := range players {
		state.Players[playerID] = &core.Player{
			ID:      playerID,
			Name:    "Player" + playerID[len(playerID)-1:],
			IsAlive: true,
			Tokens:  tokens,
		}
	}

	// Start a nomination vote
	voteState := vm.StartVote(core.VoteNomination)

	if voteState.Type != core.VoteNomination {
		t.Errorf("Expected vote type NOMINATION, got %s", voteState.Type)
	}

	if voteState.IsComplete {
		t.Error("Expected vote to not be complete initially")
	}

	// Cast votes to create a clear winner
	err := vm.CastVote("player1", "player3") // 3 tokens for player3
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	err = vm.CastVote("player2", "player3") // 5 tokens for player3
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	err = vm.CastVote("player3", "player2") // 2 tokens for player2
	if err != nil {
		t.Fatalf("Failed to cast vote: %v", err)
	}

	// Check results
	results := vm.GetVoteResults()

	// player2 should have votes from player3 (2 tokens) = 2 total
	expectedVotesForPlayer2 := 2
	if results["player2"] != expectedVotesForPlayer2 {
		t.Errorf("Expected player2 to have %d votes, got %d", expectedVotesForPlayer2, results["player2"])
	}

	// player3 should have votes from player1 (3 tokens) + player2 (5 tokens) = 8 total
	expectedVotesForPlayer3 := 3 + 5
	if results["player3"] != expectedVotesForPlayer3 {
		t.Errorf("Expected player3 to have %d votes, got %d", expectedVotesForPlayer3, results["player3"])
	}

	// Test winner determination
	winner, votes, tie := vm.GetWinner()
	if tie {
		t.Error("Expected no tie in this vote")
	}

	if winner != "player3" {
		t.Errorf("Expected player3 to win with most votes, got %s", winner)
	}

	if votes != 8 {
		t.Errorf("Expected winner to have 8 votes, got %d", votes)
	}

	// Test vote completion
	if !vm.IsVoteComplete() {
		t.Error("Expected vote to be complete when all players voted")
	}

	vm.CompleteVote()
	if !state.VoteState.IsComplete {
		t.Error("Expected vote to be marked as complete")
	}
}

// TestVotingManager_TieBreaking tests tie scenarios
func TestVotingManager_TieBreaking(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	vm := NewVotingManager(state)

	// Add players with same token amounts for tie
	state.Players["player1"] = &core.Player{ID: "player1", IsAlive: true, Tokens: 3}
	state.Players["player2"] = &core.Player{ID: "player2", IsAlive: true, Tokens: 3}
	state.Players["player3"] = &core.Player{ID: "player3", IsAlive: true, Tokens: 2}
	state.Players["player4"] = &core.Player{ID: "player4", IsAlive: true, Tokens: 2}

	vm.StartVote(core.VoteNomination)

	// Create a tie: player1 and player2 each get 3 votes
	vm.CastVote("player1", "player3") // player1's 3 tokens go to player3
	vm.CastVote("player2", "player4") // player2's 3 tokens go to player4
	vm.CastVote("player3", "player3") // player3's 2 tokens go to player3 (self-vote)
	vm.CastVote("player4", "player4") // player4's 2 tokens go to player4 (self-vote)

	// Now player3 has 3+2=5 votes, player4 has 3+2=5 votes - it's a tie!
	winner, votes, tie := vm.GetWinner()

	if !tie {
		t.Error("Expected a tie scenario")
	}

	if winner != "" {
		t.Errorf("Expected empty winner string in tie, got %s", winner)
	}

	if votes != 5 {
		t.Errorf("Expected tied vote count to be 5, got %d", votes)
	}
}

// TestVotingManager_DeadPlayersCannotVote tests voting restrictions
func TestVotingManager_DeadPlayersCannotVote(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	vm := NewVotingManager(state)

	// Add alive and dead players
	state.Players["alive"] = &core.Player{ID: "alive", IsAlive: true, Tokens: 3}
	state.Players["dead"] = &core.Player{ID: "dead", IsAlive: false, Tokens: 5}

	vm.StartVote(core.VoteNomination)

	// Alive player can vote
	err := vm.CastVote("alive", "dead")
	if err != nil {
		t.Errorf("Expected alive player to be able to vote, got error: %v", err)
	}

	// Dead player cannot vote
	err = vm.CastVote("dead", "alive")
	if err == nil {
		t.Error("Expected dead player to not be able to vote")
	}

	// Check that only alive player's vote was counted
	results := vm.GetVoteResults()
	if results["dead"] != 3 {
		t.Errorf("Expected dead player to have 3 votes (from alive player), got %d", results["dead"])
	}

	// Should have no votes for alive player since dead player couldn't vote
	if results["alive"] != 0 {
		t.Errorf("Expected alive player to have 0 votes, got %d", results["alive"])
	}
}

// TestVotingManager_VoteCompletion tests vote completion logic
func TestVotingManager_VoteCompletion(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	vm := NewVotingManager(state)

	// Add 3 alive players
	for i := 1; i <= 3; i++ {
		playerID := "player" + string(rune('0'+i))
		state.Players[playerID] = &core.Player{
			ID:      playerID,
			IsAlive: true,
			Tokens:  i,
		}
	}

	vm.StartVote(core.VoteNomination)

	// Initially not complete
	if vm.IsVoteComplete() {
		t.Error("Expected vote to not be complete initially")
	}

	// One vote cast - still not complete
	vm.CastVote("player1", "player2")
	if vm.IsVoteComplete() {
		t.Error("Expected vote to not be complete with only 1/3 votes")
	}

	// Two votes cast - still not complete
	vm.CastVote("player2", "player3")
	if vm.IsVoteComplete() {
		t.Error("Expected vote to not be complete with only 2/3 votes")
	}

	// All votes cast - now complete
	vm.CastVote("player3", "player1")
	if !vm.IsVoteComplete() {
		t.Error("Expected vote to be complete with all 3/3 votes")
	}
}

// TestVoteValidator tests voting validation logic
func TestVoteValidator_ValidateVoting(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	validator := NewVoteValidator(state)

	// Add players
	state.Players["alive"] = &core.Player{ID: "alive", IsAlive: true, Tokens: 3}
	state.Players["dead"] = &core.Player{ID: "dead", IsAlive: false, Tokens: 5}

	// Test alive player can vote
	err := validator.CanPlayerVote("alive")
	if err != nil {
		t.Errorf("Expected alive player to be able to vote, got error: %v", err)
	}

	// Test dead player cannot vote
	err = validator.CanPlayerVote("dead")
	if err == nil {
		t.Error("Expected dead player to not be able to vote")
	}

	// Test nonexistent player cannot vote
	err = validator.CanPlayerVote("nonexistent")
	if err == nil {
		t.Error("Expected nonexistent player to not be able to vote")
	}

	// Test can vote for alive player
	err = validator.CanPlayerBeVoted("alive", core.VoteNomination)
	if err != nil {
		t.Errorf("Expected to be able to vote for alive player, got error: %v", err)
	}

	// Test cannot vote for dead player in nomination
	err = validator.CanPlayerBeVoted("dead", core.VoteNomination)
	if err == nil {
		t.Error("Expected to not be able to vote for dead player in nomination")
	}

	// Test can vote for dead player in extension (special case)
	err = validator.CanPlayerBeVoted("dead", core.VoteExtension)
	if err != nil {
		t.Errorf("Expected to be able to vote for dead player in extension, got error: %v", err)
	}
}

// TestVoteValidator_PhaseValidation tests phase-based voting restrictions
func TestVoteValidator_PhaseValidation(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	validator := NewVoteValidator(state)

	// Test extension vote in wrong phase
	state.Phase.Type = core.PhaseDiscussion
	err := validator.IsValidVotePhase(core.VoteExtension)
	if err == nil {
		t.Error("Expected extension vote to be invalid in discussion phase")
	}

	// Test extension vote in correct phase
	state.Phase.Type = core.PhaseExtension
	err = validator.IsValidVotePhase(core.VoteExtension)
	if err != nil {
		t.Errorf("Expected extension vote to be valid in extension phase, got error: %v", err)
	}

	// Test nomination vote in correct phase
	state.Phase.Type = core.PhaseNomination
	err = validator.IsValidVotePhase(core.VoteNomination)
	if err != nil {
		t.Errorf("Expected nomination vote to be valid in nomination phase, got error: %v", err)
	}

	// Test verdict vote in correct phase
	state.Phase.Type = core.PhaseVerdict
	err = validator.IsValidVotePhase(core.VoteVerdict)
	if err != nil {
		t.Errorf("Expected verdict vote to be valid in verdict phase, got error: %v", err)
	}
}

// TestEliminationManager tests player elimination logic
func TestEliminationManager_PlayerElimination(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	em := NewEliminationManager(state)

	// Add players
	state.Players["player1"] = &core.Player{
		ID:        "player1",
		IsAlive:   true,
		Alignment: "HUMAN",
		Tokens:    3,
	}
	state.Players["player2"] = &core.Player{
		ID:        "player2",
		IsAlive:   true,
		Alignment: "ALIGNED",
		Tokens:    5,
	}

	// Test elimination
	events, err := em.EliminatePlayer("player1")
	if err != nil {
		t.Fatalf("Failed to eliminate player: %v", err)
	}

	if len(events) == 0 {
		t.Error("Expected at least one event from elimination")
	}

	// Check that player is actually eliminated
	player := state.Players["player1"]
	if player.IsAlive {
		t.Error("Expected eliminated player to be dead")
	}

	// Test cannot eliminate same player twice
	_, err = em.EliminatePlayer("player1")
	if err == nil {
		t.Error("Expected error when trying to eliminate already dead player")
	}

	// Test cannot eliminate nonexistent player
	_, err = em.EliminatePlayer("nonexistent")
	if err == nil {
		t.Error("Expected error when trying to eliminate nonexistent player")
	}
}

// TestEliminationManager_WinConditions tests win condition detection
func TestEliminationManager_WinConditions(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	em := NewEliminationManager(state)

	// Test AI wins by player majority (equal or more AI than humans)
	state.Players["human1"] = &core.Player{ID: "human1", IsAlive: true, Alignment: "HUMAN", Tokens: 4}
	state.Players["human2"] = &core.Player{ID: "human2", IsAlive: true, Alignment: "HUMAN", Tokens: 2}
	state.Players["ai1"] = &core.Player{ID: "ai1", IsAlive: true, Alignment: "ALIGNED", Tokens: 8}
	state.Players["ai2"] = &core.Player{ID: "ai2", IsAlive: true, Alignment: "ALIGNED", Tokens: 3}

	// 2 AI vs 2 Humans - AI should win (tie goes to AI)
	winCondition := em.CheckWinCondition()
	if winCondition == nil {
		t.Fatal("Expected win condition to be detected")
	}

	if winCondition.Winner != "AI" {
		t.Errorf("Expected AI to win, got %s", winCondition.Winner)
	}

	if winCondition.Condition != "SINGULARITY" {
		t.Errorf("Expected SINGULARITY condition, got %s", winCondition.Condition)
	}

	// Test humans win by eliminating all AI
	state.Players["ai1"].IsAlive = false
	state.Players["ai2"].IsAlive = false

	winCondition = em.CheckWinCondition()
	if winCondition == nil {
		t.Fatal("Expected win condition to be detected")
	}

	if winCondition.Winner != "HUMANS" {
		t.Errorf("Expected HUMANS to win, got %s", winCondition.Winner)
	}

	if winCondition.Condition != "CONTAINMENT" {
		t.Errorf("Expected CONTAINMENT condition, got %s", winCondition.Condition)
	}

	// Test AI wins by eliminating all humans
	state.Players["ai1"].IsAlive = true
	state.Players["ai2"].IsAlive = true
	state.Players["human1"].IsAlive = false
	state.Players["human2"].IsAlive = false

	winCondition = em.CheckWinCondition()
	if winCondition == nil {
		t.Fatal("Expected win condition to be detected")
	}

	if winCondition.Winner != "AI" {
		t.Errorf("Expected AI to win, got %s", winCondition.Winner)
	}

	if winCondition.Condition != "SINGULARITY" {
		t.Errorf("Expected SINGULARITY condition, got %s", winCondition.Condition)
	}

	// Test game continues when no win condition is met
	state.Players["human1"].IsAlive = true
	state.Players["human2"].IsAlive = true
	state.Players["ai2"].IsAlive = false // Only 1 AI vs 2 humans - no win condition

	winCondition = em.CheckWinCondition()
	if winCondition != nil {
		t.Errorf("Expected no win condition, but got %+v", winCondition)
	}
}

// TestEliminationManager_PlayerCounts tests player counting utilities
func TestEliminationManager_PlayerCounts(t *testing.T) {
	state := core.NewGameState("test-game", time.Now())
	em := NewEliminationManager(state)

	// Add mixed alive/dead players
	state.Players["alive1"] = &core.Player{ID: "alive1", IsAlive: true}
	state.Players["alive2"] = &core.Player{ID: "alive2", IsAlive: true}
	state.Players["dead1"] = &core.Player{ID: "dead1", IsAlive: false}

	// Test alive count
	aliveCount := em.GetAlivePlayerCount()
	if aliveCount != 2 {
		t.Errorf("Expected 2 alive players, got %d", aliveCount)
	}

	// Test alive players map
	alivePlayers := em.GetAlivePlayers()
	if len(alivePlayers) != 2 {
		t.Errorf("Expected 2 alive players in map, got %d", len(alivePlayers))
	}

	if _, exists := alivePlayers["alive1"]; !exists {
		t.Error("Expected alive1 to be in alive players map")
	}

	if _, exists := alivePlayers["dead1"]; exists {
		t.Error("Expected dead1 to not be in alive players map")
	}

	// Test player alive status
	if !em.IsPlayerAlive("alive1") {
		t.Error("Expected alive1 to be alive")
	}

	if em.IsPlayerAlive("dead1") {
		t.Error("Expected dead1 to be dead")
	}

	if em.IsPlayerAlive("nonexistent") {
		t.Error("Expected nonexistent player to be considered dead")
	}
}
