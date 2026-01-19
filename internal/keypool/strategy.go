// Package keypool provides API key pool management with multiple selection strategies.
package keypool

import (
	"math"
	"math/rand"
	"sync"
	"sync/atomic"

	"muxueTools/internal/types"
)

// ==================== Strategy Interface ====================

// Strategy defines the interface for key selection algorithms.
type Strategy interface {
	// Select picks a key from the provided slice of keys.
	// Returns nil if no available key can be selected.
	Select(keys []*types.Key) *types.Key

	// Name returns the strategy identifier.
	Name() string
}

// StrategyFactory creates a strategy based on the configuration.
func StrategyFactory(strategyName types.PoolStrategy) Strategy {
	switch strategyName {
	case types.PoolStrategyRandom:
		return NewRandomStrategy()
	case types.PoolStrategyLeastUsed:
		return NewLeastUsedStrategy()
	case types.PoolStrategyWeighted:
		return NewWeightedStrategy()
	default:
		return NewRoundRobinStrategy()
	}
}

// ==================== Round Robin Strategy ====================

// RoundRobinStrategy implements a round-robin key selection algorithm.
// Keys are selected in sequential order, cycling back to the start.
type RoundRobinStrategy struct {
	index uint64 // atomic counter for thread-safe round-robin
}

// NewRoundRobinStrategy creates a new round-robin strategy instance.
func NewRoundRobinStrategy() *RoundRobinStrategy {
	return &RoundRobinStrategy{}
}

// Select picks the next available key in round-robin order.
func (s *RoundRobinStrategy) Select(keys []*types.Key) *types.Key {
	availableKeys := filterAvailable(keys)
	if len(availableKeys) == 0 {
		return nil
	}

	// Atomically increment and get index
	idx := atomic.AddUint64(&s.index, 1) - 1
	selected := availableKeys[idx%uint64(len(availableKeys))]
	return selected
}

// Name returns the strategy identifier.
func (s *RoundRobinStrategy) Name() string {
	return string(types.PoolStrategyRoundRobin)
}

// ==================== Random Strategy ====================

// RandomStrategy implements random key selection.
// Each call picks a random available key with uniform probability.
type RandomStrategy struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewRandomStrategy creates a new random strategy instance.
func NewRandomStrategy() *RandomStrategy {
	return &RandomStrategy{
		rng: rand.New(rand.NewSource(rand.Int63())),
	}
}

// Select picks a random available key.
func (s *RandomStrategy) Select(keys []*types.Key) *types.Key {
	availableKeys := filterAvailable(keys)
	if len(availableKeys) == 0 {
		return nil
	}

	s.mu.Lock()
	idx := s.rng.Intn(len(availableKeys))
	s.mu.Unlock()

	return availableKeys[idx]
}

// Name returns the strategy identifier.
func (s *RandomStrategy) Name() string {
	return string(types.PoolStrategyRandom)
}

// ==================== Least Used Strategy ====================

// LeastUsedStrategy implements selection based on request count.
// The key with the fewest total requests is preferred.
type LeastUsedStrategy struct{}

// NewLeastUsedStrategy creates a new least-used strategy instance.
func NewLeastUsedStrategy() *LeastUsedStrategy {
	return &LeastUsedStrategy{}
}

// Select picks the key with the lowest request count.
func (s *LeastUsedStrategy) Select(keys []*types.Key) *types.Key {
	availableKeys := filterAvailable(keys)
	if len(availableKeys) == 0 {
		return nil
	}

	minKey := availableKeys[0]
	minCount := minKey.Stats.RequestCount

	for _, key := range availableKeys[1:] {
		if key.Stats.RequestCount < minCount {
			minKey = key
			minCount = key.Stats.RequestCount
		}
	}

	return minKey
}

// Name returns the strategy identifier.
func (s *LeastUsedStrategy) Name() string {
	return string(types.PoolStrategyLeastUsed)
}

// ==================== Weighted Strategy ====================

// WeightedStrategy implements weighted selection based on success rate.
// Keys with higher success rates are more likely to be selected.
// New keys (0 requests) are given a fair default weight.
type WeightedStrategy struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewWeightedStrategy creates a new weighted strategy instance.
func NewWeightedStrategy() *WeightedStrategy {
	return &WeightedStrategy{
		rng: rand.New(rand.NewSource(rand.Int63())),
	}
}

// Select picks a key with probability proportional to its success rate.
func (s *WeightedStrategy) Select(keys []*types.Key) *types.Key {
	availableKeys := filterAvailable(keys)
	if len(availableKeys) == 0 {
		return nil
	}

	// Calculate weights based on success rate
	weights := make([]float64, len(availableKeys))
	totalWeight := 0.0

	for i, key := range availableKeys {
		weight := calculateWeight(key)
		weights[i] = weight
		totalWeight += weight
	}

	// Weighted random selection
	s.mu.Lock()
	randomValue := s.rng.Float64() * totalWeight
	s.mu.Unlock()

	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randomValue <= cumulative {
			return availableKeys[i]
		}
	}

	// Fallback to last key (should not happen normally)
	return availableKeys[len(availableKeys)-1]
}

// Name returns the strategy identifier.
func (s *WeightedStrategy) Name() string {
	return string(types.PoolStrategyWeighted)
}

// calculateWeight computes the selection weight for a key.
// Uses success rate with a minimum baseline to ensure all keys get a chance.
func calculateWeight(key *types.Key) float64 {
	const (
		minWeight     = 0.1  // Minimum weight to ensure selection chance
		defaultWeight = 0.5  // Default weight for new keys
		maxWeight     = 1.0  // Maximum weight
	)

	if key.Stats.RequestCount == 0 {
		// New keys get a fair default weight
		return defaultWeight
	}

	// Calculate success rate (0.0 to 1.0)
	successRate := float64(key.Stats.SuccessCount) / float64(key.Stats.RequestCount)

	// Apply minimum threshold to avoid 0 weight
	weight := math.Max(minWeight, successRate)

	// Cap at maximum
	return math.Min(weight, maxWeight)
}

// ==================== Helper Functions ====================

// filterAvailable returns only the keys that are currently available for use.
func filterAvailable(keys []*types.Key) []*types.Key {
	if len(keys) == 0 {
		return nil
	}

	// Pre-allocate with capacity estimate
	available := make([]*types.Key, 0, len(keys))

	for _, key := range keys {
		if key != nil && key.IsAvailable() {
			available = append(available, key)
		}
	}

	return available
}
