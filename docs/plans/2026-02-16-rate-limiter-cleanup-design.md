# Rate Limiter Cleanup Scheduler Design

**Issue:** #59 — Add rate limiter cleanup scheduler to prevent memory exhaustion
**Date:** 2026-02-16
**Status:** Approved

## Problem

The rate limiter uses an in-memory `map[string][]time.Time` that grows unbounded. The `Cleanup()` method exists but is never called automatically. An attacker could exhaust server memory by sending requests from many unique IPs.

## Scope

Harden the `RateLimiter` itself. Wiring into auth handlers is Issue #58.

## Design

### API Changes

```go
// Constructor starts a background cleanup goroutine.
func NewRateLimiter(cleanupInterval, maxAge time.Duration, maxEntries int) *RateLimiter

// Stops the cleanup goroutine. Safe to call multiple times.
func (rl *RateLimiter) Stop()

// Returns current number of tracked keys.
func (rl *RateLimiter) Len() int
```

### Struct

```go
type RateLimiter struct {
    mu         sync.Mutex
    attempts   map[string][]time.Time
    maxEntries int
    stopCh     chan struct{}
    stopped    sync.Once
}
```

### Behavior

1. **Periodic cleanup:** A goroutine ticks at `cleanupInterval` and calls `Cleanup(maxAge)` to remove entries with no recent attempts.
2. **Hard cap:** On `CheckLimit`, if the map is at `maxEntries` and the key is new, evict the entry whose most-recent attempt is oldest.
3. **Stop:** Closes `stopCh`, goroutine exits. `sync.Once` ensures idempotency.
4. **Len:** Returns `len(attempts)` under lock.

### Suggested Defaults

- `cleanupInterval`: 5 minutes
- `maxAge`: 15 minutes
- `maxEntries`: 100,000

### Metrics

Expose `Len()` method only. Prometheus/structured logging deferred to Issue #32.

### Tests

- `TestCleanupRemovesStaleEntries`
- `TestMaxEntriesEviction`
- `TestStopHaltsCleanup`
- `TestConcurrentAccess`
- `TestLen`

## Alternatives Considered

- **Sharded map:** Better concurrency but premature optimization for auth-rate endpoints.
- **Third-party LRU cache:** LRU semantics don't match (active attackers stay cached). Unnecessary dependency.
