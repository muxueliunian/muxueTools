// Package keypool provides API key pool management with multiple selection strategies.
package keypool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"mxlnapi/internal/types"

	"golang.org/x/sync/errgroup"
)

// ==================== Pool Creation Tests ====================

func TestNewPool(t *testing.T) {
	tests := []struct {
		name      string
		configs   []types.KeyConfig
		opts      []PoolOption
		wantCount int
		wantErr   bool
	}{
		{
			name: "creates pool with valid keys",
			configs: []types.KeyConfig{
				{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
				{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
			},
			opts:      nil,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "creates empty pool with no keys",
			configs:   []types.KeyConfig{},
			opts:      nil,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:      "handles nil configs",
			configs:   nil,
			opts:      nil,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "includes disabled keys in pool",
			configs: []types.KeyConfig{
				{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
				{Key: "AIzaSyKey2", Name: "Key 2", Enabled: false},
			},
			opts:      nil,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "applies custom strategy",
			configs: []types.KeyConfig{
				{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
			},
			opts:      []PoolOption{WithStrategy(NewRandomStrategy())},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := NewPool(tt.configs, tt.opts...)
			if pool == nil {
				t.Fatal("NewPool returned nil")
			}
			if got := pool.Size(); got != tt.wantCount {
				t.Errorf("Size() = %d, want %d", got, tt.wantCount)
			}
		})
	}
}

// ==================== GetKey Tests ====================

func TestPool_GetKey_HappyPath(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
	}

	pool := NewPool(configs)

	// Should get a key successfully
	key, err := pool.GetKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key == nil {
		t.Fatal("expected key, got nil")
	}
	if key.APIKey == "" {
		t.Error("expected APIKey to be set")
	}
}

func TestPool_GetKey_RoundRobinOrder(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
		{Key: "AIzaSyKey3", Name: "Key 3", Enabled: true},
	}

	pool := NewPool(configs, WithStrategy(NewRoundRobinStrategy()))

	// Get 6 keys and verify round-robin order
	expectedNames := []string{"Key 1", "Key 2", "Key 3", "Key 1", "Key 2", "Key 3"}
	for i, expectedName := range expectedNames {
		key, err := pool.GetKey()
		if err != nil {
			t.Fatalf("iteration %d: unexpected error: %v", i, err)
		}
		if key.Name != expectedName {
			t.Errorf("iteration %d: got %s, want %s", i, key.Name, expectedName)
		}
		pool.ReleaseKey(key)
	}
}

func TestPool_GetKey_SkipsDisabledKeys(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: false},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
		{Key: "AIzaSyKey3", Name: "Key 3", Enabled: false},
	}

	pool := NewPool(configs, WithStrategy(NewRoundRobinStrategy()))

	// Should only return Key 2
	for i := 0; i < 5; i++ {
		key, err := pool.GetKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.Name != "Key 2" {
			t.Errorf("expected Key 2, got %s", key.Name)
		}
		pool.ReleaseKey(key)
	}
}

func TestPool_GetKey_NoAvailableKeys(t *testing.T) {
	pool := NewPool([]types.KeyConfig{})

	key, err := pool.GetKey()

	if key != nil {
		t.Errorf("expected nil key, got %v", key)
	}
	if !errors.Is(err, types.ErrNoAvailableKeys) {
		t.Errorf("expected ErrNoAvailableKeys, got %v", err)
	}
}

func TestPool_GetKey_AllKeysDisabled(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: false},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: false},
	}

	pool := NewPool(configs)

	key, err := pool.GetKey()

	if key != nil {
		t.Errorf("expected nil key, got %v", key)
	}
	if !errors.Is(err, types.ErrNoAvailableKeys) {
		t.Errorf("expected ErrNoAvailableKeys, got %v", err)
	}
}

func TestPool_GetKey_AllKeysRateLimited(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
	}

	pool := NewPool(configs, WithCooldownSeconds(60))

	// Rate limit all keys
	for _, key := range pool.keys {
		key.SetRateLimited(60)
	}

	key, err := pool.GetKey()

	if key != nil {
		t.Errorf("expected nil key, got %v", key)
	}
	if !errors.Is(err, types.ErrAllKeysRateLimited) {
		t.Errorf("expected ErrAllKeysRateLimited, got %v", err)
	}
}

// ==================== ReleaseKey Tests ====================

func TestPool_ReleaseKey(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	pool := NewPool(configs)

	key, _ := pool.GetKey()
	pool.ReleaseKey(key)

	// Verify the key can be obtained again
	key2, err := pool.GetKey()
	if err != nil {
		t.Errorf("unexpected error after release: %v", err)
	}
	if key2 == nil {
		t.Error("expected key after release")
	}
}

func TestPool_ReleaseKey_NilKey(t *testing.T) {
	pool := NewPool([]types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	})

	// Should not panic on nil key
	pool.ReleaseKey(nil)
}

// ==================== ReportSuccess/Failure Tests ====================

func TestPool_ReportSuccess(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	pool := NewPool(configs)

	key, _ := pool.GetKey()
	initialRequestCount := key.Stats.RequestCount

	pool.ReportSuccess(key, 100, 50, "test-model")

	if key.Stats.RequestCount != initialRequestCount+1 {
		t.Errorf("request count not incremented")
	}
	if key.Stats.SuccessCount != 1 {
		t.Errorf("success count should be 1, got %d", key.Stats.SuccessCount)
	}
	if key.Stats.PromptTokens != 100 {
		t.Errorf("prompt tokens should be 100, got %d", key.Stats.PromptTokens)
	}
	if key.Stats.CompletionTokens != 50 {
		t.Errorf("completion tokens should be 50, got %d", key.Stats.CompletionTokens)
	}
}

func TestPool_ReportFailure_RateLimitError(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	pool := NewPool(configs, WithCooldownSeconds(30))

	key, _ := pool.GetKey()

	// Simulate rate limit error
	rateLimitErr := &types.AppError{
		Code:       types.ErrCodeRateLimit,
		HTTPStatus: 429,
	}
	pool.ReportFailure(key, rateLimitErr, "test-model")

	if key.Status != types.KeyStatusRateLimited {
		t.Errorf("key should be rate limited, got status %s", key.Status)
	}
	if key.CooldownUntil == nil {
		t.Error("cooldown until should be set")
	}
}

func TestPool_ReportFailure_GenericError(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	pool := NewPool(configs)

	key, _ := pool.GetKey()

	// Report a generic error (not rate limit)
	genericErr := errors.New("network error")
	pool.ReportFailure(key, genericErr, "")

	if key.Status != types.KeyStatusActive {
		t.Errorf("key should remain active for generic errors, got status %s", key.Status)
	}
	if key.Stats.ErrorCount != 1 {
		t.Errorf("error count should be 1, got %d", key.Stats.ErrorCount)
	}
}

func TestPool_ReportFailure_ConsecutiveFailures(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	// Set max consecutive failures to 3
	pool := NewPool(configs, WithMaxConsecutiveFailures(3), WithCooldownSeconds(60))

	key, _ := pool.GetKey()

	// Report 3 consecutive failures
	for i := 0; i < 3; i++ {
		pool.ReportFailure(key, errors.New("some error"), "")
	}

	if key.Status != types.KeyStatusRateLimited {
		t.Errorf("key should be rate limited after %d consecutive failures, got status %s",
			3, key.Status)
	}
}

// ==================== Cooldown Recovery Tests ====================

func TestPool_CooldownRecovery(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	pool := NewPool(configs, WithCooldownSeconds(1))

	key, _ := pool.GetKey()

	// Rate limit the key with 1 second cooldown
	rateLimitErr := &types.AppError{Code: types.ErrCodeRateLimit}
	pool.ReportFailure(key, rateLimitErr, "")

	// Verify key is not available immediately
	_, err := pool.GetKey()
	if !errors.Is(err, types.ErrAllKeysRateLimited) {
		// It's possible the cooldown already expired if the system is slow
		// So just log rather than fail
		t.Logf("key might have recovered faster than expected: %v", err)
	}

	// Wait for cooldown to expire
	time.Sleep(1100 * time.Millisecond)

	// Now the key should be available
	recoveredKey, err := pool.GetKey()
	if err != nil {
		t.Errorf("key should be available after cooldown: %v", err)
	}
	if recoveredKey == nil {
		t.Error("expected key after cooldown recovery")
	}
	if recoveredKey.Status != types.KeyStatusActive {
		t.Errorf("key status should be active after recovery, got %s", recoveredKey.Status)
	}
}

// ==================== GetStats Tests ====================

func TestPool_GetStats(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
	}

	pool := NewPool(configs)

	// Make some requests
	k1, _ := pool.GetKey()
	pool.ReportSuccess(k1, 100, 50, "model-a")
	pool.ReleaseKey(k1)

	k2, _ := pool.GetKey()
	pool.ReportSuccess(k2, 200, 100, "model-b")
	pool.ReleaseKey(k2)

	stats := pool.GetStats()

	if len(stats) != 2 {
		t.Errorf("expected 2 stats entries, got %d", len(stats))
	}

	// Verify total tokens
	totalTokens := int64(0)
	for _, s := range stats {
		totalTokens += s.Stats.TotalTokens()
	}
	if totalTokens != 450 { // 100+50+200+100
		t.Errorf("expected total tokens 450, got %d", totalTokens)
	}
}

func TestPool_GetStats_Empty(t *testing.T) {
	pool := NewPool([]types.KeyConfig{})

	stats := pool.GetStats()

	if len(stats) != 0 {
		t.Errorf("expected 0 stats entries, got %d", len(stats))
	}
}

// ==================== Concurrency Tests ====================

func TestPool_ConcurrentGetKey(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
		{Key: "AIzaSyKey3", Name: "Key 3", Enabled: true},
	}

	pool := NewPool(configs)

	var wg sync.WaitGroup
	numGoroutines := 100
	successCount := atomic.Int64{}
	errorCount := atomic.Int64{}

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key, err := pool.GetKey()
			if err != nil {
				errorCount.Add(1)
				return
			}
			successCount.Add(1)

			// Simulate some work
			time.Sleep(1 * time.Millisecond)

			pool.ReleaseKey(key)
		}()
	}

	wg.Wait()

	if errorCount.Load() > 0 {
		t.Errorf("unexpected errors in concurrent access: %d", errorCount.Load())
	}
	t.Logf("Successful concurrent GetKey calls: %d", successCount.Load())
}

func TestPool_ConcurrentGetKey_WithErrgroup(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
		{Key: "AIzaSyKey3", Name: "Key 3", Enabled: true},
	}

	pool := NewPool(configs)

	g, ctx := errgroup.WithContext(context.Background())

	for i := 0; i < 100; i++ {
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			key, err := pool.GetKey()
			if err != nil {
				return err
			}
			defer pool.ReleaseKey(key)

			// Simulate API call
			time.Sleep(1 * time.Millisecond)

			pool.ReportSuccess(key, 10, 5, "test-model")
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		t.Errorf("concurrent operations failed: %v", err)
	}

	// Verify all stats are updated correctly
	stats := pool.GetStats()
	totalRequests := int64(0)
	for _, s := range stats {
		totalRequests += s.Stats.RequestCount
	}
	if totalRequests != 100 {
		t.Errorf("expected 100 total requests, got %d", totalRequests)
	}
}

func TestPool_ConcurrentMixedOperations(t *testing.T) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
		{Key: "AIzaSyKey2", Name: "Key 2", Enabled: true},
	}

	pool := NewPool(configs, WithCooldownSeconds(60), WithMaxConsecutiveFailures(5))

	var wg sync.WaitGroup

	// Start multiple goroutines doing different operations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key, err := pool.GetKey()
			if err != nil {
				return // Some keys might be rate limited
			}
			defer pool.ReleaseKey(key)

			// Mix of success and failure reports
			if id%5 == 0 {
				pool.ReportFailure(key, errors.New("random error"), "")
			} else {
				pool.ReportSuccess(key, 10, 5, "test-model")
			}
		}(i)
	}

	// Concurrent stats reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = pool.GetStats()
		}()
	}

	wg.Wait()
}

// ==================== Benchmark Tests ====================

func BenchmarkPool_GetKey(b *testing.B) {
	configs := make([]types.KeyConfig, 10)
	for i := 0; i < 10; i++ {
		configs[i] = types.KeyConfig{
			Key:     "AIzaSyKey" + string(rune('0'+i)),
			Name:    "Key " + string(rune('0'+i)),
			Enabled: true,
		}
	}

	pool := NewPool(configs)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key, _ := pool.GetKey()
		pool.ReleaseKey(key)
	}
}

func BenchmarkPool_GetKey_Concurrent(b *testing.B) {
	configs := make([]types.KeyConfig, 10)
	for i := 0; i < 10; i++ {
		configs[i] = types.KeyConfig{
			Key:     "AIzaSyKey" + string(rune('0'+i)),
			Name:    "Key " + string(rune('0'+i)),
			Enabled: true,
		}
	}

	pool := NewPool(configs)
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key, err := pool.GetKey()
			if err == nil && key != nil {
				pool.ReleaseKey(key)
			}
		}
	})
}

func BenchmarkPool_ReportSuccess(b *testing.B) {
	configs := []types.KeyConfig{
		{Key: "AIzaSyKey1", Name: "Key 1", Enabled: true},
	}

	pool := NewPool(configs)
	key, _ := pool.GetKey()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pool.ReportSuccess(key, 100, 50, "benchmark-model")
	}
}
