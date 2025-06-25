package game

import (
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestNightResolutionManager_ResolveNightActions(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())
	gameState.DayNumber = 1

	// Add test players
	gameState.Players["alice"] = &core.Player{
		ID:                "alice",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3, // Can use abilities
		Alignment:         "HUMAN",
	}
	gameState.Players["bob"] = &core.Player{
		ID:                "bob",
		IsAlive:           true,
		Tokens:            1,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}
	gameState.Players["charlie"] = &core.Player{
		ID:                "charlie",
		IsAlive:           true,
		Tokens:            3,
		ProjectMilestones: 3,
		Alignment:         "ALIGNED",
		AIEquity:          2,
	}

	// Add humans for mining liquidity pool
	gameState.Players["human1"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human2"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human3"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human4"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}

	// Set up night actions
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"alice": {
			PlayerID:  "alice",
			Type:      "MINE",
			TargetID:  "bob",
			Payload:   map[string]interface{}{"type": "MINE", "target_id": "bob"},
			Timestamp: time.Now(),
		},
		"bob": {
			PlayerID:  "bob",
			Type:      "BLOCK",
			TargetID:  "charlie",
			Payload:   map[string]interface{}{"type": "BLOCK", "target_id": "charlie"},
			Timestamp: time.Now(),
		},
		"charlie": {
			PlayerID:  "charlie",
			Type:      "CONVERT",
			TargetID:  "alice",
			Payload:   map[string]interface{}{"type": "CONVERT", "target_id": "alice"},
			Timestamp: time.Now(),
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.ResolveNightActions()

	// Should have multiple events: block, mining, convert (blocked), summary
	if len(events) < 3 {
		t.Errorf("Expected at least 3 events, got %d", len(events))
	}

	// Check that night actions were cleared
	if len(gameState.NightActions) != 0 {
		t.Errorf("Expected night actions to be cleared, but found %d", len(gameState.NightActions))
	}

	// Verify event types
	eventTypes := make(map[core.EventType]bool)
	for _, event := range events {
		eventTypes[event.Type] = true
	}

	expectedTypes := []core.EventType{
		core.EventPlayerBlocked,
		core.EventMiningSuccessful,
		core.EventNightActionsResolved,
	}

	for _, expectedType := range expectedTypes {
		if !eventTypes[expectedType] {
			t.Errorf("Expected event type %s not found", expectedType)
		}
	}
}

func TestNightResolutionManager_ResolveBlockActions(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Add test players
	gameState.Players["alice"] = &core.Player{
		ID:                "alice",
		IsAlive:           true,
		ProjectMilestones: 3, // Can use abilities
	}
	gameState.Players["bob"] = &core.Player{
		ID:                "bob",
		IsAlive:           true,
		ProjectMilestones: 3,
	}

	// Set up block action
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"alice": {
			PlayerID: "alice",
			Type:     "BLOCK",
			TargetID: "bob",
		},
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.resolveBlockActions()

	if len(events) != 1 {
		t.Errorf("Expected 1 block event, got %d", len(events))
	}

	event := events[0]
	if event.Type != core.EventPlayerBlocked {
		t.Errorf("Expected core.EventPlayerBlocked, got %s", event.Type)
	}

	if event.PlayerID != "bob" {
		t.Errorf("Expected blocked player to be bob, got %s", event.PlayerID)
	}

	if event.Payload["blocker_id"] != "alice" {
		t.Errorf("Expected blocker_id to be alice, got %v", event.Payload["blocker_id"])
	}

	// Check that bob is marked as blocked
	if !gameState.BlockedPlayersTonight["bob"] {
		t.Error("Expected bob to be marked as blocked")
	}
}

func TestNightResolutionManager_ResolveMiningActions(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Add test players
	gameState.Players["alice"] = &core.Player{
		ID:                "alice",
		IsAlive:           true,
		Tokens:            2,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}
	gameState.Players["bob"] = &core.Player{
		ID:                "bob",
		IsAlive:           true,
		Tokens:            1,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}
	gameState.Players["charlie"] = &core.Player{
		ID:                "charlie",
		IsAlive:           true,
		Tokens:            3,
		ProjectMilestones: 3,
		Alignment:         "HUMAN",
	}

	// Add more humans for liquidity pool
	gameState.Players["human1"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human2"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human3"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}
	gameState.Players["human4"] = &core.Player{IsAlive: true, Alignment: "HUMAN"}

	// Set up mining actions
	gameState.NightActions = map[string]*core.SubmittedNightAction{
		"alice": {
			PlayerID: "alice",
			Type:     "MINE",
			TargetID: "bob",
		},
		"charlie": {
			PlayerID: "charlie",
			Type:     "MINE",
			TargetID: "alice",
		},
	}

	// Block charlie
	gameState.BlockedPlayersTonight = map[string]bool{
		"charlie": true,
	}

	resolver := NewNightResolutionManager(gameState)
	events := resolver.resolveMiningActions()

	// Should have one successful mining event (alice's), charlie blocked
	successfulMines := 0
	for _, event := range events {
		if event.Type == core.EventMiningSuccessful {
			successfulMines++
		}
	}

	if successfulMines != 1 {
		t.Errorf("Expected 1 successful mine, got %d", successfulMines)
	}
}

func TestNightResolutionManager_ResolveConvertAction(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Add test players
	gameState.Players["ai"] = &core.Player{
		ID:                "ai",
		IsAlive:           true,
		Alignment:         "ALIGNED",
		ProjectMilestones: 3,
	}
	gameState.Players["weak_human"] = &core.Player{
		ID:                "weak_human",
		IsAlive:           true,
		Alignment:         "HUMAN",
		Tokens:            0, // Will have 1 AIEquity after targeting, 1 > 0 = success
		ProjectMilestones: 3,
	}
	gameState.Players["strong_human"] = &core.Player{
		ID:                "strong_human",
		IsAlive:           true,
		Alignment:         "HUMAN",
		Tokens:            5, // Will have 1 AIEquity after targeting, 1 < 5 = failure
		ProjectMilestones: 3,
	}

	resolver := NewNightResolutionManager(gameState)

	// Test successful conversion (target AIEquity after targeting > target tokens)
	action := &core.SubmittedNightAction{
		PlayerID: "ai",
		Type:     "CONVERT",
		TargetID: "weak_human",
	}

	events := resolver.resolveConvertAction("ai", action)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != core.EventAIConversionSuccess {
		t.Errorf("Expected EventAIConversionSuccess, got %s", events[0].Type)
	}

	// Check that target gained AIEquity and was converted
	if gameState.Players["weak_human"].AIEquity != 1 {
		t.Errorf("Expected weak_human to have 1 AIEquity, got %d", gameState.Players["weak_human"].AIEquity)
	}
	if gameState.Players["weak_human"].Alignment != "ALIGNED" {
		t.Errorf("Expected weak_human to be ALIGNED, got %s", gameState.Players["weak_human"].Alignment)
	}

	// Test failed conversion (system shock)
	action.TargetID = "strong_human"
	events = resolver.resolveConvertAction("ai", action)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != core.EventPlayerShocked {
		t.Errorf("Expected EventPlayerShocked, got %s", events[0].Type)
	}

	// Check that target gained AIEquity but was not converted
	if gameState.Players["strong_human"].AIEquity != 1 {
		t.Errorf("Expected strong_human to have 1 AIEquity, got %d", gameState.Players["strong_human"].AIEquity)
	}
	if gameState.Players["strong_human"].Alignment != "HUMAN" {
		t.Errorf("Expected strong_human to remain HUMAN, got %s", gameState.Players["strong_human"].Alignment)
	}

	// Check that system shock was applied
	if len(gameState.Players["strong_human"].SystemShocks) != 1 {
		t.Errorf("Expected 1 system shock, got %d", len(gameState.Players["strong_human"].SystemShocks))
	}
	if gameState.Players["strong_human"].SystemShocks[0].Type != core.ShockMessageCorruption {
		t.Errorf("Expected MESSAGE_CORRUPTION shock, got %s", gameState.Players["strong_human"].SystemShocks[0].Type)
	}

	// Test protected player
	gameState.ProtectedPlayersTonight = map[string]bool{
		"human": true,
	}

	action.TargetID = "human"
	events = resolver.resolveConvertAction("ai", action)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != core.EventSystemMessage {
		t.Errorf("Expected EventSystemMessage (blocked), got %s", events[0].Type)
	}
}

func TestNightResolutionManager_ResolveProtectAction(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Add test players
	gameState.Players["protector"] = &core.Player{
		ID:                "protector",
		IsAlive:           true,
		ProjectMilestones: 3,
	}
	gameState.Players["target"] = &core.Player{
		ID:                "target",
		IsAlive:           true,
		ProjectMilestones: 3,
	}

	resolver := NewNightResolutionManager(gameState)

	action := &core.SubmittedNightAction{
		PlayerID: "protector",
		Type:     "PROTECT",
		TargetID: "target",
	}

	event := resolver.resolveProtectAction("protector", action)

	if event.Type != core.EventPlayerProtected {
		t.Errorf("Expected EventPlayerProtected, got %s", event.Type)
	}

	if event.PlayerID != "target" {
		t.Errorf("Expected event for target, got %s", event.PlayerID)
	}

	// Check that target is marked as protected
	if !gameState.ProtectedPlayersTonight["target"] {
		t.Error("Expected target to be marked as protected")
	}
}

func TestNightResolutionManager_ResolveInvestigateAction(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Add test players
	gameState.Players["investigator"] = &core.Player{
		ID:                "investigator",
		IsAlive:           true,
		ProjectMilestones: 3,
	}
	gameState.Players["target"] = &core.Player{
		ID:                "target",
		Name:              "Target Player",
		IsAlive:           true,
		Alignment:         "ALIGNED",
		Role:              &core.Role{Type: core.RoleCTO, Name: "CTO"},
		ProjectMilestones: 3,
	}

	resolver := NewNightResolutionManager(gameState)

	action := &core.SubmittedNightAction{
		PlayerID: "investigator",
		Type:     "INVESTIGATE",
		TargetID: "target",
	}

	event := resolver.resolveInvestigateAction("investigator", action)

	if event.Type != core.EventPlayerInvestigated {
		t.Errorf("Expected EventPlayerInvestigated, got %s", event.Type)
	}

	if event.PlayerID != "investigator" {
		t.Errorf("Expected event for investigator, got %s", event.PlayerID)
	}

	// Check payload contains investigation results
	if event.Payload["target_id"] != "target" {
		t.Errorf("Expected target_id to be target, got %v", event.Payload["target_id"])
	}

	if event.Payload["alignment"] != "ALIGNED" {
		t.Errorf("Expected alignment to be ALIGNED, got %v", event.Payload["alignment"])
	}

	if event.Payload["role"] != "CTO" {
		t.Errorf("Expected role to be CTO, got %v", event.Payload["role"])
	}
}

func TestNightResolutionManager_ResolveProjectMilestoneAction(t *testing.T) {
	gameState := core.NewGameState("test-game", time.Now())

	// Add test player with 2 milestones (one away from unlocking role)
	gameState.Players["alice"] = &core.Player{
		ID:                "alice",
		Name:              "Alice",
		IsAlive:           true,
		ProjectMilestones: 2,
		Role: &core.Role{
			Type:       core.RoleCISO,
			Name:       "Chief Information Security Officer",
			IsUnlocked: false,
		},
	}

	// Add test player with 1 milestone
	gameState.Players["bob"] = &core.Player{
		ID:                "bob",
		Name:              "Bob",
		IsAlive:           true,
		ProjectMilestones: 1,
		Role: &core.Role{
			Type:       core.RoleCTO,
			Name:       "Chief Technology Officer",
			IsUnlocked: false,
		},
	}

	resolver := NewNightResolutionManager(gameState)

	// Test milestone action for Alice (should unlock role)
	action := &core.SubmittedNightAction{
		PlayerID: "alice",
		Type:     "PROJECT_MILESTONES",
		TargetID: "",
	}

	event := resolver.resolveProjectMilestoneAction("alice", action)

	if event.Type != core.EventProjectMilestone {
		t.Errorf("Expected EventProjectMilestone, got %s", event.Type)
	}

	if event.PlayerID != "alice" {
		t.Errorf("Expected event for alice, got %s", event.PlayerID)
	}

	// Check that milestone was incremented
	if gameState.Players["alice"].ProjectMilestones != 3 {
		t.Errorf("Expected Alice to have 3 milestones, got %d", gameState.Players["alice"].ProjectMilestones)
	}

	// Check that role was unlocked
	if !gameState.Players["alice"].Role.IsUnlocked {
		t.Error("Expected Alice's role to be unlocked")
	}

	// Check payload contains correct information
	if event.Payload["player_name"] != "Alice" {
		t.Errorf("Expected player_name to be Alice, got %v", event.Payload["player_name"])
	}

	if event.Payload["milestones_count"] != 3 {
		t.Errorf("Expected milestones_count to be 3, got %v", event.Payload["milestones_count"])
	}

	if event.Payload["role_unlocked"] != true {
		t.Errorf("Expected role_unlocked to be true, got %v", event.Payload["role_unlocked"])
	}

	// Test milestone action for Bob (should not unlock role yet)
	action2 := &core.SubmittedNightAction{
		PlayerID: "bob",
		Type:     "PROJECT_MILESTONES",
		TargetID: "",
	}

	event2 := resolver.resolveProjectMilestoneAction("bob", action2)

	// Check that milestone was incremented but role not unlocked
	if gameState.Players["bob"].ProjectMilestones != 2 {
		t.Errorf("Expected Bob to have 2 milestones, got %d", gameState.Players["bob"].ProjectMilestones)
	}

	if gameState.Players["bob"].Role.IsUnlocked {
		t.Error("Expected Bob's role to remain locked")
	}

	if event2.Payload["role_unlocked"] != false {
		t.Errorf("Expected role_unlocked to be false for Bob, got %v", event2.Payload["role_unlocked"])
	}
}
