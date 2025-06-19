package actors

import (
	"log"
	"sync"
	"time"
)

// Supervisor manages all game actors and provides fault isolation
type Supervisor struct {
	actors    map[string]*GameActor
	mutex     sync.RWMutex
	shutdown  chan struct{}
	
	// Dependencies
	datastore   DataStore
	broadcaster Broadcaster
}

// NewSupervisor creates a new supervisor
func NewSupervisor(datastore DataStore, broadcaster Broadcaster) *Supervisor {
	return &Supervisor{
		actors:      make(map[string]*GameActor),
		shutdown:    make(chan struct{}),
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
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	log.Println("Supervisor: Shutting down all actors")
	
	// Stop all actors
	for gameID, actor := range s.actors {
		log.Printf("Supervisor: Stopping actor %s", gameID)
		actor.Stop()
	}
	
	// Clear actors map
	s.actors = make(map[string]*GameActor)
	
	// Signal shutdown
	close(s.shutdown)
}

// CreateGame creates a new game actor
func (s *Supervisor) CreateGame(gameID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	// Check if game already exists
	if _, exists := s.actors[gameID]; exists {
		log.Printf("Supervisor: Game %s already exists", gameID)
		return ErrGameAlreadyExists
	}
	
	// Create new actor
	actor := NewGameActor(gameID, s.datastore, s.broadcaster)
	s.actors[gameID] = actor
	
	// Start the actor
	actor.Start()
	
	log.Printf("Supervisor: Created and started game actor %s", gameID)
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
		case <-s.shutdown:
			return
		}
	}
}

// checkActorHealth monitors actor health and restarts failed ones
func (s *Supervisor) checkActorHealth() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	for gameID, actor := range s.actors {
		// Check if actor is still running (simplified check)
		select {
		case <-actor.shutdown:
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
	
	// Create new actor
	actor := NewGameActor(gameID, s.datastore, s.broadcaster)
	
	// TODO: Restore state from persistence layer
	// This would involve loading the latest snapshot and replaying events
	
	s.actors[gameID] = actor
	actor.Start()
	
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