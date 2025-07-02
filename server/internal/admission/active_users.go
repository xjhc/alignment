package admission

import (
	"fmt"
	"sync"
	"time"
)

// ActiveUserSession represents a user's active session
type ActiveUserSession struct {
	UserID    string    `json:"user_id"`
	GameID    string    `json:"game_id"`
	PlayerID  string    `json:"player_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ActiveUserTracker tracks which users are currently in active game sessions
type ActiveUserTracker struct {
	// Map from user_id to their active session
	activeSessions map[string]*ActiveUserSession
	mutex          sync.RWMutex
}

// Ensure ActiveUserTracker implements the required interface methods
var _ interface {
	RemoveUserSessionByGameAndPlayer(gameID, playerID string)
} = (*ActiveUserTracker)(nil)

// NewActiveUserTracker creates a new active user tracker
func NewActiveUserTracker() *ActiveUserTracker {
	return &ActiveUserTracker{
		activeSessions: make(map[string]*ActiveUserSession),
	}
}

// CheckSingleSession checks if a user can join/create a game without violating single-session rule
func (aut *ActiveUserTracker) CheckSingleSession(userID, gameID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	aut.mutex.RLock()
	defer aut.mutex.RUnlock()

	// Check if user has an active session
	existingSession, exists := aut.activeSessions[userID]
	if !exists {
		// User has no active session, they can join
		return nil
	}

	// If they're trying to rejoin the same game, that's allowed
	if existingSession.GameID == gameID {
		return nil
	}

	// User is already in a different game, reject
	return fmt.Errorf("user is already in an active game session (game: %s)", existingSession.GameID)
}

// AddUserSession adds a user to an active session
func (aut *ActiveUserTracker) AddUserSession(userID, gameID, playerID string) error {
	if userID == "" || gameID == "" || playerID == "" {
		return fmt.Errorf("userID, gameID, and playerID are required")
	}

	aut.mutex.Lock()
	defer aut.mutex.Unlock()

	session := &ActiveUserSession{
		UserID:    userID,
		GameID:    gameID,
		PlayerID:  playerID,
		Timestamp: time.Now(),
	}

	aut.activeSessions[userID] = session
	return nil
}

// RemoveUserSession removes a user from active sessions
func (aut *ActiveUserTracker) RemoveUserSession(userID string) {
	if userID == "" {
		return
	}

	aut.mutex.Lock()
	defer aut.mutex.Unlock()

	delete(aut.activeSessions, userID)
}

// RemoveUserSessionByGameAndPlayer removes a user session by game ID and player ID
func (aut *ActiveUserTracker) RemoveUserSessionByGameAndPlayer(gameID, playerID string) {
	if gameID == "" || playerID == "" {
		return
	}

	aut.mutex.Lock()
	defer aut.mutex.Unlock()

	// Find the user with this game/player combination
	for userID, session := range aut.activeSessions {
		if session.GameID == gameID && session.PlayerID == playerID {
			delete(aut.activeSessions, userID)
			break
		}
	}
}

// GetUserSession gets a user's active session
func (aut *ActiveUserTracker) GetUserSession(userID string) (*ActiveUserSession, bool) {
	aut.mutex.RLock()
	defer aut.mutex.RUnlock()

	session, exists := aut.activeSessions[userID]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	sessionCopy := *session
	return &sessionCopy, true
}

// GetActiveSessionCount returns the number of active sessions
func (aut *ActiveUserTracker) GetActiveSessionCount() int {
	aut.mutex.RLock()
	defer aut.mutex.RUnlock()

	return len(aut.activeSessions)
}

// GetAllActiveSessions returns a snapshot of all active sessions
func (aut *ActiveUserTracker) GetAllActiveSessions() map[string]*ActiveUserSession {
	aut.mutex.RLock()
	defer aut.mutex.RUnlock()

	// Return a deep copy to avoid race conditions
	sessions := make(map[string]*ActiveUserSession)
	for userID, session := range aut.activeSessions {
		sessionCopy := *session
		sessions[userID] = &sessionCopy
	}

	return sessions
}

// CleanupStaleSessions removes sessions older than the specified duration
func (aut *ActiveUserTracker) CleanupStaleSessions(maxAge time.Duration) int {
	aut.mutex.Lock()
	defer aut.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var staleUserIDs []string

	for userID, session := range aut.activeSessions {
		if session.Timestamp.Before(cutoff) {
			staleUserIDs = append(staleUserIDs, userID)
		}
	}

	for _, userID := range staleUserIDs {
		delete(aut.activeSessions, userID)
	}

	return len(staleUserIDs)
}