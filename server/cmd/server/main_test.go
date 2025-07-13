package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/xjhc/alignment/server/internal/ratelimit"
	"golang.org/x/time/rate"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name         string
		remoteAddr   string
		xForwardedFor string
		xRealIP      string
		expected     string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For takes precedence",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "192.168.1.1, 10.0.0.1",
			expected:      "192.168.1.1",
		},
		{
			name:       "X-Real-IP takes precedence over RemoteAddr",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "192.168.1.1",
			expected:   "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For takes precedence over X-Real-IP",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "192.168.1.1",
			xRealIP:       "192.168.1.2",
			expected:      "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     make(http.Header),
			}
			
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			
			result := getClientIP(req)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestRateLimiterCleanupIntegration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Create a server with the context
	server := &Server{
		ctx:                  ctx,
		joinLobbyRateLimiter: ratelimit.NewIPRateLimiter(rate.Every(time.Second), 10),
	}
	
	// Start the cleanup routine
	done := make(chan bool)
	go func() {
		server.startRateLimiterCleanup()
		done <- true
	}()
	
	// Let it run briefly
	time.Sleep(100 * time.Millisecond)
	
	// Cancel the context to stop cleanup
	cancel()
	
	// Wait for cleanup to finish
	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Error("Cleanup routine did not exit when context was cancelled")
	}
}