package medjed

import (
	"context"
	"sync"
	"testing"
	"time"
)

// mockLimiter simulates the sliding window algorithm in memory.
type mockLimiter struct {
	mu       sync.Mutex
	requests map[string][]int64
	limit    int64
	window   time.Duration
}

func newMockLimiter(limit int64, window time.Duration) *mockLimiter {
	return &mockLimiter{
		requests: make(map[string][]int64),
		limit:    limit,
		window:   window,
	}
}

func (m *mockLimiter) Allow(_ context.Context, key string) int64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UnixNano()
	windowStart := now - m.window.Nanoseconds()

	// Prune expired
	var active []int64
	for _, t := range m.requests[key] {
		if t > windowStart {
			active = append(active, t)
		}
	}
	m.requests[key] = active

	if int64(len(m.requests[key])) < m.limit {
		m.requests[key] = append(m.requests[key], now)
		return 1
	}
	return 0
}

func TestAllow_UnderLimit(t *testing.T) {
	limiter := newMockLimiter(5, 1*time.Minute)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if got := limiter.Allow(ctx, "user1"); got != 1 {
			t.Errorf("request %d: got %d, want 1", i+1, got)
		}
	}
}

func TestAllow_OverLimit(t *testing.T) {
	limiter := newMockLimiter(3, 1*time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		limiter.Allow(ctx, "user1")
	}

	if got := limiter.Allow(ctx, "user1"); got != 0 {
		t.Errorf("over-limit request: got %d, want 0", got)
	}
}

func TestAllow_DifferentKeys(t *testing.T) {
	limiter := newMockLimiter(1, 1*time.Minute)
	ctx := context.Background()

	if got := limiter.Allow(ctx, "user1"); got != 1 {
		t.Errorf("user1: got %d, want 1", got)
	}
	if got := limiter.Allow(ctx, "user2"); got != 1 {
		t.Errorf("user2: got %d, want 1", got)
	}
}

// ---- Benchmarks ----

// BenchmarkAllow_SingleKey measures throughput for a single client key.
func BenchmarkAllow_SingleKey(b *testing.B) {
	limiter := newMockLimiter(int64(b.N+1), 1*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, "bench-user")
	}
}

// BenchmarkAllow_MultiKey measures throughput across many distinct keys.
func BenchmarkAllow_MultiKey(b *testing.B) {
	limiter := newMockLimiter(1000, 1*time.Minute)
	ctx := context.Background()
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = "user-" + string(rune('A'+i%26)) + string(rune('0'+i%10))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		limiter.Allow(ctx, keys[i%len(keys)])
	}
}

// BenchmarkAllow_Parallel measures throughput under concurrent load.
func BenchmarkAllow_Parallel(b *testing.B) {
	limiter := newMockLimiter(1000000, 1*time.Minute)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			limiter.Allow(ctx, "parallel-user-"+string(rune('A'+i%26)))
			i++
		}
	})
}
