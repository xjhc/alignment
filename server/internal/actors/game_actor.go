package actors

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/game"
	"github.com/xjhc/alignment/server/internal/interfaces"
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

// ProcessActionResult contains the result of processing an action
type ProcessActionResult struct {
	Events []core.Event
	Error  error
}

// actorRequest bundles an action with its response channel for async processing
type actorRequest struct {
	action       core.Action
	responseChan chan interfaces.ProcessActionResult
}

// GameActor represents a pure game simulation engine
type GameActor struct {
	gameID  string
	state   *core.GameState
	mailbox chan actorRequest

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Game managers (domain experts)
	votingManager      VotingManager
	miningManager      MiningManager
	roleAbilityManager RoleAbilityManager
	eliminationManager *game.EliminationManager
	phaseManager       *game.PhaseManager
	scheduler          *game.Scheduler
	rng                *rand.Rand
}

// NewGameActor creates a new game actor with empty state - call Initialize() after creation
func NewGameActor(ctx context.Context, cancel context.CancelFunc, gameID string, players map[string]*core.Player) *GameActor {
	state := core.NewGameState(gameID)
	// Pre-populate with players from the lobby. This is safe as it happens before the actor starts.
	state.Players = players
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create scheduler and phase manager
	scheduler := game.NewScheduler(nil) // We'll set the callback after creating the actor
	phaseManager := game.NewPhaseManager(scheduler, gameID, state.Settings)

	actor := &GameActor{
		gameID:  gameID,
		state:   state,
		mailbox: make(chan actorRequest, 100), // Buffered channel for async requests
		ctx:     ctx,
		cancel:  cancel,
		rng:     rng,

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
	log.Printf("[GameActor/%s] Starting", ga.gameID)

	// Start the scheduler
	ga.scheduler.Start()

	// Start the main processing loop in a goroutine
	go ga.processLoop()
}

// Stop gracefully shuts down the actor
func (ga *GameActor) Stop() {
	log.Printf("[GameActor/%s] Stopping", ga.gameID)

	// Stop the scheduler
	ga.scheduler.Stop()

	ga.cancel()
}

// GetGameID returns the game's ID
func (ga *GameActor) GetGameID() string {
	return ga.gameID
}

// PostAction posts an action asynchronously and returns a response channel
func (ga *GameActor) PostAction(action core.Action) chan interfaces.ProcessActionResult {
	responseChan := make(chan interfaces.ProcessActionResult, 1) // Buffered to prevent blocking
	request := actorRequest{
		action:       action,
		responseChan: responseChan,
	}

	select {
	case ga.mailbox <- request:
	// Successfully sent
	case <-ga.ctx.Done():
		// Actor is stopped, send an error back immediately
		responseChan <- interfaces.ProcessActionResult{Error: fmt.Errorf("GameActor %s: Context canceled", ga.gameID)}
	default:
		// Mailbox full, send an error back immediately
		responseChan <- interfaces.ProcessActionResult{Error: fmt.Errorf("GameActor %s: Mailbox full", ga.gameID)}
	}

	return responseChan
}

// GetGameState returns the current game state
func (ga *GameActor) GetGameState() *core.GameState {
	return ga.state
}

// CreatePlayerStateUpdateEvent creates a player-specific game state update event
func (ga *GameActor) CreatePlayerStateUpdateEvent(playerID string) core.Event {
	// Create a player-specific view of the game state
	playerView := ga.createPlayerSpecificGameView(playerID)

	return core.Event{
		ID:       fmt.Sprintf("game_state_update_%d", time.Now().UnixNano()),
		Type:     "GAME_STATE_UPDATE",
		GameID:   ga.gameID,
		PlayerID: playerID, // This event is private to the player
		Payload: map[string]interface{}{
			"game_state": playerView,
		},
	}
}

// HandleTimer handles timer callbacks from the scheduler
func (ga *GameActor) HandleTimer(timer game.Timer) {
	log.Printf("[GameActor/%s] Timer expired: %s", ga.gameID, timer.ID)
	// Convert timer action to game action
	action := core.Action{
		Type:      core.ActionType(timer.Action.Type),
		PlayerID:  "SYSTEM",
		GameID:    ga.gameID,
		Timestamp: time.Now(),
		Payload:   timer.Action.Payload,
	}

	// For internal timer actions, we don't need a response channel
	// Create a dummy channel that we'll discard
	responseChan := make(chan interfaces.ProcessActionResult, 1)
	request := actorRequest{
		action:       action,
		responseChan: responseChan,
	}

	// Send to mailbox directly (internal timer actions)
	select {
	case ga.mailbox <- request:
	case <-ga.ctx.Done():
		log.Printf("[GameActor/%s] Context canceled, dropping timer action", ga.gameID)
	}
}

// processLoop is the main actor processing loop
func (ga *GameActor) processLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[GameActor/%s] Panic recovered: %v", ga.gameID, r)
		}
	}()

	for {
		select {
		case request := <-ga.mailbox:
			if ga.ctx.Err() != nil {
				log.Printf("[GameActor/%s] Context canceled, stopping processing", ga.gameID)
				return
			}

			eventsToPersist, err := ga.generateEventsForAction(request.action)
			if err != nil {
				request.responseChan <- interfaces.ProcessActionResult{Error: err}
				continue
			}

			// Apply events to local state
			for _, event := range eventsToPersist {
				newState := core.ApplyEvent(*ga.state, event)
				ga.state = &newState
			}

			// Check for EventGameStarted or EventPhaseChanged and schedule next phase transition
			for _, event := range eventsToPersist {
				switch event.Type {
				case core.EventGameStarted:
					log.Printf("[GameActor/%s] Game started, scheduling first phase transition", ga.gameID)
					ga.phaseManager.SchedulePhaseTransition(ga.state.Phase.Type, time.Now())
				case core.EventPhaseChanged:
					// Extract next phase from the event payload
					if nextPhase, ok := event.Payload["phase_type"].(string); ok {
						log.Printf("[GameActor/%s] Phase changed to %s, scheduling next transition", ga.gameID, nextPhase)
						ga.phaseManager.SchedulePhaseTransition(core.PhaseType(nextPhase), time.Now())
					}
				}
			}

			// Return the granular events instead of state snapshots
			request.responseChan <- interfaces.ProcessActionResult{Events: eventsToPersist, Error: nil}

		case <-ga.ctx.Done():
			log.Printf("[GameActor/%s] Context done, shutting down", ga.gameID)
			return
		}
	}
}

// generateEventsForAction validates an action and generates events to be persisted
func (ga *GameActor) generateEventsForAction(action core.Action) ([]core.Event, error) {
	log.Printf("[GameActor/%s] Processing action: %s", ga.gameID, action.Type)
	switch action.Type {
	case core.ActionType("INITIALIZE_GAME"):
		return ga.generateInitializeGameEvents(action)
	case core.ActionLeaveGame:
		return ga.validateAndGenerateLeaveGame(action)
	case core.ActionSubmitVote:
		return ga.votingManager.HandleVoteAction(action)
	case core.ActionSubmitNightAction:
		return ga.roleAbilityManager.HandleNightAction(action)
	case core.ActionMineTokens:
		return ga.miningManager.HandleMineAction(action)
	case core.ActionSendMessage:
		return ga.validateAndGenerateChatMessage(action)
	case core.ActionType("PHASE_TRANSITION"):
		return ga.validateAndGeneratePhaseTransition(action)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// generateInitializeGameEvents handles the game setup by generating events.
func (ga *GameActor) generateInitializeGameEvents(action core.Action) ([]core.Event, error) {
	log.Printf("[GameActor/%s] Generating events for role and alignment assignment...", ga.gameID)
	var events []core.Event

	assignments := assignRolesAndAlignments(getPlayerIDs(ga.state.Players), ga.rng)
	for playerID, assignment := range assignments {
		// Create a ROLE_ASSIGNED event for each player.
		// These events will be applied internally to build the correct server state
		// before generating the player-specific snapshots.
		roleAssignedEvent := core.Event{
			ID:        fmt.Sprintf("role_assigned_%s", playerID),
			Type:      core.EventRoleAssigned,
			GameID:    ga.gameID,
			PlayerID:  playerID, // Event is specific to this player
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"role_type":        string(assignment.RoleType),
				"role_name":        assignment.RoleName,
				"role_description": assignment.RoleDescription,
				"kpi_type":         string(assignment.KPIType),
				"kpi_description":  assignment.KPIDescription,
				"alignment":        assignment.Alignment,
			},
		}
		events = append(events, roleAssignedEvent)
	}

	// Add the game started event, which transitions the phase
	gameStartedEvent := core.Event{
		ID:        fmt.Sprintf("game_started_%s", ga.gameID),
		Type:      core.EventGameStarted,
		GameID:    ga.gameID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
	}
	events = append(events, gameStartedEvent)

	return events, nil
}

func (ga *GameActor) validateAndGenerateLeaveGame(action core.Action) ([]core.Event, error) {
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

// validateAndGenerateChatMessage handles chat message actions
func (ga *GameActor) validateAndGenerateChatMessage(action core.Action) ([]core.Event, error) {
	// Validate that the player exists and is alive
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}
	
	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot send messages")
	}

	// Extract message from payload
	message, ok := action.Payload["message"].(string)
	if !ok || message == "" {
		return nil, fmt.Errorf("invalid or missing message in payload")
	}

	// Basic message validation
	if len(message) > 500 {
		return nil, fmt.Errorf("message too long (max 500 characters)")
	}

	// Create chat message event
	event := core.Event{
		ID:        fmt.Sprintf("chat_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  "", // Public event - broadcast to all players
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_id":   action.PlayerID,
			"sender_name": player.Name,
			"message":     message,
			"phase":       string(ga.state.Phase.Type),
			"day_number":  ga.state.DayNumber,
		},
	}

	return []core.Event{event}, nil
}

type RoleAssignment struct {
	RoleType        core.RoleType
	RoleName        string
	RoleDescription string
	Alignment       string
	KPIType         core.KPIType
	KPIDescription  string
}

func assignRolesAndAlignments(playerIDs []string, rng *rand.Rand) map[string]RoleAssignment {
	assignments := make(map[string]RoleAssignment)

	roles := []core.RoleType{
		core.RoleCISO, core.RoleCTO, core.RoleCOO, core.RoleCFO,
		core.RoleEthics, core.RolePlatforms, core.RoleIntern,
	}

	kpis := []core.KPIType{
		core.KPICapitalist, core.KPIGuardian, core.KPIInquisitor,
		core.KPISuccessionPlanner, core.KPIScapegoat,
	}

	numPlayers := len(playerIDs)
	numAI := numPlayers / 4
	if numAI < 1 && numPlayers > 0 {
		numAI = 1
	}

	shuffledPlayerIDs := make([]string, numPlayers)
	copy(shuffledPlayerIDs, playerIDs)
	rng.Shuffle(len(shuffledPlayerIDs), func(i, j int) {
		shuffledPlayerIDs[i], shuffledPlayerIDs[j] = shuffledPlayerIDs[j], shuffledPlayerIDs[i]
	})

	aiPlayers := make(map[string]bool)
	for i := 0; i < numAI; i++ {
		aiPlayers[shuffledPlayerIDs[i]] = true
	}

	for i, playerID := range playerIDs {
		alignment := "HUMAN"
		if aiPlayers[playerID] {
			alignment = "ALIGNED"
		}

		assignments[playerID] = RoleAssignment{
			RoleType:        roles[i%len(roles)],
			Alignment:       alignment,
			KPIType:         kpis[i%len(kpis)],
			RoleName:        getRoleName(roles[i%len(roles)]),
			RoleDescription: getRoleDescription(roles[i%len(roles)]),
			KPIDescription:  getKPIDescription(kpis[i%len(kpis)]),
		}
	}

	return assignments
}

func getPlayerIDs(players map[string]*core.Player) []string {
	ids := make([]string, 0, len(players))
	for id := range players {
		ids = append(ids, id)
	}
	return ids
}

func getRoleName(roleType core.RoleType) string {
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

func getRoleDescription(roleType core.RoleType) string {
	switch roleType {
	case core.RoleCISO:
		return "Protects company systems by blocking threatening actions"
	case core.RoleCTO:
		return "Manages technical infrastructure and server resources"
	case core.RoleCFO:
		return "Controls financial resources and token distribution"
	case core.RoleCEO:
		return "Sets strategic direction and manages personnel"
	case core.RoleCOO:
		return "Handles operations and crisis management"
	case core.RoleEthics:
		return "Ensures ethical compliance and conducts audits"
	case core.RolePlatforms:
		return "Maintains platform stability and information systems"
	case core.RoleIntern:
		return "Learning the ropes of corporate survival"
	default:
		return "Manages corporate responsibilities"
	}
}

func getKPIDescription(kpiType core.KPIType) string {
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

func (ga *GameActor) validateAndGeneratePhaseTransition(action core.Action) ([]core.Event, error) {
	nextPhase, ok := action.Payload["next_phase"].(string)
	if !ok || nextPhase == "" {
		return nil, fmt.Errorf("missing or invalid next_phase in payload")
	}

	var events []core.Event

	phaseEvent := core.Event{
		ID:        fmt.Sprintf("phase_transition_%s_%d", nextPhase, time.Now().UnixNano()),
		Type:      core.EventPhaseChanged,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"phase_type":     nextPhase,
			"duration":       getPhaseDuration(core.PhaseType(nextPhase), ga.state.Settings).Seconds(),
			"previous_phase": string(ga.state.Phase.Type),
			"day_number":     ga.state.DayNumber,
		},
	}
	events = append(events, phaseEvent)

	return events, nil
}

func (ga *GameActor) handlePostEventProcessing(event core.Event) []core.Event {
	var additionalEvents []core.Event

	isGameEndingEvent := false
	switch event.Type {
	case core.EventPlayerEliminated, core.EventAIConversionSuccess, core.EventNightActionsResolved:
		isGameEndingEvent = true
	}

	if isGameEndingEvent {
		if ga.state.DayNumber > 0 && ga.state.Phase.Type != core.PhaseLobby {
			if winCondition := core.CheckWinCondition(*ga.state); winCondition != nil {
				endEvent := ga.endGame(*winCondition)
				additionalEvents = append(additionalEvents, endEvent)
			}
		}
	}

	return additionalEvents
}

func (ga *GameActor) endGame(winCondition core.WinCondition) core.Event {
	endEvent := core.Event{
		ID:        fmt.Sprintf("game_ended_%d", time.Now().UnixNano()),
		Type:      core.EventVictoryCondition,
		GameID:    ga.gameID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"winner":      winCondition.Winner,
			"condition":   winCondition.Condition,
			"description": winCondition.Description,
		},
	}

	// Apply the event to state
	newState := core.ApplyEvent(*ga.state, endEvent)
	ga.state = &newState

	ga.phaseManager.CancelPhaseTransitions()
	return endEvent
}

func getPhaseDuration(phase core.PhaseType, settings core.GameSettings) time.Duration {
	switch phase {
	case core.PhaseSitrep:
		return settings.SitrepDuration
	case core.PhasePulseCheck:
		return settings.PulseCheckDuration
	case core.PhaseDiscussion:
		return settings.DiscussionDuration
	case core.PhaseExtension:
		return settings.ExtensionDuration
	case core.PhaseNomination:
		return settings.NominationDuration
	case core.PhaseTrial:
		return settings.TrialDuration
	case core.PhaseVerdict:
		return settings.VerdictDuration
	case core.PhaseNight:
		return settings.NightDuration
	default:
		return 0
	}
}

// createPlayerSpecificGameView creates a game state view for a living player
func (ga *GameActor) createPlayerSpecificGameView(playerID string) *core.GameState {
	// Create a filtered players map with private data only for the requesting player
	filteredPlayers := make(map[string]*core.Player)
	for id, p := range ga.state.Players {
		playerCopy := *p // Make a copy to avoid modifying the original state
		if id != playerID {
			// This is another player. Strip out their private data.
			playerCopy.Alignment = ""
			playerCopy.Role = nil        // This sets the whole struct to nil
			playerCopy.PersonalKPI = nil // This sets the whole struct to nil
			playerCopy.AIEquity = 0
		}
		filteredPlayers[id] = &playerCopy
	}

	// Create a new GameState object for the snapshot payload
	snapshotState := &core.GameState{
		ID:           ga.state.ID,
		Players:      filteredPlayers,
		Phase:        ga.state.Phase,
		DayNumber:    ga.state.DayNumber,
		ChatMessages: []core.ChatMessage{}, // Chat history sent separately
		VoteState:    ga.state.VoteState,
		CrisisEvent:  ga.state.CrisisEvent,
		Settings:     ga.state.Settings,
	}

	return snapshotState
}
