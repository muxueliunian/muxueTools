// Package types defines all data transfer objects and core types for MuxueTools.
package types

import "time"

// ==================== Session Types ====================

// Session represents a chat session.
type Session struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Title        string    `json:"title" gorm:"type:varchar(255)"`
	Model        string    `json:"model" gorm:"type:varchar(100)"`
	MessageCount int       `json:"message_count" gorm:"default:0"`
	TotalTokens  int       `json:"total_tokens" gorm:"default:0"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for Session.
func (Session) TableName() string {
	return "sessions"
}

// ChatMessage represents a stored message in a session.
// Named ChatMessage to avoid conflict with types.Message (OpenAI request format).
type ChatMessage struct {
	ID               string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	SessionID        string    `json:"session_id" gorm:"type:varchar(36);index"`
	Role             string    `json:"role" gorm:"type:varchar(20)"` // user/assistant/system
	Content          string    `json:"content" gorm:"type:text"`
	PromptTokens     int       `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int       `json:"completion_tokens" gorm:"default:0"`
	CreatedAt        time.Time `json:"created_at"`
}

// TableName specifies the table name for ChatMessage.
func (ChatMessage) TableName() string {
	return "messages"
}

// TotalTokens returns the total tokens for this message.
func (m *ChatMessage) TotalTokens() int {
	return m.PromptTokens + m.CompletionTokens
}

// ==================== Session API DTOs ====================

// CreateSessionRequest represents the request body for POST /api/sessions.
type CreateSessionRequest struct {
	Title string `json:"title"`
	Model string `json:"model"`
}

// CreateSessionResponse represents the response for POST /api/sessions.
type CreateSessionResponse struct {
	Success bool    `json:"success"`
	Data    Session `json:"data"`
}

// SessionListResponse represents the response for GET /api/sessions.
type SessionListResponse struct {
	Success  bool      `json:"success"`
	Sessions []Session `json:"sessions"`
	Total    int       `json:"total"`
}

// SessionDetailResponse represents the response for GET /api/sessions/:id.
type SessionDetailResponse struct {
	Success  bool          `json:"success"`
	Session  Session       `json:"session"`
	Messages []ChatMessage `json:"messages"`
}

// AddMessageRequest represents the request body for POST /api/sessions/:id/messages.
type AddMessageRequest struct {
	Role             string `json:"role" binding:"required"`
	Content          string `json:"content" binding:"required"`
	PromptTokens     int    `json:"prompt_tokens,omitempty"`
	CompletionTokens int    `json:"completion_tokens,omitempty"`
}

// AddMessageResponse represents the response for POST /api/sessions/:id/messages.
type AddMessageResponse struct {
	Success bool        `json:"success"`
	Data    ChatMessage `json:"data"`
}

// UpdateSessionRequest represents the request body for PUT /api/sessions/:id.
type UpdateSessionRequest struct {
	Title *string `json:"title,omitempty"`
	Model *string `json:"model,omitempty"`
}

// UpdateSessionResponse represents the response for PUT /api/sessions/:id.
type UpdateSessionResponse struct {
	Success bool    `json:"success"`
	Data    Session `json:"data"`
}

// DeleteSessionResponse represents the response for DELETE /api/sessions/:id.
type DeleteSessionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
