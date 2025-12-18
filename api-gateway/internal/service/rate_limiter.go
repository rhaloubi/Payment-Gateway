package service

import (
	"sync"
	"time"

	"github.com/rhaloubi/api-gateway/internal/config"
)

type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	config  *config.Config
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	limit      int
	window     time.Duration
}

func NewRateLimiter(cfg *config.Config) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		config:  cfg,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) Allow(key string, limit int, window time.Duration) (bool, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     limit,
			lastRefill: time.Now(),
			limit:      limit,
			window:     window,
		}
		rl.buckets[key] = b
	}

	now := time.Now()
	if now.Sub(b.lastRefill) >= b.window {
		b.tokens = b.limit
		b.lastRefill = now
	}

	if b.tokens > 0 {
		b.tokens--
		resetTime := b.lastRefill.Add(b.window)
		return true, b.tokens, resetTime
	}

	resetTime := b.lastRefill.Add(b.window)
	return false, 0, resetTime
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			if now.Sub(b.lastRefill) > 1*time.Hour {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}
