// Package config provides legacy data migration detection for MuxueTools.
package config

import (
	"os"

	"github.com/sirupsen/logrus"
)

// legacyPaths defines old relative paths used by previous versions.
var legacyPaths = []string{
	"data/MuxueTools.db",
	"data/muxuetools.db",
}

// CheckLegacyData checks if legacy data exists from previous versions.
// Returns the legacy path if found, empty string otherwise.
//
// Migration is only suggested if:
//  1. New data path doesn't exist yet (clean install to new location)
//  2. Legacy data exists in the old relative path
func CheckLegacyData() string {
	newDBPath := GetDatabasePath()

	// If data already exists at new path, no migration needed
	if fileExists(newDBPath) {
		return ""
	}

	// Check each legacy path for existing data
	for _, oldPath := range legacyPaths {
		if fileExists(oldPath) {
			return oldPath
		}
	}

	return ""
}

// LogMigrationHint logs a warning message about legacy data and migration.
// This is informational only; automatic migration is not performed.
func LogMigrationHint(logger *logrus.Logger, oldPath string) {
	newPath := GetDatabasePath()
	logger.Warn("========================================")
	logger.Warnf("发现旧版本数据: %s", oldPath)
	logger.Warnf("新数据路径: %s", newPath)
	logger.Warn("请手动迁移数据或继续使用旧路径")
	logger.Warn("========================================")
}

// fileExists checks if a path exists and is a file (not a directory).
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
