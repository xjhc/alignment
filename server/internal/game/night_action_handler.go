package game

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// NightActionHandler handles night action logic
type NightActionHandler struct {
	miningManager      interface{ HandleMineAction(action core.Action) ([]core.Event, error) }
	roleAbilityManager interface{ HandleNightAction(action core.Action) ([]core.Event, error) }
}

// NewNightActionHandler creates a new NightActionHandler
func NewNightActionHandler(miningManager interface{ HandleMineAction(action core.Action) ([]core.Event, error) }, roleAbilityManager interface{ HandleNightAction(action core.Action) ([]core.Event, error) }) *NightActionHandler {
	return &NightActionHandler{
		miningManager:      miningManager,
		roleAbilityManager: roleAbilityManager,
	}
}

// Handle processes night actions
func (h *NightActionHandler) Handle(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	actionType, _ := action.Payload["action_type"].(string)

	switch actionType {
	case "MINE_TOKENS", "MINE":
		return h.miningManager.HandleMineAction(action)
	case "ATTEMPT_CONVERSION":
		// Store the night action for later resolution
		return h.storeNightAction(state, action, acker)
	case "PROJECT_MILESTONES":
		// Store the night action for later resolution
		return h.storeNightAction(state, action, acker)
	default:
		// Handle other night actions through role ability manager
		return h.roleAbilityManager.HandleNightAction(action)
	}
}

// storeNightAction stores a night action for resolution at night end
func (h *NightActionHandler) storeNightAction(state *core.GameState, action core.Action, acker ActionAcker) ([]core.Event, error) {
	// Note: In the Strategy Pattern, we don't mutate the state directly
	// Instead, we create an event that will be processed to update the state
	
	actionType, _ := action.Payload["action_type"].(string)
	targetID, _ := action.Payload["target_player_id"].(string)

	// Create night action submitted event
	event := core.Event{
		ID:        fmt.Sprintf("night_action_%s_%d", action.PlayerID, time.Now().UnixNano()),
		Type:      core.EventNightActionSubmitted,
		GameID:    acker.GetGameID(),
		PlayerID:  action.PlayerID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"action_type": actionType,
			"target_id":   targetID,
			"full_payload": action.Payload,
		},
	}

	return []core.Event{event}, nil
}