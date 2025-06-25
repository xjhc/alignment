package game

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// LiaisonProtocolManager handles the LIAISON Protocol catch-up mechanic
type LiaisonProtocolManager struct {
	gameState *core.GameState
	rng       *rand.Rand
}

// NewLiaisonProtocolManager creates a new LIAISON protocol manager
func NewLiaisonProtocolManager(gameState *core.GameState) *LiaisonProtocolManager {
	return &LiaisonProtocolManager{
		gameState: gameState,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CheckProtocolTrigger evaluates if the LIAISON Protocol should be activated
func (lpm *LiaisonProtocolManager) CheckProtocolTrigger() bool {
	alivePlayers := 0
	aiPlayers := 0

	for _, player := range lpm.gameState.Players {
		if player.IsAlive {
			alivePlayers++
			if player.Alignment == "ai" {
				aiPlayers++
			}
		}
	}

	if alivePlayers == 0 {
		return false
	}

	// Check if AI faction is at 40% or more of living players
	aiPercentage := float64(aiPlayers) / float64(alivePlayers)
	return aiPercentage >= 0.40
}

// ActivateProtocol triggers the LIAISON Protocol and returns the generated events
func (lpm *LiaisonProtocolManager) ActivateProtocol() []core.Event {
	var events []core.Event

	log.Printf("[LiaisonProtocol] LIAISON Protocol activated - AI faction at 40%% of living players")

	// 1. Generate system alert message
	alertEvent := core.Event{
		ID:        fmt.Sprintf("liaison_alert_%d", time.Now().UnixNano()),
		Type:      core.EventSystemMessage,
		GameID:    lpm.gameState.ID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message":     "[LIAISON ALERT] Corporate security protocol activated. Enhanced intelligence gathering authorized.",
			"message_type": "liaison_protocol",
			"priority":    "high",
		},
	}
	events = append(events, alertEvent)

	// 2. Reveal a random non-AI night action from previous night
	revealEvent := lpm.generateActionRevealEvent()
	if revealEvent != nil {
		events = append(events, *revealEvent)
	}

	// 3. Mark protocol as active for mining bonus (handled by mining manager)
	lpm.markProtocolActive()

	return events
}

// generateActionRevealEvent randomly selects and reveals a non-AI night action
func (lpm *LiaisonProtocolManager) generateActionRevealEvent() *core.Event {
	if lpm.gameState.NightActions == nil || len(lpm.gameState.NightActions) == 0 {
		log.Printf("[LiaisonProtocol] No night actions to reveal")
		return nil
	}

	// Collect non-AI night actions
	var eligibleActions []string
	for playerID, action := range lpm.gameState.NightActions {
		if player, exists := lpm.gameState.Players[playerID]; exists {
			if player.Alignment == "human" && action != nil {
				eligibleActions = append(eligibleActions, playerID)
			}
		}
	}

	if len(eligibleActions) == 0 {
		log.Printf("[LiaisonProtocol] No human night actions to reveal")
		return nil
	}

	// Randomly select an action to reveal
	selectedPlayerID := eligibleActions[lpm.rng.Intn(len(eligibleActions))]
	selectedAction := lpm.gameState.NightActions[selectedPlayerID]
	selectedPlayer := lpm.gameState.Players[selectedPlayerID]

	actionDescription := lpm.getActionDescription(selectedAction)

	revealEvent := &core.Event{
		ID:        fmt.Sprintf("liaison_action_reveal_%d", time.Now().UnixNano()),
		Type:      core.EventSystemMessage,
		GameID:    lpm.gameState.ID,
		PlayerID:  "", // Public event
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": fmt.Sprintf("[LIAISON INTEL] Corporate security detected: %s performed %s last night.",
				selectedPlayer.Name, actionDescription),
			"message_type":   "liaison_reveal",
			"revealed_player": selectedPlayerID,
			"revealed_action": selectedAction.Type,
			"priority":       "high",
		},
	}

	log.Printf("[LiaisonProtocol] Revealed action: %s by %s", actionDescription, selectedPlayer.Name)
	return revealEvent
}

// getActionDescription returns a human-readable description of a night action
func (lpm *LiaisonProtocolManager) getActionDescription(action *core.SubmittedNightAction) string {
	if action == nil {
		return "an unknown action"
	}

	switch action.Type {
	case "MINE_TOKENS":
		return "token mining operations"
	case "RUN_AUDIT":
		return "a security audit"
	case "OVERCLOCK_SERVERS":
		return "server overclocking"
	case "ISOLATE_NODE":
		return "network isolation procedures"
	case "PERFORMANCE_REVIEW":
		return "a performance review"
	case "REALLOCATE_BUDGET":
		return "budget reallocation"
	case "PIVOT":
		return "strategic pivoting"
	case "DEPLOY_HOTFIX":
		return "hotfix deployment"
	default:
		if action.TargetID != "" {
			return fmt.Sprintf("targeted action on %s", action.TargetID)
		}
		return "administrative actions"
	}
}

// markProtocolActive sets a flag that the mining manager can check for bonus slots
func (lpm *LiaisonProtocolManager) markProtocolActive() {
	// Set a temporary flag in game state for this night
	if lpm.gameState.Settings.CustomSettings == nil {
		lpm.gameState.Settings.CustomSettings = make(map[string]interface{})
	}
	lpm.gameState.Settings.CustomSettings["liaison_protocol_active"] = true
	log.Printf("[LiaisonProtocol] Mining bonus activated for current night")
}

// IsProtocolActive checks if the LIAISON Protocol bonus is currently active
func (lpm *LiaisonProtocolManager) IsProtocolActive() bool {
	if lpm.gameState.Settings.CustomSettings == nil {
		return false
	}
	
	if active, exists := lpm.gameState.Settings.CustomSettings["liaison_protocol_active"]; exists {
		if isActive, ok := active.(bool); ok {
			return isActive
		}
	}
	
	return false
}

// ClearProtocolFlag clears the LIAISON Protocol flag (called at end of night)
func (lpm *LiaisonProtocolManager) ClearProtocolFlag() {
	if lpm.gameState.Settings.CustomSettings != nil {
		delete(lpm.gameState.Settings.CustomSettings, "liaison_protocol_active")
		log.Printf("[LiaisonProtocol] Mining bonus cleared")
	}
}

// GetMiningBonusSlots returns the number of bonus mining slots from LIAISON Protocol
func (lpm *LiaisonProtocolManager) GetMiningBonusSlots() int {
	if lpm.IsProtocolActive() {
		return 2 // +2 slots as specified in the epic
	}
	return 0
}