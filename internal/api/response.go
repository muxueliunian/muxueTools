// Package api provides HTTP API handlers and routing for MxlnAPI.
package api

import (
	"net/http"

	"mxlnapi/internal/types"

	"github.com/gin-gonic/gin"
)

// ==================== Standard Response Types ====================

// JSONResult represents a unified JSON response structure.
type JSONResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ==================== Response Helpers ====================

// RespondSuccess sends a successful JSON response.
func RespondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, JSONResult{
		Success: true,
		Data:    data,
	})
}

// RespondSuccessWithMessage sends a successful JSON response with a message.
func RespondSuccessWithMessage(c *gin.Context, data interface{}, message string) {
	c.JSON(http.StatusOK, JSONResult{
		Success: true,
		Data:    data,
		Message: message,
	})
}

// RespondError sends an error response using AppError.
func RespondError(c *gin.Context, err *types.AppError) {
	c.JSON(err.HTTPStatus, err.ToAPIError())
}

// RespondErrorWithStatus sends an error response with a custom HTTP status.
func RespondErrorWithStatus(c *gin.Context, status int, err *types.AppError) {
	c.JSON(status, err.ToAPIError())
}

// RespondBadRequest sends a 400 Bad Request error.
func RespondBadRequest(c *gin.Context, message string) {
	RespondError(c, types.NewInvalidRequestError(message))
}

// RespondNotFound sends a 404 Not Found error.
func RespondNotFound(c *gin.Context, resource string) {
	RespondError(c, types.NewNotFoundError(resource))
}

// RespondInternalError sends a 500 Internal Server Error.
func RespondInternalError(c *gin.Context, message string) {
	RespondError(c, types.NewInternalError(message))
}

// RespondRateLimited sends a 429 Too Many Requests error.
func RespondRateLimited(c *gin.Context, retryAfter int) {
	err := types.NewRateLimitError(retryAfter)
	c.Header("Retry-After", string(rune(retryAfter)))
	RespondError(c, err)
}

// RespondNoContent sends a 204 No Content response.
func RespondNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// RespondCreated sends a 201 Created response.
func RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, JSONResult{
		Success: true,
		Data:    data,
	})
}

// ==================== OpenAI Format Responses ====================

// RespondOpenAI sends a response in OpenAI API format (without wrapping).
func RespondOpenAI(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondOpenAIError sends an error in OpenAI API format.
func RespondOpenAIError(c *gin.Context, err *types.AppError) {
	c.JSON(err.HTTPStatus, err.ToAPIError())
}

// ==================== Streaming Response Helpers ====================

// SSEWriter provides helper methods for Server-Sent Events.
type SSEWriter struct {
	c *gin.Context
}

// NewSSEWriter creates a new SSE writer.
func NewSSEWriter(c *gin.Context) *SSEWriter {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	return &SSEWriter{c: c}
}

// WriteEvent writes a single SSE event.
func (w *SSEWriter) WriteEvent(data []byte) error {
	_, err := w.c.Writer.Write([]byte("data: "))
	if err != nil {
		return err
	}
	_, err = w.c.Writer.Write(data)
	if err != nil {
		return err
	}
	_, err = w.c.Writer.Write([]byte("\n\n"))
	if err != nil {
		return err
	}
	w.c.Writer.Flush()
	return nil
}

// WriteString writes a string as SSE data.
func (w *SSEWriter) WriteString(data string) error {
	return w.WriteEvent([]byte(data))
}

// WriteDone writes the SSE [DONE] marker.
func (w *SSEWriter) WriteDone() error {
	return w.WriteString("[DONE]")
}

// Flush flushes the response writer.
func (w *SSEWriter) Flush() {
	w.c.Writer.Flush()
}
