package game

import (
	"testing"

	"github.com/xjhc/alignment/core"
)

func TestKPIManager_TrackPlayerEliminated_InquisitorProgress(t *testing.T) {
	// Create test game state
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Players: map[string]*core.Player{
			"voter1": {
				ID:        "voter1",
				Name:      "Test Voter",
				IsAlive:   true,
				Alignment: "HUMAN",
				PersonalKPI: &core.PersonalKPI{
					Type:        core.KPIInquisitor,
					Progress:    0,
					Target:      3,
					IsCompleted: false,
				},
			},
			"eliminated": {
				ID:        "eliminated",
				Name:      "AI Player",
				IsAlive:   false,
				Alignment: "ALIGNED",
			},
		},
		VoteState: &core.VoteState{
			Votes: map[string]string{
				"voter1": "eliminated",
			},
		},
	}

	// Create KPI manager
	manager := NewKPIManager(gameState)

	// Track elimination
	events := manager.TrackPlayerEliminated("eliminated")

	// Should generate progress and notification events
	if len(events) != 2 {
		t.Fatalf("Expected 2 events (progress + notification), got %d", len(events))
	}

	// Check progress event
	progressEvent := events[0]
	if progressEvent.Type != core.EventKPIProgress {
		t.Errorf("Expected KPI_PROGRESS event, got %s", progressEvent.Type)
	}
	if progressEvent.PlayerID != "voter1" {
		t.Errorf("Expected event for voter1, got %s", progressEvent.PlayerID)
	}

	// Check notification event
	notificationEvent := events[1]
	if notificationEvent.Type != core.EventPrivateNotification {
		t.Errorf("Expected PRIVATE_NOTIFICATION event, got %s", notificationEvent.Type)
	}
	if notificationEvent.PlayerID != "voter1" {
		t.Errorf("Expected notification for voter1, got %s", notificationEvent.PlayerID)
	}

	// Check notification payload
	notifType, ok := notificationEvent.Payload["type"].(string)
	if !ok || notifType != "kpi_progress" {
		t.Errorf("Expected notification type 'kpi_progress', got %v", notifType)
	}
}

func TestKPIManager_TrackPlayerEliminated_InquisitorCompletion(t *testing.T) {
	// Create test game state with player at 2/3 progress
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 1,
		Players: map[string]*core.Player{
			"voter1": {
				ID:        "voter1",
				Name:      "Test Voter",
				IsAlive:   true,
				Alignment: "HUMAN",
				PersonalKPI: &core.PersonalKPI{
					Type:        core.KPIInquisitor,
					Progress:    2, // Almost complete
					Target:      3,
					IsCompleted: false,
				},
			},
			"eliminated": {
				ID:        "eliminated",
				Name:      "AI Player",
				IsAlive:   false,
				Alignment: "ALIGNED",
			},
		},
		VoteState: &core.VoteState{
			Votes: map[string]string{
				"voter1": "eliminated",
			},
		},
	}

	// Create KPI manager
	manager := NewKPIManager(gameState)

	// Track elimination (should complete KPI)
	events := manager.TrackPlayerEliminated("eliminated")

	// Should generate progress, notification, completion, completion notification, and token events
	if len(events) != 5 {
		t.Fatalf("Expected 5 events (progress + notification + completion + completion notification + tokens), got %d", len(events))
	}

	// Check that KPI completion event is generated
	var completionEvent *core.Event
	for i := range events {
		if events[i].Type == core.EventKPICompleted {
			completionEvent = &events[i]
			break
		}
	}

	if completionEvent == nil {
		t.Fatal("Expected KPI_COMPLETED event not found")
	}

	if completionEvent.PlayerID != "voter1" {
		t.Errorf("Expected completion event for voter1, got %s", completionEvent.PlayerID)
	}
}

func TestKPIManager_TrackNightSurvival_GuardianProgress(t *testing.T) {
	// Create test game state with CISO alive
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Players: map[string]*core.Player{
			"guardian": {
				ID:        "guardian",
				Name:      "Guardian Player",
				IsAlive:   true,
				Alignment: "HUMAN",
				PersonalKPI: &core.PersonalKPI{
					Type:        core.KPIGuardian,
					Progress:    1,
					Target:      4,
					IsCompleted: false,
				},
			},
			"ciso": {
				ID:        "ciso",
				Name:      "CISO Player",
				IsAlive:   true,
				Alignment: "HUMAN",
				Role: &core.Role{
					Type: core.RoleCISO,
				},
			},
		},
	}

	// Create KPI manager
	manager := NewKPIManager(gameState)

	// Track night survival
	events := manager.TrackNightSurvival()

	// Should generate progress and notification events
	if len(events) != 2 {
		t.Fatalf("Expected 2 events (progress + notification), got %d", len(events))
	}

	// Check progress event
	progressEvent := events[0]
	if progressEvent.Type != core.EventKPIProgress {
		t.Errorf("Expected KPI_PROGRESS event, got %s", progressEvent.Type)
	}
	if progressEvent.PlayerID != "guardian" {
		t.Errorf("Expected event for guardian, got %s", progressEvent.PlayerID)
	}
}

func TestKPIManager_TrackNightSurvival_GuardianCompletion(t *testing.T) {
	// Create test game state with CISO alive on day 4
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 4,
		Players: map[string]*core.Player{
			"guardian": {
				ID:        "guardian",
				Name:      "Guardian Player",
				IsAlive:   true,
				Alignment: "HUMAN",
				PersonalKPI: &core.PersonalKPI{
					Type:        core.KPIGuardian,
					Progress:    3,
					Target:      4,
					IsCompleted: false,
				},
			},
			"ciso": {
				ID:        "ciso",
				Name:      "CISO Player",
				IsAlive:   true,
				Alignment: "HUMAN",
				Role: &core.Role{
					Type: core.RoleCISO,
				},
			},
		},
	}

	// Create KPI manager
	manager := NewKPIManager(gameState)

	// Track night survival (should complete KPI)
	events := manager.TrackNightSurvival()

	// Should generate completion and completion notification events
	if len(events) != 2 {
		t.Fatalf("Expected 2 events (completion + completion notification), got %d", len(events))
	}

	// Check that KPI completion event is generated
	var completionEvent *core.Event
	for i := range events {
		if events[i].Type == core.EventKPICompleted {
			completionEvent = &events[i]
			break
		}
	}

	if completionEvent == nil {
		t.Fatal("Expected KPI_COMPLETED event not found")
	}

	if completionEvent.PlayerID != "guardian" {
		t.Errorf("Expected completion event for guardian, got %s", completionEvent.PlayerID)
	}
}

func TestKPIManager_CheckGameEndKPIs_SuccessionPlanner(t *testing.T) {
	// Create test game state with exactly 2 humans alive
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 3,
		Players: map[string]*core.Player{
			"planner": {
				ID:        "planner",
				Name:      "Succession Planner",
				IsAlive:   true,
				Alignment: "HUMAN",
				PersonalKPI: &core.PersonalKPI{
					Type:        core.KPISuccessionPlanner,
					Progress:    0,
					Target:      1,
					IsCompleted: false,
				},
			},
			"human2": {
				ID:        "human2",
				Name:      "Human 2",
				IsAlive:   true,
				Alignment: "HUMAN",
			},
			"dead_human": {
				ID:        "dead_human",
				Name:      "Dead Human",
				IsAlive:   false,
				Alignment: "HUMAN",
			},
		},
	}

	// Create KPI manager
	manager := NewKPIManager(gameState)

	// Check game end KPIs
	events := manager.CheckGameEndKPIs()

	// Should generate completion and completion notification events
	if len(events) != 2 {
		t.Fatalf("Expected 2 events (completion + completion notification), got %d", len(events))
	}

	// Check that KPI completion event is generated
	var completionEvent *core.Event
	for i := range events {
		if events[i].Type == core.EventKPICompleted {
			completionEvent = &events[i]
			break
		}
	}

	if completionEvent == nil {
		t.Fatal("Expected KPI_COMPLETED event not found")
	}

	if completionEvent.PlayerID != "planner" {
		t.Errorf("Expected completion event for planner, got %s", completionEvent.PlayerID)
	}
}

func TestKPIManager_TrackUnanimousElimination_Scapegoat(t *testing.T) {
	// Create test game state with scapegoat player eliminated unanimously
	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Players: map[string]*core.Player{
			"scapegoat": {
				ID:        "scapegoat",
				Name:      "Scapegoat Player",
				IsAlive:   false,
				Alignment: "HUMAN",
				PersonalKPI: &core.PersonalKPI{
					Type:        core.KPIScapegoat,
					Progress:    0,
					Target:      1,
					IsCompleted: false,
				},
			},
		},
		VoteState: &core.VoteState{
			Votes: map[string]string{
				"voter1": "scapegoat",
				"voter2": "scapegoat",
				"voter3": "scapegoat",
			},
		},
	}

	// Create KPI manager
	manager := NewKPIManager(gameState)

	// Track unanimous elimination
	events := manager.TrackUnanimousElimination("scapegoat")

	// Should generate completion and completion notification events
	if len(events) != 2 {
		t.Fatalf("Expected 2 events (completion + completion notification), got %d", len(events))
	}

	// Check that KPI completion event is generated
	var completionEvent *core.Event
	for i := range events {
		if events[i].Type == core.EventKPICompleted {
			completionEvent = &events[i]
			break
		}
	}

	if completionEvent == nil {
		t.Fatal("Expected KPI_COMPLETED event not found")
	}

	if completionEvent.PlayerID != "scapegoat" {
		t.Errorf("Expected completion event for scapegoat, got %s", completionEvent.PlayerID)
	}
}

// Helper function to override time for testing (mimicking the pattern in the main code)
func init() {
	// For consistency with the actual implementation, we use the current time
	// In a more sophisticated test setup, we might want to mock this
}