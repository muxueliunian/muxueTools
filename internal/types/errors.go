// Package types defines all data transfer objects and core types for MxlnAPI.
package types

import (
	"fmt"
	"net/http"
)

// ==================== Error Codes ====================

// Error codes following the ARCHITECTURE.md specification.
const (
	// 4xx Client Errors
	ErrCodeInvalidRequest   = 40001 // Invalid request format
	ErrCodeUnsupportedModel = 40002 // Unsupported model
	ErrCodeInvalidMessages  = 40003 // Invalid messages format
	ErrCodeAuthentication   = 40101 // Missing or invalid API key
	ErrCodePermission       = 40301 // Key disabled or access denied
	ErrCodeNotFound         = 40401 // Resource not found
	ErrCodeRateLimit        = 42901 // All keys rate limited

	// 5xx Server Errors
	ErrCodeInternal           = 50001 // Internal server error
	ErrCodeUpstream           = 50201 // Upstream (Gemini) API error
	ErrCodeServiceUnavailable = 50301 // Service temporarily unavailable
)

// ==================== Error Types ====================

// Error type strings for categorization.
const (
	ErrTypeInvalidRequest     = "invalid_request_error"
	ErrTypeAuthentication     = "authentication_error"
	ErrTypePermission         = "permission_error"
	ErrTypeNotFound           = "not_found_error"
	ErrTypeRateLimit          = "rate_limit_error"
	ErrTypeServer             = "server_error"
	ErrTypeUpstream           = "upstream_error"
	ErrTypeServiceUnavailable = "service_unavailable"
)

// ==================== API Error Response ====================

// APIError represents the standard error response format.
// This is the top-level structure returned to clients.
type APIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains detailed error information.
type ErrorDetail struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
	Param      string `json:"param,omitempty"`       // Related parameter (if applicable)
	RetryAfter int    `json:"retry_after,omitempty"` // Seconds to wait (for 429)
}

// ==================== AppError ====================

// AppError is the internal error type used throughout the application.
// It satisfies the error interface and contains additional context.
type AppError struct {
	Code       int
	Message    string
	Type       string
	HTTPStatus int
	Param      string
	RetryAfter int
	Cause      error // Underlying error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// ToAPIError converts AppError to the client-facing APIError format.
func (e *AppError) ToAPIError() *APIError {
	return &APIError{
		Error: ErrorDetail{
			Code:       e.Code,
			Message:    e.Message,
			Type:       e.Type,
			Param:      e.Param,
			RetryAfter: e.RetryAfter,
		},
	}
}

// WithCause attaches an underlying error and returns the AppError.
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithParam attaches a parameter name and returns the AppError.
func (e *AppError) WithParam(param string) *AppError {
	e.Param = param
	return e
}

// WithMessage replaces the message and returns the AppError.
func (e *AppError) WithMessage(message string) *AppError {
	e.Message = message
	return e
}

// ==================== Pre-defined Errors ====================

// NewInvalidRequestError creates an error for invalid request format.
func NewInvalidRequestError(message string) *AppError {
	if message == "" {
		message = "Invalid request format"
	}
	return &AppError{
		Code:       ErrCodeInvalidRequest,
		Message:    message,
		Type:       ErrTypeInvalidRequest,
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewUnsupportedModelError creates an error for unsupported model.
func NewUnsupportedModelError(model string) *AppError {
	return &AppError{
		Code:       ErrCodeUnsupportedModel,
		Message:    fmt.Sprintf("The specified model '%s' is not supported", model),
		Type:       ErrTypeInvalidRequest,
		HTTPStatus: http.StatusBadRequest,
		Param:      "model",
	}
}

// NewInvalidMessagesError creates an error for invalid messages format.
func NewInvalidMessagesError(message string) *AppError {
	if message == "" {
		message = "Invalid messages format"
	}
	return &AppError{
		Code:       ErrCodeInvalidMessages,
		Message:    message,
		Type:       ErrTypeInvalidRequest,
		HTTPStatus: http.StatusBadRequest,
		Param:      "messages",
	}
}

// NewAuthenticationError creates an error for authentication failure.
func NewAuthenticationError(message string) *AppError {
	if message == "" {
		message = "Invalid or missing API key"
	}
	return &AppError{
		Code:       ErrCodeAuthentication,
		Message:    message,
		Type:       ErrTypeAuthentication,
		HTTPStatus: http.StatusUnauthorized,
	}
}

// NewPermissionError creates an error for permission denied.
func NewPermissionError(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return &AppError{
		Code:       ErrCodePermission,
		Message:    message,
		Type:       ErrTypePermission,
		HTTPStatus: http.StatusForbidden,
	}
}

// NewNotFoundError creates an error for resource not found.
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Code:       ErrCodeNotFound,
		Message:    fmt.Sprintf("Resource not found: %s", resource),
		Type:       ErrTypeNotFound,
		HTTPStatus: http.StatusNotFound,
	}
}

// NewRateLimitError creates an error when all keys are rate limited.
func NewRateLimitError(retryAfter int) *AppError {
	return &AppError{
		Code:       ErrCodeRateLimit,
		Message:    "All API keys are currently rate limited",
		Type:       ErrTypeRateLimit,
		HTTPStatus: http.StatusTooManyRequests,
		RetryAfter: retryAfter,
	}
}

// NewInternalError creates an error for internal server errors.
func NewInternalError(message string) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return &AppError{
		Code:       ErrCodeInternal,
		Message:    message,
		Type:       ErrTypeServer,
		HTTPStatus: http.StatusInternalServerError,
	}
}

// NewUpstreamError creates an error for Gemini API errors.
func NewUpstreamError(message string) *AppError {
	if message == "" {
		message = "Upstream API error"
	}
	return &AppError{
		Code:       ErrCodeUpstream,
		Message:    message,
		Type:       ErrTypeUpstream,
		HTTPStatus: http.StatusBadGateway,
	}
}

// NewServiceUnavailableError creates an error for service unavailability.
func NewServiceUnavailableError(message string) *AppError {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	return &AppError{
		Code:       ErrCodeServiceUnavailable,
		Message:    message,
		Type:       ErrTypeServiceUnavailable,
		HTTPStatus: http.StatusServiceUnavailable,
	}
}

// ==================== Sentinel Errors ====================

// Pre-defined sentinel errors for common error cases.
var (
	// ErrNoAvailableKeys indicates the key pool has no usable keys.
	ErrNoAvailableKeys = NewServiceUnavailableError("No available API keys in the pool")

	// ErrAllKeysRateLimited indicates all keys are in cooldown.
	ErrAllKeysRateLimited = NewRateLimitError(60)

	// ErrEmptyMessages indicates the messages array is empty.
	ErrEmptyMessages = NewInvalidMessagesError("Messages array cannot be empty")

	// ErrMissingModel indicates the model field is required.
	ErrMissingModel = NewInvalidRequestError("Model field is required").WithParam("model")

	// ErrKeyNotFound indicates a key was not found in the database.
	ErrKeyNotFound = NewNotFoundError("Key")

	// ErrSessionNotFound indicates a session was not found in the database.
	ErrSessionNotFound = NewNotFoundError("Session")
)

// ==================== Error Helpers ====================

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// AsAppError attempts to convert an error to an AppError.
// If the error is already an AppError, it returns it directly.
// Otherwise, it wraps it in a generic internal error.
func AsAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return NewInternalError(err.Error()).WithCause(err)
}

// HTTPStatusFromError extracts the HTTP status code from an error.
// Returns 500 for non-AppError types.
func HTTPStatusFromError(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
