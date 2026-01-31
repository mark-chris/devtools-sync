package auth

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements rate limiting for authentication endpoints
type RateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		attempts: make(map[string][]time.Time),
	}
}

// CheckLimit checks if the key has exceeded the rate limit
// Returns error if limit exceeded, nil otherwise
func (rl *RateLimiter) CheckLimit(key string, maxAttempts int, window time.Duration) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	attempts := rl.attempts[key]

	// Filter to attempts within window
	var recent []time.Time
	for _, t := range attempts {
		if now.Sub(t) < window {
			recent = append(recent, t)
		}
	}

	if len(recent) >= maxAttempts {
		return fmt.Errorf("too many attempts, try again in %v", window)
	}

	rl.attempts[key] = append(recent, now)
	return nil
}

// ResetLimit clears the rate limit for a key
func (rl *RateLimiter) ResetLimit(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.attempts, key)
}

// Cleanup removes old entries from the rate limiter
func (rl *RateLimiter) Cleanup(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, attempts := range rl.attempts {
		var recent []time.Time
		for _, t := range attempts {
			if now.Sub(t) < maxAge {
				recent = append(recent, t)
			}
		}

		if len(recent) == 0 {
			delete(rl.attempts, key)
		} else {
			rl.attempts[key] = recent
		}
	}
}
