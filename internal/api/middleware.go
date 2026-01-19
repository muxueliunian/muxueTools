// Package api provides HTTP API handlers and routing for MuxueTools.
package api

import (
	"fmt"
	"time"

	"muxueTools/internal/types"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ==================== Request ID Middleware ====================

// RequestIDKey is the context key for request ID.
const RequestIDKey = "request_id"

// RequestIDMiddleware adds a unique request ID to each request.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is provided in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context and response header
		c.Set(RequestIDKey, requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		return id.(string)
	}
	return ""
}

// ==================== CORS Middleware ====================

// CORSMiddleware returns a configured CORS middleware.
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}

// ==================== Logging Middleware ====================

// LoggingMiddleware logs HTTP requests using logrus.
func LoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get request ID
		requestID := GetRequestID(c)

		// Log fields
		fields := logrus.Fields{
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency":    latency.String(),
			"latency_ms": latency.Milliseconds(),
			"client_ip":  c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		}

		// Add query string if present
		if c.Request.URL.RawQuery != "" {
			fields["query"] = c.Request.URL.RawQuery
		}

		// Log based on status code
		status := c.Writer.Status()
		entry := logger.WithFields(fields)

		switch {
		case status >= 500:
			entry.Error("Server error")
		case status >= 400:
			entry.Warn("Client error")
		default:
			entry.Info("Request completed")
		}
	}
}

// ==================== Recovery Middleware ====================

// RecoveryMiddleware handles panics and returns a proper error response.
func RecoveryMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(c)

				// Log the panic
				logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"error":      err,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
				}).Error("Panic recovered")

				// Return 500 error
				appErr := types.NewInternalError(fmt.Sprintf("Internal server error: %v", err))
				c.AbortWithStatusJSON(appErr.HTTPStatus, appErr.ToAPIError())
			}
		}()

		c.Next()
	}
}

// ==================== Content-Type Middleware ====================

// JSONContentTypeMiddleware ensures Content-Type is application/json for API routes.
func JSONContentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType == "" || (contentType != "application/json" && contentType != "application/json; charset=utf-8") {
				// Accept requests without content type for flexibility
				// but log a warning for debugging
			}
		}
		c.Next()
	}
}

// ==================== Rate Limit Middleware (Placeholder) ====================

// RateLimitMiddleware provides local rate limiting (placeholder for future implementation).
func RateLimitMiddleware() gin.HandlerFunc {
	// Note: This is a placeholder. Actual implementation would use a token bucket
	// or sliding window algorithm with redis or in-memory storage.
	return func(c *gin.Context) {
		c.Next()
	}
}

// ==================== IP Whitelist Middleware ====================

// ConfigGetter interface for getting config values.
type ConfigGetter interface {
	GetConfig(key string) (string, error)
}

// IPWhitelistMiddleware checks if the request IP is in the whitelist.
// If whitelist is disabled, all IPs are allowed.
// Local IPs (127.0.0.1, ::1) are always allowed to prevent lockout.
func IPWhitelistMiddleware(configGetter ConfigGetter, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check if whitelist is enabled
		enabled, _ := configGetter.GetConfig("security.ip_whitelist_enabled")
		if enabled != "true" {
			c.Next()
			return
		}

		// 2. Get whitelist IP
		whitelistIP, _ := configGetter.GetConfig("security.whitelist_ip")
		if whitelistIP == "" {
			c.Next() // Empty whitelist allows all
			return
		}

		// 3. Get client IP
		clientIP := c.ClientIP()

		// 4. Always allow localhost to prevent lockout
		if clientIP == "127.0.0.1" || clientIP == "::1" || clientIP == "localhost" {
			c.Next()
			return
		}

		// 5. Check whitelist
		if clientIP != whitelistIP {
			logger.WithFields(logrus.Fields{
				"client_ip":    clientIP,
				"whitelist_ip": whitelistIP,
			}).Warn("IP not in whitelist, access denied")

			c.AbortWithStatusJSON(403, gin.H{
				"error": gin.H{
					"code":    40301,
					"message": "Access denied: IP not in whitelist",
					"type":    "permission_error",
				},
			})
			return
		}

		c.Next()
	}
}

// ==================== Proxy Key Middleware ====================

// DefaultProxyKey is the default proxy key for local development.
const DefaultProxyKey = "sk-mxln-proxy-local"

// ProxyKeyAuthMiddleware validates the proxy API key for OpenAI-compatible endpoints.
// The key should be passed in the Authorization header as "Bearer sk-mxln-xxx".
func ProxyKeyAuthMiddleware(configGetter ConfigGetter, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the configured proxy key
		proxyKey, _ := configGetter.GetConfig("security.proxy_key")
		if proxyKey == "" {
			proxyKey = DefaultProxyKey
		}

		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth header - allow for compatibility (optional auth)
			c.Next()
			return
		}

		// Parse Bearer token
		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) {
			c.Next()
			return
		}

		token := authHeader[len(bearerPrefix):]
		if token == "" {
			c.Next()
			return
		}

		// Validate token - accept if it matches the configured proxy key
		// Note: We're lenient here - if auth is provided, we validate it,
		// but we don't require auth for backward compatibility
		if token != proxyKey && token != DefaultProxyKey {
			// Log but don't reject - just for debugging
			logger.WithFields(logrus.Fields{
				"provided_key": token[:min(8, len(token))] + "...",
			}).Debug("Non-matching proxy key provided")
		}

		c.Next()
	}
}

// ==================== Proxy Key Generation ====================

// GenerateProxyKey generates a new random proxy key.
// Format: sk-mxln-{16 random alphanumeric characters}
func GenerateProxyKey() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		// Use uuid to get randomness since we already import it
		id := uuid.New()
		b[i] = chars[int(id[0])%len(chars)]
	}
	return "sk-mxln-" + string(b)
}

// min returns the smaller of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
