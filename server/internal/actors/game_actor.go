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

// GameActor represents a single game instance running in its own goroutine
type GameActor struct {
	gameID  string
	state   *core.GameState
	mailbox chan core.Action

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
	return &GameActor{
		gameID:      gameID,
		state:       state,
		mailbox:     make(chan core.Action, 100), // Buffered channel
		ctx:         ctx,
		cancel:      cancel,
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
}

// Stop gracefully shuts down the actor
func (ga *GameActor) Stop() {
	log.Printf("GameActor %s: Stopping", ga.gameID)
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
			// Check if context is canceled before processing action
			if ga.ctx.Err() != nil {
				log.Printf("GameActor %s: Context canceled, stopping processing", ga.gameID)
				return
			}
			ga.handleAction(action)
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
	}
}

// generateEventsForAction validates an action and generates events without modifying state
func (ga *GameActor) generateEventsForAction(action core.Action) ([]core.Event, error) {
	switch action.Type {
	case core.ActionJoinGame:
		return ga.validateAndGenerateJoinGame(action)
	case core.ActionLeaveGame:
		return ga.validateAndGenerateLeaveGame(action)
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
			"previous_phase": string(ga.state.Phase.Type),
			"next_phase":     nextPhase,
			"day_number":     ga.state.DayNumber,
		},
	}
	events = append(events, phaseEvent)

	return events, nil
}
