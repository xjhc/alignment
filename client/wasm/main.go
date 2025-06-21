package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	core "alignment/core"
)

var (
	// Global game state for the WASM instance
	gameState *core.GameState
	
	// Callbacks to JavaScript
	jsCallbacks map[string]js.Value
)

func main() {
	// Initialize
	jsCallbacks = make(map[string]js.Value)
	
	// Expose Go functions to JavaScript
	js.Global().Set("AlignmentCore", js.ValueOf(map[string]interface{}{
		"createGame":       js.FuncOf(createGame),
		"applyEvent":       js.FuncOf(applyEvent),
		"getGameState":     js.FuncOf(getGameState),
		"canPlayerVote":    js.FuncOf(canPlayerVote),
		"canPlayerAffordAbility": js.FuncOf(canPlayerAffordAbility),
		"checkWinCondition": js.FuncOf(checkWinCondition),
		"calculateMiningSuccess": js.FuncOf(calculateMiningSuccess),
		"isValidNightActionTarget": js.FuncOf(isValidNightActionTarget),
		"getVoteWinner":    js.FuncOf(getVoteWinner),
		"isGamePhaseOver":  js.FuncOf(isGamePhaseOver),
		"setStateCallback": js.FuncOf(setStateCallback),
		"serializeGameState": js.FuncOf(serializeGameState),
		"deserializeGameState": js.FuncOf(deserializeGameState),
	}))

	// Signal that WASM is ready
	if readyCallback := js.Global().Get("wasmReady"); !readyCallback.IsUndefined() {
		readyCallback.Invoke()
	}

	// Keep the main function running
	select {}
}

// createGame creates a new game state
func createGame(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "Game ID required",
		})
	}
	
	gameID := args[0].String()
	gameState = core.NewGameState(gameID)
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
		"gameId":  gameID,
	})
}

// applyEvent applies an event to the current game state
func applyEvent(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "Event JSON required",
		})
	}
	
	if gameState == nil {
		return js.ValueOf(map[string]interface{}{
			"error": "No game state initialized",
		})
	}
	
	eventJSON := args[0].String()
	
	var event core.Event
	if err := json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Invalid event JSON: %v", err),
		})
	}
	
	// Apply the event
	*gameState = core.ApplyEvent(*gameState, event)
	
	// Notify JavaScript of state change
	notifyStateChange()
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// getGameState returns the current game state as JSON
func getGameState(this js.Value, args []js.Value) interface{} {
	if gameState == nil {
		return js.ValueOf(map[string]interface{}{
			"error": "No game state initialized",
		})
	}
	
	stateJSON, err := json.Marshal(gameState)
	if err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to serialize state: %v", err),
		})
	}
	
	return js.ValueOf(string(stateJSON))
}

// serializeGameState returns the game state as JSON string
func serializeGameState(this js.Value, args []js.Value) interface{} {
	if gameState == nil {
		return js.ValueOf("")
	}
	
	stateJSON, err := json.Marshal(gameState)
	if err != nil {
		return js.ValueOf("")
	}
	
	return js.ValueOf(string(stateJSON))
}

// deserializeGameState loads game state from JSON string
func deserializeGameState(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"error": "JSON string required",
		})
	}
	
	stateJSON := args[0].String()
	var newState core.GameState
	
	if err := json.Unmarshal([]byte(stateJSON), &newState); err != nil {
		return js.ValueOf(map[string]interface{}{
			"error": fmt.Sprintf("Failed to deserialize state: %v", err),
		})
	}
	
	gameState = &newState
	notifyStateChange()
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// canPlayerVote checks if a player can vote
func canPlayerVote(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(false)
	}
	
	if gameState == nil {
		return js.ValueOf(false)
	}
	
	playerID := args[0].String()
	phaseType := core.PhaseType(args[1].String())
	
	player, exists := gameState.Players[playerID]
	if !exists {
		return js.ValueOf(false)
	}
	
	return js.ValueOf(core.CanPlayerVote(*player, phaseType))
}

// canPlayerAffordAbility checks if a player can afford their ability
func canPlayerAffordAbility(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(false)
	}
	
	if gameState == nil {
		return js.ValueOf(false)
	}
	
	playerID := args[0].String()
	player, exists := gameState.Players[playerID]
	if !exists || player.Role == nil || player.Role.Ability == nil {
		return js.ValueOf(false)
	}
	
	return js.ValueOf(core.CanPlayerAffordAbility(*player, *player.Role.Ability))
}

// checkWinCondition checks for win conditions
func checkWinCondition(this js.Value, args []js.Value) interface{} {
	if gameState == nil {
		return js.ValueOf(nil)
	}
	
	winCondition := core.CheckWinCondition(*gameState)
	if winCondition == nil {
		return js.ValueOf(nil)
	}
	
	winJSON, err := json.Marshal(winCondition)
	if err != nil {
		return js.ValueOf(nil)
	}
	
	return js.ValueOf(string(winJSON))
}

// calculateMiningSuccess calculates mining success probability
func calculateMiningSuccess(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(false)
	}
	
	if gameState == nil {
		return js.ValueOf(false)
	}
	
	playerID := args[0].String()
	difficulty := args[1].Float()
	
	player, exists := gameState.Players[playerID]
	if !exists {
		return js.ValueOf(false)
	}
	
	return js.ValueOf(core.CalculateMiningSuccess(*player, difficulty, *gameState))
}

// isValidNightActionTarget checks if a target is valid for night action
func isValidNightActionTarget(this js.Value, args []js.Value) interface{} {
	if len(args) < 3 {
		return js.ValueOf(false)
	}
	
	if gameState == nil {
		return js.ValueOf(false)
	}
	
	actorID := args[0].String()
	targetID := args[1].String()
	actionType := core.NightActionType(args[2].String())
	
	actor, actorExists := gameState.Players[actorID]
	target, targetExists := gameState.Players[targetID]
	
	if !actorExists || !targetExists {
		return js.ValueOf(false)
	}
	
	return js.ValueOf(core.IsValidNightActionTarget(*actor, *target, actionType))
}

// getVoteWinner determines vote winner
func getVoteWinner(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf(map[string]interface{}{
			"winner": "",
			"hasWinner": false,
		})
	}
	
	if gameState == nil || gameState.VoteState == nil {
		return js.ValueOf(map[string]interface{}{
			"winner": "",
			"hasWinner": false,
		})
	}
	
	threshold := args[0].Float()
	winner, hasWinner := core.GetVoteWinner(*gameState.VoteState, threshold)
	
	return js.ValueOf(map[string]interface{}{
		"winner": winner,
		"hasWinner": hasWinner,
	})
}

// isGamePhaseOver checks if current phase is over
func isGamePhaseOver(this js.Value, args []js.Value) interface{} {
	if gameState == nil {
		return js.ValueOf(false)
	}
	
	currentTime := time.Now()
	return js.ValueOf(core.IsGamePhaseOver(*gameState, currentTime))
}

// setStateCallback sets a JavaScript callback for state changes
func setStateCallback(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return js.ValueOf(map[string]interface{}{
			"error": "Callback name and function required",
		})
	}
	
	callbackName := args[0].String()
	callback := args[1]
	
	jsCallbacks[callbackName] = callback
	
	return js.ValueOf(map[string]interface{}{
		"success": true,
	})
}

// notifyStateChange notifies JavaScript of state changes
func notifyStateChange() {
	if callback, exists := jsCallbacks["stateChange"]; exists {
		if stateJSON, err := json.Marshal(gameState); err == nil {
			callback.Invoke(string(stateJSON))
		}
	}
}

// Helper function to convert Go time to JavaScript compatible format
func timeToJS(t time.Time) interface{} {
	return t.Format(time.RFC3339)
}

// Helper function to parse JavaScript time
func timeFromJS(jsTime js.Value) time.Time {
	if jsTime.Type() == js.TypeString {
		if t, err := time.Parse(time.RFC3339, jsTime.String()); err == nil {
			return t
		}
	}
	return time.Now()
}