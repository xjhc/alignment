package game

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// GameManagementHandler handles game lifecycle and management actions
type GameManagementHandler struct {
	gameLifecycleManager GameLifecycleManager
}

// GameLifecycleManager interface for game lifecycle operations
type GameLifecycleManager interface {
	HandleInitializeGame(action core.Action) ([]core.Event, error)
	HandleLeaveGame(action core.Action) ([]core.Event, error)
	HandleAbandonGame(action core.Action) ([]core.Event, error)
	HandleSetPlayerConnectionStatus(action core.Action) ([]core.Event, error)
	HandleAbandonPlayer(action core.Action) ([]core.Event, error)
	HandlePhaseTransition(action core.Action) ([]core.Event, error)
	HandleSyncLobbyState(action core.Action) ([]core.Event, error)
}

// NewGameManagementHandler creates a new GameManagementHandler
func NewGameManagementHandler(gameLifecycleManager GameLifecycleManager) *GameManagementHandler {
	return &GameManagementHandler{
		gameLifecycleManager: gameLifecycleManager,
	}
}

// Handle processes game management actions
func (h *GameManagementHandler) Handle(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	switch action.Type {
	case core.ActionType("INITIALIZE_GAME"):
		return h.gameLifecycleManager.HandleInitializeGame(action)
	case core.ActionLeaveGame:
		return h.gameLifecycleManager.HandleLeaveGame(action)
	case core.ActionAbandonGame:
		return h.gameLifecycleManager.HandleAbandonGame(action)
	case core.ActionSetPlayerConnectionStatus:
		return h.gameLifecycleManager.HandleSetPlayerConnectionStatus(action)
	case core.ActionReconnect:
		return h.handleReconnect(action)
	case core.ActionAbandonPlayer:
		return h.gameLifecycleManager.HandleAbandonPlayer(action)
	case core.ActionType("PHASE_TRANSITION"):
		return h.gameLifecycleManager.HandlePhaseTransition(action)
	case core.ActionSyncLobbyState:
		return h.gameLifecycleManager.HandleSyncLobbyState(action)
	case core.ActionAssignCorporateMandate:
		return h.handleAssignCorporateMandate(state, action, acker)
	default:
		return nil, fmt.Errorf("unsupported game management action type: %s", action.Type)
	}
}

// handleAssignCorporateMandate assigns a corporate mandate to the game
func (h *GameManagementHandler) handleAssignCorporateMandate(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Only system actions are allowed
	if action.PlayerID != "SYSTEM" {
		return nil, fmt.Errorf("only system can assign corporate mandates")
	}

	// Create corporate mandate manager
	mandateManager := NewCorporateMandateManager(state)

	// Assign a random mandate
	mandate := mandateManager.AssignRandomMandate()
	if mandate == nil {
		return nil, fmt.Errorf("failed to assign corporate mandate")
	}

	// Generate mandate activated event
	event := core.Event{
		ID:        fmt.Sprintf("mandate_activated_%s_%d", action.GameID, time.Now().UnixNano()),
		Type:      core.EventMandateActivated,
		GameID:    action.GameID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"mandate_type":        mandate.Type,
			"mandate_name":        mandate.Name,
			"mandate_description": mandate.Description,
			"mandate_effects":     mandate.Effects,
			"day_number":          state.DayNumber,
		},
	}

	return []core.Event{event}, nil
}

// handleReconnect generates a player reconnected event
func (h *GameManagementHandler) handleReconnect(action core.Action) ([]core.Event, error) {
	// Generate a reconnection event to update the player's connection status
	event := core.Event{
		ID:        fmt.Sprintf("player_reconnected_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventPlayerReconnected,
		GameID:    action.GameID,
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"player_id": action.PlayerID,
		},
	}

	return []core.Event{event}, nil
}