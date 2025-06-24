package ai

import (
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
)

// AIManager handles AI player decisions and actions
type AIManager struct {
	gameState   *core.GameState
	rulesEngine *RulesEngine
}

// NewAIManager creates a new AI manager
func NewAIManager(gameState *core.GameState) *AIManager {
	return &AIManager{
		gameState:   gameState,
		rulesEngine: NewRulesEngine(),
	}
}

// GetAIPlayers returns all AI-controlled players in the game
func (aim *AIManager) GetAIPlayers() []*core.Player {
	var aiPlayers []*core.Player
	for _, player := range aim.gameState.Players {
		if player.ControlType == "AI" && player.IsAlive {
			aiPlayers = append(aiPlayers, player)
		}
	}
	return aiPlayers
}

// ProcessAIActions determines what actions AI players should take based on the current phase
func (aim *AIManager) ProcessAIActions() []core.Action {
	var actions []core.Action
	aiPlayers := aim.GetAIPlayers()
	
	if len(aiPlayers) == 0 {
		return actions
	}

	currentPhase := aim.gameState.Phase.Type

	for _, aiPlayer := range aiPlayers {
		switch currentPhase {
		case core.PhaseNight:
			action := aim.generateNightAction(aiPlayer)
			if action != nil {
				actions = append(actions, *action)
			}
		case core.PhaseNomination:
			action := aim.generateVoteAction(aiPlayer, "NOMINATION")
			if action != nil {
				actions = append(actions, *action)
			}
		case core.PhaseVerdict:
			action := aim.generateVoteAction(aiPlayer, "VERDICT")
			if action != nil {
				actions = append(actions, *action)
			}
		case core.PhaseDiscussion:
			action := aim.generateChatAction(aiPlayer)
			if action != nil {
				actions = append(actions, *action)
			}
		}
	}

	return actions
}

// generateNightAction creates a night action for the AI player
func (aim *AIManager) generateNightAction(aiPlayer *core.Player) *core.Action {
	// Simple random strategy: try to convert a random human
	humanPlayers := aim.getHumanPlayers()
	if len(humanPlayers) == 0 {
		return nil
	}

	// Select a random human to convert
	targetPlayer := humanPlayers[rand.Intn(len(humanPlayers))]

	return &core.Action{
		Type:      core.ActionAttemptConversion,
		PlayerID:  aiPlayer.ID,
		GameID:    aim.gameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id": targetPlayer.ID,
		},
	}
}

// generateVoteAction creates a voting action for the AI player
func (aim *AIManager) generateVoteAction(aiPlayer *core.Player, voteType string) *core.Action {
	var targetID string

	if voteType == "NOMINATION" {
		// For nomination votes, select a random human player to nominate
		humanPlayers := aim.getHumanPlayers()
		if len(humanPlayers) > 0 {
			targetPlayer := humanPlayers[rand.Intn(len(humanPlayers))]
			targetID = targetPlayer.ID
		}
	} else if voteType == "VERDICT" {
		// For verdict votes, vote INNOCENT to try to save the nominated player
		// (if they're AI-aligned) or GUILTY (if they're human)
		if aim.gameState.NominatedPlayer != "" {
			nominatedPlayer := aim.gameState.Players[aim.gameState.NominatedPlayer]
			if nominatedPlayer != nil {
				// Vote GUILTY if the nominated player is human
				if nominatedPlayer.Alignment == "HUMAN" {
					targetID = "GUILTY"
				} else {
					targetID = "INNOCENT"
				}
			}
		}
	}

	if targetID == "" {
		return nil
	}

	return &core.Action{
		Type:      core.ActionSubmitVote,
		PlayerID:  aiPlayer.ID,
		GameID:    aim.gameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"target_id":  targetID,
			"vote_type":  voteType,
		},
	}
}

// generateChatAction creates a chat action for the AI player
func (aim *AIManager) generateChatAction(aiPlayer *core.Player) *core.Action {
	// Simple chat: just say hello once per discussion phase
	// Check if the AI has already spoken this phase by looking at recent chat messages
	recentMessages := aim.getRecentChatMessages()
	for _, msg := range recentMessages {
		if msg.PlayerID == aiPlayer.ID {
			// AI has already spoken this phase
			return nil
		}
	}

	return &core.Action{
		Type:      core.ActionSendMessage,
		PlayerID:  aiPlayer.ID,
		GameID:    aim.gameState.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"message": "Hello world! I'm analyzing the situation...",
		},
	}
}

// getHumanPlayers returns all living human players
func (aim *AIManager) getHumanPlayers() []*core.Player {
	var humanPlayers []*core.Player
	for _, player := range aim.gameState.Players {
		if player.ControlType == "HUMAN" && player.IsAlive && player.Alignment == "HUMAN" {
			humanPlayers = append(humanPlayers, player)
		}
	}
	return humanPlayers
}

// getRecentChatMessages returns chat messages from the current discussion phase
func (aim *AIManager) getRecentChatMessages() []core.ChatMessage {
	// For simplicity, return the last 10 messages
	messages := aim.gameState.ChatMessages
	if len(messages) > 10 {
		return messages[len(messages)-10:]
	}
	return messages
}