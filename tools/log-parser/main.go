
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// LogEntry represents a structured log entry from the server
type LogEntry struct {
	Level    string                 `json:"level"`
	Time     string                 `json:"time"`
	Message  string                 `json:"msg"`
	GameID   string                 `json:"game_id,omitempty"`
	PlayerID string                 `json:"player_id,omitempty"`
	Action   string                 `json:"action,omitempty"`
	Error    string                 `json:"error,omitempty"`
	Details  map[string]interface{} `json:"details,omitempty"`
}

// PlayerTimeline holds a player's actions and events in chronological order
type PlayerTimeline struct {
	PlayerID string
	Events   []LogEntry
}

// GameSession aggregates all log entries for a single game
type GameSession struct {
	GameID    string
	StartTime time.Time
	EndTime   time.Time
	Entries   []LogEntry
}

func main() {
	// CLI flags
	logFile := flag.String("file", "", "Path to the log file to parse")
	gameID := flag.String("game", "", "Filter by a specific game ID")
	playerID := flag.String("player", "", "Filter by a specific player ID")
	level := flag.String("level", "", "Filter by log level (info, warn, error)")

	flag.Parse()

	if *logFile == "" {
		fmt.Println("Usage: go run main.go --file <path_to_log.json>")
		os.Exit(1)
	}

	// Read and parse the log file
	sessions, err := parseLogFile(*logFile)
	if err != nil {
		fmt.Printf("Error parsing log file: %v\n", err)
		os.Exit(1)
	}

	// Filter sessions based on flags
	filteredSessions := filterSessions(sessions, *gameID)

	// Display results
	if len(filteredSessions) == 0 {
		fmt.Println("No matching game sessions found.")
		return
	}

	for _, session := range filteredSessions {
		fmt.Printf("=== Game: %s ===\n", session.GameID)
		fmt.Printf("Duration: %v\n", session.EndTime.Sub(session.StartTime))

		for _, entry := range session.Entries {
			// Apply filters
			if *level != "" && strings.ToLower(entry.Level) != strings.ToLower(*level) {
				continue
			}

			if *playerID != "" && entry.PlayerID != *playerID {
				continue
			}

			// Print formatted log entry
			fmt.Printf("[%s] [%-5s] %-15s: %s\n",
				entry.Time,
				entry.Level,
				entry.PlayerID,
				entry.Message)

			if entry.Error != "" {
				fmt.Printf("  ERROR: %s\n", entry.Error)
			}
		}

		fmt.Println()
	}
}

func parseLogFile(filePath string) (map[string]*GameSession, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	sessions := make(map[string]*GameSession)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var entry LogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			// Ignore non-JSON lines
			continue
		}

		if entry.GameID != "" {
			// Find or create session for this game
			session, exists := sessions[entry.GameID]
			if !exists {```
				session = &GameSession{
					GameID: entry.GameID,
				}
				sessions[entry.GameID] = session
			}

			// Add entry and update start/end times
			session.Entries = append(session.Entries, entry)

			entryTime, err := time.Parse(time.RFC3339, entry.Time)
			if err == nil {
				if session.StartTime.IsZero() || entryTime.Before(session.StartTime) {
					session.StartTime = entryTime
				}
				if entryTime.After(session.EndTime) {
					session.EndTime = entryTime
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	// Sort entries by time for each session
	for _, session := range sessions {
		sort.Slice(session.Entries, func(i, j int) bool {
			timeA, _ := time.Parse(time.RFC3339, session.Entries[i].Time)
			timeB, _ := time.Parse(time.RFC3339, session.Entries[j].Time)
			return timeA.Before(timeB)
		})
	}

	return sessions, nil
}

func filterSessions(sessions map[string]*GameSession, gameID string) []*GameSession {
	if gameID == "" {
		// Return all sessions if no filter
		var all []*GameSession
		for _, s := range sessions {
			all = append(all, s)
		}
		return all
	}

	if session, exists := sessions[gameID]; exists {
		return []*GameSession{session}
	}

	return []*GameSession{}
}