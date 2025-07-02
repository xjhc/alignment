package game

import (
	"math/rand"
	"testing"

	"github.com/xjhc/alignment/core"
)

func TestAssignPersonas_DefaultSettings(t *testing.T) {
	// Create test players
	players := make(map[string]*core.Player)
	for i := 0; i < 5; i++ {
		playerID := string(rune('A' + i))
		players[playerID] = &core.Player{
			ID:   playerID,
			Name: "Player" + playerID,
		}
	}

	// Create default settings (InitialAlignedHumanCount = 0)
	settings := core.GameSettings{
		InitialAlignedHumanCount: 0,
	}

	// Use deterministic random generator for consistent test results
	rng := rand.New(rand.NewSource(42))

	// Call AssignPersonas
	assignments := AssignPersonas(players, settings, rng)

	// Verify we got assignments for all players
	if len(assignments) != len(players) {
		t.Fatalf("Expected %d assignments, got %d", len(players), len(assignments))
	}

	// Count alignments
	var aiCount, alignedCount, humanCount int
	for _, assignment := range assignments {
		switch assignment.Alignment {
		case "AI":
			aiCount++
		case "ALIGNED":
			alignedCount++
		case "HUMAN":
			humanCount++
		}
	}

	// Verify there's exactly 1 AI, 0 aligned, and 4 humans
	if aiCount != 1 {
		t.Errorf("Expected 1 AI player, got %d", aiCount)
	}
	if alignedCount != 0 {
		t.Errorf("Expected 0 aligned players, got %d", alignedCount)
	}
	if humanCount != 4 {
		t.Errorf("Expected 4 human players, got %d", humanCount)
	}
}

func TestAssignPersonas_WithAlignedHumans(t *testing.T) {
	// Create test players
	players := make(map[string]*core.Player)
	for i := 0; i < 6; i++ {
		playerID := string(rune('A' + i))
		players[playerID] = &core.Player{
			ID:   playerID,
			Name: "Player" + playerID,
		}
	}

	// Create settings with 2 initial aligned humans
	settings := core.GameSettings{
		InitialAlignedHumanCount: 2,
	}

	// Use deterministic random generator for consistent test results
	rng := rand.New(rand.NewSource(42))

	// Call AssignPersonas
	assignments := AssignPersonas(players, settings, rng)

	// Verify we got assignments for all players
	if len(assignments) != len(players) {
		t.Fatalf("Expected %d assignments, got %d", len(players), len(assignments))
	}

	// Count alignments
	var aiCount, alignedCount, humanCount int
	for _, assignment := range assignments {
		switch assignment.Alignment {
		case "AI":
			aiCount++
		case "ALIGNED":
			alignedCount++
		case "HUMAN":
			humanCount++
		}
	}

	// Verify there's exactly 1 AI, 2 aligned, and 3 humans
	if aiCount != 1 {
		t.Errorf("Expected 1 AI player, got %d", aiCount)
	}
	if alignedCount != 2 {
		t.Errorf("Expected 2 aligned players, got %d", alignedCount)
	}
	if humanCount != 3 {
		t.Errorf("Expected 3 human players, got %d", humanCount)
	}
}

func TestAssignPersonas_ExcessiveAlignedHumans(t *testing.T) {
	// Create test players (only 3 players)
	players := make(map[string]*core.Player)
	for i := 0; i < 3; i++ {
		playerID := string(rune('A' + i))
		players[playerID] = &core.Player{
			ID:   playerID,
			Name: "Player" + playerID,
		}
	}

	// Try to set 5 aligned humans (more than available)
	settings := core.GameSettings{
		InitialAlignedHumanCount: 5,
	}

	// Use deterministic random generator for consistent test results
	rng := rand.New(rand.NewSource(42))

	// Call AssignPersonas
	assignments := AssignPersonas(players, settings, rng)

	// Count alignments
	var aiCount, alignedCount, humanCount int
	for _, assignment := range assignments {
		switch assignment.Alignment {
		case "AI":
			aiCount++
		case "ALIGNED":
			alignedCount++
		case "HUMAN":
			humanCount++
		}
	}

	// Should cap at maximum possible (1 AI + 2 aligned = 3 total players)
	if aiCount != 1 {
		t.Errorf("Expected 1 AI player, got %d", aiCount)
	}
	if alignedCount != 2 {
		t.Errorf("Expected 2 aligned players (capped), got %d", alignedCount)
	}
	if humanCount != 0 {
		t.Errorf("Expected 0 human players, got %d", humanCount)
	}
}

func TestAssignPersonas_SinglePlayer(t *testing.T) {
	// Create single test player
	players := make(map[string]*core.Player)
	players["A"] = &core.Player{
		ID:   "A",
		Name: "PlayerA",
	}

	// Create default settings
	settings := core.GameSettings{
		InitialAlignedHumanCount: 0,
	}

	// Use deterministic random generator
	rng := rand.New(rand.NewSource(42))

	// Call AssignPersonas
	assignments := AssignPersonas(players, settings, rng)

	// Verify single player gets AI role
	if len(assignments) != 1 {
		t.Fatalf("Expected 1 assignment, got %d", len(assignments))
	}

	assignment := assignments["A"]
	if assignment.Alignment != "AI" {
		t.Errorf("Expected single player to be AI, got %s", assignment.Alignment)
	}
}

func TestAssignPersonas_EmptyPlayers(t *testing.T) {
	// Empty players map
	players := make(map[string]*core.Player)

	// Create settings
	settings := core.GameSettings{
		InitialAlignedHumanCount: 0,
	}

	// Use deterministic random generator
	rng := rand.New(rand.NewSource(42))

	// Call AssignPersonas
	assignments := AssignPersonas(players, settings, rng)

	// Should return empty assignments
	if len(assignments) != 0 {
		t.Errorf("Expected 0 assignments for empty players, got %d", len(assignments))
	}
}

func TestAssignPersonas_PersonaAssignment(t *testing.T) {
	// Create test players
	players := make(map[string]*core.Player)
	for i := 0; i < 3; i++ {
		playerID := string(rune('A' + i))
		players[playerID] = &core.Player{
			ID:   playerID,
			Name: "Player" + playerID,
		}
	}

	// Create settings
	settings := core.GameSettings{
		InitialAlignedHumanCount: 1,
	}

	// Use deterministic random generator
	rng := rand.New(rand.NewSource(42))

	// Call AssignPersonas
	assignments := AssignPersonas(players, settings, rng)

	// Verify each player got a unique persona
	usedPersonas := make(map[string]bool)
	for _, assignment := range assignments {
		personaKey := assignment.Persona.Name + assignment.Persona.JobTitle
		if usedPersonas[personaKey] {
			t.Errorf("Duplicate persona assigned: %s - %s", assignment.Persona.Name, assignment.Persona.JobTitle)
		}
		usedPersonas[personaKey] = true

		// Verify persona has all required fields
		if assignment.Persona.Name == "" {
			t.Error("Persona name is empty")
		}
		if assignment.Persona.JobTitle == "" {
			t.Error("Persona job title is empty")
		}
		if assignment.Persona.Role == "" {
			t.Error("Persona role is empty")
		}

		// Verify lobby handle is preserved
		if assignment.LobbyHandle == "" {
			t.Error("Lobby handle is empty")
		}
	}
}

func TestAssignPersonas_DecoupledNamesAndRoles(t *testing.T) {
	// Create enough test players to verify randomization
	players := make(map[string]*core.Player)
	for i := 0; i < 8; i++ {
		playerID := string(rune('A' + i))
		players[playerID] = &core.Player{
			ID:   playerID,
			Name: "Player" + playerID,
		}
	}

	// Create settings with some aligned humans for variety
	settings := core.GameSettings{
		InitialAlignedHumanCount: 2,
	}

	// Run assignment multiple times with different seeds to verify randomization
	var allAssignments []map[string]PersonaAssignment
	for seed := int64(1); seed <= 5; seed++ {
		rng := rand.New(rand.NewSource(seed))
		assignments := AssignPersonas(players, settings, rng)
		allAssignments = append(allAssignments, assignments)
	}

	// Verify that names and roles are decoupled by checking if we see different combinations
	nameRoleCombinations := make(map[string]int)
	for _, assignments := range allAssignments {
		for _, assignment := range assignments {
			combination := assignment.Persona.Name + ":" + string(assignment.Persona.Role)
			nameRoleCombinations[combination]++
		}
	}

	// Since names and roles are shuffled independently, we should see varied combinations
	// With 8 players across 5 runs, we should see different name-role pairings
	if len(nameRoleCombinations) < 10 {
		t.Errorf("Expected varied name-role combinations indicating decoupling, got only %d unique combinations", len(nameRoleCombinations))
	}

	// Verify that each assignment has valid data
	for runIndex, assignments := range allAssignments {
		usedNames := make(map[string]bool)
		usedRoles := make(map[core.RoleType]bool)

		for _, assignment := range assignments {
			// Check for duplicate names within a single run
			if usedNames[assignment.Persona.Name] {
				t.Errorf("Run %d: Duplicate name assigned: %s", runIndex+1, assignment.Persona.Name)
			}
			usedNames[assignment.Persona.Name] = true

			// Check for duplicate roles within a single run
			if usedRoles[assignment.Persona.Role] {
				t.Errorf("Run %d: Duplicate role assigned: %s", runIndex+1, assignment.Persona.Role)
			}
			usedRoles[assignment.Persona.Role] = true

			// Verify job title matches role
			rolePool := GetRolePool()
			expectedJobTitle := rolePool[assignment.Persona.Role]
			if assignment.Persona.JobTitle != expectedJobTitle {
				t.Errorf("Run %d: Job title mismatch for role %s. Expected: %s, Got: %s", 
					runIndex+1, assignment.Persona.Role, expectedJobTitle, assignment.Persona.JobTitle)
			}

			// Verify name is from the name pool
			namePool := GetNamePool()
			nameFound := false
			for _, validName := range namePool {
				if assignment.Persona.Name == validName {
					nameFound = true
					break
				}
			}
			if !nameFound {
				t.Errorf("Run %d: Name %s not found in name pool", runIndex+1, assignment.Persona.Name)
			}
		}
	}
}