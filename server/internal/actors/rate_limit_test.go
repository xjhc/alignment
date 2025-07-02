package actors

import (
	"context"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/xjhc/alignment/core"
)

func TestPlayerActor_ChatRateLimit(t *testing.T) {
	// Create a mock websocket connection
	conn := &websocket.Conn{}
	
	// Create a PlayerActor
	ctx := context.Background()
	actor := NewPlayerActor(ctx, "test-player", "Test Player", "👤", "test-token", conn)
	
	// Test that the rate limiter was initialized
	if actor.chatLimiter == nil {
		t.Fatal("Chat limiter was not initialized")
	}
	
	
	// Set up a mock game state
	actor.gameID = "test-game"
	
	// Test rate limiting by calling the action multiple times quickly
	// Now rate limiting applies to batches, not individual messages
	sent := 0
	for i := 0; i < 10; i++ {
		// Check if the limiter allows the batch action
		if actor.chatLimiter.Allow() {
			sent++
		}
	}
	
	// The limiter should allow the initial burst (3 batches) but then start rejecting
	if sent > 3 {
		t.Errorf("Rate limiter allowed too many batches: expected <= 3, got %d", sent)
	}
	
	if sent == 0 {
		t.Error("Rate limiter rejected all batches, expected to allow at least the burst capacity")
	}
}

func TestPlayerActor_RateLimitRefill(t *testing.T) {
	// Create a mock websocket connection
	conn := &websocket.Conn{}
	
	// Create a PlayerActor
	ctx := context.Background()
	actor := NewPlayerActor(ctx, "test-player", "Test Player", "👤", "test-token", conn)
	
	// Exhaust the initial burst capacity
	for i := 0; i < 3; i++ {
		if !actor.chatLimiter.Allow() {
			t.Fatal("Expected initial burst to be allowed")
		}
	}
	
	// Next message should be rejected
	if actor.chatLimiter.Allow() {
		t.Error("Expected message to be rate limited after burst")
	}
	
	// Wait for token refill (rate is 0.5/sec, so wait 2.1 seconds for 1 token)
	time.Sleep(2100 * time.Millisecond)
	
	// Now a message should be allowed again
	if !actor.chatLimiter.Allow() {
		t.Error("Expected message to be allowed after token refill")
	}
}

func TestPlayerActor_BulkMessageProcessing(t *testing.T) {
	// Create a mock websocket connection
	conn := &websocket.Conn{}
	
	// Create a PlayerActor
	ctx := context.Background()
	actor := NewPlayerActor(ctx, "test-player", "Test Player", "👤", "test-token", conn)
	
	// Test bulk message processing
	testCases := []struct {
		name          string
		payload       map[string]interface{}
		expectError   bool
		expectedMsgs  []string
	}{
		{
			name: "Valid bulk messages",
			payload: map[string]interface{}{
				"messages": []interface{}{"Hello", "World", "Test"},
			},
			expectError:  false,
			expectedMsgs: []string{"Hello", "World", "Test"},
		},
		{
			name: "Single message (backward compatibility)",
			payload: map[string]interface{}{
				"message": "Single message",
			},
			expectError:  false,
			expectedMsgs: []string{"Single message"},
		},
		{
			name: "Empty messages array",
			payload: map[string]interface{}{
				"messages": []interface{}{},
			},
			expectError: true,
		},
		{
			name: "Too many messages",
			payload: map[string]interface{}{
				"messages": []interface{}{"1", "2", "3", "4", "5", "6"}, // 6 messages > 5 limit
			},
			expectError: true,
		},
		{
			name: "Invalid message type",
			payload: map[string]interface{}{
				"messages": []interface{}{"valid", 123, "another"},
			},
			expectError: true,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			action := core.Action{
				Type:     core.ActionSendMessage,
				PlayerID: "test-player",
				GameID:   "test-game",
				Payload:  tc.payload,
			}
			
			err := actor.processBulkMessages(&action)
			
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			
			if !tc.expectError {
				// Check that messages were properly processed
				if messages, ok := action.Payload["messages"].([]string); ok {
					if len(messages) != len(tc.expectedMsgs) {
						t.Errorf("Expected %d messages, got %d", len(tc.expectedMsgs), len(messages))
					}
					for i, expected := range tc.expectedMsgs {
						if i < len(messages) && messages[i] != expected {
							t.Errorf("Expected message %d to be %q, got %q", i, expected, messages[i])
						}
					}
				} else {
					t.Error("Expected messages array in payload")
				}
			}
		})
	}
}