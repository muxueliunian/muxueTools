// Package api provides HTTP API handlers and routing for MuxueTools.
package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"muxueTools/internal/gemini"
	"muxueTools/internal/keypool"
	"muxueTools/internal/storage"
	"muxueTools/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ==================== Server ====================

// Server represents the MuxueTools HTTP server.
type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
	config     *types.Config
	pool       *keypool.Pool
	client     *gemini.Client
	storage    *storage.Storage
	logger     *logrus.Logger
	version    string
	webRoot    string // Path to static web files (for desktop mode)
}

// ServerOption is a functional option for configuring the Server.
type ServerOption func(*Server)

// WithVersion sets the server version.
func WithVersion(version string) ServerOption {
	return func(s *Server) {
		s.version = version
	}
}

// WithLogger sets the server logger.
func WithLogger(logger *logrus.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}

// WithWebRoot sets the path to static web files for serving the frontend.
func WithWebRoot(path string) ServerOption {
	return func(s *Server) {
		s.webRoot = path
	}
}

// NewServer creates a new MuxueTools server.
func NewServer(cfg *types.Config, opts ...ServerOption) (*Server, error) {
	// Create server with defaults
	server := &Server{
		config:  cfg,
		version: "dev",
		logger:  logrus.New(),
	}

	// Apply options
	for _, opt := range opts {
		opt(server)
	}

	// Configure logger
	server.configureLogger()

	// Initialize storage (if configured)
	if cfg.Database.Path != "" {
		st, err := storage.NewStorage(cfg.Database.Path)
		if err != nil {
			server.logger.WithError(err).Warn("Failed to initialize storage, running in memory-only mode")
		} else {
			server.storage = st
			server.logger.WithField("path", cfg.Database.Path).Info("Storage initialized")
		}
	}

	// Initialize key pool
	pool, err := server.initializePool()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize key pool: %w", err)
	}
	server.pool = pool

	// Initialize Gemini client
	clientOpts := []gemini.ClientOption{
		gemini.WithRequestTimeout(time.Duration(cfg.Advanced.RequestTimeout) * time.Second),
	}

	// Add model settings getter if storage is available
	if server.storage != nil {
		clientOpts = append(clientOpts, gemini.WithModelSettings(func() *types.ModelSettingsConfig {
			return server.getModelSettings()
		}))
	}

	server.client = gemini.NewClient(pool, clientOpts...)

	// Create router
	routerConfig := &RouterConfig{
		Config:  cfg,
		Pool:    pool,
		Client:  server.client,
		Storage: server.storage,
		Logger:  server.logger,
		Version: server.version,
		WebRoot: server.webRoot,
	}
	server.engine = NewRouter(routerConfig)

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         cfg.Server.Addr(),
		Handler:      server.engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second, // Longer for streaming
		IdleTimeout:  120 * time.Second,
	}

	return server, nil
}

// configureLogger sets up the logger based on configuration.
func (s *Server) configureLogger() {
	// Set log level
	switch s.config.Logging.Level {
	case types.LogLevelDebug:
		s.logger.SetLevel(logrus.DebugLevel)
	case types.LogLevelInfo:
		s.logger.SetLevel(logrus.InfoLevel)
	case types.LogLevelWarn:
		s.logger.SetLevel(logrus.WarnLevel)
	case types.LogLevelError:
		s.logger.SetLevel(logrus.ErrorLevel)
	default:
		s.logger.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	if s.config.Logging.Format == types.LogFormatJSON {
		s.logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		s.logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set log output (file or stdout)
	if s.config.Logging.File != "" {
		file, err := os.OpenFile(s.config.Logging.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			s.logger.Warnf("Failed to open log file %s, using stdout: %v", s.config.Logging.File, err)
		} else {
			s.logger.SetOutput(file)
		}
	}
}

// initializePool creates and initializes the key pool.
func (s *Server) initializePool() (*keypool.Pool, error) {
	// Get strategy from config
	var strategy keypool.Strategy
	switch s.config.Pool.Strategy {
	case types.PoolStrategyRoundRobin:
		strategy = keypool.NewRoundRobinStrategy()
	case types.PoolStrategyRandom:
		strategy = keypool.NewRandomStrategy()
	case types.PoolStrategyLeastUsed:
		strategy = keypool.NewLeastUsedStrategy()
	case types.PoolStrategyWeighted:
		strategy = keypool.NewWeightedStrategy()
	default:
		strategy = keypool.NewRoundRobinStrategy()
	}

	// Build pool options
	poolOpts := []keypool.PoolOption{
		keypool.WithStrategy(strategy),
		keypool.WithCooldownSeconds(s.config.Pool.CooldownSeconds),
	}

	// Add storage if available
	if s.storage != nil {
		poolOpts = append(poolOpts, keypool.WithStorage(s.storage))
	}

	// Create pool with configuration
	pool := keypool.NewPool(s.config.Keys, poolOpts...)

	// If storage is configured, sync config keys to DB and load from DB
	if s.storage != nil {
		// Sync config keys to storage (first-time setup)
		synced, err := pool.SyncConfigToStorage(s.config.Keys)
		if err != nil {
			s.logger.WithError(err).Warn("Failed to sync config keys to storage")
		} else if synced > 0 {
			s.logger.WithField("synced", synced).Info("Synced config keys to storage")
		}

		// Load all keys from storage
		if err := pool.LoadFromStorage(); err != nil {
			s.logger.WithError(err).Warn("Failed to load keys from storage")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"key_count": pool.Size(),
		"strategy":  s.config.Pool.Strategy,
		"storage":   s.storage != nil,
	}).Info("Key pool initialized")

	return pool, nil
}

// Run starts the HTTP server.
func (s *Server) Run() error {
	s.logger.WithFields(logrus.Fields{
		"addr":    s.config.Server.Addr(),
		"version": s.version,
	}).Info("Starting MuxueTools server")

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Info("Server stopped gracefully")
	return nil
}

// ShutdownWithTimeout shuts down the server with a timeout.
func (s *Server) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}

// ==================== Accessors ====================

// Engine returns the underlying Gin engine.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// Pool returns the key pool.
func (s *Server) Pool() *keypool.Pool {
	return s.pool
}

// Client returns the Gemini client.
func (s *Server) Client() *gemini.Client {
	return s.client
}

// Logger returns the server logger.
func (s *Server) Logger() *logrus.Logger {
	return s.logger
}

// Config returns the server configuration.
func (s *Server) Config() *types.Config {
	return s.config
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.config.Server.Addr()
}

// Storage returns the storage instance.
func (s *Server) Storage() *storage.Storage {
	return s.storage
}

// getModelSettings reads model settings from storage.
func (s *Server) getModelSettings() *types.ModelSettingsConfig {
	if s.storage == nil {
		return nil
	}

	settings := &types.ModelSettingsConfig{}

	// Read System Prompt
	if sp, _ := s.storage.GetConfig("model_settings.system_prompt"); sp != "" {
		settings.SystemPrompt = sp
	}

	// Read Temperature
	if temp, _ := s.storage.GetConfig("model_settings.temperature"); temp != "" {
		if parsed, err := parseFloat64(temp); err == nil {
			settings.Temperature = &parsed
		}
	}

	// Read Max Output Tokens
	if tokens, _ := s.storage.GetConfig("model_settings.max_output_tokens"); tokens != "" {
		if parsed, err := parseInt(tokens); err == nil {
			settings.MaxOutputTokens = &parsed
		}
	}

	// Read Top-P
	if topP, _ := s.storage.GetConfig("model_settings.top_p"); topP != "" {
		if parsed, err := parseFloat64(topP); err == nil {
			settings.TopP = &parsed
		}
	}

	// Read Top-K
	if topK, _ := s.storage.GetConfig("model_settings.top_k"); topK != "" {
		if parsed, err := parseInt(topK); err == nil {
			settings.TopK = &parsed
		}
	}

	// Read Thinking Level
	if level, _ := s.storage.GetConfig("model_settings.thinking_level"); level != "" {
		settings.ThinkingLevel = &level
	}

	// Read Media Resolution
	if resolution, _ := s.storage.GetConfig("model_settings.media_resolution"); resolution != "" {
		settings.MediaResolution = &resolution
	}

	return settings
}

// Helper functions for parsing
func parseFloat64(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func parseInt(s string) (int, error) {
	var i int
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

// Close closes the server and all resources.
func (s *Server) Close() error {
	if s.storage != nil {
		return s.storage.Close()
	}
	return nil
}
