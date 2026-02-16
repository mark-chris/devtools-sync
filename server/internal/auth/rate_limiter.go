package auth

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements rate limiting for authentication endpoints.
// It automatically cleans up stale entries and enforces a maximum map size
// to prevent memory exhaustion under sustained attack.
type RateLimiter struct {
	mu         sync.Mutex
	attempts   map[string][]time.Time
	maxEntries int
	stopCh     chan struct{}
	stopped    sync.Once
}

// NewRateLimiter creates a new rate limiter that automatically cleans up
// stale entries. cleanupInterval controls how often cleanup runs, maxAge
// controls how long entries are kept, and maxEntries caps the map size.
func NewRateLimiter(cleanupInterval, maxAge time.Duration, maxEntries int) *RateLimiter {
	rl := &RateLimiter{
		attempts:   make(map[string][]time.Time),
		maxEntries: maxEntries,
		stopCh:     make(chan struct{}),
	}

	go rl.cleanupLoop(cleanupInterval, maxAge)

	return rl
}

func (rl *RateLimiter) cleanupLoop(interval, maxAge time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.Cleanup(maxAge)
		case <-rl.stopCh:
			return
		}
	}
}

// Stop halts the background cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Stop() {
	rl.stopped.Do(func() {
		close(rl.stopCh)
	})
}

// Len returns the current number of tracked keys.
func (rl *RateLimiter) Len() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return len(rl.attempts)
}

// CheckLimit checks if the key has exceeded the rate limit.
// Returns error if limit exceeded, nil otherwise.
// If the map is at capacity and the key is new, the entry with the
// oldest most-recent attempt is evicted.
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

	// Evict oldest entry if at capacity and this is a new key
	if _, exists := rl.attempts[key]; !exists && len(rl.attempts) >= rl.maxEntries {
		rl.evictOldest()
	}

	rl.attempts[key] = append(recent, now)
	return nil
}

// evictOldest removes the entry whose most-recent attempt is the oldest.
// Must be called with mu held.
func (rl *RateLimiter) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, attempts := range rl.attempts {
		if len(attempts) == 0 {
			delete(rl.attempts, key)
			return
		}
		mostRecent := attempts[len(attempts)-1]
		if oldestKey == "" || mostRecent.Before(oldestTime) {
			oldestKey = key
			oldestTime = mostRecent
		}
	}

	if oldestKey != "" {
		delete(rl.attempts, oldestKey)
	}
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
