package mcp

import (
	"log"
	"strings"

	"github.com/xjhc/alignment/server/internal/interfaces"
)

// handleListResources responds to a `resources/list` request.
func handleListResources(req Request) Response {
	templates := []Resource{
		{
			URITemplate: "game://alignment/{game_id}",
			Name:        "Alignment Game State",
			Description: "Provides a read-only, public view of the game state for a given game ID.",
		},
	}
	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ListResourcesResult{ResourceTemplates: templates},
	}
}

// handleReadResource responds to a `resources/read` request.
func handleReadResource(req Request, glm interfaces.GameLifecycleManagerInterface, aiPlayerID string) Response {
	params, ok := req.Params.(map[string]interface{})
	if !ok {
		return createErrorResponse(req.ID, InvalidParams, "Invalid params format", nil)
	}
	uri, _ := params["uri"].(string)
	if !strings.HasPrefix(uri, "game://alignment/") {
		return createErrorResponse(req.ID, InvalidParams, "Invalid resource URI", nil)
	}

	gameID := strings.TrimPrefix(uri, "game://alignment/")

	// Use the GameLifecycleManager to get the GameActor
	gameActor, exists := glm.GetGameActor(gameID)
	if !exists {
		log.Printf("MCP: Attempted to read resource for non-existent game: %s", gameID)
		return createErrorResponse(req.ID, InternalError, "Game not found", nil)
	}

	// Get the current state from the actor
	fullGameState := gameActor.GetGameState()
	if fullGameState == nil {
		log.Printf("MCP: Game actor for %s returned nil state", gameID)
		return createErrorResponse(req.ID, InternalError, "Could not retrieve game state", nil)
	}

	// Filter the state to only include public information
	publicState := FilterGameStateForAI(fullGameState, aiPlayerID)

	content := ResourceContent{
		URI:  uri,
		Text: publicState,
	}

	return Response{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ReadResourceResult{Contents: []ResourceContent{content}},
	}
}