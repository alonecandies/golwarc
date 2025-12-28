package libs

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter provides rate limiting functionality using token bucket algorithm
type RateLimiter struct {
	limiter *rate.Limiter
	mu      sync.Mutex
}

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerSecond int // Number of requests per second
	Burst             int // Maximum burst size
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config RateLimiterConfig) *RateLimiter {
	if config.RequestsPerSecond <= 0 {
		config.RequestsPerSecond = 10 // Default to 10 requests per second
	}
	if config.Burst <= 0 {
		config.Burst = config.RequestsPerSecond // Default burst equals RPS
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst),
	}
}

// Wait blocks until the rate limiter allows an event to proceed
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Allow checks if an event can proceed without blocking
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// Reserve reserves a slot and returns a reservation
func (rl *RateLimiter) Reserve() *rate.Reservation {
	return rl.limiter.Reserve()
}

// SetLimit updates the rate limit
func (rl *RateLimiter) SetLimit(requestsPerSecond int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limiter.SetLimit(rate.Limit(requestsPerSecond))
}

// SetBurst updates the burst size
func (rl *RateLimiter) SetBurst(burst int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.limiter.SetBurst(burst)
}

// WaitN blocks until n events can proceed
func (rl *RateLimiter) WaitN(ctx context.Context, n int) error {
	return rl.limiter.WaitN(ctx, n)
}

// WaitWithTimeout waits for rate limiter with a timeout
func (rl *RateLimiter) WaitWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return rl.Wait(ctx)
}
