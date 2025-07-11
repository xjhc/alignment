package mcp

import (
	"fmt"
	"log"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/interfaces"
)

// Define schemas for the tools.
var (
	sendChatMessageTool = Tool{
		Name:        "send_chat_message",
		Description: "Sends a message to the public in-game chat.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"game_id": map[string]string{"type": "string", "description": "The ID of the game to send the message to."},
				"message": map[string]string{"type": "string", "description": "The content of the chat message."},
			},
			"required": []string{"game_id", "message"},
		},
	}

	// --- Placeholders for future RulesEngine tools ---
	// voteForPlayerTool = Tool{ ... }
	// useAbilityTool = Tool{ ... }
)

// handleListTools responds to a `tools/list` request.
func handleListTools(req Request) Response {
	// For now, only expose the chat tool.
	// The RulesEngine tools will be added here later.
	tools := []Tool{sendChatMessageTool}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ListToolsResult{Tools: tools},
	}
}

// handleCallTool routes a `tools/call` request to the correct handler.
func handleCallTool(req Request, glm interfaces.GameLifecycleManagerInterface, aiPlayerID string) Response {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return createErrorResponse(req.ID, InvalidParams, "Invalid params format", nil)
	}

	toolName, _ := params["name"].(string)
	args, _ := params["arguments"].(map[string]interface{})

	switch toolName {
	case "send_chat_message":
		return callSendChatMessage(req.ID, args, glm, aiPlayerID)
	// --- Placeholders for future RulesEngine tools ---
	// case "vote_for_player":
	// 	return callVoteForPlayer(req.ID, args, glm, aiPlayerID)
	// case "use_ability":
	//  return callUseAbility(req.ID, args, glm, aiPlayerID)
	default:
		return createErrorResponse(req.ID, MethodNotFound, fmt.Sprintf("Tool '%s' not found", toolName), nil)
	}
}

// callSendChatMessage executes the logic for the chat tool.
func callSendChatMessage(id interface{}, args map[string]interface{}, glm interfaces.GameLifecycleManagerInterface, aiPlayerID string) Response {
	gameID, ok := args["game_id"].(string)
	if !ok {
		return createErrorResponse(id, InvalidParams, "Missing or invalid 'game_id' argument", nil)
	}
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return createErrorResponse(id, InvalidParams, "Missing or invalid 'message' argument", nil)
	}

	action := core.Action{
		Type:     core.ActionSendMessage,
		PlayerID: aiPlayerID, // The AI player is the one sending the message
		GameID:   gameID,
		Payload:  map[string]interface{}{"content": message},
	}

	resultChan, err := glm.SendActionToGame(gameID, action)
	if err != nil {
		log.Printf("MCP: Error sending chat action to game: %v", err)
		return createErrorResponse(id, InternalError, "Failed to send message to game", err.Error())
	}

	// Wait for the result (MCP operations should be synchronous)
	select {
	case result := <-resultChan:
		if result.Error != nil {
			log.Printf("MCP: Chat action rejected by game: %v", result.Error)
			return createErrorResponse(id, InternalError, "Message rejected by game", result.Error.Error())
		}
		// Broadcast the events if successful
		if len(result.Events) > 0 {
			if err := glm.BroadcastEventsToGame(gameID, result.Events); err != nil {
				log.Printf("MCP: Failed to broadcast events: %v", err)
			}
		}
	case <-time.After(5 * time.Second):
		log.Printf("MCP: Chat action timed out")
		return createErrorResponse(id, InternalError, "Message timed out", "The server is busy")
	}

	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Result: CallToolResult{
			Content: []ToolResultContent{{Type: "text", Text: "Message sent successfully."}},
		},
	}
}