package actors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/events"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/store"
)

// Supervisor manages all game actors and provides fault isolation
type Supervisor struct {
	actors map[string]*GameActor
	mutex  sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc

	// Dependencies
	datastore     interfaces.DataStore
	postgresStore *store.PostgresStore
	broadcaster   interfaces.Broadcaster
	eventBus      *events.EventBus
}

// NewSupervisor creates a new supervisor
func NewSupervisor(ctx context.Context, datastore interfaces.DataStore, postgresStore *store.PostgresStore, broadcaster interfaces.Broadcaster, eventBus *events.EventBus) *Supervisor {
	supervisorCtx, cancel := context.WithCancel(ctx)
	return &Supervisor{
		actors:        make(map[string]*GameActor),
		ctx:           supervisorCtx,
		cancel:        cancel,
		datastore:     datastore,
		postgresStore: postgresStore,
		broadcaster:   broadcaster,
		eventBus:      eventBus,
	}
}

// SetBroadcaster sets the broadcaster dependency (for dependency injection)
func (s *Supervisor) SetBroadcaster(broadcaster interfaces.Broadcaster) {
	s.broadcaster = broadcaster
}

// Start begins the supervisor's monitoring loop
func (s *Supervisor) Start() {
	go s.monitoringLoop()
}

// Stop gracefully shuts down all actors
func (s *Supervisor) Stop() {
	log.Println("Supervisor: Shutting down all actors")
	s.cancel()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for gameID, actor := range s.actors {
		log.Printf("Supervisor: Stopping actor %s", gameID)
		actor.Stop()
	}
	s.actors = make(map[string]*GameActor)
}

// CreateGame creates a new game actor, but is deprecated
func (s *Supervisor) CreateGame(gameID string) error {
	return fmt.Errorf("CreateGame is deprecated, use CreateGameWithPlayers")
}

func (s *Supervisor) CreateGameWithPlayers(gameID string, players map[string]*core.Player) (interfaces.GameActorInterface, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.actors[gameID]; exists {
		log.Printf("[Supervisor] Game %s already exists", gameID)
		// Return the existing actor instead of an error
		return s.actors[gameID], nil
	}

	actorCtx, actorCancel := context.WithCancel(s.ctx)
	actor := NewGameActor(actorCtx, actorCancel, gameID, players, s.postgresStore)
	
	return s.setupGameActor(gameID, actor)
}

func (s *Supervisor) CreateGameWithPlayersAndSettings(gameID string, players map[string]*core.Player, settings core.GameSettings) (interfaces.GameActorInterface, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.actors[gameID]; exists {
		log.Printf("[Supervisor] Game %s already exists", gameID)
		// Return the existing actor instead of an error
		return s.actors[gameID], nil
	}

	actorCtx, actorCancel := context.WithCancel(s.ctx)
	actor := NewGameActorWithSettings(actorCtx, actorCancel, gameID, players, settings, s.postgresStore)
	
	return s.setupGameActor(gameID, actor)
}

// setupGameActor configures a newly created game actor
func (s *Supervisor) setupGameActor(gameID string, actor *GameActor) (interfaces.GameActorInterface, error) {
	// Inject EventBus dependency for system event publishing
	if s.eventBus != nil {
		actor.SetEventBus(s.eventBus)
	}
	
	// Set up event callback for timer-generated events
	actor.SetEventCallback(func(gameID string, events []core.Event) {
		// Broadcast events to players using the supervisor's broadcaster
		for _, event := range events {
			if event.PlayerID != "" { // Private event
				s.broadcaster.SendToPlayer(gameID, event.PlayerID, event)
			} else { // Public event
				s.broadcaster.BroadcastToGame(gameID, event)
			}
		}
	})
	
	s.actors[gameID] = actor

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Supervisor] GameActor %s panicked: %v", gameID, r)
				s.RemoveGame(gameID)
			}
		}()
		actor.Start()
	}()

	log.Printf("[Supervisor] Created game actor %s", gameID)
	return actor, nil
}

// GetActor returns a game actor by ID
func (s *Supervisor) GetActor(gameID string) (interfaces.GameActorInterface, bool) {
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
		log.Printf("[Supervisor] Removing game actor %s", gameID)
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
			log.Println("[Supervisor] Monitoring loop stopped")
			return
		}
	}
}

// checkActorHealth monitors actor health and restarts failed ones
func (s *Supervisor) checkActorHealth() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for gameID, actor := range s.actors {
		select {
		case <-actor.ctx.Done():
			log.Printf("[Supervisor] Detected failed actor %s, attempting restart", gameID)
			s.restartActor(gameID)
		default:
		}
	}
}

// restartActor attempts to restart a failed actor
func (s *Supervisor) restartActor(gameID string) {
	delete(s.actors, gameID)
	actorCtx, actorCancel := context.WithCancel(s.ctx)
	// For restarted actors, we'll start with empty players map - they'll rejoin if still connected
	actor := NewGameActor(actorCtx, actorCancel, gameID, make(map[string]*core.Player), s.postgresStore)
	s.actors[gameID] = actor

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[Supervisor] Restarted GameActor %s panicked: %v", gameID, r)
				s.RemoveGame(gameID)
			}
		}()
		actor.Start()
	}()
	log.Printf("[Supervisor] Restarted actor %s", gameID)
}

// GetStats returns supervisor statistics
func (s *Supervisor) GetStats() SupervisorStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	gameInfos := make([]GameInfo, 0, len(s.actors))
	for gameID, actor := range s.actors {
		state := actor.GetGameState()
		gameInfo := GameInfo{
			GameID:      gameID,
			PlayerCount: len(state.Players),
			Status:      string(state.Phase.Type),
		}
		gameInfos = append(gameInfos, gameInfo)
	}
	
	return SupervisorStats{
		ActiveGames: len(s.actors),
		Games:       gameInfos,
	}
}

// SupervisorStats contains supervisor statistics
type SupervisorStats struct {
	ActiveGames int        `json:"active_games"`
	Games       []GameInfo `json:"games,omitempty"`
}

// GameInfo represents basic information about a game
type GameInfo struct {
	GameID      string `json:"game_id"`
	PlayerCount int    `json:"player_count"`
	Status      string `json:"status"`
}

// GetGameState returns the state of a specific game
func (s *Supervisor) GetGameState(gameID string) (*core.GameState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	actor, exists := s.actors[gameID]
	if !exists {
		return nil, ErrGameNotFound
	}
	
	return actor.GetGameState(), nil
}

// TerminateGame forcefully terminates a game and removes its actor
func (s *Supervisor) TerminateGame(gameID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	actor, exists := s.actors[gameID]
	if !exists {
		return ErrGameNotFound
	}
	
	log.Printf("[Supervisor] Forcefully terminating game %s", gameID)
	actor.Stop()
	delete(s.actors, gameID)
	
	return nil
}

// Custom errors
var (
	ErrGameAlreadyExists = fmt.Errorf("game already exists")
	ErrGameNotFound      = fmt.Errorf("game not found")
)
