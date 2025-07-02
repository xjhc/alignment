package store

import (
	"time"
)

// Player represents a persistent player profile
type Player struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	DisplayName  string    `json:"displayName" db:"display_name"`
	Email        string    `json:"email" db:"email"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	LastActiveAt time.Time `json:"lastActiveAt" db:"last_active_at"`
	
	// Profile customization
	EquippedAvatar string `json:"equippedAvatar" db:"equipped_avatar"`
	EquippedTitle  string `json:"equippedTitle" db:"equipped_title"`
	
	// Statistics
	TotalGamesPlayed int `json:"totalGamesPlayed" db:"total_games_played"`
	TotalGamesWon    int `json:"totalGamesWon" db:"total_games_won"`
	TotalTokensMined int `json:"totalTokensMined" db:"total_tokens_mined"`
	KudosReceived    int `json:"kudosReceived" db:"kudos_received"`
	
	// Unlocked content
	UnlockedAchievements []string `json:"unlockedAchievements" db:"unlocked_achievements"`
	UnlockedAvatars      []string `json:"unlockedAvatars" db:"unlocked_avatars"`
	UnlockedTitles       []string `json:"unlockedTitles" db:"unlocked_titles"`
	
	// Social features
	BlockedPlayers []string `json:"blockedPlayers" db:"blocked_players"`
	
	// FTUE (First Time User Experience) settings
	DisableLoebmateHints bool                   `json:"disableLoebmateHints" db:"disable_loebmate_hints"`
	SeenHints            map[string]interface{} `json:"seenHints" db:"seen_hints"`
}

// GameHistory represents a completed game record
type GameHistory struct {
	ID          string    `json:"id" db:"id"`
	GameID      string    `json:"gameId" db:"game_id"`
	PlayerID    string    `json:"playerId" db:"player_id"`
	Role        string    `json:"role" db:"role"`
	Alignment   string    `json:"alignment" db:"alignment"`
	WinnerFaction string  `json:"winnerFaction" db:"winner_faction"`
	IsWinner    bool      `json:"isWinner" db:"is_winner"`
	TokensMined int       `json:"tokensMined" db:"tokens_mined"`
	DaysSurvived int      `json:"daysSurvived" db:"days_survived"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	
	// Performance metrics
	CorrectVotes  int `json:"correctVotes" db:"correct_votes"`
	Conversions   int `json:"conversions" db:"conversions"`
	AbilitiesUsed int `json:"abilitiesUsed" db:"abilities_used"`
}

// Achievement represents an unlockable achievement
type Achievement struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	IconURL     string `json:"iconUrl" db:"icon_url"`
	Rarity      string `json:"rarity" db:"rarity"` // "common", "rare", "epic", "legendary"
	
	// Unlock criteria (stored as JSON)
	UnlockCriteria map[string]interface{} `json:"unlockCriteria" db:"unlock_criteria"`
	
	// Rewards
	AvatarReward string `json:"avatarReward" db:"avatar_reward"`
	TitleReward  string `json:"titleReward" db:"title_reward"`
}

// PlayerReport represents a player behavior report
type PlayerReport struct {
	ID           string    `json:"id" db:"id"`
	ReporterID   string    `json:"reporterId" db:"reporter_id"`
	ReportedID   string    `json:"reportedId" db:"reported_id"`
	GameID       string    `json:"gameId" db:"game_id"`
	Reason       string    `json:"reason" db:"reason"`
	Description  string    `json:"description" db:"description"`
	Status       string    `json:"status" db:"status"` // "pending", "reviewed", "dismissed"
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	ReviewedAt   *time.Time `json:"reviewedAt" db:"reviewed_at"`
	ReviewedBy   *string   `json:"reviewedBy" db:"reviewed_by"`
}

// KudosGiven represents kudos given between players
type KudosGiven struct {
	ID         string    `json:"id" db:"id"`
	GiverID    string    `json:"giverId" db:"giver_id"`
	ReceiverID string    `json:"receiverId" db:"receiver_id"`
	GameID     string    `json:"gameId" db:"game_id"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

// Avatar represents an unlockable avatar
type Avatar struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	IconEmoji   string `json:"iconEmoji" db:"icon_emoji"`
	Description string `json:"description" db:"description"`
	Rarity      string `json:"rarity" db:"rarity"`
	UnlockType  string `json:"unlockType" db:"unlock_type"` // "default", "achievement", "special"
}

// Title represents an unlockable title
type Title struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Color       string `json:"color" db:"color"` // CSS color for display
	Rarity      string `json:"rarity" db:"rarity"`
	UnlockType  string `json:"unlockType" db:"unlock_type"`
}

// FriendRequest represents a pending friend request
type FriendRequest struct {
	ID          string    `json:"id" db:"id"`
	RequesterID string    `json:"requesterId" db:"requester_id"`
	RecipientID string    `json:"recipientId" db:"recipient_id"`
	Status      string    `json:"status" db:"status"` // "pending", "accepted", "declined"
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

// Friend represents a friendship between two players
type Friend struct {
	ID        string    `json:"id" db:"id"`
	Player1ID string    `json:"player1Id" db:"player1_id"`
	Player2ID string    `json:"player2Id" db:"player2_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// PlayerPresence represents a player's current online status
type PlayerPresence struct {
	PlayerID  string     `json:"playerId" db:"player_id"`
	Status    string     `json:"status" db:"status"` // "offline", "online", "in_lobby", "in_game"
	LobbyID   *string    `json:"lobbyId" db:"lobby_id"`
	GameID    *string    `json:"gameId" db:"game_id"`
	LastSeen  time.Time  `json:"lastSeen" db:"last_seen"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// FriendWithStatus includes friend info with their current presence status
type FriendWithStatus struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"displayName"`
	Status      string    `json:"status"`
	LobbyID     *string   `json:"lobbyId"`
	GameID      *string   `json:"gameId"`
	LastSeen    time.Time `json:"lastSeen"`
}