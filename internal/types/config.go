// Package types defines all data transfer objects and core types for MuxueTools.
package types

import (
	"fmt"
	"time"
)

// ==================== Configuration Root ====================

// Config represents the complete application configuration.
type Config struct {
	Server        ServerConfig        `mapstructure:"server" yaml:"server"`
	Keys          []KeyConfig         `mapstructure:"keys" yaml:"keys"`
	Pool          PoolConfig          `mapstructure:"pool" yaml:"pool"`
	Models        ModelMappings       `mapstructure:"model_mappings" yaml:"model_mappings"`
	Logging       LoggingConfig       `mapstructure:"logging" yaml:"logging"`
	Update        UpdateConfig        `mapstructure:"update" yaml:"update"`
	Database      DatabaseConfig      `mapstructure:"database" yaml:"database"`
	Advanced      AdvancedConfig      `mapstructure:"advanced" yaml:"advanced"`
	ModelSettings ModelSettingsConfig `mapstructure:"model_settings" yaml:"model_settings"`
}

// ==================== Server Configuration ====================

// ServerConfig contains HTTP server settings.
type ServerConfig struct {
	Port int    `mapstructure:"port" yaml:"port"`
	Host string `mapstructure:"host" yaml:"host"`
}

// DefaultServerConfig returns the default server configuration.
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Port: 8080,
		Host: "0.0.0.0",
	}
}

// Addr returns the full address string (host:port).
func (c *ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// ==================== Pool Configuration ====================

// PoolStrategy defines the key selection strategy.
type PoolStrategy string

const (
	PoolStrategyRoundRobin PoolStrategy = "round_robin"
	PoolStrategyRandom     PoolStrategy = "random"
	PoolStrategyLeastUsed  PoolStrategy = "least_used"
	PoolStrategyWeighted   PoolStrategy = "weighted"
)

// IsValid returns true if the strategy is a valid PoolStrategy value.
func (s PoolStrategy) IsValid() bool {
	switch s {
	case PoolStrategyRoundRobin, PoolStrategyRandom, PoolStrategyLeastUsed, PoolStrategyWeighted:
		return true
	}
	return false
}

// PoolConfig contains key pool settings.
type PoolConfig struct {
	Strategy        PoolStrategy `mapstructure:"strategy" yaml:"strategy"`
	CooldownSeconds int          `mapstructure:"cooldown_seconds" yaml:"cooldown_seconds"`
	MaxRetries      int          `mapstructure:"max_retries" yaml:"max_retries"`
}

// DefaultPoolConfig returns the default pool configuration.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		Strategy:        PoolStrategyRoundRobin,
		CooldownSeconds: 60,
		MaxRetries:      3,
	}
}

// ==================== Model Mappings ====================

// ModelMappings maps OpenAI model names to Gemini model names.
type ModelMappings map[string]string

// DefaultModelMappings returns the default model name mappings.
func DefaultModelMappings() ModelMappings {
	return ModelMappings{
		// OpenAI -> Gemini
		"gpt-4":                "gemini-1.5-pro-latest",
		"gpt-4-turbo":          "gemini-1.5-pro-latest",
		"gpt-4-vision-preview": "gemini-1.5-pro-latest",
		"gpt-4o":               "gemini-1.5-flash-latest",
		"gpt-4o-mini":          "gemini-1.5-flash-8b-latest",
		"gpt-3.5-turbo":        "gemini-1.5-flash-latest",
		// Gemini aliases
		"gemini-pro":       "gemini-1.5-pro-latest",
		"gemini-flash":     "gemini-1.5-flash-latest",
		"gemini-2.0-flash": "gemini-2.0-flash",
		"gemini-2.5-pro":   "gemini-2.5-pro-preview",
	}
}

// MapModel returns the Gemini model name for a given request model.
// If no mapping exists, it returns the original model name (pass-through).
func (m ModelMappings) MapModel(requestModel string) string {
	if geminiModel, ok := m[requestModel]; ok {
		return geminiModel
	}
	return requestModel
}

// ==================== Logging Configuration ====================

// LogLevel defines the logging verbosity level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// IsValid returns true if the level is a valid LogLevel value.
func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelDebug, LogLevelInfo, LogLevelWarn, LogLevelError:
		return true
	}
	return false
}

// LogFormat defines the log output format.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level  LogLevel  `mapstructure:"level" yaml:"level"`
	File   string    `mapstructure:"file" yaml:"file"`
	Format LogFormat `mapstructure:"format" yaml:"format"`
}

// DefaultLoggingConfig returns the default logging configuration.
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		Level:  LogLevelInfo,
		File:   "",
		Format: LogFormatText,
	}
}

// ==================== Database Configuration ====================

// DatabaseConfig contains database settings.
type DatabaseConfig struct {
	Path string `mapstructure:"path" yaml:"path"`
}

// DefaultDatabaseConfig returns the default database configuration.
// Path is left empty to indicate that the platform-specific path should be used.
// Use config.GetEffectiveDatabasePath(Path) to get the actual path.
func DefaultDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Path: "", // Empty means use config.GetEffectiveDatabasePath()
	}
}

// ==================== Update Configuration ====================

// UpdateConfig contains GitHub update detection settings.
type UpdateConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled"`
	CheckInterval string `mapstructure:"check_interval" yaml:"check_interval"`
	GithubRepo    string `mapstructure:"github_repo" yaml:"github_repo"`
}

// DefaultUpdateConfig returns the default update configuration.
func DefaultUpdateConfig() UpdateConfig {
	return UpdateConfig{
		Enabled:       true,
		CheckInterval: "24h",
		GithubRepo:    "muxueliunian/MuxueTools",
	}
}

// CheckIntervalDuration parses the check interval string as a Duration.
func (c *UpdateConfig) CheckIntervalDuration() time.Duration {
	d, err := time.ParseDuration(c.CheckInterval)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

// ==================== Advanced Configuration ====================

// AdvancedConfig contains advanced/optional settings.
type AdvancedConfig struct {
	RequestTimeout      int `mapstructure:"request_timeout" yaml:"request_timeout"`             // Seconds
	StreamFlushInterval int `mapstructure:"stream_flush_interval" yaml:"stream_flush_interval"` // Milliseconds
	StatsRetentionDays  int `mapstructure:"stats_retention_days" yaml:"stats_retention_days"`
}

// DefaultAdvancedConfig returns the default advanced configuration.
func DefaultAdvancedConfig() AdvancedConfig {
	return AdvancedConfig{
		RequestTimeout:      120,
		StreamFlushInterval: 100,
		StatsRetentionDays:  30,
	}
}

// RequestTimeoutDuration returns the request timeout as a Duration.
func (c *AdvancedConfig) RequestTimeoutDuration() time.Duration {
	return time.Duration(c.RequestTimeout) * time.Second
}

// ==================== Model Settings Configuration ====================

// ModelSettingsConfig contains global model generation settings.
type ModelSettingsConfig struct {
	SystemPrompt    string   `mapstructure:"system_prompt" yaml:"system_prompt" json:"system_prompt"`
	Temperature     *float64 `mapstructure:"temperature" yaml:"temperature" json:"temperature,omitempty"`
	MaxOutputTokens *int     `mapstructure:"max_output_tokens" yaml:"max_output_tokens" json:"max_output_tokens,omitempty"`
	TopP            *float64 `mapstructure:"top_p" yaml:"top_p" json:"top_p,omitempty"`
	TopK            *int     `mapstructure:"top_k" yaml:"top_k" json:"top_k,omitempty"`
	ThinkingLevel   *string  `mapstructure:"thinking_level" yaml:"thinking_level" json:"thinking_level,omitempty"`
	MediaResolution *string  `mapstructure:"media_resolution" yaml:"media_resolution" json:"media_resolution,omitempty"`
	StreamOutput    *bool    `mapstructure:"stream_output" yaml:"stream_output" json:"stream_output,omitempty"` // Default: true
}

// DefaultModelSettingsConfig returns the default model settings configuration.
func DefaultModelSettingsConfig() ModelSettingsConfig {
	defaultStreamOutput := true
	return ModelSettingsConfig{
		SystemPrompt: "",
		StreamOutput: &defaultStreamOutput,
		// nil means use Gemini's defaults for other fields
	}
}

// ==================== Admin API DTOs ====================

// ConfigResponse represents the response for GET /api/config.
type ConfigResponse struct {
	Success bool       `json:"success"`
	Data    ConfigData `json:"data"`
}

// ConfigData contains the configuration data returned to clients.
// Excludes sensitive information like API keys.
type ConfigData struct {
	Server        ServerConfig        `json:"server"`
	Pool          PoolConfig          `json:"pool"`
	Logging       LoggingConfig       `json:"logging"`
	Update        UpdateConfig        `json:"update"`
	ModelSettings ModelSettingsConfig `json:"model_settings"`
}

// UpdateConfigRequest represents the request for PUT /api/config.
type UpdateConfigRequest struct {
	Server        *ServerConfigUpdate  `json:"server,omitempty"`
	Pool          *PoolConfigUpdate    `json:"pool,omitempty"`
	Logging       *LoggingConfigUpdate `json:"logging,omitempty"`
	ModelSettings *ModelSettingsConfig `json:"model_settings,omitempty"`
}

// ServerConfigUpdate contains partial server configuration updates.
type ServerConfigUpdate struct {
	Port *int `json:"port,omitempty"`
}

// PoolConfigUpdate contains partial pool configuration updates.
type PoolConfigUpdate struct {
	Strategy        *string `json:"strategy,omitempty"`
	CooldownSeconds *int    `json:"cooldown_seconds,omitempty"`
	MaxRetries      *int    `json:"max_retries,omitempty"`
}

// LoggingConfigUpdate contains partial logging configuration updates.
type LoggingConfigUpdate struct {
	Level *string `json:"level,omitempty"`
}

// UpdateConfigResponse represents the response for PUT /api/config.
type UpdateConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ==================== Update Check DTOs ====================

// UpdateCheckResponse represents the response for GET /api/update/check.
type UpdateCheckResponse struct {
	Success bool            `json:"success"`
	Data    UpdateCheckData `json:"data"`
}

// UpdateCheckData contains version information.
type UpdateCheckData struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	DownloadURL    string `json:"download_url"`
	Changelog      string `json:"changelog"`
	PublishedAt    string `json:"published_at"`
}

// ==================== Default Config ====================

// DefaultConfig returns a configuration with all default values.
func DefaultConfig() Config {
	return Config{
		Server:        DefaultServerConfig(),
		Keys:          []KeyConfig{},
		Pool:          DefaultPoolConfig(),
		Models:        DefaultModelMappings(),
		Logging:       DefaultLoggingConfig(),
		Update:        DefaultUpdateConfig(),
		Database:      DefaultDatabaseConfig(),
		Advanced:      DefaultAdvancedConfig(),
		ModelSettings: DefaultModelSettingsConfig(),
	}
}
