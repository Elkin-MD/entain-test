package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func newTestLimiter(burst, refill int) *RateLimiter {
	return &RateLimiter{
		tokens: make(map[uint64]int),
		burst:  burst,
		refill: refill,
	}
}

func Test_RateLimiter_Allow_ReturnsFalse_InCaseBurstExceeded(t *testing.T) {
	// Arrange
	limiter := newTestLimiter(3, 2)

	// Act + Assert
	require.True(t, limiter.Allow(1))
	require.True(t, limiter.Allow(1))
	require.True(t, limiter.Allow(1))
	require.False(t, limiter.Allow(1))
}

func Test_RateLimiter_Allow_ReturnsTrue_InCaseRefilled(t *testing.T) {
	// Arrange
	limiter := newTestLimiter(3, 2)
	limiter.Allow(1)
	limiter.Allow(1)
	limiter.Allow(1)
	require.False(t, limiter.Allow(1))

	// Act
	limiter.refillTokens()

	// Assert
	require.True(t, limiter.Allow(1))
	require.True(t, limiter.Allow(1))
	require.False(t, limiter.Allow(1))
}

func Test_RateLimiter_Allow_IsolatesUsers_InCaseDifferentUsers(t *testing.T) {
	// Arrange
	limiter := newTestLimiter(1, 1)

	// Act + Assert
	require.True(t, limiter.Allow(1))
	require.False(t, limiter.Allow(1))
	require.True(t, limiter.Allow(2))
}
