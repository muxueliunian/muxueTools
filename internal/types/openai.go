// Package types defines all data transfer objects and core types for MxlnAPI.
// This package contains type definitions only - no business logic.
package types

import "encoding/json"

// ==================== Chat Completion Request ====================

// ChatCompletionRequest represents an OpenAI-compatible chat completion request.
type ChatCompletionRequest struct {
	Model            string         `json:"model"`
	Messages         []Message      `json:"messages"`
	Temperature      *float64       `json:"temperature,omitempty"`
	TopP             *float64       `json:"top_p,omitempty"`
	MaxTokens        *int           `json:"max_tokens,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	Stop             StopSequence   `json:"stop,omitempty"`
	PresencePenalty  *float64       `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64       `json:"frequency_penalty,omitempty"`
	N                *int           `json:"n,omitempty"`
	User             string         `json:"user,omitempty"`
}

// Message represents a single message in the conversation.
type Message struct {
	Role    string          `json:"role"`    // "system", "user", or "assistant"
	Content json.RawMessage `json:"content"` // string or []ContentPart
}

// ContentPart represents a multimodal content part (text or image).
type ContentPart struct {
	Type     string    `json:"type"`               // "text" or "image_url"
	Text     string    `json:"text,omitempty"`     // for type="text"
	ImageURL *ImageURL `json:"image_url,omitempty"` // for type="image_url"
}

// ImageURL represents an image input with optional detail level.
type ImageURL struct {
	URL    string `json:"url"`              // base64 data URI or HTTP URL
	Detail string `json:"detail,omitempty"` // "low", "high", or "auto"
}

// StopSequence can be either a single string or an array of strings.
type StopSequence []string

// UnmarshalJSON implements custom unmarshaling for StopSequence.
// It handles both string and []string formats.
func (s *StopSequence) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a single string first
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}
	// Try to unmarshal as an array of strings
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*s = arr
	return nil
}

// MarshalJSON implements custom marshaling for StopSequence.
func (s StopSequence) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

// ==================== Chat Completion Response (Non-Streaming) ====================

// ChatCompletionResponse represents an OpenAI-compatible chat completion response.
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`  // "chat.completion"
	Created int64    `json:"created"` // Unix timestamp
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a single completion choice.
type Choice struct {
	Index        int              `json:"index"`
	Message      ResponseMessage  `json:"message"`
	FinishReason string           `json:"finish_reason"` // "stop", "length", "content_filter"
}

// ResponseMessage represents the assistant's response message.
type ResponseMessage struct {
	Role    string `json:"role"`    // Always "assistant"
	Content string `json:"content"`
}

// Usage represents token consumption statistics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ==================== Chat Completion Response (Streaming) ====================

// ChatCompletionChunk represents a single SSE chunk in streaming response.
type ChatCompletionChunk struct {
	ID      string        `json:"id"`
	Object  string        `json:"object"`  // "chat.completion.chunk"
	Created int64         `json:"created"` // Unix timestamp
	Model   string        `json:"model"`
	Choices []ChunkChoice `json:"choices"`
}

// ChunkChoice represents a single choice in a streaming chunk.
type ChunkChoice struct {
	Index        int    `json:"index"`
	Delta        Delta  `json:"delta"`
	FinishReason string `json:"finish_reason,omitempty"` // null or "stop", "length"
}

// Delta represents incremental content in a streaming chunk.
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// ==================== Models Endpoint ====================

// ModelsResponse represents the response for GET /v1/models.
type ModelsResponse struct {
	Object string      `json:"object"` // "list"
	Data   []ModelInfo `json:"data"`
}

// ModelInfo represents information about a single model.
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`  // "model"
	Created int64  `json:"created"` // Unix timestamp
	OwnedBy string `json:"owned_by"` // "google" for Gemini models
}

// ==================== Health Check ====================

// HealthResponse represents the response for GET /health.
type HealthResponse struct {
	Status  string         `json:"status"` // "ok" or "degraded"
	Version string         `json:"version"`
	Uptime  int64          `json:"uptime"` // Seconds since start
	Keys    KeyHealthStats `json:"keys"`
}

// KeyHealthStats provides a summary of key pool status.
type KeyHealthStats struct {
	Total       int `json:"total"`
	Active      int `json:"active"`
	RateLimited int `json:"rate_limited"`
	Disabled    int `json:"disabled"`
}

// ==================== Helper Methods ====================

// GetContentAsString attempts to extract the message content as a plain string.
// Returns the string and true if successful, empty string and false otherwise.
func (m *Message) GetContentAsString() (string, bool) {
	var s string
	if err := json.Unmarshal(m.Content, &s); err == nil {
		return s, true
	}
	return "", false
}

// GetContentAsParts attempts to extract the message content as []ContentPart.
// Returns the parts and true if successful, nil and false otherwise.
func (m *Message) GetContentAsParts() ([]ContentPart, bool) {
	var parts []ContentPart
	if err := json.Unmarshal(m.Content, &parts); err == nil {
		return parts, true
	}
	return nil, false
}

// NewTextContent creates a Message with plain text content.
func NewTextContent(role, text string) Message {
	content, _ := json.Marshal(text)
	return Message{
		Role:    role,
		Content: content,
	}
}

// NewMultimodalContent creates a Message with multimodal content parts.
func NewMultimodalContent(role string, parts []ContentPart) Message {
	content, _ := json.Marshal(parts)
	return Message{
		Role:    role,
		Content: content,
	}
}
