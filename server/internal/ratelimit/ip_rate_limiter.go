package ratelimit

import (
	"sync"
	"golang.org/x/time/rate"
)

// IPRateLimiter manages per-IP rate limiting
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewIPRateLimiter creates a new per-IP rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for a given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// CleanupStale removes stale limiters (called periodically)
func (i *IPRateLimiter) CleanupStale() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for ip, limiter := range i.limiters {
		// Remove limiters that haven't been used recently
		if limiter.Tokens() == float64(i.burst) {
			delete(i.limiters, ip)
		}
	}
}