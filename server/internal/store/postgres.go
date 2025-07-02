package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgresStore handles persistent data storage using PostgreSQL
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL store
func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{db: db}
	
	// Run migrations
	if err := store.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("PostgreSQL store initialized successfully")
	return store, nil
}

// runMigrations runs database migrations
func (ps *PostgresStore) runMigrations() error {
	// For simplicity, we'll run the migration SQL directly
	// In a production system, you'd use a proper migration tool
	
	// First, create migrations tracking table
	migrationTrackingSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`
	
	if _, err := ps.db.Exec(migrationTrackingSQL); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	
	// Check which migrations have been applied
	var appliedMigrations []int
	rows, err := ps.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to check applied migrations: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations = append(appliedMigrations, version)
	}
	
	// Migration 1: Basic tables (players, etc.)
	if !contains(appliedMigrations, 1) {
		log.Println("Running migration 1: Basic tables")
		// This would normally be loaded from 001_create_players_table.sql
		// For now, we'll assume it's already been run manually or the table exists
		if _, err := ps.db.Exec("INSERT INTO schema_migrations (version) VALUES (1)"); err != nil {
			return fmt.Errorf("failed to record migration 1: %w", err)
		}
	}
	
	// Migration 2: Friends system
	if !contains(appliedMigrations, 2) {
		log.Println("Running migration 2: Friends system")
		friendsSystemSQL := `
		-- Create friend_requests table
		CREATE TABLE IF NOT EXISTS friend_requests (
			id VARCHAR(36) PRIMARY KEY,
			requester_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
			recipient_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
			status VARCHAR(20) DEFAULT 'pending',  -- 'pending', 'accepted', 'declined'
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			
			-- Prevent duplicate requests
			UNIQUE(requester_id, recipient_id)
		);

		-- Create friends table (bidirectional friendship)
		CREATE TABLE IF NOT EXISTS friends (
			id VARCHAR(36) PRIMARY KEY,
			player1_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
			player2_id VARCHAR(36) NOT NULL REFERENCES players(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			
			-- Ensure friendship is unique and bidirectional
			UNIQUE(player1_id, player2_id),
			-- Prevent self-friendship
			CHECK (player1_id != player2_id),
			-- Ensure consistent ordering (player1_id < player2_id) to avoid duplicates
			CHECK (player1_id < player2_id)
		);

		-- Create player_presence table for online status tracking
		CREATE TABLE IF NOT EXISTS player_presence (
			player_id VARCHAR(36) PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
			status VARCHAR(20) DEFAULT 'offline',  -- 'offline', 'online', 'in_lobby', 'in_game'
			lobby_id VARCHAR(36),
			game_id VARCHAR(36),
			last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);

		-- Create indexes for better performance
		CREATE INDEX IF NOT EXISTS idx_friend_requests_requester ON friend_requests(requester_id);
		CREATE INDEX IF NOT EXISTS idx_friend_requests_recipient ON friend_requests(recipient_id);
		CREATE INDEX IF NOT EXISTS idx_friend_requests_status ON friend_requests(status);

		CREATE INDEX IF NOT EXISTS idx_friends_player1 ON friends(player1_id);
		CREATE INDEX IF NOT EXISTS idx_friends_player2 ON friends(player2_id);

		CREATE INDEX IF NOT EXISTS idx_player_presence_status ON player_presence(status);
		CREATE INDEX IF NOT EXISTS idx_player_presence_last_seen ON player_presence(last_seen);

		-- Create trigger for friend_requests updated_at
		CREATE TRIGGER update_friend_requests_updated_at 
			BEFORE UPDATE ON friend_requests 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();

		-- Create trigger for player_presence updated_at
		CREATE TRIGGER update_player_presence_updated_at 
			BEFORE UPDATE ON player_presence 
			FOR EACH ROW 
			EXECUTE FUNCTION update_updated_at_column();
		`
		
		if _, err := ps.db.Exec(friendsSystemSQL); err != nil {
			return fmt.Errorf("failed to run friends system migration: %w", err)
		}
		
		if _, err := ps.db.Exec("INSERT INTO schema_migrations (version) VALUES (2)"); err != nil {
			return fmt.Errorf("failed to record migration 2: %w", err)
		}
	}
	
	// Migration 3: FTUE settings
	if !contains(appliedMigrations, 3) {
		log.Println("Running migration 3: FTUE settings")
		ftueSettingsSQL := `
		-- Add FTUE (First Time User Experience) settings to players table
		ALTER TABLE players ADD COLUMN IF NOT EXISTS disable_loebmate_hints BOOLEAN DEFAULT FALSE;
		ALTER TABLE players ADD COLUMN IF NOT EXISTS seen_hints JSONB DEFAULT '{}'::jsonb;
		`
		
		if _, err := ps.db.Exec(ftueSettingsSQL); err != nil {
			return fmt.Errorf("failed to run FTUE settings migration: %w", err)
		}
		
		if _, err := ps.db.Exec("INSERT INTO schema_migrations (version) VALUES (3)"); err != nil {
			return fmt.Errorf("failed to record migration 3: %w", err)
		}
	}
	
	log.Println("Database migrations completed")
	return nil
}

// contains checks if a slice contains a value
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// Close closes the database connection
func (ps *PostgresStore) Close() error {
	return ps.db.Close()
}

// Player operations

// CreatePlayer creates a new player profile
func (ps *PostgresStore) CreatePlayer(username, displayName, email string) (*Player, error) {
	player := &Player{
		ID:                   uuid.New().String(),
		Username:             username,
		DisplayName:          displayName,
		Email:                email,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		LastActiveAt:         time.Now(),
		EquippedAvatar:       "default",
		EquippedTitle:        "",
		TotalGamesPlayed:     0,
		TotalGamesWon:        0,
		TotalTokensMined:     0,
		KudosReceived:        0,
		UnlockedAchievements: []string{},
		UnlockedAvatars:      []string{"default"},
		UnlockedTitles:       []string{},
		BlockedPlayers:       []string{},
	}

	query := `
		INSERT INTO players (
			id, username, display_name, email, created_at, updated_at, last_active_at,
			equipped_avatar, equipped_title, total_games_played, total_games_won,
			total_tokens_mined, kudos_received, unlocked_achievements,
			unlocked_avatars, unlocked_titles, blocked_players
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	achievementsJSON, _ := json.Marshal(player.UnlockedAchievements)
	avatarsJSON, _ := json.Marshal(player.UnlockedAvatars)
	titlesJSON, _ := json.Marshal(player.UnlockedTitles)
	blockedJSON, _ := json.Marshal(player.BlockedPlayers)

	_, err := ps.db.Exec(query,
		player.ID, player.Username, player.DisplayName, player.Email,
		player.CreatedAt, player.UpdatedAt, player.LastActiveAt,
		player.EquippedAvatar, player.EquippedTitle,
		player.TotalGamesPlayed, player.TotalGamesWon,
		player.TotalTokensMined, player.KudosReceived,
		achievementsJSON, avatarsJSON, titlesJSON, blockedJSON,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create player: %w", err)
	}

	return player, nil
}

// GetPlayerByID retrieves a player by ID
func (ps *PostgresStore) GetPlayerByID(playerID string) (*Player, error) {
	query := `
		SELECT id, username, display_name, email, created_at, updated_at, last_active_at,
			   equipped_avatar, equipped_title, total_games_played, total_games_won,
			   total_tokens_mined, kudos_received, unlocked_achievements,
			   unlocked_avatars, unlocked_titles, blocked_players
		FROM players WHERE id = $1
	`

	player := &Player{}
	var achievementsJSON, avatarsJSON, titlesJSON, blockedJSON []byte

	err := ps.db.QueryRow(query, playerID).Scan(
		&player.ID, &player.Username, &player.DisplayName, &player.Email,
		&player.CreatedAt, &player.UpdatedAt, &player.LastActiveAt,
		&player.EquippedAvatar, &player.EquippedTitle,
		&player.TotalGamesPlayed, &player.TotalGamesWon,
		&player.TotalTokensMined, &player.KudosReceived,
		&achievementsJSON, &avatarsJSON, &titlesJSON, &blockedJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	// Unmarshal JSON fields
	json.Unmarshal(achievementsJSON, &player.UnlockedAchievements)
	json.Unmarshal(avatarsJSON, &player.UnlockedAvatars)
	json.Unmarshal(titlesJSON, &player.UnlockedTitles)
	json.Unmarshal(blockedJSON, &player.BlockedPlayers)

	return player, nil
}

// GetPlayerByUsername retrieves a player by username
func (ps *PostgresStore) GetPlayerByUsername(username string) (*Player, error) {
	query := `
		SELECT id, username, display_name, email, created_at, updated_at, last_active_at,
			   equipped_avatar, equipped_title, total_games_played, total_games_won,
			   total_tokens_mined, kudos_received, unlocked_achievements,
			   unlocked_avatars, unlocked_titles, blocked_players
		FROM players WHERE username = $1
	`

	player := &Player{}
	var achievementsJSON, avatarsJSON, titlesJSON, blockedJSON []byte

	err := ps.db.QueryRow(query, username).Scan(
		&player.ID, &player.Username, &player.DisplayName, &player.Email,
		&player.CreatedAt, &player.UpdatedAt, &player.LastActiveAt,
		&player.EquippedAvatar, &player.EquippedTitle,
		&player.TotalGamesPlayed, &player.TotalGamesWon,
		&player.TotalTokensMined, &player.KudosReceived,
		&achievementsJSON, &avatarsJSON, &titlesJSON, &blockedJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("failed to get player: %w", err)
	}

	// Unmarshal JSON fields
	json.Unmarshal(achievementsJSON, &player.UnlockedAchievements)
	json.Unmarshal(avatarsJSON, &player.UnlockedAvatars)
	json.Unmarshal(titlesJSON, &player.UnlockedTitles)
	json.Unmarshal(blockedJSON, &player.BlockedPlayers)

	return player, nil
}

// UpdatePlayerEquipment updates a player's equipped avatar and title
func (ps *PostgresStore) UpdatePlayerEquipment(playerID, equippedAvatar, equippedTitle string) error {
	query := `
		UPDATE players 
		SET equipped_avatar = $2, equipped_title = $3, updated_at = NOW()
		WHERE id = $1
	`

	result, err := ps.db.Exec(query, playerID, equippedAvatar, equippedTitle)
	if err != nil {
		return fmt.Errorf("failed to update player equipment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found")
	}

	return nil
}

// UpdatePlayerSettings updates a player's game settings
func (ps *PostgresStore) UpdatePlayerSettings(playerID string, disableLoebmateHints bool) error {
	query := `
		UPDATE players 
		SET disable_loebmate_hints = $2, updated_at = NOW()
		WHERE id = $1
	`

	result, err := ps.db.Exec(query, playerID, disableLoebmateHints)
	if err != nil {
		return fmt.Errorf("failed to update player settings: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found")
	}

	return nil
}

// UnlockAchievement adds an achievement to a player's unlocked list
func (ps *PostgresStore) UnlockAchievement(playerID, achievementID string) error {
	query := `
		UPDATE players 
		SET unlocked_achievements = unlocked_achievements || $2::jsonb,
			updated_at = NOW()
		WHERE id = $1 AND NOT (unlocked_achievements ? $3)
	`

	achievementJSON, _ := json.Marshal([]string{achievementID})
	
	_, err := ps.db.Exec(query, playerID, achievementJSON, achievementID)
	if err != nil {
		return fmt.Errorf("failed to unlock achievement: %w", err)
	}

	return nil
}

// UnlockAvatar adds an avatar to a player's unlocked list
func (ps *PostgresStore) UnlockAvatar(playerID, avatarID string) error {
	query := `
		UPDATE players 
		SET unlocked_avatars = unlocked_avatars || $2::jsonb,
			updated_at = NOW()
		WHERE id = $1 AND NOT (unlocked_avatars ? $3)
	`

	avatarJSON, _ := json.Marshal([]string{avatarID})
	
	_, err := ps.db.Exec(query, playerID, avatarJSON, avatarID)
	if err != nil {
		return fmt.Errorf("failed to unlock avatar: %w", err)
	}

	return nil
}

// UnlockTitle adds a title to a player's unlocked list
func (ps *PostgresStore) UnlockTitle(playerID, titleID string) error {
	query := `
		UPDATE players 
		SET unlocked_titles = unlocked_titles || $2::jsonb,
			updated_at = NOW()
		WHERE id = $1 AND NOT (unlocked_titles ? $3)
	`

	titleJSON, _ := json.Marshal([]string{titleID})
	
	_, err := ps.db.Exec(query, playerID, titleJSON, titleID)
	if err != nil {
		return fmt.Errorf("failed to unlock title: %w", err)
	}

	return nil
}

// UpdatePlayerStats updates a player's statistics after a game
func (ps *PostgresStore) UpdatePlayerStats(playerID string, gamesPlayed, gamesWon, tokensMined int) error {
	query := `
		UPDATE players 
		SET total_games_played = total_games_played + $2,
			total_games_won = total_games_won + $3,
			total_tokens_mined = total_tokens_mined + $4,
			last_active_at = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := ps.db.Exec(query, playerID, gamesPlayed, gamesWon, tokensMined)
	if err != nil {
		return fmt.Errorf("failed to update player stats: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("player not found")
	}

	return nil
}

// Game History operations

// CreateGameHistory records a completed game for a player
func (ps *PostgresStore) CreateGameHistory(gameID, playerID, role, alignment, winnerFaction string, isWinner bool, stats map[string]int) error {
	history := &GameHistory{
		ID:            uuid.New().String(),
		GameID:        gameID,
		PlayerID:      playerID,
		Role:          role,
		Alignment:     alignment,
		WinnerFaction: winnerFaction,
		IsWinner:      isWinner,
		TokensMined:   stats["tokensMined"],
		DaysSurvived:  stats["daysSurvived"],
		CorrectVotes:  stats["correctVotes"],
		Conversions:   stats["conversions"],
		AbilitiesUsed: stats["abilitiesUsed"],
		CreatedAt:     time.Now(),
	}

	query := `
		INSERT INTO game_history (
			id, game_id, player_id, role, alignment, winner_faction, is_winner,
			tokens_mined, days_survived, correct_votes, conversions, abilities_used, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := ps.db.Exec(query,
		history.ID, history.GameID, history.PlayerID, history.Role,
		history.Alignment, history.WinnerFaction, history.IsWinner,
		history.TokensMined, history.DaysSurvived, history.CorrectVotes,
		history.Conversions, history.AbilitiesUsed, history.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create game history: %w", err)
	}

	return nil
}

// GetPlayerGameHistory retrieves game history for a player
func (ps *PostgresStore) GetPlayerGameHistory(playerID string, limit int) ([]GameHistory, error) {
	query := `
		SELECT id, game_id, player_id, role, alignment, winner_faction, is_winner,
			   tokens_mined, days_survived, correct_votes, conversions, abilities_used, created_at
		FROM game_history 
		WHERE player_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2
	`

	rows, err := ps.db.Query(query, playerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get game history: %w", err)
	}
	defer rows.Close()

	var history []GameHistory
	for rows.Next() {
		var h GameHistory
		err := rows.Scan(
			&h.ID, &h.GameID, &h.PlayerID, &h.Role, &h.Alignment,
			&h.WinnerFaction, &h.IsWinner, &h.TokensMined, &h.DaysSurvived,
			&h.CorrectVotes, &h.Conversions, &h.AbilitiesUsed, &h.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan game history: %w", err)
		}
		history = append(history, h)
	}

	return history, nil
}

// Social features

// GiveKudos records kudos given from one player to another
func (ps *PostgresStore) GiveKudos(giverID, receiverID, gameID string) error {
	// First, insert the kudos record
	kudosQuery := `
		INSERT INTO kudos_given (id, giver_id, receiver_id, game_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (giver_id, receiver_id, game_id) DO NOTHING
	`

	kudosID := uuid.New().String()
	result, err := ps.db.Exec(kudosQuery, kudosID, giverID, receiverID, gameID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to give kudos: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	// If kudos was actually inserted (not a duplicate), increment the receiver's count
	if rowsAffected > 0 {
		updateQuery := `
			UPDATE players 
			SET kudos_received = kudos_received + 1, updated_at = NOW()
			WHERE id = $1
		`
		
		_, err = ps.db.Exec(updateQuery, receiverID)
		if err != nil {
			return fmt.Errorf("failed to update kudos count: %w", err)
		}
	}

	return nil
}

// BlockPlayer adds a player to another player's block list
func (ps *PostgresStore) BlockPlayer(blockerID, blockedID string) error {
	query := `
		UPDATE players 
		SET blocked_players = blocked_players || $2::jsonb,
			updated_at = NOW()
		WHERE id = $1 AND NOT (blocked_players ? $3)
	`

	blockedJSON, _ := json.Marshal([]string{blockedID})
	
	_, err := ps.db.Exec(query, blockerID, blockedJSON, blockedID)
	if err != nil {
		return fmt.Errorf("failed to block player: %w", err)
	}

	return nil
}

// GetBlockedPlayers returns the list of players blocked by the given player
func (ps *PostgresStore) GetBlockedPlayers(playerID string) ([]string, error) {
	query := `
		SELECT COALESCE(blocked_players, '[]'::jsonb) 
		FROM players 
		WHERE id = $1
	`
	
	var blockedPlayersJSON []byte
	err := ps.db.QueryRow(query, playerID).Scan(&blockedPlayersJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil // Player not found, return empty list
		}
		return nil, fmt.Errorf("failed to get blocked players: %w", err)
	}
	
	var blockedPlayers []string
	if err := json.Unmarshal(blockedPlayersJSON, &blockedPlayers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocked players: %w", err)
	}
	
	return blockedPlayers, nil
}

// CreatePlayerReport creates a new player behavior report
func (ps *PostgresStore) CreatePlayerReport(reporterID, reportedID, gameID, reason, description string) error {
	report := &PlayerReport{
		ID:          uuid.New().String(),
		ReporterID:  reporterID,
		ReportedID:  reportedID,
		GameID:      gameID,
		Reason:      reason,
		Description: description,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}

	query := `
		INSERT INTO player_reports (
			id, reporter_id, reported_id, game_id, reason, description, status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := ps.db.Exec(query,
		report.ID, report.ReporterID, report.ReportedID, report.GameID,
		report.Reason, report.Description, report.Status, report.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create player report: %w", err)
	}

	return nil
}

// CheckBlockedPlayers checks if player A has blocked player B
func (ps *PostgresStore) CheckBlockedPlayers(playerA, playerB string) (bool, error) {
	query := `
		SELECT blocked_players ? $2 as is_blocked
		FROM players 
		WHERE id = $1
	`

	var isBlocked bool
	err := ps.db.QueryRow(query, playerA, playerB).Scan(&isBlocked)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check blocked players: %w", err)
	}

	return isBlocked, nil
}

// GetAllAvatars retrieves all available avatars
func (ps *PostgresStore) GetAllAvatars() ([]Avatar, error) {
	query := `
		SELECT id, name, icon_emoji, description, rarity, unlock_type
		FROM avatars
		ORDER BY rarity, name
	`

	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get avatars: %w", err)
	}
	defer rows.Close()

	var avatars []Avatar
	for rows.Next() {
		var a Avatar
		err := rows.Scan(&a.ID, &a.Name, &a.IconEmoji, &a.Description, &a.Rarity, &a.UnlockType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan avatar: %w", err)
		}
		avatars = append(avatars, a)
	}

	return avatars, nil
}

// GetAllTitles retrieves all available titles
func (ps *PostgresStore) GetAllTitles() ([]Title, error) {
	query := `
		SELECT id, name, description, color, rarity, unlock_type
		FROM titles
		ORDER BY rarity, name
	`

	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get titles: %w", err)
	}
	defer rows.Close()

	var titles []Title
	for rows.Next() {
		var t Title
		err := rows.Scan(&t.ID, &t.Name, &t.Description, &t.Color, &t.Rarity, &t.UnlockType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan title: %w", err)
		}
		titles = append(titles, t)
	}

	return titles, nil
}

// GetAllAchievements retrieves all available achievements
func (ps *PostgresStore) GetAllAchievements() ([]Achievement, error) {
	query := `
		SELECT id, name, description, icon_url, rarity, unlock_criteria, avatar_reward, title_reward
		FROM achievements
		ORDER BY rarity, name
	`

	rows, err := ps.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}
	defer rows.Close()

	var achievements []Achievement
	for rows.Next() {
		var a Achievement
		var criteriaJSON []byte
		var avatarReward, titleReward sql.NullString

		err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.IconURL, &a.Rarity, 
			&criteriaJSON, &avatarReward, &titleReward)
		if err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}

		json.Unmarshal(criteriaJSON, &a.UnlockCriteria)
		if avatarReward.Valid {
			a.AvatarReward = avatarReward.String
		}
		if titleReward.Valid {
			a.TitleReward = titleReward.String
		}

		achievements = append(achievements, a)
	}

	return achievements, nil
}

// Friend Request operations

// SendFriendRequest creates a new friend request
func (ps *PostgresStore) SendFriendRequest(requesterID, recipientID string) error {
	// Check if players are already friends
	if areFriends, err := ps.AreFriends(requesterID, recipientID); err != nil {
		return fmt.Errorf("failed to check friendship status: %w", err)
	} else if areFriends {
		return fmt.Errorf("players are already friends")
	}

	// Check if a request already exists
	existingRequest, err := ps.GetFriendRequest(requesterID, recipientID)
	if err == nil && existingRequest != nil {
		return fmt.Errorf("friend request already exists")
	}

	request := &FriendRequest{
		ID:          uuid.New().String(),
		RequesterID: requesterID,
		RecipientID: recipientID,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO friend_requests (id, requester_id, recipient_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err = ps.db.Exec(query, request.ID, request.RequesterID, request.RecipientID, 
		request.Status, request.CreatedAt, request.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to send friend request: %w", err)
	}

	return nil
}

// GetFriendRequest retrieves a friend request between two players
func (ps *PostgresStore) GetFriendRequest(requesterID, recipientID string) (*FriendRequest, error) {
	query := `
		SELECT id, requester_id, recipient_id, status, created_at, updated_at
		FROM friend_requests
		WHERE requester_id = $1 AND recipient_id = $2 AND status = 'pending'
	`

	request := &FriendRequest{}
	err := ps.db.QueryRow(query, requesterID, recipientID).Scan(
		&request.ID, &request.RequesterID, &request.RecipientID,
		&request.Status, &request.CreatedAt, &request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get friend request: %w", err)
	}

	return request, nil
}

// GetFriendRequestByID retrieves a friend request by its ID
func (ps *PostgresStore) GetFriendRequestByID(requestID string) (*FriendRequest, error) {
	query := `
		SELECT id, requester_id, recipient_id, status, created_at, updated_at
		FROM friend_requests
		WHERE id = $1
	`

	request := &FriendRequest{}
	err := ps.db.QueryRow(query, requestID).Scan(
		&request.ID, &request.RequesterID, &request.RecipientID,
		&request.Status, &request.CreatedAt, &request.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("friend request not found")
		}
		return nil, fmt.Errorf("failed to get friend request: %w", err)
	}

	return request, nil
}

// GetIncomingFriendRequests retrieves all pending friend requests for a player
func (ps *PostgresStore) GetIncomingFriendRequests(playerID string) ([]FriendRequest, error) {
	query := `
		SELECT fr.id, fr.requester_id, fr.recipient_id, fr.status, fr.created_at, fr.updated_at
		FROM friend_requests fr
		WHERE fr.recipient_id = $1 AND fr.status = 'pending'
		ORDER BY fr.created_at DESC
	`

	rows, err := ps.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming friend requests: %w", err)
	}
	defer rows.Close()

	var requests []FriendRequest
	for rows.Next() {
		var req FriendRequest
		err := rows.Scan(&req.ID, &req.RequesterID, &req.RecipientID,
			&req.Status, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// GetOutgoingFriendRequests retrieves all pending friend requests sent by a player
func (ps *PostgresStore) GetOutgoingFriendRequests(playerID string) ([]FriendRequest, error) {
	query := `
		SELECT fr.id, fr.requester_id, fr.recipient_id, fr.status, fr.created_at, fr.updated_at
		FROM friend_requests fr
		WHERE fr.requester_id = $1 AND fr.status = 'pending'
		ORDER BY fr.created_at DESC
	`

	rows, err := ps.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outgoing friend requests: %w", err)
	}
	defer rows.Close()

	var requests []FriendRequest
	for rows.Next() {
		var req FriendRequest
		err := rows.Scan(&req.ID, &req.RequesterID, &req.RecipientID,
			&req.Status, &req.CreatedAt, &req.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// AcceptFriendRequest accepts a friend request and creates a friendship
func (ps *PostgresStore) AcceptFriendRequest(requestID string) error {
	// Start transaction
	tx, err := ps.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the friend request
	var req FriendRequest
	query := `
		SELECT requester_id, recipient_id 
		FROM friend_requests 
		WHERE id = $1 AND status = 'pending'
	`
	err = tx.QueryRow(query, requestID).Scan(&req.RequesterID, &req.RecipientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("friend request not found")
		}
		return fmt.Errorf("failed to get friend request: %w", err)
	}

	// Update request status
	updateQuery := `
		UPDATE friend_requests 
		SET status = 'accepted', updated_at = NOW()
		WHERE id = $1
	`
	_, err = tx.Exec(updateQuery, requestID)
	if err != nil {
		return fmt.Errorf("failed to update friend request: %w", err)
	}

	// Create friendship (ensure consistent ordering)
	player1ID, player2ID := req.RequesterID, req.RecipientID
	if player1ID > player2ID {
		player1ID, player2ID = player2ID, player1ID
	}

	friendshipQuery := `
		INSERT INTO friends (id, player1_id, player2_id, created_at)
		VALUES ($1, $2, $3, NOW())
	`
	_, err = tx.Exec(friendshipQuery, uuid.New().String(), player1ID, player2ID)
	if err != nil {
		return fmt.Errorf("failed to create friendship: %w", err)
	}

	// Commit transaction
	return tx.Commit()
}

// DeclineFriendRequest declines a friend request
func (ps *PostgresStore) DeclineFriendRequest(requestID string) error {
	query := `
		UPDATE friend_requests 
		SET status = 'declined', updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`

	result, err := ps.db.Exec(query, requestID)
	if err != nil {
		return fmt.Errorf("failed to decline friend request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("friend request not found")
	}

	return nil
}

// Friendship operations

// AreFriends checks if two players are friends
func (ps *PostgresStore) AreFriends(player1ID, player2ID string) (bool, error) {
	// Ensure consistent ordering
	if player1ID > player2ID {
		player1ID, player2ID = player2ID, player1ID
	}

	query := `
		SELECT COUNT(*) 
		FROM friends 
		WHERE player1_id = $1 AND player2_id = $2
	`

	var count int
	err := ps.db.QueryRow(query, player1ID, player2ID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check friendship: %w", err)
	}

	return count > 0, nil
}

// GetFriends retrieves all friends for a player with their status
func (ps *PostgresStore) GetFriends(playerID string) ([]FriendWithStatus, error) {
	query := `
		SELECT 
			CASE 
				WHEN f.player1_id = $1 THEN p.id
				ELSE p.id
			END as friend_id,
			p.username,
			p.display_name,
			COALESCE(pr.status, 'offline') as status,
			pr.lobby_id,
			pr.game_id,
			COALESCE(pr.last_seen, p.last_active_at) as last_seen
		FROM friends f
		JOIN players p ON (
			(f.player1_id = $1 AND p.id = f.player2_id) OR
			(f.player2_id = $1 AND p.id = f.player1_id)
		)
		LEFT JOIN player_presence pr ON p.id = pr.player_id
		ORDER BY 
			CASE WHEN COALESCE(pr.status, 'offline') = 'offline' THEN 1 ELSE 0 END,
			COALESCE(pr.last_seen, p.last_active_at) DESC
	`

	rows, err := ps.db.Query(query, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends: %w", err)
	}
	defer rows.Close()

	var friends []FriendWithStatus
	for rows.Next() {
		var friend FriendWithStatus
		err := rows.Scan(&friend.ID, &friend.Username, &friend.DisplayName,
			&friend.Status, &friend.LobbyID, &friend.GameID, &friend.LastSeen)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend: %w", err)
		}
		friends = append(friends, friend)
	}

	return friends, nil
}

// RemoveFriend removes a friendship between two players
func (ps *PostgresStore) RemoveFriend(player1ID, player2ID string) error {
	// Ensure consistent ordering
	if player1ID > player2ID {
		player1ID, player2ID = player2ID, player1ID
	}

	query := `
		DELETE FROM friends 
		WHERE player1_id = $1 AND player2_id = $2
	`

	result, err := ps.db.Exec(query, player1ID, player2ID)
	if err != nil {
		return fmt.Errorf("failed to remove friend: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("friendship not found")
	}

	return nil
}

// Player Presence operations

// UpdatePlayerPresence updates or creates a player's presence status
func (ps *PostgresStore) UpdatePlayerPresence(playerID, status string, lobbyID, gameID *string) error {
	query := `
		INSERT INTO player_presence (player_id, status, lobby_id, game_id, last_seen, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (player_id) 
		DO UPDATE SET 
			status = EXCLUDED.status,
			lobby_id = EXCLUDED.lobby_id,
			game_id = EXCLUDED.game_id,
			last_seen = NOW(),
			updated_at = NOW()
	`

	_, err := ps.db.Exec(query, playerID, status, lobbyID, gameID)
	if err != nil {
		return fmt.Errorf("failed to update player presence: %w", err)
	}

	return nil
}

// GetPlayerPresence retrieves a player's current presence status
func (ps *PostgresStore) GetPlayerPresence(playerID string) (*PlayerPresence, error) {
	query := `
		SELECT player_id, status, lobby_id, game_id, last_seen, updated_at
		FROM player_presence
		WHERE player_id = $1
	`

	presence := &PlayerPresence{}
	err := ps.db.QueryRow(query, playerID).Scan(
		&presence.PlayerID, &presence.Status, &presence.LobbyID,
		&presence.GameID, &presence.LastSeen, &presence.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &PlayerPresence{
				PlayerID: playerID,
				Status:   "offline",
				LastSeen: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get player presence: %w", err)
	}

	return presence, nil
}

// GetMultiplePlayerPresence retrieves presence for multiple players
func (ps *PostgresStore) GetMultiplePlayerPresence(playerIDs []string) (map[string]*PlayerPresence, error) {
	if len(playerIDs) == 0 {
		return make(map[string]*PlayerPresence), nil
	}

	// Create placeholders for IN clause
	placeholders := make([]string, len(playerIDs))
	args := make([]interface{}, len(playerIDs))
	for i, id := range playerIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT player_id, status, lobby_id, game_id, last_seen, updated_at
		FROM player_presence
		WHERE player_id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := ps.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple player presence: %w", err)
	}
	defer rows.Close()

	presenceMap := make(map[string]*PlayerPresence)
	
	// Initialize all players as offline
	for _, id := range playerIDs {
		presenceMap[id] = &PlayerPresence{
			PlayerID: id,
			Status:   "offline",
			LastSeen: time.Now(),
		}
	}

	// Update with actual presence data
	for rows.Next() {
		var presence PlayerPresence
		err := rows.Scan(&presence.PlayerID, &presence.Status, &presence.LobbyID,
			&presence.GameID, &presence.LastSeen, &presence.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan player presence: %w", err)
		}
		presenceMap[presence.PlayerID] = &presence
	}

	return presenceMap, nil
}