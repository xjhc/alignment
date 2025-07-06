
package game

import (
	"fmt"

	"github.com/xjhc/alignment/core"
)

// EliminationManager handles player elimination and related events
type EliminationManager struct {
	gameState *core.GameState
}

// NewEliminationManager creates a new elimination manager
func NewEliminationManager(gameState *core.GameState) *EliminationManager {
	return &EliminationManager{
		gameState: gameState,
	}
}

// EliminatePlayer processes the elimination of a player and generates events
func (em *EliminationManager) EliminatePlayer(playerID string) ([]core.Event, error) {
	player, exists := em.gameState.Players[playerID]
	if !exists {
		return nil, fmt.Errorf("player %s not found", playerID)
	}

	if !player.IsAlive {
		return nil, fmt.Errorf("player %s is already eliminated", playerID)
	}

	var events []core.Event

	// Check for Scapegoat KPI achievement before elimination
	if em.gameState.VoteState != nil && core.CheckScapegoatKPI(*player, *em.gameState.VoteState) {
		kpiEvent := core.Event{
			ID:        fmt.Sprintf("kpi_completed_%s_%d", playerID, getCurrentTime().UnixNano()),
			Type:      core.EventKPICompleted,
			GameID:    em.gameState.ID,
			PlayerID:  playerID,
			Timestamp: getCurrentTime(),
			Payload: map[string]interface{}{
				"kpi_type":    string(core.KPIScapegoat),
				"achievement": "Eliminated by unanimous vote",
			},
		}
		events = append(events, kpiEvent)

		// Award tokens for KPI completion
		tokenReward := core.CalculateTokenReward(core.EventKPICompleted, *player, *em.gameState)
		if tokenReward > 0 {
			tokenEvent := core.Event{
				ID:        fmt.Sprintf("tokens_awarded_%s_%d", playerID, getCurrentTime().UnixNano()),
				Type:      core.EventTokensAwarded,
				GameID:    em.gameState.ID,
				PlayerID:  playerID,
				Timestamp: getCurrentTime(),
				Payload: map[string]interface{}{
					"amount": tokenReward,
					"reason": "KPI completion reward",
				},
			}
			events = append(events, tokenEvent)
		}
	}

	// Create elimination event
	roleType := ""
	if player.Role != nil {
		roleType = string(player.Role.Type)
	}

	eliminationEvent := core.Event{
		ID:        fmt.Sprintf("player_eliminated_%s_%d", playerID, getCurrentTime().UnixNano()),
		Type:      core.EventPlayerEliminated,
		GameID:    em.gameState.ID,
		PlayerID:  playerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"role_type":    roleType,
			"alignment":    player.Alignment,
			"parting_shot": player.PartingShot,
			"tokens":       player.Tokens,
		},
	}
	events = append(events, eliminationEvent)

	// Generate role revelation event for elimination
	roleRevealedEvent := core.Event{
		ID:        fmt.Sprintf("role_revealed_%s_%d", playerID, getCurrentTime().UnixNano()),
		Type:      core.EventPlayerRoleRevealed,
		GameID:    em.gameState.ID,
		PlayerID:  playerID,
		Timestamp: getCurrentTime(),
		Payload: map[string]interface{}{
			"player_id": playerID,
			"role_type": roleType,
			"alignment": player.Alignment,
			"reason": "elimination",
		},
	}
	events = append(events, roleRevealedEvent)

	// Update player state (this is done in ApplyEvent)

	return events, nil
}

// CheckWinCondition evaluates if either faction has won after an elimination
func (em *EliminationManager) CheckWinCondition() *core.WinCondition {
	return core.CheckWinCondition(*em.gameState)
}

// GetAlivePlayerCount returns the number of living players
func (em *EliminationManager) GetAlivePlayerCount() int {
	count := 0
	for _, player := range em.gameState.Players {
		if player.IsAlive {
			count++
		}
	}
	return count
}

// GetAlivePlayers returns all living players
func (em *EliminationManager) GetAlivePlayers() map[string]*core.Player {
	alive := make(map[string]*core.Player)
	for id, player := range em.gameState.Players {
		if player.IsAlive {
			alive[id] = player
		}
	}
	return alive
}

// IsPlayerAlive checks if a player is still alive
func (em *EliminationManager) IsPlayerAlive(playerID string) bool {
	if player, exists := em.gameState.Players[playerID]; exists {
		return player.IsAlive
	}
	return false
}