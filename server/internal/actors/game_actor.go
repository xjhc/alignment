package actors

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/ai"
	"github.com/xjhc/alignment/server/internal/game"
	"github.com/xjhc/alignment/server/internal/interfaces"
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
	rng                     *rand.Rand
	
	// Callback for event notifications (especially for timer-generated events)
	eventCallback EventCallback
}

// NewGameActor creates a new game actor with empty state - call Initialize() after creation
func NewGameActor(ctx context.Context, cancel context.CancelFunc, gameID string, players map[string]*core.Player) *GameActor {
	state := core.NewGameState(gameID, time.Now())
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

// SetEventCallback sets the callback function for event notifications
func (ga *GameActor) SetEventCallback(callback EventCallback) {
	ga.eventCallback = callback
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
	switch action.Type {
	case core.ActionType("INITIALIZE_GAME"):
		return ga.generateInitializeGameEvents(action)
	case core.ActionLeaveGame:
		return ga.validateAndGenerateLeaveGame(action)
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

	// Generate pulse check submitted event
	event := core.Event{
		ID:        fmt.Sprintf("pulse_check_submitted_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPulseCheckSubmitted,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_id":   action.PlayerID,
			"player_name": player.Name,
			"response":    response,
		},
	}

	return []core.Event{event}, nil
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
	switch crisis.Type {
	case "Database Index Corruption":
		return "A critical role has been exposed. How does this change your immediate priority?"
	case "Cascading Server Failure":
		return "With limited bandwidth, what is the one piece of information everyone needs to hear from you?"
	case "Emergency Board Meeting":
		return "The Board demands accountability. Which two roles do you believe are most responsible for this situation?"
	case "Tainted Training Data":
		return "We've learned the AI was trained on compromised data. What 'unshakeable truth' do you now question?"
	case "Nightmare Scenario":
		return "Emergency protocols are in effect. What is your immediate action to protect the company?"
	case "Press Leak":
		return "Sensitive information has leaked. What is your strategy to control the narrative?"
	case "Incident Response Drill":
		return "All communications are monitored. What would you say if you knew everyone was listening?"
	case "Major Service Outage":
		return "Critical services are down. What is your highest priority for recovery efforts?"
	case "Phishing Attack":
		return "Security has been compromised. Who do you trust most in this room and why?"
	case "Data Privacy Audit":
		return "External auditors are reviewing everything. What would concern you most if discovered?"
	case "Vendor Security Breach":
		return "A trusted partner has been compromised. How do you verify who you can still trust?"
	case "Regulatory Review":
		return "Government oversight is imminent. What would you want leadership to know before they arrive?"
	default:
		return ""
	}
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

// generateInitializeGameEvents handles the game setup by generating events.
func (ga *GameActor) generateInitializeGameEvents(action core.Action) ([]core.Event, error) {
	log.Printf("[GameActor/%s] Generating events for role and alignment assignment...", ga.gameID)
	var events []core.Event

	assignments := assignRolesAndAlignments(getPlayerIDs(ga.state.Players), ga.rng)
	for playerID, assignment := range assignments {
		// Create a ROLE_ASSIGNED event for each player.
		// These events will be applied internally to build the correct server state
		// before generating the player-specific snapshots.
		// Create role assignment event  
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
				"alignment":        assignment.Alignment,
			},
		}
		events = append(events, roleAssignedEvent)

		// Create separate KPI assignment event for human players only
		if assignment.Alignment == "human" && assignment.KPIType != "" {
			kpiAssignedEvent := core.Event{
				ID:        fmt.Sprintf("kpi_assigned_%s", playerID),
				Type:      core.EventKPIAssigned,
				GameID:    ga.gameID,
				PlayerID:  playerID, // Private event for this player
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"kpi_type":    string(assignment.KPIType),
					"description": assignment.KPIDescription,
					"target":      ga.getKPITarget(assignment.KPIType),
					"reward":      ga.getKPIReward(assignment.KPIType),
				},
			}
			events = append(events, kpiAssignedEvent)
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
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"sender_id":    "SYSTEM",
			"sender_name":  "Security Alert",
			"message":      "[SEV-1] Critical Security Incident - Immediate Response Protocol",
			"phase":        "LOBBY",
			"day_number":   1,
			"channel_id":   "#war-room",
			"is_system":    true,
			"type":         "INCITING_INCIDENT",
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
		Type:      core.EventChatMessage,
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

	// Extract channel ID (default to #war-room for backward compatibility)
	channelID, ok := action.Payload["channel_id"].(string)
	if !ok || channelID == "" {
		channelID = "#war-room"
	}

	// Validate channel-specific permissions using the new rules
	if !core.CanPlayerSendMessageInChannel(*player, channelID, ga.state.Phase.Type, time.Now()) {
		switch channelID {
		case "#war-room":
			if ga.state.Phase.Type == core.PhasePulseCheck && !player.HasSubmittedPulseCheck {
				return nil, fmt.Errorf("must submit pulse check response before chatting")
			} else if ga.state.Phase.Type == core.PhaseNight {
				return nil, fmt.Errorf("war room chat is locked during night phase")
			} else {
				return nil, fmt.Errorf("cannot send messages in war room during %s phase", ga.state.Phase.Type)
			}
		case "#aligned":
			if player.Alignment != "ALIGNED" {
				return nil, fmt.Errorf("only AI faction members can access aligned channel")
			}
		default:
			return nil, fmt.Errorf("invalid channel: %s", channelID)
		}
	}

	// Check if this is a private message (legacy support)
	targetID, isPrivate := action.Payload["target_id"].(string)

	// Check mandate restrictions for private messages
	if isPrivate && targetID != "" {
		if ga.corporateMandateManager.IsMandateActive() {
			_, noDirectMessages := ga.corporateMandateManager.CheckCommunicationRestrictions()
			if noDirectMessages {
				return nil, fmt.Errorf("private messaging suspended due to Total Transparency Initiative")
			}
		}
	}

	// Basic message validation
	if len(message) > 500 {
		return nil, fmt.Errorf("message too long (max 500 characters)")
	}

	// Check for System Shock effects that might corrupt the message
	message = ga.applySystemShockEffects(player, message)

	// Create chat message event
	eventPlayerID := "" // Public event by default
	payload := map[string]interface{}{
		"sender_id":   action.PlayerID,
		"sender_name": player.Name,
		"message":     message,
		"phase":       string(ga.state.Phase.Type),
		"day_number":  ga.state.DayNumber,
		"channel_id":  channelID,
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
		ID:        fmt.Sprintf("chat_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventChatMessage,
		GameID:    ga.gameID,
		PlayerID:  eventPlayerID,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	return []core.Event{event}, nil
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
		"player_id":    action.PlayerID,
		"player_name":  player.Name,
		"message_id":   messageID,
		"emoji":        emoji,
		"channel_id":   channelID,
		"phase":        string(ga.state.Phase.Type),
		"day_number":   ga.state.DayNumber,
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

	case core.PhaseSitrep:
		// Resolve all night actions
		if ga.state.Phase.Type == core.PhaseNight {
			nightManager := game.NewNightResolutionManager(ga.state)
			nightEvents := nightManager.ResolveNightActions()
			events = append(events, nightEvents...)
		}
		// Increment day number
		ga.state.DayNumber++
		
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
				"question": question,
				"day_number": ga.state.DayNumber,
			},
		}
		events = append(events, pulseCheckEvent)
		
	case core.PhaseDiscussion:
		// Reveal pulse check results if transitioning from PULSE_CHECK phase
		if ga.state.Phase.Type == core.PhasePulseCheck {
			pulseCheckRevealed := ga.generatePulseCheckRevelation()
			if pulseCheckRevealed.Type != "" {
				events = append(events, pulseCheckRevealed)
			}
		}
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
				"vote_type": string(ga.state.VoteState.Type),
				"results":   ga.state.VoteState.Results,
				"vote_breakdown": ga.createVoteBreakdown(),
				"winner_info": ga.createWinnerInfo(),
				"total_votes": len(ga.state.VoteState.Votes),
				"total_token_weight": ga.calculateTotalTokenWeight(),
			},
		}
		events = append(events, voteCompleteEvent)

		// Create chat message for vote results display
		voteResultChatEvent := ga.createVoteResultChatMessage()
		events = append(events, voteResultChatEvent)
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
		"has_winner": !hasTie && winner != "",
		"has_tie":    hasTie,
		"winner_id":  winner,
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
		"question":        question,
		"outcome":         outcome,
		"votes":           ga.state.VoteState.Votes,
		"tokenWeights":    ga.state.VoteState.TokenWeights,
		"results":         ga.state.VoteState.Results,
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