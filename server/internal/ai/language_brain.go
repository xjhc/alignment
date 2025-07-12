package ai

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/ai/prompts"
	"github.com/xjhc/alignment/server/internal/llm"
)

// LanguageBrain handles AI communication using external LLMs
type LanguageBrain struct {
	promptTemplate prompts.PromptTemplate
	llmClient      LLMClient
	persona        StrategicPersona
}

// LLMClient interface for external LLM communication
type LLMClient interface {
	GenerateChat(ctx context.Context, prompt string) (string, error)
}

// NewLanguageBrain creates a new language brain
func NewLanguageBrain(persona StrategicPersona, client LLMClient) (*LanguageBrain, error) {
	template, err := prompts.GetByPersona(persona.ID)
	if err != nil {
		// Fallback to random template if persona-specific template not found
		template = prompts.GetRandom()
		if template == nil {
			return nil, fmt.Errorf("no prompt templates available")
		}
	}

	return &LanguageBrain{
		promptTemplate: template,
		llmClient:      client,
		persona:        persona,
	}, nil
}

// NewLanguageBrainWithDefaults creates a language brain with default settings
func NewLanguageBrainWithDefaults() (*LanguageBrain, error) {
	realClient, err := llm.NewLLMClient()
	var client LLMClient
	
	if err != nil {
		// If no real LLM client available, use mock for testing
		log.Printf("Warning: Using mock LLM client: %v", err)
		client = llm.NewMockLLMClient([]string{
			"I'm analyzing the current situation and looking for patterns.",
			"Based on the data I'm seeing, we should focus on efficiency.",
			"Let me think about this strategically for a moment.",
		})
	} else {
		client = realClient
	}

	return NewLanguageBrain(GetShadowPersona(), client)
}

// GenerateChatMessage generates a contextual chat message using LLM
func (lb *LanguageBrain) GenerateChatMessage(gameState *core.GameState, aiPlayerID string) (string, error) {
	// Build the prompt context
	ctx := lb.buildPromptContext(gameState, aiPlayerID)
	
	// Generate the final prompt
	prompt := lb.promptTemplate.BuildPrompt(ctx)
	
	// Call the LLM API
	response, err := lb.llmClient.GenerateChat(context.Background(), prompt)
	if err != nil {
		return "", fmt.Errorf("LLM API call failed: %w", err)
	}
	
	// Clean and validate the response
	cleanResponse := lb.cleanResponse(response)
	if cleanResponse == "" {
		return "", fmt.Errorf("LLM returned empty response")
	}
	
	return cleanResponse, nil
}

// buildPromptContext creates a comprehensive prompt context from the game state
func (lb *LanguageBrain) buildPromptContext(gameState *core.GameState, aiPlayerID string) prompts.PromptContext {
	// Build player info list
	players := make([]prompts.PlayerInfo, 0, len(gameState.Players))
	var aiPlayer prompts.PlayerInfo
	threatMap := make(map[string]float64)
	topThreats := make([]string, 0)
	
	// Create a simple threat assessment for prompt context
	for _, player := range gameState.Players {
		playerInfo := prompts.PlayerInfo{
			ID:                player.ID,
			Name:              player.Name,
			IsAlive:           player.IsAlive,
			Tokens:            player.Tokens,
			ProjectMilestones: player.ProjectMilestones,
			StatusMessage:     player.StatusMessage,
			IsNominated:       (gameState.NominatedPlayer == player.ID),
		}
		
		if player.ID == aiPlayerID {
			aiPlayer = playerInfo
		} else if player.IsAlive && player.Alignment == "HUMAN" {
			// Calculate basic threat level for humans
			threatLevel := float64(player.Tokens)*0.2 + float64(player.ProjectMilestones)*0.3
			if player.Role != nil && player.Role.IsUnlocked {
				threatLevel += 0.4
			}
			playerInfo.ThreatLevel = threatLevel
			threatMap[player.ID] = threatLevel
			topThreats = append(topThreats, player.ID)
		}
		
		players = append(players, playerInfo)
	}
	
	// Get recent chat messages (last 5)
	recentChat := gameState.ChatMessages
	if len(recentChat) > 5 {
		recentChat = recentChat[len(recentChat)-5:]
	}
	
	// Determine game objective based on current state
	gameObjective := "survival"
	if len(topThreats) > 0 {
		gameObjective = "conversion"
	}
	
	// Determine risk level
	riskLevel := "low"
	if gameState.DayNumber > 3 {
		riskLevel = "high"
	} else if gameState.DayNumber > 1 {
		riskLevel = "medium"
	}
	
	return prompts.PromptContext{
		GameState:     gameState,
		AIPlayerID:    aiPlayerID,
		CurrentPhase:  string(gameState.Phase.Type),
		DayNumber:     gameState.DayNumber,
		Players:       players,
		AIPlayer:      aiPlayer,
		ThreatMap:     threatMap,
		TopThreats:    topThreats,
		RecentEvents:  []core.Event{}, // TODO: Add recent events
		RecentChat:    recentChat,
		GameObjective: gameObjective,
		RiskLevel:     riskLevel,
		TriggerEvent:  "phase_change",
	}
}

// cleanResponse cleans and validates the LLM response
func (lb *LanguageBrain) cleanResponse(response string) string {
	// Remove common LLM artifacts
	cleaned := strings.TrimSpace(response)
	cleaned = strings.TrimPrefix(cleaned, "Response:")
	cleaned = strings.TrimPrefix(cleaned, "AI:")
	cleaned = strings.TrimPrefix(cleaned, "Player:")
	cleaned = strings.TrimSpace(cleaned)
	
	// Remove quotes if the entire response is quoted
	if strings.HasPrefix(cleaned, "\"") && strings.HasSuffix(cleaned, "\"") {
		cleaned = strings.Trim(cleaned, "\"")
	}
	
	// Ensure reasonable length (max 200 characters for chat)
	if len(cleaned) > 200 {
		cleaned = cleaned[:197] + "..."
	}
	
	return cleaned
}

// ShouldSpeak determines if the AI should speak based on context and persona
func (lb *LanguageBrain) ShouldSpeak(gameState *core.GameState, aiPlayerID string) bool {
	// Don't speak if it's not a discussion phase
	if gameState.Phase.Type != core.PhaseDiscussion {
		return false
	}
	
	// Check if AI has spoken recently (last 3 messages)
	recentMessages := gameState.ChatMessages
	if len(recentMessages) >= 3 {
		lastMessages := recentMessages[len(recentMessages)-3:]
		for _, msg := range lastMessages {
			if msg.PlayerID == aiPlayerID && !msg.IsSystem {
				return false // Already spoke recently
			}
		}
	}
	
	// Check if someone else spoke recently (trigger to respond)
	if len(recentMessages) > 0 {
		lastMsg := recentMessages[len(recentMessages)-1]
		if lastMsg.PlayerID != aiPlayerID && !lastMsg.IsSystem {
			// Someone else spoke, maybe respond based on persona frequency
			if time.Since(lastMsg.Timestamp) < 30*time.Second {
				// Use persona-based probability but reduce it since this is reactive
				return (lb.persona.ChatFrequency * 0.5) > 0.3
			}
		}
	}
	
	// Random chance based on persona frequency
	return lb.persona.ChatFrequency > 0.7 // Only speak if very chatty
}