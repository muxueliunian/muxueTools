// Package config provides cross-platform path utilities for MuxueTools.
// It follows XDG Base Directory Specification on Linux and uses
// APPDATA on Windows for proper system integration.
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// AppName is the application name used for directory paths.
// Windows uses PascalCase, Linux uses lowercase.
const AppName = "MuxueTools"

// appNameLower is the lowercase app name for Linux paths.
const appNameLower = "muxuetools"

// GetDataDir returns the user data directory.
// This is automatically scanned and should not be customized by users.
//
// Paths:
//   - Windows: %APPDATA%\MuxueTools\data\
//   - Linux:   ~/.local/share/muxuetools/
//   - Other:   ./data (fallback)
func GetDataDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, AppName, "data")
	case "linux":
		// Follow XDG Base Directory Specification
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return filepath.Join(xdgData, appNameLower)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", appNameLower)
	default:
		return "data"
	}
}

// GetConfigDir returns the configuration directory.
// This is where config.yaml should be placed.
//
// Paths:
//   - Windows: %APPDATA%\MuxueTools\
//   - Linux:   ~/.config/muxuetools/
//   - Other:   . (current directory)
func GetConfigDir() string {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, _ := os.UserHomeDir()
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, AppName)
	case "linux":
		// Follow XDG Base Directory Specification
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return filepath.Join(xdgConfig, appNameLower)
		}
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".config", appNameLower)
	default:
		return "."
	}
}

// GetLogDir returns the log directory.
//
// Paths:
//   - Windows: %APPDATA%\MuxueTools\data\logs\
//   - Linux:   ~/.local/share/muxuetools/logs/
//   - Other:   ./data/logs
func GetLogDir() string {
	return filepath.Join(GetDataDir(), "logs")
}

// GetDatabasePath returns the SQLite database file path.
//
// Paths:
//   - Windows: %APPDATA%\MuxueTools\data\muxuetools.db
//   - Linux:   ~/.local/share/muxuetools/muxuetools.db
//   - Other:   ./data/muxuetools.db
func GetDatabasePath() string {
	return filepath.Join(GetDataDir(), "muxuetools.db")
}

// GetConfigFilePath returns the full path to the config file.
//
// Paths:
//   - Windows: %APPDATA%\MuxueTools\config.yaml
//   - Linux:   ~/.config/muxuetools/config.yaml
//   - Other:   ./config.yaml
func GetConfigFilePath() string {
	return filepath.Join(GetConfigDir(), "config.yaml")
}

// EnsureDirectories creates all required directories if they don't exist.
// Returns an error if any directory creation fails.
func EnsureDirectories() error {
	dirs := []string{
		GetConfigDir(),
		GetDataDir(),
		GetLogDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// IsPortableMode checks if the application is running in portable mode.
// Portable mode is active when a config.yaml file exists in the current directory.
func IsPortableMode() bool {
	if _, err := os.Stat("config.yaml"); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join("configs", "config.yaml")); err == nil {
		return true
	}
	return false
}

// GetEffectiveDataDir returns the data directory respecting portable mode.
// In portable mode, it returns "data" (relative path).
// In standard mode, it returns the platform-specific path.
func GetEffectiveDataDir() string {
	if IsPortableMode() {
		return "data"
	}
	return GetDataDir()
}

// GetEffectiveDatabasePath returns the database path respecting portable mode.
// If the config has an explicit path set, it returns that path.
// Otherwise, it returns the platform-specific default path.
func GetEffectiveDatabasePath(configuredPath string) string {
	if configuredPath != "" {
		return configuredPath
	}
	if IsPortableMode() {
		return filepath.Join("data", "muxuetools.db")
	}
	return GetDatabasePath()
}
