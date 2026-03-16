package safety

import (
	"fmt"
	"time"
)

type RateLimiter struct {
	perMinute int
	hits      []time.Time
}

func NewRateLimiter(perMinute int) *RateLimiter {
	if perMinute <= 0 {
		perMinute = 60
	}
	return &RateLimiter{perMinute: perMinute}
}

func (r *RateLimiter) Allow() error {
	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)
	filtered := r.hits[:0]
	for _, t := range r.hits {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	r.hits = filtered
	if len(r.hits) >= r.perMinute {
		return fmt.Errorf("rate limit exceeded: %d/%d requests per minute", len(r.hits), r.perMinute)
	}
	r.hits = append(r.hits, now)
	return nil
}
