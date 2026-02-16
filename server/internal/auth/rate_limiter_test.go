package auth

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiterStartsCleanup(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 100*time.Millisecond, 1000)
	defer rl.Stop()

	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
}

func TestCleanupRemovesStaleEntries(t *testing.T) {
	// Use a short maxAge so entries expire quickly
	rl := NewRateLimiter(50*time.Millisecond, 100*time.Millisecond, 1000)
	defer rl.Stop()

	// Add some entries
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("ip-%d", i)
		if err := rl.CheckLimit(key, 100, time.Hour); err != nil {
			t.Fatalf("CheckLimit failed: %v", err)
		}
	}

	if rl.Len() != 10 {
		t.Fatalf("expected 10 entries, got %d", rl.Len())
	}

	// Wait for cleanup to run (maxAge=100ms, cleanup interval=50ms)
	time.Sleep(250 * time.Millisecond)

	if rl.Len() != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", rl.Len())
	}
}

func TestMaxEntriesEviction(t *testing.T) {
	maxEntries := 5
	rl := NewRateLimiter(time.Hour, time.Hour, maxEntries)
	defer rl.Stop()

	// Fill to capacity
	for i := 0; i < maxEntries; i++ {
		key := fmt.Sprintf("ip-%d", i)
		if err := rl.CheckLimit(key, 100, time.Hour); err != nil {
			t.Fatalf("CheckLimit failed for key %s: %v", key, err)
		}
	}

	if rl.Len() != maxEntries {
		t.Fatalf("expected %d entries, got %d", maxEntries, rl.Len())
	}

	// Add one more — should evict oldest and stay at maxEntries
	if err := rl.CheckLimit("new-ip", 100, time.Hour); err != nil {
		t.Fatalf("CheckLimit failed for new key: %v", err)
	}

	if rl.Len() != maxEntries {
		t.Errorf("expected %d entries after eviction, got %d", maxEntries, rl.Len())
	}
}

func TestMaxEntriesEvictsOldest(t *testing.T) {
	maxEntries := 3
	rl := NewRateLimiter(time.Hour, time.Hour, maxEntries)
	defer rl.Stop()

	// Add entries with known ordering
	// ip-0 is added first (oldest)
	for i := 0; i < maxEntries; i++ {
		key := fmt.Sprintf("ip-%d", i)
		if err := rl.CheckLimit(key, 100, time.Hour); err != nil {
			t.Fatalf("CheckLimit failed: %v", err)
		}
		// Small sleep to ensure distinct timestamps
		time.Sleep(5 * time.Millisecond)
	}

	// Add a new entry — ip-0 (oldest) should be evicted
	if err := rl.CheckLimit("new-ip", 100, time.Hour); err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}

	// ip-0 should be gone — CheckLimit for it with max=1 and short window
	// should succeed (no prior attempts), proving it was evicted
	if err := rl.CheckLimit("ip-0", 1, time.Hour); err != nil {
		t.Errorf("expected ip-0 to be evicted but it still has entries: %v", err)
	}
}

func TestStopHaltsCleanup(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, 100*time.Millisecond, 1000)

	// Add entries
	for i := 0; i < 5; i++ {
		if err := rl.CheckLimit(fmt.Sprintf("ip-%d", i), 100, time.Hour); err != nil {
			t.Fatalf("CheckLimit failed: %v", err)
		}
	}

	// Stop before cleanup can run
	rl.Stop()

	// Wait long enough for cleanup to have run if goroutine was still alive
	time.Sleep(200 * time.Millisecond)

	// Entries should still be present since cleanup goroutine was stopped
	if rl.Len() != 5 {
		t.Errorf("expected 5 entries after Stop (cleanup should not run), got %d", rl.Len())
	}

	// Stop should be idempotent
	rl.Stop()
}

func TestLen(t *testing.T) {
	rl := NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	if rl.Len() != 0 {
		t.Errorf("expected 0 for empty limiter, got %d", rl.Len())
	}

	if err := rl.CheckLimit("ip-1", 100, time.Hour); err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}

	if rl.Len() != 1 {
		t.Errorf("expected 1, got %d", rl.Len())
	}

	if err := rl.CheckLimit("ip-2", 100, time.Hour); err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}

	if rl.Len() != 2 {
		t.Errorf("expected 2, got %d", rl.Len())
	}

	// Same key again shouldn't increase count
	if err := rl.CheckLimit("ip-1", 100, time.Hour); err != nil {
		t.Fatalf("CheckLimit failed: %v", err)
	}

	if rl.Len() != 2 {
		t.Errorf("expected 2 (same key), got %d", rl.Len())
	}
}

func TestConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(50*time.Millisecond, time.Second, 10000)
	defer rl.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("ip-%d", n%20)
			_ = rl.CheckLimit(key, 1000, time.Second)
		}(i)
	}
	wg.Wait()

	// Should not panic or deadlock — just verify it completes and Len is sane
	length := rl.Len()
	if length < 1 || length > 20 {
		t.Errorf("expected between 1 and 20 unique keys, got %d", length)
	}
}

func TestCheckLimitStillEnforcesRateLimit(t *testing.T) {
	// Verify the existing rate limiting behavior still works with new constructor
	rl := NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	maxAttempts := 3
	window := time.Second

	for i := 0; i < maxAttempts; i++ {
		if err := rl.CheckLimit("test-ip", maxAttempts, window); err != nil {
			t.Fatalf("attempt %d should succeed: %v", i+1, err)
		}
	}

	// Next attempt should be rate limited
	if err := rl.CheckLimit("test-ip", maxAttempts, window); err == nil {
		t.Error("expected rate limit error, got nil")
	}
}

func TestResetLimitStillWorks(t *testing.T) {
	rl := NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	// Hit the limit
	for i := 0; i < 3; i++ {
		_ = rl.CheckLimit("test-ip", 3, time.Hour)
	}

	// Should be limited
	if err := rl.CheckLimit("test-ip", 3, time.Hour); err == nil {
		t.Fatal("expected rate limit error")
	}

	// Reset
	rl.ResetLimit("test-ip")

	// Should work again
	if err := rl.CheckLimit("test-ip", 3, time.Hour); err != nil {
		t.Errorf("expected success after reset: %v", err)
	}
}
