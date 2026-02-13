package server

import (
	"context"
	"sync"
	"time"
)

// RateLimiter tracks request counts per IP using a sliding window.
type RateLimiter struct {
	mu       sync.Mutex
	limit    int // requests per minute
	requests map[string][]time.Time
}

// NewRateLimiter creates a rate limiter that allows limit requests per minute per IP.
func NewRateLimiter(limit int) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		requests: make(map[string][]time.Time),
	}
}

// Allow checks if the given key is within the default rate limit.
func (rl *RateLimiter) Allow(key string) bool {
	return rl.AllowN(key, rl.limit)
}

// AllowN checks if the given key is within the specified per-minute limit.
func (rl *RateLimiter) AllowN(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-time.Minute)

	// Filter expired entries
	times := rl.requests[key]
	valid := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= limit {
		rl.requests[key] = valid
		return false
	}

	rl.requests[key] = append(valid, now)
	return true
}

// Cleanup runs a background goroutine to purge stale entries every 60s.
func (rl *RateLimiter) Cleanup(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-time.Minute)
			for ip, times := range rl.requests {
				valid := times[:0]
				for _, t := range times {
					if t.After(cutoff) {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(rl.requests, ip)
				} else {
					rl.requests[ip] = valid
				}
			}
			rl.mu.Unlock()
		}
	}
}
