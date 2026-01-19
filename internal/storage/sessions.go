// Package storage provides SQLite-based persistence layer for MuxueTools.
package storage

import (
	"fmt"
	"time"

	"muxueTools/internal/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ==================== Session Storage Methods ====================

// CreateSession creates a new session in the database.
func (s *Storage) CreateSession(session *types.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	session.UpdatedAt = time.Now()

	if err := s.db.Create(session).Error; err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetSession retrieves a session by ID.
func (s *Storage) GetSession(id string) (*types.Session, error) {
	var session types.Session
	if err := s.db.Where("id = ?", id).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, types.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

// ListSessions retrieves sessions with pagination.
// Returns sessions, total count, and error.
func (s *Storage) ListSessions(limit, offset int) ([]types.Session, int, error) {
	var sessions []types.Session
	var total int64

	// Get total count
	if err := s.db.Model(&types.Session{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Get paginated results
	query := s.db.Order("updated_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&sessions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, int(total), nil
}

// UpdateSession updates an existing session.
func (s *Storage) UpdateSession(session *types.Session) error {
	session.UpdatedAt = time.Now()
	result := s.db.Model(&types.Session{}).Where("id = ?", session.ID).Updates(map[string]interface{}{
		"title":         session.Title,
		"model":         session.Model,
		"message_count": session.MessageCount,
		"total_tokens":  session.TotalTokens,
		"updated_at":    session.UpdatedAt,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return types.ErrSessionNotFound
	}
	return nil
}

// DeleteSession deletes a session and all its messages.
func (s *Storage) DeleteSession(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete all messages first
		if err := tx.Where("session_id = ?", id).Delete(&types.ChatMessage{}).Error; err != nil {
			return fmt.Errorf("failed to delete session messages: %w", err)
		}

		// Delete the session
		result := tx.Where("id = ?", id).Delete(&types.Session{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete session: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return types.ErrSessionNotFound
		}

		return nil
	})
}

// ==================== Message Storage Methods ====================

// AddMessage adds a new message to a session.
func (s *Storage) AddMessage(message *types.ChatMessage) error {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	if message.CreatedAt.IsZero() {
		message.CreatedAt = time.Now()
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Insert the message
		if err := tx.Create(message).Error; err != nil {
			return fmt.Errorf("failed to add message: %w", err)
		}

		// Update session statistics
		result := tx.Model(&types.Session{}).Where("id = ?", message.SessionID).Updates(map[string]interface{}{
			"message_count": gorm.Expr("message_count + 1"),
			"total_tokens":  gorm.Expr("total_tokens + ?", message.TotalTokens()),
			"updated_at":    time.Now(),
		})
		if result.Error != nil {
			return fmt.Errorf("failed to update session stats: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return types.ErrSessionNotFound
		}

		return nil
	})
}

// GetMessages retrieves all messages for a session.
func (s *Storage) GetMessages(sessionID string) ([]types.ChatMessage, error) {
	var messages []types.ChatMessage
	if err := s.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	return messages, nil
}

// DeleteMessages deletes all messages for a session.
func (s *Storage) DeleteMessages(sessionID string) error {
	if err := s.db.Where("session_id = ?", sessionID).Delete(&types.ChatMessage{}).Error; err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}
	return nil
}

// GetMessageCount returns the number of messages in a session.
func (s *Storage) GetMessageCount(sessionID string) (int, error) {
	var count int64
	if err := s.db.Model(&types.ChatMessage{}).Where("session_id = ?", sessionID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}
	return int(count), nil
}
