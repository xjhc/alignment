package ai

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
	"github.com/xjhc/alignment/server/internal/llm"
)

func TestLanguageBrain_GenerateChatMessage(t *testing.T) {
	// Create mock LLM client
	mockClient := llm.NewMockLLMClient([]string{
		"I'm analyzing the current situation carefully.",
		"Based on the data, we should focus on efficiency.",
	})

	brain, err := NewLanguageBrain(GetShadowPersona(), mockClient)
	if err != nil {
		t.Fatalf("Failed to create LanguageBrain: %v", err)
	}

	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseDiscussion},
		Players: map[string]*core.Player{
			"ai-1": {
				ID:        "ai-1",
				Name:      "AI Player",
				IsAlive:   true,
				Alignment: "AI",
				Tokens:    3,
			},
			"human-1": {
				ID:        "human-1",
				Name:      "Human Player",
				IsAlive:   true,
				Alignment: "HUMAN",
				Tokens:    4,
			},
		},
		ChatMessages: []core.ChatMessage{
			{
				ID:        "msg-1",
				PlayerID:  "human-1",
				Message:   "What do you think about the current situation?",
				Timestamp: time.Now().Add(-1 * time.Minute),
			},
		},
	}

	message, err := brain.GenerateChatMessage(gameState, "ai-1")
	if err != nil {
		t.Fatalf("Failed to generate chat message: %v", err)
	}

	if message == "" {
		t.Error("Expected non-empty chat message")
	}

	// Should be cleaned and reasonably short
	if len(message) > 200 {
		t.Errorf("Message too long: %d characters", len(message))
	}

	// Should not contain common LLM artifacts
	if strings.HasPrefix(message, "Response:") || strings.HasPrefix(message, "AI:") {
		t.Errorf("Message contains LLM artifacts: %s", message)
	}
}

func TestLanguageBrain_ShouldSpeak(t *testing.T) {
	brain, err := NewLanguageBrainWithDefaults()
	if err != nil {
		t.Fatalf("Failed to create LanguageBrain: %v", err)
	}

	tests := []struct {
		name      string
		phase     core.PhaseType
		messages  []core.ChatMessage
		aiPlayerID string
		expected  bool
	}{
		{
			name:      "Should not speak during night phase",
			phase:     core.PhaseNight,
			messages:  []core.ChatMessage{},
			aiPlayerID: "ai-1",
			expected:  false,
		},
		{
			name:     "Should not speak if recently spoke",
			phase:    core.PhaseDiscussion,
			aiPlayerID: "ai-1",
			messages: []core.ChatMessage{
				{
					PlayerID:  "ai-1",
					Message:   "I just spoke",
					Timestamp: time.Now().Add(-10 * time.Second),
				},
			},
			expected: false,
		},
		{
			name:     "Should consider speaking in discussion with no recent messages",
			phase:    core.PhaseDiscussion,
			aiPlayerID: "ai-1",
			messages: []core.ChatMessage{
				{
					PlayerID:  "human-1",
					Message:   "Old message",
					Timestamp: time.Now().Add(-5 * time.Minute),
				},
			},
			expected: false, // Depends on persona frequency, likely false for Shadow
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gameState := &core.GameState{
				Phase: core.Phase{Type: tt.phase},
				ChatMessages: tt.messages,
			}

			result := brain.ShouldSpeak(gameState, tt.aiPlayerID)
			if result != tt.expected {
				t.Errorf("Expected ShouldSpeak = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestLanguageBrain_CleanResponse(t *testing.T) {
	brain, _ := NewLanguageBrainWithDefaults()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Response: This is a test message",
			expected: "This is a test message",
		},
		{
			input:    "AI: I think we should vote carefully",
			expected: "I think we should vote carefully",
		},
		{
			input:    "\"I'm thinking about the situation\"",
			expected: "I'm thinking about the situation",
		},
		{
			input:    "   Whitespace padded message   ",
			expected: "Whitespace padded message",
		},
	}

	for _, tt := range tests {
		result := brain.cleanResponse(tt.input)
		if result != tt.expected {
			t.Errorf("cleanResponse(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestLanguageBrain_BuildPromptContext(t *testing.T) {
	brain, _ := NewLanguageBrainWithDefaults()

	gameState := &core.GameState{
		ID:        "test-game",
		DayNumber: 4,
		Phase:     core.Phase{Type: core.PhaseDiscussion},
		Players: map[string]*core.Player{
			"ai-1": {
				ID:                "ai-1",
				Name:              "AI Player",
				IsAlive:           true,
				Alignment:         "AI",
				Tokens:            3,
				ProjectMilestones: 2,
			},
			"human-1": {
				ID:                "human-1",
				Name:              "Human Player",
				IsAlive:           true,
				Alignment:         "HUMAN",
				Tokens:            5,
				ProjectMilestones: 3,
				Role: &core.Role{
					Type:       core.RoleCISO,
					IsUnlocked: true,
				},
			},
		},
		ChatMessages: []core.ChatMessage{
			{PlayerID: "human-1", Message: "Test message", Timestamp: time.Now()},
		},
	}

	ctx := brain.buildPromptContext(gameState, "ai-1")

	if ctx.AIPlayerID != "ai-1" {
		t.Errorf("Expected AI player ID ai-1, got %s", ctx.AIPlayerID)
	}

	if ctx.DayNumber != 4 {
		t.Errorf("Expected day number 4, got %d", ctx.DayNumber)
	}

	if ctx.CurrentPhase != "DISCUSSION" {
		t.Errorf("Expected phase DISCUSSION, got %s", ctx.CurrentPhase)
	}

	if len(ctx.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(ctx.Players))
	}

	// Should have threat analysis for human players
	if len(ctx.ThreatMap) != 1 {
		t.Errorf("Expected 1 threat entry, got %d", len(ctx.ThreatMap))
	}

	if ctx.RiskLevel != "high" { // Day 3 should be high risk
		t.Errorf("Expected high risk level, got %s", ctx.RiskLevel)
	}
}

// Mock LLM client for testing
type testLLMClient struct {
	responses []string
	callCount int
}

func (t *testLLMClient) GenerateChat(ctx context.Context, prompt string) (string, error) {
	if t.callCount >= len(t.responses) {
		return "Default test response", nil
	}
	
	response := t.responses[t.callCount]
	t.callCount++
	return response, nil
}

func TestNewLanguageBrainWithDefaults(t *testing.T) {
	brain, err := NewLanguageBrainWithDefaults()
	if err != nil {
		t.Fatalf("Expected NewLanguageBrainWithDefaults to succeed, got error: %v", err)
	}

	if brain.persona.ID != "shadow" {
		t.Errorf("Expected default persona to be shadow, got %s", brain.persona.ID)
	}

	if brain.promptTemplate == nil {
		t.Error("Expected prompt template to be initialized")
	}

	if brain.llmClient == nil {
		t.Error("Expected LLM client to be initialized")
	}
}

func BenchmarkGenerateChatMessage(b *testing.B) {
	mockClient := llm.NewMockLLMClient([]string{"Benchmark response"})
	brain, _ := NewLanguageBrain(GetShadowPersona(), mockClient)

	gameState := &core.GameState{
		ID:        "bench-game",
		DayNumber: 2,
		Phase:     core.Phase{Type: core.PhaseDiscussion},
		Players: map[string]*core.Player{
			"ai-1": {ID: "ai-1", IsAlive: true, Alignment: "AI"},
			"human-1": {ID: "human-1", IsAlive: true, Alignment: "HUMAN"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		brain.GenerateChatMessage(gameState, "ai-1")
	}
}