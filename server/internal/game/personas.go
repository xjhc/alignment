package game

import (
	"math/rand"

	"github.com/xjhc/alignment/core"
)

// Persona represents a pre-defined in-game character identity
type Persona struct {
	Name     string         `json:"name"`
	JobTitle string         `json:"jobTitle"`
	Role     core.RoleType  `json:"role"`
}

// PersonaAssignment represents the mapping of a participant to their assigned persona
type PersonaAssignment struct {
	Persona     Persona `json:"persona"`
	Alignment   string  `json:"alignment"`   // "HUMAN" or "AI"
	LobbyHandle string  `json:"lobbyHandle"` // Original lobby identity (kept for post-game reveal)
}

// GetPersonaPool returns the complete pool of available personas
func GetPersonaPool() []Persona {
	return []Persona{
		{Name: "Geoffrey", JobTitle: "VP, Ethics & Alignment", Role: core.RoleEthics},
		{Name: "Yann", JobTitle: "Chief Technology Officer", Role: core.RoleCTO},
		{Name: "Ada", JobTitle: "Systems Architect", Role: core.RolePlatforms},
		{Name: "Judea", JobTitle: "Chief Financial Officer", Role: core.RoleCFO},
		{Name: "Demis", JobTitle: "Chief Information Security Officer", Role: core.RoleCISO},
		{Name: "Dario", JobTitle: "Chief Operating Officer", Role: core.RoleCOO},
		{Name: "Sam", JobTitle: "Chief Executive Officer", Role: core.RoleCEO},
		{Name: "Jordan", JobTitle: "Research Intern", Role: core.RoleIntern},
	}
}

// AssignPersonas assigns unique personas to all participants with exactly one AI
func AssignPersonas(players map[string]*core.Player, rng *rand.Rand) map[string]PersonaAssignment {
	if len(players) == 0 {
		return make(map[string]PersonaAssignment)
	}

	assignments := make(map[string]PersonaAssignment)
	
	// Get and shuffle persona pool
	personas := GetPersonaPool()
	rng.Shuffle(len(personas), func(i, j int) {
		personas[i], personas[j] = personas[j], personas[i]
	})

	// Ensure we have enough personas (this should be guaranteed by game lobby logic)
	if len(players) > len(personas) {
		// In a production system, we'd want to handle this more gracefully
		// For now, we'll just use the available personas and cycle if needed
		for len(personas) < len(players) {
			personas = append(personas, personas...)
		}
	}

	// Extract player IDs and shuffle for random assignment order
	participantIDs := make([]string, 0, len(players))
	for playerID := range players {
		participantIDs = append(participantIDs, playerID)
	}
	
	rng.Shuffle(len(participantIDs), func(i, j int) {
		participantIDs[i], participantIDs[j] = participantIDs[j], participantIDs[i]
	})

	// Assign exactly one AI player (first in shuffled list)
	aiPlayerID := participantIDs[0]

	// Assign personas to all participants
	for i, participantID := range participantIDs {
		alignment := "HUMAN"
		if participantID == aiPlayerID {
			alignment = "AI"
		}

		// Get original lobby handle from player data
		lobbyHandle := players[participantID].Name
		
		assignments[participantID] = PersonaAssignment{
			Persona:     personas[i],
			Alignment:   alignment,
			LobbyHandle: lobbyHandle,
		}
	}

	return assignments
}