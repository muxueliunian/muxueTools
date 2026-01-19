// Package types defines all data transfer objects and core types for MuxueTools.
package types

import (
	"strings"
	"time"
)

// ==================== Key Status Enum ====================

// KeyStatus represents the current operational status of an API key.
type KeyStatus string

const (
	// KeyStatusActive indicates the key is available for use.
	KeyStatusActive KeyStatus = "active"
	// KeyStatusRateLimited indicates the key is cooling down after hitting rate limit.
	KeyStatusRateLimited KeyStatus = "rate_limited"
	// KeyStatusDisabled indicates the key has been manually disabled.
	KeyStatusDisabled KeyStatus = "disabled"
)

// IsValid returns true if the status is a valid KeyStatus value.
func (s KeyStatus) IsValid() bool {
	switch s {
	case KeyStatusActive, KeyStatusRateLimited, KeyStatusDisabled:
		return true
	}
	return false
}

// ==================== Core Key Types ====================

// Key represents an API key with its metadata and statistics.
type Key struct {
	ID            string     `json:"id"`
	APIKey        string     `json:"-"`   // Never serialize to JSON
	MaskedKey     string     `json:"key"` // Display only (e.g., "AIzaSy...xxx")
	Name          string     `json:"name"`
	Status        KeyStatus  `json:"status"`
	Enabled       bool       `json:"enabled"`
	Tags          []string   `json:"tags"`
	Provider      string     `json:"provider"`      // e.g., "google_aistudio"
	DefaultModel  string     `json:"default_model"` // e.g., "gemini-1.5-pro-latest"
	Stats         KeyStats   `json:"stats"`
	CooldownUntil *time.Time `json:"cooldown_until,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// KeyStats holds usage statistics for a single key.
type KeyStats struct {
	RequestCount     int64            `json:"request_count"`
	SuccessCount     int64            `json:"success_count"`
	ErrorCount       int64            `json:"error_count"`
	PromptTokens     int64            `json:"prompt_tokens"`
	CompletionTokens int64            `json:"completion_tokens"`
	LastUsedAt       *time.Time       `json:"last_used_at,omitempty"`
	ModelUsage       map[string]int64 `json:"model_usage,omitempty"`
}

// ==================== Key Configuration ====================

// KeyConfig represents a key entry in the configuration file.
type KeyConfig struct {
	Key     string   `mapstructure:"key" yaml:"key"`
	Name    string   `mapstructure:"name" yaml:"name"`
	Enabled bool     `mapstructure:"enabled" yaml:"enabled"`
	Tags    []string `mapstructure:"tags" yaml:"tags"`
}

// ==================== Admin API DTOs ====================

// KeyListResponse represents the response for GET /api/keys.
type KeyListResponse struct {
	Success bool  `json:"success"`
	Data    []Key `json:"data"`
	Total   int   `json:"total"`
}

// KeyInfo is an alias for Key used in API responses.
type KeyInfo = Key

// CreateKeyRequest represents the request body for POST /api/keys.
type CreateKeyRequest struct {
	Key          string   `json:"key" binding:"required"`
	Name         string   `json:"name,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	Provider     string   `json:"provider,omitempty"`
	DefaultModel string   `json:"default_model,omitempty"`
}

// CreateKeyResponse represents the response for POST /api/keys.
type CreateKeyResponse struct {
	Success bool `json:"success"`
	Data    Key  `json:"data"`
}

// DeleteKeyResponse represents the response for DELETE /api/keys/:id.
type DeleteKeyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// TestKeyResponse represents the response for POST /api/keys/:id/test.
type TestKeyResponse struct {
	Success bool          `json:"success"`
	Data    TestKeyResult `json:"data"`
}

// TestKeyResult contains the result of testing a key's validity.
type TestKeyResult struct {
	Valid     bool     `json:"valid"`
	LatencyMs int64    `json:"latency_ms"`
	Models    []string `json:"models,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// ImportKeyItem represents a single key entry in the import request.
type ImportKeyItem struct {
	Key  string   `json:"key" binding:"required"`
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`
}

// ImportKeysRequest represents the request body for POST /api/keys/import.
type ImportKeysRequest struct {
	Keys []ImportKeyItem `json:"keys" binding:"required"` // List of keys to import
}

// ImportKeysResponse represents the response for POST /api/keys/import.
type ImportKeysResponse struct {
	Success bool             `json:"success"`
	Data    ImportKeysResult `json:"data"`
}

// ImportKeysResult contains the result of batch importing keys.
type ImportKeysResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"` // Duplicate keys
	Errors   []string `json:"errors"`
}

// ==================== Statistics DTOs ====================

// KeyStatsResponse represents the response for GET /api/stats/keys.
type KeyStatsResponse struct {
	Success bool          `json:"success"`
	Data    []KeyStatItem `json:"data"`
}

// KeyStatItem represents statistics for a single key.
type KeyStatItem struct {
	KeyID        string  `json:"key_id"`
	KeyName      string  `json:"key_name"`
	RequestCount int64   `json:"request_count"`
	SuccessRate  float64 `json:"success_rate"` // Percentage (0-100)
	TokenUsage   int64   `json:"token_usage"`
	AvgLatencyMs float64 `json:"avg_latency_ms"`
}

// StatsResponse represents the response for GET /api/stats.
type StatsResponse struct {
	Success bool      `json:"success"`
	Data    StatsData `json:"data"`
}

// StatsData contains aggregate statistics.
type StatsData struct {
	Period       StatsPeriod  `json:"period"`
	Requests     RequestStats `json:"requests"`
	Tokens       TokenStats   `json:"tokens"`
	AvgLatencyMs float64      `json:"avg_latency_ms"`
}

// StatsPeriod defines the time range for statistics.
type StatsPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RequestStats contains request-level statistics.
type RequestStats struct {
	Total       int64 `json:"total"`
	Success     int64 `json:"success"`
	Error       int64 `json:"error"`
	RateLimited int64 `json:"rate_limited"`
}

// TokenStats contains token consumption statistics.
type TokenStats struct {
	Prompt     int64 `json:"prompt"`
	Completion int64 `json:"completion"`
	Total      int64 `json:"total"`
}

// ==================== Trend & Model Usage DTOs ====================

// StatsTimeRange represents supported time ranges for statistics queries.
type StatsTimeRange string

const (
	// StatsTimeRange24H represents the last 24 hours.
	StatsTimeRange24H StatsTimeRange = "24h"
	// StatsTimeRange7D represents the last 7 days.
	StatsTimeRange7D StatsTimeRange = "7d"
	// StatsTimeRange30D represents the last 30 days.
	StatsTimeRange30D StatsTimeRange = "30d"
)

// TrendDataPoint represents a single data point in the trend.
type TrendDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Requests  int64     `json:"requests"`
	Tokens    int64     `json:"tokens"`
	Errors    int64     `json:"errors"`
}

// TrendResponse represents the response for GET /api/stats/trend.
type TrendResponse struct {
	Success   bool             `json:"success"`
	Data      []TrendDataPoint `json:"data"`
	TimeRange string           `json:"time_range"` // 当前查询的时间范围
}

// ModelUsageItem represents usage statistics for a single model.
type ModelUsageItem struct {
	Model        string  `json:"model"`
	RequestCount int64   `json:"request_count"`
	TokenUsage   int64   `json:"token_usage"`
	Percentage   float64 `json:"percentage"` // 基于请求数计算的百分比 (0-100)
}

// ModelUsageResponse represents the response for GET /api/stats/models.
type ModelUsageResponse struct {
	Success bool             `json:"success"`
	Data    []ModelUsageItem `json:"data"`
}

// ==================== Helper Methods ====================

// MaskAPIKey returns a masked version of an API key for display.
// Example: "AIzaSyABC123xyz" -> "AIzaSy...xyz"
func MaskAPIKey(apiKey string) string {
	if len(apiKey) < 12 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:6] + "..." + apiKey[len(apiKey)-3:]
}

// IsAvailable returns true if the key can be used for a request right now.
func (k *Key) IsAvailable() bool {
	if !k.Enabled || k.Status == KeyStatusDisabled {
		return false
	}
	if k.Status == KeyStatusRateLimited {
		if k.CooldownUntil != nil && time.Now().Before(*k.CooldownUntil) {
			return false
		}
		// Cooldown has expired, key should be reset to active
	}
	return true
}

// SetRateLimited marks the key as rate limited with a cooldown period.
func (k *Key) SetRateLimited(cooldownSeconds int) {
	k.Status = KeyStatusRateLimited
	cooldownUntil := time.Now().Add(time.Duration(cooldownSeconds) * time.Second)
	k.CooldownUntil = &cooldownUntil
}

// ResetCooldown resets the key to active status if cooldown has expired.
func (k *Key) ResetCooldown() bool {
	if k.Status != KeyStatusRateLimited {
		return false
	}
	if k.CooldownUntil == nil || time.Now().After(*k.CooldownUntil) {
		k.Status = KeyStatusActive
		k.CooldownUntil = nil
		return true
	}
	return false
}

// IncrementStats updates the key's statistics after a request.
// model: the actual model used in this request (for usage tracking)
func (k *Key) IncrementStats(success bool, promptTokens, completionTokens int, model string) {
	k.Stats.RequestCount++
	if success {
		k.Stats.SuccessCount++
	} else {
		k.Stats.ErrorCount++
	}
	k.Stats.PromptTokens += int64(promptTokens)
	k.Stats.CompletionTokens += int64(completionTokens)
	now := time.Now()
	k.Stats.LastUsedAt = &now
	k.UpdatedAt = now

	// Track model usage
	if model != "" {
		if k.Stats.ModelUsage == nil {
			k.Stats.ModelUsage = make(map[string]int64)
		}
		k.Stats.ModelUsage[model]++
	}
}

// TotalTokens returns the total token consumption for this key.
func (s *KeyStats) TotalTokens() int64 {
	return s.PromptTokens + s.CompletionTokens
}

// SuccessRate calculates the success rate as a percentage (0-100).
func (s *KeyStats) SuccessRate() float64 {
	if s.RequestCount == 0 {
		return 0
	}
	return float64(s.SuccessCount) / float64(s.RequestCount) * 100
}

// ==================== Key Validation DTOs ====================

// ValidateKeyRequest represents the request for POST /api/keys/validate.
type ValidateKeyRequest struct {
	Key      string `json:"key" binding:"required"`
	Provider string `json:"provider,omitempty"` // 默认 "google_aistudio"
}

// ValidateKeyResponse represents the response for POST /api/keys/validate.
type ValidateKeyResponse struct {
	Success bool              `json:"success"`
	Data    ValidateKeyResult `json:"data"`
}

// ValidateKeyResult contains the validation result.
type ValidateKeyResult struct {
	Valid     bool     `json:"valid"`
	LatencyMs int64    `json:"latency_ms"`
	Models    []string `json:"models"`
	Error     string   `json:"error,omitempty"`
}
