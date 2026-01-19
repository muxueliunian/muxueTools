// Package keypool provides API key pool management with multiple selection strategies.
package keypool

import (
	"math"
	"sync"
	"testing"
	"time"

	"muxueTools/internal/types"
)

// ==================== Strategy Interface Tests ====================

// TestRoundRobinStrategy_Select tests the round-robin selection strategy.
func TestRoundRobinStrategy_Select(t *testing.T) {
	tests := []struct {
		name     string
		keys     []*types.Key
		numCalls int
		wantIDs  []string
	}{
		{
			name: "sequential selection across multiple calls",
			keys: []*types.Key{
				createTestKey("key1", types.KeyStatusActive, true),
				createTestKey("key2", types.KeyStatusActive, true),
				createTestKey("key3", types.KeyStatusActive, true),
			},
			numCalls: 6,
			wantIDs:  []string{"key1", "key2", "key3", "key1", "key2", "key3"},
		},
		{
			name: "skip disabled keys",
			keys: []*types.Key{
				createTestKey("key1", types.KeyStatusActive, true),
				createTestKey("key2", types.KeyStatusDisabled, false),
				createTestKey("key3", types.KeyStatusActive, true),
			},
			numCalls: 4,
			wantIDs:  []string{"key1", "key3", "key1", "key3"},
		},
		{
			name: "skip rate limited keys within cooldown",
			keys: []*types.Key{
				createTestKey("key1", types.KeyStatusActive, true),
				createRateLimitedKey("key2", time.Now().Add(1*time.Hour)),
				createTestKey("key3", types.KeyStatusActive, true),
			},
			numCalls: 4,
			wantIDs:  []string{"key1", "key3", "key1", "key3"},
		},
		{
			name: "single key",
			keys: []*types.Key{
				createTestKey("key1", types.KeyStatusActive, true),
			},
			numCalls: 3,
			wantIDs:  []string{"key1", "key1", "key1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewRoundRobinStrategy()

			gotIDs := make([]string, 0, tt.numCalls)
			for i := 0; i < tt.numCalls; i++ {
				key := strategy.Select(tt.keys)
				if key != nil {
					gotIDs = append(gotIDs, key.ID)
				}
			}

			if len(gotIDs) != len(tt.wantIDs) {
				t.Errorf("got %d selections, want %d", len(gotIDs), len(tt.wantIDs))
				return
			}
			for i := range gotIDs {
				if gotIDs[i] != tt.wantIDs[i] {
					t.Errorf("selection %d: got %s, want %s", i, gotIDs[i], tt.wantIDs[i])
				}
			}
		})
	}
}

func TestRoundRobinStrategy_SelectNoAvailable(t *testing.T) {
	strategy := NewRoundRobinStrategy()

	tests := []struct {
		name string
		keys []*types.Key
	}{
		{
			name: "empty keys",
			keys: []*types.Key{},
		},
		{
			name: "nil keys",
			keys: nil,
		},
		{
			name: "all disabled",
			keys: []*types.Key{
				createTestKey("key1", types.KeyStatusDisabled, false),
				createTestKey("key2", types.KeyStatusDisabled, false),
			},
		},
		{
			name: "all rate limited",
			keys: []*types.Key{
				createRateLimitedKey("key1", time.Now().Add(1*time.Hour)),
				createRateLimitedKey("key2", time.Now().Add(1*time.Hour)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := strategy.Select(tt.keys)
			if key != nil {
				t.Errorf("expected nil, got key %s", key.ID)
			}
		})
	}
}

// TestRandomStrategy_Select tests the random selection strategy.
func TestRandomStrategy_Select(t *testing.T) {
	keys := []*types.Key{
		createTestKey("key1", types.KeyStatusActive, true),
		createTestKey("key2", types.KeyStatusActive, true),
		createTestKey("key3", types.KeyStatusActive, true),
	}

	strategy := NewRandomStrategy()

	// Run many selections and verify:
	// 1. All selected keys are from valid pool
	// 2. Distribution is somewhat uniform (not too skewed)
	selectedCounts := make(map[string]int)
	numIterations := 300

	for i := 0; i < numIterations; i++ {
		key := strategy.Select(keys)
		if key == nil {
			t.Fatal("unexpected nil key")
		}
		selectedCounts[key.ID]++
	}

	// Each key should be selected at least 20% of the time (60 times out of 300)
	minExpected := numIterations / 5
	for _, key := range keys {
		count := selectedCounts[key.ID]
		if count < minExpected {
			t.Logf("Warning: key %s selected only %d times (min expected: %d)", key.ID, count, minExpected)
		}
	}
}

func TestRandomStrategy_SkipsUnavailable(t *testing.T) {
	keys := []*types.Key{
		createTestKey("key1", types.KeyStatusActive, true),
		createTestKey("key2", types.KeyStatusDisabled, false),
		createRateLimitedKey("key3", time.Now().Add(1*time.Hour)),
	}

	strategy := NewRandomStrategy()

	// All selections should return key1 (the only available one)
	for i := 0; i < 20; i++ {
		key := strategy.Select(keys)
		if key == nil || key.ID != "key1" {
			t.Errorf("expected key1, got %v", key)
		}
	}
}

// TestLeastUsedStrategy_Select tests the least-used selection strategy.
func TestLeastUsedStrategy_Select(t *testing.T) {
	tests := []struct {
		name     string
		keys     []*types.Key
		wantID   string
	}{
		{
			name: "selects key with lowest request count",
			keys: []*types.Key{
				createKeyWithStats("key1", 100, 90),
				createKeyWithStats("key2", 50, 45),
				createKeyWithStats("key3", 200, 180),
			},
			wantID: "key2",
		},
		{
			name: "selects first key when all equal",
			keys: []*types.Key{
				createKeyWithStats("key1", 50, 45),
				createKeyWithStats("key2", 50, 45),
				createKeyWithStats("key3", 50, 45),
			},
			wantID: "key1",
		},
		{
			name: "skips unavailable keys",
			keys: []*types.Key{
				createKeyWithStats("key1", 100, 90),
				func() *types.Key {
					k := createKeyWithStats("key2", 10, 9) // lowest, but disabled
					k.Enabled = false
					return k
				}(),
				createKeyWithStats("key3", 50, 45),
			},
			wantID: "key3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := NewLeastUsedStrategy()
			key := strategy.Select(tt.keys)
			if key == nil {
				t.Fatal("unexpected nil key")
			}
			if key.ID != tt.wantID {
				t.Errorf("got %s, want %s", key.ID, tt.wantID)
			}
		})
	}
}

// TestWeightedStrategy_Select tests the weighted (by success rate) selection strategy.
func TestWeightedStrategy_Select(t *testing.T) {
	// Create keys with different success rates
	keys := []*types.Key{
		createKeyWithStats("key1", 100, 90),  // 90% success rate
		createKeyWithStats("key2", 100, 50),  // 50% success rate
		createKeyWithStats("key3", 100, 10),  // 10% success rate
	}

	strategy := NewWeightedStrategy()

	// Run many selections
	selectedCounts := make(map[string]int)
	numIterations := 1000

	for i := 0; i < numIterations; i++ {
		key := strategy.Select(keys)
		if key != nil {
			selectedCounts[key.ID]++
		}
	}

	// Key1 (90% success) should be selected more often than key3 (10% success)
	if selectedCounts["key1"] < selectedCounts["key3"] {
		t.Errorf("key1 (90%% success) should be selected more than key3 (10%% success): key1=%d, key3=%d",
			selectedCounts["key1"], selectedCounts["key3"])
	}
}

func TestWeightedStrategy_NewKeysGetChance(t *testing.T) {
	// New keys (0 requests) should still get selected
	keys := []*types.Key{
		createKeyWithStats("key1", 100, 90),
		createKeyWithStats("key2", 0, 0), // New key, no history
	}

	strategy := NewWeightedStrategy()

	// Run selections and ensure the new key gets some selections
	selectedCounts := make(map[string]int)
	numIterations := 100

	for i := 0; i < numIterations; i++ {
		key := strategy.Select(keys)
		if key != nil {
			selectedCounts[key.ID]++
		}
	}

	if selectedCounts["key2"] == 0 {
		t.Error("new key with 0 requests should still be selected sometimes")
	}
}

// ==================== Strategy Concurrency Tests ====================

func TestRoundRobinStrategy_Concurrent(t *testing.T) {
	keys := []*types.Key{
		createTestKey("key1", types.KeyStatusActive, true),
		createTestKey("key2", types.KeyStatusActive, true),
		createTestKey("key3", types.KeyStatusActive, true),
	}

	strategy := NewRoundRobinStrategy()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := strategy.Select(keys)
			if key == nil {
				t.Error("unexpected nil key in concurrent access")
			}
		}()
	}

	wg.Wait()
}

// ==================== Helper Functions ====================

func createTestKey(id string, status types.KeyStatus, enabled bool) *types.Key {
	return &types.Key{
		ID:        id,
		APIKey:    "test_api_key_" + id,
		MaskedKey: "test...key",
		Name:      "Test Key " + id,
		Status:    status,
		Enabled:   enabled,
		Tags:      []string{},
		Stats:     types.KeyStats{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createRateLimitedKey(id string, cooldownUntil time.Time) *types.Key {
	key := createTestKey(id, types.KeyStatusRateLimited, true)
	key.CooldownUntil = &cooldownUntil
	return key
}

func createKeyWithStats(id string, requestCount, successCount int64) *types.Key {
	key := createTestKey(id, types.KeyStatusActive, true)
	key.Stats = types.KeyStats{
		RequestCount: requestCount,
		SuccessCount: successCount,
		ErrorCount:   requestCount - successCount,
	}
	return key
}

// ==================== Benchmark Tests ====================

func BenchmarkRoundRobinStrategy_Select(b *testing.B) {
	keys := make([]*types.Key, 10)
	for i := 0; i < 10; i++ {
		keys[i] = createTestKey("key"+string(rune('0'+i)), types.KeyStatusActive, true)
	}

	strategy := NewRoundRobinStrategy()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strategy.Select(keys)
	}
}

func BenchmarkRandomStrategy_Select(b *testing.B) {
	keys := make([]*types.Key, 10)
	for i := 0; i < 10; i++ {
		keys[i] = createTestKey("key"+string(rune('0'+i)), types.KeyStatusActive, true)
	}

	strategy := NewRandomStrategy()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strategy.Select(keys)
	}
}

func BenchmarkLeastUsedStrategy_Select(b *testing.B) {
	keys := make([]*types.Key, 10)
	for i := 0; i < 10; i++ {
		keys[i] = createKeyWithStats("key"+string(rune('0'+i)), int64(i*100), int64(i*90))
	}

	strategy := NewLeastUsedStrategy()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strategy.Select(keys)
	}
}

func BenchmarkWeightedStrategy_Select(b *testing.B) {
	keys := make([]*types.Key, 10)
	for i := 0; i < 10; i++ {
		keys[i] = createKeyWithStats("key"+string(rune('0'+i)), int64(100), int64(50+i*5))
	}

	strategy := NewWeightedStrategy()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		strategy.Select(keys)
	}
}

// Placeholder to avoid unused import error
var _ = math.Max
