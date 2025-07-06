
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/xjhc/alignment/core"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path_to_game_state.json>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read game state file: %v", err)
	}

	var gameState core.GameState
	if err := json.Unmarshal(data, &gameState); err != nil {
		log.Fatalf("Failed to parse game state JSON: %v", err)
	}

	visualize(gameState)
}

func visualize(state core.GameState) {
	fmt.Printf("===== Game State: %s =====\n", state.ID)
	fmt.Printf("Day: %d | Phase: %s\n", state.DayNumber, state.Phase.Type)

	if state.WinCondition != nil {
		fmt.Printf("\n--- WIN CONDITION MET ---\n")
		fmt.Printf("Winner: %s\n", state.WinCondition.Winner)
		fmt.Printf("Reason: %s\n", state.WinCondition.Description)
		fmt.Println("-------------------------")
	}

	fmt.Println("\n--- Players ---")
	for _, player := range state.Players {
		status := "ALIVE"
		if !player.IsAlive {
			status = "ELIMINATED"
		}

		role := "Unknown"
		if player.Role != nil {
			role = string(player.Role.Type)
		}

		fmt.Printf("  - %s (%s): %s | Tokens: %d | Milestones: %d | Alignment: %s\n",
			player.Name, role, status, player.Tokens, player.ProjectMilestones, player.Alignment)

		if player.PersonalKPI != nil {
			fmt.Printf("    KPI: %s (%d/%d)\n", player.PersonalKPI.Description, player.PersonalKPI.Progress, player.PersonalKPI.Target)
		}
	}

	if state.CrisisEvent != nil {
		fmt.Println("\n--- Active Crisis ---")
		fmt.Printf("  Title: %s\n", state.CrisisEvent.Title)
		fmt.Printf("  Description: %s\n", state.CrisisEvent.Description)
	}

	if state.VoteState != nil {
		fmt.Println("\n--- Active Vote ---")
		fmt.Printf("  Type: %s\n", state.VoteState.Type)
		fmt.Printf("  Results: %+v\n", state.VoteState.Results)
	}

	fmt.Println("\n--- Chat History ---")
	for i, msg := range state.ChatMessages {
		if i > 10 { // Limit to last 10 messages
			continue
		}
		fmt.Printf("  [%s] %s: %s\n", msg.Timestamp.Format("15:04:05"), msg.PlayerName, msg.Message)
	}
}