package rt

import (
	"sync"
	"time"
)

type RateLimiter struct {
	N     int
	Every time.Duration
	mu    sync.Mutex
	last  time.Time
}

func NewRateLimiter(n int, every time.Duration) *RateLimiter {
	return &RateLimiter{N: n, Every: every}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if time.Since(r.last) >= r.Every {
		r.last = time.Now()
		return true
	}
	return false
}
