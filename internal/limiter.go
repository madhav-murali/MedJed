package internal

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// interface for rate limiter to be implemented later to allow interchangable caches

type Limiter interface {
	Allow(ctx context.Context, user_key string) int64
}

type SlidingWindowLimiter struct {
	storage LimiterStorage
	limit   int64
	window  time.Duration
}

func NewSlidingWindowLimiter(addr string, limit int64, window time.Duration) *SlidingWindowLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &SlidingWindowLimiter{
		storage: client,
		limit:   limit,
		window:  window,
	}
}

func (r *SlidingWindowLimiter) Allow(ctx context.Context, user_key string) int64 {
	now := time.Now().UnixNano()
	window_start := now - r.window.Nanoseconds()
	reqId := uuid.New().String()
	result, err := r.storage.Eval(
		ctx,
		luaScript,
		[]string{user_key},
		r.limit,
		window_start,
		now,
		r.window.Nanoseconds(),
		reqId,
	).Result()
	if err != nil {
		log.Printf("Error executing script: %v , returning 1", err)
		return 1
	}

	return result.(int64)
}
