package main

import (
	"context"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestIPRateLimiter_CleanupStale(t *testing.T) {
	// Create a rate limiter with a burst of 5
	limiter := NewIPRateLimiter(rate.Every(time.Second), 5)

	// Add multiple IP addresses to the limiter
	ips := []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4", "192.168.1.5"}
	for _, ip := range ips {
		_ = limiter.GetLimiter(ip)
	}

	// Verify initial state - should have 5 limiters
	limiter.mu.RLock()
	initialCount := len(limiter.limiters)
	limiter.mu.RUnlock()
	
	if initialCount != 5 {
		t.Errorf("Expected 5 limiters, got %d", initialCount)
	}

	// Use some limiters to drain their tokens
	for i := 0; i < 3; i++ {
		limiter.GetLimiter("192.168.1.1").Allow()
		limiter.GetLimiter("192.168.1.2").Allow()
	}

	// Call CleanupStale - should remove unused limiters (those with full tokens)
	limiter.CleanupStale()

	// Check the size after cleanup
	limiter.mu.RLock()
	afterCleanupCount := len(limiter.limiters)
	limiter.mu.RUnlock()

	// Should have removed limiters that weren't used (had full tokens)
	// The used limiters (192.168.1.1 and 192.168.1.2) should remain
	if afterCleanupCount >= initialCount {
		t.Errorf("Expected cleanup to reduce limiter count, but got %d (was %d)", afterCleanupCount, initialCount)
	}

	// Verify that used limiters are still present
	limiter.mu.RLock()
	_, exists1 := limiter.limiters["192.168.1.1"]
	_, exists2 := limiter.limiters["192.168.1.2"]
	limiter.mu.RUnlock()

	if !exists1 || !exists2 {
		t.Error("Expected used limiters to remain after cleanup")
	}
}

func TestIPRateLimiter_GetLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 10)
	
	// Test getting a limiter for a new IP
	ip := "192.168.1.100"
	rateLimiter := limiter.GetLimiter(ip)
	
	if rateLimiter == nil {
		t.Error("Expected GetLimiter to return a rate limiter")
	}
	
	// Test getting the same limiter again
	rateLimiter2 := limiter.GetLimiter(ip)
	
	if rateLimiter != rateLimiter2 {
		t.Error("Expected GetLimiter to return the same limiter for the same IP")
	}
	
	// Verify limiter is stored in map
	limiter.mu.RLock()
	_, exists := limiter.limiters[ip]
	limiter.mu.RUnlock()
	
	if !exists {
		t.Error("Expected limiter to be stored in map")
	}
}

func TestIPRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Millisecond*10), 1)
	
	// Test concurrent access to the rate limiter
	done := make(chan bool)
	
	// Start multiple goroutines accessing the same IP
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = limiter.GetLimiter("192.168.1.1")
			}
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify only one limiter was created
	limiter.mu.RLock()
	count := len(limiter.limiters)
	limiter.mu.RUnlock()
	
	if count != 1 {
		t.Errorf("Expected 1 limiter, got %d", count)
	}
}

func TestServer_startRateLimiterCleanup_ContextCancellation(t *testing.T) {
	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create a server with the context
	server := &Server{
		ctx:                  ctx,
		joinLobbyRateLimiter: NewIPRateLimiter(rate.Every(time.Second), 10),
	}
	
	// Start the cleanup routine
	done := make(chan bool)
	go func() {
		server.startRateLimiterCleanup()
		done <- true
	}()
	
	// Cancel the context after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()
	
	// Wait for the cleanup routine to stop
	select {
	case <-done:
		// Success - the cleanup routine stopped when context was cancelled
	case <-time.After(1 * time.Second):
		t.Error("Cleanup routine did not stop after context cancellation")
	}
}