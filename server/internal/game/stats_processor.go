package game

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/xjhc/alignment/core"
)

// StatsProcessor analyzes game data and generates comprehensive post-game statistics
type StatsProcessor struct{}

// NewStatsProcessor creates a new stats processor
func NewStatsProcessor() *StatsProcessor {
	return &StatsProcessor{}
}

// ProcessGameAnalysis analyzes a completed game and generates comprehensive analysis data
func (sp *StatsProcessor) ProcessGameAnalysis(finalState *core.GameState, events []core.Event) *core.GameAnalysis {
	log.Printf("[StatsProcessor] Processing analysis for game %s with %d events", finalState.ID, len(events))

	analysis := &core.GameAnalysis{
		GameID:    finalState.ID,
		CreatedAt: time.Now(),
	}

	// Analyze different aspects of the game
	analysis.MVP = sp.calculateMVP(finalState, events)
	analysis.KeyMoment = sp.identifyKeyMoment(finalState, events)
	analysis.Timeline = sp.generateTimeline(finalState, events)
	analysis.PlayerStats = sp.calculatePlayerStats(finalState, events)
	analysis.CommunicationHighlights = sp.analyzeCommunication(finalState, events)
	analysis.PartingShots = sp.extractPartingShots(finalState, events)

	return analysis
}

// calculateMVP determines the Most Valuable Personnel based on game contributions
func (sp *StatsProcessor) calculateMVP(finalState *core.GameState, events []core.Event) core.MVPResult {
	playerScores := make(map[string]int)
	playerInfo := make(map[string]*core.Player)

	// Copy player information
	for id, player := range finalState.Players {
		playerInfo[id] = player
		playerScores[id] = 0
	}

	// Score players based on their contributions
	for _, event := range events {
		switch event.Type {
		case core.EventMiningSuccessful:
			// +2 points for mining tokens (helps team)
			if player := playerInfo[event.PlayerID]; player != nil {
				playerScores[event.PlayerID] += 2
			}

		case core.EventPlayerEliminated:
			// Check if this was a correct elimination vote
			if targetPlayer := playerInfo[event.PlayerID]; targetPlayer != nil {
				// +3 points for each human who voted for an AI player
				// -1 point for each human who voted for another human
				if targetPlayer.Alignment == "AI" {
					// This was a correct elimination - reward voters
					sp.rewardCorrectVoters(finalState, event, playerScores, 3)
				} else if targetPlayer.Alignment == "HUMAN" {
					// This was an incorrect elimination - penalize voters
					sp.rewardCorrectVoters(finalState, event, playerScores, -1)
				}
			}

		case core.EventAIConversionBlocked:
			// +4 points for blocking an AI conversion attempt
			if player := playerInfo[event.PlayerID]; player != nil {
				playerScores[event.PlayerID] += 4
			}

		case core.EventPlayerInvestigated:
			// +2 points for using investigation abilities
			if player := playerInfo[event.PlayerID]; player != nil {
				playerScores[event.PlayerID] += 2
			}

		case core.EventAIConversionSuccess:
			// +3 points for successful AI conversion
			if player := playerInfo[event.PlayerID]; player != nil && player.Alignment == "AI" {
				playerScores[event.PlayerID] += 3
			}
		}
	}

	// Find the highest scoring player
	maxScore := -999
	var mvpPlayer *core.Player
	mvpPlayerID := ""

	for playerID, score := range playerScores {
		if score > maxScore {
			maxScore = score
			mvpPlayer = playerInfo[playerID]
			mvpPlayerID = playerID
		}
	}

	// Generate MVP reason description
	reason := sp.generateMVPReason(mvpPlayer, maxScore, finalState, events)

	return core.MVPResult{
		PlayerID:     mvpPlayerID,
		PlayerName:   mvpPlayer.Name,
		PlayerAvatar: "👤", // Default avatar - could be extracted from player data
		Score:        maxScore,
		Reason:       reason,
	}
}

// rewardCorrectVoters awards points to players who voted correctly
func (sp *StatsProcessor) rewardCorrectVoters(finalState *core.GameState, eliminationEvent core.Event, playerScores map[string]int, points int) {
	// Extract votes from the current vote state or event payload
	if finalState.VoteState != nil {
		for voterID, targetID := range finalState.VoteState.Votes {
			if targetID == eliminationEvent.PlayerID {
				playerScores[voterID] += points
			}
		}
	}
}

// generateMVPReason creates a human-readable explanation for why a player won MVP
func (sp *StatsProcessor) generateMVPReason(player *core.Player, score int, finalState *core.GameState, events []core.Event) string {
	if player == nil {
		return "Outstanding performance throughout the game"
	}

	reasons := []string{}
	
	// Analyze their contributions
	tokensMined := sp.countPlayerAction(player.ID, events, core.EventMiningSuccessful)
	if tokensMined > 0 {
		reasons = append(reasons, fmt.Sprintf("%d tokens mined for team", tokensMined))
	}

	correctVotes := sp.countCorrectVotes(player.ID, finalState, events)
	if correctVotes > 0 {
		reasons = append(reasons, fmt.Sprintf("%d correct elimination votes", correctVotes))
	}

	conversions := sp.countPlayerAction(player.ID, events, core.EventAIConversionSuccess)
	if conversions > 0 {
		reasons = append(reasons, fmt.Sprintf("%d successful conversions", conversions))
	}

	if len(reasons) == 0 {
		return fmt.Sprintf("Scored %d points through strategic gameplay", score)
	}

	return fmt.Sprintf("Scored %d points: %s", score, strings.Join(reasons, " + "))
}

// identifyKeyMoment finds the most impactful moment in the game
func (sp *StatsProcessor) identifyKeyMoment(finalState *core.GameState, events []core.Event) core.KeyMoment {
	// Look for the most impactful events in reverse chronological order
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		
		switch event.Type {
		case core.EventAIConversionBlocked:
			return core.KeyMoment{
				Title:       "🛡️ Conversion Blocked",
				Description: sp.describeConversionBlock(event, finalState),
				EventType:   string(event.Type),
				DayNumber:   finalState.DayNumber,
				Timestamp:   event.Timestamp,
			}

		case core.EventPlayerInvestigated:
			if result, ok := event.Payload["result"].(string); ok && result == "AI" {
				return core.KeyMoment{
					Title:       "🔍 AI Discovery",
					Description: sp.describeInvestigation(event, finalState),
					EventType:   string(event.Type),
					DayNumber:   finalState.DayNumber,
					Timestamp:   event.Timestamp,
				}
			}

		case core.EventPlayerEliminated:
			if player := finalState.Players[event.PlayerID]; player != nil && player.Alignment == "AI" {
				return core.KeyMoment{
					Title:       "🎯 AI Elimination",
					Description: sp.describeAIElimination(event, finalState),
					EventType:   string(event.Type),
					DayNumber:   finalState.DayNumber,
					Timestamp:   event.Timestamp,
				}
			}
		}
	}

	// Fallback to first significant event
	return core.KeyMoment{
		Title:       "🎮 Game Started",
		Description: "The corporate investigation began as employees searched for the rogue AI among them.",
		EventType:   "GAME_STARTED",
		DayNumber:   1,
		Timestamp:   finalState.CreatedAt,
	}
}

// generateTimeline creates a chronological timeline of major events
func (sp *StatsProcessor) generateTimeline(finalState *core.GameState, events []core.Event) []core.TimelineEvent {
	timeline := []core.TimelineEvent{}

	for _, event := range events {
		switch event.Type {
		case core.EventPlayerEliminated:
			timeline = append(timeline, core.TimelineEvent{
				Type:        "elimination",
				Icon:        "💀",
				Day:         sp.formatDayNumber(event, finalState),
				Description: sp.describeElimination(event, finalState),
				IconClass:   "elimination",
				Timestamp:   event.Timestamp,
			})

		case core.EventAIConversionSuccess:
			timeline = append(timeline, core.TimelineEvent{
				Type:        "conversion",
				Icon:        "🔄",
				Day:         sp.formatNightNumber(event, finalState),
				Description: sp.describeConversion(event, finalState),
				IconClass:   "conversion",
				Timestamp:   event.Timestamp,
			})

		case core.EventPlayerInvestigated, core.EventAIConversionBlocked:
			timeline = append(timeline, core.TimelineEvent{
				Type:        "ability",
				Icon:        "🛡️",
				Day:         sp.formatNightNumber(event, finalState),
				Description: sp.describeAbility(event, finalState),
				IconClass:   "ability",
				Timestamp:   event.Timestamp,
			})

		case core.EventCrisisTriggered:
			timeline = append(timeline, core.TimelineEvent{
				Type:        "crisis",
				Icon:        "⚠️",
				Day:         sp.formatDayNumber(event, finalState),
				Description: sp.describeCrisis(event, finalState),
				IconClass:   "crisis",
				Timestamp:   event.Timestamp,
			})
		}
	}

	// Sort timeline by timestamp
	sort.Slice(timeline, func(i, j int) bool {
		return timeline[i].Timestamp.Before(timeline[j].Timestamp)
	})

	return timeline
}

// calculatePlayerStats generates detailed statistics for each player
func (sp *StatsProcessor) calculatePlayerStats(finalState *core.GameState, events []core.Event) []core.PlayerStat {
	stats := []core.PlayerStat{}

	for _, player := range finalState.Players {
		stat := core.PlayerStat{
			PlayerID:  player.ID,
			Name:      player.Name,
			Avatar:    "👤", // Default avatar
			Role:      player.JobTitle,
			Alignment: strings.ToLower(player.Alignment),
			Stats: core.PlayerPerformanceStats{
				TokensMined:       sp.countPlayerAction(player.ID, events, core.EventMiningSuccessful),
				CorrectVotes:      sp.countCorrectVotes(player.ID, finalState, events),
				Conversions:       sp.countPlayerAction(player.ID, events, core.EventAIConversionSuccess),
				Nominations:       sp.countNominations(player.ID, events),
				DaysSurvived:      sp.calculateDaysSurvived(player.ID, finalState, events),
				AbilitiesUsed:     sp.countAbilityUses(player.ID, events),
				MessagesPosted:    sp.countMessages(player.ID, events),
				ReactionsReceived: sp.countReactionsReceived(player.ID, events),
			},
		}
		stats = append(stats, stat)
	}

	return stats
}

// analyzeCommunication examines chat messages and reactions
func (sp *StatsProcessor) analyzeCommunication(finalState *core.GameState, events []core.Event) core.CommunicationHighlights {
	messageEvents := []core.Event{}
	reactionEvents := []core.Event{}

	// Collect communication events
	for _, event := range events {
		if event.Type == core.EventChatMessage {
			messageEvents = append(messageEvents, event)
		} else if event.Type == core.EventMessageReaction {
			reactionEvents = append(reactionEvents, event)
		}
	}

	return core.CommunicationHighlights{
		MostReacted:   sp.findMostReactedMessage(messageEvents, reactionEvents, finalState),
		NotableQuotes: sp.findNotableQuotes(messageEvents, finalState),
		Stats: core.CommunicationStats{
			TotalMessages:            len(messageEvents),
			EmojiReactions:           len(reactionEvents),
			DirectAccusations:        sp.countAccusations(messageEvents),
			CorrectAIIdentifications: sp.countCorrectAccusations(messageEvents, finalState),
		},
	}
}

// extractPartingShots collects final messages from eliminated players
func (sp *StatsProcessor) extractPartingShots(finalState *core.GameState, events []core.Event) []core.PartingShot {
	partingShots := []core.PartingShot{}

	for _, event := range events {
		if event.Type == core.EventPartingShotSet {
			if player := finalState.Players[event.PlayerID]; player != nil {
				if message, ok := event.Payload["parting_shot"].(string); ok {
					partingShots = append(partingShots, core.PartingShot{
						PlayerID:     player.ID,
						PlayerName:   player.Name,
						PlayerAvatar: "👤", // Default avatar
						Message:      message,
						Timestamp:    event.Timestamp,
					})
				}
			}
		}
	}

	return partingShots
}

// Helper functions for detailed analysis

func (sp *StatsProcessor) countPlayerAction(playerID string, events []core.Event, eventType core.EventType) int {
	count := 0
	for _, event := range events {
		if event.PlayerID == playerID && event.Type == eventType {
			count++
		}
	}
	return count
}

func (sp *StatsProcessor) countCorrectVotes(playerID string, finalState *core.GameState, events []core.Event) int {
	// Implementation would track votes and compare against actual alignments
	// This is a simplified version
	return 1 // Placeholder
}

func (sp *StatsProcessor) countNominations(playerID string, events []core.Event) int {
	count := 0
	for _, event := range events {
		if event.Type == core.EventPlayerNominated {
			if nominator, ok := event.Payload["nominator"].(string); ok && nominator == playerID {
				count++
			}
		}
	}
	return count
}

func (sp *StatsProcessor) calculateDaysSurvived(playerID string, finalState *core.GameState, events []core.Event) int {
	// Find elimination day or return total days if survived
	for _, event := range events {
		if event.Type == core.EventPlayerEliminated && event.PlayerID == playerID {
			if day, ok := event.Payload["day"].(float64); ok {
				return int(day)
			}
		}
	}
	return finalState.DayNumber
}

func (sp *StatsProcessor) countAbilityUses(playerID string, events []core.Event) int {
	count := 0
	abilityEvents := []core.EventType{
		core.EventRunAudit, core.EventOverclockServers, core.EventIsolateNode,
		core.EventPerformanceReview, core.EventReallocateBudget, core.EventPivot,
		core.EventDeployHotfix,
	}
	
	for _, event := range events {
		if event.PlayerID == playerID {
			for _, abilityEvent := range abilityEvents {
				if event.Type == abilityEvent {
					count++
					break
				}
			}
		}
	}
	return count
}

func (sp *StatsProcessor) countMessages(playerID string, events []core.Event) int {
	return sp.countPlayerAction(playerID, events, core.EventChatMessage)
}

func (sp *StatsProcessor) countReactionsReceived(playerID string, events []core.Event) int {
	count := 0
	for _, event := range events {
		if event.Type == core.EventMessageReaction {
			if targetPlayer, ok := event.Payload["target_player"].(string); ok && targetPlayer == playerID {
				count++
			}
		}
	}
	return count
}

func (sp *StatsProcessor) findMostReactedMessage(messageEvents, reactionEvents []core.Event, finalState *core.GameState) core.MostReactedMessage {
	// Find the message with the most reactions
	reactionCounts := make(map[string]int)
	
	for _, reaction := range reactionEvents {
		if messageID, ok := reaction.Payload["message_id"].(string); ok {
			reactionCounts[messageID]++
		}
	}
	
	// Find the message with highest count
	maxReactions := 0
	var mostReactedEvent core.Event
	
	for _, message := range messageEvents {
		if count := reactionCounts[message.ID]; count > maxReactions {
			maxReactions = count
			mostReactedEvent = message
		}
	}
	
	if mostReactedEvent.ID == "" {
		// Return empty result if no messages found
		return core.MostReactedMessage{}
	}
	
	player := finalState.Players[mostReactedEvent.PlayerID]
	message, _ := mostReactedEvent.Payload["message"].(string)
	
	return core.MostReactedMessage{
		PlayerID:     mostReactedEvent.PlayerID,
		PlayerName:   player.Name,
		PlayerAvatar: "👤",
		Timestamp:    sp.formatTimestamp(mostReactedEvent.Timestamp),
		Message:      message,
		Reactions:    []core.EmojiReactionSummary{{Emoji: "🤔", Count: maxReactions}},
	}
}

func (sp *StatsProcessor) findNotableQuotes(messageEvents []core.Event, finalState *core.GameState) []core.NotableQuote {
	quotes := []core.NotableQuote{}
	
	// Look for messages containing AI accusations or key strategic moments
	for _, event := range messageEvents {
		if message, ok := event.Payload["message"].(string); ok {
			if sp.isNotableMessage(message) {
				if player := finalState.Players[event.PlayerID]; player != nil {
					quotes = append(quotes, core.NotableQuote{
						PlayerID:     event.PlayerID,
						PlayerName:   player.Name,
						PlayerAvatar: "👤",
						Timestamp:    sp.formatTimestamp(event.Timestamp),
						Message:      message,
					})
				}
			}
		}
		
		// Limit to top 3 quotes
		if len(quotes) >= 3 {
			break
		}
	}
	
	return quotes
}

func (sp *StatsProcessor) countAccusations(messageEvents []core.Event) int {
	count := 0
	for _, event := range messageEvents {
		if message, ok := event.Payload["message"].(string); ok {
			if sp.containsAccusation(message) {
				count++
			}
		}
	}
	return count
}

func (sp *StatsProcessor) countCorrectAccusations(messageEvents []core.Event, finalState *core.GameState) int {
	// This would need to analyze messages and cross-reference with actual AI players
	// Simplified for now
	return 1
}

// Helper functions for event descriptions

func (sp *StatsProcessor) describeConversionBlock(event core.Event, finalState *core.GameState) string {
	return "A critical AI conversion attempt was blocked, preventing the AI faction from gaining another member."
}

func (sp *StatsProcessor) describeInvestigation(event core.Event, finalState *core.GameState) string {
	return "A secret investigation revealed crucial alignment information, changing the course of the game."
}

func (sp *StatsProcessor) describeAIElimination(event core.Event, finalState *core.GameState) string {
	player := finalState.Players[event.PlayerID]
	if player != nil {
		return fmt.Sprintf("%s was eliminated, revealed to be AI-aligned, dealing a critical blow to the AI faction.", player.Name)
	}
	return "An AI player was successfully identified and eliminated."
}

func (sp *StatsProcessor) describeElimination(event core.Event, finalState *core.GameState) string {
	player := finalState.Players[event.PlayerID]
	if player != nil {
		return fmt.Sprintf("%s was eliminated. Revealed alignment: %s.", player.Name, strings.ToUpper(player.Alignment))
	}
	return "A player was eliminated."
}

func (sp *StatsProcessor) describeConversion(event core.Event, finalState *core.GameState) string {
	return "An AI conversion attempt succeeded, bringing another player to the AI faction."
}

func (sp *StatsProcessor) describeAbility(event core.Event, finalState *core.GameState) string {
	return "A role ability was used, potentially affecting the game's outcome."
}

func (sp *StatsProcessor) describeCrisis(event core.Event, finalState *core.GameState) string {
	if title, ok := event.Payload["title"].(string); ok {
		return fmt.Sprintf("Crisis event triggered: %s", title)
	}
	return "A crisis event was triggered, changing the game rules."
}

// Utility functions

func (sp *StatsProcessor) formatDayNumber(event core.Event, finalState *core.GameState) string {
	return fmt.Sprintf("Day %d", finalState.DayNumber)
}

func (sp *StatsProcessor) formatNightNumber(event core.Event, finalState *core.GameState) string {
	return fmt.Sprintf("Night %d", finalState.DayNumber)
}

func (sp *StatsProcessor) formatTimestamp(timestamp time.Time) string {
	return timestamp.Format("Day 1, 03:04 PM")
}

func (sp *StatsProcessor) isNotableMessage(message string) bool {
	lowerMsg := strings.ToLower(message)
	keywords := []string{"ai", "artificial", "suspect", "eliminate", "vote", "guilty", "innocent"}
	
	for _, keyword := range keywords {
		if strings.Contains(lowerMsg, keyword) {
			return true
		}
	}
	return false
}

func (sp *StatsProcessor) containsAccusation(message string) bool {
	lowerMsg := strings.ToLower(message)
	accusations := []string{"is the ai", "suspect", "vote", "eliminate", "guilty"}
	
	for _, accusation := range accusations {
		if strings.Contains(lowerMsg, accusation) {
			return true
		}
	}
	return false
}