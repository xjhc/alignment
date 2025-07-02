package core

import "time"

// GameAnalysis contains comprehensive post-game analysis data
type GameAnalysis struct {
	GameID     string     `json:"gameId"`
	CreatedAt  time.Time  `json:"createdAt"`
	MVP        MVPResult  `json:"mvp"`
	KeyMoment  KeyMoment  `json:"keyMoment"`
	Timeline   []TimelineEvent `json:"timeline"`
	PlayerStats []PlayerStat `json:"playerStats"`
	CommunicationHighlights CommunicationHighlights `json:"communicationHighlights"`
	PartingShots []PartingShot `json:"partingShots"`
}

// MVPResult represents the Most Valuable Personnel award
type MVPResult struct {
	PlayerID string `json:"playerId"`
	PlayerName string `json:"playerName"`
	PlayerAvatar string `json:"playerAvatar"`
	Score int `json:"score"`
	Reason string `json:"reason"`
}

// KeyMoment represents the most impactful moment of the game
type KeyMoment struct {
	Title string `json:"title"`
	Description string `json:"description"`
	EventType string `json:"eventType"`
	DayNumber int `json:"dayNumber"`
	Timestamp time.Time `json:"timestamp"`
}

// TimelineEvent represents a significant event in the game timeline
type TimelineEvent struct {
	Type string `json:"type"` // "elimination", "conversion", "ability", "crisis"
	Icon string `json:"icon"`
	Day string `json:"day"`
	Description string `json:"description"`
	IconClass string `json:"iconClass"`
	Timestamp time.Time `json:"timestamp"`
}

// PlayerStat represents individual player performance statistics
type PlayerStat struct {
	PlayerID string `json:"playerId"`
	Name string `json:"name"`
	Avatar string `json:"avatar"`
	Role string `json:"role"`
	Alignment string `json:"alignment"` // "human", "ai", "aligned"
	Stats PlayerPerformanceStats `json:"stats"`
}

// PlayerPerformanceStats contains detailed performance metrics
type PlayerPerformanceStats struct {
	TokensMined int `json:"tokensMined"`
	CorrectVotes int `json:"correctVotes"`
	Conversions int `json:"conversions,omitempty"` // For AI players
	Nominations int `json:"nominations"`
	DaysSurvived int `json:"daysSurvived"`
	AbilitiesUsed int `json:"abilitiesUsed"`
	MessagesPosted int `json:"messagesPosted"`
	ReactionsReceived int `json:"reactionsReceived"`
}

// PartingShot represents a player's final message before elimination
type PartingShot struct {
	PlayerID string `json:"playerId"`
	PlayerName string `json:"playerName"`
	PlayerAvatar string `json:"playerAvatar"`
	Message string `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// CommunicationHighlights contains notable chat/communication moments
type CommunicationHighlights struct {
	MostReacted MostReactedMessage `json:"mostReacted"`
	NotableQuotes []NotableQuote `json:"notableQuotes"`
	Stats CommunicationStats `json:"stats"`
}

// MostReactedMessage represents the message with the most emoji reactions
type MostReactedMessage struct {
	PlayerID string `json:"playerId"`
	PlayerName string `json:"playerName"`
	PlayerAvatar string `json:"playerAvatar"`
	Timestamp string `json:"timestamp"`
	Message string `json:"message"`
	Reactions []EmojiReactionSummary `json:"reactions"`
}

// EmojiReactionSummary represents aggregated emoji reaction data
type EmojiReactionSummary struct {
	Emoji string `json:"emoji"`
	Count int `json:"count"`
}

// NotableQuote represents a significant or impactful message
type NotableQuote struct {
	PlayerID string `json:"playerId"`
	PlayerName string `json:"playerName"`
	PlayerAvatar string `json:"playerAvatar"`
	Timestamp string `json:"timestamp"`
	Message string `json:"message"`
	Context string `json:"context,omitempty"` // Additional context about why this quote is notable
}

// CommunicationStats contains aggregate communication statistics
type CommunicationStats struct {
	TotalMessages int `json:"totalMessages"`
	EmojiReactions int `json:"emojiReactions"`
	DirectAccusations int `json:"directAccusations"`
	CorrectAIIdentifications int `json:"correctAIIdentifications"`
}