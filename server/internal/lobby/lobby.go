package lobby

import (
	"errors"
	"sync"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// LobbyStateUpdate represents lobby state changes
type LobbyStateUpdate struct {
	LobbyID      string
	Players      []PlayerInfo
	HostID       string
	CanStart     bool
	LobbyName    string
	GameSettings core.GameSettings
}

// Lobby represents a pre-game waiting room as a simple data structure
type Lobby struct {
	ID              string
	Name            string
	HostPlayerID    string
	Players         map[string]*PlayerInfo                     // Map of playerID -> PlayerInfo (persistent)
	PlayerActors    map[string]interfaces.PlayerActorInterface // Map of playerID -> PlayerActor (live connections)
	MaxPlayers      int
	MinPlayers      int
	CreatedAt       time.Time
	LastActivity    time.Time // Tracks when any player last joined or left
	Status          string
	IsPrivate       bool // If true, lobby won't appear in public listings
	GameSettings    core.GameSettings
	mutex           sync.RWMutex
}

// NewLobby creates a new lobby with the host player
func NewLobby(id, name, hostPlayerID string, hostActor interfaces.PlayerActorInterface, isPrivate bool, settings core.GameSettings) *Lobby {
	players := make(map[string]*PlayerInfo)
	playerActors := make(map[string]interfaces.PlayerActorInterface)
	
	now := time.Now()
	
	// Create persistent player info for host
	players[hostPlayerID] = &PlayerInfo{
		ID:               hostPlayerID,
		Name:             hostActor.GetPlayerName(),
		Avatar:           hostActor.GetPlayerAvatar(),
		JoinedAt:         now,
		ConnectionStatus: "CONNECTED",
	}
	
	// Store live connection for host
	playerActors[hostPlayerID] = hostActor

	return &Lobby{
		ID:           id,
		Name:         name,
		HostPlayerID: hostPlayerID,
		Players:      players,
		PlayerActors: playerActors,
		MaxPlayers:   8,
		MinPlayers:   2,
		CreatedAt:    now,
		LastActivity: now,
		Status:       "WAITING",
		IsPrivate:    isPrivate,
		GameSettings: settings,
	}
}

// createStateUpdate_unsafe creates a state update under lock
func (l *Lobby) createStateUpdate_unsafe() LobbyStateUpdate {
	var infos []PlayerInfo
	for _, playerInfo := range l.Players {
		infos = append(infos, *playerInfo)
	}

	return LobbyStateUpdate{
		LobbyID:      l.ID,
		Players:      infos,
		HostID:       l.HostPlayerID,
		CanStart:     len(l.Players) >= l.MinPlayers && (l.Status == "WAITING" || l.Status == "COUNTDOWN"),
		LobbyName:    l.Name,
		GameSettings: l.GameSettings,
	}
}

// copyPlayers_unsafe copies players map under lock
func (l *Lobby) copyPlayers_unsafe() map[string]interfaces.PlayerActorInterface {
	players := make(map[string]interfaces.PlayerActorInterface, len(l.PlayerActors))
	for id, actor := range l.PlayerActors {
		players[id] = actor
	}
	return players
}

// AddPlayer adds a player to the lobby
func (l *Lobby) AddPlayer(playerActor interfaces.PlayerActorInterface) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Allow players to join if lobby is waiting or if it's the host connecting for the first time
	if l.Status != "WAITING" && l.Status != "WAITING_FOR_HOST" {
		return ErrLobbyNotAcceptingPlayers
	}

	// Special case: if status is WAITING_FOR_HOST, only the host can join
	if l.Status == "WAITING_FOR_HOST" && l.HostPlayerID != playerActor.GetPlayerID() {
		return ErrLobbyNotAcceptingPlayers
	}

	if len(l.Players) >= l.MaxPlayers {
		return ErrLobbyFull
	}

	playerID := playerActor.GetPlayerID()
	now := time.Now()
	
	// Check if player already exists (rejoining)
	if existingInfo, exists := l.Players[playerID]; exists {
		// Update existing player info and mark as connected
		existingInfo.ConnectionStatus = "CONNECTED"
		existingInfo.Name = playerActor.GetPlayerName() // Update in case name changed
		existingInfo.Avatar = playerActor.GetPlayerAvatar() // Update in case avatar changed
	} else {
		// Create new persistent player info
		l.Players[playerID] = &PlayerInfo{
			ID:               playerID,
			Name:             playerActor.GetPlayerName(),
			Avatar:           playerActor.GetPlayerAvatar(),
			JoinedAt:         now,
			ConnectionStatus: "CONNECTED",
		}
	}
	
	// Store live connection
	l.PlayerActors[playerID] = playerActor
	l.LastActivity = now // Update last activity when player joins

	// Create the update and broadcast it to all players in the lobby
	l.broadcastStateUpdate()

	return nil
}

// DisconnectPlayer marks a player as disconnected but keeps their slot
func (l *Lobby) DisconnectPlayer(playerID string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if playerInfo, exists := l.Players[playerID]; exists {
		playerInfo.ConnectionStatus = "DISCONNECTED"
		delete(l.PlayerActors, playerID) // Remove live connection
		l.LastActivity = time.Now()
		l.broadcastStateUpdate()
	}
}

// ReconnectPlayer marks a player as reconnected and restores their connection
func (l *Lobby) ReconnectPlayer(playerID string, playerActor interfaces.PlayerActorInterface) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if playerInfo, exists := l.Players[playerID]; exists {
		playerInfo.ConnectionStatus = "CONNECTED"
		l.PlayerActors[playerID] = playerActor // Restore live connection
		l.LastActivity = time.Now()
		l.broadcastStateUpdate()
	}
}

// RemovePlayer permanently removes a player from the lobby
func (l *Lobby) RemovePlayer(playerID string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, exists := l.Players[playerID]; !exists {
		return
	}

	delete(l.Players, playerID)
	delete(l.PlayerActors, playerID)
	l.LastActivity = time.Now() // Update last activity when player leaves

	// Create the update and broadcast it to all players in the lobby
	l.broadcastStateUpdate()
}

// CanStart returns whether the lobby can start a game
func (l *Lobby) CanStart() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return len(l.Players) >= l.MinPlayers && (l.Status == "WAITING" || l.Status == "COUNTDOWN")
}

// GetPlayerActors returns a copy of the player actors map
func (l *Lobby) GetPlayerActors() map[string]interfaces.PlayerActorInterface {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	players := make(map[string]interfaces.PlayerActorInterface)
	for id, actor := range l.PlayerActors {
		players[id] = actor
	}
	return players
}

// GetPlayerInfos returns player information for state updates
func (l *Lobby) GetPlayerInfos() []PlayerInfo {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	var infos []PlayerInfo
	for _, playerInfo := range l.Players {
		infos = append(infos, *playerInfo)
	}
	return infos
}

// broadcastStateUpdate sends lobby state to all players
// NOTE: This method assumes the caller already holds the lobby lock
func (l *Lobby) broadcastStateUpdate() {
	var playerInfos []PlayerInfo
	for _, playerInfo := range l.Players {
		playerInfos = append(playerInfos, *playerInfo)
	}

	update := LobbyStateUpdate{
		LobbyID:      l.ID,
		Players:      playerInfos,
		HostID:       l.HostPlayerID,
		CanStart:     len(l.Players) >= l.MinPlayers && (l.Status == "WAITING" || l.Status == "COUNTDOWN"),
		LobbyName:    l.Name,
		GameSettings: l.GameSettings,
	}

	// This event is now a struct, not a core.Event
	// We'll wrap it in a core.Event for consistency
	event := core.Event{
		Type:      "LOBBY_STATE_UPDATE",
		GameID:    l.ID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"lobby_id":       update.LobbyID,
			"players":        update.Players,
			"host_id":        update.HostID,
			"can_start":      update.CanStart,
			"name":           update.LobbyName,
			"game_settings":  update.GameSettings,
		},
	}

	// Send to all connected players only
	for _, actor := range l.PlayerActors {
		// The PlayerActor will handle marshaling this event to JSON
		actor.SendServerMessage(event)
	}
}

// SetStatus updates the lobby status
func (l *Lobby) SetStatus(status string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.Status = status
}

// SetPrivacy updates the lobby privacy setting
func (l *Lobby) SetPrivacy(isPrivate bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.IsPrivate = isPrivate
}

// IsPrivateLobby returns whether the lobby is private
func (l *Lobby) IsPrivateLobby() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.IsPrivate
}

func (l *Lobby) Lock() {
	 l.mutex.Lock()
}

func (l *Lobby) Unlock() {
	 l.mutex.Unlock()
}

func (l *Lobby) RLock() {
	 l.mutex.RLock()
}

func (l *Lobby) RUnlock() {
	 l.mutex.RUnlock()
}

// TransferHostToNextPlayer transfers host to the player who joined earliest (excluding current host)
// Returns the new host player ID, or empty string if no other players exist
// NOTE: This method assumes the caller already holds the lobby lock
func (l *Lobby) TransferHostToNextPlayer() string {
	if len(l.Players) <= 1 {
		return "" // No other players to transfer to
	}

	var earliestJoinTime time.Time
	var newHostID string

	// Find the player (excluding current host) who joined earliest
	for playerID, playerInfo := range l.Players {
		if playerID == l.HostPlayerID {
			continue // Skip current host
		}
		if newHostID == "" || playerInfo.JoinedAt.Before(earliestJoinTime) {
			earliestJoinTime = playerInfo.JoinedAt
			newHostID = playerID
		}
	}

	if newHostID != "" {
		l.HostPlayerID = newHostID
	}

	return newHostID
}

// IsPlayerDisconnected checks if a player exists but is disconnected
func (l *Lobby) IsPlayerDisconnected(playerID string) bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	
	if playerInfo, exists := l.Players[playerID]; exists {
		return playerInfo.ConnectionStatus == "DISCONNECTED"
	}
	return false
}

// PlayerInfo holds basic info for a player in the lobby
type PlayerInfo struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Avatar           string    `json:"avatar"`
	JoinedAt         time.Time `json:"joinedAt"`
	ConnectionStatus string    `json:"connectionStatus"` // "CONNECTED", "DISCONNECTED"
}

// Custom errors
var (
	ErrLobbyNotAcceptingPlayers = errors.New("lobby is not accepting new players")
	ErrLobbyFull                = errors.New("lobby is full")
)