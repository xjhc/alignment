package mocks

import (
	"sync"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockLobbyManager is a mock implementation of the LobbyManagerInterface
type MockLobbyManager struct {
	sync.Mutex  // Add mutex for thread-safe access to mock data
	Wg sync.WaitGroup  // Add WaitGroup for synchronization in tests

	JoinLobbyCalls           []JoinLobbyCall
	JoinLobbyWithActorCalls  []JoinLobbyWithActorCall
	LeaveLobbyCall           []LeaveLobbyCall
	StartGameCalls           []StartGameCall
	CreateLobbyCalls         []CreateLobbyCall
	GetLobbyListCalls        []GetLobbyListCall
	GetLobbyCalls            []GetLobbyCall
	ValidateSessionCalls     []ValidateSessionCall
	GetPlayerInfoCalls       []GetPlayerInfoCall

	JoinLobbyResults         []JoinLobbyResult
	JoinLobbyWithActorResults []error
	LeaveLobbyResults        []error
	StartGameResults         []error
	CreateLobbyResults       []CreateLobbyResult
	GetLobbyListResults      [][]interface{}
	GetLobbyResults          []GetLobbyResult
	ValidateSessionResults   []bool
	GetPlayerInfoResults     []GetPlayerInfoResult
}

type JoinLobbyCall struct {
	GameID       string
	PlayerName   string
	PlayerAvatar string
}

type JoinLobbyResult struct {
	LobbyID      string
	SessionToken string
	Error        error
}

type JoinLobbyWithActorCall struct {
	LobbyID string
	Player  interfaces.PlayerActorInterface
}

type LeaveLobbyCall struct {
	LobbyID  string
	PlayerID string
}

type StartGameCall struct {
	LobbyID      string
	HostPlayerID string
}

type CreateLobbyCall struct {
	HostPlayer interfaces.PlayerActorInterface
	LobbyName  string
}

type CreateLobbyResult struct {
	LobbyID string
	Error   error
}

type GetLobbyListCall struct{}

type GetLobbyCall struct {
	LobbyID string
}

type GetLobbyResult struct {
	Lobby interface{}
	Found bool
}

type ValidateSessionCall struct {
	GameID       string
	PlayerID     string
	SessionToken string
}

type GetPlayerInfoCall struct {
	GameID   string
	PlayerID string
}

type GetPlayerInfoResult struct {
	PlayerName   string
	PlayerAvatar string
	Error        error
}

// Ensure MockLobbyManager implements the interface at compile time
var _ interfaces.LobbyManagerInterface = (*MockLobbyManager)(nil)

func (m *MockLobbyManager) JoinLobby(gameID, playerName, playerAvatar string) (string, string, error) {
	m.JoinLobbyCalls = append(m.JoinLobbyCalls, JoinLobbyCall{
		GameID:       gameID,
		PlayerName:   playerName,
		PlayerAvatar: playerAvatar,
	})

	if len(m.JoinLobbyResults) > 0 {
		result := m.JoinLobbyResults[0]
		if len(m.JoinLobbyResults) > 1 {
			m.JoinLobbyResults = m.JoinLobbyResults[1:]
		}
		return result.LobbyID, result.SessionToken, result.Error
	}

	return "", "", nil
}

func (m *MockLobbyManager) JoinLobbyWithActor(lobbyID string, player interfaces.PlayerActorInterface) error {
	m.JoinLobbyWithActorCalls = append(m.JoinLobbyWithActorCalls, JoinLobbyWithActorCall{
		LobbyID: lobbyID,
		Player:  player,
	})

	if len(m.JoinLobbyWithActorResults) > 0 {
		result := m.JoinLobbyWithActorResults[0]
		if len(m.JoinLobbyWithActorResults) > 1 {
			m.JoinLobbyWithActorResults = m.JoinLobbyWithActorResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockLobbyManager) LeaveLobby(lobbyID string, playerID string) error {
	m.LeaveLobbyCall = append(m.LeaveLobbyCall, LeaveLobbyCall{
		LobbyID:  lobbyID,
		PlayerID: playerID,
	})

	if len(m.LeaveLobbyResults) > 0 {
		result := m.LeaveLobbyResults[0]
		if len(m.LeaveLobbyResults) > 1 {
			m.LeaveLobbyResults = m.LeaveLobbyResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockLobbyManager) StartGame(lobbyID string, hostPlayerID string) error {
	defer m.Wg.Done() // Signal that this function has completed
	
	m.Lock()
	m.StartGameCalls = append(m.StartGameCalls, StartGameCall{
		LobbyID:      lobbyID,
		HostPlayerID: hostPlayerID,
	})

	var result error
	if len(m.StartGameResults) > 0 {
		result = m.StartGameResults[0]
		if len(m.StartGameResults) > 1 {
			m.StartGameResults = m.StartGameResults[1:]
		}
	}
	m.Unlock()

	return result
}

func (m *MockLobbyManager) CreateLobby(hostPlayer interfaces.PlayerActorInterface, lobbyName string) (string, error) {
	m.CreateLobbyCalls = append(m.CreateLobbyCalls, CreateLobbyCall{
		HostPlayer: hostPlayer,
		LobbyName:  lobbyName,
	})

	if len(m.CreateLobbyResults) > 0 {
		result := m.CreateLobbyResults[0]
		if len(m.CreateLobbyResults) > 1 {
			m.CreateLobbyResults = m.CreateLobbyResults[1:]
		}
		return result.LobbyID, result.Error
	}

	return "", nil
}

func (m *MockLobbyManager) GetLobbyList() []interface{} {
	m.GetLobbyListCalls = append(m.GetLobbyListCalls, GetLobbyListCall{})

	if len(m.GetLobbyListResults) > 0 {
		result := m.GetLobbyListResults[0]
		if len(m.GetLobbyListResults) > 1 {
			m.GetLobbyListResults = m.GetLobbyListResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockLobbyManager) GetLobby(lobbyID string) (interface{}, bool) {
	m.GetLobbyCalls = append(m.GetLobbyCalls, GetLobbyCall{
		LobbyID: lobbyID,
	})

	if len(m.GetLobbyResults) > 0 {
		result := m.GetLobbyResults[0]
		if len(m.GetLobbyResults) > 1 {
			m.GetLobbyResults = m.GetLobbyResults[1:]
		}
		return result.Lobby, result.Found
	}

	return nil, false
}

func (m *MockLobbyManager) ValidateSession(gameID, playerID, sessionToken string) bool {
	m.ValidateSessionCalls = append(m.ValidateSessionCalls, ValidateSessionCall{
		GameID:       gameID,
		PlayerID:     playerID,
		SessionToken: sessionToken,
	})

	if len(m.ValidateSessionResults) > 0 {
		result := m.ValidateSessionResults[0]
		if len(m.ValidateSessionResults) > 1 {
			m.ValidateSessionResults = m.ValidateSessionResults[1:]
		}
		return result
	}

	return false
}

func (m *MockLobbyManager) GetPlayerInfo(gameID, playerID string) (string, string, error) {
	m.GetPlayerInfoCalls = append(m.GetPlayerInfoCalls, GetPlayerInfoCall{
		GameID:   gameID,
		PlayerID: playerID,
	})

	if len(m.GetPlayerInfoResults) > 0 {
		result := m.GetPlayerInfoResults[0]
		if len(m.GetPlayerInfoResults) > 1 {
			m.GetPlayerInfoResults = m.GetPlayerInfoResults[1:]
		}
		return result.PlayerName, result.PlayerAvatar, result.Error
	}

	return "", "", nil
}