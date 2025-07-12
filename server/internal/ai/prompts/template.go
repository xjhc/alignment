package prompts

import (
	"fmt"
	
	"github.com/xjhc/alignment/core"
)

// PromptTemplate defines the interface for AI prompt templates
type PromptTemplate interface {
	// ID returns the unique identifier for this template
	ID() string
	
	// Name returns the human-readable name for this template
	Name() string
	
	// Description returns a description of this template's personality and approach
	Description() string
	
	// BuildPrompt constructs the final prompt string using the provided context
	BuildPrompt(ctx PromptContext) string
}

// PromptContext contains all the data needed to build an AI prompt
type PromptContext struct {
	// Game state information
	GameState    *core.GameState `json:"game_state"`
	AIPlayerID   string          `json:"ai_player_id"`
	CurrentPhase string          `json:"current_phase"`
	DayNumber    int             `json:"day_number"`
	
	// Player information (filtered for AI visibility)
	Players      []PlayerInfo    `json:"players"`
	AIPlayer     PlayerInfo      `json:"ai_player"`
	
	// Threat analysis from the Rules Engine
	ThreatMap    map[string]float64 `json:"threat_map"`
	TopThreats   []string           `json:"top_threats"`
	
	// Recent game events for context
	RecentEvents []core.Event       `json:"recent_events"`
	RecentChat   []core.ChatMessage `json:"recent_chat"`
	
	// Strategic context
	GameObjective string `json:"game_objective"` // "conversion", "elimination", "survival"
	RiskLevel     string `json:"risk_level"`     // "low", "medium", "high"
	
	// Conversation trigger
	TriggerEvent string `json:"trigger_event"` // "phase_change", "player_spoke", "random"
}

// PlayerInfo represents public information about a player for prompt building
type PlayerInfo struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	IsAlive           bool   `json:"is_alive"`
	Tokens            int    `json:"tokens"`
	ProjectMilestones int    `json:"project_milestones"`
	StatusMessage     string `json:"status_message"`
	IsNominated       bool   `json:"is_nominated"`
	ThreatLevel       float64 `json:"threat_level,omitempty"` // Only for non-AI players
}

// BaseTemplate provides common functionality for all prompt templates
type BaseTemplate struct {
	id          string
	name        string
	description string
}

// ID returns the template identifier
func (bt *BaseTemplate) ID() string {
	return bt.id
}

// Name returns the template name
func (bt *BaseTemplate) Name() string {
	return bt.name
}

// Description returns the template description
func (bt *BaseTemplate) Description() string {
	return bt.description
}

// BuildGameStateContext creates a formatted string describing the current game state
func BuildGameStateContext(ctx PromptContext) string {
	result := fmt.Sprintf("=== ALIGNMENT GAME STATE ===\n")
	result += fmt.Sprintf("Day %d - Phase: %s\n", ctx.DayNumber, ctx.CurrentPhase)
	result += fmt.Sprintf("Your Identity: %s (AI Player)\n\n", ctx.AIPlayer.Name)
	
	result += "=== PLAYERS ===\n"
	for _, player := range ctx.Players {
		status := "ALIVE"
		if !player.IsAlive {
			status = "ELIMINATED"
		}
		if player.IsNominated {
			status += " (NOMINATED)"
		}
		
		result += fmt.Sprintf("- %s: %d tokens, %d milestones [%s]\n", 
			player.Name, player.Tokens, player.ProjectMilestones, status)
		
		if player.StatusMessage != "" {
			result += fmt.Sprintf("  Status: %s\n", player.StatusMessage)
		}
	}
	
	if len(ctx.RecentChat) > 0 {
		result += "\n=== RECENT CONVERSATION ===\n"
		for _, msg := range ctx.RecentChat {
			if !msg.IsSystem {
				result += fmt.Sprintf("%s: %s\n", msg.PlayerName, msg.Message)
			}
		}
	}
	
	return result
}

// BuildThreatContext creates a formatted string describing threat analysis
func BuildThreatContext(ctx PromptContext) string {
	if len(ctx.TopThreats) == 0 {
		return "No significant threats identified.\n"
	}
	
	result := "=== THREAT ANALYSIS ===\n"
	for i, playerID := range ctx.TopThreats {
		if i >= 3 { // Limit to top 3 threats
			break
		}
		
		var playerName string
		for _, player := range ctx.Players {
			if player.ID == playerID {
				playerName = player.Name
				break
			}
		}
		
		threatLevel := ctx.ThreatMap[playerID]
		result += fmt.Sprintf("- %s: %.2f threat level\n", playerName, threatLevel)
	}
	
	return result
}

// Helper function to safely format strings
func SafeFormat(format string, args ...interface{}) string {
	defer func() {
		if r := recover(); r != nil {
			// If formatting fails, return the format string
			fmt.Printf("Prompt formatting error: %v\n", r)
		}
	}()
	return fmt.Sprintf(format, args...)
}