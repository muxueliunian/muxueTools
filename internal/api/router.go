// Package api provides HTTP API handlers and routing for MuxueTools.
package api

import (
	"muxueTools/internal/gemini"
	"muxueTools/internal/keypool"
	"muxueTools/internal/storage"
	"muxueTools/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ==================== Router Configuration ====================

// RouterConfig holds the dependencies needed to create the router.
type RouterConfig struct {
	Config  *types.Config
	Pool    *keypool.Pool
	Client  *gemini.Client
	Storage *storage.Storage // Optional: for session persistence
	Logger  *logrus.Logger
	Version string
	WebRoot string // Optional: path to static web files (e.g., "web/dist")
}

// NewRouter creates and configures a new Gin router with all routes.
func NewRouter(cfg *RouterConfig) *gin.Engine {
	// Set Gin mode based on log level
	if cfg.Config.Logging.Level == types.LogLevelDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create engine without default middleware
	engine := gin.New()

	// Apply custom middleware
	engine.Use(RequestIDMiddleware())
	engine.Use(CORSMiddleware())
	engine.Use(RecoveryMiddleware(cfg.Logger))
	engine.Use(LoggingMiddleware(cfg.Logger))

	// Create handlers
	openaiHandler := NewOpenAIHandler(cfg.Client, cfg.Pool, cfg.Logger)
	healthHandler := NewHealthHandler(cfg.Pool, cfg.Version)
	adminHandler := NewAdminHandler(cfg.Pool, cfg.Logger, cfg.Storage)

	// ==================== OpenAI Compatible Routes ====================
	// Apply IP whitelist middleware to protect API endpoints
	v1 := engine.Group("/v1")
	v1.Use(IPWhitelistMiddleware(cfg.Storage, cfg.Logger))
	{
		// Chat completions
		v1.POST("/chat/completions", openaiHandler.ChatCompletions)

		// Models
		v1.GET("/models", openaiHandler.ListModels)
	}

	// ==================== Health & Status Routes ====================
	engine.GET("/health", healthHandler.Health)
	engine.GET("/ping", Ping)

	// ==================== Admin API Routes ====================
	api := engine.Group("/api")
	{
		// Key management
		keys := api.Group("/keys")
		{
			keys.GET("", adminHandler.ListKeys)
			keys.POST("", adminHandler.AddKey)
			keys.DELETE("/:id", adminHandler.DeleteKey)
			keys.POST("/:id/test", adminHandler.TestKey)
			keys.POST("/validate", adminHandler.ValidateKey)
			keys.POST("/import", adminHandler.ImportKeys)
			keys.GET("/export", adminHandler.ExportKeys)
		}

		// Models
		api.GET("/models", adminHandler.ListAvailableModels)

		// Statistics
		api.GET("/stats", adminHandler.GetStats)
		api.GET("/stats/keys", adminHandler.GetKeyStats)
		api.GET("/stats/trend", adminHandler.GetStatsTrend)
		api.GET("/stats/models", adminHandler.GetStatsModels)
		api.DELETE("/stats/reset", adminHandler.ResetStats)

		// Configuration
		api.GET("/config", adminHandler.GetConfig)
		api.PUT("/config", adminHandler.UpdateConfig)
		api.POST("/config/regenerate-proxy-key", adminHandler.RegenerateProxyKey)

		// Update check
		api.GET("/update/check", adminHandler.CheckUpdate)

		// Session management (only if storage is configured)
		if cfg.Storage != nil {
			sessionHandler := NewSessionHandler(cfg.Storage)
			sessions := api.Group("/sessions")
			{
				sessions.GET("", sessionHandler.ListSessions)
				sessions.POST("", sessionHandler.CreateSession)
				sessions.DELETE("", sessionHandler.DeleteAllSessions) // 清空所有
				sessions.GET("/:id", sessionHandler.GetSession)
				sessions.PUT("/:id", sessionHandler.UpdateSession)
				sessions.DELETE("/:id", sessionHandler.DeleteSession)
				sessions.POST("/:id/messages", sessionHandler.AddMessage)
			}
		}
	}

	// ==================== Static Files & Root Route ====================
	if cfg.WebRoot != "" {
		// Serve static files from WebRoot directory
		engine.Static("/assets", cfg.WebRoot+"/assets")

		// Serve root-level static files
		engine.StaticFile("/logo.png", cfg.WebRoot+"/logo.png")
		engine.StaticFile("/favicon.png", cfg.WebRoot+"/favicon.png")
		engine.StaticFile("/vite.svg", cfg.WebRoot+"/vite.svg")

		// Serve index.html for root and SPA fallback
		engine.GET("/", func(c *gin.Context) {
			c.File(cfg.WebRoot + "/index.html")
		})

		// SPA fallback: serve index.html for unmatched routes
		engine.NoRoute(func(c *gin.Context) {
			// Only serve index.html for non-API/non-asset requests
			path := c.Request.URL.Path
			if len(path) > 0 && path[0] == '/' {
				// Check if it's an API route
				if len(path) >= 3 && (path[:3] == "/v1" || path[:4] == "/api") {
					c.JSON(404, gin.H{"error": "Not found"})
					return
				}
				// Serve index.html for SPA routes
				c.File(cfg.WebRoot + "/index.html")
				return
			}
			c.JSON(404, gin.H{"error": "Not found"})
		})

		cfg.Logger.WithField("webroot", cfg.WebRoot).Info("Static file serving enabled")
	} else {
		// No WebRoot: serve API info at root
		engine.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"name":    "MuxueTools",
				"version": cfg.Version,
				"docs":    "/v1/models",
				"health":  "/health",
				"admin":   "/api/keys",
			})
		})
	}

	return engine
}

// ==================== Route Group Builders ====================

// SetupOpenAIRoutes sets up OpenAI-compatible routes on the given router group.
func SetupOpenAIRoutes(group *gin.RouterGroup, handler *OpenAIHandler) {
	group.POST("/chat/completions", handler.ChatCompletions)
	group.GET("/models", handler.ListModels)
}

// SetupAdminRoutes sets up admin routes on the given router group.
func SetupAdminRoutes(group *gin.RouterGroup, handler *AdminHandler) {
	// Key management
	keys := group.Group("/keys")
	{
		keys.GET("", handler.ListKeys)
		keys.POST("", handler.AddKey)
		keys.DELETE("/:id", handler.DeleteKey)
		keys.POST("/:id/test", handler.TestKey)
		keys.POST("/validate", handler.ValidateKey)
		keys.POST("/import", handler.ImportKeys)
		keys.GET("/export", handler.ExportKeys)
	}

	// Statistics
	group.GET("/stats", handler.GetStats)
	group.GET("/stats/keys", handler.GetKeyStats)
	group.GET("/stats/trend", handler.GetStatsTrend)
	group.GET("/stats/models", handler.GetStatsModels)

	// Configuration
	group.GET("/config", handler.GetConfig)
	group.PUT("/config", handler.UpdateConfig)

	// Update check
	group.GET("/update/check", handler.CheckUpdate)
}
