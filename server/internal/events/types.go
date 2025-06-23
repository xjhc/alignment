package events

// PlayerDisconnectedEvent is published when a player's WebSocket connection is lost
type PlayerDisconnectedEvent struct {
	PlayerID string
	LobbyID  string
	GameID   string
}

func (e PlayerDisconnectedEvent) EventType() string {
	return "player_disconnected"
}

// GameEndedEvent is published when a game reaches its natural conclusion
type GameEndedEvent struct {
	GameID string
	Reason string // "completed", "abandoned", "error"
}

func (e GameEndedEvent) EventType() string {
	return "game_ended"
}

// LobbyCreatedEvent is published when a new lobby is created
type LobbyCreatedEvent struct {
	LobbyID      string
	HostPlayerID string
	LobbyName    string
}

func (e LobbyCreatedEvent) EventType() string {
	return "lobby_created"
}

// GameStartedEvent is published when a lobby transitions to an active game
type GameStartedEvent struct {
	GameID   string
	LobbyID  string
	PlayerIDs []string
}

func (e GameStartedEvent) EventType() string {
	return "game_started"
}

// PlayerJoinedLobbyEvent is published when a player joins a lobby
type PlayerJoinedLobbyEvent struct {
	PlayerID string
	LobbyID  string
}

func (e PlayerJoinedLobbyEvent) EventType() string {
	return "player_joined_lobby"
}

// PlayerLeftLobbyEvent is published when a player leaves a lobby
type PlayerLeftLobbyEvent struct {
	PlayerID string
	LobbyID  string
}

func (e PlayerLeftLobbyEvent) EventType() string {
	return "player_left_lobby"
}