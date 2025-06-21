package actors

import (
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

// GameActor represents a single game instance running in its own goroutine
type GameActor struct {
	gameID   string
	state    *core.GameState
	mailbox  chan core.Action
	events   chan core.Event
	shutdown chan struct{}

	// Dependencies (interfaces for testing)
	datastore   DataStore
	broadcaster Broadcaster

	// Game managers (domain experts)
	votingManager      VotingManager
	miningManager      MiningManager
	roleAbilityManager RoleAbilityManager
	eliminationManager *game.EliminationManager
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
func NewGameActor(gameID string, datastore DataStore, broadcaster Broadcaster) *GameActor {
	state := core.NewGameState(gameID)
	return &GameActor{
		gameID:      gameID,
		state:       state,
		mailbox:     make(chan core.Action, 100), // Buffered channel
		events:      make(chan core.Event, 100),
		shutdown:    make(chan struct{}),
		datastore:   datastore,
		broadcaster: broadcaster,

		// Initialize managers with shared state
		votingManager:      game.NewVotingManager(state),
		miningManager:      game.NewMiningManager(state),
		roleAbilityManager: game.NewRoleAbilityManager(state),
		eliminationManager: game.NewEliminationManager(state),
	}
}

// Start begins the actor's main processing loop
func (ga *GameActor) Start() {
	log.Printf("GameActor %s: Starting", ga.gameID)

	// Start the main processing loop in a goroutine
	go ga.processLoop()

	// Start the event persistence loop
	go ga.eventLoop()
}

// Stop gracefully shuts down the actor
func (ga *GameActor) Stop() {
	log.Printf("GameActor %s: Stopping", ga.gameID)
	close(ga.shutdown)
}

// SendAction sends an action to the actor's mailbox
func (ga *GameActor) SendAction(action core.Action) {
	select {
	case ga.mailbox <- action:
		// Action queued successfully
	default:
		log.Printf("GameActor %s: Mailbox full, dropping action %s", ga.gameID, action.Type)
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
		case action := <-ga.mailbox:
			ga.handleAction(action)
		case <-ga.shutdown:
			log.Printf("GameActor %s: Shutting down", ga.gameID)
			return
		}
	}
}

// eventLoop handles event persistence and broadcasting
func (ga *GameActor) eventLoop() {
	for {
		select {
		case event := <-ga.events:
			// Persist the event
			if err := ga.datastore.AppendEvent(ga.gameID, event); err != nil {
				log.Printf("GameActor %s: Failed to persist event: %v", ga.gameID, err)
			}

			// Broadcast to clients
			if err := ga.broadcaster.BroadcastToGame(ga.gameID, event); err != nil {
				log.Printf("GameActor %s: Failed to broadcast event: %v", ga.gameID, err)
			}

		case <-ga.shutdown:
			return
		}
	}
}

// handleAction processes a single action and generates events
func (ga *GameActor) handleAction(action core.Action) {
	log.Printf("GameActor %s: Processing action %s from player %s", ga.gameID, action.Type, action.PlayerID)

	var events []core.Event

	switch action.Type {
	case core.ActionJoinGame:
		events = ga.handleJoinGame(action)
	case core.ActionLeaveGame:
		events = ga.handleLeaveGame(action)
	case core.ActionSubmitVote:
		events = ga.handleSubmitVote(action)
	case core.ActionSubmitNightAction:
		events = ga.handleSubmitNightAction(action)
	case core.ActionMineTokens:
		events = ga.handleMineTokens(action)
	case core.ActionType("PHASE_TRANSITION"):
		events = ga.handlePhaseTransition(action)
	default:
		log.Printf("GameActor %s: Unknown action type: %s", ga.gameID, action.Type)
		return
	}

	// Apply events to state and send to event loop
	ga.applyAndBroadcast(events)
}

// applyAndBroadcast applies events to state and queues them for persistence/broadcast
func (ga *GameActor) applyAndBroadcast(events []core.Event) {
	for _, event := range events {
		newState := core.ApplyEvent(*ga.state, event)
		ga.state = &newState

		// Send to event loop for persistence and broadcasting
		select {
		case ga.events <- event:
			// Event queued successfully
		default:
			log.Printf("GameActor %s: Event queue full, dropping event", ga.gameID)
		}
	}
}

func (ga *GameActor) handleJoinGame(action core.Action) []core.Event {
	playerName, _ := action.Payload["name"].(string)
	jobTitle, _ := action.Payload["job_title"].(string)

	// Check if game is full
	if len(ga.state.Players) >= ga.state.Settings.MaxPlayers {
		return nil // Game full, ignore join request
	}

	// Check if player already joined
	if _, exists := ga.state.Players[action.PlayerID]; exists {
		return nil // Player already in game
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

	return []core.Event{event}
}

func (ga *GameActor) handleLeaveGame(action core.Action) []core.Event {
	// Check if player is in game
	if _, exists := ga.state.Players[action.PlayerID]; !exists {
		return nil // Player not in game
	}

	event := core.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      core.EventPlayerLeft,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}

	return []core.Event{event}
}

func (ga *GameActor) handleSubmitVote(action core.Action) []core.Event {
	// Delegate to VotingManager for complex business logic
	events, err := ga.votingManager.HandleVoteAction(action)
	if err != nil {
		log.Printf("GameActor %s: Invalid vote action from player %s: %v", ga.gameID, action.PlayerID, err)
		// Could send a private error event back to the player here
		return nil
	}
	
	return events
}

func (ga *GameActor) handleSubmitNightAction(action core.Action) []core.Event {
	// Delegate to RoleAbilityManager for complex business logic
	events, err := ga.roleAbilityManager.HandleNightAction(action)
	if err != nil {
		log.Printf("GameActor %s: Invalid night action from player %s: %v", ga.gameID, action.PlayerID, err)
		// Could send a private error event back to the player here
		return nil
	}
	
	return events
}

func (ga *GameActor) handleMineTokens(action core.Action) []core.Event {
	// Delegate to MiningManager for complex business logic
	events, err := ga.miningManager.HandleMineAction(action)
	if err != nil {
		log.Printf("GameActor %s: Mining action error from player %s: %v", ga.gameID, action.PlayerID, err)
		return nil
	}
	
	return events
}

func (ga *GameActor) handlePhaseTransition(action core.Action) []core.Event {
	nextPhase, _ := action.Payload["next_phase"].(string)

	var events []core.Event

	// If we're transitioning FROM night phase, resolve night actions first
	if ga.state.Phase.Type == core.PhaseNight {
		// Simplified night resolution - in full implementation would use game package
		// Clear temporary night resolution state
		ga.state.BlockedPlayersTonight = nil
		ga.state.ProtectedPlayersTonight = nil
	}

	// Create phase transition event
	phaseEvent := core.Event{
		ID:        fmt.Sprintf("phase_transition_%s_%d", nextPhase, time.Now().UnixNano()),
		Type:      core.EventPhaseChanged,
		GameID:    ga.gameID,
		PlayerID:  "",
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"previous_phase": string(ga.state.Phase.Type),
			"next_phase":     nextPhase,
			"day_number":     ga.state.DayNumber,
		},
	}
	events = append(events, phaseEvent)

	return events
}
