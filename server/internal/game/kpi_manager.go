package game

import (
	"fmt"
	"log"
	"time"

	"github.com/xjhc/alignment/core"
)

// KPIManager handles tracking and progression of personal KPIs
type KPIManager struct {
	gameState *core.GameState
}

// NewKPIManager creates a new KPI manager
func NewKPIManager(gameState *core.GameState) *KPIManager {
	return &KPIManager{
		gameState: gameState,
	}
}

// TrackPlayerEliminated handles KPI progress when a player is eliminated
func (km *KPIManager) TrackPlayerEliminated(eliminatedPlayerID string) []core.Event {
	var events []core.Event

	// Track Inquisitor KPI progress for players who voted correctly
	if km.gameState.VoteState != nil {
		eliminatedPlayer := km.gameState.Players[eliminatedPlayerID]
		if eliminatedPlayer != nil {
			// Determine if this was a correct vote (AI player eliminated = correct vote for humans)
			voteWasCorrect := eliminatedPlayer.Alignment == "ai"

			// Check all voters and update Inquisitor KPI progress
			for voterID := range km.gameState.VoteState.Votes {
				voter := km.gameState.Players[voterID]
				if voter != nil && voter.PersonalKPI != nil && voter.PersonalKPI.Type == core.KPIInquisitor {
					if voteWasCorrect && voter.Alignment == "human" {
						// Human voted to eliminate AI - correct vote
						newProgress := voter.PersonalKPI.Progress + 1
						log.Printf("[KPIManager] Player %s made correct vote, Inquisitor progress: %d/%d", 
							voterID, newProgress, voter.PersonalKPI.Target)

						progressEvent := core.Event{
							ID:        fmt.Sprintf("kpi_progress_%s_%d", voterID, getKPICurrentTime().UnixNano()),
							Type:      core.EventKPIProgress,
							GameID:    km.gameState.ID,
							PlayerID:  voterID,
							Timestamp: getKPICurrentTime(),
							Payload: map[string]interface{}{
								"kpi_type": string(core.KPIInquisitor),
								"progress": newProgress,
								"reason":   "correct elimination vote",
							},
						}
						events = append(events, progressEvent)

						// Check if KPI is completed
						if newProgress >= voter.PersonalKPI.Target {
							completedEvent := km.generateKPICompletedEvent(voterID, core.KPIInquisitor)
							events = append(events, completedEvent)

							// Award tokens for Inquisitor KPI completion
							tokenReward := newProgress * 2 // 2 tokens per correct vote
							tokenEvent := core.Event{
								ID:        fmt.Sprintf("tokens_awarded_%s_%d", voterID, getKPICurrentTime().UnixNano()),
								Type:      core.EventTokensAwarded,
								GameID:    km.gameState.ID,
								PlayerID:  voterID,
								Timestamp: getKPICurrentTime(),
								Payload: map[string]interface{}{
									"amount": tokenReward,
									"reason": "Inquisitor KPI completion bonus",
								},
							}
							events = append(events, tokenEvent)
						}
					}
				}
			}
		}
	}

	return events
}

// TrackNightSurvival handles Guardian KPI progress when CISO survives the night
func (km *KPIManager) TrackNightSurvival() []core.Event {
	var events []core.Event

	// Find CISO player
	var cisoPlayer *core.Player
	for _, player := range km.gameState.Players {
		if player.Role != nil && player.Role.Type == core.RoleCISO && player.IsAlive {
			cisoPlayer = player
			break
		}
	}

	if cisoPlayer == nil {
		return events // CISO is dead or doesn't exist
	}

	// Check all players with Guardian KPI
	for playerID, player := range km.gameState.Players {
		if player.PersonalKPI != nil && player.PersonalKPI.Type == core.KPIGuardian {
			// Check if we've reached the target day
			if km.gameState.DayNumber >= player.PersonalKPI.Target {
				log.Printf("[KPIManager] Player %s completed Guardian KPI - CISO survived to Day %d", 
					playerID, km.gameState.DayNumber)

				completedEvent := km.generateKPICompletedEvent(playerID, core.KPIGuardian)
				events = append(events, completedEvent)
			} else {
				// Update progress (days survived)
				progressEvent := core.Event{
					ID:        fmt.Sprintf("kpi_progress_%s_%d", playerID, getKPICurrentTime().UnixNano()),
					Type:      core.EventKPIProgress,
					GameID:    km.gameState.ID,
					PlayerID:  playerID,
					Timestamp: getKPICurrentTime(),
					Payload: map[string]interface{}{
						"kpi_type": string(core.KPIGuardian),
						"progress": km.gameState.DayNumber,
						"reason":   "CISO survived another day",
					},
				}
				events = append(events, progressEvent)
			}
		}
	}

	return events
}

// CheckGameEndKPIs evaluates KPIs that are resolved at game end
func (km *KPIManager) CheckGameEndKPIs() []core.Event {
	var events []core.Event

	// Count alive humans
	aliveHumans := 0
	var alivePlayers []*core.Player
	for _, player := range km.gameState.Players {
		if player.IsAlive {
			alivePlayers = append(alivePlayers, player)
			if player.Alignment == "human" {
				aliveHumans++
			}
		}
	}

	// Check Succession Planner KPI (exactly 2 humans alive)
	for playerID, player := range km.gameState.Players {
		if player.PersonalKPI != nil && !player.PersonalKPI.IsCompleted {
			switch player.PersonalKPI.Type {
			case core.KPISuccessionPlanner:
				if aliveHumans == 2 {
					log.Printf("[KPIManager] Player %s completed Succession Planner KPI - exactly 2 humans remain", playerID)
					completedEvent := km.generateKPICompletedEvent(playerID, core.KPISuccessionPlanner)
					events = append(events, completedEvent)
				}

			case core.KPICapitalist:
				// Find player with most tokens
				maxTokens := -1
				var richestPlayer *core.Player
				for _, p := range alivePlayers {
					if p.Tokens > maxTokens {
						maxTokens = p.Tokens
						richestPlayer = p
					}
				}

				if richestPlayer != nil && richestPlayer.ID == playerID {
					log.Printf("[KPIManager] Player %s completed Capitalist KPI - has most tokens (%d)", playerID, maxTokens)
					completedEvent := km.generateKPICompletedEvent(playerID, core.KPICapitalist)
					events = append(events, completedEvent)
				}
			}
		}
	}

	return events
}

// TrackUnanimousElimination handles Scapegoat KPI when a player is eliminated unanimously
func (km *KPIManager) TrackUnanimousElimination(eliminatedPlayerID string) []core.Event {
	var events []core.Event

	eliminatedPlayer := km.gameState.Players[eliminatedPlayerID]
	if eliminatedPlayer == nil || eliminatedPlayer.PersonalKPI == nil || 
		eliminatedPlayer.PersonalKPI.Type != core.KPIScapegoat {
		return events
	}

	// Check if the vote was unanimous
	if km.gameState.VoteState != nil {
		totalVoters := 0
		votesForEliminated := 0

		for _, targetID := range km.gameState.VoteState.Votes {
			totalVoters++
			if targetID == eliminatedPlayerID {
				votesForEliminated++
			}
		}

		if totalVoters > 0 && votesForEliminated == totalVoters {
			log.Printf("[KPIManager] Player %s completed Scapegoat KPI - eliminated unanimously", eliminatedPlayerID)
			completedEvent := km.generateKPICompletedEvent(eliminatedPlayerID, core.KPIScapegoat)
			events = append(events, completedEvent)
		}
	}

	return events
}

// generateKPICompletedEvent creates a KPI completion event
func (km *KPIManager) generateKPICompletedEvent(playerID string, kpiType core.KPIType) core.Event {
	return core.Event{
		ID:        fmt.Sprintf("kpi_completed_%s_%d", playerID, getKPICurrentTime().UnixNano()),
		Type:      core.EventKPICompleted,
		GameID:    km.gameState.ID,
		PlayerID:  playerID,
		Timestamp: getKPICurrentTime(),
		Payload: map[string]interface{}{
			"kpi_type": string(kpiType),
			"reward":   km.getKPIReward(kpiType),
		},
	}
}

// getKPIReward returns the reward description for completing a KPI
func (km *KPIManager) getKPIReward(kpiType core.KPIType) string {
	switch kpiType {
	case core.KPIInquisitor:
		return "Gain 2 extra tokens for each correct vote"
	case core.KPIGuardian:
		return "Alternate win condition: Win if CISO survives to Day 4"
	case core.KPISuccessionPlanner:
		return "Alternate win condition: Win if exactly 2 humans remain"
	case core.KPICapitalist:
		return "Alternate win condition: Win if you have the most tokens"
	case core.KPIScapegoat:
		return "Alternate win condition: Win if eliminated unanimously"
	default:
		return "KPI completion bonus"
	}
}

// Helper function to get current time (can be mocked for testing)
func getKPICurrentTime() time.Time {
	return time.Now()
}