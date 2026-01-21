// Package api provides HTTP API handlers and routing for MuxueTools.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"muxueTools/internal/config"
	"muxueTools/internal/keypool"
	"muxueTools/internal/storage"
	"muxueTools/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ==================== Admin Handler ====================

// AdminHandler handles administrative API endpoints.
type AdminHandler struct {
	pool    *keypool.Pool
	logger  *logrus.Logger
	storage *storage.Storage
}

// NewAdminHandler creates a new admin handler.
func NewAdminHandler(pool *keypool.Pool, logger *logrus.Logger, store *storage.Storage) *AdminHandler {
	return &AdminHandler{
		pool:    pool,
		logger:  logger,
		storage: store,
	}
}

// ==================== Models ====================

// ListAvailableModels handles GET /api/models - Get available models from Gemini API.
// Uses a valid key from the pool to query the Gemini models API.
func (h *AdminHandler) ListAvailableModels(c *gin.Context) {
	// Get a valid key from the pool
	key, err := h.pool.GetKey()
	if err != nil {
		h.logger.WithError(err).Warn("Failed to get key for models list")
		RespondSuccess(c, []string{})
		return
	}

	// Create HTTP client with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Call Gemini models.list API
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + key.APIKey
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create models request")
		RespondSuccess(c, []string{})
		return
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		h.logger.WithError(err).Warn("Models list request failed")
		RespondSuccess(c, []string{})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read models response")
		RespondSuccess(c, []string{})
		return
	}

	if resp.StatusCode != http.StatusOK {
		h.logger.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(body),
		}).Warn("Gemini API returned error for models list")
		RespondSuccess(c, []string{})
		return
	}

	// Parse models list
	var result struct {
		Models []struct {
			Name                       string   `json:"name"`
			SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.WithError(err).Error("Failed to parse models response")
		RespondSuccess(c, []string{})
		return
	}

	// Filter to only include models that support generateContent (chat)
	modelNames := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		// Only include models that support "generateContent"
		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				modelNames = append(modelNames, strings.TrimPrefix(m.Name, "models/"))
				break
			}
		}
	}

	h.logger.WithFields(logrus.Fields{
		"model_count": len(modelNames),
	}).Debug("Retrieved available models from Gemini API")

	RespondSuccess(c, modelNames)
}

// ==================== Key Management ====================

// ListKeys handles GET /api/keys - List all keys with masked display.
func (h *AdminHandler) ListKeys(c *gin.Context) {
	stats := h.pool.GetStats()

	resp := types.KeyListResponse{
		Success: true,
		Data:    stats,
		Total:   len(stats),
	}

	c.JSON(http.StatusOK, resp)
}

// AddKey handles POST /api/keys - Add a new key.
func (h *AdminHandler) AddKey(c *gin.Context) {
	var req types.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Validate key format (basic check)
	if len(req.Key) < 10 {
		RespondBadRequest(c, "Invalid API key format")
		return
	}

	// Create key object
	newKey := &types.Key{
		ID:           uuid.New().String(),
		APIKey:       req.Key,
		MaskedKey:    types.MaskAPIKey(req.Key),
		Name:         req.Name,
		Status:       types.KeyStatusActive,
		Enabled:      true,
		Tags:         req.Tags,
		Provider:     req.Provider,
		DefaultModel: req.DefaultModel,
		Stats:        types.KeyStats{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if newKey.Tags == nil {
		newKey.Tags = []string{}
	}

	// Set default provider
	if newKey.Provider == "" {
		newKey.Provider = "google_aistudio"
	}

	// Add to pool (will also persist to DB if storage is configured)
	if err := h.pool.AddKey(newKey); err != nil {
		if err.Error() == "key already exists" {
			RespondBadRequest(c, "API key already exists")
			return
		}
		h.logger.WithError(err).Error("Failed to add key")
		RespondInternalError(c, "Failed to add key")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"key_id":     newKey.ID,
		"key_name":   req.Name,
		"masked_key": newKey.MaskedKey,
	}).Info("Key added successfully")

	c.JSON(http.StatusCreated, types.CreateKeyResponse{
		Success: true,
		Data:    *newKey,
	})
}

// ValidateKey handles POST /api/keys/validate - Validate a key and get available models.
func (h *AdminHandler) ValidateKey(c *gin.Context) {
	var req types.ValidateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Create HTTP client with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Call Gemini models.list API
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + req.Key
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create request")
		c.JSON(http.StatusOK, types.ValidateKeyResponse{
			Success: true,
			Data:    types.ValidateKeyResult{Valid: false, Error: fmt.Sprintf("Request creation failed: %v", err)},
		})
		return
	}

	start := time.Now()
	resp, err := http.DefaultClient.Do(httpReq)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		h.logger.WithError(err).Warn("Key validation request failed")
		c.JSON(http.StatusOK, types.ValidateKeyResponse{
			Success: true,
			Data:    types.ValidateKeyResult{Valid: false, LatencyMs: latency, Error: fmt.Sprintf("Request failed: %v", err)},
		})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.WithError(err).Error("Failed to read response body")
		c.JSON(http.StatusOK, types.ValidateKeyResponse{
			Success: true,
			Data:    types.ValidateKeyResult{Valid: false, LatencyMs: latency, Error: "Failed to read response"},
		})
		return
	}

	if resp.StatusCode != http.StatusOK {
		// Parse error response
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		errMsg := "Invalid API Key"
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			errMsg = errResp.Error.Message
		}

		c.JSON(http.StatusOK, types.ValidateKeyResponse{
			Success: true,
			Data:    types.ValidateKeyResult{Valid: false, LatencyMs: latency, Error: errMsg},
		})
		return
	}

	// Parse models list
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.WithError(err).Error("Failed to parse models response")
		c.JSON(http.StatusOK, types.ValidateKeyResponse{
			Success: true,
			Data:    types.ValidateKeyResult{Valid: false, LatencyMs: latency, Error: "Failed to parse response"},
		})
		return
	}

	// Extract model names (remove "models/" prefix)
	modelNames := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		modelNames = append(modelNames, strings.TrimPrefix(m.Name, "models/"))
	}

	h.logger.WithFields(logrus.Fields{
		"latency_ms":  latency,
		"model_count": len(modelNames),
	}).Info("Key validated successfully")

	c.JSON(http.StatusOK, types.ValidateKeyResponse{
		Success: true,
		Data: types.ValidateKeyResult{
			Valid:     true,
			LatencyMs: latency,
			Models:    modelNames,
		},
	})
}

// DeleteKey handles DELETE /api/keys/:id - Delete a key.
func (h *AdminHandler) DeleteKey(c *gin.Context) {
	keyID := c.Param("id")
	if keyID == "" {
		RespondBadRequest(c, "Key ID is required")
		return
	}

	// Remove from pool (will also delete from DB if storage is configured)
	if err := h.pool.RemoveKey(keyID); err != nil {
		if err == types.ErrKeyNotFound {
			RespondNotFound(c, "Key")
			return
		}
		h.logger.WithError(err).Error("Failed to delete key")
		RespondInternalError(c, "Failed to delete key")
		return
	}

	h.logger.WithFields(logrus.Fields{
		"key_id": keyID,
	}).Info("Key deleted successfully")

	c.JSON(http.StatusOK, types.DeleteKeyResponse{
		Success: true,
		Message: "Key deleted successfully",
	})
}

// TestKey handles POST /api/keys/:id/test - Test key validity.
func (h *AdminHandler) TestKey(c *gin.Context) {
	keyID := c.Param("id")
	if keyID == "" {
		RespondBadRequest(c, "Key ID is required")
		return
	}

	// Find key in pool
	stats := h.pool.GetStats()
	var targetKey *types.Key
	for _, key := range stats {
		if key.ID == keyID {
			targetKey = &key
			break
		}
	}

	if targetKey == nil {
		RespondNotFound(c, "Key")
		return
	}

	// Note: Actual key testing would require making a test request to Gemini API
	// For now, return a mock response based on key status
	result := types.TestKeyResult{
		Valid:     targetKey.Status == types.KeyStatusActive && targetKey.Enabled,
		LatencyMs: 150, // Mock latency
		Models:    []string{"gemini-1.5-pro-latest", "gemini-1.5-flash-latest"},
	}

	if !result.Valid {
		if targetKey.Status == types.KeyStatusRateLimited {
			result.Error = "Key is currently rate limited"
		} else if targetKey.Status == types.KeyStatusDisabled || !targetKey.Enabled {
			result.Error = "Key is disabled"
		}
	}

	c.JSON(http.StatusOK, types.TestKeyResponse{
		Success: true,
		Data:    result,
	})
}

// ImportKeys handles POST /api/keys/import - Batch import keys.
func (h *AdminHandler) ImportKeys(c *gin.Context) {
	var req types.ImportKeysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	result := types.ImportKeysResult{
		Imported: 0,
		Skipped:  0,
		Errors:   []string{},
	}

	for i, item := range req.Keys {
		// Basic validation
		if len(item.Key) < 10 {
			result.Errors = append(result.Errors, fmt.Sprintf("Item %d: Key too short", i+1))
			continue
		}

		// Create key object
		newKey := &types.Key{
			ID:        uuid.New().String(),
			APIKey:    item.Key,
			MaskedKey: types.MaskAPIKey(item.Key),
			Name:      item.Name,
			Status:    types.KeyStatusActive,
			Enabled:   true,
			Tags:      item.Tags,
			Provider:  "google_aistudio", // Default provider, could be inferred or passed
			Stats:     types.KeyStats{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if newKey.Tags == nil {
			newKey.Tags = []string{}
		}

		// Add to pool
		if err := h.pool.AddKey(newKey); err != nil {
			if err.Error() == "key already exists" {
				result.Skipped++
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Item %d: %v", i+1, err))
			}
			continue
		}

		result.Imported++
	}

	h.logger.WithFields(logrus.Fields{
		"imported": result.Imported,
		"skipped":  result.Skipped,
		"errors":   len(result.Errors),
	}).Info("Keys import completed")

	c.JSON(http.StatusOK, types.ImportKeysResponse{
		Success: true,
		Data:    result,
	})
}

// ExportKeys handles GET /api/keys/export - Export all keys as text.
func (h *AdminHandler) ExportKeys(c *gin.Context) {
	// Note: This would export the actual API keys
	// For security, this might require additional authentication
	// For now, return a placeholder message

	c.Header("Content-Type", "text/plain")
	c.Header("Content-Disposition", "attachment; filename=keys.txt")

	stats := h.pool.GetStats()
	var keys []string
	for _, key := range stats {
		// In a real implementation, we would export the actual API key
		// For now, export masked keys as a safety measure
		keys = append(keys, key.MaskedKey)
	}

	c.String(http.StatusOK, strings.Join(keys, "\n"))
}

// ==================== Statistics ====================

// GetStats handles GET /api/stats - Get aggregate statistics.
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats := h.pool.GetStats()

	// Calculate aggregate statistics
	var totalRequests, successCount, errorCount, rateLimitedCount int64
	var totalPromptTokens, totalCompletionTokens int64

	for _, key := range stats {
		totalRequests += key.Stats.RequestCount
		successCount += key.Stats.SuccessCount
		errorCount += key.Stats.ErrorCount
		totalPromptTokens += key.Stats.PromptTokens
		totalCompletionTokens += key.Stats.CompletionTokens
	}

	resp := types.StatsResponse{
		Success: true,
		Data: types.StatsData{
			Period: types.StatsPeriod{
				Start: time.Now().AddDate(0, 0, -7), // Last 7 days
				End:   time.Now(),
			},
			Requests: types.RequestStats{
				Total:       totalRequests,
				Success:     successCount,
				Error:       errorCount,
				RateLimited: rateLimitedCount,
			},
			Tokens: types.TokenStats{
				Prompt:     totalPromptTokens,
				Completion: totalCompletionTokens,
				Total:      totalPromptTokens + totalCompletionTokens,
			},
			AvgLatencyMs: 0, // Would be calculated from request logs
		},
	}

	c.JSON(http.StatusOK, resp)
}

// GetKeyStats handles GET /api/stats/keys - Get per-key statistics.
func (h *AdminHandler) GetKeyStats(c *gin.Context) {
	stats := h.pool.GetStats()

	var keyStats []types.KeyStatItem
	for _, key := range stats {
		keyStats = append(keyStats, types.KeyStatItem{
			KeyID:        key.ID,
			KeyName:      key.Name,
			RequestCount: key.Stats.RequestCount,
			SuccessRate:  key.Stats.SuccessRate(),
			TokenUsage:   key.Stats.TotalTokens(),
			AvgLatencyMs: 0, // Would be calculated from request logs
		})
	}

	c.JSON(http.StatusOK, types.KeyStatsResponse{
		Success: true,
		Data:    keyStats,
	})
}

// GetStatsTrend handles GET /api/stats/trend - Get request trend over time.
// Query params:
//   - range: 24h | 7d | 30d (default: 7d)
func (h *AdminHandler) GetStatsTrend(c *gin.Context) {
	rangeStr := c.DefaultQuery("range", "7d")

	// Validate range
	validRanges := map[string]bool{"24h": true, "7d": true, "30d": true}
	if !validRanges[rangeStr] {
		rangeStr = "7d"
	}

	// Generate time points based on range
	var points []types.TrendDataPoint
	now := time.Now()

	switch rangeStr {
	case "24h":
		// 24 points (hourly)
		for i := 23; i >= 0; i-- {
			t := now.Add(-time.Duration(i) * time.Hour)
			points = append(points, types.TrendDataPoint{
				Timestamp: t.Truncate(time.Hour),
				Requests:  0,
				Tokens:    0,
				Errors:    0,
			})
		}
	case "7d":
		// 7 points (daily)
		for i := 6; i >= 0; i-- {
			t := now.AddDate(0, 0, -i)
			points = append(points, types.TrendDataPoint{
				Timestamp: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()),
				Requests:  0,
				Tokens:    0,
				Errors:    0,
			})
		}
	case "30d":
		// 30 points (daily)
		for i := 29; i >= 0; i-- {
			t := now.AddDate(0, 0, -i)
			points = append(points, types.TrendDataPoint{
				Timestamp: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()),
				Requests:  0,
				Tokens:    0,
				Errors:    0,
			})
		}
	}

	// Populate with current aggregated data (simplified approach)
	// Note: Without historical data storage, we distribute current stats to the last point
	stats := h.pool.GetStats()
	if len(points) > 0 && len(stats) > 0 {
		lastIdx := len(points) - 1
		for _, key := range stats {
			points[lastIdx].Requests += key.Stats.RequestCount
			points[lastIdx].Tokens += key.Stats.TotalTokens()
			points[lastIdx].Errors += key.Stats.ErrorCount
		}
	}

	c.JSON(http.StatusOK, types.TrendResponse{
		Success:   true,
		Data:      points,
		TimeRange: rangeStr,
	})
}

// GetStatsModels handles GET /api/stats/models - Get usage by model.
// Priority: Use KeyStats.ModelUsage (request-level tracking) if available,
// otherwise fall back to Key.DefaultModel.
func (h *AdminHandler) GetStatsModels(c *gin.Context) {
	stats := h.pool.GetStats()

	// Aggregate model usage across all keys
	modelMap := make(map[string]*types.ModelUsageItem)
	var totalRequests int64

	for _, key := range stats {
		// Priority 1: Use request-level ModelUsage data if available
		if len(key.Stats.ModelUsage) > 0 {
			for model, count := range key.Stats.ModelUsage {
				if _, exists := modelMap[model]; !exists {
					modelMap[model] = &types.ModelUsageItem{Model: model}
				}
				modelMap[model].RequestCount += count
				totalRequests += count
			}
			// Token usage is distributed proportionally (simplified: assign to first model)
			// In practice, tokens are not tracked per-model, so we skip TokenUsage here
		} else if key.Stats.RequestCount > 0 {
			// Priority 2: Fall back to DefaultModel for keys without ModelUsage
			model := key.DefaultModel
			if model == "" {
				model = "unknown"
			}

			if _, exists := modelMap[model]; !exists {
				modelMap[model] = &types.ModelUsageItem{Model: model}
			}

			modelMap[model].RequestCount += key.Stats.RequestCount
			modelMap[model].TokenUsage += key.Stats.TotalTokens()
			totalRequests += key.Stats.RequestCount
		}
	}

	// Calculate percentages and convert to slice
	result := make([]types.ModelUsageItem, 0, len(modelMap))
	for _, item := range modelMap {
		if totalRequests > 0 {
			item.Percentage = float64(item.RequestCount) / float64(totalRequests) * 100
		}
		result = append(result, *item)
	}

	// Sort by request count descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].RequestCount > result[j].RequestCount
	})

	c.JSON(http.StatusOK, types.ModelUsageResponse{
		Success: true,
		Data:    result,
	})
}

// ==================== Configuration ====================

// GetConfig handles GET /api/config - Get current configuration (sanitized).
func (h *AdminHandler) GetConfig(c *gin.Context) {
	cfg := config.GetSafe()
	if cfg == nil {
		RespondInternalError(c, "Configuration not loaded")
		return
	}

	// Get security config from storage
	var ipWhitelistEnabled bool
	var whitelistIP, proxyKey string
	if h.storage != nil {
		if val, _ := h.storage.GetConfig("security.ip_whitelist_enabled"); val == "true" {
			ipWhitelistEnabled = true
		}
		whitelistIP, _ = h.storage.GetConfig("security.whitelist_ip")
		proxyKey, _ = h.storage.GetConfig("security.proxy_key")
	}
	if proxyKey == "" {
		proxyKey = DefaultProxyKey
	}

	// Get update source from storage
	var updateSource string
	if h.storage != nil {
		updateSource, _ = h.storage.GetConfig("update.source")
	}
	if updateSource == "" {
		updateSource = "mxln"
	}

	// Get request timeout from storage or config
	requestTimeout := cfg.Advanced.RequestTimeout
	if h.storage != nil {
		if storedTimeout, _ := h.storage.GetConfig("advanced.request_timeout"); storedTimeout != "" {
			if parsed, err := strconv.Atoi(storedTimeout); err == nil && parsed > 0 {
				requestTimeout = parsed
			}
		}
	}
	if requestTimeout == 0 {
		requestTimeout = 120 // default
	}

	// Get stored port from storage (for user configuration, requires restart)
	storedPort := cfg.Server.Port
	if h.storage != nil {
		if storedPortStr, _ := h.storage.GetConfig("server.port"); storedPortStr != "" {
			if parsed, err := strconv.Atoi(storedPortStr); err == nil && parsed > 0 {
				storedPort = parsed
			}
		}
	}

	// Get pool config from storage (override config.yaml values)
	poolStrategy := string(cfg.Pool.Strategy)
	poolCooldown := cfg.Pool.CooldownSeconds
	poolMaxRetries := cfg.Pool.MaxRetries
	if h.storage != nil {
		if storedStrategy, _ := h.storage.GetConfig("pool.strategy"); storedStrategy != "" {
			poolStrategy = storedStrategy
		}
		if storedCooldown, _ := h.storage.GetConfig("pool.cooldown_seconds"); storedCooldown != "" {
			if parsed, err := strconv.Atoi(storedCooldown); err == nil {
				poolCooldown = parsed
			}
		}
		if storedMaxRetries, _ := h.storage.GetConfig("pool.max_retries"); storedMaxRetries != "" {
			if parsed, err := strconv.Atoi(storedMaxRetries); err == nil {
				poolMaxRetries = parsed
			}
		}
	}

	// Get logging level from storage (override config.yaml values)
	loggingLevel := string(cfg.Logging.Level)
	if h.storage != nil {
		if storedLevel, _ := h.storage.GetConfig("logging.level"); storedLevel != "" {
			loggingLevel = storedLevel
		}
	}

	// Get model settings from storage
	var modelSettings gin.H = gin.H{
		"system_prompt":     "",
		"temperature":       nil,
		"max_output_tokens": nil,
		"top_p":             nil,
		"top_k":             nil,
		"thinking_level":    nil,
		"media_resolution":  nil,
	}
	if h.storage != nil {
		if sp, _ := h.storage.GetConfig("model_settings.system_prompt"); sp != "" {
			modelSettings["system_prompt"] = sp
		}
		if temp, _ := h.storage.GetConfig("model_settings.temperature"); temp != "" {
			if parsed, err := strconv.ParseFloat(temp, 64); err == nil {
				modelSettings["temperature"] = parsed
			}
		}
		if tokens, _ := h.storage.GetConfig("model_settings.max_output_tokens"); tokens != "" {
			if parsed, err := strconv.Atoi(tokens); err == nil {
				modelSettings["max_output_tokens"] = parsed
			}
		}
		if topP, _ := h.storage.GetConfig("model_settings.top_p"); topP != "" {
			if parsed, err := strconv.ParseFloat(topP, 64); err == nil {
				modelSettings["top_p"] = parsed
			}
		}
		if topK, _ := h.storage.GetConfig("model_settings.top_k"); topK != "" {
			if parsed, err := strconv.Atoi(topK); err == nil {
				modelSettings["top_k"] = parsed
			}
		}
		if level, _ := h.storage.GetConfig("model_settings.thinking_level"); level != "" {
			modelSettings["thinking_level"] = level
		}
		if resolution, _ := h.storage.GetConfig("model_settings.media_resolution"); resolution != "" {
			modelSettings["media_resolution"] = resolution
		}
	}

	// Return sanitized config (without sensitive data)
	sanitized := gin.H{
		"server": gin.H{
			"port":        cfg.Server.Port,
			"stored_port": storedPort,
			"host":        cfg.Server.Host,
		},
		"pool": gin.H{
			"strategy":         poolStrategy,
			"cooldown_seconds": poolCooldown,
			"max_retries":      poolMaxRetries,
		},
		"logging": gin.H{
			"level": loggingLevel,
		},
		"update": gin.H{
			"enabled":        cfg.Update.Enabled,
			"check_interval": cfg.Update.CheckInterval,
			"source":         updateSource,
		},
		"security": gin.H{
			"ip_whitelist_enabled": ipWhitelistEnabled,
			"whitelist_ip":         whitelistIP,
			"proxy_key":            proxyKey,
		},
		"advanced": gin.H{
			"request_timeout": requestTimeout,
		},
		"model_settings": modelSettings,
	}

	RespondSuccess(c, sanitized)
}

// UpdateConfigRequest represents the request to update configuration.
type UpdateConfigRequest struct {
	Server        *ServerConfigUpdate        `json:"server,omitempty"`
	Pool          *PoolConfigUpdate          `json:"pool,omitempty"`
	Logging       *LoggingConfigUpdate       `json:"logging,omitempty"`
	Update        *UpdateConfigUpdate        `json:"update,omitempty"`
	Security      *SecurityConfigUpdate      `json:"security,omitempty"`
	Advanced      *AdvancedConfigUpdate      `json:"advanced,omitempty"`
	ModelSettings *ModelSettingsConfigUpdate `json:"model_settings,omitempty"`
}

// ServerConfigUpdate represents server configuration updates.
type ServerConfigUpdate struct {
	Port *int `json:"port,omitempty"`
}

// PoolConfigUpdate represents pool configuration updates.
type PoolConfigUpdate struct {
	Strategy        *string `json:"strategy,omitempty"`
	CooldownSeconds *int    `json:"cooldown_seconds,omitempty"`
	MaxRetries      *int    `json:"max_retries,omitempty"`
}

// LoggingConfigUpdate represents logging configuration updates.
type LoggingConfigUpdate struct {
	Level *string `json:"level,omitempty"`
}

// UpdateConfigUpdate represents update service configuration.
type UpdateConfigUpdate struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Source  *string `json:"source,omitempty"`
}

// SecurityConfigUpdate represents security configuration updates.
type SecurityConfigUpdate struct {
	IPWhitelistEnabled *bool   `json:"ip_whitelist_enabled,omitempty"`
	WhitelistIP        *string `json:"whitelist_ip,omitempty"`
	ProxyKey           *string `json:"proxy_key,omitempty"`
}

// AdvancedConfigUpdate represents advanced configuration updates.
type AdvancedConfigUpdate struct {
	RequestTimeout *int `json:"request_timeout,omitempty"`
}

// ModelSettingsConfigUpdate represents model settings configuration updates.
type ModelSettingsConfigUpdate struct {
	SystemPrompt    *string  `json:"system_prompt,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"max_output_tokens,omitempty"`
	TopP            *float64 `json:"top_p,omitempty"`
	TopK            *int     `json:"top_k,omitempty"`
	ThinkingLevel   *string  `json:"thinking_level,omitempty"`
	MediaResolution *string  `json:"media_resolution,omitempty"`
}

// UpdateConfig handles PUT /api/config - Update configuration.
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	updated := make(map[string]interface{})

	// Process server configuration updates (requires restart)
	if req.Server != nil && req.Server.Port != nil {
		port := *req.Server.Port
		if port < 1024 || port > 65535 {
			RespondBadRequest(c, "Port must be between 1024 and 65535")
			return
		}

		if h.storage != nil {
			_ = h.storage.SetConfig("server.port", strconv.Itoa(port))
		}
		updated["server.port"] = port
	}

	// Process pool configuration updates
	if req.Pool != nil {
		// Update Strategy
		if req.Pool.Strategy != nil {
			strategyName := *req.Pool.Strategy
			validStrategies := map[string]bool{"round_robin": true, "random": true, "least_used": true, "weighted": true}
			if !validStrategies[strategyName] {
				RespondBadRequest(c, "Invalid strategy: "+strategyName)
				return
			}

			// Apply hot update to KeyPool
			var strategy keypool.Strategy
			switch strategyName {
			case "round_robin":
				strategy = keypool.NewRoundRobinStrategy()
			case "random":
				strategy = keypool.NewRandomStrategy()
			case "least_used":
				strategy = keypool.NewLeastUsedStrategy()
			case "weighted":
				strategy = keypool.NewWeightedStrategy()
			}
			h.pool.SetStrategy(strategy)

			// Persist to storage
			if h.storage != nil {
				_ = h.storage.SetConfig("pool.strategy", strategyName)
			}
			updated["pool.strategy"] = strategyName
		}

		// Update Cooldown Seconds
		if req.Pool.CooldownSeconds != nil {
			cooldown := *req.Pool.CooldownSeconds
			if cooldown < 0 {
				RespondBadRequest(c, "cooldown_seconds must be non-negative")
				return
			}

			h.pool.SetCooldownSeconds(cooldown)

			if h.storage != nil {
				_ = h.storage.SetConfig("pool.cooldown_seconds", strconv.Itoa(cooldown))
			}
			updated["pool.cooldown_seconds"] = cooldown
		}

		// Update Max Retries
		if req.Pool.MaxRetries != nil {
			maxRetries := *req.Pool.MaxRetries
			if maxRetries < 0 {
				RespondBadRequest(c, "max_retries must be non-negative")
				return
			}

			h.pool.SetMaxConsecutiveFailures(maxRetries)

			if h.storage != nil {
				_ = h.storage.SetConfig("pool.max_retries", strconv.Itoa(maxRetries))
			}
			updated["pool.max_retries"] = maxRetries
		}
	}

	// Process logging configuration updates
	if req.Logging != nil && req.Logging.Level != nil {
		level := *req.Logging.Level
		validLevels := map[string]logrus.Level{
			"debug": logrus.DebugLevel,
			"info":  logrus.InfoLevel,
			"warn":  logrus.WarnLevel,
			"error": logrus.ErrorLevel,
		}

		logLevel, ok := validLevels[level]
		if !ok {
			RespondBadRequest(c, "Invalid log level: "+level)
			return
		}

		// Apply hot update to logger
		h.logger.SetLevel(logLevel)

		if h.storage != nil {
			_ = h.storage.SetConfig("logging.level", level)
		}
		updated["logging.level"] = level
	}

	// Process update configuration
	if req.Update != nil {
		if req.Update.Enabled != nil {
			if h.storage != nil {
				_ = h.storage.SetConfig("update.enabled", strconv.FormatBool(*req.Update.Enabled))
			}
			updated["update.enabled"] = *req.Update.Enabled
		}

		if req.Update.Source != nil {
			source := *req.Update.Source
			if source != "mxln" && source != "github" {
				RespondBadRequest(c, "Invalid update source: "+source)
				return
			}

			if h.storage != nil {
				_ = h.storage.SetConfig("update.source", source)
			}
			updated["update.source"] = source
		}
	}

	// Process security configuration
	if req.Security != nil {
		if req.Security.IPWhitelistEnabled != nil {
			if h.storage != nil {
				_ = h.storage.SetConfig("security.ip_whitelist_enabled", strconv.FormatBool(*req.Security.IPWhitelistEnabled))
			}
			updated["security.ip_whitelist_enabled"] = *req.Security.IPWhitelistEnabled
		}

		if req.Security.WhitelistIP != nil {
			if h.storage != nil {
				_ = h.storage.SetConfig("security.whitelist_ip", *req.Security.WhitelistIP)
			}
			updated["security.whitelist_ip"] = *req.Security.WhitelistIP
		}

		if req.Security.ProxyKey != nil {
			proxyKey := *req.Security.ProxyKey
			// Validate proxy key format (must start with sk-mxln-)
			if proxyKey != "" && len(proxyKey) < 8 {
				RespondBadRequest(c, "Proxy key must be at least 8 characters")
				return
			}

			if h.storage != nil {
				_ = h.storage.SetConfig("security.proxy_key", proxyKey)
			}
			updated["security.proxy_key"] = proxyKey
		}
	}

	// Process advanced configuration
	if req.Advanced != nil {
		if req.Advanced.RequestTimeout != nil {
			timeout := *req.Advanced.RequestTimeout
			if timeout < 30 || timeout > 600 {
				RespondBadRequest(c, "Request timeout must be between 30 and 600 seconds")
				return
			}

			if h.storage != nil {
				_ = h.storage.SetConfig("advanced.request_timeout", strconv.Itoa(timeout))
			}
			updated["advanced.request_timeout"] = timeout
		}
	}

	// Process model settings configuration
	if req.ModelSettings != nil {
		if req.ModelSettings.SystemPrompt != nil {
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.system_prompt", *req.ModelSettings.SystemPrompt)
			}
			updated["model_settings.system_prompt"] = *req.ModelSettings.SystemPrompt
		}

		if req.ModelSettings.Temperature != nil {
			temp := *req.ModelSettings.Temperature
			if temp < 0 || temp > 2 {
				RespondBadRequest(c, "Temperature must be between 0 and 2")
				return
			}
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.temperature", strconv.FormatFloat(temp, 'f', 2, 64))
			}
			updated["model_settings.temperature"] = temp
		}

		if req.ModelSettings.MaxOutputTokens != nil {
			tokens := *req.ModelSettings.MaxOutputTokens
			if tokens < 1 || tokens > 65536 {
				RespondBadRequest(c, "Max output tokens must be between 1 and 65536")
				return
			}
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.max_output_tokens", strconv.Itoa(tokens))
			}
			updated["model_settings.max_output_tokens"] = tokens
		}

		if req.ModelSettings.TopP != nil {
			topP := *req.ModelSettings.TopP
			if topP < 0 || topP > 1 {
				RespondBadRequest(c, "Top-P must be between 0 and 1")
				return
			}
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.top_p", strconv.FormatFloat(topP, 'f', 2, 64))
			}
			updated["model_settings.top_p"] = topP
		}

		if req.ModelSettings.TopK != nil {
			topK := *req.ModelSettings.TopK
			if topK < 1 || topK > 100 {
				RespondBadRequest(c, "Top-K must be between 1 and 100")
				return
			}
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.top_k", strconv.Itoa(topK))
			}
			updated["model_settings.top_k"] = topK
		}

		if req.ModelSettings.ThinkingLevel != nil {
			level := *req.ModelSettings.ThinkingLevel
			validLevels := map[string]bool{"LOW": true, "MEDIUM": true, "HIGH": true, "": true}
			if !validLevels[level] {
				RespondBadRequest(c, "Thinking level must be LOW, MEDIUM, HIGH, or empty")
				return
			}
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.thinking_level", level)
			}
			updated["model_settings.thinking_level"] = level
		}

		if req.ModelSettings.MediaResolution != nil {
			resolution := *req.ModelSettings.MediaResolution
			validResolutions := map[string]bool{
				"MEDIA_RESOLUTION_LOW":    true,
				"MEDIA_RESOLUTION_MEDIUM": true,
				"MEDIA_RESOLUTION_HIGH":   true,
				"":                        true,
			}
			if !validResolutions[resolution] {
				RespondBadRequest(c, "Invalid media resolution")
				return
			}
			if h.storage != nil {
				_ = h.storage.SetConfig("model_settings.media_resolution", resolution)
			}
			updated["model_settings.media_resolution"] = resolution
		}
	}

	h.logger.WithFields(logrus.Fields{
		"updated": updated,
	}).Info("Configuration updated successfully")

	RespondSuccessWithMessage(c, gin.H{"updated": updated}, "Configuration updated successfully")
}

// ==================== Update Check ====================

// Update source URLs
const (
	MxlnUpdateURL   = "https://mxlnuma.space/muxueTools/update/latest.json"
	GitHubUpdateURL = "https://api.github.com/repos/muxueliunian/muxueTools/releases/latest"
)

// UpdateCheckResponse represents the update check result.
type UpdateCheckResponse struct {
	Success bool            `json:"success"`
	Data    UpdateCheckData `json:"data"`
}

// UpdateCheckData contains update information.
type UpdateCheckData struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	DownloadURL    string `json:"download_url,omitempty"`
	Changelog      string `json:"changelog,omitempty"`
	ReleaseDate    string `json:"release_date,omitempty"`
	Source         string `json:"source,omitempty"`
}

// MxlnUpdateInfo represents the structure of latest.json from mxln server.
type MxlnUpdateInfo struct {
	Version             string            `json:"version"`
	ReleaseDate         string            `json:"release_date"`
	Changelog           string            `json:"changelog"`
	Downloads           map[string]string `json:"downloads"`
	MinSupportedVersion string            `json:"min_supported_version"`
	IsCritical          bool              `json:"is_critical"`
	Announcement        string            `json:"announcement"`
}

// GitHubReleaseInfo represents the GitHub API release response.
type GitHubReleaseInfo struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// CheckUpdate handles GET /api/update/check - Check for updates.
func (h *AdminHandler) CheckUpdate(c *gin.Context) {
	// Get current version (set at build time or default to "dev")
	currentVersion := config.GetVersion()
	if currentVersion == "" {
		currentVersion = "dev"
	}

	// Get update source from storage
	source := "mxln" // default
	if h.storage != nil {
		if storedSource, _ := h.storage.GetConfig("update.source"); storedSource != "" {
			source = storedSource
		}
	}

	var data UpdateCheckData
	var err error

	if source == "github" {
		data, err = h.fetchGitHubUpdate(currentVersion)
	} else {
		data, err = h.fetchMxlnUpdate(currentVersion)
	}

	if err != nil {
		h.logger.WithError(err).WithField("source", source).Warn("Failed to check for updates")
		// Return current version info without update
		data = UpdateCheckData{
			CurrentVersion: currentVersion,
			LatestVersion:  currentVersion,
			HasUpdate:      false,
			Source:         source,
		}
	}

	c.JSON(http.StatusOK, UpdateCheckResponse{
		Success: true,
		Data:    data,
	})
}

// fetchMxlnUpdate fetches update info from mxln server.
func (h *AdminHandler) fetchMxlnUpdate(currentVersion string) (UpdateCheckData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", MxlnUpdateURL, nil)
	if err != nil {
		return UpdateCheckData{}, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return UpdateCheckData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateCheckData{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var info MxlnUpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return UpdateCheckData{}, err
	}

	// Get download URL for windows-amd64
	downloadURL := ""
	if url, ok := info.Downloads["windows-amd64"]; ok {
		downloadURL = url
	}

	return UpdateCheckData{
		CurrentVersion: currentVersion,
		LatestVersion:  info.Version,
		HasUpdate:      compareVersions(info.Version, currentVersion) > 0,
		DownloadURL:    downloadURL,
		Changelog:      info.Changelog,
		ReleaseDate:    info.ReleaseDate,
		Source:         "mxln",
	}, nil
}

// fetchGitHubUpdate fetches update info from GitHub Releases.
func (h *AdminHandler) fetchGitHubUpdate(currentVersion string) (UpdateCheckData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", GitHubUpdateURL, nil)
	if err != nil {
		return UpdateCheckData{}, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "MuxueTools-Updater")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return UpdateCheckData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UpdateCheckData{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var info GitHubReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return UpdateCheckData{}, err
	}

	// Extract version from tag (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(info.TagName, "v")

	// Find windows-amd64 asset
	downloadURL := info.HTMLURL
	for _, asset := range info.Assets {
		if strings.Contains(asset.Name, "windows-amd64") {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	return UpdateCheckData{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		HasUpdate:      compareVersions(latestVersion, currentVersion) > 0,
		DownloadURL:    downloadURL,
		Changelog:      info.Body,
		ReleaseDate:    info.PublishedAt,
		Source:         "github",
	}, nil
}

// compareVersions compares two semantic versions.
// Returns: 1 if v1 > v2, -1 if v1 < v2, 0 if equal.
func compareVersions(v1, v2 string) int {
	// Handle "dev" version
	if v1 == "dev" || v2 == "dev" {
		if v1 == v2 {
			return 0
		}
		if v1 == "dev" {
			return -1 // dev is always "older"
		}
		return 1
	}

	// Parse version strings
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			n1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			n2, _ = strconv.Atoi(parts2[i])
		}

		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}

	return 0
}

// RegenerateProxyKey handles POST /api/config/regenerate-proxy-key - Generate a new proxy key.
func (h *AdminHandler) RegenerateProxyKey(c *gin.Context) {
	// Generate new proxy key
	newKey := GenerateProxyKey()

	// Save to storage
	if h.storage != nil {
		if err := h.storage.SetConfig("security.proxy_key", newKey); err != nil {
			h.logger.WithError(err).Error("Failed to save new proxy key")
			RespondInternalError(c, "Failed to save new proxy key")
			return
		}
	}

	h.logger.Info("Proxy key regenerated successfully")

	RespondSuccess(c, gin.H{
		"proxy_key": newKey,
	})
}

// ResetStats handles DELETE /api/stats/reset - Reset all key statistics.
func (h *AdminHandler) ResetStats(c *gin.Context) {
	if h.storage == nil {
		RespondInternalError(c, "Storage not configured")
		return
	}

	count, err := h.storage.ResetKeyStats()
	if err != nil {
		h.logger.WithError(err).Error("Failed to reset stats")
		RespondInternalError(c, "Failed to reset statistics")
		return
	}

	h.logger.WithField("keys_affected", count).Info("Statistics reset successfully")

	RespondSuccess(c, gin.H{
		"message":       "Statistics reset successfully",
		"keys_affected": count,
	})
}
