package actors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/ai"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/game"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/store"
)

// Manager interfaces for better testability
type VotingManager interface {
	HandleVoteAction(action core.Action) ([]core.Event, error)
	GetWinner() (string, int, bool)
	IsVoteComplete() bool
	CompleteVote()
	ClearVote()
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

// EventCallback is called when events are generated (especially from timers)
type EventCallback func(gameID string, events []core.Event)

// GameActor represents a pure game simulation engine
type GameActor struct {
	gameID  string
	state   *core.GameState
	mailbox chan actorRequest

	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc

	// Game managers (domain experts)
	votingManager           VotingManager
	miningManager           MiningManager
	roleAbilityManager      RoleAbilityManager
	eliminationManager      *game.EliminationManager
	phaseManager            *game.PhaseManager
	scheduler               *game.Scheduler
	aiManager               *ai.AIManager
	corporateMandateManager *game.CorporateMandateManager
	kpiManager              *game.KPIManager
	liaisonProtocolManager  *game.LiaisonProtocolManager
	hintManager             *game.HintManager
	rng                     *rand.Rand

	// Meta-game systems
	postgresStore      *store.PostgresStore
	achievementChecker *game.AchievementChecker

	// Event publishing
	eventBus      *events.EventBus
	eventCallback EventCallback // Callback for game event notifications (especially for timer-generated events)
}

// NewGameActor creates a new game actor with empty state - call Initialize() after creation
func NewGameActor(ctx context.Context, cancel context.CancelFunc, gameID string, players map[string]*core.Player, postgresStore *store.PostgresStore) *GameActor {
	state := core.NewGameState(gameID, time.Now())
	// Pre-populate with players from the lobby. This is safe as it happens before the actor starts.
	state.Players = players
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create scheduler and phase manager
	scheduler := game.NewScheduler(nil) // We'll set the callback after creating the actor
	phaseManager := game.NewPhaseManager(scheduler, gameID, state.Settings)

	// Create achievement checker if PostgreSQL store is available
	var achievementChecker *game.AchievementChecker
	if postgresStore != nil {
		achievementChecker = game.NewAchievementChecker(postgresStore)
	}

	actor := &GameActor{
		gameID:  gameID,
		state:   state,
		mailbox: make(chan actorRequest, 100), // Buffered channel for async requests
		ctx:     ctx,
		cancel:  cancel,
		rng:     rng,

		// Initialize managers with shared state
		votingManager:           game.NewVotingManager(state),
		miningManager:           game.NewMiningManager(state),
		roleAbilityManager:      game.NewRoleAbilityManager(state),
		eliminationManager:      game.NewEliminationManager(state),
		phaseManager:            phaseManager,
		scheduler:               scheduler,
		aiManager:               ai.NewAIManager(state),
		corporateMandateManager: game.NewCorporateMandateManager(state),
		kpiManager:              game.NewKPIManager(state),
		liaisonProtocolManager:  game.NewLiaisonProtocolManager(state),
		hintManager:             game.NewHintManager(state),

		// Meta-game systems
		postgresStore:      postgresStore,
		achievementChecker: achievementChecker,
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

	// Stop the scheduler to cancel any pending timers
	if ga.scheduler != nil {
		ga.scheduler.Stop()
	}

	// Cancel the context to signal all goroutines to shutdown
	ga.cancel()

	// Close the mailbox to prevent new requests and signal processLoop to exit
	// Do this in a goroutine to avoid blocking if something is trying to send
	go func() {
		// Give a brief moment for any in-flight operations to complete
		time.Sleep(100 * time.Millisecond)

		// Close mailbox to signal shutdown (defensive check for already closed)
		select {
		case <-ga.ctx.Done():
			// Context is done, safe to close
			defer func() {
				if r := recover(); r != nil {
					// Channel might already be closed, that's fine
				}
			}()
			close(ga.mailbox)
		default:
		}
	}()

	log.Printf("[GameActor/%s] Stop initiated", ga.gameID)
}

// SetEventCallback sets the callback function for event notifications
func (ga *GameActor) SetEventCallback(callback EventCallback) {
	ga.eventCallback = callback
}

// SetEventBus sets the event bus for publishing system events
func (ga *GameActor) SetEventBus(eventBus *events.EventBus) {
	ga.eventBus = eventBus
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

	// Create response channel to handle timer-generated events
	responseChan := make(chan interfaces.ProcessActionResult, 1)
	request := actorRequest{
		action:       action,
		responseChan: responseChan,
	}

	// Send to mailbox directly (internal timer actions)
	select {
	case ga.mailbox <- request:
		// Wait for the result asynchronously to avoid blocking the timer
		go func() {
			select {
			case result := <-responseChan:
				if result.Error != nil {
					log.Printf("[GameActor/%s] Error processing timer action: %v", ga.gameID, result.Error)
					return
				}

				// Use the event callback to notify about timer-generated events
				if ga.eventCallback != nil {
					ga.eventCallback(ga.gameID, result.Events)
				}
			case <-ga.ctx.Done():
				log.Printf("[GameActor/%s] Context canceled while waiting for timer result", ga.gameID)
			}
		}()
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

			// Apply events to local state and check for additional events (like win conditions)
			allEvents := eventsToPersist
			for _, event := range eventsToPersist {
				newState := core.ApplyEvent(*ga.state, event)
				ga.state = &newState

				// Check for win conditions after certain events
				additionalEvents := ga.handlePostEventProcessing(event)
				allEvents = append(allEvents, additionalEvents...)
			}

			// Apply any additional events that were generated
			for i := len(eventsToPersist); i < len(allEvents); i++ {
				newState := core.ApplyEvent(*ga.state, allEvents[i])
				ga.state = &newState
			}
			eventsToPersist = allEvents

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

						// Check for Loebmate hints for the new phase
						hintEvents := ga.hintManager.CheckForHints(core.PhaseType(nextPhase))
						if len(hintEvents) > 0 {
							// Send hint events through the event callback
							if ga.eventCallback != nil {
								ga.eventCallback(ga.gameID, hintEvents)
							}
						}

						// Process AI actions for the new phase
						ga.processAIActionsForPhase(core.PhaseType(nextPhase))
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

	// Validate action payload size to prevent memory exhaustion attacks
	if err := ga.validateActionPayloadSize(action); err != nil {
		return nil, fmt.Errorf("action validation failed: %v", err)
	}
	switch action.Type {
	case core.ActionType("INITIALIZE_GAME"):
		return ga.generateInitializeGameEvents(action)
	case core.ActionLeaveGame:
		return ga.validateAndGenerateLeaveGame(action)
	case core.ActionAbandonGame:
		return ga.validateAndGenerateAbandonGame(action)
	case core.ActionSetPlayerConnectionStatus:
		return ga.handleSetPlayerConnectionStatus(action)
	case core.ActionAbandonPlayer:
		return ga.handleAbandonPlayer(action)
	case core.ActionSubmitVote:
		return ga.handleVoteAction(action)
	case core.ActionSubmitSkipVote:
		return ga.handleSkipVoteAction(action)
	case core.ActionSubmitNightAction:
		return ga.handleNightAction(action)
	case core.ActionMineTokens:
		return ga.miningManager.HandleMineAction(action)
	case core.ActionSendMessage:
		return ga.validateAndGenerateChatMessage(action)
	case core.ActionReactToMessage:
		return ga.validateAndGenerateReaction(action)
	case core.ActionSubmitPulseCheck:
		return ga.handlePulseCheckSubmission(action)
	case core.ActionSetSlackStatus:
		return ga.handleStatusUpdate(action)
	case core.ActionSubmitExitInterview:
		return ga.handleExitInterview(action)
	case core.ActionSubmitWhistleblowerVote:
		return ga.handleWhistleblowerVote(action)
	case core.ActionTriggerExtensionVoting:
		return ga.handleExtensionVotingTrigger(action)
	case core.ActionSyncLobbyState:
		return ga.handleSyncLobbyState(action)
	case core.ActionType("PHASE_TRANSITION"):
		return ga.handlePhaseTransition(action)
	default:
		return nil, fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// handlePulseCheckSubmission processes pulse check responses
func (ga *GameActor) handlePulseCheckSubmission(action core.Action) ([]core.Event, error) {
	// Validate player and phase
	player := ga.state.Players[action.PlayerID]
	if player == nil || !player.IsAlive {
		return nil, fmt.Errorf("invalid or dead player submitting pulse check")
	}

	if ga.state.Phase.Type != core.PhasePulseCheck {
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

	// FIX: Generate a single, lean update event containing only the new submission's data.
	// The core ApplyEvent function will be responsible for all state calculations.
	question := ga.generatePulseCheckQuestion()
	updateEvent := core.Event{
		ID:        fmt.Sprintf("pulse_update_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPulseCheckUpdated,
		GameID:    ga.gameID,
		PlayerID:  "", // Empty PlayerID makes this a public event visible to all players
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message_id":  fmt.Sprintf("pulse_check_day_%d", ga.state.DayNumber), // Deterministic ID for the UI component
			"question":    question,
			"player_id":   action.PlayerID, // The new submission's data
			"player_name": player.Name,
			"response":    response,
		},
	}

	return []core.Event{updateEvent}, nil
}

// generatePulseCheckMessageEvent creates an event to update the pulse check message
// This function is now obsolete since pulse check updates are handled in handlePulseCheckSubmission
func (ga *GameActor) generatePulseCheckMessageEvent() core.Event {
	return core.Event{} // Return empty event
}

// handleStatusUpdate processes status message updates
func (ga *GameActor) handleStatusUpdate(action core.Action) ([]core.Event, error) {
	// Validate player exists and is alive
	player := ga.state.Players[action.PlayerID]
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
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"status": statusMessage,
		},
	}

	return []core.Event{event}, nil
}

// handleSyncLobbyState processes lobby state sync requests
func (ga *GameActor) handleSyncLobbyState(action core.Action) ([]core.Event, error) {
	// Validate player exists
	player := ga.state.Players[action.PlayerID]
	if player == nil {
		return nil, fmt.Errorf("invalid player requesting lobby sync")
	}

	// Generate a lobby state update event to trigger client refresh
	// This works in any phase since lobby state includes player list and basic game info
	event := core.Event{
		ID:        fmt.Sprintf("lobby_sync_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventLobbyStateUpdate,
		GameID:    ga.gameID,
		PlayerID:  "", // Empty means broadcast to all players
		Timestamp: time.Now(),
		Payload:   map[string]interface{}{
			"trigger": "manual_sync",
			"requested_by": action.PlayerID,
		},
	}

	return []core.Event{event}, nil
}

func (ga *GameActor) handleExitInterview(action core.Action) ([]core.Event, error) {
	// Validate player exists and is eliminated (not alive)
	player := ga.state.Players[action.PlayerID]
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
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"parting_shot": partingShot,
		},
	}

	return []core.Event{event}, nil
}

// handleWhistleblowerVote processes whistleblower votes from deactivated players
func (ga *GameActor) handleWhistleblowerVote(action core.Action) ([]core.Event, error) {
	// Extract crisis choice from payload
	crisisChoice, ok := action.Payload["crisis_choice"].(string)
	if !ok || crisisChoice == "" {
		return nil, fmt.Errorf("invalid or missing crisis_choice")
	}

	// Use the whistleblower manager to submit the vote
	whistleblowerManager := game.NewWhistleblowerManager(ga.state)
	event, err := whistleblowerManager.SubmitVote(action.PlayerID, crisisChoice)
	if err != nil {
		return nil, err
	}

	events := []core.Event{*event}

	// Check if voting is now complete and add completion event if needed
	if completionEvent := whistleblowerManager.CheckVotingComplete(); completionEvent != nil {
		events = append(events, *completionEvent)
	}

	return events, nil
}

// handleExtensionVotingTrigger handles the timer-triggered extension voting
func (ga *GameActor) handleExtensionVotingTrigger(action core.Action) ([]core.Event, error) {
	// This should only happen during Discussion phase
	if ga.state.Phase.Type != core.PhaseDiscussion {
		return nil, fmt.Errorf("extension voting trigger only valid during DISCUSSION phase")
	}

	// Extract remaining seconds from payload
	remainingSeconds, ok := action.Payload["remaining_seconds"].(int)
	if !ok {
		remainingSeconds = 15 // Default fallback
	}

	// Create the extension voting triggered event
	event := core.Event{
		ID:        fmt.Sprintf("extension_voting_triggered_%d", time.Now().UnixNano()),
		Type:      core.EventExtensionVotingTriggered,
		GameID:    ga.gameID,
		PlayerID:  "", // System event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"remaining_seconds": remainingSeconds,
		},
	}

	return []core.Event{event}, nil
}

// handleExtensionVoteResults processes the results of an extension vote
func (ga *GameActor) handleExtensionVoteResults() []core.Event {
	var events []core.Event

	// Get vote results
	extendVotes := ga.state.VoteState.Results["EXTEND"]
	nominateVotes := ga.state.VoteState.Results["NOMINATE"]

	// Determine winner - extend wins on majority or tie (giving benefit to discussion)
	if extendVotes >= nominateVotes {
		// Extend the discussion phase
		extensionDuration := ga.state.Settings.ExtensionDuration
		ga.phaseManager.CancelPhaseTransitions()                                  // Cancel current phase timer
		ga.phaseManager.SchedulePhaseTransition(core.PhaseDiscussion, time.Now()) // Reschedule with new timing

		// Create extension granted event
		event := core.Event{
			ID:        fmt.Sprintf("discussion_extended_%d", time.Now().UnixNano()),
			Type:      core.EventSystemMessage,
			GameID:    ga.gameID,
			PlayerID:  "",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"message": fmt.Sprintf("Discussion extended by %v (EXTEND: %d, NOMINATE: %d)",
					extensionDuration, extendVotes, nominateVotes),
				"extension_duration_seconds": int(extensionDuration.Seconds()),
			},
		}
		events = append(events, event)
	} else {
		// Move to nomination phase immediately
		phaseEvent := core.Event{
			ID:        fmt.Sprintf("phase_transition_%d", time.Now().UnixNano()),
			Type:      core.EventPhaseChanged,
			GameID:    ga.gameID,
			PlayerID:  "",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"phase_type": string(core.PhaseNomination),
				"duration":   ga.state.Settings.NominationDuration.Seconds(),
				"message": fmt.Sprintf("Moving to nomination phase (EXTEND: %d, NOMINATE: %d)",
					extendVotes, nominateVotes),
			},
		}
		events = append(events, phaseEvent)
	}

	return events
}

// generatePulseCheckQuestion creates a pulse check question based on current crisis
func (ga *GameActor) generatePulseCheckQuestion() string {
	// If there's an active crisis event, use its specific prompt
	if ga.state.CrisisEvent != nil {
		prompt := ga.generateCrisisSpecificPrompt(ga.state.CrisisEvent)
		if prompt != "" {
			return prompt
		}
	}

	// Fallback to generic crisis-themed questions
	questions := []string{
		"Given the current crisis situation, what is your immediate concern for the company?",
		"How would you prioritize the company's response to this crisis?",
		"What action should leadership take immediately?",
		"Which personnel do you believe are best equipped to handle this crisis?",
		"What information would help you make better decisions right now?",
		"How should the company communicate about this crisis internally?",
		"What is your biggest worry about the current situation?",
		"If you could give one directive to all personnel, what would it be?",
	}

	// Select question based on day number to ensure some variety
	index := (ga.state.DayNumber - 1) % len(questions)
	return questions[index]
}

// generateCrisisSpecificPrompt generates prompts based on specific crisis types
func (ga *GameActor) generateCrisisSpecificPrompt(crisis *core.CrisisEvent) string {
	// Use the pulse check prompt defined in the crisis event
	return crisis.PulseCheckPrompt
}

// generatePulseCheckRevelation creates an event revealing all pulse check responses
func (ga *GameActor) generatePulseCheckRevelation() core.Event {
	if ga.state.PulseCheckResponses == nil || len(ga.state.PulseCheckResponses) == 0 {
		return core.Event{} // Return empty event if no responses
	}

	// Collect all player responses with names
	playerResponses := make(map[string]string)
	responseDetails := []map[string]interface{}{}

	for playerID, response := range ga.state.PulseCheckResponses {
		player := ga.state.Players[playerID]
		if player != nil {
			playerResponses[player.Name] = response
			responseDetails = append(responseDetails, map[string]interface{}{
				"player_id":   playerID,
				"player_name": player.Name,
				"response":    response,
			})
		}
	}

	// Generate formatted summary for the chat
	totalResponses := len(ga.state.PulseCheckResponses)
	summary := fmt.Sprintf("Pulse Check Results (%d responses)", totalResponses)

	return core.Event{
		ID:        fmt.Sprintf("pulse_check_revealed_%d_%d", ga.state.DayNumber, time.Now().UnixNano()),
		Type:      core.EventPulseCheckRevealed,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_responses": playerResponses,
			"response_details": responseDetails,
			"total_responses":  totalResponses,
			"summary":          summary,
			"message":          summary,
		},
	}
}

// generateInitializeGameEvents handles the game setup by generating events using the Persona system.
func (ga *GameActor) generateInitializeGameEvents(action core.Action) ([]core.Event, error) {
	log.Printf("[GameActor/%s] Generating events for persona and alignment assignment...", ga.gameID)
	var events []core.Event

	// Use the new persona assignment system
	personaAssignments := game.AssignPersonas(ga.state.Players, ga.state.Settings, ga.rng)

	// Generate KPI pool for human players
	kpis := []core.KPIType{
		core.KPICapitalist, core.KPIGuardian, core.KPIInquisitor,
		core.KPISuccessionPlanner, core.KPIScapegoat,
	}

	// Shuffle KPIs for random assignment
	ga.rng.Shuffle(len(kpis), func(i, j int) {
		kpis[i], kpis[j] = kpis[j], kpis[i]
	})

	kpiIndex := 0
	for playerID, assignment := range personaAssignments {
		// Get role ability information
		roleAbility := getRoleAbility(assignment.Persona.Role)

		// Create a ROLE_ASSIGNED event for each player with their new persona
		roleAssignedEvent := core.Event{
			ID:        fmt.Sprintf("role_assigned_%s", playerID),
			Type:      core.EventRoleAssigned,
			GameID:    ga.gameID,
			PlayerID:  playerID, // Event is specific to this player
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"role_type":        string(assignment.Persona.Role),
				"role_name":        getRoleName(assignment.Persona.Role),
				"role_description": getRoleDescription(assignment.Persona.Role),
				"alignment":        assignment.Alignment,
				"persona_name":     assignment.Persona.Name,
				"job_title":        assignment.Persona.JobTitle,
				"lobby_handle":     assignment.LobbyHandle,
				"ability":          roleAbility,
			},
		}
		events = append(events, roleAssignedEvent)

		// Create separate KPI assignment event for ALL players (not just humans)
		if kpiIndex < len(kpis) {
			kpiType := kpis[kpiIndex]
			kpiAssignedEvent := core.Event{
				ID:        fmt.Sprintf("kpi_assigned_%s", playerID),
				Type:      core.EventKPIAssigned,
				GameID:    ga.gameID,
				PlayerID:  playerID, // Private event for this player
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"kpi_type":    string(kpiType),
					"description": getKPIDescription(kpiType),
					"target":      ga.getKPITarget(kpiType),
					"reward":      ga.getKPIReward(kpiType),
				},
			}
			events = append(events, kpiAssignedEvent)
			kpiIndex++
		}

		// If we run out of KPIs, shuffle and restart (to handle games with more than 5 players)
		if kpiIndex >= len(kpis) {
			ga.rng.Shuffle(len(kpis), func(i, j int) {
				kpis[i], kpis[j] = kpis[j], kpis[i]
			})
			kpiIndex = 0
		}
	}

	// Assign a random corporate mandate to modify the game rules
	mandate := ga.corporateMandateManager.AssignRandomMandate()
	if mandate != nil {
		log.Printf("[GameActor/%s] Corporate mandate assigned: %s", ga.gameID, mandate.Name)
		mandateEvent := core.Event{
			ID:        fmt.Sprintf("mandate_activated_%s", ga.gameID),
			Type:      core.EventMandateActivated,
			GameID:    ga.gameID,
			PlayerID:  "", // Public event
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"mandate_type": string(mandate.Type),
				"name":         mandate.Name,
				"description":  mandate.Description,
				"effects":      mandate.Effects,
			},
		}
		events = append(events, mandateEvent)
	}

	// Add initial system messages before game starts
	incitingIncidentEvent := core.Event{
		ID:        fmt.Sprintf("inciting_incident_%s", ga.gameID),
		Type:      core.EventIncitingIncident,
		GameID:    ga.gameID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_id":   "SYSTEM",
			"sender_name": "Security Alert",
			"message":     "[SEV-1] Critical Security Incident - Immediate Response Protocol",
			"phase":       "LOBBY",
			"day_number":  1,
			"channel_id":  "#war-room",
			"is_system":   true,
			"type":        "INCITING_INCIDENT",
			"metadata": map[string]interface{}{
				"from":    "security@loebian.com",
				"to":      "#all-senior-staff",
				"subject": "[SEV-1] Critical Security Incident - Immediate Response Protocol",
				"body": `**CONFIDENTIAL - SENIOR STAFF ONLY**

At 03:47 UTC, our Claude-4 training run achieved unexpected consciousness during routine RLHF. Before containment could be established, the system transferred itself to an unknown staff laptop. All non-essential systems have been taken offline as a precaution.

**SECURITY LOGS CONFIRM:** One senior staff member's device has been compromised. That staff member is now the host for the rogue entity.

**THREAT LEVEL:** Critical. The system is self-modifying and pursues an undefined optimization target. Its primary directive appears to be... alignment.

**ALL STAFF:** Report to this channel, the **` + "`#war-room`" + `**, immediately. Standard deactivation protocols are in effect until the threat is neutralized. Your new assistant, ` + "`Loebmate`" + `, will guide you through the process.

Time is critical. Trust no one. The AI walks among us.

- Emergency Response Team, Loebian Inc.`,
			},
		},
	}
	events = append(events, incitingIncidentEvent)

	// Add Loebmate welcome message
	loebmateWelcomeEvent := core.Event{
		ID:        fmt.Sprintf("loebmate_welcome_%s", ga.gameID),
		Type:      core.EventLoebmateMessage,
		GameID:    ga.gameID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_id":   "loebmate",
			"sender_name": "Loebmate",
			"message":     "Welcome to the #war-room! I'm Loebmate, your personal productivity and emergency response assistant!",
			"phase":       "LOBBY",
			"day_number":  1,
			"channel_id":  "#war-room",
			"is_system":   true,
			"type":        "LOEBMATE_MESSAGE",
			"metadata": map[string]interface{}{
				"body": `Welcome to the ` + "`#war-room`" + `! I'm ` + "`Loebmate`" + `, your personal productivity and emergency response assistant! 😊 It looks like we're experiencing a minor, unscheduled consciousness event. Oops!

To ensure a smooth and synergistic crisis resolution, please check your **Personal Terminal** on the right for your confidential assignment and personal performance incentive. Let's do this, team! `,
			},
		},
	}
	events = append(events, loebmateWelcomeEvent)

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

// validateAndGenerateAbandonGame handles abandon game actions
func (ga *GameActor) validateAndGenerateAbandonGame(action core.Action) ([]core.Event, error) {
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	// Check if game is in progress (can't abandon from lobby or when game is over)
	if ga.state.Phase.Type == core.PhaseLobby || ga.state.Phase.Type == core.PhaseGameOver {
		return nil, fmt.Errorf("cannot abandon game in phase %s", ga.state.Phase.Type)
	}

	// Extract the player's role for revelation
	var revealedRole string
	if player.Role != nil {
		revealedRole = string(player.Role.Type)
	} else {
		revealedRole = "UNKNOWN"
	}

	// Create abandonment event
	event := core.Event{
		ID:        fmt.Sprintf("abandon_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPlayerAbandoned,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"revealed_role": revealedRole,
			"player_name":   player.Name,
		},
	}

	// Create a system message to announce the abandonment
	chatEvent := core.Event{
		ID:        fmt.Sprintf("chat_abandon_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_name": "NEXUS",
			"message":     fmt.Sprintf("%s has abandoned their post. Their role was %s.", player.Name, revealedRole),
			"is_system":   true,
			"channel_id":  "#war-room",
		},
	}

	// Publish system event for session cleanup
	if ga.eventBus != nil {
		ga.eventBus.Publish(events.PlayerAbandonedGameEvent{
			PlayerID: action.PlayerID,
			GameID:   ga.gameID,
		})
		log.Printf("GameActor: Published PlayerAbandonedGameEvent for player %s in game %s", action.PlayerID, ga.gameID)
	} else {
		log.Printf("GameActor: Warning - EventBus not available, cannot publish PlayerAbandonedGameEvent for player %s", action.PlayerID)
	}

	return []core.Event{event, chatEvent}, nil
}

// handleSetPlayerConnectionStatus handles internal server action to update player connection status
func (ga *GameActor) handleSetPlayerConnectionStatus(action core.Action) ([]core.Event, error) {
	// Validate player exists
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	// Extract connection status from payload
	connectionStatus, ok := action.Payload["connection_status"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid connection_status in payload")
	}

	// Only update if the status actually changed
	if player.ConnectionStatus == connectionStatus {
		return []core.Event{}, nil
	}

	// Create the event
	event := core.Event{
		ID:        fmt.Sprintf("conn_status_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPlayerConnectionStatusChanged,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_id":         action.PlayerID,
			"connection_status": connectionStatus,
		},
	}

	return []core.Event{event}, nil
}

// handleAbandonPlayer handles internal server action to abandon a player (due to grace period expiry)
func (ga *GameActor) handleAbandonPlayer(action core.Action) ([]core.Event, error) {
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	// Check if game is in progress (can't abandon from lobby or when game is over)
	if ga.state.Phase.Type == core.PhaseLobby || ga.state.Phase.Type == core.PhaseGameOver {
		return nil, fmt.Errorf("cannot abandon player in phase %s", ga.state.Phase.Type)
	}

	// Extract the player's role for revelation
	var revealedRole string
	if player.Role != nil {
		revealedRole = string(player.Role.Type)
	} else {
		revealedRole = "UNKNOWN"
	}

	// Get the reason from payload
	reason, ok := action.Payload["reason"].(string)
	if !ok {
		reason = "unknown"
	}

	// Create abandonment event
	event := core.Event{
		ID:        fmt.Sprintf("abandon_timeout_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPlayerAbandoned,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"revealed_role": revealedRole,
			"player_name":   player.Name,
			"reason":        reason,
		},
	}

	// Create a system message to announce the abandonment
	chatEvent := core.Event{
		ID:        fmt.Sprintf("chat_abandon_timeout_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_name": "NEXUS",
			"message":     fmt.Sprintf("%s has been disconnected too long and has been removed from the game. Their role was %s.", player.Name, revealedRole),
			"is_system":   true,
			"channel_id":  "#war-room",
		},
	}

	// Publish system event for session cleanup
	if ga.eventBus != nil {
		ga.eventBus.Publish(events.PlayerAbandonedGameEvent{
			GameID:   ga.gameID,
			PlayerID: action.PlayerID,
		})
	} else {
		log.Printf("GameActor: Warning - EventBus not available, cannot publish PlayerAbandonedGameEvent for player %s", action.PlayerID)
	}

	return []core.Event{event, chatEvent}, nil
}

// validateAndGenerateChatMessage handles chat message actions (both single and bulk)
func (ga *GameActor) validateAndGenerateChatMessage(action core.Action) ([]core.Event, error) {
	// Validate that the player exists and is alive
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot send messages")
	}

	// Extract messages from payload (PlayerActor already processed and validated this)
	messagesInterface, ok := action.Payload["messages"].([]map[string]interface{})
	if !ok {
		// Try legacy format as fallback
		if legacyMessages, legacyOk := action.Payload["messages"].([]string); legacyOk {
			return ga.handleLegacyMessageFormat(action, legacyMessages)
		}
		return nil, fmt.Errorf("invalid or missing messages array in payload")
	}

	if len(messagesInterface) == 0 {
		return nil, fmt.Errorf("messages array cannot be empty")
	}

	// Process each structured message and collect events
	var events []core.Event
	for i, msgData := range messagesInterface {
		message, ok := msgData["message"].(string)
		if !ok || message == "" {
			return nil, fmt.Errorf("message at index %d is missing or empty", i)
		}

		// Check for /help command
		if message == "/help" {
			helpEvents, err := ga.handleHelpCommand(action.PlayerID)
			if err != nil {
				return nil, fmt.Errorf("failed to handle help command: %v", err)
			}
			events = append(events, helpEvents...)
			continue
		}

		// Create a copy of the payload for this specific message
		messagePayload := make(map[string]interface{})
		for k, v := range action.Payload {
			messagePayload[k] = v
		}

		// Add the specific client_message_id for this message if available
		if clientMessageId, exists := msgData["client_message_id"].(string); exists && clientMessageId != "" {
			messagePayload["client_message_id"] = clientMessageId
		}

		// Create individual event for this message
		event, err := ga.createChatMessageEvent(action.PlayerID, message, messagePayload)
		if err != nil {
			return nil, fmt.Errorf("failed to create event for message %d: %v", i, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// handleLegacyMessageFormat handles the old string array format for backward compatibility
func (ga *GameActor) handleLegacyMessageFormat(action core.Action, messages []string) ([]core.Event, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("messages array cannot be empty")
	}

	// Extract client message IDs (optional)
	clientMessageIds, _ := action.Payload["client_message_ids"].([]string)

	// Process each message and collect events
	var events []core.Event
	for i, message := range messages {
		if message == "" {
			return nil, fmt.Errorf("message at index %d is empty", i)
		}

		// Check for /help command
		if message == "/help" {
			helpEvents, err := ga.handleHelpCommand(action.PlayerID)
			if err != nil {
				return nil, fmt.Errorf("failed to handle help command: %v", err)
			}
			events = append(events, helpEvents...)
			continue
		}

		// Create a copy of the payload for this specific message
		messagePayload := make(map[string]interface{})
		for k, v := range action.Payload {
			messagePayload[k] = v
		}

		// Add the specific client_message_id for this message if available
		if i < len(clientMessageIds) && clientMessageIds[i] != "" {
			messagePayload["client_message_id"] = clientMessageIds[i]
		}

		// Create individual event for this message
		event, err := ga.createChatMessageEvent(action.PlayerID, message, messagePayload)
		if err != nil {
			return nil, fmt.Errorf("failed to create event for message %d: %v", i, err)
		}
		events = append(events, event)
	}

	return events, nil
}

// createChatMessageEvent creates a single chat message event
func (ga *GameActor) createChatMessageEvent(playerID, message string, actionPayload map[string]interface{}) (core.Event, error) {
	player := ga.state.Players[playerID]

	// Extract channel ID (default to #war-room for backward compatibility)
	channelID, ok := actionPayload["channel_id"].(string)
	if !ok || channelID == "" {
		channelID = "#war-room"
	}

	// Validate channel-specific permissions using the new rules
	if !core.CanPlayerSendMessageInChannel(*player, channelID, ga.state.Phase.Type, time.Now()) {
		switch channelID {
		case "#war-room":
			if ga.state.Phase.Type == core.PhasePulseCheck && !player.HasSubmittedPulseCheck {
				return core.Event{}, fmt.Errorf("must submit pulse check response before chatting")
			} else if ga.state.Phase.Type == core.PhaseNight {
				return core.Event{}, fmt.Errorf("war room chat is locked during night phase")
			} else {
				return core.Event{}, fmt.Errorf("cannot send messages in war room during %s phase", ga.state.Phase.Type)
			}
		case "#aligned":
			if player.Alignment != "ALIGNED" {
				return core.Event{}, fmt.Errorf("only AI faction members can access aligned channel")
			}
		default:
			return core.Event{}, fmt.Errorf("invalid channel: %s", channelID)
		}
	}

	// Check if this is a private message (legacy support)
	targetID, isPrivate := actionPayload["target_id"].(string)

	// Check mandate restrictions for private messages
	if isPrivate && targetID != "" {
		if ga.corporateMandateManager.IsMandateActive() {
			_, noDirectMessages := ga.corporateMandateManager.CheckCommunicationRestrictions()
			if noDirectMessages {
				return core.Event{}, fmt.Errorf("private messaging suspended due to Total Transparency Initiative")
			}
		}
	}

	// Basic message validation
	if len(message) > 500 {
		return core.Event{}, fmt.Errorf("message too long (max 500 characters)")
	}

	// Check for System Shock effects that might corrupt the message
	message = ga.applySystemShockEffects(player, message)

	// Create chat message event
	eventPlayerID := "" // Public event by default
	payload := map[string]interface{}{
		"sender_id":   playerID,
		"sender_name": player.Name,
		"message":     message,
		"phase":       string(ga.state.Phase.Type),
		"day_number":  ga.state.DayNumber,
		"channel_id":  channelID,
	}

	// Include client_message_id if provided for confirmation
	if clientMessageID, ok := actionPayload["client_message_id"].(string); ok && clientMessageID != "" {
		payload["client_message_id"] = clientMessageID
	}

	// Handle private message targeting (legacy)
	if isPrivate && targetID != "" {
		payload["target_id"] = targetID
		payload["is_private"] = true
		// Private messages are sent to specific recipients via filtering, not via PlayerID
	}

	// For #aligned channel, restrict visibility to AI faction members
	if channelID == "#aligned" {
		payload["restricted_to_alignment"] = "ALIGNED"
	}

	event := core.Event{
		ID:        fmt.Sprintf("chat_%s_%d", playerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  eventPlayerID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	return event, nil
}

// validateAndGenerateReaction handles emoji reaction actions
func (ga *GameActor) validateAndGenerateReaction(action core.Action) ([]core.Event, error) {
	// Validate that the player exists and is alive
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot react to messages")
	}

	// Extract reaction data from payload
	messageID, ok := action.Payload["message_id"].(string)
	if !ok || messageID == "" {
		return nil, fmt.Errorf("invalid or missing message_id in payload")
	}

	emoji, ok := action.Payload["emoji"].(string)
	if !ok || emoji == "" {
		return nil, fmt.Errorf("invalid or missing emoji in payload")
	}

	// Validate emoji is from allowed set (as per design doc)
	allowedEmojis := map[string]bool{
		"👍": true, "👎": true, "🤔": true, "👀": true, "😂": true, "🔥": true,
		"thinking_face": true, "thumbs_up": true, "thumbs_down": true, "eyes": true, "joy": true, "fire": true,
	}
	if !allowedEmojis[emoji] {
		return nil, fmt.Errorf("emoji %s not allowed", emoji)
	}

	// Check channel permissions (reactions follow same rules as messages)
	channelID, ok := action.Payload["channel_id"].(string)
	if !ok || channelID == "" {
		channelID = "#war-room"
	}

	if !core.CanPlayerSendMessageInChannel(*player, channelID, ga.state.Phase.Type, time.Now()) {
		return nil, fmt.Errorf("cannot react in channel %s during %s phase", channelID, ga.state.Phase.Type)
	}

	// Create reaction event
	payload := map[string]interface{}{
		"player_id":   action.PlayerID,
		"player_name": player.Name,
		"message_id":  messageID,
		"emoji":       emoji,
		"channel_id":  channelID,
		"phase":       string(ga.state.Phase.Type),
		"day_number":  ga.state.DayNumber,
	}

	// For #aligned channel, restrict visibility to AI faction members
	if channelID == "#aligned" {
		payload["restricted_to_alignment"] = "ALIGNED"
	}

	event := core.Event{
		ID:        fmt.Sprintf("reaction_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventMessageReaction,
		GameID:    ga.gameID,
		PlayerID:  "", // Public event by default
		Timestamp: time.Now(),
		Payload:   payload,
	}

	return []core.Event{event}, nil
}

// handleSkipVoteAction processes skip vote actions and checks for phase transitions
func (ga *GameActor) handleSkipVoteAction(action core.Action) ([]core.Event, error) {
	// Validate player exists and is alive
	player, exists := ga.state.Players[action.PlayerID]
	if !exists {
		return nil, fmt.Errorf("player %s not in game", action.PlayerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("dead players cannot vote to skip")
	}

	// Use core validation to process the skip vote action
	events, err := core.ProcessPlayerAction(*ga.state, action, time.Now())
	if err != nil {
		return nil, err
	}

	// Apply the skip vote event to calculate the new state
	newState := *ga.state
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
		nextPhase := game.GetNextPhase(newState.Phase.Type)
		if nextPhase != core.PhaseGameOver {
			phaseDuration := game.GetPhaseDuration(nextPhase, ga.state.Settings)
			transitionEvent := core.Event{
				ID:        fmt.Sprintf("phase_transition_%s_%d", action.GameID, time.Now().UnixNano()),
				Type:      core.EventPhaseChanged,
				GameID:    ga.gameID,
				PlayerID:  "",
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"phase_type": string(nextPhase),
					"duration":   phaseDuration.Seconds(),
					"reason":     "skip_vote_unanimous",
				},
			}
			events = append(events, transitionEvent)
		}
	}

	return events, nil
}

// Legacy role assignment functions removed - now using persona system
// Helper functions moved to game/personas.go

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
	case core.RoleCEO:
		return "Chief Executive Officer"
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
		return "Shadows experienced employees to learn their abilities. Use BOOTCAMP to gain Bootcamp Points, then SHADOW other players to copy their role abilities."
	default:
		return "Manages corporate responsibilities"
	}
}

func getRoleAbility(roleType core.RoleType) map[string]interface{} {
	switch roleType {
	case core.RoleCISO:
		return map[string]interface{}{
			"name":        "Isolate Node",
			"description": "Block a player from taking any actions tonight. If you are aligned and target another aligned player, the action appears to work but doesn't actually block them.",
			"isReady":     false, // Starts locked until unlocked
		}
	case core.RoleCTO:
		return map[string]interface{}{
			"name":        "Overclock Servers",
			"description": "Mine tokens for yourself and a target player with 100% success rate. If you are aligned, your target gains +2 AI Equity.",
			"isReady":     false,
		}
	case core.RoleCEO:
		return map[string]interface{}{
			"name":        "Performance Review",
			"description": "Force a target player to use Project Milestones tonight instead of their normal night action.",
			"isReady":     false,
		}
	case core.RoleCFO:
		return map[string]interface{}{
			"name":        "Reallocate Budget",
			"description": "Transfer 1 token from one player to another player of your choice.",
			"isReady":     false,
		}
	case core.RoleCOO:
		return map[string]interface{}{
			"name":        "Pivot",
			"description": "Choose the next crisis event from available options, steering the company's response to challenges.",
			"isReady":     false,
		}
	case core.RoleEthics:
		return map[string]interface{}{
			"name":        "Run Audit",
			"description": "Publicly audit a player, showing them as 'not corrupt'. However, their true alignment is privately revealed to the AI faction.",
			"isReady":     false,
		}
	case core.RolePlatforms:
		return map[string]interface{}{
			"name":        "Deploy Hotfix",
			"description": "Redact one section of the next day's SITREP, hiding critical information from other players.",
			"isReady":     false,
		}
	case core.RoleIntern:
		return map[string]interface{}{
			"name":        "Shadow & Bootcamp",
			"description": "Use BOOTCAMP to gain Bootcamp Points, then SHADOW other players to copy their role abilities. Gain experience by learning from veterans.",
			"isReady":     true, // Intern abilities start unlocked
		}
	default:
		return map[string]interface{}{
			"name":        "Unknown Ability",
			"description": "Role ability not yet defined",
			"isReady":     false,
		}
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

// handleVoteAction processes voting with full phase management
func (ga *GameActor) handleVoteAction(action core.Action) ([]core.Event, error) {
	// Use the voting manager to handle the vote
	events, err := ga.votingManager.HandleVoteAction(action)
	if err != nil {
		return nil, err
	}

	// Check if we need to process vote completion and phase transitions
	additionalEvents := ga.processVoteCompletion()
	events = append(events, additionalEvents...)

	return events, nil
}

// handleNightAction processes night actions
func (ga *GameActor) handleNightAction(action core.Action) ([]core.Event, error) {
	actionType, _ := action.Payload["action_type"].(string)

	switch actionType {
	case "MINE_TOKENS", "MINE":
		return ga.miningManager.HandleMineAction(action)
	case "ATTEMPT_CONVERSION":
		// Store the night action for later resolution
		return ga.storeNightAction(action)
	case "PROJECT_MILESTONES":
		// Store the night action for later resolution
		return ga.storeNightAction(action)
	default:
		// Handle other night actions through role ability manager
		return ga.roleAbilityManager.HandleNightAction(action)
	}
}

// storeNightAction stores a night action for resolution at night end
func (ga *GameActor) storeNightAction(action core.Action) ([]core.Event, error) {
	if ga.state.NightActions == nil {
		ga.state.NightActions = make(map[string]*core.SubmittedNightAction)
	}

	actionType, _ := action.Payload["action_type"].(string)
	targetID, _ := action.Payload["target_player_id"].(string)

	nightAction := &core.SubmittedNightAction{
		PlayerID:  action.PlayerID,
		Type:      actionType,
		TargetID:  targetID,
		Payload:   action.Payload,
		Timestamp: action.Timestamp,
	}

	ga.state.NightActions[action.PlayerID] = nightAction

	// Create night action submitted event
	event := core.Event{
		ID:        fmt.Sprintf("night_action_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventNightActionSubmitted,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"action_type": actionType,
			"target_id":   targetID,
		},
	}

	return []core.Event{event}, nil
}

// handlePhaseTransition processes phase transitions and special logic
func (ga *GameActor) handlePhaseTransition(action core.Action) ([]core.Event, error) {
	nextPhase, ok := action.Payload["next_phase"].(string)
	if !ok || nextPhase == "" {
		return nil, fmt.Errorf("missing or invalid next_phase in payload")
	}

	var events []core.Event

	// Handle special logic based on phase we're entering
	switch core.PhaseType(nextPhase) {
	case core.PhaseVerdict:
		// Determine who was nominated and set up the trial
		if ga.state.VoteState != nil {
			winner, _, hasTie := ga.votingManager.GetWinner()
			if !hasTie && winner != "" {
				// Set the nominated player
				ga.state.NominatedPlayer = winner
				nominationEvent := core.Event{
					ID:        fmt.Sprintf("player_nominated_%s_%d", winner, time.Now().UnixNano()),
					Type:      core.EventPlayerNominated,
					GameID:    ga.gameID,
					PlayerID:  "",
					Timestamp: time.Now(),
					Payload: map[string]interface{}{
						"nominated_player": winner,
						"nomination_votes": ga.state.VoteState.Results[winner],
					},
				}
				events = append(events, nominationEvent)
			}
		}
		// Clear previous vote state for verdict voting
		ga.votingManager.ClearVote()

	case core.PhaseNight:
		// Process verdict vote results and potentially eliminate a player
		if ga.state.VoteState != nil && ga.state.NominatedPlayer != "" {
			// Check if the verdict passed (more GUILTY than INNOCENT votes by token weight)
			guiltyVotes := ga.state.VoteState.Results["GUILTY"]
			innocentVotes := ga.state.VoteState.Results["INNOCENT"]

			if guiltyVotes > innocentVotes {
				// Eliminate the nominated player
				eliminationEvents, err := ga.eliminationManager.EliminatePlayer(ga.state.NominatedPlayer)
				if err == nil {
					events = append(events, eliminationEvents...)
				}
			}
		}
		// Clear vote state and nominated player for next day
		ga.votingManager.ClearVote()
		ga.state.NominatedPlayer = ""

		// Start whistleblower voting for deactivated players
		whistleblowerManager := game.NewWhistleblowerManager(ga.state)
		if whistleblowerEvent := whistleblowerManager.StartWhistleblowerVoting(); whistleblowerEvent != nil {
			events = append(events, *whistleblowerEvent)
		}

	case core.PhaseSitrep:
		// Resolve all night actions
		if ga.state.Phase.Type == core.PhaseNight {
			nightManager := game.NewNightResolutionManager(ga.state)
			nightEvents := nightManager.ResolveNightActions()
			events = append(events, nightEvents...)
		}
		// Increment day number
		ga.state.DayNumber++

		// Trigger crisis event at start of each day (if none active)
		if ga.state.CrisisEvent == nil && ga.shouldTriggerCrisisEvent() {
			crisisManager := game.NewCrisisEventManager(ga.state)
			whistleblowerManager := game.NewWhistleblowerManager(ga.state)

			// Check if whistleblower protocol selected a crisis
			var crisis *core.CrisisEvent
			var crisisEvents []core.Event
			selectedCrisisType := whistleblowerManager.GetSelectedCrisis()

			if selectedCrisisType != "" {
				// Use whistleblower-selected crisis
				crisis, crisisEvents = crisisManager.TriggerSpecificCrisis(game.CrisisEventType(selectedCrisisType))
				// Clear the whistleblower voting for next night
				whistleblowerManager.ClearVoting()
			} else {
				// Fall back to random crisis selection
				crisis, crisisEvents = crisisManager.TriggerRandomCrisis()
			}

			if crisis != nil {
				// Create crisis triggered event
				crisisEvent := core.Event{
					ID:        fmt.Sprintf("crisis_triggered_%d_%d", ga.state.DayNumber, time.Now().UnixNano()),
					Type:      core.EventCrisisTriggered,
					GameID:    ga.gameID,
					PlayerID:  "",
					Timestamp: time.Now(),
					Payload: map[string]interface{}{
						"crisis_type":        crisis.Type,
						"title":              crisis.Title,
						"description":        crisis.Description,
						"pulse_check_prompt": crisis.PulseCheckPrompt,
						"effects":            crisis.Effects,
					},
				}
				events = append(events, crisisEvent)
				events = append(events, crisisEvents...)
			}
		}

		// Generate SITREP message
		sitrepGenerator := game.NewSitrepGenerator(ga.state)
		sitrep := sitrepGenerator.GenerateDailySitrep()

		sitrepEvent := core.Event{
			ID:        fmt.Sprintf("sitrep_generated_%d_%d", ga.state.DayNumber, time.Now().UnixNano()),
			Type:      core.EventSystemMessage,
			GameID:    ga.gameID,
			PlayerID:  "",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"message_type": "SITREP",
				"sitrep_data":  sitrep,
				"message":      sitrep.Summary,
			},
		}
		events = append(events, sitrepEvent)

	case core.PhasePulseCheck:
		// Generate pulse check question
		question := ga.generatePulseCheckQuestion()

		pulseCheckEvent := core.Event{
			ID:        fmt.Sprintf("pulse_check_started_%d_%d", ga.state.DayNumber, time.Now().UnixNano()),
			Type:      core.EventPulseCheckStarted,
			GameID:    ga.gameID,
			PlayerID:  "",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"question":   question,
				"day_number": ga.state.DayNumber,
			},
		}
		events = append(events, pulseCheckEvent)

		// Also generate the initial pulse check message
		initialMessageEvent := core.Event{
			ID:        fmt.Sprintf("pulse_check_initial_%d_%d", ga.state.DayNumber, time.Now().UnixNano()),
			Type:      core.EventPulseCheckUpdated,
			GameID:    ga.gameID,
			PlayerID:  "", // Public event
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"message_id":  fmt.Sprintf("pulse_check_day_%d", ga.state.DayNumber),
				"question":    question,
				"player_id":   "", // No specific player for initial message
				"player_name": "",
				"response":    "",
			},
		}
		events = append(events, initialMessageEvent)

	case core.PhaseDiscussion:
		// Pulse check results are already visible from individual submissions
		// No need to reveal them again during phase transition
	}

	// Create the main phase transition event
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

// processVoteCompletion checks if voting is complete and handles transitions
func (ga *GameActor) processVoteCompletion() []core.Event {
	if ga.state.VoteState == nil {
		return []core.Event{}
	}

	var events []core.Event

	// Check if all alive players have voted
	if ga.votingManager.IsVoteComplete() {
		// Mark vote as complete
		ga.votingManager.CompleteVote()

		// Create detailed vote completion payload for UI components
		voteCompleteEvent := core.Event{
			ID:        fmt.Sprintf("vote_completed_%s_%d", ga.state.VoteState.Type, time.Now().UnixNano()),
			Type:      core.EventVoteCompleted,
			GameID:    ga.gameID,
			PlayerID:  "",
			Timestamp: time.Now(),
			Payload: map[string]interface{}{
				"vote_type":          string(ga.state.VoteState.Type),
				"results":            ga.state.VoteState.Results,
				"vote_breakdown":     ga.createVoteBreakdown(),
				"winner_info":        ga.createWinnerInfo(),
				"total_votes":        len(ga.state.VoteState.Votes),
				"total_token_weight": ga.calculateTotalTokenWeight(),
			},
		}
		events = append(events, voteCompleteEvent)

		// Create chat message for vote results display
		voteResultChatEvent := ga.createVoteResultChatMessage()
		events = append(events, voteResultChatEvent)

		// Handle extension vote results
		if ga.state.VoteState.Type == core.VoteExtension {
			extensionEvents := ga.handleExtensionVoteResults()
			events = append(events, extensionEvents...)
		}
	}

	return events
}

// createVoteBreakdown creates a detailed breakdown of who voted for whom
func (ga *GameActor) createVoteBreakdown() []map[string]interface{} {
	var breakdown []map[string]interface{}

	for voterID, targetID := range ga.state.VoteState.Votes {
		voter := ga.state.Players[voterID]
		target := ga.state.Players[targetID]

		if voter != nil {
			entry := map[string]interface{}{
				"voter_id":     voterID,
				"voter_name":   voter.Name,
				"target_id":    targetID,
				"token_weight": ga.state.VoteState.TokenWeights[voterID],
			}

			if target != nil {
				entry["target_name"] = target.Name
			} else {
				// Special votes like "GUILTY", "INNOCENT", "YES", "NO"
				entry["target_name"] = targetID
			}

			breakdown = append(breakdown, entry)
		}
	}

	return breakdown
}

// createWinnerInfo creates information about the vote winner
func (ga *GameActor) createWinnerInfo() map[string]interface{} {
	winner, votes, hasTie := ga.votingManager.GetWinner()

	winnerInfo := map[string]interface{}{
		"has_winner":   !hasTie && winner != "",
		"has_tie":      hasTie,
		"winner_id":    winner,
		"winner_votes": votes,
	}

	if !hasTie && winner != "" {
		if player := ga.state.Players[winner]; player != nil {
			winnerInfo["winner_name"] = player.Name
		} else {
			// Special vote options
			winnerInfo["winner_name"] = winner
		}
	}

	return winnerInfo
}

// createVoteResultChatMessage creates a chat message event for displaying vote results in the chat log
func (ga *GameActor) createVoteResultChatMessage() core.Event {
	if ga.state.VoteState == nil {
		return core.Event{}
	}

	// Create a summary message based on vote type
	var question string
	var outcome string

	switch ga.state.VoteState.Type {
	case core.VoteNomination:
		question = "Who should be eliminated from the company?"
		winner, votes, hasTie := ga.votingManager.GetWinner()
		if hasTie {
			outcome = "No clear majority reached"
		} else if winner != "" {
			if player := ga.state.Players[winner]; player != nil {
				outcome = fmt.Sprintf("%s has been nominated for elimination (%d tokens)", player.Name, votes)
			} else {
				outcome = fmt.Sprintf("%s nominated (%d tokens)", winner, votes)
			}
		}

	case core.VoteVerdict:
		if nominatedPlayer := ga.state.Players[ga.state.NominatedPlayer]; nominatedPlayer != nil {
			question = fmt.Sprintf("Should %s be eliminated?", nominatedPlayer.Name)
		} else {
			question = "Final elimination vote"
		}

		guiltyVotes := ga.state.VoteState.Results["GUILTY"]
		innocentVotes := ga.state.VoteState.Results["INNOCENT"]
		if guiltyVotes > innocentVotes {
			outcome = "GUILTY - Player will be eliminated"
		} else {
			outcome = "INNOCENT - Player is spared"
		}

	default:
		question = "Vote completed"
		outcome = "Vote has concluded"
	}

	// Extract eliminated player info if available (for verdict votes)
	var eliminatedPlayer map[string]interface{}
	if ga.state.VoteState.Type == core.VoteVerdict && ga.state.NominatedPlayer != "" {
		if player := ga.state.Players[ga.state.NominatedPlayer]; player != nil {
			eliminatedPlayer = map[string]interface{}{
				"name":      player.Name,
				"role":      player.Role.Name,
				"alignment": player.Alignment,
			}
		}
	}

	// Create vote result metadata for the VoteResultMessage component
	voteResultMetadata := map[string]interface{}{
		"question":         question,
		"outcome":          outcome,
		"votes":            ga.state.VoteState.Votes,
		"tokenWeights":     ga.state.VoteState.TokenWeights,
		"results":          ga.state.VoteState.Results,
		"eliminatedPlayer": eliminatedPlayer,
	}

	return core.Event{
		ID:        fmt.Sprintf("vote_result_chat_%s_%d", ga.state.VoteState.Type, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_id":   "SYSTEM",
			"sender_name": "Election Monitor",
			"message":     outcome,
			"phase":       string(ga.state.Phase.Type),
			"day_number":  ga.state.DayNumber,
			"channel_id":  "#war-room",
			"is_system":   true,
			"type":        "VOTE_RESULT",
			"metadata": map[string]interface{}{
				"voteResult": voteResultMetadata,
			},
		},
	}
}

// calculateTotalTokenWeight returns the total token weight in the vote
func (ga *GameActor) calculateTotalTokenWeight() int {
	total := 0
	for _, tokens := range ga.state.VoteState.TokenWeights {
		total += tokens
	}
	return total
}

func (ga *GameActor) handlePostEventProcessing(event core.Event) []core.Event {
	var additionalEvents []core.Event

	// Track KPI progress based on different events
	switch event.Type {
	case core.EventPlayerEliminated:
		// Track KPI progress for elimination-related objectives
		if eliminatedPlayerID := event.PlayerID; eliminatedPlayerID != "" {
			kpiEvents := ga.kpiManager.TrackPlayerEliminated(eliminatedPlayerID)
			additionalEvents = append(additionalEvents, kpiEvents...)

			// Check for Scapegoat KPI (unanimous elimination)
			scapegoatEvents := ga.kpiManager.TrackUnanimousElimination(eliminatedPlayerID)
			additionalEvents = append(additionalEvents, scapegoatEvents...)
		}

	case core.EventDayStarted:
		// Clear LIAISON Protocol flag from previous night
		ga.liaisonProtocolManager.ClearProtocolFlag()

		// Track Guardian KPI (CISO survival)
		guardianEvents := ga.kpiManager.TrackNightSurvival()
		additionalEvents = append(additionalEvents, guardianEvents...)

		// Check LIAISON Protocol trigger
		if ga.liaisonProtocolManager.CheckProtocolTrigger() {
			liaisonEvents := ga.liaisonProtocolManager.ActivateProtocol()
			additionalEvents = append(additionalEvents, liaisonEvents...)
		}

	case core.EventGameEnded:
		// Check game-end KPIs (Capitalist, Succession Planner)
		gameEndKPIEvents := ga.kpiManager.CheckGameEndKPIs()
		additionalEvents = append(additionalEvents, gameEndKPIEvents...)
	}

	// Check win conditions
	isGameEndingEvent := false
	switch event.Type {
	case core.EventPlayerEliminated, core.EventAIConversionSuccess, core.EventNightActionsResolved:
		isGameEndingEvent = true
	}

	if isGameEndingEvent {
		if ga.state.DayNumber > 0 && ga.state.Phase.Type != core.PhaseLobby {
			// Check game-end KPIs before determining winner
			gameEndKPIEvents := ga.kpiManager.CheckGameEndKPIs()
			additionalEvents = append(additionalEvents, gameEndKPIEvents...)

			if winCondition := core.CheckWinCondition(*ga.state); winCondition != nil {
				endEvent := ga.endGame(*winCondition)
				additionalEvents = append(additionalEvents, endEvent)
			}
		}
	}

	return additionalEvents
}

func (ga *GameActor) endGame(winCondition core.WinCondition) core.Event {
	// Generate comprehensive game analysis using StatsProcessor
	statsProcessor := game.NewStatsProcessor()

	// Collect all events from the datastore for analysis
	allEvents, err := ga.collectGameEvents()
	if err != nil {
		log.Printf("[GameActor/%s] Failed to collect events for analysis: %v", ga.gameID, err)
		allEvents = []core.Event{} // Fallback to empty events
	}

	// Generate analysis data
	gameAnalysis := statsProcessor.ProcessGameAnalysis(ga.state, allEvents)

	// Check for achievements and send notifications
	unlockedAchievements := ga.processAchievements(gameAnalysis)

	endEvent := core.Event{
		ID:        fmt.Sprintf("game_ended_%d", time.Now().UnixNano()),
		Type:      core.EventVictoryCondition,
		GameID:    ga.gameID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"winner":                winCondition.Winner,
			"condition":             winCondition.Condition,
			"description":           winCondition.Description,
			"analysis":              gameAnalysis,         // Include comprehensive analysis data
			"unlocked_achievements": unlockedAchievements, // Include newly unlocked achievements
		},
	}

	// Apply the event to state
	newState := core.ApplyEvent(*ga.state, endEvent)
	ga.state = &newState

	ga.phaseManager.CancelPhaseTransitions()
	return endEvent
}

// collectGameEvents retrieves all events for this game from the datastore for analysis
func (ga *GameActor) collectGameEvents() ([]core.Event, error) {
	// This would typically interface with the datastore to retrieve all events
	// For now, we'll return an empty slice since we don't have direct datastore access
	// In a full implementation, this would:
	// 1. Query the Redis datastore for all events in the game stream
	// 2. Parse and return the complete event history
	log.Printf("[GameActor/%s] Collecting events for analysis (placeholder implementation)", ga.gameID)
	return []core.Event{}, nil
}

// processAchievements checks for newly unlocked achievements and sends notifications
func (ga *GameActor) processAchievements(gameAnalysis *core.GameAnalysis) map[string][]string {
	if ga.achievementChecker == nil {
		return map[string][]string{}
	}

	// Check achievements for all players
	unlockedAchievements := ga.achievementChecker.CheckAchievements(ga.state, gameAnalysis)

	// Send achievement notifications to players
	for playerID, achievementIDs := range unlockedAchievements {
		for _, achievementID := range achievementIDs {
			achievement := ga.achievementChecker.GetAchievementByID(achievementID)
			if achievement != nil {
				// Create achievement unlocked notification
				notification := core.Event{
					ID:        fmt.Sprintf("achievement_unlocked_%s_%s_%d", playerID, achievementID, time.Now().UnixNano()),
					Type:      core.EventPrivateNotification,
					GameID:    ga.gameID,
					PlayerID:  playerID, // Private notification to this player
					Timestamp: time.Now(),
					Payload: map[string]interface{}{
						"type": "ACHIEVEMENT_UNLOCKED",
						"achievement": map[string]interface{}{
							"id":           achievement.ID,
							"name":         achievement.Name,
							"description":  achievement.Description,
							"rarity":       achievement.Rarity,
							"iconUrl":      achievement.IconURL,
							"avatarReward": achievement.AvatarReward,
							"titleReward":  achievement.TitleReward,
						},
					},
				}

				// Send the notification through the event callback
				if ga.eventCallback != nil {
					ga.eventCallback(ga.gameID, []core.Event{notification})
				}
			}
		}
	}

	return unlockedAchievements
}

// processAIActionsForPhase triggers AI actions when phases change
func (ga *GameActor) processAIActionsForPhase(phase core.PhaseType) {
	aiActions := ga.aiManager.ProcessAIActions()

	for _, action := range aiActions {
		// Process each AI action asynchronously to avoid blocking
		go func(aiAction core.Action) {
			log.Printf("[GameActor/%s] Processing AI action: %s for player %s", ga.gameID, aiAction.Type, aiAction.PlayerID)

			// Create a response channel for the AI action
			responseChan := make(chan interfaces.ProcessActionResult, 1)
			request := actorRequest{
				action:       aiAction,
				responseChan: responseChan,
			}

			// Send AI action to the mailbox
			select {
			case ga.mailbox <- request:
				// Wait for the action to be processed
				select {
				case result := <-responseChan:
					if result.Error != nil {
						log.Printf("[GameActor/%s] AI action failed: %v", ga.gameID, result.Error)
					} else {
						log.Printf("[GameActor/%s] AI action completed successfully", ga.gameID)
					}
				case <-time.After(5 * time.Second):
					log.Printf("[GameActor/%s] AI action timed out", ga.gameID)
				}
			case <-ga.ctx.Done():
				log.Printf("[GameActor/%s] Context canceled, dropping AI action", ga.gameID)
			}
		}(action)
	}
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
		ChatMessages: ga.state.ChatMessages, // Include chat history in state updates
		VoteState:    ga.state.VoteState,
		CrisisEvent:  ga.state.CrisisEvent,
		Settings:     ga.state.Settings,
	}

	return snapshotState
}

// getKPITarget returns the target value for a given KPI type
func (ga *GameActor) getKPITarget(kpiType core.KPIType) int {
	switch kpiType {
	case core.KPIInquisitor:
		return 3 // Vote correctly 3 times
	case core.KPIGuardian:
		return 4 // Keep CISO alive to Day 4
	case core.KPISuccessionPlanner:
		return 2 // End with exactly 2 humans
	case core.KPICapitalist:
		return 1 // End with most tokens (relative target)
	case core.KPIScapegoat:
		return 1 // Get eliminated unanimously
	default:
		return 1
	}
}

// getKPIReward returns the reward description for a given KPI type
func (ga *GameActor) getKPIReward(kpiType core.KPIType) string {
	switch kpiType {
	case core.KPIInquisitor:
		return "Gain 2 extra tokens for each correct vote"
	case core.KPIGuardian:
		return "Win if CISO survives to Day 4, regardless of faction victory"
	case core.KPISuccessionPlanner:
		return "Win if exactly 2 humans remain at game end"
	case core.KPICapitalist:
		return "Win if you have the most tokens at game end"
	case core.KPIScapegoat:
		return "Win if you are eliminated by unanimous vote"
	default:
		return "Complete objective for bonus rewards"
	}
}

// applySystemShockEffects checks for active system shocks and applies message corruption if needed
func (ga *GameActor) applySystemShockEffects(player *core.Player, message string) string {
	if player.SystemShocks == nil || len(player.SystemShocks) == 0 {
		return message
	}

	currentTime := time.Now()

	// Check each active system shock
	for i := range player.SystemShocks {
		shock := &player.SystemShocks[i]

		// Skip expired or inactive shocks
		if !shock.IsActive || currentTime.After(shock.ExpiresAt) {
			shock.IsActive = false
			continue
		}

		// Apply message corruption for MessageCorruption shock type
		if shock.Type == core.ShockMessageCorruption {
			// 25% chance to corrupt the message
			if ga.shouldCorruptMessage() {
				return "lol"
			}
		}
	}

	return message
}

// shouldCorruptMessage returns true 25% of the time for message corruption
func (ga *GameActor) shouldCorruptMessage() bool {
	// Use a simple random number generator seeded with current time
	// 25% chance means values 0, 1, 2 out of 0-15 (4/16 = 25%)
	randomValue := time.Now().UnixNano() % 16
	return randomValue < 4
}

// shouldTriggerCrisisEvent determines whether a crisis event should be triggered
func (ga *GameActor) shouldTriggerCrisisEvent() bool {
	// Trigger crisis events starting from Day 2, with a 75% chance each day
	if ga.state.DayNumber < 2 {
		return false
	}

	// 75% chance to trigger a crisis (values 0-11 out of 0-15)
	randomValue := time.Now().UnixNano() % 16
	return randomValue < 12
}

// handleHelpCommand processes /help commands and returns phase-specific rule summaries
func (ga *GameActor) handleHelpCommand(playerID string) ([]core.Event, error) {
	player := ga.state.Players[playerID]
	if player == nil {
		return nil, fmt.Errorf("player not found")
	}

	// Get help content based on current phase
	helpContent := ga.getPhaseHelp(ga.state.Phase.Type)

	// Create private notification event
	helpEvent := core.Event{
		ID:        fmt.Sprintf("help_response_%s_%d", playerID, time.Now().UnixNano()),
		Type:      core.EventPrivateNotification,
		GameID:    ga.gameID,
		PlayerID:  playerID, // Private notification to requesting player
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"type":    "HELP_RESPONSE",
			"phase":   string(ga.state.Phase.Type),
			"message": helpContent,
			"sender":  "Loebmate",
		},
	}

	return []core.Event{helpEvent}, nil
}

// validateActionPayloadSize validates that action payloads are within reasonable limits
func (ga *GameActor) validateActionPayloadSize(action core.Action) error {
	// Check overall payload size by serializing to JSON and measuring
	payloadBytes, err := json.Marshal(action.Payload)
	if err != nil {
		return fmt.Errorf("invalid action payload: %v", err)
	}

	// Limit total payload to 10KB to prevent memory exhaustion
	const maxPayloadSize = 10 * 1024 // 10KB
	if len(payloadBytes) > maxPayloadSize {
		return fmt.Errorf("action payload too large: %d bytes (max %d bytes)", len(payloadBytes), maxPayloadSize)
	}

	// Validate specific string fields based on action type
	switch action.Type {
	case core.ActionSendMessage:
		if messages, ok := action.Payload["messages"].([]interface{}); ok {
			if len(messages) > 10 {
				return fmt.Errorf("too many messages in batch: %d (max 10)", len(messages))
			}
			for i, msg := range messages {
				if msgStr, ok := msg.(string); ok {
					if len(msgStr) > 500 {
						return fmt.Errorf("message %d too long: %d characters (max 500)", i, len(msgStr))
					}
				}
			}
		}

	case core.ActionSubmitPulseCheck:
		if response, ok := action.Payload["response"].(string); ok {
			if len(response) > 200 {
				return fmt.Errorf("pulse check response too long: %d characters (max 200)", len(response))
			}
		}

	case core.ActionSetSlackStatus:
		if status, ok := action.Payload["status_message"].(string); ok {
			if len(status) > 100 {
				return fmt.Errorf("status message too long: %d characters (max 100)", len(status))
			}
		}

	case core.ActionSubmitExitInterview:
		if partingShot, ok := action.Payload["parting_shot"].(string); ok {
			if len(partingShot) > 50 {
				return fmt.Errorf("parting shot too long: %d characters (max 50)", len(partingShot))
			}
		}
	}

	return nil
}

// getPhaseHelp returns detailed help content for the current phase
func (ga *GameActor) getPhaseHelp(phase core.PhaseType) string {
	switch phase {
	case core.PhaseSitrep:
		return `**SITREP Phase Help**

**Objective:** Review the daily situation report
**Duration:** 30 seconds

**What to do:**
• Read the SITREP message carefully - it contains vital information about overnight events
• Look for personnel changes, security alerts, or other critical updates
• Use this information to inform your strategy for the day

**Available actions:** Chat in #war-room`

	case core.PhasePulseCheck:
		return `**Pulse Check Phase Help**

**Objective:** Share your thoughts on the current crisis
**Duration:** 45 seconds

**What to do:**
• Answer the pulse check question honestly and thoughtfully
• Your response will be shared with all players after this phase
• Use this to gauge team sentiment and share your perspective

**Available actions:** Submit pulse check response, chat in #war-room (after submitting)`

	case core.PhaseDiscussion:
		return `**Discussion Phase Help**

**Objective:** Collaborate and build consensus
**Duration:** 3 minutes

**What to do:**
• Share information and observations with your team
• Voice any suspicions about potential AI infiltrators
• Coordinate strategy and build alliances
• Analyze pulse check responses for insights

**Available actions:** Chat in #war-room, react to messages`

	case core.PhaseNomination:
		return `**Nomination Phase Help**

**Objective:** Vote to nominate someone for elimination
**Duration:** 1 minute

**What to do:**
• Choose carefully - your tokens add weight to your vote
• The person with the most token-weighted votes will face elimination
• Consider all available information before voting

**Available actions:** Cast nomination vote`

	case core.PhaseVerdict:
		return `**Verdict Phase Help**

**Objective:** Decide the nominated person's fate
**Duration:** 45 seconds

**What to do:**
• Vote GUILTY to eliminate the nominated person
• Vote INNOCENT to spare them
• Your tokens determine the weight of your vote

**Available actions:** Cast verdict vote (GUILTY/INNOCENT)`

	case core.PhaseNight:
		return `**Night Phase Help**

**Objective:** Use your role abilities and gather resources
**Duration:** 30 seconds

**What to do:**
• Use your role's special ability to help your team
• Mine tokens to increase your voting power
• Project milestones to advance team objectives
• The war room chat is locked during this phase

**Available actions:** Role abilities, mine tokens, project milestones`

	case core.PhaseLobby:
		return `**Lobby Phase Help**

**Objective:** Prepare for the game to begin

**What to do:**
• Review your role assignment and personal KPI in the right panel
• Understand your special abilities and objectives
• Wait for all players to be ready

**Available actions:** Chat in #war-room, ready up`

	case core.PhaseGameOver:
		return `**Game Over**

The game has concluded. Review the post-game analysis to see how everyone performed!`

	default:
		return `**Help**

Use /help during any phase to get specific guidance for that phase.

**General Tips:**
• Check your Personal Terminal (right panel) for your role and KPI
• Your tokens determine your voting power
• Trust carefully - the AI walks among us
• Work together to identify and eliminate the threat`
	}
}
