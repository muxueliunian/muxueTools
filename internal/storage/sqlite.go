// Package storage provides SQLite-based persistence layer for MuxueTools.
package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"muxueTools/internal/types"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Storage provides database access methods.
type Storage struct {
	db *gorm.DB
}

// NewStorage creates a new Storage instance with SQLite backend.
func NewStorage(dbPath string) (*Storage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	// Open database connection
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxOpenConns(1) // SQLite only supports one writer at a time
	sqlDB.SetMaxIdleConns(1)

	storage := &Storage{db: db}

	// Run migrations
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return storage, nil
}

// NewStorageWithDB creates a Storage with an existing GORM DB instance.
// Useful for testing with in-memory database.
func NewStorageWithDB(db *gorm.DB) (*Storage, error) {
	storage := &Storage{db: db}

	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return storage, nil
}

// migrate runs database migrations.
func (s *Storage) migrate() error {
	return s.db.AutoMigrate(
		&DBKey{},
		&types.Session{},
		&types.ChatMessage{},
		&DBConfig{}, // 新增配置表
	)
}

// Close closes the database connection.
func (s *Storage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Ping checks if the database connection is alive.
func (s *Storage) Ping() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// DB returns the underlying GORM DB instance.
// Use with caution - prefer using Storage methods.
func (s *Storage) DB() *gorm.DB {
	return s.db
}

// ==================== Database Models ====================

// DBKey is the database model for API keys.
// Separate from types.Key to handle GORM-specific concerns.
type DBKey struct {
	ID               string `gorm:"primaryKey;type:varchar(36)"`
	APIKey           string `gorm:"type:text;not null;uniqueIndex"`
	Name             string `gorm:"type:varchar(100)"`
	Tags             string `gorm:"type:text"` // JSON array
	Enabled          bool   `gorm:"default:true"`
	RequestCount     int64  `gorm:"default:0"`
	SuccessCount     int64  `gorm:"default:0"`
	ErrorCount       int64  `gorm:"default:0"`
	PromptTokens     int64  `gorm:"default:0"`
	CompletionTokens int64  `gorm:"default:0"`
	ModelUsage       string `gorm:"type:text"`    // JSON map[string]int64
	LastUsedAt       *int64 `gorm:"type:integer"` // Unix timestamp
	CreatedAt        int64  `gorm:"autoCreateTime"`
	UpdatedAt        int64  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for DBKey.
func (DBKey) TableName() string {
	return "keys"
}

// DBConfig is the database model for application configuration.
// Uses Key-Value pattern for flexibility.
type DBConfig struct {
	Key       string `gorm:"primaryKey;type:varchar(100)"`
	Value     string `gorm:"type:text;not null"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for DBConfig.
func (DBConfig) TableName() string {
	return "app_config"
}

// ==================== Configuration Methods ====================

// GetConfig retrieves a configuration value by key.
// Returns empty string if not found.
func (s *Storage) GetConfig(key string) (string, error) {
	var cfg DBConfig
	result := s.db.Where("key = ?", key).First(&cfg)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", nil // Not found, return empty
		}
		return "", fmt.Errorf("failed to get config %s: %w", key, result.Error)
	}
	return cfg.Value, nil
}

// SetConfig saves or updates a configuration value.
func (s *Storage) SetConfig(key, value string) error {
	cfg := DBConfig{
		Key:   key,
		Value: value,
	}
	// Upsert: update if exists, insert if not
	result := s.db.Save(&cfg)
	if result.Error != nil {
		return fmt.Errorf("failed to set config %s: %w", key, result.Error)
	}
	return nil
}

// GetAllConfig retrieves all configuration values as a map.
func (s *Storage) GetAllConfig() (map[string]string, error) {
	var configs []DBConfig
	result := s.db.Find(&configs)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all config: %w", result.Error)
	}

	configMap := make(map[string]string, len(configs))
	for _, cfg := range configs {
		configMap[cfg.Key] = cfg.Value
	}
	return configMap, nil
}

// DeleteConfig removes a configuration value by key.
func (s *Storage) DeleteConfig(key string) error {
	result := s.db.Where("key = ?", key).Delete(&DBConfig{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete config %s: %w", key, result.Error)
	}
	return nil
}

// ==================== Data Cleanup Methods ====================

// DeleteAllSessions deletes all sessions and their messages.
// Returns the number of sessions deleted.
func (s *Storage) DeleteAllSessions() (int64, error) {
	// Delete all messages first (foreign key constraint)
	result := s.db.Delete(&types.ChatMessage{}, "1=1")
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete messages: %w", result.Error)
	}

	// Delete all sessions
	result = s.db.Delete(&types.Session{}, "1=1")
	if result.Error != nil {
		return 0, fmt.Errorf("failed to delete sessions: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// ResetKeyStats resets statistics for all keys (request count, success count, etc.)
// Returns the number of keys affected.
func (s *Storage) ResetKeyStats() (int64, error) {
	result := s.db.Model(&DBKey{}).Updates(map[string]interface{}{
		"request_count":     0,
		"success_count":     0,
		"error_count":       0,
		"prompt_tokens":     0,
		"completion_tokens": 0,
		"model_usage":       "",
		"last_used_at":      nil,
	})
	if result.Error != nil {
		return 0, fmt.Errorf("failed to reset key stats: %w", result.Error)
	}
	return result.RowsAffected, nil
}
