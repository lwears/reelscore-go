package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	redis       *redis.Client
	maxRequests int
	window      time.Duration
	isProduction bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redis.Client, maxRequests int, window time.Duration, isProduction bool) *RateLimiter {
	return &RateLimiter{
		redis:       redis,
		maxRequests: maxRequests,
		window:      window,
		isProduction: isProduction,
	}
}

// Limit returns a middleware that rate limits requests
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get identifier (user ID if authenticated, IP otherwise)
		identifier := rl.getIdentifier(r)

		// Check rate limit
		allowed, err := rl.checkRateLimit(r.Context(), identifier)
		if err != nil {
			// Log error but don't block request
			http.Error(w, `{"error":"Rate limit check failed"}`, http.StatusInternalServerError)
			return
		}

		if !allowed {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"Too many requests. Please try again later."}`)
			return
		}

		// Request allowed, continue
		next.ServeHTTP(w, r)
	})
}

// getIdentifier returns the identifier for rate limiting
func (rl *RateLimiter) getIdentifier(r *http.Request) string {
	// Try to get user ID from context
	if userID, ok := GetUserIDFromContext(r.Context()); ok {
		return fmt.Sprintf("user:%s", userID.String())
	}

	// Fallback to IP address
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	return fmt.Sprintf("ip:%s", ip)
}

// checkRateLimit checks if the request should be allowed
func (rl *RateLimiter) checkRateLimit(ctx context.Context, identifier string) (bool, error) {
	// Skip rate limiting in local/dev mode for easier testing
	if !rl.isProduction {
		return true, nil
	}

	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now().Unix()
	windowStart := now - int64(rl.window.Seconds())

	// Use Redis sorted set for sliding window
	pipe := rl.redis.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count requests in current window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d", now),
	})

	// Set expiry on the key
	pipe.Expire(ctx, key, rl.window)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Check if count exceeds limit
	count := countCmd.Val()
	return count < int64(rl.maxRequests), nil
}
