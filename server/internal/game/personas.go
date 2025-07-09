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

// GetNamePool returns the pool of available first names for random assignment
func GetNamePool() []string {
	return []string{
		"Ada", "Alan", "David", "Dario", "Demis", "Eliza", "Geoffrey", "Grace",
		"Jordan", "Judea", "Michael", "Sara", "Sam", "Yann", "Alex", "Blake",
		"Cameron", "Dana", "Evelyn", "Felix", "Harper", "Ian", "Jules", "Kelly",
		"Logan", "Morgan", "Nolan", "Parker", "Quinn", "River", "Sage", "Taylor",
	}
}

// GetRolePool returns the pool of available roles with their job titles
func GetRolePool() map[core.RoleType]string {
	return map[core.RoleType]string{
		core.RoleCISO:      "Chief Information Security Officer",
		core.RoleCTO:       "Chief Technology Officer",
		core.RoleCOO:       "Chief Operating Officer",
		core.RoleCFO:       "Chief Financial Officer",
		core.RoleCEO:       "Chief Executive Officer",
		core.RoleEthics:    "VP, Ethics & Alignment",
		core.RolePlatforms: "VP, Platforms",
		core.RoleIntern:    "Research Intern",
	}
}

// GetPersonaPool returns the complete pool of available personas (DEPRECATED)
// This function is maintained for backward compatibility but will be removed
// in favor of the new decoupled name/role assignment system
func GetPersonaPool() []Persona {
	return []Persona{
		{Name: "Geoffrey", JobTitle: "VP, Ethics & Alignment", Role: core.RoleEthics},
		{Name: "Yann", JobTitle: "Chief Technology Officer", Role: core.RoleCTO},
		{Name: "Ada", JobTitle: "VP, Platforms", Role: core.RolePlatforms},
		{Name: "Judea", JobTitle: "Chief Financial Officer", Role: core.RoleCFO},
		{Name: "Demis", JobTitle: "Chief Information Security Officer", Role: core.RoleCISO},
		{Name: "Dario", JobTitle: "Chief Operating Officer", Role: core.RoleCOO},
		{Name: "Sam", JobTitle: "Chief Executive Officer", Role: core.RoleCEO},
		{Name: "Jordan", JobTitle: "Research Intern", Role: core.RoleIntern},
	}
}

// AssignPersonas assigns unique personas to all participants with exactly one AI and configurable aligned humans
func AssignPersonas(players map[string]*core.Player, settings core.GameSettings, rng *rand.Rand) map[string]PersonaAssignment {
	if len(players) == 0 {
		return make(map[string]PersonaAssignment)
	}

	assignments := make(map[string]PersonaAssignment)

	// Get and shuffle name pool independently
	names := GetNamePool()
	rng.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})

	// Get role pool and create shuffled list of roles
	rolePool := GetRolePool()
	roles := make([]core.RoleType, 0, len(rolePool))
	for roleType := range rolePool {
		roles = append(roles, roleType)
	}
	rng.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})

	// Ensure we have enough names and roles (cycle if needed)
	for len(names) < len(players) {
		names = append(names, names...)
	}
	for len(roles) < len(players) {
		roles = append(roles, roles...)
	}

	// Extract player IDs and shuffle for random assignment order
	participantIDs := make([]string, 0, len(players))
	for playerID := range players {
		participantIDs = append(participantIDs, playerID)
	}

	rng.Shuffle(len(participantIDs), func(i, j int) {
		participantIDs[i], participantIDs[j] = participantIDs[j], participantIDs[i]
	})

	// Assign roles based on game settings
	var originalAIPlayerID string
	if settings.PlayAsAI {
		// In "Play as AI" mode, a random human player becomes the AI
		originalAIPlayerID = participantIDs[0]
	} else {
		// In normal mode, a separate AI entity is implied or would be added
		// For simulation, we'll assign the first player as AI if not in PlayAsAI mode
		// In a real game, this might be a dedicated AI bot
		originalAIPlayerID = participantIDs[0]
	}

	// Determine how many additional players should be Aligned humans
	alignedHumanCount := settings.InitialAlignedHumanCount
	// Ensure we don't exceed available players (minus the one Original AI)
	if alignedHumanCount > len(participantIDs)-1 {
		alignedHumanCount = len(participantIDs) - 1
	}

	// Assign aligned humans (next N players in shuffled list after the Original AI)
	alignedHumanIDs := make(map[string]bool)
	aiPlayerIndex := -1
	for i, id := range participantIDs {
		if id == originalAIPlayerID {
			aiPlayerIndex = i
			break
		}
	}

	for i := 1; i <= alignedHumanCount; i++ {
		// Wrap around the list to select players if we reach the end
		alignedIndex := (aiPlayerIndex + i) % len(participantIDs)
		alignedHumanIDs[participantIDs[alignedIndex]] = true
	}


	// Assign decoupled personas to all participants
	for i, participantID := range participantIDs {
		var alignment string
		if participantID == originalAIPlayerID {
			alignment = "AI" // Original AI
		} else if alignedHumanIDs[participantID] {
			alignment = "ALIGNED" // Aligned humans (converted to AI faction)
		} else {
			alignment = "HUMAN" // Regular humans
		}

		// Get original lobby handle from player data
		lobbyHandle := players[participantID].Name

		// Create persona with decoupled name and role
		assignedRole := roles[i]
		assignedName := names[i]
		assignedJobTitle := rolePool[assignedRole]

		assignments[participantID] = PersonaAssignment{
			Persona: Persona{
				Name:     assignedName,
				JobTitle: assignedJobTitle,
				Role:     assignedRole,
			},
			Alignment:   alignment,
			LobbyHandle: lobbyHandle,
		}
	}

	return assignments
}