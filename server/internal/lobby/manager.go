package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/store"
)

// LobbyInfo represents lobby information for listing
type LobbyInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PlayerCount int       `json:"player_count"`
	MaxPlayers  int       `json:"max_players"`
	MinPlayers  int       `json:"min_players"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	CanJoin     bool      `json:"can_join"`
	IsPrivate   bool      `json:"is_private"`
}

// LobbyManager manages pre-game lobbies using PlayerActors
type LobbyManager struct {
	lobbies map[string]*Lobby
	tokens  map[string]*JoinToken
	mutex   sync.RWMutex

	// Dependencies
	sessionManager interfaces.SessionManagerInterface
	postgresStore  *store.PostgresStore
}

// Import interfaces to avoid circular dependency

// JoinToken represents a session token for a player in a lobby
type JoinToken struct {
	Token        string    `json:"token"`
	LobbyID      string    `json:"lobby_id"`
	PlayerID     string    `json:"player_id"`
	PlayerName   string    `json:"player_name"`
	PlayerAvatar string    `json:"player_avatar"`
	LobbyName    string    `json:"lobby_name"`
	IsHost       bool      `json:"is_host"`
	IsPrivate    bool      `json:"is_private"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewLobbyManager creates a new lobby manager
func NewLobbyManager(sessionManager interfaces.SessionManagerInterface, postgresStore *store.PostgresStore) *LobbyManager {
	lm := &LobbyManager{
		lobbies:        make(map[string]*Lobby),
		tokens:         make(map[string]*JoinToken),
		sessionManager: sessionManager,
		postgresStore:  postgresStore,
	}

	// Start cleanup routine for stale lobbies
	go lm.startCleanupRoutine()

	return lm
}

// startCleanupRoutine periodically removes stale lobbies
func (lm *LobbyManager) startCleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		lm.cleanupStaleLobbies()
	}
}

// cleanupStaleLobbies removes lobbies that have been stale for too long
func (lm *LobbyManager) cleanupStaleLobbies() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	now := time.Now()
	var staleLobbyIDs []string
	
	const (
		waitingForHostTimeout = 10 * time.Minute // Lobbies waiting for host
		emptyLobbyTimeout     = 30 * time.Minute // Empty lobbies 
		staleLobbyTimeout     = 15 * time.Minute // Lobbies with no activity
	)

	for lobbyID, lobby := range lm.lobbies {
		lobby.mutex.RLock()
		playerCount := len(lobby.Players)
		status := lobby.Status
		createdAt := lobby.CreatedAt
		lastActivity := lobby.LastActivity
		lobby.mutex.RUnlock()

		// CRITERIA FOR DELETION:
		shouldDelete := false
		reason := ""

		// 1. Lobby is waiting for host for too long
		if status == "WAITING_FOR_HOST" && now.Sub(createdAt) > waitingForHostTimeout {
			shouldDelete = true
			reason = "waiting for host timeout"
		}
		// 2. Lobby is empty and has existed for a while
		if playerCount == 0 && now.Sub(createdAt) > emptyLobbyTimeout {
			shouldDelete = true
			reason = "empty lobby timeout"
		}
		// 3. Lobby has seen no activity (joins/leaves) for a long time, regardless of player count
		if now.Sub(lastActivity) > staleLobbyTimeout {
			shouldDelete = true
			reason = "no recent activity"
		}

		if shouldDelete {
			staleLobbyIDs = append(staleLobbyIDs, lobbyID)
			log.Printf("[GC] Marking lobby %s for removal: %s (players: %d, created: %v ago, last activity: %v ago)", 
				lobbyID, reason, playerCount, now.Sub(createdAt), now.Sub(lastActivity))
		}
	}

	// Remove stale lobbies and their associated tokens
	for _, lobbyID := range staleLobbyIDs {
		lobby := lm.lobbies[lobbyID]
		
		// Stop all associated player actors if any linger (defensive cleanup)
		if lobby != nil {
			lobby.mutex.RLock()
			for _, actor := range lobby.PlayerActors {
				if actor != nil {
					actor.Stop()
				}
			}
			lobby.mutex.RUnlock()
		}
		
		delete(lm.lobbies, lobbyID)

		// Clean up associated tokens
		var staleTokens []string
		for tokenStr, token := range lm.tokens {
			if token.LobbyID == lobbyID {
				staleTokens = append(staleTokens, tokenStr)
			}
		}
		for _, tokenStr := range staleTokens {
			delete(lm.tokens, tokenStr)
		}

		log.Printf("[GC] Cleaned up stale lobby %s", lobbyID)
	}

	// Also clean up expired tokens
	var expiredTokens []string
	for tokenStr, token := range lm.tokens {
		if now.After(token.ExpiresAt) {
			expiredTokens = append(expiredTokens, tokenStr)
		}
	}
	for _, tokenStr := range expiredTokens {
		delete(lm.tokens, tokenStr)
	}
	
	if len(expiredTokens) > 0 {
		log.Printf("[GC] Cleaned up %d expired tokens", len(expiredTokens))
	}
}

// CreateLobbyViaHTTP creates a lobby and returns all necessary info for the host to connect
func (lm *LobbyManager) CreateLobbyViaHTTP(hostPlayerName, lobbyName, playerAvatar string, isPrivate bool) (string, string, string, error) {
	lobbyID := uuid.New().String()
	hostPlayerID := fmt.Sprintf("player_%s_%d", hostPlayerName, time.Now().UnixNano())

	// CRITICAL FIX: Do not create the lobby struct here or hold any locks.
	// The lobby will be created when the host's WebSocket connects.
	// This prevents the deadlock where CreateLobbyViaHTTP holds a write lock
	// while the WebSocket handler tries to acquire a read lock.

	// Generate the session token for the host with lobby creation info
	sessionToken, err := lm.generateSessionTokenWithLobbyInfo(lobbyID, hostPlayerID, hostPlayerName, playerAvatar, lobbyName, true, isPrivate)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	log.Printf("LobbyManager: Generated credentials for lobby %s, waiting for host %s to connect", lobbyID, hostPlayerID)

	// Return everything the client needs to connect
	return lobbyID, hostPlayerID, sessionToken, nil
}

// CreateLobby creates a new lobby with the host player actor
func (lm *LobbyManager) CreateLobby(hostActor interfaces.PlayerActorInterface, lobbyName string, isPrivate bool) (string, error) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lobbyID := uuid.New().String()
	hostPlayerID := hostActor.GetPlayerID()

	defaultSettings := core.GameSettings{
		MaxPlayers:               8,
		MinPlayers:               4,
		SitrepDuration:           time.Minute * 2,
		PulseCheckDuration:       time.Minute * 1,
		DiscussionDuration:       time.Minute * 5,
		ExtensionDuration:        time.Minute * 2,
		NominationDuration:       time.Minute * 2,
		TrialDuration:            time.Minute * 3,
		VerdictDuration:          time.Minute * 2,
		NightDuration:            time.Minute * 2,
		StartingTokens:           10,
		VotingThreshold:          0.5,
		InitialAlignedHumanCount: 0,
		PlayAsAI:                 false,
	}
	lobby := NewLobby(lobbyID, lobbyName, hostPlayerID, hostActor, isPrivate, defaultSettings)
	lm.lobbies[lobbyID] = lobby

	// Transition the host actor to lobby state
	err := hostActor.TransitionToLobby(lobbyID)
	if err != nil {
		delete(lm.lobbies, lobbyID)
		return "", fmt.Errorf("failed to transition host to lobby: %w", err)
	}

	// Generate session token for host
	sessionToken, err := lm.generateSessionTokenWithLobbyInfo(lobbyID, hostPlayerID, hostActor.GetPlayerName(), "", lobbyName, false, isPrivate)
	if err != nil {
		delete(lm.lobbies, lobbyID)
		return "", fmt.Errorf("failed to generate host session token: %w", err)
	}

	return sessionToken, nil
}

// JoinLobby by gameID, playerName, avatar - this is for HTTP API
func (lm *LobbyManager) JoinLobby(gameID, playerName, playerAvatar string) (string, string, error) {
	lm.mutex.RLock()
	lobby, exists := lm.lobbies[gameID]
	if !exists {
		lm.mutex.RUnlock()
		return "", "", fmt.Errorf("lobby not found")
	}

	// Check lobby status safely
	lobby.mutex.RLock()
	status := lobby.Status
	playerCount := len(lobby.Players)
	maxPlayers := lobby.MaxPlayers
	lobby.mutex.RUnlock()
	lm.mutex.RUnlock()

	if status != "WAITING" {
		return "", "", ErrLobbyNotAcceptingPlayers
	}

	if playerCount >= maxPlayers {
		return "", "", ErrLobbyFull
	}

	// Generate a unique player ID and session token
	playerID := fmt.Sprintf("player_%s_%d", playerName, time.Now().UnixNano())
	sessionToken, err := lm.generateSessionTokenWithLobbyInfo(gameID, playerID, playerName, playerAvatar, "", false, lobby.IsPrivate)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	return playerID, sessionToken, nil
}

// JoinLobbyWithActor adds a player actor to the lobby, creating it if needed
func (lm *LobbyManager) JoinLobbyWithActor(lobbyID string, playerActor interfaces.PlayerActorInterface) error {
	lm.mutex.Lock()
	lobby, exists := lm.lobbies[lobbyID]
	if !exists {
		// Lobby doesn't exist, check if this is a host trying to create it
		playerID := playerActor.GetPlayerID()

		// Find the token for this player to get lobby creation info
		var hostToken *JoinToken
		for _, token := range lm.tokens {
			if token.LobbyID == lobbyID && token.PlayerID == playerID {
				hostToken = token
				break
			}
		}

		if hostToken == nil {
			lm.mutex.Unlock()
			return fmt.Errorf("lobby not found and no valid token available")
		}

		// Create the lobby now that the first player (likely host) is connecting
		lobbyName := hostToken.LobbyName
		if lobbyName == "" {
			lobbyName = hostToken.PlayerName + "'s Game"
		}
		defaultSettings := core.GameSettings{
			MaxPlayers:               8,
			MinPlayers:               4,
			SitrepDuration:           time.Minute * 2,
			PulseCheckDuration:       time.Minute * 1,
			DiscussionDuration:       time.Minute * 5,
			ExtensionDuration:        time.Minute * 2,
			NominationDuration:       time.Minute * 2,
			TrialDuration:            time.Minute * 3,
			VerdictDuration:          time.Minute * 2,
			NightDuration:            time.Minute * 2,
			StartingTokens:           10,
			VotingThreshold:          0.5,
			InitialAlignedHumanCount: 0,
			PlayAsAI:                 false,
		}
		lobby = NewLobby(lobbyID, lobbyName, hostToken.PlayerID, playerActor, hostToken.IsPrivate, defaultSettings)
		lobby.Status = "WAITING" // Host is connected, so it's waiting for players
		lm.lobbies[lobbyID] = lobby
		log.Printf("[LobbyManager] Created lobby %s for player %s", lobbyID, playerID)
	}
	lm.mutex.Unlock()

	// Check for blocked players before allowing the join
	err := lm.checkForBlockedPlayers(playerActor.GetPlayerID(), lobby)
	if err != nil {
		return err
	}

	// Add player to lobby (this will handle validation and broadcasting)
	err = lobby.AddPlayer(playerActor)
	if err != nil {
		return err
	}

	// Transition the player actor to lobby state
	return playerActor.TransitionToLobby(lobbyID)
}

// LeaveLobby removes a player from the lobby
func (lm *LobbyManager) LeaveLobby(lobbyID string, playerID string) error {
	lm.mutex.RLock()
	lobby, exists := lm.lobbies[lobbyID]
	lm.mutex.RUnlock()

	if !exists {
		return fmt.Errorf("lobby not found")
	}

	lobby.RemovePlayer(playerID)
	return nil
}

// StartGame initiates the atomic transition from lobby to game
func (lm *LobbyManager) StartGame(lobbyID string, hostPlayerID string) error {
	log.Printf("[LobbyManager] Received START_GAME from host %s for lobby %s", hostPlayerID, lobbyID)
	lm.mutex.RLock()
	lobby, exists := lm.lobbies[lobbyID]
	if !exists {
		lm.mutex.RUnlock()
		return fmt.Errorf("lobby not found")
	}
	lm.mutex.RUnlock()

	// Lock the specific lobby for the duration of the check and data copy
	lobby.mutex.Lock()

	// Verify the host is starting the game
	if lobby.HostPlayerID != hostPlayerID {
		lobby.mutex.Unlock()
		return fmt.Errorf("only the host can start the game")
	}

	// BUG FIX: Check start conditions directly since we already hold the lock.
	// Calling lobby.CanStart() here would cause a deadlock because it tries to
	// acquire a RLock while this function holds a WLock.
	if len(lobby.Players) < lobby.MinPlayers || lobby.Status != "WAITING" {
		lobby.mutex.Unlock()
		return fmt.Errorf("lobby cannot start: not enough players or invalid state")
	}

	// Mark lobby as transitioning to prevent more players from joining
	lobby.Status = "STARTING"

	// Copy the players out so we can release the lobby lock
	playerActors := make(map[string]interfaces.PlayerActorInterface)
	for id, actor := range lobby.PlayerActors {
		playerActors[id] = actor
	}

	lobby.mutex.Unlock() // Release the lobby lock BEFORE the long operation

	// Create the game with atomic transition
	err := lm.sessionManager.CreateGameFromLobby(lobbyID, playerActors)
	if err != nil {
		lobby.mutex.Lock()
		lobby.Status = "WAITING" // Revert on failure
		lobby.mutex.Unlock()
		return fmt.Errorf("failed to create game: %w", err)
	}

	// Remove the lobby after successful transition
	lm.mutex.Lock()
	delete(lm.lobbies, lobbyID)
	lm.mutex.Unlock()

	return nil
}

// ValidateSession validates a session token for a player in a game
func (lm *LobbyManager) ValidateSession(gameID, playerID, sessionToken string) bool {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	token, exists := lm.tokens[sessionToken]
	if !exists {
		return false
	}

	if time.Now().After(token.ExpiresAt) {
		return false
	}

	return token.LobbyID == gameID && token.PlayerID == playerID
}

// GetPlayerInfo returns player information for token validation
func (lm *LobbyManager) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	// Look up player info from the token
	for _, token := range lm.tokens {
		if token.LobbyID == gameID && token.PlayerID == playerID {
			// Return name and avatar from the token
			return token.PlayerName, token.PlayerAvatar, nil
		}
	}

	return "", "", fmt.Errorf("player not found")
}

// GetLobbyList returns a list of available lobbies
func (lm *LobbyManager) GetLobbyList() []interface{} {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var lobbies []interface{}
	for _, lobby := range lm.lobbies {
		// Use fine-grained locking to read lobby state safely
		lobby.mutex.RLock()
		if lobby.Status == "WAITING" && !lobby.IsPrivate {
			playerActors := len(lobby.Players) // Read directly to avoid extra lock
			lobbies = append(lobbies, LobbyInfo{
				ID:          lobby.ID,
				Name:        lobby.Name,
				PlayerCount: playerActors,
				MaxPlayers:  lobby.MaxPlayers,
				MinPlayers:  lobby.MinPlayers,
				CreatedAt:   lobby.CreatedAt,
				Status:      lobby.Status,
				CanJoin:     playerActors < lobby.MaxPlayers,
				IsPrivate:   lobby.IsPrivate,
			})
		}
		lobby.mutex.RUnlock()
	}

	return lobbies
}

// GetLobby returns a lobby by ID
func (lm *LobbyManager) GetLobby(lobbyID string) (interface{}, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	lobby, exists := lm.lobbies[lobbyID]
	return lobby, exists
}

// GenerateJoinToken creates a public session token for lobby joining
func (lm *LobbyManager) GenerateJoinToken(lobbyID, playerID string) (string, error) {
	return lm.generateSessionToken(lobbyID, playerID)
}

// generateSessionToken creates a session token for a player in a lobby (legacy method)
func (lm *LobbyManager) generateSessionToken(lobbyID, playerID string) (string, error) {
	return lm.generateSessionTokenWithLobbyInfo(lobbyID, playerID, "Unknown", "", "", false, false)
}

// generateSessionTokenWithLobbyInfo creates a session token with full lobby info
func (lm *LobbyManager) generateSessionTokenWithLobbyInfo(lobbyID, playerID, playerName, playerAvatar, lobbyName string, isHost, isPrivate bool) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(bytes)

	token := &JoinToken{
		Token:        tokenStr,
		LobbyID:      lobbyID,
		PlayerID:     playerID,
		PlayerName:   playerName,
		PlayerAvatar: playerAvatar,
		LobbyName:    lobbyName,
		IsHost:       isHost,
		IsPrivate:    isPrivate,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	lm.mutex.Lock()
	lm.tokens[tokenStr] = token
	lm.mutex.Unlock()
	return tokenStr, nil
}

// checkForBlockedPlayers checks if the joining player or existing players have blocked each other
func (lm *LobbyManager) checkForBlockedPlayers(joiningPlayerID string, lobby *Lobby) error {
	if lm.postgresStore == nil {
		// If PostgreSQL is not available, allow all joins (graceful degradation)
		return nil
	}

	// Get the list of blocked players for the joining player
	joiningPlayerBlocked, err := lm.postgresStore.GetBlockedPlayers(joiningPlayerID)
	if err != nil {
		log.Printf("Warning: Failed to get blocked players for %s: %v", joiningPlayerID, err)
		// Don't block the join if we can't check - graceful degradation
		return nil
	}

	// Check each existing player in the lobby
	lobby.mutex.RLock()
	defer lobby.mutex.RUnlock()

	for existingPlayerID := range lobby.Players {
		// Check if the joining player has blocked this existing player
		for _, blockedID := range joiningPlayerBlocked {
			if blockedID == existingPlayerID {
				return fmt.Errorf("cannot join lobby: you have blocked a player in this game")
			}
		}

		// Check if the existing player has blocked the joining player
		existingPlayerBlocked, err := lm.postgresStore.GetBlockedPlayers(existingPlayerID)
		if err != nil {
			log.Printf("Warning: Failed to get blocked players for %s: %v", existingPlayerID, err)
			continue // Skip this check if we can't verify
		}

		for _, blockedID := range existingPlayerBlocked {
			if blockedID == joiningPlayerID {
				return fmt.Errorf("cannot join lobby: another player has blocked you")
			}
		}
	}

	return nil
}
