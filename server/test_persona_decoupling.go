package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/game"
)

func main() {
	fmt.Println("=== Demonstrating Decoupled Persona Generation ===")

	// Create test players
	players := make(map[string]*core.Player)
	playerNames := []string{"Alice", "Bob", "Charlie", "Diana", "Eve", "Frank", "Grace", "Henry"}
	
	for i, name := range playerNames {
		playerID := fmt.Sprintf("player_%d", i+1)
		players[playerID] = &core.Player{
			ID:   playerID,
			Name: name,
		}
	}

	// Game settings
	settings := core.GameSettings{
		InitialAlignedHumanCount: 1,
	}

	fmt.Printf("Testing with %d players and %d aligned humans\n\n", len(players), settings.InitialAlignedHumanCount)

	// Run multiple assignments to show randomization
	for run := 1; run <= 3; run++ {
		fmt.Printf("--- Game %d ---\n", run)
		
		// Use different seed for each run to show variety
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(run)))
		assignments := game.AssignPersonas(players, settings, rng)

		// Display assignments
		for playerID, assignment := range assignments {
			originalName := players[playerID].Name
			fmt.Printf("  %s (lobby: %s) → %s, %s [%s]\n",
				assignment.Persona.Name,
				originalName,
				assignment.Persona.JobTitle,
				assignment.Persona.Role,
				assignment.Alignment)
		}
		fmt.Println()
	}

	fmt.Println("=== Demonstrating Name and Role Pools ===")

	namePool := game.GetNamePool()
	rolePool := game.GetRolePool()

	fmt.Printf("Available Names (%d):\n", len(namePool))
	for i, name := range namePool {
		if i > 0 && i%8 == 0 {
			fmt.Println()
		}
		fmt.Printf("  %-10s", name)
	}
	fmt.Println()

	fmt.Printf("Available Roles (%d):\n", len(rolePool))
	for role, jobTitle := range rolePool {
		fmt.Printf("  %-10s → %s\n", role, jobTitle)
	}

	fmt.Println("\n=== Key Benefits ===")
	fmt.Println("✓ Names and roles are completely decoupled")
	fmt.Println("✓ Each game has unique, unpredictable combinations")
	fmt.Println("✓ Players can't learn name-role associations")
	fmt.Println("✓ Maximum replayability and strategic depth")
	fmt.Println("✓ Preserves original lobby handles for post-game reveal")
}