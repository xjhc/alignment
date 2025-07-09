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
	Players         map[string]interfaces.PlayerActorInterface // Map of playerID -> PlayerActor
	PlayerJoinTimes map[string]time.Time                   // Map of playerID -> join timestamp
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
	players := make(map[string]interfaces.PlayerActorInterface)
	players[hostPlayerID] = hostActor

	playerJoinTimes := make(map[string]time.Time)
	now := time.Now()
	playerJoinTimes[hostPlayerID] = now

	return &Lobby{
		ID:              id,
		Name:            name,
		HostPlayerID:    hostPlayerID,
		Players:         players,
		PlayerJoinTimes: playerJoinTimes,
		MaxPlayers:      8,
		MinPlayers:      2,
		CreatedAt:       now,
		LastActivity:    now,
		Status:          "WAITING",
		IsPrivate:       isPrivate,
		GameSettings:    settings,
	}
}

// createStateUpdate_unsafe creates a state update under lock
func (l *Lobby) createStateUpdate_unsafe() LobbyStateUpdate {
	var infos []PlayerInfo
	for _, actor := range l.Players {
		playerID := actor.GetPlayerID()
		joinTime, exists := l.PlayerJoinTimes[playerID]
		if !exists {
			joinTime = time.Now() // Fallback to current time if not tracked
		}
		infos = append(infos, PlayerInfo{
			ID:       playerID,
			Name:     actor.GetPlayerName(),
			Avatar:   actor.GetPlayerAvatar(),
			JoinedAt: joinTime,
		})
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
	players := make(map[string]interfaces.PlayerActorInterface, len(l.Players))
	for id, actor := range l.Players {
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
	l.Players[playerID] = playerActor
	l.PlayerJoinTimes[playerID] = now
	l.LastActivity = now // Update last activity when player joins

	// Create the update and broadcast it to all players in the lobby
	l.broadcastStateUpdate()

	return nil
}

// RemovePlayer removes a player from the lobby
func (l *Lobby) RemovePlayer(playerID string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, exists := l.Players[playerID]; !exists {
		return
	}

	delete(l.Players, playerID)
	delete(l.PlayerJoinTimes, playerID)
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
	for id, actor := range l.Players {
		players[id] = actor
	}
	return players
}

// GetPlayerInfos returns player information for state updates
func (l *Lobby) GetPlayerInfos() []PlayerInfo {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	var infos []PlayerInfo
	for _, actor := range l.Players {
		playerID := actor.GetPlayerID()
		joinTime, exists := l.PlayerJoinTimes[playerID]
		if !exists {
			joinTime = time.Now() // Fallback to current time if not tracked
		}
		infos = append(infos, PlayerInfo{
			ID:       playerID,
			Name:     actor.GetPlayerName(),
			Avatar:   actor.GetPlayerAvatar(),
			JoinedAt: joinTime,
		})
	}
	return infos
}

// broadcastStateUpdate sends lobby state to all players
// NOTE: This method assumes the caller already holds the lobby lock
func (l *Lobby) broadcastStateUpdate() {
	var playerInfos []PlayerInfo
	for _, actor := range l.Players {
		playerID := actor.GetPlayerID()
		joinTime, exists := l.PlayerJoinTimes[playerID]
		if !exists {
			joinTime = time.Now() // Fallback to current time if not tracked
		}
		playerInfos = append(playerInfos, PlayerInfo{
			ID:       playerID,
			Name:     actor.GetPlayerName(),
			Avatar:   actor.GetPlayerAvatar(),
			JoinedAt: joinTime,
		})
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

	for _, actor := range l.Players {
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
	for playerID, joinTime := range l.PlayerJoinTimes {
		if playerID == l.HostPlayerID {
			continue // Skip current host
		}
		if _, exists := l.Players[playerID]; !exists {
			continue // Skip if player no longer in lobby
		}
		if newHostID == "" || joinTime.Before(earliestJoinTime) {
			earliestJoinTime = joinTime
			newHostID = playerID
		}
	}

	if newHostID != "" {
		l.HostPlayerID = newHostID
	}

	return newHostID
}

// PlayerInfo holds basic info for a player in the lobby
type PlayerInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Avatar   string    `json:"avatar"`
	JoinedAt time.Time `json:"joinedAt"`
}

// Custom errors
var (
	ErrLobbyNotAcceptingPlayers = errors.New("lobby is not accepting new players")
	ErrLobbyFull                = errors.New("lobby is full")
)