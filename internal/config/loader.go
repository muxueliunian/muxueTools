// Package config provides configuration loading and management for MuxueTools.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"muxueTools/internal/types"

	"github.com/spf13/viper"
)

// Global configuration instance with thread-safe access.
var (
	globalConfig *types.Config
	configMu     sync.RWMutex
	once         sync.Once
	appVersion   = "dev" // Set at build time via -ldflags
)

// SetVersion sets the application version (called at startup).
func SetVersion(version string) {
	if version != "" {
		appVersion = version
	}
}

// GetVersion returns the current application version.
func GetVersion() string {
	return appVersion
}

// Loader handles configuration loading and management.
type Loader struct {
	v        *viper.Viper
	filePath string
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	return &Loader{v: v}
}

// SetConfigPath sets the path to the configuration file.
func (l *Loader) SetConfigPath(path string) {
	l.filePath = path
	if path != "" {
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		name := strings.TrimSuffix(base, ext)

		l.v.SetConfigName(name)
		l.v.AddConfigPath(dir)
		if ext != "" {
			l.v.SetConfigType(strings.TrimPrefix(ext, "."))
		}
	}
}

// AddSearchPath adds a directory to search for config files.
func (l *Loader) AddSearchPath(path string) {
	l.v.AddConfigPath(path)
}

// setupDefaults sets all default configuration values.
func (l *Loader) setupDefaults() {
	defaults := types.DefaultConfig()

	// Server defaults
	l.v.SetDefault("server.port", defaults.Server.Port)
	l.v.SetDefault("server.host", defaults.Server.Host)

	// Pool defaults
	l.v.SetDefault("pool.strategy", string(defaults.Pool.Strategy))
	l.v.SetDefault("pool.cooldown_seconds", defaults.Pool.CooldownSeconds)
	l.v.SetDefault("pool.max_retries", defaults.Pool.MaxRetries)

	// Logging defaults
	l.v.SetDefault("logging.level", string(defaults.Logging.Level))
	l.v.SetDefault("logging.file", defaults.Logging.File)
	l.v.SetDefault("logging.format", string(defaults.Logging.Format))

	// Database defaults
	l.v.SetDefault("database.path", defaults.Database.Path)

	// Update defaults
	l.v.SetDefault("update.enabled", defaults.Update.Enabled)
	l.v.SetDefault("update.check_interval", defaults.Update.CheckInterval)
	l.v.SetDefault("update.github_repo", defaults.Update.GithubRepo)

	// Advanced defaults
	l.v.SetDefault("advanced.request_timeout", defaults.Advanced.RequestTimeout)
	l.v.SetDefault("advanced.stream_flush_interval", defaults.Advanced.StreamFlushInterval)
	l.v.SetDefault("advanced.stats_retention_days", defaults.Advanced.StatsRetentionDays)
}

// setupEnvBindings configures environment variable bindings.
// Environment variables use the prefix MXLN_ and replace dots with underscores.
// Example: MXLN_SERVER_PORT overrides server.port
func (l *Loader) setupEnvBindings() {
	l.v.SetEnvPrefix("MXLN")
	l.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	l.v.AutomaticEnv()
}

// Load reads the configuration from file and environment variables.
// If the config file doesn't exist, it uses default values.
func (l *Loader) Load() (*types.Config, error) {
	l.setupDefaults()
	l.setupEnvBindings()

	// Add default search paths (priority from high to low)
	// 1. Current directory (portable mode)
	// 2. ./configs (development mode)
	// 3. User config directory (platform-specific)
	l.v.AddConfigPath(".")
	l.v.AddConfigPath("./configs")
	l.v.AddConfigPath(GetConfigDir())

	// Try to read config file (not an error if it doesn't exist)
	if err := l.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error occurred
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, will use defaults
	}

	// Unmarshal into config struct
	cfg := &types.Config{}
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply default model mappings if not set
	if len(cfg.Models) == 0 {
		cfg.Models = types.DefaultModelMappings()
	}

	// Validate configuration
	if err := Validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a specific file path.
func (l *Loader) LoadFromFile(path string) (*types.Config, error) {
	l.SetConfigPath(path)
	return l.Load()
}

// Validate checks if the configuration is valid.
func Validate(cfg *types.Config) error {
	// Validate server config
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	// Validate pool config
	if !cfg.Pool.Strategy.IsValid() {
		return fmt.Errorf("pool.strategy is invalid: %s", cfg.Pool.Strategy)
	}
	if cfg.Pool.CooldownSeconds < 0 {
		return fmt.Errorf("pool.cooldown_seconds must be >= 0, got %d", cfg.Pool.CooldownSeconds)
	}
	if cfg.Pool.MaxRetries < 1 {
		return fmt.Errorf("pool.max_retries must be >= 1, got %d", cfg.Pool.MaxRetries)
	}

	// Validate logging config
	if !cfg.Logging.Level.IsValid() {
		return fmt.Errorf("logging.level is invalid: %s", cfg.Logging.Level)
	}

	// Validate keys (if any)
	for i, key := range cfg.Keys {
		if key.Key == "" {
			return fmt.Errorf("keys[%d].key cannot be empty", i)
		}
	}

	// Validate advanced config
	if cfg.Advanced.RequestTimeout < 1 {
		return fmt.Errorf("advanced.request_timeout must be >= 1, got %d", cfg.Advanced.RequestTimeout)
	}

	return nil
}

// ==================== Global Config Access ====================

// Get returns the global configuration.
// Panics if configuration has not been initialized.
func Get() *types.Config {
	configMu.RLock()
	defer configMu.RUnlock()
	if globalConfig == nil {
		panic("config: configuration not initialized, call Init() first")
	}
	return globalConfig
}

// GetSafe returns the global configuration or nil if not initialized.
func GetSafe() *types.Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

// Set sets the global configuration.
func Set(cfg *types.Config) {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = cfg
}

// Init initializes the global configuration.
// This is typically called once at application startup.
func Init(paths ...string) error {
	loader := NewLoader()

	// Add custom search paths
	for _, p := range paths {
		loader.AddSearchPath(p)
	}

	cfg, err := loader.Load()
	if err != nil {
		return err
	}

	Set(cfg)
	return nil
}

// InitFromFile initializes the global configuration from a specific file.
func InitFromFile(path string) error {
	loader := NewLoader()
	cfg, err := loader.LoadFromFile(path)
	if err != nil {
		return err
	}
	Set(cfg)
	return nil
}

// Reset clears the global configuration (primarily for testing).
func Reset() {
	configMu.Lock()
	defer configMu.Unlock()
	globalConfig = nil
}

// ==================== Config File Generation ====================

// GenerateDefaultConfig creates a new config file with default values at the specified path.
func GenerateDefaultConfig(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}

	// Read the example config from embedded or default location
	examplePath := filepath.Join("configs", "config.example.yaml")
	content, err := os.ReadFile(examplePath)
	if err != nil {
		// If example doesn't exist, generate minimal config
		content = []byte(minimalConfigTemplate)
	}

	// Write the config file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// minimalConfigTemplate is used when the example config is not available.
const minimalConfigTemplate = `# MuxueTools Configuration
server:
  port: 8080
  host: "0.0.0.0"

keys: []

pool:
  strategy: "round_robin"
  cooldown_seconds: 60
  max_retries: 3

logging:
  level: "info"

database:
  path: "data/MuxueTools.db"

update:
  enabled: true
  github_repo: "muxueliunian/MuxueTools"
`
