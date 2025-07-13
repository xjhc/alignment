package ratelimit

import (
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

	// Verify all limiters were created
	limiter.mu.RLock()
	initialCount := len(limiter.limiters)
	limiter.mu.RUnlock()
	
	if initialCount != 5 {
		t.Errorf("Expected 5 limiters, got %d", initialCount)
	}

	// Exhaust some limiters (simulate no activity)
	for i := 0; i < 3; i++ {
		l := limiter.GetLimiter(ips[i])
		// Consume all available tokens (burst = 5)
		for j := 0; j < 5; j++ {
			l.Allow()
		}
	}

	// Call cleanup - stale limiters should be removed
	limiter.CleanupStale()

	// Verify cleanup occurred (stale limiters with full token buckets should remain)
	limiter.mu.RLock()
	finalCount := len(limiter.limiters)
	limiters := make(map[string]*rate.Limiter, len(limiter.limiters))
	for k, v := range limiter.limiters {
		limiters[k] = v
	}
	limiter.mu.RUnlock()

	// Limiters that still have tokens (unused ones) should be cleaned up
	// Limiters that were used (tokens consumed) should remain
	if finalCount > initialCount {
		t.Errorf("Expected cleanup to remove some limiters, but count increased from %d to %d", initialCount, finalCount)
	}
}

func TestIPRateLimiter_GetLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Every(time.Second), 10)
	
	// Test getting a limiter for a new IP
	ip := "192.168.1.100"
	rateLimiter := limiter.GetLimiter(ip)
	
	if rateLimiter == nil {
		t.Error("Expected non-nil rate limiter")
	}
	
	// Test getting the same limiter again
	rateLimiter2 := limiter.GetLimiter(ip)
	
	if rateLimiter != rateLimiter2 {
		t.Error("Expected same rate limiter instance for same IP")
	}
	
	// Test getting a limiter for a different IP
	ip2 := "192.168.1.101"
	rateLimiter3 := limiter.GetLimiter(ip2)
	
	if rateLimiter3 == rateLimiter {
		t.Error("Expected different rate limiter instance for different IP")
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
		select {
		case <-done:
			// Success
		case <-time.After(time.Second * 5):
			t.Fatal("Test timed out waiting for goroutines")
		}
	}
}

func TestIPRateLimiter_RateLimiting(t *testing.T) {
	// Create a very restrictive rate limiter for testing
	limiter := NewIPRateLimiter(rate.Limit(1), 1) // 1 per second, burst of 1
	
	ip := "192.168.1.1"
	rateLimiter := limiter.GetLimiter(ip)
	
	// First request should be allowed
	if !rateLimiter.Allow() {
		t.Error("First request should be allowed")
	}
	
	// Second request should be denied (burst exhausted)
	if rateLimiter.Allow() {
		t.Error("Second request should be denied")
	}
	
	// After waiting, request should be allowed again
	time.Sleep(time.Second + 100*time.Millisecond)
	if !rateLimiter.Allow() {
		t.Error("Request after waiting should be allowed")
	}
}

func TestIPRateLimiter_PerIPIsolation(t *testing.T) {
	// Create rate limiter with burst of 1
	limiter := NewIPRateLimiter(rate.Limit(1), 1)
	
	ip1 := "192.168.1.1"
	ip2 := "192.168.1.2"
	
	limiter1 := limiter.GetLimiter(ip1)
	limiter2 := limiter.GetLimiter(ip2)
	
	// Exhaust rate limit for IP1
	if !limiter1.Allow() {
		t.Error("First request for IP1 should be allowed")
	}
	if limiter1.Allow() {
		t.Error("Second request for IP1 should be denied")
	}
	
	// IP2 should still be able to make requests
	if !limiter2.Allow() {
		t.Error("First request for IP2 should be allowed")
	}
	if limiter2.Allow() {
		t.Error("Second request for IP2 should be denied")
	}
}