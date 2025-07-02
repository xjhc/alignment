package game

import (
	"log"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/store"
)

// AchievementChecker analyzes game completion data and awards achievements
type AchievementChecker struct {
	postgresStore *store.PostgresStore
	achievements  []store.Achievement
}

// NewAchievementChecker creates a new achievement checker
func NewAchievementChecker(postgresStore *store.PostgresStore) *AchievementChecker {
	checker := &AchievementChecker{
		postgresStore: postgresStore,
	}

	// Load all achievements from database
	if achievements, err := postgresStore.GetAllAchievements(); err == nil {
		checker.achievements = achievements
	} else {
		log.Printf("Warning: Failed to load achievements: %v", err)
		checker.achievements = []store.Achievement{}
	}

	return checker
}

// CheckAchievements checks if any players earned achievements after a game ends
func (ac *AchievementChecker) CheckAchievements(finalState *core.GameState, analysis *core.GameAnalysis) map[string][]string {
	if ac.postgresStore == nil {
		return map[string][]string{}
	}

	unlockedAchievements := make(map[string][]string) // playerID -> []achievementIDs

	for _, player := range finalState.Players {
		// Get player's persistent profile to check current achievement status
		profile, err := ac.postgresStore.GetPlayerByID(player.ID)
		if err != nil {
			log.Printf("Failed to get player profile for %s: %v", player.ID, err)
			continue
		}

		// Check each achievement
		for _, achievement := range ac.achievements {
			// Skip if player already has this achievement
			if ac.hasAchievement(profile.UnlockedAchievements, achievement.ID) {
				continue
			}

			// Check if player meets the criteria for this achievement
			if ac.checkAchievementCriteria(achievement, player, finalState, analysis, profile) {
				unlockedAchievements[player.ID] = append(unlockedAchievements[player.ID], achievement.ID)
				
				// Unlock the achievement in the database
				if err := ac.postgresStore.UnlockAchievement(player.ID, achievement.ID); err != nil {
					log.Printf("Failed to unlock achievement %s for player %s: %v", achievement.ID, player.ID, err)
				}

				// Unlock associated rewards
				if achievement.AvatarReward != "" {
					if err := ac.postgresStore.UnlockAvatar(player.ID, achievement.AvatarReward); err != nil {
						log.Printf("Failed to unlock avatar reward %s for player %s: %v", achievement.AvatarReward, player.ID, err)
					}
				}

				if achievement.TitleReward != "" {
					if err := ac.postgresStore.UnlockTitle(player.ID, achievement.TitleReward); err != nil {
						log.Printf("Failed to unlock title reward %s for player %s: %v", achievement.TitleReward, player.ID, err)
					}
				}

				log.Printf("Player %s unlocked achievement: %s", player.Name, achievement.Name)
			}
		}
	}

	return unlockedAchievements
}

// hasAchievement checks if a player already has an achievement
func (ac *AchievementChecker) hasAchievement(unlockedAchievements []string, achievementID string) bool {
	for _, id := range unlockedAchievements {
		if id == achievementID {
			return true
		}
	}
	return false
}

// checkAchievementCriteria checks if a player meets the criteria for a specific achievement
func (ac *AchievementChecker) checkAchievementCriteria(achievement store.Achievement, player *core.Player, finalState *core.GameState, analysis *core.GameAnalysis, profile *store.Player) bool {
	criteria := achievement.UnlockCriteria

	// Check different types of criteria
	switch achievement.ID {
	case "first_win":
		return ac.checkFirstWin(criteria, player, finalState, profile)
	case "win_as_ai":
		return ac.checkWinAsAI(criteria, player, finalState, analysis)
	case "survive_5_days":
		return ac.checkSurvivalDays(criteria, player, finalState)
	case "perfect_detective":
		return ac.checkPerfectVoting(criteria, player, finalState, analysis)
	case "social_butterfly":
		return ac.checkKudosReceived(criteria, profile)
	case "legend_status":
		return ac.checkTotalWins(criteria, profile)
	default:
		// Generic criteria checking
		return ac.checkGenericCriteria(criteria, player, finalState, analysis, profile)
	}
}

// checkFirstWin checks if this is the player's first win
func (ac *AchievementChecker) checkFirstWin(criteria map[string]interface{}, player *core.Player, finalState *core.GameState, profile *store.Player) bool {
	if finalState.WinCondition == nil {
		return false
	}

	// Check if the player is on the winning side
	isWinner := (finalState.WinCondition.Winner == "HUMANS" && player.Alignment == "HUMAN") ||
		(finalState.WinCondition.Winner == "AI" && (player.Alignment == "AI" || player.Alignment == "ALIGNED"))

	return isWinner && profile.TotalGamesWon == 0 // This will be their first win after the game is processed
}

// checkWinAsAI checks for AI-specific win conditions
func (ac *AchievementChecker) checkWinAsAI(criteria map[string]interface{}, player *core.Player, finalState *core.GameState, analysis *core.GameAnalysis) bool {
	if finalState.WinCondition == nil || finalState.WinCondition.Winner != "AI" {
		return false
	}

	if player.Alignment != "AI" {
		return false
	}

	// Check if the AI won without receiving any votes
	// This would need to be tracked in the analysis data
	if analysis != nil {
		for _, playerStat := range analysis.PlayerStats {
			if playerStat.PlayerID == player.ID {
				// If they received no nominations/votes against them
				return playerStat.Stats.Nominations == 0
			}
		}
	}

	return false
}

// checkSurvivalDays checks if a player survived a certain number of days
func (ac *AchievementChecker) checkSurvivalDays(criteria map[string]interface{}, player *core.Player, finalState *core.GameState) bool {
	requiredDays, ok := criteria["days_survived"].(float64)
	if !ok {
		return false
	}

	return player.IsAlive && finalState.DayNumber >= int(requiredDays)
}

// checkPerfectVoting checks if a player voted correctly in all elimination rounds in a winning game
func (ac *AchievementChecker) checkPerfectVoting(criteria map[string]interface{}, player *core.Player, finalState *core.GameState, analysis *core.GameAnalysis) bool {
	if finalState.WinCondition == nil {
		return false
	}

	// Check if the player is on the winning side
	isWinner := (finalState.WinCondition.Winner == "HUMANS" && player.Alignment == "HUMAN") ||
		(finalState.WinCondition.Winner == "AI" && (player.Alignment == "AI" || player.Alignment == "ALIGNED"))

	if !isWinner {
		return false
	}

	// Check perfect voting record from analysis
	if analysis != nil {
		for _, playerStat := range analysis.PlayerStats {
			if playerStat.PlayerID == player.ID {
				// They must have voted correctly in all votes they participated in
				// This is a simplified check - in reality you'd need to track voting accuracy more precisely
				return playerStat.Stats.CorrectVotes > 0 && playerStat.Stats.CorrectVotes >= finalState.DayNumber-1
			}
		}
	}

	return false
}

// checkKudosReceived checks if a player has received enough kudos
func (ac *AchievementChecker) checkKudosReceived(criteria map[string]interface{}, profile *store.Player) bool {
	requiredKudos, ok := criteria["kudos_received"].(float64)
	if !ok {
		return false
	}

	return profile.KudosReceived >= int(requiredKudos)
}

// checkTotalWins checks if a player has won enough games
func (ac *AchievementChecker) checkTotalWins(criteria map[string]interface{}, profile *store.Player) bool {
	requiredWins, ok := criteria["wins"].(float64)
	if !ok {
		return false
	}

	return profile.TotalGamesWon >= int(requiredWins)
}

// checkGenericCriteria handles generic achievement criteria
func (ac *AchievementChecker) checkGenericCriteria(criteria map[string]interface{}, player *core.Player, finalState *core.GameState, analysis *core.GameAnalysis, profile *store.Player) bool {
	// This is where you'd implement more complex or custom achievement logic
	// For now, we'll just check basic win conditions
	if wins, ok := criteria["wins"].(float64); ok {
		return profile.TotalGamesWon >= int(wins)
	}

	return false
}

// GetAchievementByID returns an achievement by its ID
func (ac *AchievementChecker) GetAchievementByID(achievementID string) *store.Achievement {
	for _, achievement := range ac.achievements {
		if achievement.ID == achievementID {
			return &achievement
		}
	}
	return nil
}