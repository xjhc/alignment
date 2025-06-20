package game

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/xjhc/alignment/core"
)

// Timer represents a scheduled event
type Timer struct {
	ID        string
	GameID    string
	Type      TimerType
	ExpiresAt time.Time
	Action    TimerAction
}

// TimerType represents different types of timers
type TimerType string

const (
	TimerPhaseEnd  TimerType = "PHASE_END"
	TimerGameStart TimerType = "GAME_START"
	TimerHeartbeat TimerType = "HEARTBEAT"
)

// TimerAction represents an action to execute when timer expires
type TimerAction struct {
	Type    core.ActionType             `json:"type"`
	Payload map[string]interface{} `json:"payload"`
}

// TimerCallback is called when a timer expires
type TimerCallback func(timer Timer)

// Scheduler manages game timers using a timing wheel algorithm
type Scheduler struct {
	timers   map[string]*Timer
	mutex    sync.RWMutex
	callback TimerCallback
	ctx      context.Context
	cancel   context.CancelFunc
	ticker   *time.Ticker
	running  bool
}

// NewScheduler creates a new scheduler
func NewScheduler(callback TimerCallback) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		timers:   make(map[string]*Timer),
		callback: callback,
		ctx:      ctx,
		cancel:   cancel,
		ticker:   time.NewTicker(1 * time.Second), // Check every second
		running:  false,
	}
}

// Start begins the scheduler's timing wheel
func (s *Scheduler) Start() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return
	}

	s.running = true
	go s.run()
	log.Println("Scheduler: Started")
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.cancel()
	s.ticker.Stop()
	log.Println("Scheduler: Stopped")
}

// ScheduleTimer adds a new timer
func (s *Scheduler) ScheduleTimer(timer Timer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.timers[timer.ID] = &timer
	log.Printf("Scheduler: Scheduled timer %s for %v", timer.ID, timer.ExpiresAt)
}

// CancelTimer removes a timer
func (s *Scheduler) CancelTimer(timerID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.timers[timerID]; exists {
		delete(s.timers, timerID)
		log.Printf("Scheduler: Cancelled timer %s", timerID)
	}
}

// CancelGameTimers removes all timers for a specific game
func (s *Scheduler) CancelGameTimers(gameID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for timerID, timer := range s.timers {
		if timer.GameID == gameID {
			delete(s.timers, timerID)
		}
	}
	log.Printf("Scheduler: Cancelled all timers for game %s", gameID)
}

// run is the main scheduler loop
func (s *Scheduler) run() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-s.ticker.C:
			s.processExpiredTimers(now)
		}
	}
}

// processExpiredTimers checks for and executes expired timers
func (s *Scheduler) processExpiredTimers(now time.Time) {
	s.mutex.Lock()
	expiredTimers := make([]*Timer, 0)

	// Find expired timers
	for timerID, timer := range s.timers {
		if now.After(timer.ExpiresAt) || now.Equal(timer.ExpiresAt) {
			expiredTimers = append(expiredTimers, timer)
			delete(s.timers, timerID)
		}
	}
	s.mutex.Unlock()

	// Execute expired timers (outside the lock to avoid deadlock)
	for _, timer := range expiredTimers {
		log.Printf("Scheduler: Executing expired timer %s", timer.ID)
		if s.callback != nil {
			go s.callback(*timer) // Execute in goroutine to avoid blocking
		}
	}
}

// GetActiveTimers returns all active timers for debugging
func (s *Scheduler) GetActiveTimers() map[string]*Timer {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make(map[string]*Timer)
	for id, timer := range s.timers {
		result[id] = timer
	}
	return result
}

// PhaseManager handles automatic phase transitions
type PhaseManager struct {
	scheduler *Scheduler
	gameID    string
	settings  core.GameSettings
}

// NewPhaseManager creates a new phase manager
func NewPhaseManager(scheduler *Scheduler, gameID string, settings core.GameSettings) *PhaseManager {
	return &PhaseManager{
		scheduler: scheduler,
		gameID:    gameID,
		settings:  settings,
	}
}

// SchedulePhaseTransition schedules the next phase transition
func (pm *PhaseManager) SchedulePhaseTransition(currentPhase core.PhaseType, phaseStartTime time.Time) {
	duration := getPhaseDuration(currentPhase, pm.settings)
	nextPhase := getNextPhase(currentPhase)

	if duration == 0 || nextPhase == core.PhaseGameOver {
		return // Unknown phase or end of game, don't schedule
	}

	timerID := pm.gameID + "_phase_" + string(currentPhase)
	expiresAt := phaseStartTime.Add(duration)

	timer := Timer{
		ID:        timerID,
		GameID:    pm.gameID,
		Type:      TimerPhaseEnd,
		ExpiresAt: expiresAt,
		Action: TimerAction{
			Type: core.ActionType("PHASE_TRANSITION"),
			Payload: map[string]interface{}{
				"next_phase": string(nextPhase),
				"duration":   getPhaseDuration(nextPhase, pm.settings).Seconds(),
			},
		},
	}

	pm.scheduler.ScheduleTimer(timer)
}

// CancelPhaseTransitions cancels all phase timers for this game
func (pm *PhaseManager) CancelPhaseTransitions() {
	pm.scheduler.CancelGameTimers(pm.gameID)
}

// getPhaseDuration returns the duration for a specific phase
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

// getNextPhase returns the next phase after the current one
func getNextPhase(currentPhase core.PhaseType) core.PhaseType {
	switch currentPhase {
	case core.PhaseSitrep:
		return core.PhasePulseCheck
	case core.PhasePulseCheck:
		return core.PhaseDiscussion
	case core.PhaseDiscussion:
		return core.PhaseExtension
	case core.PhaseExtension:
		return core.PhaseNomination
	case core.PhaseNomination:
		return core.PhaseTrial
	case core.PhaseTrial:
		return core.PhaseVerdict
	case core.PhaseVerdict:
		return core.PhaseNight
	case core.PhaseNight:
		return core.PhaseSitrep
	default:
		return core.PhaseGameOver
	}
}
