package actors

import (
	"fmt"
	"log"
	"time"

	"github.com/alignment/server/internal/game"
)

// GameActor represents a single game instance running in its own goroutine
type GameActor struct {
	gameID   string
	state    *game.GameState
	mailbox  chan game.Action
	events   chan game.Event
	shutdown chan struct{}
	
	// Dependencies (interfaces for testing)
	datastore   DataStore
	broadcaster Broadcaster
}

// DataStore interface for persistence
type DataStore interface {
	AppendEvent(gameID string, event game.Event) error
	SaveSnapshot(gameID string, state *game.GameState) error
	LoadEvents(gameID string, afterSequence int) ([]game.Event, error)
	LoadSnapshot(gameID string) (*game.GameState, error)
}

// Broadcaster interface for sending events to clients
type Broadcaster interface {
	BroadcastToGame(gameID string, event game.Event) error
	SendToPlayer(gameID, playerID string, event game.Event) error
}

// NewGameActor creates a new game actor
func NewGameActor(gameID string, datastore DataStore, broadcaster Broadcaster) *GameActor {
	return &GameActor{
		gameID:      gameID,
		state:       game.NewGameState(gameID),
		mailbox:     make(chan game.Action, 100), // Buffered channel
		events:      make(chan game.Event, 100),
		shutdown:    make(chan struct{}),
		datastore:   datastore,
		broadcaster: broadcaster,
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
func (ga *GameActor) SendAction(action game.Action) {
	select {
	case ga.mailbox <- action:
		// Action queued successfully
	default:
		log.Printf("GameActor %s: Mailbox full, dropping action %s", ga.gameID, action.Type)
	}
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
func (ga *GameActor) handleAction(action game.Action) {
	log.Printf("GameActor %s: Processing action %s from player %s", ga.gameID, action.Type, action.PlayerID)
	
	var events []game.Event
	
	switch action.Type {
	case game.ActionJoinGame:
		events = ga.handleJoinGame(action)
	case game.ActionLeaveGame:
		events = ga.handleLeaveGame(action)
	case game.ActionSubmitVote:
		events = ga.handleSubmitVote(action)
	case game.ActionMineTokens:
		events = ga.handleMineTokens(action)
	default:
		log.Printf("GameActor %s: Unknown action type: %s", ga.gameID, action.Type)
		return
	}
	
	// Apply events to state and send to event loop
	for _, event := range events {
		if err := ga.state.ApplyEvent(event); err != nil {
			log.Printf("GameActor %s: Failed to apply event: %v", ga.gameID, err)
			continue
		}
		
		// Send to event loop for persistence and broadcasting
		select {
		case ga.events <- event:
			// Event queued successfully
		default:
			log.Printf("GameActor %s: Event queue full, dropping event", ga.gameID)
		}
	}
}

func (ga *GameActor) handleJoinGame(action game.Action) []game.Event {
	playerName, _ := action.Payload["name"].(string)
	
	// Check if game is full
	if len(ga.state.Players) >= ga.state.Settings.MaxPlayers {
		return nil // Game full, ignore join request
	}
	
	// Check if player already joined
	if _, exists := ga.state.Players[action.PlayerID]; exists {
		return nil // Player already in game
	}
	
	event := game.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      game.EventPlayerJoined,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"name": playerName,
		},
	}
	
	return []game.Event{event}
}

func (ga *GameActor) handleLeaveGame(action game.Action) []game.Event {
	// Check if player is in game
	if _, exists := ga.state.Players[action.PlayerID]; !exists {
		return nil // Player not in game
	}
	
	event := game.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      game.EventPlayerLeft,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload:   make(map[string]interface{}),
	}
	
	return []game.Event{event}
}

func (ga *GameActor) handleSubmitVote(action game.Action) []game.Event {
	targetID, _ := action.Payload["target_id"].(string)
	
	// Validate vote (game phase, player active, etc.)
	if ga.state.Phase.Type != game.PhaseVoting {
		return nil // Not in voting phase
	}
	
	if _, exists := ga.state.Players[action.PlayerID]; !exists {
		return nil // Player not in game
	}
	
	event := game.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      game.EventPlayerVoted,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": targetID,
		},
	}
	
	return []game.Event{event}
}

func (ga *GameActor) handleMineTokens(action game.Action) []game.Event {
	// Simple mining logic - award 1 token
	event := game.Event{
		ID:        fmt.Sprintf("event_%d", time.Now().UnixNano()),
		Type:      game.EventMiningSuccessful,
		GameID:    ga.gameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"amount": 1,
		},
	}
	
	return []game.Event{event}
}