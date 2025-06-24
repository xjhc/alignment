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
		AIEquity:          3,
		ProjectMilestones: 3,
	}
	gameState.Players["human"] = &core.Player{
		ID:                "human",
		IsAlive:           true,
		Alignment:         "HUMAN",
		Tokens:            2, // Less than AI equity
		ProjectMilestones: 3,
	}
	gameState.Players["strong_human"] = &core.Player{
		ID:                "strong_human",
		IsAlive:           true,
		Alignment:         "HUMAN",
		Tokens:            5, // More than AI equity
		ProjectMilestones: 3,
	}

	resolver := NewNightResolutionManager(gameState)

	// Test successful conversion (AI equity > human tokens)
	action := &core.SubmittedNightAction{
		PlayerID: "ai",
		Type:     "CONVERT",
		TargetID: "human",
	}

	events := resolver.resolveConvertAction("ai", action)

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	if events[0].Type != core.EventAIConversionSuccess {
		t.Errorf("Expected EventAIConversionSuccess, got %s", events[0].Type)
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
