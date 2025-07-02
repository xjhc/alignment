package actors

import (
	"context"
	"testing"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/semaphore"
)

func TestPlayerActor_GeneralRateLimit(t *testing.T) {
	// Create a mock websocket connection
	conn := &websocket.Conn{}
	
	// Create a PlayerActor
	ctx := context.Background()
	actor := NewPlayerActor(ctx, "test-player", "Test Player", "👤", "test-token", conn)
	
	// Test that the general rate limiter was initialized
	if actor.generalLimiter == nil {
		t.Fatal("General rate limiter was not initialized")
	}
	
	// Test rate limiting by calling the action handler multiple times quickly
	sent := 0
	for i := 0; i < 25; i++ { // Try more than the burst capacity (20)
		// Check if the limiter allows the action
		if actor.generalLimiter.Allow() {
			sent++
		}
	}
	
	// The limiter should allow the burst (20) but then start rejecting
	if sent > 20 {
		t.Errorf("General rate limiter allowed too many actions: expected <= 20, got %d", sent)
	}
	
	if sent < 15 { // Should allow at least most of the burst
		t.Errorf("General rate limiter rejected too many actions, expected to allow at least 15, got %d", sent)
	}
}

func TestPlayerActor_ActionSemaphore(t *testing.T) {
	// Create a mock websocket connection
	conn := &websocket.Conn{}
	
	// Create a PlayerActor
	ctx := context.Background()
	actor := NewPlayerActor(ctx, "test-player", "Test Player", "👤", "test-token", conn)
	
	// Create a small semaphore for testing
	sem := semaphore.NewWeighted(2)
	actor.SetActionSemaphore(sem)
	
	// Test that the semaphore was set
	if actor.actionSemaphore == nil {
		t.Fatal("Action semaphore was not set")
	}
	
	// Acquire all semaphore permits
	sem.Acquire(context.Background(), 2)
	
	// Now the semaphore should be full and actions should be rejected
	// We can't easily test handleClientAction without a lot of setup,
	// but we can test that the semaphore is correctly configured
	if sem.TryAcquire(1) {
		t.Error("Expected semaphore to be at capacity")
	}
}

func TestInputValidation(t *testing.T) {
	// Test basic security functionality
	t.Run("Player name validation", func(t *testing.T) {
		// Valid names
		if err := validatePlayerName("John_Doe"); err != nil {
			t.Errorf("Expected valid name to pass: %v", err)
		}
		
		// Invalid - empty
		if err := validatePlayerName(""); err == nil {
			t.Error("Expected empty name to be rejected")
		}
		
		// Invalid - too long
		if err := validatePlayerName(string(make([]byte, 100))); err == nil {
			t.Error("Expected long name to be rejected")
		}
	})
	
	t.Run("Chat message validation", func(t *testing.T) {
		// Valid message
		if err := validateChatMessage("Hello world"); err != nil {
			t.Errorf("Expected valid message to pass: %v", err)
		}
		
		// Invalid - empty
		if err := validateChatMessage(""); err == nil {
			t.Error("Expected empty message to be rejected")
		}
		
		// Invalid - too long
		if err := validateChatMessage(string(make([]byte, 600))); err == nil {
			t.Error("Expected long message to be rejected")
		}
	})
}

func TestSanitizeString(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"Normal string", "Hello world", 20, "Hello world"},
		{"String with control chars", "Hello\x00\x01world", 20, "Helloworld"},
		{"String too long", "This is a very long string", 10, "This is a "},
		{"String with tabs and newlines", "Hello\tworld\n", 20, "Hello\tworld"},
		{"Empty string", "", 10, ""},
		{"Whitespace trimming", "  spaced  ", 20, "spaced"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sanitizeString(tc.input, tc.maxLen)
			if result != tc.expected {
				t.Errorf("sanitizeString() = %q, expected %q", result, tc.expected)
			}
		})
	}
}