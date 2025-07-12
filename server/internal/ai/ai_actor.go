package ai

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/xjhc/alignment/core"
)

// AITrigger represents different types of AI action triggers
type AITrigger string

const (
	TriggerPhaseChange   AITrigger = "PHASE_CHANGE"
	TriggerPlayerSpoke   AITrigger = "PLAYER_SPOKE"
	TriggerVotingStart   AITrigger = "VOTING_START"
	TriggerNightStart    AITrigger = "NIGHT_START"
	TriggerRandomChat    AITrigger = "RANDOM_CHAT"
	TriggerGameEnd       AITrigger = "GAME_END"
)

// AITriggerEvent represents a trigger event sent to the AI Actor
type AITriggerEvent struct {
	Type       AITrigger
	GameState  *core.GameState
	PlayerID   string
	Context    map[string]interface{}
	ResultChan chan *core.Action // For returning the AI's decision
}

// AIActor represents a supervised sidecar goroutine that manages AI decisions
type AIActor struct {
	playerID       string
	gameID         string
	rulesEngine    *EnhancedRulesEngine
	languageBrain  *LanguageBrain
	
	// Communication channels
	triggerChan    chan AITriggerEvent
	shutdownChan   chan struct{}
	
	// Concurrency management
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	
	// Configuration
	isActive       bool
	responseTimeout time.Duration
}

// NewAIActor creates a new AI actor for a specific AI player
func NewAIActor(playerID, gameID string, persona StrategicPersona) (*AIActor, error) {
	// Create enhanced rules engine with persona
	rulesEngine := NewEnhancedRulesEngine(persona)
	
	// Create language brain with persona
	languageBrain, err := NewLanguageBrain(persona, nil) // Will use default/mock client
	if err != nil {
		log.Printf("Warning: Failed to create LanguageBrain for AI %s: %v", playerID, err)
		// Continue without language brain for strategic actions
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &AIActor{
		playerID:       playerID,
		gameID:         gameID,
		rulesEngine:    rulesEngine,
		languageBrain:  languageBrain,
		triggerChan:    make(chan AITriggerEvent, 10), // Buffered to prevent blocking
		shutdownChan:   make(chan struct{}),
		ctx:            ctx,
		cancel:         cancel,
		isActive:       true,
		responseTimeout: 5 * time.Second,
	}, nil
}

// Start begins the AI actor's processing loop in a supervised goroutine
func (ai *AIActor) Start() {
	ai.wg.Add(1)
	go ai.supervisedProcessLoop()
}

// Stop gracefully shuts down the AI actor
func (ai *AIActor) Stop() {
	log.Printf("[AIActor/%s] Stopping AI actor for player %s", ai.gameID, ai.playerID)
	
	ai.isActive = false
	ai.cancel()
	
	// Signal shutdown and wait for goroutine to finish
	close(ai.shutdownChan)
	ai.wg.Wait()
	
	// Close channels
	close(ai.triggerChan)
	
	log.Printf("[AIActor/%s] AI actor stopped for player %s", ai.gameID, ai.playerID)
}

// TriggerAction sends a trigger to the AI actor and returns a channel for the result
func (ai *AIActor) TriggerAction(trigger AITrigger, gameState *core.GameState, context map[string]interface{}) chan *core.Action {
	resultChan := make(chan *core.Action, 1)
	
	if !ai.isActive {
		close(resultChan)
		return resultChan
	}
	
	event := AITriggerEvent{
		Type:       trigger,
		GameState:  gameState,
		PlayerID:   ai.playerID,
		Context:    context,
		ResultChan: resultChan,
	}
	
	select {
	case ai.triggerChan <- event:
		// Successfully queued trigger
	case <-ai.ctx.Done():
		// AI actor is shutting down
		close(resultChan)
	default:
		// Channel is full, drop the trigger to prevent blocking
		log.Printf("[AIActor/%s] Trigger channel full, dropping %s trigger", ai.gameID, trigger)
		close(resultChan)
	}
	
	return resultChan
}

// supervisedProcessLoop runs the main AI processing loop with panic recovery
func (ai *AIActor) supervisedProcessLoop() {
	defer ai.wg.Done()
	
	// Panic recovery wrapper (ADR-008 compliance)
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AIActor/%s] PANIC RECOVERED in AI actor for player %s: %v", ai.gameID, ai.playerID, r)
			// The AI actor is now considered failed and will not process more events
			ai.isActive = false
		}
	}()
	
	log.Printf("[AIActor/%s] Starting AI processing loop for player %s", ai.gameID, ai.playerID)
	
	for {
		select {
		case event := <-ai.triggerChan:
			ai.handleTriggerEvent(event)
			
		case <-ai.shutdownChan:
			log.Printf("[AIActor/%s] Received shutdown signal for player %s", ai.gameID, ai.playerID)
			return
			
		case <-ai.ctx.Done():
			log.Printf("[AIActor/%s] Context cancelled for player %s", ai.gameID, ai.playerID)
			return
		}
	}
}

// handleTriggerEvent processes a single trigger event
func (ai *AIActor) handleTriggerEvent(event AITriggerEvent) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AIActor/%s] PANIC in trigger handler for %s: %v", ai.gameID, event.Type, r)
			// Close the result channel to signal failure
			close(event.ResultChan)
		}
	}()
	
	// Set a timeout for AI decision making
	ctx, cancel := context.WithTimeout(ai.ctx, ai.responseTimeout)
	defer cancel()
	
	// Process the trigger in a separate goroutine with timeout
	done := make(chan *core.Action, 1)
	go func() {
		action := ai.processAIDecision(event)
		done <- action
	}()
	
	select {
	case action := <-done:
		event.ResultChan <- action
		close(event.ResultChan)
		
	case <-ctx.Done():
		log.Printf("[AIActor/%s] AI decision timeout for trigger %s", ai.gameID, event.Type)
		event.ResultChan <- nil
		close(event.ResultChan)
	}
}

// processAIDecision makes the actual AI decision based on the trigger
func (ai *AIActor) processAIDecision(event AITriggerEvent) *core.Action {
	switch event.Type {
	case TriggerPhaseChange:
		return ai.handlePhaseChange(event)
		
	case TriggerVotingStart, TriggerNightStart:
		return ai.handleStrategicPhase(event)
		
	case TriggerPlayerSpoke, TriggerRandomChat:
		return ai.handleChatOpportunity(event)
		
	case TriggerGameEnd:
		return nil // No action needed on game end
		
	default:
		log.Printf("[AIActor/%s] Unknown trigger type: %s", ai.gameID, event.Type)
		return nil
	}
}

// handlePhaseChange processes phase transition triggers
func (ai *AIActor) handlePhaseChange(event AITriggerEvent) *core.Action {
	phase := event.GameState.Phase.Type
	
	log.Printf("[AIActor/%s] Processing phase change to %s for player %s", ai.gameID, phase, ai.playerID)
	
	// Use the enhanced rules engine for strategic decisions
	return ai.rulesEngine.DecideAction(event.GameState, ai.playerID)
}

// handleStrategicPhase processes voting and night action phases
func (ai *AIActor) handleStrategicPhase(event AITriggerEvent) *core.Action {
	log.Printf("[AIActor/%s] Processing strategic phase %s for player %s", ai.gameID, event.Type, ai.playerID)
	
	// Use the enhanced rules engine for strategic decisions
	return ai.rulesEngine.DecideAction(event.GameState, ai.playerID)
}

// handleChatOpportunity processes chat opportunities
func (ai *AIActor) handleChatOpportunity(event AITriggerEvent) *core.Action {
	if ai.languageBrain == nil {
		// No language brain available, skip chat
		return nil
	}
	
	// Check if the AI should speak
	if !ai.languageBrain.ShouldSpeak(event.GameState, ai.playerID) {
		return nil
	}
	
	log.Printf("[AIActor/%s] Generating chat message for player %s", ai.gameID, ai.playerID)
	
	// Generate chat message using language brain
	message, err := ai.languageBrain.GenerateChatMessage(event.GameState, ai.playerID)
	if err != nil {
		log.Printf("[AIActor/%s] Failed to generate chat message: %v", ai.gameID, err)
		return nil
	}
	
	if message == "" {
		return nil
	}
	
	// Create chat action
	return &core.Action{
		Type:      core.ActionSendMessage,
		PlayerID:  ai.playerID,
		GameID:    ai.gameID,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"content": message,
		},
	}
}

// GetStatus returns the current status of the AI actor
func (ai *AIActor) GetStatus() string {
	if !ai.isActive {
		return "inactive"
	}
	
	select {
	case <-ai.ctx.Done():
		return "shutting_down"
	default:
		return "active"
	}
}

// GetPlayerID returns the player ID this AI actor manages
func (ai *AIActor) GetPlayerID() string {
	return ai.playerID
}

// UpdateGameState allows the AI actor to be notified of game state changes
// This could be used for more sophisticated AI that tracks game history
func (ai *AIActor) UpdateGameState(gameState *core.GameState) {
	// TODO: Could be used to build game history for more context-aware AI
	// For now, the AI gets the current state with each trigger
}