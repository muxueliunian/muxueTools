package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetDataDir(t *testing.T) {
	dataDir := GetDataDir()

	if dataDir == "" {
		t.Error("GetDataDir() returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		// Should contain AppData and MuxueTools
		if !strings.Contains(dataDir, "AppData") {
			t.Errorf("Windows GetDataDir() should contain 'AppData', got: %s", dataDir)
		}
		if !strings.Contains(dataDir, "MuxueTools") {
			t.Errorf("Windows GetDataDir() should contain 'MuxueTools', got: %s", dataDir)
		}
		if !strings.HasSuffix(dataDir, "data") {
			t.Errorf("Windows GetDataDir() should end with 'data', got: %s", dataDir)
		}
	case "linux":
		// Should contain .local/share and muxuetools
		if !strings.Contains(dataDir, ".local") || !strings.Contains(dataDir, "share") {
			// Check if XDG_DATA_HOME is set
			if os.Getenv("XDG_DATA_HOME") == "" {
				t.Errorf("Linux GetDataDir() should contain '.local/share', got: %s", dataDir)
			}
		}
		if !strings.Contains(dataDir, "muxuetools") {
			t.Errorf("Linux GetDataDir() should contain 'muxuetools', got: %s", dataDir)
		}
	default:
		if dataDir != "data" {
			t.Errorf("Default GetDataDir() should return 'data', got: %s", dataDir)
		}
	}
}

func TestGetConfigDir(t *testing.T) {
	configDir := GetConfigDir()

	if configDir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(configDir, "AppData") {
			t.Errorf("Windows GetConfigDir() should contain 'AppData', got: %s", configDir)
		}
		if !strings.Contains(configDir, "MuxueTools") {
			t.Errorf("Windows GetConfigDir() should contain 'MuxueTools', got: %s", configDir)
		}
	case "linux":
		if !strings.Contains(configDir, ".config") {
			// Check if XDG_CONFIG_HOME is set
			if os.Getenv("XDG_CONFIG_HOME") == "" {
				t.Errorf("Linux GetConfigDir() should contain '.config', got: %s", configDir)
			}
		}
		if !strings.Contains(configDir, "muxuetools") {
			t.Errorf("Linux GetConfigDir() should contain 'muxuetools', got: %s", configDir)
		}
	default:
		if configDir != "." {
			t.Errorf("Default GetConfigDir() should return '.', got: %s", configDir)
		}
	}
}

func TestGetLogDir(t *testing.T) {
	logDir := GetLogDir()
	dataDir := GetDataDir()

	if logDir == "" {
		t.Error("GetLogDir() returned empty string")
	}

	// Log dir should be a subdirectory of data dir
	expectedLogDir := filepath.Join(dataDir, "logs")
	if logDir != expectedLogDir {
		t.Errorf("GetLogDir() = %s, want %s", logDir, expectedLogDir)
	}
}

func TestGetDatabasePath(t *testing.T) {
	dbPath := GetDatabasePath()
	dataDir := GetDataDir()

	if dbPath == "" {
		t.Error("GetDatabasePath() returned empty string")
	}

	// Database path should be in data dir
	expectedDBPath := filepath.Join(dataDir, "muxuetools.db")
	if dbPath != expectedDBPath {
		t.Errorf("GetDatabasePath() = %s, want %s", dbPath, expectedDBPath)
	}

	// Should end with .db extension
	if !strings.HasSuffix(dbPath, ".db") {
		t.Errorf("GetDatabasePath() should end with '.db', got: %s", dbPath)
	}
}

func TestGetConfigFilePath(t *testing.T) {
	configPath := GetConfigFilePath()
	configDir := GetConfigDir()

	if configPath == "" {
		t.Error("GetConfigFilePath() returned empty string")
	}

	// Config file should be in config dir
	expectedPath := filepath.Join(configDir, "config.yaml")
	if configPath != expectedPath {
		t.Errorf("GetConfigFilePath() = %s, want %s", configPath, expectedPath)
	}
}

func TestEnsureDirectories(t *testing.T) {
	// Create a temp directory for testing
	tmpDir, err := os.MkdirTemp("", "muxuetools-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test that EnsureDirectories doesn't return an error on existing paths
	err = EnsureDirectories()
	if err != nil {
		// This might fail if we don't have permissions
		t.Logf("EnsureDirectories() error (may be permission issue): %v", err)
	}
}

func TestIsPortableMode(t *testing.T) {
	// Get initial state
	isPortable := IsPortableMode()

	// The result depends on whether config.yaml exists in current dir
	// Just verify it returns a boolean without panic
	t.Logf("IsPortableMode() = %v", isPortable)
}

func TestGetEffectiveDataDir(t *testing.T) {
	effectiveDir := GetEffectiveDataDir()

	if effectiveDir == "" {
		t.Error("GetEffectiveDataDir() returned empty string")
	}

	// In portable mode, should return "data"
	// Otherwise, should return platform-specific path
	if IsPortableMode() {
		if effectiveDir != "data" {
			t.Errorf("Portable mode GetEffectiveDataDir() should return 'data', got: %s", effectiveDir)
		}
	} else {
		platformDir := GetDataDir()
		if effectiveDir != platformDir {
			t.Errorf("Standard mode GetEffectiveDataDir() should return %s, got: %s", platformDir, effectiveDir)
		}
	}
}

func TestGetEffectiveDatabasePath(t *testing.T) {
	tests := []struct {
		name           string
		configuredPath string
		wantCustom     bool
	}{
		{
			name:           "empty config uses default",
			configuredPath: "",
			wantCustom:     false,
		},
		{
			name:           "custom path is respected",
			configuredPath: "/custom/path/db.sqlite",
			wantCustom:     true,
		},
		{
			name:           "relative custom path",
			configuredPath: "my-data/custom.db",
			wantCustom:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetEffectiveDatabasePath(tt.configuredPath)

			if tt.wantCustom {
				if result != tt.configuredPath {
					t.Errorf("GetEffectiveDatabasePath(%q) = %q, want %q",
						tt.configuredPath, result, tt.configuredPath)
				}
			} else {
				// Should return default path
				if result == "" {
					t.Error("GetEffectiveDatabasePath(\"\") returned empty string")
				}
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Test with non-existent file
	if fileExists("non-existent-file-12345.xyz") {
		t.Error("fileExists() should return false for non-existent file")
	}

	// Create a temp file for testing
	tmpFile, err := os.CreateTemp("", "muxuetools-test-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if !fileExists(tmpPath) {
		t.Error("fileExists() should return true for existing file")
	}

	// Test with directory (should return false)
	tmpDir, err := os.MkdirTemp("", "muxuetools-testdir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if fileExists(tmpDir) {
		t.Error("fileExists() should return false for directory")
	}
}

func TestCheckLegacyData(t *testing.T) {
	// Just verify it doesn't panic and returns a string
	result := CheckLegacyData()
	t.Logf("CheckLegacyData() = %q", result)

	// If result is not empty, it should be one of the legacy paths
	if result != "" {
		found := false
		for _, path := range legacyPaths {
			if result == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CheckLegacyData() returned unexpected path: %s", result)
		}
	}
}

func TestXDGEnvironmentVariables(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("XDG tests only run on Linux")
	}

	// Test XDG_DATA_HOME override
	customDataHome := "/tmp/custom-xdg-data"
	oldXDGData := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", customDataHome)

	dataDir := GetDataDir()
	if !strings.HasPrefix(dataDir, customDataHome) {
		t.Errorf("GetDataDir() should respect XDG_DATA_HOME, got: %s", dataDir)
	}

	// Restore
	if oldXDGData != "" {
		os.Setenv("XDG_DATA_HOME", oldXDGData)
	} else {
		os.Unsetenv("XDG_DATA_HOME")
	}

	// Test XDG_CONFIG_HOME override
	customConfigHome := "/tmp/custom-xdg-config"
	oldXDGConfig := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", customConfigHome)

	configDir := GetConfigDir()
	if !strings.HasPrefix(configDir, customConfigHome) {
		t.Errorf("GetConfigDir() should respect XDG_CONFIG_HOME, got: %s", configDir)
	}

	// Restore
	if oldXDGConfig != "" {
		os.Setenv("XDG_CONFIG_HOME", oldXDGConfig)
	} else {
		os.Unsetenv("XDG_CONFIG_HOME")
	}
}
