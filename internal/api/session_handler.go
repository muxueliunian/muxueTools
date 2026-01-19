// Package api provides HTTP API handlers and routing for MuxueTools.
package api

import (
	"net/http"
	"strconv"
	"time"

	"muxueTools/internal/storage"
	"muxueTools/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ==================== Session Handler ====================

// SessionHandler handles session management API endpoints.
type SessionHandler struct {
	storage *storage.Storage
}

// NewSessionHandler creates a new session handler.
func NewSessionHandler(storage *storage.Storage) *SessionHandler {
	return &SessionHandler{
		storage: storage,
	}
}

// ListSessions handles GET /api/sessions - List all sessions with pagination.
func (h *SessionHandler) ListSessions(c *gin.Context) {
	// Parse pagination parameters
	limit := 20 // Default limit
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100 // Max limit
			}
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	sessions, total, err := h.storage.ListSessions(limit, offset)
	if err != nil {
		RespondInternalError(c, "Failed to list sessions")
		return
	}

	c.JSON(http.StatusOK, types.SessionListResponse{
		Success:  true,
		Sessions: sessions,
		Total:    total,
	})
}

// CreateSession handles POST /api/sessions - Create a new session.
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req types.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Set defaults
	if req.Title == "" {
		req.Title = "New Chat"
	}
	if req.Model == "" {
		req.Model = "gemini-1.5-flash"
	}

	session := &types.Session{
		ID:        uuid.New().String(),
		Title:     req.Title,
		Model:     req.Model,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.storage.CreateSession(session); err != nil {
		RespondInternalError(c, "Failed to create session")
		return
	}

	c.JSON(http.StatusCreated, types.CreateSessionResponse{
		Success: true,
		Data:    *session,
	})
}

// GetSession handles GET /api/sessions/:id - Get session with messages.
func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		RespondBadRequest(c, "Session ID is required")
		return
	}

	session, err := h.storage.GetSession(sessionID)
	if err != nil {
		if err == types.ErrSessionNotFound {
			RespondNotFound(c, "Session")
			return
		}
		RespondInternalError(c, "Failed to get session")
		return
	}

	messages, err := h.storage.GetMessages(sessionID)
	if err != nil {
		RespondInternalError(c, "Failed to get messages")
		return
	}

	c.JSON(http.StatusOK, types.SessionDetailResponse{
		Success:  true,
		Session:  *session,
		Messages: messages,
	})
}

// UpdateSession handles PUT /api/sessions/:id - Update session.
func (h *SessionHandler) UpdateSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		RespondBadRequest(c, "Session ID is required")
		return
	}

	var req types.UpdateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Get existing session
	session, err := h.storage.GetSession(sessionID)
	if err != nil {
		if err == types.ErrSessionNotFound {
			RespondNotFound(c, "Session")
			return
		}
		RespondInternalError(c, "Failed to get session")
		return
	}

	// Apply updates
	if req.Title != nil {
		session.Title = *req.Title
	}
	if req.Model != nil {
		session.Model = *req.Model
	}

	if err := h.storage.UpdateSession(session); err != nil {
		RespondInternalError(c, "Failed to update session")
		return
	}

	c.JSON(http.StatusOK, types.UpdateSessionResponse{
		Success: true,
		Data:    *session,
	})
}

// DeleteSession handles DELETE /api/sessions/:id - Delete session and messages.
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		RespondBadRequest(c, "Session ID is required")
		return
	}

	if err := h.storage.DeleteSession(sessionID); err != nil {
		if err == types.ErrSessionNotFound {
			RespondNotFound(c, "Session")
			return
		}
		RespondInternalError(c, "Failed to delete session")
		return
	}

	c.JSON(http.StatusOK, types.DeleteSessionResponse{
		Success: true,
		Message: "Session deleted successfully",
	})
}

// AddMessage handles POST /api/sessions/:id/messages - Add a message to session.
func (h *SessionHandler) AddMessage(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		RespondBadRequest(c, "Session ID is required")
		return
	}

	var req types.AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Validate role
	if req.Role != "user" && req.Role != "assistant" && req.Role != "system" {
		RespondBadRequest(c, "Invalid role: must be 'user', 'assistant', or 'system'")
		return
	}

	message := &types.ChatMessage{
		ID:               uuid.New().String(),
		SessionID:        sessionID,
		Role:             req.Role,
		Content:          req.Content,
		PromptTokens:     req.PromptTokens,
		CompletionTokens: req.CompletionTokens,
		CreatedAt:        time.Now(),
	}

	if err := h.storage.AddMessage(message); err != nil {
		if err == types.ErrSessionNotFound {
			RespondNotFound(c, "Session")
			return
		}
		RespondInternalError(c, "Failed to add message")
		return
	}

	c.JSON(http.StatusCreated, types.AddMessageResponse{
		Success: true,
		Data:    *message,
	})
}

// DeleteAllSessions handles DELETE /api/sessions - Delete all sessions and messages.
func (h *SessionHandler) DeleteAllSessions(c *gin.Context) {
	count, err := h.storage.DeleteAllSessions()
	if err != nil {
		RespondInternalError(c, "Failed to delete sessions")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All sessions deleted successfully",
		"deleted": count,
	})
}
