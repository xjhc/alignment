// Package admission provides request admission control and waitlist management
package admission

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xjhc/alignment/server/internal/health"
	"github.com/xjhc/alignment/server/internal/logger"
)

const (
	// WaitlistKey is the Redis key for the game creation waitlist
	WaitlistKey = "game_waitlist"
	// WaitlistProcessInterval is how often to process the waitlist
	WaitlistProcessInterval = 10 * time.Second
)

// WaitlistEntry represents an entry in the game creation waitlist
type WaitlistEntry struct {
	UserID    string    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
	Request   string    `json:"request"` // JSON-encoded original request
}

// Controller manages admission control and waitlist processing
type Controller struct {
	healthMonitor *health.Monitor
	redisClient   *redis.Client
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *slog.Logger
}

// NewController creates a new admission controller
func NewController(healthMonitor *health.Monitor, redisClient *redis.Client) *Controller {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &Controller{
		healthMonitor: healthMonitor,
		redisClient:   redisClient,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger.WithField("component", "admission_controller"),
	}
}

// Start begins the waitlist processing goroutine
func (c *Controller) Start() {
	c.logger.Info("Starting admission controller", "process_interval", WaitlistProcessInterval)
	
	go c.processWaitlistLoop()
}

// Stop stops the admission controller
func (c *Controller) Stop() {
	c.logger.Info("Stopping admission controller")
	c.cancel()
}

// CheckAdmission checks if a request should be admitted or waitlisted
func (c *Controller) CheckAdmission(userID string, requestBody []byte) error {
	if c.healthMonitor.IsHealthy() {
		return nil // Allow the request to proceed
	}

	// Server is overloaded, add to waitlist
	entry := WaitlistEntry{
		UserID:    userID,
		Timestamp: time.Now(),
		Request:   string(requestBody),
	}

	entryJSON, err := json.Marshal(entry)
	if err != nil {
		c.logger.Error("Failed to marshal waitlist entry", "error", err, "user_id", userID)
		return fmt.Errorf("internal error processing request")
	}

	// Add to Redis list (LPUSH for FIFO when using RPOP)
	err = c.redisClient.LPush(c.ctx, WaitlistKey, entryJSON).Err()
	if err != nil {
		c.logger.Error("Failed to add user to waitlist", "error", err, "user_id", userID)
		return fmt.Errorf("internal error processing request")
	}

	c.logger.Info("User added to waitlist", "user_id", userID)
	return &WaitlistError{
		Message: "Server is currently overloaded. You have been added to the waitlist and will be processed when capacity is available.",
		UserID:  userID,
	}
}

// WaitlistError represents an error when a user is waitlisted
type WaitlistError struct {
	Message string
	UserID  string
}

func (e *WaitlistError) Error() string {
	return e.Message
}

// IsWaitlistError checks if an error is a waitlist error
func IsWaitlistError(err error) bool {
	_, ok := err.(*WaitlistError)
	return ok
}

// WriteWaitlistResponse writes a 503 Service Unavailable response for waitlisted requests
func WriteWaitlistResponse(w http.ResponseWriter, err *WaitlistError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	
	response := map[string]interface{}{
		"error":   err.Message,
		"user_id": err.UserID,
		"retry_after": 30, // Suggest retrying after 30 seconds
	}
	
	json.NewEncoder(w).Encode(response)
}

// processWaitlistLoop processes the waitlist periodically
func (c *Controller) processWaitlistLoop() {
	ticker := time.NewTicker(WaitlistProcessInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.processWaitlist()
		}
	}
}

// processWaitlist processes entries from the waitlist when server is healthy
func (c *Controller) processWaitlist() {
	if !c.healthMonitor.IsHealthy() {
		c.logger.Debug("Server still overloaded, skipping waitlist processing")
		return
	}

	// Get the length of the waitlist
	length, err := c.redisClient.LLen(c.ctx, WaitlistKey).Result()
	if err != nil {
		c.logger.Error("Failed to get waitlist length", "error", err)
		return
	}

	if length == 0 {
		return // No entries to process
	}

	c.logger.Info("Processing waitlist", "entries", length)

	// Process a batch of entries (limit to avoid overwhelming the server)
	batchSize := int64(10)
	if length < batchSize {
		batchSize = length
	}

	for i := int64(0); i < batchSize; i++ {
		// Pop entry from the right (FIFO)
		entryJSON, err := c.redisClient.RPop(c.ctx, WaitlistKey).Result()
		if err == redis.Nil {
			break // No more entries
		}
		if err != nil {
			c.logger.Error("Failed to pop from waitlist", "error", err)
			break
		}

		var entry WaitlistEntry
		if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
			c.logger.Error("Failed to unmarshal waitlist entry", "error", err)
			continue
		}

		c.logger.Info("Processing waitlist entry", 
			"user_id", entry.UserID, 
			"age", time.Since(entry.Timestamp))

		// Here you would typically send the request to the actual handler
		// For now, we just log that it would be processed
		// In a real implementation, you might use a callback function or message queue
	}
}