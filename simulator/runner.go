package simulator

import (
	"fmt"
	"time"

	"github.com/xjhc/alignment/core"
)

// SimulationResult contains the outcome of a single game simulation
type SimulationResult struct {
	GameID       string               `json:"game_id"`
	Winner       string               `json:"winner"`       // "HUMANS" or "AI"
	Condition    string               `json:"condition"`    // Win condition type
	Duration     time.Duration        `json:"duration"`     // Game duration
	DayNumber    int                  `json:"day_number"`   // Final day number
	EventLog     []core.Event         `json:"event_log"`    // Complete event history
	FinalState   core.GameState       `json:"final_state"`  // Final game state
	PlayerStats  map[string]PlayerStat `json:"player_stats"` // Per-player statistics
}

// PlayerStat tracks statistics for individual players
type PlayerStat struct {
	PlayerID         string `json:"player_id"`
	PersonaType      string `json:"persona_type"`
	Alignment        string `json:"alignment"`
	Survived         bool   `json:"survived"`
	FinalTokens      int    `json:"final_tokens"`
	FinalMilestones  int    `json:"final_milestones"`
	ActionsPerformed int    `json:"actions_performed"`
	VotesCast        int    `json:"votes_cast"`
	MessagesExempt   int    `json:"messages_sent"`
}

// SimulationConfig defines parameters for running simulations
type SimulationConfig struct {
	PlayerCount    int                    `json:"player_count"`
	AICount        int                    `json:"ai_count"`
	PersonaWeights map[string]float64     `json:"persona_weights"` // Probability weights for persona selection
	Seed           int64                  `json:"seed"`
	TimeAcceleration int                  `json:"time_acceleration"` // How fast to advance time (in seconds per tick)
}

// DefaultSimulationConfig returns a balanced configuration for testing
func DefaultSimulationConfig() SimulationConfig {
	return SimulationConfig{
		PlayerCount:    6,
		AICount:        2,
		PersonaWeights: map[string]float64{
			"cautious_human":  0.4,
			"aggressive_human": 0.4,
			"deceptive_ai":    1.0, // AI always uses deceptive persona
		},
		Seed:             time.Now().UnixNano(),
		TimeAcceleration: 30, // 30 seconds of game time per simulation tick
	}
}

// SimulationRunner orchestrates a complete headless game simulation
type SimulationRunner struct {
	config        SimulationConfig
	gameState     *core.GameState
	personas      map[string]BotPersona
	eventLog      []core.Event
	simulatedTime time.Time
}

// NewSimulationRunner creates a new simulation runner with the given configuration
func NewSimulationRunner(config SimulationConfig) *SimulationRunner {
	return &SimulationRunner{
		config:        config,
		personas:      make(map[string]BotPersona),
		eventLog:      make([]core.Event, 0),
		simulatedTime: time.Now(),
	}
}

// RunSimulation executes a complete game simulation and returns the result
func (sr *SimulationRunner) RunSimulation() (*SimulationResult, error) {
	// Initialize the game
	if err := sr.initializeGame(); err != nil {
		return nil, fmt.Errorf("failed to initialize game: %w", err)
	}

	// Assign personas to players
	if err := sr.assignPersonas(); err != nil {
		return nil, fmt.Errorf("failed to assign personas: %w", err)
	}

	// Run the main simulation loop
	if err := sr.runMainLoop(); err != nil {
		return nil, fmt.Errorf("simulation failed: %w", err)
	}

	// Generate the final result
	result := sr.generateResult()
	return result, nil
}

// initializeGame creates a new game state and adds players
func (sr *SimulationRunner) initializeGame() error {
	gameID := fmt.Sprintf("sim_%d", sr.simulatedTime.UnixNano())
	sr.gameState = core.NewGameState(gameID, sr.simulatedTime)

	// Create players
	for i := 0; i < sr.config.PlayerCount; i++ {
		playerID := fmt.Sprintf("player_%d", i+1)
		
		// Determine alignment
		alignment := "HUMAN"
		if i < sr.config.AICount {
			alignment = "ALIGNED"
		}

		// Add player to game
		event := core.Event{
			ID:        fmt.Sprintf("join_%s_%d", playerID, sr.simulatedTime.UnixNano()),
			Type:      core.EventPlayerJoined,
			PlayerID:  playerID,
			GameID:    gameID,
			Timestamp: sr.simulatedTime,
			Payload: map[string]interface{}{
				"player_name": fmt.Sprintf("Bot %d", i+1),
				"alignment":   alignment,
			},
		}

		sr.applyEvent(event)
	}

	// Start the game
	startEvent := core.Event{
		ID:        fmt.Sprintf("start_%s_%d", gameID, sr.simulatedTime.UnixNano()),
		Type:      core.EventGameStarted,
		GameID:    gameID,
		Timestamp: sr.simulatedTime,
		Payload:   map[string]interface{}{},
	}
	sr.applyEvent(startEvent)

	return nil
}

// assignPersonas assigns bot personas to all players
func (sr *SimulationRunner) assignPersonas() error {
	for playerID, player := range sr.gameState.Players {
		var persona BotPersona
		
		if player.Alignment == "ALIGNED" {
			// AI players always use DeceptiveAI persona
			persona = NewDeceptiveAI(sr.config.Seed + int64(len(playerID)))
		} else {
			// Human players get assigned based on weights
			if sr.shouldAssignPersona("aggressive_human") {
				persona = NewAggressiveHuman(sr.config.Seed + int64(len(playerID)))
			} else {
				persona = NewCautiousHuman(sr.config.Seed + int64(len(playerID)))
			}
		}
		
		sr.personas[playerID] = persona
	}

	return nil
}

// shouldAssignPersona determines if a persona should be assigned based on weights
func (sr *SimulationRunner) shouldAssignPersona(personaType string) bool {
	weight, exists := sr.config.PersonaWeights[personaType]
	if !exists {
		return false
	}
	
	// Simple probability check (could be more sophisticated)
	return sr.simulatedTime.UnixNano()%100 < int64(weight*100)
}

// runMainLoop executes the main simulation loop until a win condition is met
func (sr *SimulationRunner) runMainLoop() error {
	maxIterations := 1000 // Prevent infinite loops
	iterations := 0

	for iterations < maxIterations {
		iterations++

		// Check for win condition
		if winCondition := core.CheckWinCondition(*sr.gameState); winCondition != nil {
			// Apply win condition event
			winEvent := core.Event{
				ID:        fmt.Sprintf("win_%s_%d", sr.gameState.ID, sr.simulatedTime.UnixNano()),
				Type:      core.EventGameEnded,
				GameID:    sr.gameState.ID,
				Timestamp: sr.simulatedTime,
				Payload: map[string]interface{}{
					"winner":      winCondition.Winner,
					"condition":   winCondition.Condition,
					"description": winCondition.Description,
				},
			}
			sr.applyEvent(winEvent)
			break
		}

		// Process current phase
		if err := sr.processCurrentPhase(); err != nil {
			return fmt.Errorf("error processing phase %s: %w", sr.gameState.Phase.Type, err)
		}

		// Advance time and check for phase transitions
		sr.advanceTime()
		if err := sr.checkPhaseTransition(); err != nil {
			return fmt.Errorf("error in phase transition: %w", err)
		}
	}

	if iterations >= maxIterations {
		return fmt.Errorf("simulation exceeded maximum iterations (%d)", maxIterations)
	}

	return nil
}

// processCurrentPhase handles actions for the current game phase
func (sr *SimulationRunner) processCurrentPhase() error {
	switch sr.gameState.Phase.Type {
	case core.PhaseLobby:
		// Nothing to do in lobby after game start
		return nil
	case core.PhaseSitrep:
		return sr.processSitrepPhase()
	case core.PhaseNomination:
		return sr.processNominationPhase()
	case core.PhaseVerdict:
		return sr.processVerdictPhase()
	case core.PhaseExtension:
		return sr.processExtensionPhase()
	case core.PhaseNight:
		return sr.processNightPhase()
	default:
		return fmt.Errorf("unknown phase type: %s", sr.gameState.Phase.Type)
	}
}

// processSitrepPhase handles pulse check responses
func (sr *SimulationRunner) processSitrepPhase() error {
	for playerID, player := range sr.gameState.Players {
		if !player.IsAlive {
			continue
		}

		persona := sr.personas[playerID]
		if persona == nil {
			continue
		}

		action := persona.DecideAction(*sr.gameState, playerID)
		if action != nil && action.Type == core.ActionSubmitPulseCheck {
			events, err := core.ProcessPlayerAction(*sr.gameState, *action, sr.simulatedTime)
			if err != nil {
				continue // Skip invalid actions
			}
			sr.applyEvents(events)
		}
	}
	return nil
}

// processNominationPhase handles player nominations
func (sr *SimulationRunner) processNominationPhase() error {
	for playerID, player := range sr.gameState.Players {
		if !player.IsAlive {
			continue
		}

		persona := sr.personas[playerID]
		if persona == nil {
			continue
		}

		action := persona.DecideAction(*sr.gameState, playerID)
		if action != nil && action.Type == core.ActionSubmitVote {
			events, err := core.ProcessPlayerAction(*sr.gameState, *action, sr.simulatedTime)
			if err != nil {
				continue // Skip invalid actions
			}
			sr.applyEvents(events)
			// Only one nomination needed
			break
		}
	}
	return nil
}

// processVerdictPhase handles voting on the nominated player
func (sr *SimulationRunner) processVerdictPhase() error {
	for playerID, player := range sr.gameState.Players {
		if !player.IsAlive {
			continue
		}

		persona := sr.personas[playerID]
		if persona == nil {
			continue
		}

		action := persona.DecideAction(*sr.gameState, playerID)
		if action != nil && action.Type == core.ActionSubmitVote {
			events, err := core.ProcessPlayerAction(*sr.gameState, *action, sr.simulatedTime)
			if err != nil {
				continue // Skip invalid actions
			}
			sr.applyEvents(events)
		}
	}
	return nil
}

// processExtensionPhase handles discussion extension voting
func (sr *SimulationRunner) processExtensionPhase() error {
	// Similar to verdict phase
	return sr.processVerdictPhase()
}

// processNightPhase handles night actions (mining, conversion, abilities)
func (sr *SimulationRunner) processNightPhase() error {
	// Collect all night actions
	nightActions := make([]core.Action, 0)

	for playerID, player := range sr.gameState.Players {
		if !player.IsAlive {
			continue
		}

		persona := sr.personas[playerID]
		if persona == nil {
			continue
		}

		action := persona.DecideAction(*sr.gameState, playerID)
		if action != nil {
			nightActions = append(nightActions, *action)
		}
	}

	// Process all night actions
	for _, action := range nightActions {
		events, err := core.ProcessPlayerAction(*sr.gameState, action, sr.simulatedTime)
		if err != nil {
			continue // Skip invalid actions
		}
		sr.applyEvents(events)
	}

	return nil
}

// advanceTime moves the simulation time forward
func (sr *SimulationRunner) advanceTime() {
	sr.simulatedTime = sr.simulatedTime.Add(time.Duration(sr.config.TimeAcceleration) * time.Second)
}

// checkPhaseTransition determines if the current phase should end
func (sr *SimulationRunner) checkPhaseTransition() error {
	if core.IsGamePhaseOver(*sr.gameState, sr.simulatedTime) {
		// Transition to next phase
		nextPhase := sr.determineNextPhase()
		
		phaseEvent := core.Event{
			ID:        fmt.Sprintf("phase_%s_%d", sr.gameState.ID, sr.simulatedTime.UnixNano()),
			Type:      core.EventPhaseChanged,
			GameID:    sr.gameState.ID,
			Timestamp: sr.simulatedTime,
			Payload: map[string]interface{}{
				"previous_phase": string(sr.gameState.Phase.Type),
				"new_phase":      string(nextPhase),
			},
		}
		
		sr.applyEvent(phaseEvent)
	}
	
	return nil
}

// determineNextPhase calculates what the next phase should be
func (sr *SimulationRunner) determineNextPhase() core.PhaseType {
	switch sr.gameState.Phase.Type {
	case core.PhaseLobby:
		return core.PhaseSitrep
	case core.PhaseSitrep:
		return core.PhaseNomination
	case core.PhaseNomination:
		return core.PhaseVerdict
	case core.PhaseVerdict:
		return core.PhaseExtension
	case core.PhaseExtension:
		return core.PhaseNight
	case core.PhaseNight:
		return core.PhaseSitrep
	default:
		return core.PhaseSitrep
	}
}

// applyEvent applies a single event to the game state and logs it
func (sr *SimulationRunner) applyEvent(event core.Event) {
	*sr.gameState = core.ApplyEvent(*sr.gameState, event)
	sr.eventLog = append(sr.eventLog, event)
}

// applyEvents applies multiple events to the game state and logs them
func (sr *SimulationRunner) applyEvents(events []core.Event) {
	for _, event := range events {
		sr.applyEvent(event)
	}
}

// generateResult creates the final simulation result
func (sr *SimulationRunner) generateResult() *SimulationResult {
	duration := sr.simulatedTime.Sub(sr.gameState.CreatedAt)
	
	// Extract winner information
	winner := "UNKNOWN"
	condition := "UNKNOWN"
	if sr.gameState.WinCondition != nil {
		winner = sr.gameState.WinCondition.Winner
		condition = sr.gameState.WinCondition.Condition
	}

	// Generate player statistics
	playerStats := make(map[string]PlayerStat)
	for playerID, player := range sr.gameState.Players {
		persona := sr.personas[playerID]
		personaType := "unknown"
		if persona != nil {
			personaType = persona.GetID()
		}

		// Count actions and votes from event log
		actionsPerformed := 0
		votesCast := 0
		messagesSent := 0
		
		for _, event := range sr.eventLog {
			if event.PlayerID == playerID {
				switch event.Type {
				case core.EventVoteCast:
					votesCast++
					actionsPerformed++
				case core.EventChatMessage:
					messagesSent++
					actionsPerformed++
				case core.EventMiningAttempted, core.EventNightActionSubmitted:
					actionsPerformed++
				}
			}
		}

		playerStats[playerID] = PlayerStat{
			PlayerID:         playerID,
			PersonaType:      personaType,
			Alignment:        player.Alignment,
			Survived:         player.IsAlive,
			FinalTokens:      player.Tokens,
			FinalMilestones:  player.ProjectMilestones,
			ActionsPerformed: actionsPerformed,
			VotesCast:        votesCast,
			MessagesExempt:   messagesSent,
		}
	}

	return &SimulationResult{
		GameID:      sr.gameState.ID,
		Winner:      winner,
		Condition:   condition,
		Duration:    duration,
		DayNumber:   sr.gameState.DayNumber,
		EventLog:    sr.eventLog,
		FinalState:  *sr.gameState,
		PlayerStats: playerStats,
	}
}