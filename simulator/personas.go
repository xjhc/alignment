package simulator

import (
	"github.com/xjhc/alignment/core"
)

// BotPersona defines the interface for AI bot behavior
type BotPersona interface {
	// GetID returns the unique identifier for this persona type
	GetID() string
	
	// GetDescription returns a human-readable description of this persona's behavior
	GetDescription() string
	
	// DecideAction determines what action this persona wants to take given the current game state
	// Returns nil if the persona doesn't want to take any action at this time
	DecideAction(gameState core.GameState, playerID string) *core.Action
	
	// ShouldNominate decides whether to nominate a player during nomination phase
	ShouldNominate(gameState core.GameState, playerID string) (bool, string)
	
	// DecideVote determines how to vote during voting phases
	DecideVote(gameState core.GameState, playerID string, nominatedPlayer string) string
	
	// DecideNightAction determines night actions to take
	DecideNightAction(gameState core.GameState, playerID string) *core.Action
	
	// DecideChatMessage determines whether to send a chat message and what to say
	DecideChatMessage(gameState core.GameState, playerID string) *string
}
