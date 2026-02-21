package nodes

import (
	"sync"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// RateLimiter manages rate limits per MCP server connection.
type RateLimiter struct {
	mu       sync.Mutex
	limiters map[uuid.UUID]*rate.Limiter

	limit rate.Limit
	burst int
}

// NewRateLimiter creates a new RateLimiter.
// limit: events per second. burst: max burst size.
func NewRateLimiter(l rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[uuid.UUID]*rate.Limiter),
		limit:    l,
		burst:    b,
	}
}

// Allow checks if a request is allowed for the given server ID.
func (rl *RateLimiter) Allow(serverID uuid.UUID) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[serverID]
	if !exists {
		limiter = rate.NewLimiter(rl.limit, rl.burst)
		rl.limiters[serverID] = limiter
	}

	return limiter.Allow()
}

// RemoveLimiter cleans up after a connection closes.
func (rl *RateLimiter) RemoveLimiter(serverID uuid.UUID) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limiters, serverID)
}
