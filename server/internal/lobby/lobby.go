package lobby

import (
	"errors"
	"sync"
	"time"

	"github.com/xjhc/alignment/server/internal/interfaces"
)

// LobbyStateUpdate represents lobby state changes
type LobbyStateUpdate struct {
	LobbyID   string
	Players   []PlayerInfo
	HostID    string
	CanStart  bool
	LobbyName string
}

// Lobby represents a pre-game waiting room as a simple data structure
type Lobby struct {
	ID           string
	Name         string
	HostPlayerID string
	Players      map[string]interfaces.PlayerActorInterface // Map of playerID -> PlayerActor
	MaxPlayers   int
	MinPlayers   int
	CreatedAt    time.Time
	Status       string
	mutex        sync.RWMutex
}

// NewLobby creates a new lobby with the host player
func NewLobby(id, name, hostPlayerID string, hostActor interfaces.PlayerActorInterface) *Lobby {
	players := make(map[string]interfaces.PlayerActorInterface)
	players[hostPlayerID] = hostActor

	return &Lobby{
		ID:           id,
		Name:         name,
		HostPlayerID: hostPlayerID,
		Players:      players,
		MaxPlayers:   8,
		MinPlayers:   2,
		CreatedAt:    time.Now(),
		Status:       "WAITING",
	}
}

// createStateUpdate_unsafe creates a state update under lock
func (l *Lobby) createStateUpdate_unsafe() LobbyStateUpdate {
	var infos []PlayerInfo
	for _, actor := range l.Players {
		infos = append(infos, PlayerInfo{
			ID:   actor.GetPlayerID(),
			Name: actor.GetPlayerName(),
		})
	}

	return LobbyStateUpdate{
		LobbyID:   l.ID,
		Players:   infos,
		HostID:    l.HostPlayerID,
		CanStart:  len(l.Players) >= l.MinPlayers && l.Status == "WAITING",
		LobbyName: l.Name,
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

	// Allow players to join if lobby is waiting or if it's the host connecting for the first time
	if l.Status != "WAITING" && l.Status != "WAITING_FOR_HOST" {
		l.mutex.Unlock()
		return ErrLobbyNotAcceptingPlayers
	}

	// Special case: if status is WAITING_FOR_HOST, only the host can join
	if l.Status == "WAITING_FOR_HOST" && l.HostPlayerID != playerActor.GetPlayerID() {
		l.mutex.Unlock()
		return ErrLobbyNotAcceptingPlayers
	}

	if len(l.Players) >= l.MaxPlayers {
		l.mutex.Unlock()
		return ErrLobbyFull
	}

	playerID := playerActor.GetPlayerID()
	l.Players[playerID] = playerActor

	// Create the update and a copy of the players *inside* the lock
	update := l.createStateUpdate_unsafe()
	playersToNotify := l.copyPlayers_unsafe()

	l.mutex.Unlock() // Unlock *before* broadcasting

	// Broadcast outside the lock
	for _, actor := range playersToNotify {
		actor.SendServerMessage(update)
	}

	return nil
}

// RemovePlayer removes a player from the lobby
func (l *Lobby) RemovePlayer(playerID string) {
	l.mutex.Lock()

	if _, exists := l.Players[playerID]; !exists {
		l.mutex.Unlock()
		return
	}

	delete(l.Players, playerID)

	update := l.createStateUpdate_unsafe()
	playersToNotify := l.copyPlayers_unsafe()

	l.mutex.Unlock() // Unlock *before* broadcasting

	for _, actor := range playersToNotify {
		actor.SendServerMessage(update)
	}
}

// CanStart returns whether the lobby can start a game
func (l *Lobby) CanStart() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return len(l.Players) >= l.MinPlayers && l.Status == "WAITING"
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
		infos = append(infos, PlayerInfo{
			ID:   actor.GetPlayerID(),
			Name: actor.GetPlayerName(),
		})
	}
	return infos
}

// broadcastStateUpdate sends lobby state to all players
// NOTE: This method assumes the caller already holds the lobby lock
func (l *Lobby) broadcastStateUpdate() {
	// Build player info list without calling GetPlayerInfos() to avoid re-entrant lock
	var playerInfos []PlayerInfo
	for _, actor := range l.Players {
		playerInfos = append(playerInfos, PlayerInfo{
			ID:   actor.GetPlayerID(),
			Name: actor.GetPlayerName(),
		})
	}

	update := LobbyStateUpdate{
		LobbyID:   l.ID,
		Players:   playerInfos,
		HostID:    l.HostPlayerID,
		CanStart:  len(l.Players) >= l.MinPlayers && l.Status == "WAITING", // Inline CanStart logic
		LobbyName: l.Name,
	}

	for _, actor := range l.Players {
		actor.SendServerMessage(update)
	}
}

// SetStatus updates the lobby status
func (l *Lobby) SetStatus(status string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.Status = status
}

// PlayerInfo holds basic info for a player in the lobby
type PlayerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Custom errors
var (
	ErrLobbyNotAcceptingPlayers = errors.New("lobby is not accepting new players")
	ErrLobbyFull               = errors.New("lobby is full")
)