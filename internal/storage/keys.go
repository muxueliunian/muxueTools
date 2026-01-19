// Package storage provides SQLite-based persistence layer for MuxueTools.
package storage

import (
	"encoding/json"
	"fmt"
	"time"

	"muxueTools/internal/types"

	"gorm.io/gorm"
)

// ==================== Key Storage Methods ====================

// CreateKey creates a new key in the database.
func (s *Storage) CreateKey(key *types.Key) error {
	dbKey := keyToDBKey(key)
	if err := s.db.Create(&dbKey).Error; err != nil {
		return fmt.Errorf("failed to create key: %w", err)
	}
	return nil
}

// GetKey retrieves a key by ID.
func (s *Storage) GetKey(id string) (*types.Key, error) {
	var dbKey DBKey
	if err := s.db.Where("id = ?", id).First(&dbKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key: %w", err)
	}
	return dbKeyToKey(&dbKey), nil
}

// GetKeyByAPIKey retrieves a key by its API key value.
func (s *Storage) GetKeyByAPIKey(apiKey string) (*types.Key, error) {
	var dbKey DBKey
	if err := s.db.Where("api_key = ?", apiKey).First(&dbKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get key by api_key: %w", err)
	}
	return dbKeyToKey(&dbKey), nil
}

// ListKeys retrieves all keys from the database.
func (s *Storage) ListKeys() ([]types.Key, error) {
	var dbKeys []DBKey
	if err := s.db.Order("created_at DESC").Find(&dbKeys).Error; err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}

	keys := make([]types.Key, 0, len(dbKeys))
	for _, dbKey := range dbKeys {
		keys = append(keys, *dbKeyToKey(&dbKey))
	}
	return keys, nil
}

// UpdateKey updates an existing key in the database.
func (s *Storage) UpdateKey(key *types.Key) error {
	dbKey := keyToDBKey(key)
	result := s.db.Model(&DBKey{}).Where("id = ?", key.ID).Updates(map[string]interface{}{
		"name":              dbKey.Name,
		"tags":              dbKey.Tags,
		"enabled":           dbKey.Enabled,
		"request_count":     dbKey.RequestCount,
		"success_count":     dbKey.SuccessCount,
		"error_count":       dbKey.ErrorCount,
		"prompt_tokens":     dbKey.PromptTokens,
		"completion_tokens": dbKey.CompletionTokens,
		"model_usage":       dbKey.ModelUsage,
		"last_used_at":      dbKey.LastUsedAt,
		"updated_at":        time.Now().Unix(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.ErrKeyNotFound
	}
	return nil
}

// DeleteKey deletes a key by ID.
func (s *Storage) DeleteKey(id string) error {
	result := s.db.Where("id = ?", id).Delete(&DBKey{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete key: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.ErrKeyNotFound
	}
	return nil
}

// ImportKeys imports multiple keys, skipping duplicates.
// Returns the number of successfully imported keys.
// Uses a transaction to ensure atomicity.
func (s *Storage) ImportKeys(keys []types.Key) (int, error) {
	imported := 0

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for i := range keys {
			key := &keys[i]

			// Check if key already exists
			var count int64
			if err := tx.Model(&DBKey{}).Where("api_key = ?", key.APIKey).Count(&count).Error; err != nil {
				return fmt.Errorf("failed to check key existence: %w", err)
			}
			if count > 0 {
				continue // Skip duplicate
			}

			// Create key using transaction-scoped db
			dbKey := keyToDBKey(key)
			if err := tx.Create(&dbKey).Error; err != nil {
				return fmt.Errorf("failed to create key: %w", err)
			}
			imported++
		}
		return nil
	})

	if err != nil {
		return 0, err
	}
	return imported, nil
}

// KeyExists checks if a key with the given API key value exists.
func (s *Storage) KeyExists(apiKey string) (bool, error) {
	var count int64
	if err := s.db.Model(&DBKey{}).Where("api_key = ?", apiKey).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// ==================== Conversion Functions ====================

// keyToDBKey converts a types.Key to a DBKey for storage.
func keyToDBKey(key *types.Key) *DBKey {
	tagsJSON, _ := json.Marshal(key.Tags)

	var lastUsedAt *int64
	if key.Stats.LastUsedAt != nil {
		ts := key.Stats.LastUsedAt.Unix()
		lastUsedAt = &ts
	}

	// Serialize ModelUsage map to JSON
	modelUsageJSON := ""
	if len(key.Stats.ModelUsage) > 0 {
		if data, err := json.Marshal(key.Stats.ModelUsage); err == nil {
			modelUsageJSON = string(data)
		}
	}

	return &DBKey{
		ID:               key.ID,
		APIKey:           key.APIKey,
		Name:             key.Name,
		Tags:             string(tagsJSON),
		Enabled:          key.Enabled,
		RequestCount:     key.Stats.RequestCount,
		SuccessCount:     key.Stats.SuccessCount,
		ErrorCount:       key.Stats.ErrorCount,
		PromptTokens:     key.Stats.PromptTokens,
		CompletionTokens: key.Stats.CompletionTokens,
		ModelUsage:       modelUsageJSON,
		LastUsedAt:       lastUsedAt,
		CreatedAt:        key.CreatedAt.Unix(),
		UpdatedAt:        key.UpdatedAt.Unix(),
	}
}

// dbKeyToKey converts a DBKey to a types.Key.
func dbKeyToKey(dbKey *DBKey) *types.Key {
	var tags []string
	if dbKey.Tags != "" {
		_ = json.Unmarshal([]byte(dbKey.Tags), &tags)
	}
	if tags == nil {
		tags = []string{}
	}

	var lastUsedAt *time.Time
	if dbKey.LastUsedAt != nil {
		t := time.Unix(*dbKey.LastUsedAt, 0)
		lastUsedAt = &t
	}

	// Deserialize ModelUsage from JSON
	var modelUsage map[string]int64
	if dbKey.ModelUsage != "" {
		_ = json.Unmarshal([]byte(dbKey.ModelUsage), &modelUsage)
	}

	return &types.Key{
		ID:        dbKey.ID,
		APIKey:    dbKey.APIKey,
		MaskedKey: types.MaskAPIKey(dbKey.APIKey),
		Name:      dbKey.Name,
		Status:    types.KeyStatusActive, // Status is runtime-only
		Enabled:   dbKey.Enabled,
		Tags:      tags,
		Stats: types.KeyStats{
			RequestCount:     dbKey.RequestCount,
			SuccessCount:     dbKey.SuccessCount,
			ErrorCount:       dbKey.ErrorCount,
			PromptTokens:     dbKey.PromptTokens,
			CompletionTokens: dbKey.CompletionTokens,
			LastUsedAt:       lastUsedAt,
			ModelUsage:       modelUsage,
		},
		CreatedAt: time.Unix(dbKey.CreatedAt, 0),
		UpdatedAt: time.Unix(dbKey.UpdatedAt, 0),
	}
}
