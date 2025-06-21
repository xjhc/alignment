package lobby

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LobbyManager manages pre-game lobbies in memory
type LobbyManager struct {
	lobbies map[string]*Lobby
	tokens  map[string]*JoinToken
	actors  map[string]*LobbyActor
	mutex   sync.RWMutex

	// Dependencies
	broadcaster Broadcaster
	supervisor  Supervisor
}

// PlayerInfo holds basic info for a player in the lobby
type PlayerInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// Supervisor manages game actors
type Supervisor interface {
	CreateGameWithPlayers(gameID string, players []PlayerInfo, hostID string) error
}

// Lobby represents a pre-game waiting room
type Lobby struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	HostPlayerID string       `json:"host_player_id"`
	Players      []PlayerInfo `json:"players"`
	MaxPlayers   int          `json:"max_players"`
	MinPlayers   int          `json:"min_players"`
	CreatedAt    time.Time    `json:"created_at"`
	Status       string       `json:"status"`
	JoinTokens   []string     `json:"-"`
	mutex        sync.RWMutex
}

// JoinToken represents a session token for a player in a lobby
type JoinToken struct {
	Token     string    `json:"token"`
	LobbyID   string    `json:"lobby_id"`
	PlayerID  string    `json:"player_id"` // Now storing playerID instead of name/avatar
	ExpiresAt time.Time `json:"expires_at"`
}

// LobbyInfo represents public lobby information
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

// NewLobbyManager creates a new lobby manager
func NewLobbyManager(broadcaster Broadcaster, supervisor Supervisor) *LobbyManager {
	return &LobbyManager{
		lobbies:     make(map[string]*Lobby),
		tokens:      make(map[string]*JoinToken),
		actors:      make(map[string]*LobbyActor),
		broadcaster: broadcaster,
		supervisor:  supervisor,
	}
}

// CreateLobby creates a new lobby and returns the lobby ID, host player ID, and host session token
func (lm *LobbyManager) CreateLobby(hostName, hostAvatar, lobbyName string) (string, string, string, error) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lobbyID := uuid.New().String()
	hostPlayerID := uuid.New().String()

	// CRITICAL FIX: Create the host's info and add them to the player list IMMEDIATELY.
	hostInfo := PlayerInfo{ID: hostPlayerID, Name: hostName, Avatar: hostAvatar}
	lobby := &Lobby{
		ID:           lobbyID,
		Name:         lobbyName,
		HostPlayerID: hostPlayerID,
		Players:      []PlayerInfo{hostInfo}, // Host is added to the list from the start.
		MaxPlayers:   8,
		MinPlayers:   4,
		CreatedAt:    time.Now(),
		Status:       "WAITING",
	}
	lm.lobbies[lobbyID] = lobby

	actor := NewLobbyActor(lobby, lm.broadcaster, lm) // Pass the pointer to the lobby
	lm.actors[lobbyID] = actor
	actor.Start()

	sessionToken, err := lm.generateSessionToken(lobbyID, hostPlayerID)
	if err != nil {
		// Clean up on failure
		delete(lm.lobbies, lobbyID)
		delete(lm.actors, lobbyID)
		return "", "", "", fmt.Errorf("failed to generate host session token: %w", err)
	}

	return lobbyID, hostPlayerID, sessionToken, nil
}

// JoinLobby atomically adds the player to the lobby state and returns a session token.
func (lm *LobbyManager) JoinLobby(lobbyID, playerName, playerAvatar string) (string, string, error) {
	lm.mutex.Lock()
	lobby, exists := lm.lobbies[lobbyID]
	lm.mutex.Unlock() // Unlock manager early, we just need the lobby pointer

	if !exists {
		return "", "", fmt.Errorf("lobby not found")
	}

	// Lock the specific lobby we're modifying
	lobby.mutex.Lock()
	defer lobby.mutex.Unlock()

	if lobby.Status != "WAITING" {
		return "", "", fmt.Errorf("lobby is not accepting new players")
	}
	if len(lobby.Players) >= lobby.MaxPlayers {
		return "", "", fmt.Errorf("lobby is full")
	}

	playerID := uuid.New().String()
	playerInfo := PlayerInfo{ID: playerID, Name: playerName, Avatar: playerAvatar}
	sessionToken, err := lm.generateSessionToken(lobbyID, playerID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate session token: %w", err)
	}

	lobby.Players = append(lobby.Players, playerInfo)

	if actor, ok := lm.actors[lobbyID]; ok {
		actor.BroadcastLobbyUpdate()
	} else {
		return "", "", fmt.Errorf("consistency error: lobby actor not found for %s", lobbyID)
	}

	return playerID, sessionToken, nil
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

// ValidateAndConsumeToken is deprecated in favor of ValidateSession
func (lm *LobbyManager) ValidateAndConsumeToken(tokenStr string) (*PlayerInfo, string, error) {
	return nil, "", fmt.Errorf("ValidateAndConsumeToken is deprecated, use ValidateSession")
}

// GetLobbyList returns a list of available lobbies
func (lm *LobbyManager) GetLobbyList() []LobbyInfo {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var lobbies []LobbyInfo
	for _, lobby := range lm.lobbies {
		if lobby.Status == "WAITING" {
			lobbies = append(lobbies, LobbyInfo{
				ID:          lobby.ID,
				Name:        lobby.Name,
				PlayerCount: len(lobby.Players),
				MaxPlayers:  lobby.MaxPlayers,
				MinPlayers:  lobby.MinPlayers,
				CreatedAt:   lobby.CreatedAt,
				Status:      lobby.Status,
				CanJoin:     len(lobby.Players) < lobby.MaxPlayers,
			})
		}
	}

	return lobbies
}

// GetLobby returns a lobby by ID
func (lm *LobbyManager) GetLobby(lobbyID string) (*Lobby, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	lobby, exists := lm.lobbies[lobbyID]
	return lobby, exists
}

// TransitionLobbyToGame handles the handoff from lobby to game
func (lm *LobbyManager) TransitionLobbyToGame(lobbyID string) error {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lobby, exists := lm.lobbies[lobbyID]
	if !exists {
		return fmt.Errorf("lobby not found to transition")
	}

	lobby.Status = "STARTING"

	// Get a copy of player info and host ID to pass to the game actor
	playerInfos := make([]PlayerInfo, len(lobby.Players))
	copy(playerInfos, lobby.Players)

	// Tell the supervisor to create the GameActor and send it the initial state
	err := lm.supervisor.CreateGameWithPlayers(lobbyID, playerInfos, lobby.HostPlayerID)
	if err != nil {
		lobby.Status = "WAITING" // Revert on failure
		return err
	}

	// Successfully created game, now stop and remove the LobbyActor
	if actor, ok := lm.actors[lobbyID]; ok {
		actor.Stop()
		delete(lm.actors, lobbyID)
	}

	// Mark lobby as started to remove it from public list
	lobby.Status = "IN_PROGRESS"

	return nil
}

// GetLobbyActor returns a lobby actor by ID (implements TokenValidator)
func (lm *LobbyManager) GetLobbyActor(lobbyID string) (*LobbyActor, bool) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	actor, exists := lm.actors[lobbyID]
	return actor, exists
}

// RemoveLobby removes a lobby and invalidates all its tokens
func (lm *LobbyManager) RemoveLobby(lobbyID string) {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	lobby, exists := lm.lobbies[lobbyID]
	if !exists {
		return
	}

	// Stop and remove lobby actor
	if actor, exists := lm.actors[lobbyID]; exists {
		actor.Stop()
		delete(lm.actors, lobbyID)
	}

	// Invalidate all tokens for this lobby
	for _, tokenStr := range lobby.JoinTokens {
		delete(lm.tokens, tokenStr)
	}

	delete(lm.lobbies, lobbyID)
}

// CleanupExpiredTokens removes expired tokens
func (lm *LobbyManager) CleanupExpiredTokens() {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	now := time.Now()
	for tokenStr, token := range lm.tokens {
		if now.After(token.ExpiresAt) {
			delete(lm.tokens, tokenStr)
		}
	}
}

// generateSessionToken creates a session token for a player in a lobby
func (lm *LobbyManager) generateSessionToken(lobbyID, playerID string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	tokenStr := hex.EncodeToString(bytes)

	token := &JoinToken{
		Token:     tokenStr,
		LobbyID:   lobbyID,
		PlayerID:  playerID,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Longer session duration
	}

	lm.tokens[tokenStr] = token
	return tokenStr, nil
}
