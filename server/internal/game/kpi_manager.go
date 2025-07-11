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

// AssignKPIs assigns a random KPI to each human player
func (km *KPIManager) AssignKPIs() []core.Event {
	var events []core.Event
	
	// Available KPI types
	kpiTypes := []core.KPIType{
		core.KPICapitalist,
		core.KPIGuardian,
		core.KPIInquisitor,
		core.KPISuccessionPlanner,
		core.KPIScapegoat,
	}
	
	// Shuffle KPI types to randomize assignment
	for i := len(kpiTypes) - 1; i > 0; i-- {
		j := int(getKPICurrentTime().UnixNano()) % (i + 1)
		kpiTypes[i], kpiTypes[j] = kpiTypes[j], kpiTypes[i]
	}
	
	// Assign KPIs to human players
	kpiIndex := 0
	for playerID, player := range km.gameState.Players {
		if player.Alignment == "HUMAN" && player.ControlType == "HUMAN" {
			if kpiIndex < len(kpiTypes) {
				kpiType := kpiTypes[kpiIndex]
				player.PersonalKPI = km.createKPI(kpiType)
				
				// Generate KPI assignment event
				assignmentEvent := core.Event{
					ID:        fmt.Sprintf("kpi_assigned_%s_%d", playerID, getKPICurrentTime().UnixNano()),
					Type:      core.EventKPIAssigned,
					GameID:    km.gameState.ID,
					PlayerID:  playerID,
					Timestamp: getKPICurrentTime(),
					Payload: map[string]interface{}{
						"kpi_type":    string(kpiType),
						"description": player.PersonalKPI.Description,
						"target":      player.PersonalKPI.Target,
						"reward":      player.PersonalKPI.Reward,
					},
				}
				events = append(events, assignmentEvent)
				
				log.Printf("[KPIManager] Assigned %s KPI to player %s", kpiType, playerID)
				kpiIndex++
			}
		}
	}
	
	return events
}

// createKPI creates a PersonalKPI struct for a given KPI type
func (km *KPIManager) createKPI(kpiType core.KPIType) *core.PersonalKPI {
	switch kpiType {
	case core.KPICapitalist:
		return &core.PersonalKPI{
			Type:        core.KPICapitalist,
			Description: "End the game with the most tokens",
			Progress:    0,
			Target:      1, // Binary: have most tokens or not
			IsCompleted: false,
			Reward:      "Alternate win condition: Win if you have the most tokens",
		}
	case core.KPIGuardian:
		return &core.PersonalKPI{
			Type:        core.KPIGuardian,
			Description: "Keep the CISO alive until Day 4",
			Progress:    0,
			Target:      4, // Target day
			IsCompleted: false,
			Reward:      "Alternate win condition: Win if CISO survives to Day 4",
		}
	case core.KPIInquisitor:
		return &core.PersonalKPI{
			Type:        core.KPIInquisitor,
			Description: "Vote correctly to eliminate AI players 3 times",
			Progress:    0,
			Target:      3, // Number of correct votes
			IsCompleted: false,
			Reward:      "Gain 2 extra tokens for each correct vote",
		}
	case core.KPISuccessionPlanner:
		return &core.PersonalKPI{
			Type:        core.KPISuccessionPlanner,
			Description: "End the game with exactly 2 humans alive",
			Progress:    0,
			Target:      2, // Exact number of humans
			IsCompleted: false,
			Reward:      "Alternate win condition: Win if exactly 2 humans remain",
		}
	case core.KPIScapegoat:
		return &core.PersonalKPI{
			Type:        core.KPIScapegoat,
			Description: "Get eliminated by unanimous vote",
			Progress:    0,
			Target:      1, // Binary: eliminated unanimously or not
			IsCompleted: false,
			Reward:      "Alternate win condition: Win if eliminated unanimously",
		}
	default:
		return &core.PersonalKPI{
			Type:        kpiType,
			Description: "Unknown KPI",
			Progress:    0,
			Target:      1,
			IsCompleted: false,
			Reward:      "KPI completion bonus",
		}
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
			voteWasCorrect := eliminatedPlayer.Alignment == "ALIGNED"

			// Check all voters and update Inquisitor KPI progress
			for voterID := range km.gameState.VoteState.Votes {
				voter := km.gameState.Players[voterID]
				if voter != nil && voter.PersonalKPI != nil && voter.PersonalKPI.Type == core.KPIInquisitor {
					if voteWasCorrect && voter.Alignment == "HUMAN" {
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

						// Create private notification for KPI progress
						notificationEvent := core.Event{
							ID:        fmt.Sprintf("private_notification_%s_%d", voterID, getKPICurrentTime().UnixNano()),
							Type:      core.EventPrivateNotification,
							GameID:    km.gameState.ID,
							PlayerID:  voterID,
							Timestamp: getKPICurrentTime(),
							Payload: map[string]interface{}{
								"type":     "kpi_progress",
								"title":    "KPI Progress",
								"message":  fmt.Sprintf("Inquisitor progress: %d/%d correct votes", newProgress, voter.PersonalKPI.Target),
								"priority": "medium",
							},
						}
						events = append(events, notificationEvent)

						// Check if KPI is completed
						if newProgress >= voter.PersonalKPI.Target {
							completedEvent := km.generateKPICompletedEvent(voterID, core.KPIInquisitor)
							events = append(events, completedEvent)

							// Create private notification for KPI completion
							completionNotificationEvent := core.Event{
								ID:        fmt.Sprintf("private_notification_complete_%s_%d", voterID, getKPICurrentTime().UnixNano()),
								Type:      core.EventPrivateNotification,
								GameID:    km.gameState.ID,
								PlayerID:  voterID,
								Timestamp: getKPICurrentTime(),
								Payload: map[string]interface{}{
									"type":     "kpi_progress",
									"title":    "KPI Completed!",
									"message":  "You've completed the Inquisitor KPI! Your final vote weight will be doubled.",
									"priority": "high",
								},
							}
							events = append(events, completionNotificationEvent)

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

				// Create private notification for Guardian KPI completion
				completionNotificationEvent := core.Event{
					ID:        fmt.Sprintf("private_notification_complete_%s_%d", playerID, getKPICurrentTime().UnixNano()),
					Type:      core.EventPrivateNotification,
					GameID:    km.gameState.ID,
					PlayerID:  playerID,
					Timestamp: getKPICurrentTime(),
					Payload: map[string]interface{}{
						"type":     "kpi_progress",
						"title":    "KPI Completed!",
						"message":  fmt.Sprintf("You've completed the Guardian KPI! The CISO survived to Day %d. You now have an alternate win condition.", km.gameState.DayNumber),
						"priority": "high",
					},
				}
				events = append(events, completionNotificationEvent)
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

				// Create private notification for Guardian KPI progress
				notificationEvent := core.Event{
					ID:        fmt.Sprintf("private_notification_%s_%d", playerID, getKPICurrentTime().UnixNano()),
					Type:      core.EventPrivateNotification,
					GameID:    km.gameState.ID,
					PlayerID:  playerID,
					Timestamp: getKPICurrentTime(),
					Payload: map[string]interface{}{
						"type":     "kpi_progress",
						"title":    "Guardian KPI Progress",
						"message":  fmt.Sprintf("The CISO survived Day %d! Progress: %d/%d days", km.gameState.DayNumber, km.gameState.DayNumber, player.PersonalKPI.Target),
						"priority": "medium",
					},
				}
				events = append(events, notificationEvent)
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
			if player.Alignment == "HUMAN" {
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

					// Create private notification for Succession Planner KPI completion
					completionNotificationEvent := core.Event{
						ID:        fmt.Sprintf("private_notification_complete_%s_%d", playerID, getKPICurrentTime().UnixNano()),
						Type:      core.EventPrivateNotification,
						GameID:    km.gameState.ID,
						PlayerID:  playerID,
						Timestamp: getKPICurrentTime(),
						Payload: map[string]interface{}{
							"type":     "kpi_progress",
							"title":    "Victory!",
							"message":  "You've completed the Succession Planner KPI! Exactly 2 humans remain. You have won the game!",
							"priority": "high",
						},
					}
					events = append(events, completionNotificationEvent)
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

					// Create private notification for Capitalist KPI completion
					completionNotificationEvent := core.Event{
						ID:        fmt.Sprintf("private_notification_complete_%s_%d", playerID, getKPICurrentTime().UnixNano()),
						Type:      core.EventPrivateNotification,
						GameID:    km.gameState.ID,
						PlayerID:  playerID,
						Timestamp: getKPICurrentTime(),
						Payload: map[string]interface{}{
							"type":     "kpi_progress",
							"title":    "Victory!",
							"message":  fmt.Sprintf("You've completed the Capitalist KPI! You have the most tokens (%d). You have won the game!", maxTokens),
							"priority": "high",
						},
					}
					events = append(events, completionNotificationEvent)
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

			// Create private notification for Scapegoat KPI completion
			completionNotificationEvent := core.Event{
				ID:        fmt.Sprintf("private_notification_complete_%s_%d", eliminatedPlayerID, getKPICurrentTime().UnixNano()),
				Type:      core.EventPrivateNotification,
				GameID:    km.gameState.ID,
				PlayerID:  eliminatedPlayerID,
				Timestamp: getKPICurrentTime(),
				Payload: map[string]interface{}{
					"type":     "kpi_progress",
					"title":    "Victory!",
					"message":  "You've completed the Scapegoat KPI! You were eliminated unanimously. You have won the game!",
					"priority": "high",
				},
			}
			events = append(events, completionNotificationEvent)
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