// Package medjed provides a distributed rate limiter backed by Redis.
//
// Usage in any Gin project:
//
//	import (
//	    "github.com/madhav-murali/medjed"
//	    "github.com/madhav-murali/medjed/middleware"
//	)
//
//	limiter := medjed.NewSlidingWindowLimiter("localhost:6379", 100, time.Minute)
//	router.Use(middleware.RateLimitMiddleware(limiter))
package medjed

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Limiter defines the interface for rate limiting.
// Implement this to use a custom backend with the middleware.
type Limiter interface {
	Allow(ctx context.Context, key string) int64
}

// SlidingWindowLimiter is a Redis-backed rate limiter using the sliding window algorithm.
type SlidingWindowLimiter struct {
	client *redis.Client
	limit  int64
	window time.Duration
}

// NewSlidingWindowLimiter creates a new rate limiter connected to the given Redis address.
func NewSlidingWindowLimiter(addr string, limit int64, window time.Duration) *SlidingWindowLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &SlidingWindowLimiter{
		client: client,
		limit:  limit,
		window: window,
	}
}

// Allow checks whether a request identified by key is within the rate limit.
// Returns 1 if allowed, 0 if rate limited.
func (r *SlidingWindowLimiter) Allow(ctx context.Context, key string) int64 {
	now := time.Now().UnixNano()
	windowStart := now - r.window.Nanoseconds()
	reqId := uuid.New().String()
	result, err := r.client.Eval(
		ctx,
		luaScript,
		[]string{key},
		r.limit,
		windowStart,
		now,
		r.window.Nanoseconds(),
		reqId,
	).Result()
	if err != nil {
		log.Printf("Error executing script: %v, allowing request", err)
		return 1
	}

	return result.(int64)
}

var luaScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window_start = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local window_ns = tonumber(ARGV[4])
local reqId = ARGV[5]

redis.call("ZREMRANGEBYSCORE", key, 0, window_start)

local count = redis.call("ZCARD", key)

if count < limit then
	redis.call("ZADD", key, now, reqId)
	redis.call("EXPIRE", key, math.ceil(window_ns / 1000000000))
	return 1
else
	return 0
end
`
