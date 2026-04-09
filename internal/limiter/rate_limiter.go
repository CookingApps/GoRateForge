package limiter

import (
	"context"
	"sync"
	"time"
)

type RateLimiter struct {
	capacity   int           // max tokens (burst size)
	tokens     int           // current available tokens
	rate       int           // tokens per second (refill rate)
	lastRefill time.Time
	mu         sync.Mutex
	// you fit add more fields if needed (e.g. ticker or refill interval)
}

func NewRateLimiter(ratePerSecond, burstCapacity int) *RateLimiter {
	
	return &RateLimiter{ capacity: burstCapacity, tokens: burstCapacity, rate: ratePerSecond ,lastRefill: time.Now(), mu: sync.Mutex{} }
}

// Main method workers will call


func (rl *RateLimiter) Wait(ctx context.Context) error {
    for {   // keep trying until we get a token or context is cancelled
        rl.mu.Lock()
        rl.refill()

        if rl.tokens > 0 {
            rl.tokens--
            rl.mu.Unlock()
            return nil
        }

        rl.mu.Unlock()   // IMPORTANT: unlock BEFORE waiting

        // Now wait a short time before trying again
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(100 * time.Millisecond):  // small backoff
        }
    }
}

// Optional helper: refill logic can be a separate method

func (rl *RateLimiter) refill() {
    now := time.Now()
    elapsed := now.Sub(rl.lastRefill).Seconds()

    tokensToAdd := int(elapsed) * rl.rate

    if tokensToAdd > 0 {
        rl.tokens += tokensToAdd
        if rl.tokens > rl.capacity {
            rl.tokens = rl.capacity
        }
        rl.lastRefill = now
    }
}