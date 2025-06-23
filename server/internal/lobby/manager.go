package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/xjhc/alignment/server/internal/interfaces"
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
}

// LobbyManager manages pre-game lobbies using PlayerActors
type LobbyManager struct {
	lobbies map[string]*Lobby
	tokens  map[string]*JoinToken
	mutex   sync.RWMutex

	// Dependencies
	sessionManager interfaces.SessionManagerInterface
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
	ExpiresAt    time.Time `json:"expires_at"`
}

// NewLobbyManager creates a new lobby manager
func NewLobbyManager(sessionManager interfaces.SessionManagerInterface) *LobbyManager {
	lm := &LobbyManager{
		lobbies:        make(map[string]*Lobby),
		tokens:         make(map[string]*JoinToken),
		sessionManager: sessionManager,
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

	for lobbyID, lobby := range lm.lobbies {
		// Remove lobbies waiting for host for more than 10 minutes
		if lobby.Status == "WAITING_FOR_HOST" && now.Sub(lobby.CreatedAt) > 10*time.Minute {
			staleLobbyIDs = append(staleLobbyIDs, lobbyID)
		}
		// Remove empty lobbies that have been around for more than 30 minutes
		if len(lobby.Players) == 0 && now.Sub(lobby.CreatedAt) > 30*time.Minute {
			staleLobbyIDs = append(staleLobbyIDs, lobbyID)
		}
	}

	// Remove stale lobbies and their associated tokens
	for _, lobbyID := range staleLobbyIDs {
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

		log.Printf("LobbyManager: Cleaned up stale lobby %s", lobbyID)
	}

	// Also clean up expired tokens
	for tokenStr, token := range lm.tokens {
		if now.After(token.ExpiresAt) {
			delete(lm.tokens, tokenStr)
		}
	}
}

// CreateLobbyViaHTTP creates a lobby and returns all necessary info for the host to connect
func (lm *LobbyManager) CreateLobbyViaHTTP(hostPlayerName, lobbyName, playerAvatar string) (string, string, string, error) {
	lobbyID := uuid.New().String()
	hostPlayerID := fmt.Sprintf("player_%s_%d", hostPlayerName, time.Now().UnixNano())

	// CRITICAL FIX: Do not create the lobby struct here or hold any locks.
	// The lobby will be created when the host's WebSocket connects.
	// This prevents the deadlock where CreateLobbyViaHTTP holds a write lock
	// while the WebSocket handler tries to acquire a read lock.

	// Generate the session token for the host with lobby creation info
	sessionToken, err := lm.generateSessionTokenWithLobbyInfo(lobbyID, hostPlayerID, hostPlayerName, playerAvatar, lobbyName, true)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	log.Printf("LobbyManager: Generated credentials for lobby %s, waiting for host %s to connect", lobbyID, hostPlayerID)

	// Return everything the client needs to connect
	return lobbyID, hostPlayerID, sessionToken, nil
}

// CreateLobby creates a new lobby with the host player actor
func (lm *LobbyManager) CreateLobby(hostActor interfaces.PlayerActorInterface, lobbyName string) (string, error) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lobbyID := uuid.New().String()
	hostPlayerID := hostActor.GetPlayerID()

	lobby := NewLobby(lobbyID, lobbyName, hostPlayerID, hostActor)
	lm.lobbies[lobbyID] = lobby

	// Transition the host actor to lobby state
	err := hostActor.TransitionToLobby(lobbyID)
	if err != nil {
		delete(lm.lobbies, lobbyID)
		return "", fmt.Errorf("failed to transition host to lobby: %w", err)
	}

	// Generate session token for host
	sessionToken, err := lm.generateSessionTokenWithLobbyInfo(lobbyID, hostPlayerID, hostActor.GetPlayerName(), "", lobbyName, false)
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
	sessionToken, err := lm.generateSessionTokenWithLobbyInfo(gameID, playerID, playerName, playerAvatar, "", false)
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
		lobby = NewLobby(lobbyID, lobbyName, hostToken.PlayerID, playerActor)
		lobby.Status = "WAITING" // Host is connected, so it's waiting for players
		lm.lobbies[lobbyID] = lobby
		log.Printf("[LobbyManager] Created lobby %s for player %s", lobbyID, playerID)
	}
	lm.mutex.Unlock()

	// Add player to lobby (this will handle validation and broadcasting)
	err := lobby.AddPlayer(playerActor)
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
func (lm *LobbyManager) StartGame(hostPlayerID, lobbyID string) error {
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
	for id, actor := range lobby.Players {
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

	// For now, we'll need to look up player info from the token
	for _, token := range lm.tokens {
		if token.LobbyID == gameID && token.PlayerID == playerID {
			// Return a default name - in a real implementation this would be stored
			return fmt.Sprintf("Player_%s", playerID[:8]), "", nil
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
		if lobby.Status == "WAITING" {
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
	return lm.generateSessionTokenWithLobbyInfo(lobbyID, playerID, "Unknown", "", "", false)
}

// generateSessionTokenWithLobbyInfo creates a session token with full lobby info
func (lm *LobbyManager) generateSessionTokenWithLobbyInfo(lobbyID, playerID, playerName, playerAvatar, lobbyName string, isHost bool) (string, error) {
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
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	lm.mutex.Lock()
	lm.tokens[tokenStr] = token
	lm.mutex.Unlock()
	return tokenStr, nil
}
