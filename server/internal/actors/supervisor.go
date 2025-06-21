package actors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xjhc/alignment/core"
)

// Supervisor manages all game actors and provides fault isolation
type Supervisor struct {
	actors map[string]*GameActor
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies
	datastore   DataStore
	broadcaster Broadcaster
}

// NewSupervisor creates a new supervisor
func NewSupervisor(ctx context.Context, datastore DataStore, broadcaster Broadcaster) *Supervisor {
	supervisorCtx, cancel := context.WithCancel(ctx)
	return &Supervisor{
		actors:      make(map[string]*GameActor),
		ctx:         supervisorCtx,
		cancel:      cancel,
		datastore:   datastore,
		broadcaster: broadcaster,
	}
}

// Start begins the supervisor's monitoring loop
func (s *Supervisor) Start() {
	go s.monitoringLoop()
}

// Stop gracefully shuts down all actors
func (s *Supervisor) Stop() {
	log.Println("Supervisor: Shutting down all actors")

	// Cancel context to signal all actors to stop
	s.cancel()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Wait for all actors to stop gracefully
	for gameID, actor := range s.actors {
		log.Printf("Supervisor: Stopping actor %s", gameID)
		actor.Stop()
	}

	// Clear actors map
	s.actors = make(map[string]*GameActor)
}

// CreateGame creates a new game actor
func (s *Supervisor) CreateGame(gameID string) error {
	return s.CreateGameWithPlayers(gameID, []string{})
}

// CreateGameWithPlayers creates a new game actor with a specific player list
func (s *Supervisor) CreateGameWithPlayers(gameID string, players []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if game already exists
	if _, exists := s.actors[gameID]; exists {
		log.Printf("Supervisor: Game %s already exists", gameID)
		return ErrGameAlreadyExists
	}

	// Create new actor with child context
	actorCtx, actorCancel := context.WithCancel(s.ctx)
	actor := NewGameActor(actorCtx, actorCancel, gameID, s.datastore, s.broadcaster)
	s.actors[gameID] = actor

	// This is the critical part from the documentation - launch supervised goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// This is where you handle the crash of a single game
				log.Printf("Supervisor: GameActor %s panicked: %v", gameID, r)
				s.RemoveGame(gameID) // Or implement a restart policy
			}
		}()
		// Each actor runs its own main loop
		actor.Start()
	}()

	// If players are provided, send join actions for each
	for _, playerID := range players {
		action := core.Action{
			Type:      core.ActionJoinGame,
			PlayerID:  playerID,
			GameID:    gameID,
			Timestamp: time.Now(),
		}
		actor.SendAction(action)
	}

	log.Printf("Supervisor: Created and started game actor %s with %d players", gameID, len(players))
	return nil
}

// GetActor returns a game actor by ID
func (s *Supervisor) GetActor(gameID string) (*GameActor, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	actor, exists := s.actors[gameID]
	return actor, exists
}

// RemoveGame removes a game actor
func (s *Supervisor) RemoveGame(gameID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if actor, exists := s.actors[gameID]; exists {
		log.Printf("Supervisor: Removing game actor %s", gameID)
		actor.Stop()
		delete(s.actors, gameID)
	}
}

// monitoringLoop periodically checks actor health
func (s *Supervisor) monitoringLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkActorHealth()
		case <-s.ctx.Done():
			log.Println("Supervisor: Monitoring loop stopped")
			return
		}
	}
}

// checkActorHealth monitors actor health and restarts failed ones
func (s *Supervisor) checkActorHealth() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for gameID, actor := range s.actors {
		// Check if actor context is done (actor has stopped)
		select {
		case <-actor.ctx.Done():
			// Actor has stopped, attempt restart
			log.Printf("Supervisor: Detected failed actor %s, attempting restart", gameID)
			s.restartActor(gameID)
		default:
			// Actor is still running
		}
	}
}

// restartActor attempts to restart a failed actor
func (s *Supervisor) restartActor(gameID string) {
	// Remove the failed actor
	delete(s.actors, gameID)

	// Create new actor with fresh context
	actorCtx, actorCancel := context.WithCancel(s.ctx)
	actor := NewGameActor(actorCtx, actorCancel, gameID, s.datastore, s.broadcaster)

	// TODO: Restore state from persistence layer
	// This would involve loading the latest snapshot and replaying events

	s.actors[gameID] = actor

	// Launch supervised goroutine for the restarted actor
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Supervisor: Restarted GameActor %s panicked: %v", gameID, r)
				s.RemoveGame(gameID)
			}
		}()
		actor.Start()
	}()

	log.Printf("Supervisor: Restarted actor %s", gameID)
}

// GetStats returns supervisor statistics
func (s *Supervisor) GetStats() SupervisorStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return SupervisorStats{
		ActiveGames: len(s.actors),
		Uptime:      time.Since(time.Now()), // Simplified
	}
}

// SupervisorStats contains supervisor statistics
type SupervisorStats struct {
	ActiveGames int           `json:"active_games"`
	Uptime      time.Duration `json:"uptime"`
}

// Custom errors
var (
	ErrGameAlreadyExists = fmt.Errorf("game already exists")
	ErrGameNotFound      = fmt.Errorf("game not found")
)
