package game

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// HintManager handles the Loebmate assistant system for guiding new players
type HintManager struct {
	gameState *core.GameState
}

// NewHintManager creates a new hint manager
func NewHintManager(gameState *core.GameState) *HintManager {
	return &HintManager{
		gameState: gameState,
	}
}

// CheckForHints checks if any players need hints for the current phase and returns events
func (hm *HintManager) CheckForHints(newPhase core.PhaseType) []core.Event {
	var events []core.Event
	
	// Get hint for the new phase
	hintText := hm.GetPhaseHint(newPhase)
	if hintText == "" {
		return events
	}
	
	// Check each player to see if they need this hint
	for playerID, player := range hm.gameState.Players {
		if !player.IsAlive || player.ControlType != "HUMAN" {
			continue
		}
		
		// Check if player should receive this hint
		if hm.shouldShowHint(playerID, newPhase) {
			hintEvent := core.Event{
				ID:        fmt.Sprintf("loebmate_hint_%s_%s_%d", playerID, newPhase, time.Now().UnixNano()),
				Type:      core.EventPrivateNotification,
				GameID:    hm.gameState.ID,
				PlayerID:  playerID, // Private notification to this player
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"type":    "LOEBMATE_HINT",
					"phase":   string(newPhase),
					"message": hintText,
					"sender":  "Loebmate",
				},
			}
			events = append(events, hintEvent)
			
			// Mark this hint as seen for the player
			hm.markHintAsSeen(playerID, newPhase)
		}
	}
	
	return events
}

// shouldShowHint determines if a player should receive a hint based on progressive disclosure
func (hm *HintManager) shouldShowHint(playerID string, phase core.PhaseType) bool {
	player := hm.gameState.Players[playerID]
	if player == nil {
		return false
	}
	
	// Only show hints to living human players
	if player.ControlType != "HUMAN" || !player.IsAlive {
		return false
	}
	
	// Check if player has disabled Loebmate hints
	if player.DisableLoebmateHints {
		return false
	}
	
	// Check if player has already seen this phase hint
	if player.SeenHints != nil {
		if seen, exists := player.SeenHints[string(phase)]; exists && seen {
			return false // Already seen this hint
		}
	}
	
	return true
}

// markHintAsSeen marks a phase hint as seen for a specific player
func (hm *HintManager) markHintAsSeen(playerID string, phase core.PhaseType) {
	player := hm.gameState.Players[playerID]
	if player == nil {
		return
	}
	
	// Initialize SeenHints map if it doesn't exist
	if player.SeenHints == nil {
		player.SeenHints = make(map[string]bool)
	}
	
	// Mark this phase hint as seen
	player.SeenHints[string(phase)] = true
}

// GetPhaseHint returns the appropriate hint text for each phase
func (hm *HintManager) GetPhaseHint(phase core.PhaseType) string {
	switch phase {
	case core.PhaseSitrep:
		return "📊 **SITREP Phase:** Review the daily situation report to understand what happened during the night. Pay attention to any personnel changes or security alerts."
		
	case core.PhasePulseCheck:
		return "💭 **Pulse Check:** Time to share your thoughts! Answer the question honestly to help gauge team sentiment. Your response will be shared with everyone after this phase."
		
	case core.PhaseDiscussion:
		return "💬 **Discussion Phase:** Collaborate with your team to analyze the situation. Share information, voice suspicions, and build consensus before the nomination vote."
		
	case core.PhaseNomination:
		return "🗳️ **Nomination Phase:** Vote to nominate someone for elimination. Choose carefully - your tokens add weight to your vote. The person with the most token-weighted votes will face elimination."
		
	case core.PhaseVerdict:
		return "⚖️ **Verdict Phase:** Time to decide the nominated person's fate. Vote GUILTY to eliminate them, or INNOCENT to spare them. Consider the evidence carefully."
		
	case core.PhaseNight:
		return "🌙 **Night Phase:** Use your role's special ability, mine tokens for future votes, or project milestones. You have 30 seconds to act. The war room chat is locked during this phase."
		
	default:
		return ""
	}
}