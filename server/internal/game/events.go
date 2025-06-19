package game

import (
	"time"
)

// Event represents a game event that changes state
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	GameID    string                 `json:"game_id"`
	PlayerID  string                 `json:"player_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// EventType represents different types of game events
type EventType string

const (
	// Game lifecycle events
	EventGameCreated   EventType = "GAME_CREATED"
	EventGameStarted   EventType = "GAME_STARTED"
	EventGameEnded     EventType = "GAME_ENDED"
	EventPhaseChanged  EventType = "PHASE_CHANGED"

	// Player events
	EventPlayerJoined  EventType = "PLAYER_JOINED"
	EventPlayerLeft    EventType = "PLAYER_LEFT"
	EventPlayerVoted   EventType = "PLAYER_VOTED"
	EventPlayerEliminated EventType = "PLAYER_ELIMINATED"

	// Token events
	EventTokensAwarded EventType = "TOKENS_AWARDED"
	EventTokensSpent   EventType = "TOKENS_SPENT"
	EventMiningSuccessful EventType = "MINING_SUCCESSFUL"

	// AI events
	EventAIAction      EventType = "AI_ACTION"
	EventAIConversion  EventType = "AI_CONVERSION"
	EventAIRevealed    EventType = "AI_REVEALED"

	// Communication events
	EventChatMessage   EventType = "CHAT_MESSAGE"
	EventSystemMessage EventType = "SYSTEM_MESSAGE"
)

// Action represents a player action that can generate events
type Action struct {
	Type      ActionType             `json:"type"`
	PlayerID  string                 `json:"player_id"`
	GameID    string                 `json:"game_id"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// ActionType represents different types of player actions
type ActionType string

const (
	ActionJoinGame    ActionType = "JOIN_GAME"
	ActionLeaveGame   ActionType = "LEAVE_GAME"
	ActionSubmitVote  ActionType = "SUBMIT_VOTE"
	ActionSendMessage ActionType = "SEND_MESSAGE"
	ActionUseAbility  ActionType = "USE_ABILITY"
	ActionMineTokens  ActionType = "MINE_TOKENS"
)

// ApplyEvent applies an event to the game state
func (gs *GameState) ApplyEvent(event Event) error {
	gs.UpdatedAt = event.Timestamp
	
	switch event.Type {
	case EventGameStarted:
		return gs.applyGameStarted(event)
	case EventPlayerJoined:
		return gs.applyPlayerJoined(event)
	case EventPlayerLeft:
		return gs.applyPlayerLeft(event)
	case EventPhaseChanged:
		return gs.applyPhaseChanged(event)
	case EventPlayerVoted:
		return gs.applyPlayerVoted(event)
	case EventTokensAwarded:
		return gs.applyTokensAwarded(event)
	case EventMiningSuccessful:
		return gs.applyMiningSuccessful(event)
	default:
		// Unknown event type - log but don't error
		return nil
	}
}

func (gs *GameState) applyGameStarted(event Event) error {
	gs.Phase = Phase{
		Type:      PhaseDay,
		StartTime: event.Timestamp,
		Duration:  gs.Settings.PhaseTimeout,
	}
	return nil
}

func (gs *GameState) applyPlayerJoined(event Event) error {
	playerID := event.PlayerID
	name, _ := event.Payload["name"].(string)
	
	gs.Players[playerID] = &Player{
		ID:       playerID,
		Name:     name,
		Tokens:   1, // Starting tokens
		IsActive: true,
		JoinedAt: event.Timestamp,
	}
	return nil
}

func (gs *GameState) applyPlayerLeft(event Event) error {
	if player, exists := gs.Players[event.PlayerID]; exists {
		player.IsActive = false
	}
	return nil
}

func (gs *GameState) applyPhaseChanged(event Event) error {
	newPhaseType, _ := event.Payload["phase_type"].(string)
	duration, _ := event.Payload["duration"].(float64)
	
	gs.Phase = Phase{
		Type:      PhaseType(newPhaseType),
		StartTime: event.Timestamp,
		Duration:  time.Duration(duration) * time.Second,
	}
	gs.Turn++
	return nil
}

func (gs *GameState) applyPlayerVoted(event Event) error {
	// Vote logic would be implemented here
	// For now, just update timestamp
	return nil
}

func (gs *GameState) applyTokensAwarded(event Event) error {
	playerID := event.PlayerID
	amount, _ := event.Payload["amount"].(float64)
	
	if player, exists := gs.Players[playerID]; exists {
		player.Tokens += int(amount)
	}
	return nil
}

func (gs *GameState) applyMiningSuccessful(event Event) error {
	playerID := event.PlayerID
	amount, _ := event.Payload["amount"].(float64)
	
	if player, exists := gs.Players[playerID]; exists {
		player.Tokens += int(amount)
	}
	return nil
}