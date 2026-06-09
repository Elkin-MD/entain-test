package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// RateLimiter is a per-user token bucket refilled on a ticker. A user may make
// up to burst requests immediately, replenished by refill tokens each second.
type RateLimiter struct {
	mu     sync.Mutex
	tokens map[uint64]int
	burst  int
	refill int
}

// NewRateLimiter creates a limiter and starts its refill ticker.
func NewRateLimiter(burst, refillPerSecond int) *RateLimiter {
	limiter := &RateLimiter{
		tokens: make(map[uint64]int),
		burst:  burst,
		refill: refillPerSecond,
	}

	go limiter.run(time.Second)

	return limiter
}

func (l *RateLimiter) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		l.refillTokens()
	}
}

func (l *RateLimiter) refillTokens() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for userID, tokens := range l.tokens {
		tokens += l.refill
		if tokens >= l.burst {
			// Fully replenished: drop the entry so it defaults to burst again.
			delete(l.tokens, userID)

			continue
		}

		l.tokens[userID] = tokens
	}
}

// Allow reports whether a request for the user may proceed, consuming a token.
func (l *RateLimiter) Allow(userID uint64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	tokens, ok := l.tokens[userID]
	if !ok {
		tokens = l.burst
	}

	if tokens <= 0 {
		l.tokens[userID] = 0

		return false
	}

	l.tokens[userID] = tokens - 1

	return true
}

// RateLimit rejects a user's requests with 429 once they exceed the allowed rate.
func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := strconv.ParseUint(r.PathValue("userId"), 10, 64)
			if err == nil && !limiter.Allow(userID) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"rate limit exceeded"}`))

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
