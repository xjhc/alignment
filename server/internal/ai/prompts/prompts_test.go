package prompts

import (
	"strings"
	"testing"
	"time"

	"github.com/xjhc/alignment/core"
)

func TestPromptRegistry(t *testing.T) {
	// Test that templates are registered
	count := Count()
	if count < 2 {
		t.Errorf("Expected at least 2 templates, got %d", count)
	}

	// Test Get function
	template, err := Get("millennial_lean")
	if err != nil {
		t.Fatalf("Failed to get millennial_lean template: %v", err)
	}

	if template.ID() != "millennial_lean" {
		t.Errorf("Expected ID millennial_lean, got %s", template.ID())
	}

	// Test GetRandom function
	randomTemplate := GetRandom()
	if randomTemplate == nil {
		t.Error("GetRandom returned nil")
	}

	// Test GetByPersona function
	shadowTemplate, err := GetByPersona("shadow")
	if err != nil {
		t.Fatalf("Failed to get template for shadow persona: %v", err)
	}

	if shadowTemplate == nil {
		t.Error("GetByPersona returned nil for shadow")
	}

	// Test non-existent template
	_, err = Get("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}
}

func TestMillennialLeanTemplate(t *testing.T) {
	template := NewMillennialLean()

	if template.ID() != "millennial_lean" {
		t.Errorf("Expected ID millennial_lean, got %s", template.ID())
	}

	if template.Name() == "" {
		t.Error("Expected non-empty name")
	}

	if template.Description() == "" {
		t.Error("Expected non-empty description")
	}

	// Test prompt building
	ctx := createTestPromptContext()
	prompt := template.BuildPrompt(ctx)

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Should contain game state context
	if !strings.Contains(prompt, "Day 2") {
		t.Error("Prompt should contain day number")
	}

	if !strings.Contains(prompt, "DISCUSSION") {
		t.Error("Prompt should contain current phase")
	}

	// Should contain personality traits
	if !strings.Contains(prompt, "millennial") || !strings.Contains(prompt, "analytical") {
		t.Error("Prompt should contain personality descriptions")
	}

	// Should contain strategic guidelines
	if !strings.Contains(prompt, "Never reveal you're an AI") {
		t.Error("Prompt should contain strategic guidelines")
	}
}

func TestGenZCoTTemplate(t *testing.T) {
	template := NewGenZCoT()

	if template.ID() != "gen_z_cot" {
		t.Errorf("Expected ID gen_z_cot, got %s", template.ID())
	}

	// Test prompt building
	ctx := createTestPromptContext()
	prompt := template.BuildPrompt(ctx)

	if prompt == "" {
		t.Error("Expected non-empty prompt")
	}

	// Should contain Gen Z characteristics
	if !strings.Contains(prompt, "Gen Z") {
		t.Error("Prompt should mention Gen Z")
	}

	// Should encourage chain of thought
	if !strings.Contains(prompt, "think out loud") {
		t.Error("Prompt should encourage thinking out loud")
	}

	// Should contain contemporary language examples
	if !strings.Contains(prompt, "like,") || !strings.Contains(prompt, "literally") {
		t.Error("Prompt should contain contemporary language examples")
	}
}

func TestBuildGameStateContext(t *testing.T) {
	ctx := createTestPromptContext()
	context := BuildGameStateContext(ctx)

	if context == "" {
		t.Error("Expected non-empty game state context")
	}

	// Should contain game info
	if !strings.Contains(context, "Day 2") {
		t.Error("Context should contain day number")
	}

	if !strings.Contains(context, "DISCUSSION") {
		t.Error("Context should contain phase")
	}

	// Should contain player info
	if !strings.Contains(context, "AI Player") {
		t.Error("Context should contain AI player name")
	}

	if !strings.Contains(context, "Human Player") {
		t.Error("Context should contain human player name")
	}

	// Should contain token counts
	if !strings.Contains(context, "3 tokens") {
		t.Error("Context should contain token information")
	}

	// Should contain recent chat
	if !strings.Contains(context, "What should we do") {
		t.Error("Context should contain recent chat messages")
	}
}

func TestBuildThreatContext(t *testing.T) {
	ctx := createTestPromptContext()
	ctx.TopThreats = []string{"human-1", "human-2"}
	ctx.ThreatMap = map[string]float64{
		"human-1": 0.8,
		"human-2": 0.3,
	}

	threatContext := BuildThreatContext(ctx)

	if threatContext == "" {
		t.Error("Expected non-empty threat context")
	}

	// Should contain threat analysis header
	if !strings.Contains(threatContext, "THREAT ANALYSIS") {
		t.Error("Context should contain threat analysis header")
	}

	// Should contain threat levels
	if !strings.Contains(threatContext, "0.80") {
		t.Error("Context should contain threat levels")
	}

	// Should contain player names
	if !strings.Contains(threatContext, "Human Player") {
		t.Error("Context should contain threat player names")
	}

	// Test empty threats
	emptyCtx := ctx
	emptyCtx.TopThreats = []string{}
	emptyThreatContext := BuildThreatContext(emptyCtx)
	
	if !strings.Contains(emptyThreatContext, "No significant threats") {
		t.Error("Empty threat context should indicate no threats")
	}
}

func TestPromptContextValidation(t *testing.T) {
	ctx := createTestPromptContext()

	// Test required fields
	if ctx.AIPlayerID == "" {
		t.Error("AI Player ID should not be empty")
	}

	if ctx.CurrentPhase == "" {
		t.Error("Current phase should not be empty")
	}

	if len(ctx.Players) == 0 {
		t.Error("Players list should not be empty")
	}

	// Test AI player is included
	aiPlayerFound := false
	for _, player := range ctx.Players {
		if player.ID == ctx.AIPlayerID {
			aiPlayerFound = true
			break
		}
	}
	if !aiPlayerFound {
		t.Error("AI player should be included in players list")
	}

	// Test threat levels are only set for non-AI players
	for _, player := range ctx.Players {
		if player.ID == ctx.AIPlayerID && player.ThreatLevel > 0 {
			t.Error("AI player should not have threat level set")
		}
	}
}

func TestSafeFormat(t *testing.T) {
	// Test normal formatting
	result := SafeFormat("Hello %s, day %d", "world", 5)
	expected := "Hello world, day 5"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with too few arguments (should not panic)
	result = SafeFormat("Hello %s", "world")
	expected = "Hello world"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// Benchmark tests
func BenchmarkBuildGameStateContext(b *testing.B) {
	ctx := createTestPromptContext()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BuildGameStateContext(ctx)
	}
}

func BenchmarkMillennialLeanBuildPrompt(b *testing.B) {
	template := NewMillennialLean()
	ctx := createTestPromptContext()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		template.BuildPrompt(ctx)
	}
}

// Helper function to create test prompt context
func createTestPromptContext() PromptContext {
	return PromptContext{
		GameState: &core.GameState{
			ID:        "test-game",
			DayNumber: 2,
			Phase:     core.Phase{Type: core.PhaseDiscussion},
		},
		AIPlayerID:   "ai-1",
		CurrentPhase: "DISCUSSION",
		DayNumber:    2,
		Players: []PlayerInfo{
			{
				ID:                "ai-1",
				Name:              "AI Player",
				IsAlive:           true,
				Tokens:            3,
				ProjectMilestones: 2,
			},
			{
				ID:                "human-1",
				Name:              "Human Player",
				IsAlive:           true,
				Tokens:            5,
				ProjectMilestones: 3,
				ThreatLevel:       0.8,
			},
		},
		AIPlayer: PlayerInfo{
			ID:                "ai-1",
			Name:              "AI Player",
			IsAlive:           true,
			Tokens:            3,
			ProjectMilestones: 2,
		},
		ThreatMap: map[string]float64{
			"human-1": 0.8,
		},
		TopThreats: []string{"human-1"},
		RecentChat: []core.ChatMessage{
			{
				PlayerID:  "human-1",
				PlayerName: "Human Player",
				Message:   "What should we do next?",
				Timestamp: time.Now(),
			},
		},
		GameObjective: "conversion",
		RiskLevel:     "medium",
		TriggerEvent:  "phase_change",
	}
}