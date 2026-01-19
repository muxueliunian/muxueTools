package config

import (
	"os"
	"path/filepath"
	"testing"

	"mxlnapi/internal/types"
)

// TestLoader_Load_DefaultValues tests that default values are used when no config file exists.
func TestLoader_Load_DefaultValues(t *testing.T) {
	// Create loader without any config file
	loader := NewLoader()
	loader.AddSearchPath(t.TempDir()) // Empty directory

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default values
	defaults := types.DefaultConfig()

	if cfg.Server.Port != defaults.Server.Port {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, defaults.Server.Port)
	}
	if cfg.Server.Host != defaults.Server.Host {
		t.Errorf("Server.Host = %s, want %s", cfg.Server.Host, defaults.Server.Host)
	}
	if cfg.Pool.Strategy != defaults.Pool.Strategy {
		t.Errorf("Pool.Strategy = %s, want %s", cfg.Pool.Strategy, defaults.Pool.Strategy)
	}
	if cfg.Pool.CooldownSeconds != defaults.Pool.CooldownSeconds {
		t.Errorf("Pool.CooldownSeconds = %d, want %d", cfg.Pool.CooldownSeconds, defaults.Pool.CooldownSeconds)
	}
	if cfg.Logging.Level != defaults.Logging.Level {
		t.Errorf("Logging.Level = %s, want %s", cfg.Logging.Level, defaults.Logging.Level)
	}
}

// TestLoader_LoadFromFile tests loading configuration from a YAML file.
func TestLoader_LoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 9090
  host: "127.0.0.1"

pool:
  strategy: "random"
  cooldown_seconds: 120
  max_retries: 5

keys:
  - key: "AIzaSyTestKey123456789012345678901234"
    name: "Test Key"
    enabled: true
    tags:
      - "test"

logging:
  level: "debug"
  format: "json"

database:
  path: "custom/path.db"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	// Verify loaded values
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want 9090", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Server.Host = %s, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Pool.Strategy != types.PoolStrategyRandom {
		t.Errorf("Pool.Strategy = %s, want random", cfg.Pool.Strategy)
	}
	if cfg.Pool.CooldownSeconds != 120 {
		t.Errorf("Pool.CooldownSeconds = %d, want 120", cfg.Pool.CooldownSeconds)
	}
	if cfg.Pool.MaxRetries != 5 {
		t.Errorf("Pool.MaxRetries = %d, want 5", cfg.Pool.MaxRetries)
	}
	if len(cfg.Keys) != 1 {
		t.Fatalf("len(Keys) = %d, want 1", len(cfg.Keys))
	}
	if cfg.Keys[0].Name != "Test Key" {
		t.Errorf("Keys[0].Name = %s, want Test Key", cfg.Keys[0].Name)
	}
	if cfg.Logging.Level != types.LogLevelDebug {
		t.Errorf("Logging.Level = %s, want debug", cfg.Logging.Level)
	}
	if cfg.Database.Path != "custom/path.db" {
		t.Errorf("Database.Path = %s, want custom/path.db", cfg.Database.Path)
	}
}

// TestLoader_EnvironmentVariables tests that environment variables override config values.
func TestLoader_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("MXLN_SERVER_PORT", "3000")
	os.Setenv("MXLN_LOGGING_LEVEL", "warn")
	defer func() {
		os.Unsetenv("MXLN_SERVER_PORT")
		os.Unsetenv("MXLN_LOGGING_LEVEL")
	}()

	loader := NewLoader()
	loader.AddSearchPath(t.TempDir())

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("Server.Port = %d, want 3000 (from env)", cfg.Server.Port)
	}
	if cfg.Logging.Level != types.LogLevelWarn {
		t.Errorf("Logging.Level = %s, want warn (from env)", cfg.Logging.Level)
	}
}

// TestValidate_ValidConfig tests that a valid configuration passes validation.
func TestValidate_ValidConfig(t *testing.T) {
	cfg := types.DefaultConfig()
	if err := Validate(&cfg); err != nil {
		t.Errorf("Validate() failed for valid config: %v", err)
	}
}

// TestValidate_InvalidPort tests that invalid port values are rejected.
func TestValidate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port zero", 0},
		{"port negative", -1},
		{"port too high", 70000},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := types.DefaultConfig()
			cfg.Server.Port = tc.port
			if err := Validate(&cfg); err == nil {
				t.Errorf("Validate() should fail for port %d", tc.port)
			}
		})
	}
}

// TestValidate_InvalidStrategy tests that invalid pool strategy is rejected.
func TestValidate_InvalidStrategy(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Pool.Strategy = "invalid_strategy"

	if err := Validate(&cfg); err == nil {
		t.Error("Validate() should fail for invalid strategy")
	}
}

// TestValidate_InvalidLogLevel tests that invalid log level is rejected.
func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Logging.Level = "invalid"

	if err := Validate(&cfg); err == nil {
		t.Error("Validate() should fail for invalid log level")
	}
}

// TestValidate_EmptyKeyInList tests that empty key in keys list is rejected.
func TestValidate_EmptyKeyInList(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Keys = []types.KeyConfig{
		{Key: "", Name: "Empty Key", Enabled: true},
	}

	if err := Validate(&cfg); err == nil {
		t.Error("Validate() should fail for empty key")
	}
}

// TestValidate_InvalidMaxRetries tests that invalid max retries is rejected.
func TestValidate_InvalidMaxRetries(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Pool.MaxRetries = 0

	if err := Validate(&cfg); err == nil {
		t.Error("Validate() should fail for max_retries = 0")
	}
}

// TestValidate_InvalidRequestTimeout tests that invalid request timeout is rejected.
func TestValidate_InvalidRequestTimeout(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Advanced.RequestTimeout = 0

	if err := Validate(&cfg); err == nil {
		t.Error("Validate() should fail for request_timeout = 0")
	}
}

// TestGlobalConfig tests the global configuration access methods.
func TestGlobalConfig(t *testing.T) {
	// Clean up
	defer Reset()

	// Test GetSafe when not initialized
	if cfg := GetSafe(); cfg != nil {
		t.Error("GetSafe() should return nil when not initialized")
	}

	// Test Get panic when not initialized
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get() should panic when not initialized")
		}
	}()
	Get()
}

// TestGlobalConfig_AfterSet tests global config after Set is called.
func TestGlobalConfig_AfterSet(t *testing.T) {
	defer Reset()

	cfg := types.DefaultConfig()
	cfg.Server.Port = 5000

	Set(&cfg)

	got := Get()
	if got.Server.Port != 5000 {
		t.Errorf("Get().Server.Port = %d, want 5000", got.Server.Port)
	}

	gotSafe := GetSafe()
	if gotSafe == nil {
		t.Error("GetSafe() should return config after Set()")
	}
}

// TestInit tests the Init function with temporary directory.
func TestInit(t *testing.T) {
	defer Reset()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
server:
  port: 7777
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	if err := Init(tmpDir); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	cfg := Get()
	if cfg.Server.Port != 7777 {
		t.Errorf("Get().Server.Port = %d, want 7777", cfg.Server.Port)
	}
}

// TestGenerateDefaultConfig tests generating a new config file.
func TestGenerateDefaultConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	if err := GenerateDefaultConfig(configPath); err != nil {
		t.Fatalf("GenerateDefaultConfig() failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}

	// Verify it can be loaded
	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load generated config: %v", err)
	}

	// Should have valid defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("Generated config Server.Port = %d, want 8080", cfg.Server.Port)
	}
}

// TestGenerateDefaultConfig_AlreadyExists tests that it fails when file exists.
func TestGenerateDefaultConfig_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create existing file
	if err := os.WriteFile(configPath, []byte("existing"), 0644); err != nil {
		t.Fatalf("Failed to write existing file: %v", err)
	}

	if err := GenerateDefaultConfig(configPath); err == nil {
		t.Error("GenerateDefaultConfig() should fail when file exists")
	}
}

// TestLoader_MalformedYAML tests handling of malformed YAML files.
func TestLoader_MalformedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Write malformed YAML
	malformedContent := `
server:
  port: "not a number"  # Should be int
  : invalid
`
	if err := os.WriteFile(configPath, []byte(malformedContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	loader := NewLoader()
	_, err := loader.LoadFromFile(configPath)
	if err == nil {
		t.Error("LoadFromFile() should fail for malformed YAML")
	}
}

// TestLoader_PartialConfig tests loading a config with only some values set.
func TestLoader_PartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Only set port, other values should use defaults
	partialContent := `
server:
  port: 4000
`
	if err := os.WriteFile(configPath, []byte(partialContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	loader := NewLoader()
	cfg, err := loader.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile() failed: %v", err)
	}

	defaults := types.DefaultConfig()

	// Custom value
	if cfg.Server.Port != 4000 {
		t.Errorf("Server.Port = %d, want 4000", cfg.Server.Port)
	}

	// Default values
	if cfg.Server.Host != defaults.Server.Host {
		t.Errorf("Server.Host = %s, want %s (default)", cfg.Server.Host, defaults.Server.Host)
	}
	if cfg.Pool.Strategy != defaults.Pool.Strategy {
		t.Errorf("Pool.Strategy = %s, want %s (default)", cfg.Pool.Strategy, defaults.Pool.Strategy)
	}
}
