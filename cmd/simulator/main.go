package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/xjhc/alignment/simulator"
)

// AggregatedResults contains statistics from multiple simulation runs
type AggregatedResults struct {
	TotalRuns         int                        `json:"total_runs"`
	SuccessfulRuns    int                        `json:"successful_runs"`
	FailedRuns        int                        `json:"failed_runs"`
	HumanWins         int                        `json:"human_wins"`
	AIWins            int                        `json:"ai_wins"`
	HumanWinRate      float64                    `json:"human_win_rate"`
	AIWinRate         float64                    `json:"ai_win_rate"`
	AverageGameDays   float64                    `json:"average_game_days"`
	AverageDuration   time.Duration              `json:"average_duration"`
	WinsByCondition   map[string]int             `json:"wins_by_condition"`
	PersonaStats      map[string]PersonaStats    `json:"persona_stats"`
	BalanceMetrics    BalanceMetrics             `json:"balance_metrics"`
	SimulationRuns    []*simulator.SimulationResult `json:"simulation_runs,omitempty"`
}

// PersonaStats tracks performance by persona type
type PersonaStats struct {
	TotalGames       int     `json:"total_games"`
	Wins             int     `json:"wins"`
	WinRate          float64 `json:"win_rate"`
	AverageSurvival  float64 `json:"average_survival_rate"`
	AverageTokens    float64 `json:"average_tokens"`
	AverageMilestones float64 `json:"average_milestones"`
}

// BalanceMetrics provides insights into game balance
type BalanceMetrics struct {
	IsBalanced           bool    `json:"is_balanced"`
	WinRateDeviation     float64 `json:"win_rate_deviation"`     // How far from 50/50 split
	DurationConsistency  float64 `json:"duration_consistency"`   // Standard deviation of game length
	PersonaBalance       bool    `json:"persona_balance"`        // Whether all personas perform reasonably
	RecommendedAdjustment string  `json:"recommended_adjustment"` // Suggested balance changes
}

// CLI flags
var (
	runs        = flag.Int("runs", 100, "Number of simulation runs to execute")
	concurrency = flag.Int("concurrency", runtime.NumCPU(), "Number of concurrent simulations")
	players     = flag.Int("players", 6, "Number of players per game")
	aiCount     = flag.Int("ai", 2, "Number of AI players per game")
	output      = flag.String("output", "", "Output file for detailed results (JSON)")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	seedBase    = flag.Int64("seed", time.Now().UnixNano(), "Base seed for reproducible results")
	ciMode      = flag.Bool("ci", false, "CI mode: fail if balance metrics indicate regression")
)

func main() {
	flag.Parse()

	fmt.Printf("🎮 Alignment Game Simulation Runner\n")
	fmt.Printf("Running %d simulations with %d players (%d AI) using %d goroutines...\n", 
		*runs, *players, *aiCount, *concurrency)

	start := time.Now()

	// Run simulations
	results, err := runBatchSimulations(*runs, *concurrency)
	if err != nil {
		log.Fatalf("Simulation batch failed: %v", err)
	}

	// Aggregate results
	aggregated := aggregateResults(results)
	
	elapsed := time.Since(start)
	fmt.Printf("\n✅ Completed %d simulations in %v\n", aggregated.SuccessfulRuns, elapsed)

	// Print summary
	printSummary(aggregated)

	// Save detailed results if requested
	if *output != "" {
		if err := saveResults(aggregated, *output); err != nil {
			log.Printf("Warning: Failed to save results to %s: %v", *output, err)
		} else {
			fmt.Printf("📊 Detailed results saved to %s\n", *output)
		}
	}

	// CI mode: check for balance regressions and save results
	if *ciMode {
		// Always save results in CI mode for later analysis
		ciOutputFile := "simulation-results.json"
		if err := saveResults(aggregated, ciOutputFile); err != nil {
			log.Printf("Warning: Failed to save CI results to %s: %v", ciOutputFile, err)
		}

		if err := checkBalanceRegression(aggregated); err != nil {
			fmt.Printf("❌ BALANCE REGRESSION DETECTED: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Balance metrics within acceptable range\n")
	}
}

// runBatchSimulations executes multiple simulations concurrently
func runBatchSimulations(totalRuns, concurrency int) ([]*simulator.SimulationResult, error) {
	results := make([]*simulator.SimulationResult, 0, totalRuns)
	resultsChan := make(chan *simulator.SimulationResult, concurrency)
	errorChan := make(chan error, concurrency)
	
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, concurrency)

	// Start workers
	for i := 0; i < totalRuns; i++ {
		wg.Add(1)
		go func(runIndex int) {
			defer wg.Done()
			
			semaphore <- struct{}{} // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			config := simulator.SimulationConfig{
				PlayerCount: *players,
				AICount:     *aiCount,
				PersonaWeights: map[string]float64{
					"cautious_human":  0.4,
					"aggressive_human": 0.4,
					"deceptive_ai":    1.0,
				},
				Seed:             *seedBase + int64(runIndex),
				TimeAcceleration: 1, // Fast simulation
			}

			runner := simulator.NewSimulationRunner(config)
			result, err := runner.RunSimulation()
			
			if err != nil {
				if *verbose {
					log.Printf("Simulation %d failed: %v", runIndex+1, err)
				}
				errorChan <- fmt.Errorf("run %d: %w", runIndex+1, err)
				return
			}

			if *verbose {
				fmt.Printf("Simulation %d: %s wins by %s on day %d\n", 
					runIndex+1, result.Winner, result.Condition, result.DayNumber)
			}

			resultsChan <- result
		}(i)
	}

	// Close channels when all workers are done
	go func() {
		wg.Wait()
		close(resultsChan)
		close(errorChan)
	}()

	// Collect results
	var errors []error
	for {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				resultsChan = nil
			} else {
				results = append(results, result)
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else {
				errors = append(errors, err)
			}
		}
		
		if resultsChan == nil && errorChan == nil {
			break
		}
	}

	if len(errors) > 0 && len(results) == 0 {
		return nil, fmt.Errorf("all simulations failed: %v", errors[0])
	}

	return results, nil
}

// aggregateResults combines individual simulation results into summary statistics
func aggregateResults(results []*simulator.SimulationResult) *AggregatedResults {
	aggregated := &AggregatedResults{
		TotalRuns:      len(results),
		SuccessfulRuns: len(results),
		WinsByCondition: make(map[string]int),
		PersonaStats:   make(map[string]PersonaStats),
	}

	if len(results) == 0 {
		return aggregated
	}

	totalDays := 0
	totalDuration := time.Duration(0)
	durations := make([]float64, 0, len(results))

	// Track persona performance
	personaGames := make(map[string]int)
	personaWins := make(map[string]int)
	personaSurvival := make(map[string]int)
	personaTokens := make(map[string][]int)
	personaMilestones := make(map[string][]int)

	for _, result := range results {
		// Count wins
		if result.Winner == "HUMANS" {
			aggregated.HumanWins++
		} else if result.Winner == "AI" {
			aggregated.AIWins++
		}

		// Track win conditions
		aggregated.WinsByCondition[result.Condition]++

		// Accumulate duration and days
		totalDays += result.DayNumber
		totalDuration += result.Duration
		durations = append(durations, result.Duration.Seconds())

		// Process player stats
		for _, playerStat := range result.PlayerStats {
			persona := playerStat.PersonaType
			personaGames[persona]++

			// Track survival
			if playerStat.Survived {
				personaSurvival[persona]++
			}

			// Track wins (if their alignment won)
			if (playerStat.Alignment == "HUMAN" && result.Winner == "HUMANS") ||
			   (playerStat.Alignment == "ALIGNED" && result.Winner == "AI") {
				personaWins[persona]++
			}

			// Track performance metrics
			if personaTokens[persona] == nil {
				personaTokens[persona] = make([]int, 0)
				personaMilestones[persona] = make([]int, 0)
			}
			personaTokens[persona] = append(personaTokens[persona], playerStat.FinalTokens)
			personaMilestones[persona] = append(personaMilestones[persona], playerStat.FinalMilestones)
		}
	}

	// Calculate win rates
	if aggregated.SuccessfulRuns > 0 {
		aggregated.HumanWinRate = float64(aggregated.HumanWins) / float64(aggregated.SuccessfulRuns)
		aggregated.AIWinRate = float64(aggregated.AIWins) / float64(aggregated.SuccessfulRuns)
		aggregated.AverageGameDays = float64(totalDays) / float64(aggregated.SuccessfulRuns)
		aggregated.AverageDuration = totalDuration / time.Duration(aggregated.SuccessfulRuns)
	}

	// Calculate persona statistics
	for persona, games := range personaGames {
		if games == 0 {
			continue
		}

		wins := personaWins[persona]
		survival := personaSurvival[persona]
		tokens := personaTokens[persona]
		milestones := personaMilestones[persona]

		// Calculate averages
		avgTokens := 0.0
		avgMilestones := 0.0
		if len(tokens) > 0 {
			sum := 0
			for _, t := range tokens {
				sum += t
			}
			avgTokens = float64(sum) / float64(len(tokens))
		}
		if len(milestones) > 0 {
			sum := 0
			for _, m := range milestones {
				sum += m
			}
			avgMilestones = float64(sum) / float64(len(milestones))
		}

		aggregated.PersonaStats[persona] = PersonaStats{
			TotalGames:       games,
			Wins:             wins,
			WinRate:          float64(wins) / float64(games),
			AverageSurvival:  float64(survival) / float64(games),
			AverageTokens:    avgTokens,
			AverageMilestones: avgMilestones,
		}
	}

	// Calculate balance metrics
	aggregated.BalanceMetrics = calculateBalanceMetrics(aggregated, durations)

	// Include detailed results if verbose
	if *verbose || *output != "" {
		aggregated.SimulationRuns = results
	}

	return aggregated
}

// calculateBalanceMetrics analyzes game balance
func calculateBalanceMetrics(aggregated *AggregatedResults, durations []float64) BalanceMetrics {
	metrics := BalanceMetrics{}

	// Win rate balance (ideal is 50/50)
	targetWinRate := 0.5
	metrics.WinRateDeviation = abs(aggregated.HumanWinRate - targetWinRate)
	
	// Consider balanced if within 5% of 50/50 split
	balanceThreshold := 0.05
	metrics.IsBalanced = metrics.WinRateDeviation <= balanceThreshold

	// Duration consistency (measure of game length variance)
	if len(durations) > 1 {
		mean := 0.0
		for _, d := range durations {
			mean += d
		}
		mean /= float64(len(durations))

		variance := 0.0
		for _, d := range durations {
			variance += (d - mean) * (d - mean)
		}
		variance /= float64(len(durations))
		metrics.DurationConsistency = variance
	}

	// Persona balance check
	metrics.PersonaBalance = true
	for persona, stats := range aggregated.PersonaStats {
		// Check if any persona is significantly over/under performing
		if stats.WinRate < 0.3 || stats.WinRate > 0.7 {
			metrics.PersonaBalance = false
			break
		}
		// Check survival rates
		if stats.AverageSurvival < 0.2 || stats.AverageSurvival > 0.8 {
			metrics.PersonaBalance = false
			break
		}
		
		_ = persona // Suppress unused variable warning
	}

	// Generate recommendations
	if !metrics.IsBalanced {
		if aggregated.HumanWinRate > 0.55 {
			metrics.RecommendedAdjustment = "AI conversion rate too low or AI detection too easy"
		} else if aggregated.AIWinRate > 0.55 {
			metrics.RecommendedAdjustment = "AI conversion rate too high or human roles too weak"
		}
	} else {
		metrics.RecommendedAdjustment = "Game balance appears optimal"
	}

	return metrics
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// printSummary displays a human-readable summary of results
func printSummary(aggregated *AggregatedResults) {
	fmt.Printf("\n📊 SIMULATION SUMMARY\n")
	fmt.Printf("==========================================\n")
	fmt.Printf("Total Runs:       %d\n", aggregated.TotalRuns)
	fmt.Printf("Successful:       %d\n", aggregated.SuccessfulRuns)
	fmt.Printf("Failed:           %d\n", aggregated.FailedRuns)
	fmt.Printf("\n🏆 WIN STATISTICS\n")
	fmt.Printf("Human Wins:       %d (%.1f%%)\n", aggregated.HumanWins, aggregated.HumanWinRate*100)
	fmt.Printf("AI Wins:          %d (%.1f%%)\n", aggregated.AIWins, aggregated.AIWinRate*100)
	fmt.Printf("Average Game Days: %.1f\n", aggregated.AverageGameDays)
	fmt.Printf("Average Duration:  %v\n", aggregated.AverageDuration)

	fmt.Printf("\n🎯 WIN CONDITIONS\n")
	for condition, count := range aggregated.WinsByCondition {
		percentage := float64(count) / float64(aggregated.SuccessfulRuns) * 100
		fmt.Printf("%-15s: %d (%.1f%%)\n", condition, count, percentage)
	}

	fmt.Printf("\n🤖 PERSONA PERFORMANCE\n")
	for persona, stats := range aggregated.PersonaStats {
		fmt.Printf("%-20s: Win Rate %.1f%%, Survival %.1f%%, Avg Tokens %.1f, Avg Milestones %.1f\n",
			persona, stats.WinRate*100, stats.AverageSurvival*100, stats.AverageTokens, stats.AverageMilestones)
	}

	fmt.Printf("\n⚖️  BALANCE ANALYSIS\n")
	balanceIcon := "✅"
	if !aggregated.BalanceMetrics.IsBalanced {
		balanceIcon = "⚠️"
	}
	fmt.Printf("Game Balance:     %s %s\n", balanceIcon, boolToString(aggregated.BalanceMetrics.IsBalanced))
	fmt.Printf("Win Rate Deviation: %.1f%% from ideal 50/50\n", aggregated.BalanceMetrics.WinRateDeviation*100)
	fmt.Printf("Persona Balance:   %s\n", boolToString(aggregated.BalanceMetrics.PersonaBalance))
	fmt.Printf("Recommendation:    %s\n", aggregated.BalanceMetrics.RecommendedAdjustment)
}

func boolToString(b bool) string {
	if b {
		return "Good"
	}
	return "Issues Detected"
}

// saveResults writes detailed results to a JSON file
func saveResults(aggregated *AggregatedResults, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(aggregated)
}

// checkBalanceRegression determines if the simulation results indicate a balance regression
func checkBalanceRegression(aggregated *AggregatedResults) error {
	// Define acceptable ranges for CI
	const (
		maxWinRateDeviation = 0.10  // 10% deviation from 50/50
		minSurvivalRate     = 0.20  // At least 20% survival rate for any persona
		maxSurvivalRate     = 0.80  // At most 80% survival rate for any persona
		minWinRate          = 0.25  // At least 25% win rate for any persona
		maxWinRate          = 0.75  // At most 75% win rate for any persona
	)

	// Check overall win rate balance
	if aggregated.BalanceMetrics.WinRateDeviation > maxWinRateDeviation {
		return fmt.Errorf("win rate deviation %.1f%% exceeds threshold %.1f%%",
			aggregated.BalanceMetrics.WinRateDeviation*100, maxWinRateDeviation*100)
	}

	// Check persona balance
	for persona, stats := range aggregated.PersonaStats {
		if stats.AverageSurvival < minSurvivalRate {
			return fmt.Errorf("persona %s has survival rate %.1f%% below threshold %.1f%%",
				persona, stats.AverageSurvival*100, minSurvivalRate*100)
		}
		if stats.AverageSurvival > maxSurvivalRate {
			return fmt.Errorf("persona %s has survival rate %.1f%% above threshold %.1f%%",
				persona, stats.AverageSurvival*100, maxSurvivalRate*100)
		}
		if stats.WinRate < minWinRate {
			return fmt.Errorf("persona %s has win rate %.1f%% below threshold %.1f%%",
				persona, stats.WinRate*100, minWinRate*100)
		}
		if stats.WinRate > maxWinRate {
			return fmt.Errorf("persona %s has win rate %.1f%% above threshold %.1f%%",
				persona, stats.WinRate*100, maxWinRate*100)
		}
	}

	return nil
}