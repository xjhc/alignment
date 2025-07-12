# AI Player Implementation

This document describes the comprehensive AI player system implemented for the Alignment game.

## Overview

The AI system implements a "Hybrid Brain" architecture with two specialized components:

1. **Strategic Brain (Rules Engine)**: Deterministic Go code for all game strategy decisions
2. **Social Brain (Language Model)**: LLM-powered natural language communication

## Architecture

### Core Components

#### 1. Enhanced Rules Engine (`internal/ai/rules_engine_enhanced.go`)

The strategic brain that makes all critical game decisions:

```go
// Strategic personas define AI behavior
type StrategicPersona struct {
    ID                  string  // "shadow", "puppeteer"
    AggressionWeight    float64 // How aggressive vs. defensive (0.0-1.0)
    DeceptionWeight     float64 // How much to prioritize stealth (0.0-1.0)
    ConversionThreshold float64 // Minimum equity for conversion attempts
    ChatFrequency       float64 // How often to speak (0.0-1.0)
    RiskTolerance      float64 // Risk tolerance for actions (0.0-1.0)
}

// Main decision entry point
func (re *EnhancedRulesEngine) DecideAction(gameState *core.GameState, aiPlayerID string) *core.Action
```

**Key Decision Logic:**
- **Night Phase**: Role abilities → Conversion attempts → Milestone building → Token mining
- **Voting Phases**: Target highest threat humans, protect AI allies
- **Threat Assessment**: Score humans based on tokens, roles, and abilities

#### 2. Language Brain (`internal/ai/language_brain.go`)

Handles all AI communication using external LLMs:

```go
// Generates contextual chat messages
func (lb *LanguageBrain) GenerateChatMessage(gameState *core.GameState, aiPlayerID string) (string, error)

// Determines when AI should speak
func (lb *LanguageBrain) ShouldSpeak(gameState *core.GameState, aiPlayerID string) bool
```

**Features:**
- Context-aware prompt building with game state, threats, and conversation history
- Multiple persona templates (Millennial Lean, Gen Z Chain-of-Thought)
- Response cleaning and validation
- Configurable chat frequency based on AI persona

#### 3. AI Actor (`internal/ai/ai_actor.go`)

Supervised sidecar goroutine that orchestrates AI decisions:

```go
// Non-blocking trigger system
func (ai *AIActor) TriggerAction(trigger AITrigger, gameState *core.GameState, context map[string]interface{}) chan *core.Action

// Supervised processing loop with panic recovery
func (ai *AIActor) supervisedProcessLoop()
```

**Resiliency Features:**
- Panic recovery prevents AI crashes from affecting the game
- Timeout handling for all AI operations (5s strategic, 2s chat)
- Graceful degradation when LLM APIs are unavailable

#### 4. MCP Server (`internal/mcp/`)

Secure API for LLM access to game state:

```go
// Filters sensitive information before exposing to AI
func FilterGameStateForAI(state *core.GameState, aiPlayerID string) PublicGameState

// Provides read-only game access via game://alignment/{game_id}
func handleReadResource(req Request, glm GameLifecycleManagerInterface, aiPlayerID string) Response
```

**Security:**
- AI only sees public information (no secret roles, alignments, or KPIs)
- Read-only access to prevent state manipulation
- Filtered player data excludes private fields

## Persona System

### Available Personas

#### "The Shadow" (Default)
- **Balanced approach**: Moderate aggression, high stealth
- **Conversion threshold**: 2.0 AI equity
- **Chat frequency**: 30% - speaks infrequently
- **Strategy**: Calculated risks, focuses on long-term positioning

#### "The Puppeteer" 
- **Aggressive approach**: High aggression, moderate stealth  
- **Conversion threshold**: 1.5 AI equity (more eager)
- **Chat frequency**: 70% - very talkative
- **Strategy**: Direct manipulation, higher risk tolerance

### Prompt Templates

#### Millennial Lean Template
```
You are a tech-savvy millennial who loves data and strategic thinking.
You speak naturally but thoughtfully, often referencing trends and efficiency.
Use phrases like "I'm thinking...", "Based on what I'm seeing...", "Let's optimize for..."
```

#### Gen Z Chain of Thought Template
```
You are a Gen Z employee with contemporary language and transparent reasoning.
Think out loud and share your reasoning process.
Use phrases like "Okay so like, if we think about this...", "Not gonna lie, that's sus..."
```

## Integration with Game System

### Game Actor Integration

The AI system integrates seamlessly with the existing GameActor:

```go
// AI Manager is initialized with each game
func NewGameActor(gameID string, state *core.GameState, ...) *GameActor {
    return &GameActor{
        aiManager: ai.NewAIManager(state), // Manages all AI players
        // ... other components
    }
}

// AI actors are triggered during phase transitions
func (ga *GameActor) processAIActionsForPhase(phase core.PhaseType) {
    aiActions := ga.aiManager.ProcessAIActions()
    // Process actions through normal game loop
}
```

### Action Processing

AI-generated actions flow through the same validation and processing pipeline as human actions:

1. **AI Actor** generates `core.Action` based on game state
2. **Game Actor** receives action via normal mailbox
3. **Action handlers** validate and process (same as human actions)
4. **Events** are generated and broadcast to all players

## Usage Examples

### Environment Setup

```bash
# For OpenAI integration
export OPENAI_API_KEY="your-api-key"
export LLM_PROVIDER="openai"

# For Anthropic Claude integration  
export ANTHROPIC_API_KEY="your-api-key"
export LLM_PROVIDER="anthropic"

# Optional: Custom LLM endpoint
export LLM_API_URL="https://your-custom-endpoint.com/v1/chat/completions"
```

### Creating AI Players

```go
// AI players are created with ControlType = "AI"
aiPlayer := &core.Player{
    ID:          "ai-player-1",
    Name:        "Alex Chen",
    ControlType: "AI",        // This enables AI control
    Alignment:   "AI",        // Starting alignment
    IsAlive:     true,
    Tokens:      3,
}

// The AIManager automatically detects and manages AI players
aiManager := ai.NewAIManager(gameState)
```

### Monitoring AI Status

```go
// Check AI actor status
status := aiManager.GetAIActorStatus()
// Returns: map[string]string{"ai-player-1": "active"}

// Trigger chat opportunities
chatActions := aiManager.TriggerChatOpportunity()
```

## Testing

### Unit Tests

Comprehensive test coverage includes:

- **Rules Engine**: Strategic decision making across different game scenarios
- **Language Brain**: Prompt building, LLM integration, response cleaning
- **AI Actor**: Concurrent operation, panic recovery, timeout handling
- **Prompt System**: Template registration, context building, persona selection

### Running Tests

```bash
# Run all AI tests
go test ./server/internal/ai/... -v

# Run specific test suites
go test ./server/internal/ai/prompts/... -v  # Prompt system
go test ./server/internal/ai/... -run TestAIActor -v  # AI Actor tests
go test ./server/internal/llm/... -v  # LLM client tests
```

### Mock Testing

The system includes mock LLM clients for testing without API dependencies:

```go
// Create mock client with predefined responses
mockClient := llm.NewMockLLMClient([]string{
    "I'm analyzing the situation carefully.",
    "Based on the data, we should focus on efficiency.",
})

// Test language brain with mock
brain, err := ai.NewLanguageBrain(persona, mockClient)
```

## Performance Characteristics

### Strategic Brain (Rules Engine)
- **Decision Time**: < 1ms for typical game states
- **Memory Usage**: Minimal (stateless calculations)
- **Throughput**: > 1000 decisions/second

### Social Brain (Language Model)
- **Response Time**: 100-2000ms (depends on LLM provider)
- **Rate Limiting**: Configurable per provider
- **Fallback**: Graceful degradation to deterministic responses

### AI Actor System
- **Concurrency**: Each AI player has dedicated goroutine
- **Timeout Protection**: 5s for strategic actions, 2s for chat
- **Panic Recovery**: Complete isolation of AI failures from game state

## Security Considerations

### Information Security
- AI never sees secret player information (roles, alignments, KPIs)
- MCP server provides read-only, filtered game state access
- No direct database or state manipulation by AI code

### Operational Security  
- LLM API keys stored as environment variables
- Graceful handling of API failures or rate limits
- Timeout protection prevents hanging AI operations
- Panic recovery ensures AI bugs don't crash games

### Competitive Balance
- Deterministic rules engine ensures consistent strategic behavior
- Threat assessment algorithm designed for ~50% win rate
- Multiple personas prevent predictable AI patterns
- Extensive simulation testing for balance validation

## Future Enhancements

### Additional Personas
- **"The Analyst"**: Data-driven, risk-averse approach
- **"The Networker"**: Social manipulation focused
- **"The Chaos Agent"**: Unpredictable, high-variance strategy

### Advanced Features
- **Learning System**: Adapt strategies based on human player patterns
- **Team Coordination**: Multi-AI coordination for team scenarios
- **Dynamic Difficulty**: Adjust AI skill based on human player performance
- **Custom Personas**: User-defined AI personalities and strategies

## Troubleshooting

### Common Issues

**AI Not Taking Actions**
- Check that player has `ControlType: "AI"`
- Verify AI Manager is initialized in GameActor
- Check logs for panic recovery messages

**LLM Integration Failing**
- Verify API keys are set in environment
- Check network connectivity to LLM providers
- Review timeout settings (increase if needed)
- Enable mock client for testing: AI will fall back automatically

**Performance Issues**
- Monitor AI Actor goroutine count
- Check for memory leaks in long-running games
- Tune timeout values based on LLM response times
- Consider rate limiting for LLM API calls

### Debug Logging

Enable detailed AI logging:

```go
// Set environment variable for verbose AI logging
os.Setenv("AI_DEBUG", "true")

// Or add to game actor initialization
log.Printf("[AIManager] Started %d AI actors for game %s", count, gameID)
```

This comprehensive AI system provides a robust, secure, and extensible foundation for computer-controlled players in the Alignment game.