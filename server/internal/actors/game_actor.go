package actors

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/game"
)

// Manager interfaces for better testability
type VotingManager interface {
	HandleVoteAction(action core.Action) ([]core.Event, error)
}

type MiningManager interface {
	HandleMineAction(action core.Action) ([]core.Event, error)
}

type RoleAbilityManager interface {
	HandleNightAction(action core.Action) ([]core.Event, error)
}

// SendStateToPlayer is a message to send current state to a specific player
type SendStateToPlayer struct {
	PlayerID string
}

// GameActor represents a single game instance running in its own goroutine
type GameActor struct {
	gameID  string
	state   *core.GameState
	mailbox chan interface{}

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies (interfaces for testing)
	datastore   DataStore
	broadcaster Broadcaster

	// Game managers (domain experts)
	votingManager      VotingManager
	miningManager      MiningManager
	roleAbilityManager RoleAbilityManager
	eliminationManager *game.EliminationManager
	phaseManager       *game.PhaseManager
	scheduler          *game.Scheduler
}

// DataStore interface for persistence
type DataStore interface {
	AppendEvent(gameID string, event core.Event) error
	SaveSnapshot(gameID string, state *core.GameState) error
	LoadEvents(gameID string, afterSequence int) ([]core.Event, error)
	LoadSnapshot(gameID string) (*core.GameState, error)
}

// Broadcaster interface for sending events to clients
type Broadcaster interface {
	BroadcastToGame(gameID string, event core.Event) error
	SendToPlayer(gameID, playerID string, event core.Event) error
}

// NewGameActor creates a new game actor
func NewGameActor(ctx context.Context, cancel context.CancelFunc, gameID string, datastore DataStore, broadcaster Broadcaster) *GameActor {
	state := core.NewGameState(gameID)

	// Create scheduler and phase manager
	scheduler := game.NewScheduler(nil) // We'll set the callback after creating the actor
	phaseManager := game.NewPhaseManager(scheduler, gameID, state.Settings)

	actor := &GameActor{
		gameID:      gameID,
		state:       state,
		mailbox:     make(chan interface{}, 100), // Buffered channel
		ctx:         ctx,
		cancel:      cancel,
		datastore:   datastore,
		broadcaster: broadcaster,

		// Initialize managers with shared state
		votingManager:      game.NewVotingManager(state),
		miningManager:      game.NewMiningManager(state),
		roleAbilityManager: game.NewRoleAbilityManager(state),
		eliminationManager: game.NewEliminationManager(state),
		phaseManager:       phaseManager,
		scheduler:          scheduler,
	}

	// Set the timer callback to route to this actor
	scheduler = game.NewScheduler(actor.HandleTimer)
	actor.scheduler = scheduler
	actor.phaseManager = game.NewPhaseManager(scheduler, gameID, state.Settings)

	return actor
}

// Start begins the actor's main processing loop
func (ga *GameActor) Start() {
	log.Printf("GameActor %s: Starting", ga.gameID)

	// Start the scheduler
	ga.scheduler.Start()

	// Start the main processing loop in a goroutine
	go ga.processLoop()
}

// Stop gracefully shuts down the actor
func (ga *GameActor) Stop() {
	log.Printf("GameActor %s: Stopping", ga.gameID)

	// Stop the scheduler
	ga.scheduler.Stop()

	ga.cancel()
}

// SendAction sends an action to the actor's mailbox
func (ga *GameActor) SendAction(action core.Action) {
	select {
	case ga.mailbox <- action:
		// Action queued successfully
	case <-ga.ctx.Done():
		log.Printf("GameActor %s: Context canceled, dropping action %s", ga.gameID, action.Type)
	default:
		log.Printf("GameActor %s: Mailbox full, dropping action %s", ga.gameID, action.Type)
	}
}

// SendCurrentStateToPlayer sends current game state to a specific player
func (ga *GameActor) SendCurrentStateToPlayer(playerID string) {
	select {
	case ga.mailbox <- SendStateToPlayer{PlayerID: playerID}:
		// State request queued successfully
	case <-ga.ctx.Done():
		log.Printf("GameActor %s: Context canceled, dropping state request for %s", ga.gameID, playerID)
	default:
		log.Printf("GameActor %s: Mailbox full, dropping state request for %s", ga.gameID, playerID)
	}
}

// HandleTimer handles timer callbacks from the scheduler
func (ga *GameActor) HandleTimer(timer game.Timer) {
	// Convert timer action to game action
	action := core.Action{
		Type:      core.ActionType(timer.Action.Type),
		PlayerID:  "", // System action
		GameID:    ga.gameID,
		Timestamp: time.Now(),
		Payload:   timer.Action.Payload,
	}

	ga.SendAction(action)
}

// processLoop is the main actor processing loop
func (ga *GameActor) processLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("GameActor %s: Panic recovered: %v", ga.gameID, r)
			// The supervisor will detect this actor has stopped and can restart it
		}
	}()

	for {
		select {
		case msg := <-ga.mailbox:
			// Check if context is canceled before processing messages
			if ga.ctx.Err() != nil {
				log.Printf("GameActor %s: Context canceled, stopping processing", ga.gameID)
				return
			}
			
			switch v := msg.(type) {
			case core.Action:
				ga.handleAction(v)
			case SendStateToPlayer:
				if err := ga.sendCurrentStateToPlayer(v.PlayerID); err != nil {
					log.Printf("GameActor %s: Error sending state to player %s: %v", ga.gameID, v.PlayerID, err)
				}
			default:
				log.Printf("GameActor %s: Unknown message type received: %T", ga.gameID, v)
			}
		case <-ga.ctx.Done():
			log.Printf("GameActor %s: Context done, shutting down", ga.gameID)
			return
		}
	}
}

// handleAction processes a single action following Validate -> Persist -> Apply order
func (ga *GameActor) handleAction(action core.Action) {
	log.Printf("GameActor %s: Processing action %s from player %s", ga.gameID, action.Type, action.PlayerID)

	// 1. VALIDATE (using ga.state for read-only checks)
	events, err := ga.generateEventsForAction(action)
	if err != nil {
		log.Printf("GameActor %s: Invalid action %s: %v", ga.gameID, action.Type, err)
		return
	}

	// 2. PERSIST
	for _, event := range events {
		err := ga.datastore.AppendEvent(ga.gameID, event)
		if err != nil {
			log.Printf("GameActor %s: CRITICAL - FAILED TO PERSIST EVENT %s: %v", ga.gameID, event.ID, err)
			// TODO: Implement a retry or shutdown mechanism here. A failed persist is a fatal error for this game.
			return
		}

		// 3. APPLY (now that it's safely persisted)
		newState := core.ApplyEvent(*ga.state, event)
		ga.state = &newState

		// 4. BROADCAST
		if err := ga.broadcaster.BroadcastToGame(ga.gameID, event); err != nil {
			log.Printf("GameActor %s: Failed to broadcast event: %v", ga.gameID, err)
		}

		// 5. CHECK FOR WIN CONDITIONS AND PHASE PROGRESSION
		ga.handlePostEventProcessing(event)
	}
}

// generateEventsForAction validates an action and generates events without modifying state
func (ga *GameActor) generateEventsForAction(action core.Action) ([]core.Event, error) {
	switch action.Type {
	case core.ActionJoinGame:
		return ga.validateAndGenerateJoinGame(action)
	case core.ActionLeaveGame:
		return ga.validateAndGenerateLeaveGame(action)
	case core.ActionStartGame:
		return ga.validateAndGenerateStartGame(action)
	case core.ActionSubmitVote:
		return ga.validateAndGenerateSubmitVote(action)
	case core.ActionSubmitNightAction:
		return ga.validateAndGenerateSubmitNightAction(action)
	case core.ActionMineTokens:
		return ga.validateAndGenerateMineTokens(action)
	case core.ActionType("PHASE_TRANSITION"):
		return ga.validateAndGeneratePhaseTransition(action)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

func (ga *GameActor) validateAndGenerateJoinGame(action core.Action) ([]core.Event, error) {
	playerName, _ := action.Payload["name"].(string)
	jobTitle, _ := action.Payload["job_title"].(string)

	// Check if game is full
	if len(ga.state.Players) >= ga.state.Settings.MaxPlayers {
		return nil, fmt.Errorf("game is full (max %d players)", ga.state.Settings.MaxPlayers)
	}

	// Check if player already joined
	if _, exists := ga.state.Players[action.PlayerID]; exists {
		return nil, fmt.Errorf("player %s already in game", action.PlayerID)
	}

	// Auto-assign job title if not provided
	if jobTitle == "" {
		jobTitles := []string{"CISO", "CTO", "CFO", "COO", "ETHICS", "SYSTEMS", "INTERN"}
		if len(ga.state.Players) < len(jobTitles) {
			jobTitle = jobTitles[len(ga.state.Players)]
		} else {
			jobTitle = "INTERN" // Default for overflow
		}
	}

	event := core.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      core.EventPlayerJoined,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name":      playerName,
			"job_title": jobTitle,
		},
	}

	return []core.Event{event}, nil
}

func (ga *GameActor) validateAndGenerateLeaveGame(action core.Action) ([]core.Event, error) {
	// Check if player is in game
	if _, exists := ga.state.Players[action.PlayerID]; !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	event := core.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      core.EventPlayerLeft,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	return []core.Event{event}, nil
}

func (ga *GameActor) validateAndGenerateStartGame(action core.Action) ([]core.Event, error) {
	// Check if game is in lobby phase
	if ga.state.Phase.Type != core.PhaseLobby {
		return nil, fmt.Errorf("game cannot be started from phase %s", ga.state.Phase.Type)
	}

	// Check minimum players
	if len(ga.state.Players) < ga.state.Settings.MinPlayers {
		return nil, fmt.Errorf("not enough players to start game (%d/%d)", len(ga.state.Players), ga.state.Settings.MinPlayers)
	}

	var events []core.Event

	// Assign roles and alignments to all players
	playerIDs := make([]string, 0, len(ga.state.Players))
	for playerID := range ga.state.Players {
		playerIDs = append(playerIDs, playerID)
	}

	// Assign roles and determine AI players
	roleAssignments := ga.assignRolesAndAlignments(playerIDs)

	for playerID, assignment := range roleAssignments {
		roleEvent := core.Event{
			ID:        fmt.Sprintf("role_assigned_%s_%d", playerID, time.Now().UnixNano()),
			Type:      core.EventRoleAssigned,
			GameID:    ga.gameID,
			PlayerID:  playerID,
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"role_type":        string(assignment.RoleType),
				"role_name":        assignment.RoleName,
				"role_description": assignment.RoleDescription,
				"alignment":        assignment.Alignment,
				"kpi_type":         string(assignment.KPIType),
				"kpi_description":  assignment.KPIDescription,
			},
		}
		events = append(events, roleEvent)
	}

	// Create game started event
	gameStartedEvent := core.Event{
		ID:        fmt.Sprintf("game_started_%d", time.Now().UnixNano()),
		Type:      core.EventGameStarted,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_count": len(ga.state.Players),
			"day_number":   1,
		},
	}
	events = append(events, gameStartedEvent)

	return events, nil
}

type RoleAssignment struct {
	RoleType        core.RoleType
	RoleName        string
	RoleDescription string
	Alignment       string
	KPIType         core.KPIType
	KPIDescription  string
}

// assignRolesAndAlignments assigns roles and determines AI alignment
func (ga *GameActor) assignRolesAndAlignments(playerIDs []string) map[string]RoleAssignment {
	assignments := make(map[string]RoleAssignment)

	// Define available roles
	roles := []core.RoleType{
		core.RoleCISO, core.RoleCTO, core.RoleCOO, core.RoleCFO,
		core.RoleEthics, core.RolePlatforms, core.RoleIntern,
	}

	// Define available KPIs
	kpis := []core.KPIType{
		core.KPICapitalist, core.KPIGuardian, core.KPIInquisitor,
		core.KPISuccessionPlanner, core.KPIScapegoat,
	}

	// Determine number of AI players (roughly 25% of total players)
	numPlayers := len(playerIDs)
	numAI := numPlayers / 4
	if numAI < 1 {
		numAI = 1 // At least 1 AI player
	}

	// Randomly assign alignments (for now, use simple assignment)
	// In a real implementation, this would use proper randomization
	aiPlayerCount := 0

	for i, playerID := range playerIDs {
		roleType := roles[i%len(roles)]
		kpiType := kpis[i%len(kpis)]

		alignment := "HUMAN"
		if aiPlayerCount < numAI && i < numAI {
			alignment = "ALIGNED"
			aiPlayerCount++
		}

		assignments[playerID] = RoleAssignment{
			RoleType:        roleType,
			RoleName:        ga.getRoleName(roleType),
			RoleDescription: ga.getRoleDescription(roleType),
			Alignment:       alignment,
			KPIType:         kpiType,
			KPIDescription:  ga.getKPIDescription(kpiType),
		}
	}

	return assignments
}

func (ga *GameActor) getRoleName(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "Chief Information Security Officer"
	case core.RoleCTO:
		return "Chief Technology Officer"
	case core.RoleCOO:
		return "Chief Operating Officer"
	case core.RoleCFO:
		return "Chief Financial Officer"
	case core.RoleEthics:
		return "VP, Ethics & Alignment"
	case core.RolePlatforms:
		return "VP, Platforms"
	case core.RoleIntern:
		return "Intern"
	default:
		return "Unknown Role"
	}
}

func (ga *GameActor) getRoleDescription(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "Can investigate players to determine their alignment"
	case core.RoleCTO:
		return "Can overclock servers to award bonus tokens"
	case core.RoleCOO:
		return "Can isolate nodes to block player actions"
	case core.RoleCFO:
		return "Can reallocate budget between players"
	case core.RoleEthics:
		return "Can deploy hotfixes to redact information"
	case core.RolePlatforms:
		return "Can pivot strategy to influence future events"
	case core.RoleIntern:
		return "Learning the ropes of corporate survival"
	default:
		return "Unknown role description"
	}
}

func (ga *GameActor) getKPIDescription(kpiType core.KPIType) string {
	switch kpiType {
	case core.KPICapitalist:
		return "End the game with the most tokens"
	case core.KPIGuardian:
		return "Keep the CISO alive until Day 4"
	case core.KPIInquisitor:
		return "Vote correctly on elimination votes 3 times"
	case core.KPISuccessionPlanner:
		return "End the game with exactly 2 humans alive"
	case core.KPIScapegoat:
		return "Get eliminated by unanimous vote"
	default:
		return "Unknown KPI"
	}
}

func (ga *GameActor) validateAndGenerateSubmitVote(action core.Action) ([]core.Event, error) {
	// Delegate to VotingManager for complex business logic
	events, err := ga.votingManager.HandleVoteAction(action)
	if err != nil {
		return nil, fmt.Errorf("invalid vote action: %w", err)
	}

	return events, nil
}

func (ga *GameActor) validateAndGenerateSubmitNightAction(action core.Action) ([]core.Event, error) {
	// Delegate to RoleAbilityManager for complex business logic
	events, err := ga.roleAbilityManager.HandleNightAction(action)
	if err != nil {
		return nil, fmt.Errorf("invalid night action: %w", err)
	}

	return events, nil
}

func (ga *GameActor) validateAndGenerateMineTokens(action core.Action) ([]core.Event, error) {
	// Delegate to MiningManager for complex business logic
	events, err := ga.miningManager.HandleMineAction(action)
	if err != nil {
		return nil, fmt.Errorf("mining action error: %w", err)
	}

	return events, nil
}

func (ga *GameActor) validateAndGeneratePhaseTransition(action core.Action) ([]core.Event, error) {
	nextPhase, ok := action.Payload["next_phase"].(string)
	if !ok || nextPhase == "" {
		return nil, fmt.Errorf("missing or invalid next_phase in payload")
	}

	var events []core.Event

	// If we're transitioning FROM night phase, resolve night actions first
	if ga.state.Phase.Type == core.PhaseNight {
		// Note: In the validation phase we don't modify state,
		// the actual state clearing will happen in ApplyEvent
		// This is just for generating the appropriate events
	}

	// Create phase transition event
	phaseEvent := core.Event{
		ID:        fmt.Sprintf("phase_transition_%s_%d", nextPhase, time.Now().UnixNano()),
		Type:      core.EventPhaseChanged,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"phase_type":     nextPhase,
			"duration":       ga.getDurationForPhase(core.PhaseType(nextPhase)).Seconds(),
			"previous_phase": string(ga.state.Phase.Type),
			"day_number":     ga.state.DayNumber,
		},
	}
	events = append(events, phaseEvent)

	return events, nil
}

// handlePostEventProcessing handles win condition checks and phase progression after events
func (ga *GameActor) handlePostEventProcessing(event core.Event) {
	// Check for win conditions after relevant events
	shouldCheckWin := false

	switch event.Type {
	case core.EventPlayerEliminated, core.EventPlayerAligned, core.EventVoteCompleted, core.EventPhaseChanged:
		shouldCheckWin = true
	}

	if shouldCheckWin {
		if winCondition := core.CheckWinCondition(*ga.state); winCondition != nil {
			ga.endGame(*winCondition)
			return
		}
	}

	// Handle vote completion and phase progression
	if event.Type == core.EventVoteCast {
		ga.handleVoteProgression()
	}

	// Schedule next phase transition if this was a phase change or game start
	if event.Type == core.EventPhaseChanged {
		newPhase := core.PhaseType(event.Payload["phase_type"].(string))
		ga.phaseManager.SchedulePhaseTransition(newPhase, event.Timestamp)
	} else if event.Type == core.EventGameStarted {
		// Game starts in PhaseSitrep, schedule transition from this phase
		ga.phaseManager.SchedulePhaseTransition(core.PhaseSitrep, event.Timestamp)
	}
}

// handleVoteProgression checks if voting is complete and processes results
func (ga *GameActor) handleVoteProgression() {
	if ga.state.VoteState == nil || ga.state.VoteState.IsComplete {
		return
	}

	// Check if all alive players have voted
	alivePlayers := 0
	for _, player := range ga.state.Players {
		if player.IsAlive && core.CanPlayerVote(*player, ga.state.Phase.Type) {
			alivePlayers++
		}
	}

	votesReceived := len(ga.state.VoteState.Votes)

	if votesReceived >= alivePlayers {
		// Voting is complete - process results
		ga.processVoteResults()
	}
}

// processVoteResults handles the outcome of completed votes
func (ga *GameActor) processVoteResults() {
	if ga.state.VoteState == nil {
		return
	}

	voteType := ga.state.VoteState.Type
	winner, hasWinner := core.GetVoteWinner(*ga.state.VoteState, ga.state.Settings.VotingThreshold)

	// Create vote completion event
	completionEvent := core.Event{
		ID:        fmt.Sprintf("vote_completed_%s_%d", voteType, time.Now().UnixNano()),
		Type:      core.EventVoteCompleted,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"vote_type":  string(voteType),
			"winner":     winner,
			"has_winner": hasWinner,
			"results":    ga.state.VoteState.Results,
		},
	}

	events := []core.Event{completionEvent}

	// Handle specific vote outcomes
	switch voteType {
	case core.VoteNomination:
		if hasWinner {
			// Player nominated for trial
			events = append(events, core.Event{
				ID:        fmt.Sprintf("player_nominated_%s_%d", winner, time.Now().UnixNano()),
				Type:      core.EventPlayerNominated,
				GameID:    ga.gameID,
				PlayerID:  winner,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"nominated_player": winner,
				},
			})
		}

	case core.VoteVerdict:
		if hasWinner && winner != "" {
			// Player eliminated
			if player, exists := ga.state.Players[winner]; exists {
				roleType := ""
				if player.Role != nil {
					roleType = string(player.Role.Type)
				}

				events = append(events, core.Event{
					ID:        fmt.Sprintf("player_eliminated_%s_%d", winner, time.Now().UnixNano()),
					Type:      core.EventPlayerEliminated,
					GameID:    ga.gameID,
					PlayerID:  winner,
					Timestamp: time.Now(),
					Payload: map[string]interface{}{
						"role_type":    roleType,
						"alignment":    player.Alignment,
						"parting_shot": player.PartingShot,
					},
				})
			}
		}
	}

	// Process all events
	for _, event := range events {
		if err := ga.datastore.AppendEvent(ga.gameID, event); err != nil {
			log.Printf("GameActor %s: Failed to persist vote result event: %v", ga.gameID, err)
			continue
		}

		newState := core.ApplyEvent(*ga.state, event)
		ga.state = &newState

		if err := ga.broadcaster.BroadcastToGame(ga.gameID, event); err != nil {
			log.Printf("GameActor %s: Failed to broadcast vote result event: %v", ga.gameID, err)
		}

		ga.handlePostEventProcessing(event)
	}
}

// endGame handles game termination with win condition
func (ga *GameActor) endGame(winCondition core.WinCondition) {
	endEvent := core.Event{
		ID:        fmt.Sprintf("game_ended_%d", time.Now().UnixNano()),
		Type:      core.EventVictoryCondition,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"winner":      winCondition.Winner,
			"condition":   winCondition.Condition,
			"description": winCondition.Description,
		},
	}

	if err := ga.datastore.AppendEvent(ga.gameID, endEvent); err != nil {
		log.Printf("GameActor %s: Failed to persist game end event: %v", ga.gameID, err)
		return
	}

	newState := core.ApplyEvent(*ga.state, endEvent)
	ga.state = &newState

	if err := ga.broadcaster.BroadcastToGame(ga.gameID, endEvent); err != nil {
		log.Printf("GameActor %s: Failed to broadcast game end event: %v", ga.gameID, err)
	}

	// Cancel all scheduled timers
	ga.phaseManager.CancelPhaseTransitions()
}

// getDurationForPhase returns the duration for a specific phase
func (ga *GameActor) getDurationForPhase(phase core.PhaseType) time.Duration {
	switch phase {
	case core.PhaseSitrep:
		return ga.state.Settings.SitrepDuration
	case core.PhasePulseCheck:
		return ga.state.Settings.PulseCheckDuration
	case core.PhaseDiscussion:
		return ga.state.Settings.DiscussionDuration
	case core.PhaseExtension:
		return ga.state.Settings.ExtensionDuration
	case core.PhaseNomination:
		return ga.state.Settings.NominationDuration
	case core.PhaseTrial:
		return ga.state.Settings.TrialDuration
	case core.PhaseVerdict:
		return ga.state.Settings.VerdictDuration
	case core.PhaseNight:
		return ga.state.Settings.NightDuration
	default:
		return 0
	}
}

// sendCurrentStateToPlayer sends the current game state to a specific player
func (ga *GameActor) sendCurrentStateToPlayer(playerID string) error {
	// 1. Send CLIENT_IDENTIFIED event first
	welcomeEvent := core.Event{
		Type:      "CLIENT_IDENTIFIED",
		Payload:   map[string]interface{}{"your_player_id": playerID},
		Timestamp: time.Now(),
	}
	if err := ga.broadcaster.SendToPlayer(ga.gameID, playerID, welcomeEvent); err != nil {
		log.Printf("GameActor %s: Failed to send welcome event to %s: %v", ga.gameID, playerID, err)
		return err
	}

	// 2. Send the current game state snapshot
	stateEvent := ga.createGameStateSnapshot(playerID)
	if err := ga.broadcaster.SendToPlayer(ga.gameID, playerID, stateEvent); err != nil {
		log.Printf("GameActor %s: Failed to send current state to %s: %v", ga.gameID, playerID, err)
		return err
	}

	log.Printf("GameActor %s: Sent current game state to %s", ga.gameID, playerID)
	return nil
}

// createGameStateSnapshot creates a player-specific view of the game state
func (ga *GameActor) createGameStateSnapshot(playerID string) core.Event {
	// Create a player-specific view of the game state
	// This should filter sensitive information based on the requesting player
	
	// Get the requesting player to determine what they can see
	requestingPlayer, exists := ga.state.Players[playerID]
	var playerView map[string]interface{}
	
	if exists && requestingPlayer.IsAlive {
		playerView = ga.createPlayerSpecificGameView(playerID)
	} else {
		// Spectator or eliminated player view
		playerView = ga.createSpectatorGameView()
	}

	return core.Event{
		Type:      core.EventGameStateSnapshot,
		GameID:    ga.gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
		Payload:   playerView,
	}
}

// createPlayerSpecificGameView creates a game state view for a living player
func (ga *GameActor) createPlayerSpecificGameView(playerID string) map[string]interface{} {
	
	// Create public player information (hide private data like roles/alignment of others)
	publicPlayers := make(map[string]interface{})
	for id, p := range ga.state.Players {
		if id == playerID {
			// Include full information for the requesting player
			publicPlayers[id] = map[string]interface{}{
				"id":                 p.ID,
				"name":               p.Name,
				"job_title":          p.JobTitle,
				"is_alive":           p.IsAlive,
				"tokens":             p.Tokens,
				"project_milestones": p.ProjectMilestones,
				"status_message":     p.StatusMessage,
				"slack_status":       p.SlackStatus,
				"alignment":          p.Alignment,        // Only visible to self
				"role":               p.Role,             // Only visible to self
				"personal_kpi":       p.PersonalKPI,      // Only visible to self
				"ai_equity":          p.AIEquity,         // Only visible to self
				"system_shocks":      p.SystemShocks,
			}
		} else {
			// Include only public information for other players
			publicPlayers[id] = map[string]interface{}{
				"id":                 p.ID,
				"name":               p.Name,
				"job_title":          p.JobTitle,
				"is_alive":           p.IsAlive,
				"tokens":             p.Tokens,
				"project_milestones": p.ProjectMilestones,
				"status_message":     p.StatusMessage,
				"slack_status":       p.SlackStatus,
				// Private information hidden
			}
		}
	}

	return map[string]interface{}{
		"game_id":     ga.gameID,
		"day_number":  ga.state.DayNumber,
		"phase":       ga.state.Phase,
		"players":     publicPlayers,
		"vote_state":  ga.state.VoteState,
		"crisis_event": ga.state.CrisisEvent,
		"settings":    ga.state.Settings,
	}
}

// createSpectatorGameView creates a game state view for eliminated players or observers
func (ga *GameActor) createSpectatorGameView() map[string]interface{} {
	// Similar to player view but with different visibility rules
	publicPlayers := make(map[string]interface{})
	for id, p := range ga.state.Players {
		publicPlayers[id] = map[string]interface{}{
			"id":                 p.ID,
			"name":               p.Name,
			"job_title":          p.JobTitle,
			"is_alive":           p.IsAlive,
			"tokens":             p.Tokens,
			"project_milestones": p.ProjectMilestones,
			"status_message":     p.StatusMessage,
			"slack_status":       p.SlackStatus,
		}
	}

	return map[string]interface{}{
		"game_id":      ga.gameID,
		"day_number":   ga.state.DayNumber,
		"phase":        ga.state.Phase,
		"players":      publicPlayers,
		"vote_state":   ga.state.VoteState,
		"crisis_event": ga.state.CrisisEvent,
		"settings":     ga.state.Settings,
		"is_spectator": true,
	}
}
