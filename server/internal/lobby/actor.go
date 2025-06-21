package lobby

import (
	"log"
	"time"

	"github.com/xjhc/alignment/core"
)

// LobbyActor manages a single lobby waiting room
type LobbyActor struct {
	lobbyID     string
	lobby       *Lobby
	mailbox     chan interface{}
	shutdown    chan struct{}
	broadcaster Broadcaster
	manager     LobbyManagerInterface
}

// Broadcaster sends events to connected clients
type Broadcaster interface {
	BroadcastToGame(gameID string, event core.Event) error
	SendToPlayer(gameID, playerID string, event core.Event) error
}

// BroadcastUpdate is a message to trigger a lobby state broadcast
type BroadcastUpdate struct{}

// SendStateToPlayer is a message to send current state to a specific player
type SendStateToPlayer struct {
	PlayerID string
}

// LobbyManagerInterface for dependency injection
type LobbyManagerInterface interface {
	TransitionLobbyToGame(lobbyID string) error
}

// NewLobbyActor creates a new lobby actor
func NewLobbyActor(lobby *Lobby, broadcaster Broadcaster, manager LobbyManagerInterface) *LobbyActor {
	return &LobbyActor{
		lobbyID:     lobby.ID,
		lobby:       lobby,
		mailbox:     make(chan interface{}, 100),
		shutdown:    make(chan struct{}),
		broadcaster: broadcaster,
		manager:     manager,
	}
}

// Start begins the lobby actor's message processing loop
func (la *LobbyActor) Start() {
	go la.run()
}

// Stop gracefully shuts down the lobby actor
func (la *LobbyActor) Stop() {
	close(la.shutdown)
}

// SendAction sends an action to the lobby actor
func (la *LobbyActor) SendAction(action core.Action) {
	select {
	case la.mailbox <- action:
	case <-la.shutdown:
		log.Printf("LobbyActor %s: Attempted to send action to stopped actor", la.lobbyID)
	}
}

// BroadcastLobbyUpdate sends a message to the actor to trigger a broadcast.
func (la *LobbyActor) BroadcastLobbyUpdate() {
	select {
	case la.mailbox <- BroadcastUpdate{}:
	case <-la.shutdown:
		log.Printf("LobbyActor %s: Attempted to broadcast update to stopped actor", la.lobbyID)
	}
}

// SendCurrentStateToPlayer tells the actor to send current state to a specific player
func (la *LobbyActor) SendCurrentStateToPlayer(playerID string) {
	select {
	case la.mailbox <- SendStateToPlayer{PlayerID: playerID}:
	case <-la.shutdown:
		log.Printf("LobbyActor %s: Attempted to send state to stopped actor", la.lobbyID)
	}
}

// run is the main message processing loop
func (la *LobbyActor) run() {
	log.Printf("LobbyActor %s: Started with %d players", la.lobbyID, len(la.lobby.Players))
	for {
		select {
		case msg := <-la.mailbox:
			switch v := msg.(type) {
			case core.Action:
				if err := la.handleAction(v); err != nil {
					log.Printf("LobbyActor %s: Error handling action %s: %v", la.lobbyID, v.Type, err)
				}
			case SendStateToPlayer:
				if err := la.sendWelcomeSequence(v.PlayerID); err != nil {
					log.Printf("LobbyActor %s: Error sending welcome sequence to %s: %v", la.lobbyID, v.PlayerID, err)
				}
			case BroadcastUpdate:
				log.Printf("LobbyActor %s: Received broadcast trigger. Broadcasting state for %d players.", la.lobbyID, len(la.lobby.Players))
				if err := la.broadcastLobbyUpdate(); err != nil {
					log.Printf("LobbyActor %s: Error broadcasting update: %v", la.lobbyID, err)
				}
			}
		case <-la.shutdown:
			log.Printf("LobbyActor %s: Shutting down", la.lobbyID)
			return
		}
	}
}

// handleAction processes individual actions
func (la *LobbyActor) handleAction(action core.Action) error {
	log.Printf("LobbyActor %s: Handling action %s from player %s", la.lobbyID, action.Type, action.PlayerID)

	switch action.Type {
	case core.ActionLeaveGame:
		return la.handlePlayerLeft(action)
	case core.ActionStartGame:
		return la.handleStartGame(action)
	default:
		log.Printf("LobbyActor %s: Unknown action type: %s", la.lobbyID, action.Type)
		return nil
	}
}

// handlePlayerLeft processes a player leaving the lobby
func (la *LobbyActor) handlePlayerLeft(action core.Action) error {
	la.lobby.mutex.Lock()
	defer la.lobby.mutex.Unlock()

	found := false
	for i, player := range la.lobby.Players {
		if player.ID == action.PlayerID {
			la.lobby.Players = append(la.lobby.Players[:i], la.lobby.Players[i+1:]...)
			found = true
			break
		}
	}

	if found {
		log.Printf("LobbyActor %s: Player %s left. Broadcasting update. Current players: %d", la.lobbyID, action.PlayerID, len(la.lobby.Players))
		// The broadcastLobbyUpdate will acquire its own RLock, which is fine.
		return la.broadcastLobbyUpdate()
	}
	return nil
}

// handleStartGame now delegates to the manager
func (la *LobbyActor) handleStartGame(action core.Action) error {
	if action.PlayerID != la.lobby.HostPlayerID {
		log.Printf("LobbyActor %s: Non-host player %s attempted to start game", la.lobbyID, action.PlayerID)
		return nil
	}

	if len(la.lobby.Players) < la.lobby.MinPlayers {
		event := core.Event{
			Type:    core.EventSystemMessage,
			GameID:  la.lobbyID,
			Payload: map[string]interface{}{"message": "Not enough players to start"},
		}
		return la.broadcaster.SendToPlayer(la.lobbyID, action.PlayerID, event)
	}

	// Delegate to manager to handle the transition
	return la.manager.TransitionLobbyToGame(la.lobbyID)
}

// createLobbyStateEvent creates a lobby state update event
func (la *LobbyActor) createLobbyStateEvent() core.Event {
	// Acquire a read lock to safely access the shared lobby state
	la.lobby.mutex.RLock()
	defer la.lobby.mutex.RUnlock()

	// Create a copy of the players to avoid data races in the payload
	playersCopy := make([]PlayerInfo, len(la.lobby.Players))
	copy(playersCopy, la.lobby.Players)

	payload := map[string]interface{}{
		"lobby_id":     la.lobby.ID,
		"name":         la.lobby.Name,
		"host_id":      la.lobby.HostPlayerID,
		"players":      playersCopy,
		"player_count": len(playersCopy),
		"max_players":  la.lobby.MaxPlayers,
		"min_players":  la.lobby.MinPlayers,
		"status":       la.lobby.Status,
		"can_start":    len(playersCopy) >= la.lobby.MinPlayers,
	}

	return core.Event{
		Type:      "LOBBY_STATE_UPDATE",
		GameID:    la.lobbyID,
		Payload:   payload,
		Timestamp: time.Now(),
	}
}

// broadcastLobbyUpdate sends the current lobby state to all connected clients
func (la *LobbyActor) broadcastLobbyUpdate() error {
	event := la.createLobbyStateEvent()
	return la.broadcaster.BroadcastToGame(la.lobbyID, event)
}

// sendWelcomeSequence sends the complete welcome sequence to a player
func (la *LobbyActor) sendWelcomeSequence(playerID string) error {
	// 1. Send CLIENT_IDENTIFIED event first
	welcomeEvent := core.Event{
		Type:      "CLIENT_IDENTIFIED",
		Payload:   map[string]interface{}{"your_player_id": playerID},
		Timestamp: time.Now(),
	}
	if err := la.broadcaster.SendToPlayer(la.lobbyID, playerID, welcomeEvent); err != nil {
		log.Printf("LobbyActor %s: Failed to send welcome event to %s: %v", la.lobbyID, playerID, err)
		return err
	}

	// 2. Then send the current lobby state
	stateEvent := la.createLobbyStateEvent()
	if err := la.broadcaster.SendToPlayer(la.lobbyID, playerID, stateEvent); err != nil {
		log.Printf("LobbyActor %s: Failed to send current state to %s: %v", la.lobbyID, playerID, err)
		return err
	}

	log.Printf("LobbyActor %s: Sent welcome sequence to %s", la.lobbyID, playerID)
	return nil
}

// GetLobbyState returns the current lobby state
func (la *LobbyActor) GetLobbyState() *Lobby {
	return la.lobby
}
