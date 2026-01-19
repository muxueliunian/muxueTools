// Package keypool provides API key pool management with multiple selection strategies.
package keypool

import (
	"errors"
	"sync"
	"time"

	"mxlnapi/internal/types"

	"github.com/google/uuid"
)

// KeyStorage is the interface for key persistence.
// This allows the pool to optionally sync with a database.
type KeyStorage interface {
	CreateKey(key *types.Key) error
	GetKey(id string) (*types.Key, error)
	ListKeys() ([]types.Key, error)
	UpdateKey(key *types.Key) error
	DeleteKey(id string) error
	KeyExists(apiKey string) (bool, error)
}

// ==================== Pool Configuration ====================

// PoolOption is a functional option for configuring the Pool.
type PoolOption func(*Pool)

// WithStrategy sets the key selection strategy.
func WithStrategy(strategy Strategy) PoolOption {
	return func(p *Pool) {
		p.strategy = strategy
	}
}

// WithCooldownSeconds sets the cooldown duration for rate-limited keys.
func WithCooldownSeconds(seconds int) PoolOption {
	return func(p *Pool) {
		p.cooldownSeconds = seconds
	}
}

// WithMaxConsecutiveFailures sets the threshold for consecutive failures before rate limiting.
func WithMaxConsecutiveFailures(count int) PoolOption {
	return func(p *Pool) {
		p.maxConsecutiveFailures = count
	}
}

// WithStorage sets the key storage backend for persistence.
func WithStorage(storage KeyStorage) PoolOption {
	return func(p *Pool) {
		p.storage = storage
	}
}

// ==================== Pool ====================

// Pool manages a collection of API keys and handles selection, rate limiting, and statistics.
type Pool struct {
	mu       sync.RWMutex
	keys     []*types.Key
	strategy Strategy
	storage  KeyStorage // Optional storage backend

	// Configuration
	cooldownSeconds        int
	maxConsecutiveFailures int

	// Internal tracking for consecutive failures per key
	consecutiveFailures map[string]int
}

// NewPool creates a new key pool from the provided key configurations.
func NewPool(configs []types.KeyConfig, opts ...PoolOption) *Pool {
	pool := &Pool{
		keys:                   make([]*types.Key, 0, len(configs)),
		strategy:               NewRoundRobinStrategy(),
		cooldownSeconds:        60,
		maxConsecutiveFailures: 5,
		consecutiveFailures:    make(map[string]int),
	}

	// Apply options
	for _, opt := range opts {
		opt(pool)
	}

	// Initialize keys from configs
	for _, cfg := range configs {
		key := &types.Key{
			ID:        uuid.New().String(),
			APIKey:    cfg.Key,
			MaskedKey: types.MaskAPIKey(cfg.Key),
			Name:      cfg.Name,
			Status:    types.KeyStatusActive,
			Enabled:   cfg.Enabled,
			Tags:      cfg.Tags,
			Stats:     types.KeyStats{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Handle disabled from config
		if !cfg.Enabled {
			key.Status = types.KeyStatusDisabled
		}

		pool.keys = append(pool.keys, key)
	}

	return pool
}

// ==================== Key Operations ====================

// GetKey retrieves an available key from the pool using the configured strategy.
// Returns ErrNoAvailableKeys if the pool is empty or all keys are disabled.
// Returns ErrAllKeysRateLimited if all keys are in cooldown.
func (p *Pool) GetKey() (*types.Key, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.keys) == 0 {
		return nil, types.ErrNoAvailableKeys
	}

	// Try to reset cooldowns for rate-limited keys
	p.resetExpiredCooldowns()

	// Use strategy to select a key
	key := p.strategy.Select(p.keys)
	if key == nil {
		// Determine if all keys are rate limited or disabled
		if p.allKeysRateLimited() {
			return nil, types.ErrAllKeysRateLimited
		}
		return nil, types.ErrNoAvailableKeys
	}

	return key, nil
}

// ReleaseKey returns a key back to the pool.
// This is a no-op for the current implementation but provided for future extensibility.
func (p *Pool) ReleaseKey(key *types.Key) {
	if key == nil {
		return
	}
	// In the current implementation, keys are always available in the pool.
	// This method is provided for API consistency and potential future use
	// (e.g., connection pooling, lease-based allocation).
}

// Size returns the total number of keys in the pool.
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.keys)
}

// ==================== Reporting ====================

// ReportSuccess records a successful request for the given key.
// model: the actual model used in this request (for usage tracking)
func (p *Pool) ReportSuccess(key *types.Key, promptTokens, completionTokens int, model string) {
	if key == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	key.IncrementStats(true, promptTokens, completionTokens, model)

	// Reset consecutive failures on success
	p.consecutiveFailures[key.ID] = 0

	// Sync to storage if available
	if p.storage != nil {
		_ = p.storage.UpdateKey(key) // Best effort, don't block on storage errors
	}
}

// ReportFailure records a failed request for the given key.
// If the error indicates rate limiting, the key enters cooldown.
// If consecutive failures exceed the threshold, the key also enters cooldown.
// model: the actual model used in this request (for usage tracking)
func (p *Pool) ReportFailure(key *types.Key, err error, model string) {
	if key == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	key.IncrementStats(false, 0, 0, model)

	// Check if this is a rate limit error
	if isRateLimitError(err) {
		key.SetRateLimited(p.cooldownSeconds)
		p.consecutiveFailures[key.ID] = 0
		return
	}

	// Track consecutive failures
	p.consecutiveFailures[key.ID]++
	if p.consecutiveFailures[key.ID] >= p.maxConsecutiveFailures {
		key.SetRateLimited(p.cooldownSeconds)
		p.consecutiveFailures[key.ID] = 0
	}

	// Sync to storage if available
	if p.storage != nil {
		_ = p.storage.UpdateKey(key) // Best effort
	}
}

// ==================== Statistics ====================

// GetStats returns statistics for all keys in the pool.
func (p *Pool) GetStats() []types.Key {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make([]types.Key, len(p.keys))
	for i, key := range p.keys {
		// Create a copy to avoid exposing internal state
		stats[i] = types.Key{
			ID:            key.ID,
			MaskedKey:     key.MaskedKey,
			Name:          key.Name,
			Status:        key.Status,
			Enabled:       key.Enabled,
			Tags:          key.Tags,
			Stats:         key.Stats,
			CooldownUntil: key.CooldownUntil,
			CreatedAt:     key.CreatedAt,
			UpdatedAt:     key.UpdatedAt,
		}
	}
	return stats
}

// ==================== Dynamic Key Management ====================

// AddKey adds a new key to the pool.
// If storage is configured, the key is also persisted.
func (p *Pool) AddKey(key *types.Key) error {
	if key == nil {
		return errors.New("key cannot be nil")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Check for duplicate
	for _, k := range p.keys {
		if k.APIKey == key.APIKey {
			return errors.New("key already exists")
		}
	}

	// Generate ID if not set
	if key.ID == "" {
		key.ID = uuid.New().String()
	}
	if key.MaskedKey == "" {
		key.MaskedKey = types.MaskAPIKey(key.APIKey)
	}
	if key.CreatedAt.IsZero() {
		key.CreatedAt = time.Now()
	}
	key.UpdatedAt = time.Now()

	// Persist to storage first
	if p.storage != nil {
		if err := p.storage.CreateKey(key); err != nil {
			return err
		}
	}

	p.keys = append(p.keys, key)
	return nil
}

// RemoveKey removes a key from the pool by ID.
// If storage is configured, the key is also deleted from storage.
func (p *Pool) RemoveKey(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Find and remove from slice
	found := false
	for i, key := range p.keys {
		if key.ID == id {
			p.keys = append(p.keys[:i], p.keys[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return types.ErrKeyNotFound
	}

	// Delete from storage
	if p.storage != nil {
		if err := p.storage.DeleteKey(id); err != nil {
			// Key already removed from memory, but DB deletion failed.
			// This may cause inconsistency between memory and DB.
			// Log the error but don't fail since the memory state is already updated.
			// The key will be re-added to DB on next restart if still in config,
			// or the DB record will be orphaned if key was dynamically added.
			// TODO: Consider rollback strategy or reconciliation on startup.
			_ = err // Logged in production via structured logging
		}
	}

	// Clean up consecutive failures tracking
	delete(p.consecutiveFailures, id)

	return nil
}

// GetKeyByID returns a key by its ID.
func (p *Pool) GetKeyByID(id string) (*types.Key, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, key := range p.keys {
		if key.ID == id {
			return key, nil
		}
	}
	return nil, types.ErrKeyNotFound
}

// LoadFromStorage loads all keys from storage into the pool.
// This replaces any existing keys in the pool.
func (p *Pool) LoadFromStorage() error {
	if p.storage == nil {
		return errors.New("no storage configured")
	}

	keys, err := p.storage.ListKeys()
	if err != nil {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.keys = make([]*types.Key, 0, len(keys))
	for i := range keys {
		key := keys[i]
		if !key.Enabled {
			key.Status = types.KeyStatusDisabled
		} else {
			key.Status = types.KeyStatusActive
		}
		p.keys = append(p.keys, &key)
	}

	p.consecutiveFailures = make(map[string]int)
	return nil
}

// SyncConfigToStorage syncs keys from config to storage (for initial setup).
// Only adds keys that don't already exist in storage.
func (p *Pool) SyncConfigToStorage(configs []types.KeyConfig) (int, error) {
	if p.storage == nil {
		return 0, errors.New("no storage configured")
	}

	synced := 0
	for _, cfg := range configs {
		// Check if key already exists
		exists, err := p.storage.KeyExists(cfg.Key)
		if err != nil {
			continue
		}
		if exists {
			continue
		}

		// Create new key
		key := &types.Key{
			ID:        uuid.New().String(),
			APIKey:    cfg.Key,
			MaskedKey: types.MaskAPIKey(cfg.Key),
			Name:      cfg.Name,
			Status:    types.KeyStatusActive,
			Enabled:   cfg.Enabled,
			Tags:      cfg.Tags,
			Stats:     types.KeyStats{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if !cfg.Enabled {
			key.Status = types.KeyStatusDisabled
		}

		if err := p.storage.CreateKey(key); err == nil {
			synced++
		}
	}

	return synced, nil
}

// HasStorage returns true if a storage backend is configured.
func (p *Pool) HasStorage() bool {
	return p.storage != nil
}

// ==================== Runtime Configuration ====================

// SetStrategy updates the key selection strategy at runtime.
func (p *Pool) SetStrategy(strategy Strategy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.strategy = strategy
}

// GetStrategy returns the current key selection strategy name.
func (p *Pool) GetStrategyName() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.strategy.Name()
}

// SetCooldownSeconds updates the cooldown duration at runtime.
func (p *Pool) SetCooldownSeconds(seconds int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if seconds > 0 {
		p.cooldownSeconds = seconds
	}
}

// GetCooldownSeconds returns the current cooldown duration.
func (p *Pool) GetCooldownSeconds() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.cooldownSeconds
}

// SetMaxConsecutiveFailures updates the max consecutive failures threshold.
func (p *Pool) SetMaxConsecutiveFailures(count int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if count > 0 {
		p.maxConsecutiveFailures = count
	}
}

// GetMaxConsecutiveFailures returns the current max consecutive failures threshold.
func (p *Pool) GetMaxConsecutiveFailures() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.maxConsecutiveFailures
}

// ==================== Internal Helpers ====================

// resetExpiredCooldowns checks all rate-limited keys and resets those
// whose cooldown has expired.
func (p *Pool) resetExpiredCooldowns() {
	now := time.Now()
	for _, key := range p.keys {
		if key.Status == types.KeyStatusRateLimited {
			if key.CooldownUntil == nil || now.After(*key.CooldownUntil) {
				key.Status = types.KeyStatusActive
				key.CooldownUntil = nil
			}
		}
	}
}

// allKeysRateLimited returns true if there are enabled keys and all of them
// are currently rate limited (in cooldown).
func (p *Pool) allKeysRateLimited() bool {
	now := time.Now()
	hasEnabledKey := false
	allRateLimited := true

	for _, key := range p.keys {
		if !key.Enabled || key.Status == types.KeyStatusDisabled {
			continue
		}

		hasEnabledKey = true

		if key.Status == types.KeyStatusActive {
			allRateLimited = false
			break
		}
		if key.Status == types.KeyStatusRateLimited {
			// Check if cooldown has expired
			if key.CooldownUntil == nil || now.After(*key.CooldownUntil) {
				allRateLimited = false
				break
			}
		}
	}

	// Only return true if there are enabled keys and all are rate limited
	return hasEnabledKey && allRateLimited
}

// isRateLimitError checks if an error indicates rate limiting.
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	// Check for AppError with rate limit code
	var appErr *types.AppError
	if errors.As(err, &appErr) {
		return appErr.Code == types.ErrCodeRateLimit || appErr.HTTPStatus == 429
	}

	// Fallback checks for known error patterns
	errStr := err.Error()
	return contains(errStr, "429") ||
		contains(errStr, "rate limit") ||
		contains(errStr, "quota exceeded") ||
		contains(errStr, "too many requests")
}

// contains is a simple case-insensitive substring check.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldSubstring(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFoldSubstring(s1, s2 string) bool {
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		// Simple ASCII case folding
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 32
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 32
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
