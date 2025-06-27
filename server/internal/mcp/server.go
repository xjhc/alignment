package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/xjhc/alignment/server/internal/interfaces"
)

// McpServer handles the Model Context Protocol communication.
type McpServer struct {
	glm interfaces.GameLifecycleManagerInterface
}

// NewMcpServer creates a new MCP server instance.
func NewMcpServer(glm interfaces.GameLifecycleManagerInterface) *McpServer {
	return &McpServer{
		glm: glm,
	}
}

// Run starts the server, listening on stdin for requests.
func (s *McpServer) Run() {
	log.Println("MCP Server: Starting on stdio...")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			log.Printf("MCP: Failed to unmarshal request: %v. Line: %s", err, line)
			continue
		}

		response := s.handleRequest(req)
		s.sendResponse(response)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("MCP: Error reading from stdin: %v", err)
	}
}

// handleRequest routes an incoming MCP request to the appropriate handler.
func (s *McpServer) handleRequest(req Request) Response {
	log.Printf("MCP: Received request method: %s", req.Method)
	// For now, we assume a single AI player with a hardcoded ID for simplicity.
	// A real implementation would manage sessions and AI player identities.
	aiPlayerID := "ai-player-mcp"

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "resources/list":
		return handleListResources(req)
	case "resources/read":
		return handleReadResource(req, s.glm, aiPlayerID)
	case "tools/list":
		return handleListTools(req)
	case "tools/call":
		return handleCallTool(req, s.glm, aiPlayerID)
	case "ping":
		return Response{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}
	default:
		if strings.HasPrefix(req.Method, "notifications/") {
			// Acknowledge notifications but do nothing
			return Response{}
		}
		log.Printf("MCP: Unknown method: %s", req.Method)
		return createErrorResponse(req.ID, MethodNotFound, fmt.Sprintf("Method '%s' not found", req.Method), nil)
	}
}

// handleInitialize responds to the client's initial handshake.
func (s *McpServer) handleInitialize(req Request) Response {
	// We don't need to parse params for this simple server.
	// Just respond with our capabilities.
	result := InitializeResult{
		ProtocolVersion: "2025-06-18",
		ServerInfo: ServerInfo{
			Name:    "alignment-mcp-server",
			Version: "0.1.0",
		},
		Capabilities: ServerCapabilities{
			Resources: map[string]interface{}{
				"listChanged": false,
				"subscribe":   false,
			},
			Tools: map[string]interface{}{
				"listChanged": false,
			},
		},
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// sendResponse marshals a response to JSON and writes it to stdout.
func (s *McpServer) sendResponse(resp Response) {
	if resp.ID == nil {
		// This is a response to a notification, do not send anything.
		return
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Printf("MCP: Failed to marshal response: %v", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(respBytes))
}

// createErrorResponse is a helper to build a JSON-RPC error response.
func createErrorResponse(id interface{}, code int, message string, data interface{}) Response {
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
}