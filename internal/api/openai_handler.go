// Package api provides HTTP API handlers and routing for MxlnAPI.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mxlnapi/internal/gemini"
	"mxlnapi/internal/keypool"
	"mxlnapi/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ==================== OpenAI Handler ====================

// OpenAIHandler handles OpenAI-compatible API endpoints.
type OpenAIHandler struct {
	client *gemini.Client
	pool   *keypool.Pool
	logger *logrus.Logger
}

// NewOpenAIHandler creates a new OpenAI handler.
func NewOpenAIHandler(client *gemini.Client, pool *keypool.Pool, logger *logrus.Logger) *OpenAIHandler {
	return &OpenAIHandler{
		client: client,
		pool:   pool,
		logger: logger,
	}
}

// ==================== Chat Completions ====================

// ChatCompletions handles POST /v1/chat/completions.
func (h *OpenAIHandler) ChatCompletions(c *gin.Context) {
	requestID := GetRequestID(c)

	// Parse request body
	var req types.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Failed to parse chat completion request")

		appErr := types.NewInvalidRequestError("Invalid request body: " + err.Error())
		RespondOpenAIError(c, appErr)
		return
	}

	// Validate request
	if err := h.validateChatRequest(&req); err != nil {
		h.logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"error":      err.Error(),
		}).Warn("Invalid chat completion request")

		RespondOpenAIError(c, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"model":      req.Model,
		"stream":     req.Stream,
		"messages":   len(req.Messages),
	}).Debug("Processing chat completion request")

	// Handle streaming vs non-streaming
	if req.Stream {
		h.handleStreamingRequest(c, &req, requestID)
	} else {
		h.handleBlockingRequest(c, &req, requestID)
	}
}

// validateChatRequest validates the chat completion request.
func (h *OpenAIHandler) validateChatRequest(req *types.ChatCompletionRequest) *types.AppError {
	if req.Model == "" {
		return types.ErrMissingModel
	}

	if len(req.Messages) == 0 {
		return types.NewInvalidMessagesError("Messages array cannot be empty")
	}

	for i, msg := range req.Messages {
		if msg.Role == "" {
			return types.NewInvalidMessagesError(fmt.Sprintf("Message at index %d is missing role", i))
		}
		if msg.Role != "system" && msg.Role != "user" && msg.Role != "assistant" {
			return types.NewInvalidMessagesError("Invalid role: " + msg.Role)
		}
	}

	return nil
}

// handleBlockingRequest handles non-streaming chat completion requests.
func (h *OpenAIHandler) handleBlockingRequest(c *gin.Context, req *types.ChatCompletionRequest, requestID string) {
	ctx := c.Request.Context()

	resp, err := h.client.ChatCompletion(ctx, req)
	if err != nil {
		h.handleError(c, err, requestID)
		return
	}

	RespondOpenAI(c, resp)
}

// handleStreamingRequest handles streaming chat completion requests.
func (h *OpenAIHandler) handleStreamingRequest(c *gin.Context, req *types.ChatCompletionRequest, requestID string) {
	ctx := c.Request.Context()

	// Get the stream channel
	eventChan, err := h.client.ChatCompletionStream(ctx, req)
	if err != nil {
		h.handleError(c, err, requestID)
		return
	}

	// Create SSE writer
	sse := NewSSEWriter(c)

	// Stream events to client
	for event := range eventChan {
		// Check for context cancellation (client disconnect)
		select {
		case <-ctx.Done():
			h.logger.WithFields(logrus.Fields{
				"request_id": requestID,
			}).Debug("Client disconnected, stopping stream")
			return
		default:
		}

		if event.Err != nil {
			// Log error but continue (client may have disconnected)
			h.logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"error":      event.Err.Error(),
			}).Warn("Stream error")

			// Send error as SSE event if possible
			if appErr, ok := event.Err.(*types.AppError); ok {
				errJSON, _ := json.Marshal(appErr.ToAPIError())
				sse.WriteEvent(errJSON)
			}
			return
		}

		if event.Done {
			// Send [DONE] marker
			sse.WriteDone()
			return
		}

		if event.Chunk != nil {
			// Marshal chunk to JSON
			chunkJSON, err := json.Marshal(event.Chunk)
			if err != nil {
				h.logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Error("Failed to marshal stream chunk")
				continue
			}

			// Write SSE event
			if err := sse.WriteEvent(chunkJSON); err != nil {
				h.logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err.Error(),
				}).Debug("Failed to write SSE event (client may have disconnected)")
				return
			}
		}
	}

	// Ensure [DONE] is sent
	sse.WriteDone()
}

// handleError converts an error to an appropriate HTTP response.
func (h *OpenAIHandler) handleError(c *gin.Context, err error, requestID string) {
	h.logger.WithFields(logrus.Fields{
		"request_id": requestID,
		"error":      err.Error(),
	}).Error("Chat completion error")

	if appErr, ok := err.(*types.AppError); ok {
		RespondOpenAIError(c, appErr)
		return
	}

	// Wrap unknown errors
	appErr := types.NewInternalError(err.Error())
	RespondOpenAIError(c, appErr)
}

// ==================== Models Endpoint ====================

// ListModels handles GET /v1/models.
func (h *OpenAIHandler) ListModels(c *gin.Context) {
	// Define available models (these are Gemini models exposed as OpenAI-compatible)
	models := []types.ModelInfo{
		{
			ID:      "gpt-4",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gpt-4-turbo",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gpt-4o",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gpt-4o-mini",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gpt-3.5-turbo",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gemini-1.5-pro-latest",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gemini-1.5-flash-latest",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
		{
			ID:      "gemini-2.0-flash",
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "google",
		},
	}

	resp := types.ModelsResponse{
		Object: "list",
		Data:   models,
	}

	RespondOpenAI(c, resp)
}

// ==================== Health Check ====================

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	pool      *keypool.Pool
	startTime time.Time
	version   string
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(pool *keypool.Pool, version string) *HealthHandler {
	return &HealthHandler{
		pool:      pool,
		startTime: time.Now(),
		version:   version,
	}
}

// Health handles GET /health.
func (h *HealthHandler) Health(c *gin.Context) {
	// Calculate uptime
	uptime := int64(time.Since(h.startTime).Seconds())

	// Get key pool statistics
	keyStats := h.calculateKeyStats()

	// Determine status
	status := "ok"
	if keyStats.Total == 0 || keyStats.Active == 0 {
		status = "degraded"
	}

	resp := types.HealthResponse{
		Status:  status,
		Version: h.version,
		Uptime:  uptime,
		Keys:    keyStats,
	}

	c.JSON(http.StatusOK, resp)
}

// calculateKeyStats calculates key pool statistics.
func (h *HealthHandler) calculateKeyStats() types.KeyHealthStats {
	stats := h.pool.GetStats()

	result := types.KeyHealthStats{
		Total: len(stats),
	}

	for _, key := range stats {
		switch key.Status {
		case types.KeyStatusActive:
			if key.Enabled {
				result.Active++
			}
		case types.KeyStatusRateLimited:
			result.RateLimited++
		case types.KeyStatusDisabled:
			result.Disabled++
		}

		// Count disabled keys from Enabled flag
		if !key.Enabled && key.Status != types.KeyStatusDisabled {
			result.Disabled++
		}
	}

	return result
}

// ==================== Ping Endpoint ====================

// Ping handles a simple ping endpoint for basic connectivity testing.
func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
		"time":    time.Now().Unix(),
	})
}

// ==================== Debug Helpers ====================

// DebugRequest logs the full request for debugging purposes.
func DebugRequest(c *gin.Context, logger *logrus.Logger) {
	// Read body
	body, _ := io.ReadAll(c.Request.Body)

	logger.WithFields(logrus.Fields{
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"query":        c.Request.URL.RawQuery,
		"headers":      c.Request.Header,
		"body":         string(body),
		"content_type": c.ContentType(),
	}).Debug("Request details")
}
