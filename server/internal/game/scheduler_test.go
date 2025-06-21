package game

import (
	"sync"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

// TestScheduler_BasicTimerScheduling tests basic timer functionality
func TestScheduler_BasicTimerScheduling(t *testing.T) {
	callbackCalled := false
	var callbackTimer Timer
	var mu sync.Mutex

	callback := func(timer Timer) {
		mu.Lock()
		defer mu.Unlock()
		callbackCalled = true
		callbackTimer = timer
	}

	scheduler := NewScheduler(callback)
	// Use faster ticker for testing
	scheduler.ticker = time.NewTicker(10 * time.Millisecond)
	scheduler.Start()
	defer scheduler.Stop()

	// Schedule a timer that expires in 50ms
	timer := Timer{
		ID:        "test-timer",
		GameID:    "test-game",
		Type:      TimerPhaseEnd,
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
		Action: TimerAction{
			Type: "TEST_ACTION",
			Payload: map[string]interface{}{
				"test": "value",
			},
		},
	}

	scheduler.ScheduleTimer(timer)

	// Wait for timer to expire
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !callbackCalled {
		t.Error("Expected timer callback to be called")
	}

	if callbackTimer.ID != "test-timer" {
		t.Errorf("Expected callback timer ID to be 'test-timer', got %s", callbackTimer.ID)
	}

	if callbackTimer.GameID != "test-game" {
		t.Errorf("Expected callback timer game ID to be 'test-game', got %s", callbackTimer.GameID)
	}
}

// TestScheduler_TimerCancellation tests timer cancellation
func TestScheduler_TimerCancellation(t *testing.T) {
	callbackCalled := false
	var mu sync.Mutex

	callback := func(timer Timer) {
		mu.Lock()
		defer mu.Unlock()
		callbackCalled = true
	}

	scheduler := NewScheduler(callback)
	scheduler.Start()
	defer scheduler.Stop()

	// Schedule a timer
	timer := Timer{
		ID:        "cancel-timer",
		GameID:    "test-game",
		Type:      TimerPhaseEnd,
		ExpiresAt: time.Now().Add(100 * time.Millisecond),
		Action:    TimerAction{Type: "TEST_ACTION"},
	}

	scheduler.ScheduleTimer(timer)

	// Cancel the timer before it expires
	scheduler.CancelTimer("cancel-timer")

	// Wait past expiration time
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if callbackCalled {
		t.Error("Expected cancelled timer callback to not be called")
	}
}

// TestScheduler_GameTimerCancellation tests cancelling all timers for a game
func TestScheduler_GameTimerCancellation(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func(timer Timer) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	scheduler := NewScheduler(callback)
	scheduler.Start()
	defer scheduler.Stop()

	// Schedule multiple timers for the same game
	for i := 0; i < 3; i++ {
		timer := Timer{
			ID:        "timer-" + string(rune('1'+i)),
			GameID:    "test-game",
			Type:      TimerPhaseEnd,
			ExpiresAt: time.Now().Add(100 * time.Millisecond),
			Action:    TimerAction{Type: "TEST_ACTION"},
		}
		scheduler.ScheduleTimer(timer)
	}

	// Schedule a timer for a different game
	otherTimer := Timer{
		ID:        "other-timer",
		GameID:    "other-game",
		Type:      TimerPhaseEnd,
		ExpiresAt: time.Now().Add(100 * time.Millisecond),
		Action:    TimerAction{Type: "TEST_ACTION"},
	}
	scheduler.ScheduleTimer(otherTimer)

	// Cancel all timers for test-game
	scheduler.CancelGameTimers("test-game")

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	// Only the timer from other-game should have fired
	if callbackCount != 1 {
		t.Errorf("Expected 1 callback (from other-game), got %d", callbackCount)
	}
}

// TestScheduler_MultipleTimers tests handling multiple timers
func TestScheduler_MultipleTimers(t *testing.T) {
	callbackCount := 0
	var mu sync.Mutex

	callback := func(timer Timer) {
		mu.Lock()
		defer mu.Unlock()
		callbackCount++
	}

	scheduler := NewScheduler(callback)
	scheduler.Start()
	defer scheduler.Stop()

	// Schedule multiple timers with different expiration times
	for i := 0; i < 5; i++ {
		timer := Timer{
			ID:        "timer-" + string(rune('1'+i)),
			GameID:    "test-game",
			Type:      TimerPhaseEnd,
			ExpiresAt: time.Now().Add(time.Duration(50*(i+1)) * time.Millisecond),
			Action:    TimerAction{Type: "TEST_ACTION"},
		}
		scheduler.ScheduleTimer(timer)
	}

	// Wait for all timers to expire
	time.Sleep(400 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if callbackCount != 5 {
		t.Errorf("Expected 5 callbacks, got %d", callbackCount)
	}
}

// TestScheduler_ActiveTimers tests active timer tracking
func TestScheduler_ActiveTimers(t *testing.T) {
	scheduler := NewScheduler(nil)
	scheduler.Start()
	defer scheduler.Stop()

	// Initially no timers
	activeTimers := scheduler.GetActiveTimers()
	if len(activeTimers) != 0 {
		t.Errorf("Expected 0 active timers initially, got %d", len(activeTimers))
	}

	// Schedule some timers
	for i := 0; i < 3; i++ {
		timer := Timer{
			ID:        "timer-" + string(rune('1'+i)),
			GameID:    "test-game",
			Type:      TimerPhaseEnd,
			ExpiresAt: time.Now().Add(1 * time.Second), // Long expiration
			Action:    TimerAction{Type: "TEST_ACTION"},
		}
		scheduler.ScheduleTimer(timer)
	}

	// Should have 3 active timers
	activeTimers = scheduler.GetActiveTimers()
	if len(activeTimers) != 3 {
		t.Errorf("Expected 3 active timers, got %d", len(activeTimers))
	}

	// Cancel one timer
	scheduler.CancelTimer("timer-1")

	// Should have 2 active timers
	activeTimers = scheduler.GetActiveTimers()
	if len(activeTimers) != 2 {
		t.Errorf("Expected 2 active timers after cancellation, got %d", len(activeTimers))
	}
}

// TestPhaseManager_PhaseTransitions tests automatic phase transitions
func TestPhaseManager_PhaseTransitions(t *testing.T) {
	// Track scheduled timers
	var scheduledTimers []Timer
	var mu sync.Mutex

	callback := func(timer Timer) {
		mu.Lock()
		defer mu.Unlock()
		scheduledTimers = append(scheduledTimers, timer)
	}

	scheduler := NewScheduler(callback)
	scheduler.Start()
	defer scheduler.Stop()

	settings := core.GameSettings{
		SitrepDuration:     15 * time.Second,
		PulseCheckDuration: 30 * time.Second,
		DiscussionDuration: 2 * time.Minute,
		ExtensionDuration:  15 * time.Second,
		NominationDuration: 30 * time.Second,
		TrialDuration:      30 * time.Second,
		VerdictDuration:    30 * time.Second,
		NightDuration:      30 * time.Second,
	}

	pm := NewPhaseManager(scheduler, "test-game", settings)

	// Test SITREP phase transition
	phaseStartTime := time.Now()
	pm.SchedulePhaseTransition(core.PhaseSitrep, phaseStartTime)

	// Wait a short time for scheduling
	time.Sleep(10 * time.Millisecond)

	// Check that timer was scheduled correctly
	activeTimers := scheduler.GetActiveTimers()
	if len(activeTimers) != 1 {
		t.Fatalf("Expected 1 active timer, got %d", len(activeTimers))
	}

	var sitrepTimer *Timer
	for _, timer := range activeTimers {
		sitrepTimer = timer
		break
	}

	if sitrepTimer.GameID != "test-game" {
		t.Errorf("Expected timer for test-game, got %s", sitrepTimer.GameID)
	}

	if sitrepTimer.Type != TimerPhaseEnd {
		t.Errorf("Expected phase end timer, got %s", sitrepTimer.Type)
	}

	// Check timer expiration time (should be start time + SITREP duration)
	expectedExpiry := phaseStartTime.Add(settings.SitrepDuration)
	if sitrepTimer.ExpiresAt.Sub(expectedExpiry).Abs() > time.Millisecond {
		t.Errorf("Expected timer to expire at %v, got %v", expectedExpiry, sitrepTimer.ExpiresAt)
	}

	// Check timer action payload
	nextPhase, exists := sitrepTimer.Action.Payload["next_phase"].(string)
	if !exists || nextPhase != string(core.PhasePulseCheck) {
		t.Errorf("Expected next phase to be PULSE_CHECK, got %v", nextPhase)
	}
}

// TestPhaseManager_AllPhaseTransitions tests the complete phase cycle
func TestPhaseManager_AllPhaseTransitions(t *testing.T) {
	scheduler := NewScheduler(nil)
	scheduler.Start()
	defer scheduler.Stop()

	settings := core.GameSettings{
		SitrepDuration:     100 * time.Millisecond,
		PulseCheckDuration: 100 * time.Millisecond,
		DiscussionDuration: 100 * time.Millisecond,
		ExtensionDuration:  100 * time.Millisecond,
		NominationDuration: 100 * time.Millisecond,
		TrialDuration:      100 * time.Millisecond,
		VerdictDuration:    100 * time.Millisecond,
		NightDuration:      100 * time.Millisecond,
	}

	pm := NewPhaseManager(scheduler, "test-game", settings)

	// Test all phase transitions
	phases := []core.PhaseType{
		core.PhaseSitrep,
		core.PhasePulseCheck,
		core.PhaseDiscussion,
		core.PhaseExtension,
		core.PhaseNomination,
		core.PhaseTrial,
		core.PhaseVerdict,
		core.PhaseNight,
	}

	expectedNextPhases := []core.PhaseType{
		core.PhasePulseCheck,
		core.PhaseDiscussion,
		core.PhaseExtension,
		core.PhaseNomination,
		core.PhaseTrial,
		core.PhaseVerdict,
		core.PhaseNight,
		core.PhaseSitrep, // Night wraps back to SITREP
	}

	for i, phase := range phases {
		pm.SchedulePhaseTransition(phase, time.Now())

		activeTimers := scheduler.GetActiveTimers()
		if len(activeTimers) != i+1 {
			t.Errorf("Expected %d active timer(s) after scheduling %s, got %d", i+1, phase, len(activeTimers))
		}

		// Find the timer for this phase
		var phaseTimer *Timer
		expectedTimerID := "test-game_phase_" + string(phase)
		for _, timer := range activeTimers {
			if timer.ID == expectedTimerID {
				phaseTimer = timer
				break
			}
		}

		if phaseTimer == nil {
			t.Fatalf("Expected to find timer with ID %s", expectedTimerID)
		}

		// Check next phase in timer action
		nextPhase, exists := phaseTimer.Action.Payload["next_phase"].(string)
		if !exists {
			t.Errorf("Expected next_phase in timer action for %s", phase)
			continue
		}

		if nextPhase != string(expectedNextPhases[i]) {
			t.Errorf("Expected next phase after %s to be %s, got %s", phase, expectedNextPhases[i], nextPhase)
		}
	}
}

// TestPhaseManager_CancelTransitions tests cancelling phase transitions
func TestPhaseManager_CancelTransitions(t *testing.T) {
	scheduler := NewScheduler(nil)
	scheduler.Start()
	defer scheduler.Stop()

	settings := core.GameSettings{
		SitrepDuration: 1 * time.Second,
	}

	pm := NewPhaseManager(scheduler, "test-game", settings)

	// Schedule a transition
	pm.SchedulePhaseTransition(core.PhaseSitrep, time.Now())

	// Verify timer was scheduled
	activeTimers := scheduler.GetActiveTimers()
	if len(activeTimers) != 1 {
		t.Fatalf("Expected 1 active timer, got %d", len(activeTimers))
	}

	// Cancel all transitions for the game
	pm.CancelPhaseTransitions()

	// Verify timer was cancelled
	activeTimers = scheduler.GetActiveTimers()
	if len(activeTimers) != 0 {
		t.Errorf("Expected 0 active timers after cancellation, got %d", len(activeTimers))
	}
}

// TestPhaseDurationHelpers tests phase duration helper functions
func TestPhaseDurationHelpers(t *testing.T) {
	settings := core.GameSettings{
		SitrepDuration:     15 * time.Second,
		PulseCheckDuration: 30 * time.Second,
		DiscussionDuration: 2 * time.Minute,
		ExtensionDuration:  15 * time.Second,
		NominationDuration: 30 * time.Second,
		TrialDuration:      30 * time.Second,
		VerdictDuration:    30 * time.Second,
		NightDuration:      30 * time.Second,
	}

	// Test getPhaseDuration
	testCases := []struct {
		phase    core.PhaseType
		expected time.Duration
	}{
		{core.PhaseSitrep, 15 * time.Second},
		{core.PhasePulseCheck, 30 * time.Second},
		{core.PhaseDiscussion, 2 * time.Minute},
		{core.PhaseExtension, 15 * time.Second},
		{core.PhaseNomination, 30 * time.Second},
		{core.PhaseTrial, 30 * time.Second},
		{core.PhaseVerdict, 30 * time.Second},
		{core.PhaseNight, 30 * time.Second},
		{core.PhaseLobby, 0}, // Unknown phase
	}

	for _, tc := range testCases {
		duration := getPhaseDuration(tc.phase, settings)
		if duration != tc.expected {
			t.Errorf("Expected duration for %s to be %v, got %v", tc.phase, tc.expected, duration)
		}
	}

	// Test getNextPhase
	transitionCases := []struct {
		current core.PhaseType
		next    core.PhaseType
	}{
		{core.PhaseSitrep, core.PhasePulseCheck},
		{core.PhasePulseCheck, core.PhaseDiscussion},
		{core.PhaseDiscussion, core.PhaseExtension},
		{core.PhaseExtension, core.PhaseNomination},
		{core.PhaseNomination, core.PhaseTrial},
		{core.PhaseTrial, core.PhaseVerdict},
		{core.PhaseVerdict, core.PhaseNight},
		{core.PhaseNight, core.PhaseSitrep},
		{core.PhaseLobby, core.PhaseGameOver}, // Unknown phase
	}

	for _, tc := range transitionCases {
		nextPhase := getNextPhase(tc.current)
		if nextPhase != tc.next {
			t.Errorf("Expected next phase after %s to be %s, got %s", tc.current, tc.next, nextPhase)
		}
	}
}
