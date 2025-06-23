package mocks

import (
	"sync"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// MockGameLifecycleManager is a mock implementation of the GameLifecycleManagerInterface
type MockGameLifecycleManager struct {
	sync.Mutex
	Wg sync.WaitGroup

	CreateLobbyViaHTTPCalls     []GLMCreateLobbyViaHTTPCall
	JoinLobbyCalls              []GLMJoinLobbyCall
	JoinLobbyWithActorCalls     []GLMJoinLobbyWithActorCall
	StartGameCalls              []GLMStartGameCall
	ValidateSessionTokenCalls   []GLMValidateSessionTokenCall
	SendActionToGameCalls       []GLMSendActionToGameCall
	StopCalls                   []GLMStopCall

	CreateLobbyViaHTTPResults   []GLMCreateLobbyViaHTTPResult
	JoinLobbyResults            []GLMJoinLobbyResult
	JoinLobbyWithActorResults   []error
	StartGameResults            []error
	ValidateSessionTokenResults []GLMValidateSessionTokenResult
	SendActionToGameResults     []error
}

type GLMCreateLobbyViaHTTPCall struct {
	HostPlayerName string
	LobbyName      string
	PlayerAvatar   string
}

type GLMCreateLobbyViaHTTPResult struct {
	LobbyID      string
	PlayerID     string
	SessionToken string
	Error        error
}

type GLMJoinLobbyCall struct {
	LobbyID      string
	PlayerName   string
	PlayerAvatar string
}

type GLMJoinLobbyResult struct {
	PlayerID     string
	SessionToken string
	Error        error
}

type GLMJoinLobbyWithActorCall struct {
	LobbyID     string
	PlayerActor interfaces.PlayerActorInterface
}

type GLMStartGameCall struct {
	LobbyID      string
	HostPlayerID string
}

type GLMValidateSessionTokenCall struct {
	Token string
}

type GLMValidateSessionTokenResult struct {
	TokenInfo interface{}
	Error     error
}

type GLMSendActionToGameCall struct {
	GameID string
	Action core.Action
}

type GLMStopCall struct{}

// Ensure MockGameLifecycleManager implements the interface at compile time
var _ interfaces.GameLifecycleManagerInterface = (*MockGameLifecycleManager)(nil)

func (m *MockGameLifecycleManager) CreateLobbyViaHTTP(hostPlayerName, lobbyName, playerAvatar string) (string, string, string, error) {
	m.Lock()
	defer m.Unlock()

	m.CreateLobbyViaHTTPCalls = append(m.CreateLobbyViaHTTPCalls, GLMCreateLobbyViaHTTPCall{
		HostPlayerName: hostPlayerName,
		LobbyName:      lobbyName,
		PlayerAvatar:   playerAvatar,
	})

	if len(m.CreateLobbyViaHTTPResults) > 0 {
		result := m.CreateLobbyViaHTTPResults[0]
		if len(m.CreateLobbyViaHTTPResults) > 1 {
			m.CreateLobbyViaHTTPResults = m.CreateLobbyViaHTTPResults[1:]
		}
		return result.LobbyID, result.PlayerID, result.SessionToken, result.Error
	}

	return "", "", "", nil
}

func (m *MockGameLifecycleManager) JoinLobby(lobbyID, playerName, playerAvatar string) (string, string, error) {
	m.Lock()
	defer m.Unlock()

	m.JoinLobbyCalls = append(m.JoinLobbyCalls, GLMJoinLobbyCall{
		LobbyID:      lobbyID,
		PlayerName:   playerName,
		PlayerAvatar: playerAvatar,
	})

	if len(m.JoinLobbyResults) > 0 {
		result := m.JoinLobbyResults[0]
		if len(m.JoinLobbyResults) > 1 {
			m.JoinLobbyResults = m.JoinLobbyResults[1:]
		}
		return result.PlayerID, result.SessionToken, result.Error
	}

	return "", "", nil
}

func (m *MockGameLifecycleManager) JoinLobbyWithActor(lobbyID string, playerActor interfaces.PlayerActorInterface) error {
	m.Lock()
	defer m.Unlock()

	m.JoinLobbyWithActorCalls = append(m.JoinLobbyWithActorCalls, GLMJoinLobbyWithActorCall{
		LobbyID:     lobbyID,
		PlayerActor: playerActor,
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

func (m *MockGameLifecycleManager) StartGame(lobbyID string, hostPlayerID string) error {
	defer m.Wg.Done() // Signal that this function has completed

	m.Lock()
	defer m.Unlock()

	m.StartGameCalls = append(m.StartGameCalls, GLMStartGameCall{
		LobbyID:      lobbyID,
		HostPlayerID: hostPlayerID,
	})

	if len(m.StartGameResults) > 0 {
		result := m.StartGameResults[0]
		if len(m.StartGameResults) > 1 {
			m.StartGameResults = m.StartGameResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockGameLifecycleManager) ValidateSessionToken(token string) (interface{}, error) {
	m.Lock()
	defer m.Unlock()

	m.ValidateSessionTokenCalls = append(m.ValidateSessionTokenCalls, GLMValidateSessionTokenCall{
		Token: token,
	})

	if len(m.ValidateSessionTokenResults) > 0 {
		result := m.ValidateSessionTokenResults[0]
		if len(m.ValidateSessionTokenResults) > 1 {
			m.ValidateSessionTokenResults = m.ValidateSessionTokenResults[1:]
		}
		return result.TokenInfo, result.Error
	}

	return nil, nil
}

func (m *MockGameLifecycleManager) SendActionToGame(gameID string, action core.Action) error {
	m.Lock()
	defer m.Unlock()

	m.SendActionToGameCalls = append(m.SendActionToGameCalls, GLMSendActionToGameCall{
		GameID: gameID,
		Action: action,
	})

	if len(m.SendActionToGameResults) > 0 {
		result := m.SendActionToGameResults[0]
		if len(m.SendActionToGameResults) > 1 {
			m.SendActionToGameResults = m.SendActionToGameResults[1:]
		}
		return result
	}

	return nil
}

func (m *MockGameLifecycleManager) Stop() {
	m.Lock()
	defer m.Unlock()

	m.StopCalls = append(m.StopCalls, GLMStopCall{})
}