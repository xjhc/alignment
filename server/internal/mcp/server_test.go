package mcp

import (
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
	"github.com/xjhc/alignment/server/internal/mocks"
)

// MockGameActor for testing McpServer
type MockGameActor struct {
	GameState *core.GameState
}

func (m *MockGameActor) GetGameID() string {
	return m.GameState.ID
}
func (m *MockGameActor) PostAction(action core.Action) chan interfaces.ProcessActionResult {
	// For testing, immediately return a successful result
	ch := make(chan interfaces.ProcessActionResult, 1)
	ch <- interfaces.ProcessActionResult{Error: nil}
	return ch
}
func (m *MockGameActor) GetGameState() *core.GameState {
	return m.GameState
}
func (m *MockGameActor) CreatePlayerStateUpdateEvent(playerID string) core.Event {
	return core.Event{}
}
func (m *MockGameActor) Stop() {}

// setupTestServer creates an McpServer with mock dependencies for testing.
func setupTestServer(t *testing.T) (*McpServer, *mocks.MockGameLifecycleManager) {
	mockGLM := &mocks.MockGameLifecycleManager{}
	server := NewMcpServer(mockGLM)
	return server, mockGLM
}

func TestMcpServer_handleInitialize(t *testing.T) {
	server, _ := setupTestServer(t)
	req := Request{
		ID:     1,
		Method: "initialize",
		Params: InitializeParams{ProtocolVersion: "2025-06-18"},
	}

	resp := server.handleRequest(req)
	require.Nil(t, resp.Error)
	require.NotNil(t, resp.Result)

	result, ok := resp.Result.(InitializeResult)
	require.True(t, ok)

	assert.Equal(t, "2025-06-18", result.ProtocolVersion)
	assert.Equal(t, "alignment-mcp-server", result.ServerInfo.Name)
	assert.NotNil(t, result.Capabilities.Resources)
	assert.NotNil(t, result.Capabilities.Tools)
}

func TestMcpServer_handleListResources(t *testing.T) {
	req := Request{ID: 1, Method: "resources/list"}
	resp := handleListResources(req)

	require.Nil(t, resp.Error)
	result, ok := resp.Result.(ListResourcesResult)
	require.True(t, ok)
	assert.Len(t, result.ResourceTemplates, 1)
	assert.Equal(t, "game://alignment/{game_id}", result.ResourceTemplates[0].URITemplate)
}

func TestMcpServer_handleReadResource(t *testing.T) {
	server, mockGLM := setupTestServer(t)

	// Setup mock game actor and state
	gameID := "test-game"
	mockActor := &MockGameActor{
		GameState: &core.GameState{
			ID: gameID,
			Players: map[string]*core.Player{
				"player1": {ID: "player1", Name: "Alice", Alignment: "HUMAN"},
				"ai-player-mcp": {ID: "ai-player-mcp", Name: "MCP_AI", Alignment: "AI"},
			},
		},
	}
	mockGLM.GetActorResults = []mocks.GetActorResult{{Actor: mockActor, Found: true}}

	req := Request{
		ID:     1,
		Method: "resources/read",
		Params: map[string]interface{}{"uri": "game://alignment/test-game"},
	}

	resp := server.handleRequest(req)
	require.Nil(t, resp.Error)
	result, ok := resp.Result.(ReadResourceResult)
	require.True(t, ok)
	assert.Len(t, result.Contents, 1)

	content, ok := result.Contents[0].Text.(PublicGameState)
	require.True(t, ok)

	assert.Equal(t, "ai-player-mcp", content.YourPlayerID)
	assert.Len(t, content.Players, 2)
}

func TestMcpServer_handleListTools(t *testing.T) {
	req := Request{ID: 1, Method: "tools/list"}
	resp := handleListTools(req)

	require.Nil(t, resp.Error)
	result, ok := resp.Result.(ListToolsResult)
	require.True(t, ok)
	assert.Len(t, result.Tools, 1)
	assert.Equal(t, "send_chat_message", result.Tools[0].Name)
}

func TestMcpServer_handleCallTool_SendChatMessage(t *testing.T) {
	server, mockGLM := setupTestServer(t)

	req := Request{
		ID:     1,
		Method: "tools/call",
		Params: map[string]interface{}{
			"name": "send_chat_message",
			"arguments": map[string]interface{}{
				"game_id": "test-game",
				"message": "Hello from the AI!",
			},
		},
	}

	// This is the important part: setting up the expectation.
	// We expect SendActionToGame to be called once, and it should return no error.
	mockGLM.SendActionToGameResults = []error{nil}

	resp := server.handleRequest(req)
	require.Nil(t, resp.Error)
	result, ok := resp.Result.(CallToolResult)
	require.True(t, ok)

	// Verify that the manager's method was called correctly
	require.Len(t, mockGLM.SendActionToGameCalls, 1)
	call := mockGLM.SendActionToGameCalls[0]
	assert.Equal(t, "test-game", call.GameID)
	assert.Equal(t, core.ActionSendMessage, call.Action.Type)
	assert.Equal(t, "ai-player-mcp", call.Action.PlayerID)
	assert.Equal(t, "Hello from the AI!", call.Action.Payload["content"])
	assert.Equal(t, "Message sent successfully.", result.Content[0].Text)
}

func TestMcpServer_handleCallTool_NotFound(t *testing.T) {
	server, _ := setupTestServer(t)
	req := Request{
		ID:     1,
		Method: "tools/call",
		Params: map[string]interface{}{"name": "non_existent_tool"},
	}

	resp := server.handleRequest(req)
	require.NotNil(t, resp.Error)
	assert.Equal(t, MethodNotFound, resp.Error.Code)
}

// TestMcpServer_Run simulates the stdin/stdout interaction.
func TestMcpServer_Run(t *testing.T) {
	server, mockGLM := setupTestServer(t)

	// --- Mock the Stdio environment ---
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdin = r
	outReader, outWriter, _ := os.Pipe()
	os.Stdout = outWriter

	// --- Run the server in a goroutine ---
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Run()
	}()

	// --- Write a request to the pipe (server's stdin) ---
	req := Request{
		ID:     123,
		Method: "initialize",
		Params: InitializeParams{ProtocolVersion: "2025-06-18"},
	}
	reqBytes, _ := json.Marshal(req)
	_, err := w.Write(append(reqBytes, '\n'))
	require.NoError(t, err)

	// --- Read the response from the pipe (server's stdout) ---
	responseBytes, err := readWithTimeout(outReader, 1*time.Second)
	require.NoError(t, err)

	var resp Response
	err = json.Unmarshal(responseBytes, &resp)
	require.NoError(t, err)
	assert.Equal(t, req.ID, resp.ID)
	assert.Nil(t, resp.Error)

	// Clean up: Close the write pipe to signal EOF to the server's scanner.
	w.Close()
	wg.Wait() // Wait for the server goroutine to finish.
	outReader.Close()
	outWriter.Close()
}

// readWithTimeout reads from a reader with a timeout.
func readWithTimeout(r io.Reader, timeout time.Duration) ([]byte, error) {
	dataChan := make(chan []byte)
	errChan := make(chan error)

	go func() {
		reader := bufio.NewReader(r)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			errChan <- err
			return
		}
		dataChan <- line
	}()

	select {
	case data := <-dataChan:
		return data, nil
	case err := <-errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("read timed out")
	}
}