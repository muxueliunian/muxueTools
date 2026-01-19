// Package storage provides SQLite-based persistence layer for MuxueTools.
package storage

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"muxueTools/internal/types"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// newTestStorage creates an in-memory storage for testing.
// Uses shared cache mode to allow concurrent access within the same process.
func newTestStorage(t *testing.T) *Storage {
	// Use file::memory:?cache=shared for concurrent access support
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	storage, err := NewStorageWithDB(db)
	require.NoError(t, err)

	return storage
}

// ==================== Key Storage Tests ====================

func TestStorage_CreateKey(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	key := &types.Key{
		ID:        uuid.New().String(),
		APIKey:    "AIzaSyTest1234567890",
		MaskedKey: types.MaskAPIKey("AIzaSyTest1234567890"),
		Name:      "Test Key",
		Status:    types.KeyStatusActive,
		Enabled:   true,
		Tags:      []string{"test", "dev"},
		Stats:     types.KeyStats{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateKey(key)
	assert.NoError(t, err)

	// Verify key was created
	retrieved, err := storage.GetKey(key.ID)
	assert.NoError(t, err)
	assert.Equal(t, key.ID, retrieved.ID)
	assert.Equal(t, key.APIKey, retrieved.APIKey)
	assert.Equal(t, key.Name, retrieved.Name)
	assert.Equal(t, key.Tags, retrieved.Tags)
}

func TestStorage_GetKey_NotFound(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	_, err := storage.GetKey("nonexistent-id")
	assert.ErrorIs(t, err, types.ErrKeyNotFound)
}

func TestStorage_GetKeyByAPIKey(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	key := &types.Key{
		ID:        uuid.New().String(),
		APIKey:    "AIzaSyUnique123",
		MaskedKey: types.MaskAPIKey("AIzaSyUnique123"),
		Name:      "Unique Key",
		Enabled:   true,
		Tags:      []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateKey(key)
	require.NoError(t, err)

	retrieved, err := storage.GetKeyByAPIKey("AIzaSyUnique123")
	assert.NoError(t, err)
	assert.Equal(t, key.ID, retrieved.ID)
}

func TestStorage_ListKeys(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	// Create multiple keys
	for i := 0; i < 3; i++ {
		key := &types.Key{
			ID:        uuid.New().String(),
			APIKey:    "AIzaSyKey" + string(rune('A'+i)) + "12345",
			Name:      "Key " + string(rune('A'+i)),
			Enabled:   true,
			Tags:      []string{},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := storage.CreateKey(key)
		require.NoError(t, err)
	}

	keys, err := storage.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, keys, 3)
}

func TestStorage_UpdateKey(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	key := &types.Key{
		ID:        uuid.New().String(),
		APIKey:    "AIzaSyUpdate123",
		Name:      "Original Name",
		Enabled:   true,
		Tags:      []string{"original"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateKey(key)
	require.NoError(t, err)

	// Update key
	key.Name = "Updated Name"
	key.Tags = []string{"updated"}
	key.Stats.RequestCount = 10
	key.Stats.SuccessCount = 9

	err = storage.UpdateKey(key)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := storage.GetKey(key.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, []string{"updated"}, retrieved.Tags)
	assert.Equal(t, int64(10), retrieved.Stats.RequestCount)
}

func TestStorage_DeleteKey(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	key := &types.Key{
		ID:        uuid.New().String(),
		APIKey:    "AIzaSyDelete123",
		Name:      "To Delete",
		Enabled:   true,
		Tags:      []string{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := storage.CreateKey(key)
	require.NoError(t, err)

	// Delete key
	err = storage.DeleteKey(key.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = storage.GetKey(key.ID)
	assert.ErrorIs(t, err, types.ErrKeyNotFound)
}

func TestStorage_ImportKeys(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	keys := []types.Key{
		{ID: uuid.New().String(), APIKey: "AIzaSyImport1", Name: "Import 1", Enabled: true, Tags: []string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New().String(), APIKey: "AIzaSyImport2", Name: "Import 2", Enabled: true, Tags: []string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: uuid.New().String(), APIKey: "AIzaSyImport1", Name: "Duplicate", Enabled: true, Tags: []string{}, CreatedAt: time.Now(), UpdatedAt: time.Now()}, // Duplicate
	}

	imported, err := storage.ImportKeys(keys)
	assert.NoError(t, err)
	assert.Equal(t, 2, imported) // Only 2 unique keys

	// Verify imported keys
	allKeys, err := storage.ListKeys()
	assert.NoError(t, err)
	assert.Len(t, allKeys, 2)
}

// ==================== Session Storage Tests ====================

func TestStorage_CreateSession(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	session := &types.Session{
		Title: "Test Session",
		Model: "gemini-1.5-pro",
	}

	err := storage.CreateSession(session)
	assert.NoError(t, err)
	assert.NotEmpty(t, session.ID)

	// Verify session was created
	retrieved, err := storage.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session.ID, retrieved.ID)
	assert.Equal(t, session.Title, retrieved.Title)
	assert.Equal(t, session.Model, retrieved.Model)
}

func TestStorage_GetSession_NotFound(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	_, err := storage.GetSession("nonexistent-id")
	assert.ErrorIs(t, err, types.ErrSessionNotFound)
}

func TestStorage_ListSessions(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	// Create multiple sessions
	for i := 0; i < 5; i++ {
		session := &types.Session{
			Title: "Session " + string(rune('A'+i)),
			Model: "gemini-1.5-flash",
		}
		err := storage.CreateSession(session)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Test pagination
	sessions, total, err := storage.ListSessions(3, 0)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, sessions, 3)

	sessions, total, err = storage.ListSessions(3, 3)
	assert.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, sessions, 2)
}

func TestStorage_UpdateSession(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	session := &types.Session{
		Title: "Original Title",
		Model: "gemini-1.5-pro",
	}

	err := storage.CreateSession(session)
	require.NoError(t, err)

	// Update session
	session.Title = "Updated Title"
	session.Model = "gemini-1.5-flash"

	err = storage.UpdateSession(session)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := storage.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, "gemini-1.5-flash", retrieved.Model)
}

func TestStorage_DeleteSession(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	// Create session with messages
	session := &types.Session{
		Title: "To Delete",
		Model: "gemini-1.5-pro",
	}
	err := storage.CreateSession(session)
	require.NoError(t, err)

	// Add some messages
	for i := 0; i < 3; i++ {
		msg := &types.ChatMessage{
			SessionID: session.ID,
			Role:      "user",
			Content:   "Message content",
		}
		err = storage.AddMessage(msg)
		require.NoError(t, err)
	}

	// Delete session (should cascade delete messages)
	err = storage.DeleteSession(session.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = storage.GetSession(session.ID)
	assert.ErrorIs(t, err, types.ErrSessionNotFound)

	// Verify messages are also deleted
	messages, err := storage.GetMessages(session.ID)
	assert.NoError(t, err)
	assert.Empty(t, messages)
}

// ==================== Message Storage Tests ====================

func TestStorage_AddMessage(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	// Create session first
	session := &types.Session{
		Title: "Chat Session",
		Model: "gemini-1.5-pro",
	}
	err := storage.CreateSession(session)
	require.NoError(t, err)

	// Add message
	msg := &types.ChatMessage{
		SessionID:        session.ID,
		Role:             "user",
		Content:          "Hello, world!",
		PromptTokens:     10,
		CompletionTokens: 0,
	}

	err = storage.AddMessage(msg)
	assert.NoError(t, err)
	assert.NotEmpty(t, msg.ID)

	// Verify session stats updated
	updated, err := storage.GetSession(session.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, updated.MessageCount)
	assert.Equal(t, 10, updated.TotalTokens)
}

func TestStorage_GetMessages(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	// Create session
	session := &types.Session{
		Title: "Chat Session",
		Model: "gemini-1.5-pro",
	}
	err := storage.CreateSession(session)
	require.NoError(t, err)

	// Add multiple messages
	roles := []string{"user", "assistant", "user"}
	for _, role := range roles {
		msg := &types.ChatMessage{
			SessionID: session.ID,
			Role:      role,
			Content:   "Message from " + role,
		}
		err = storage.AddMessage(msg)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get messages
	messages, err := storage.GetMessages(session.ID)
	assert.NoError(t, err)
	assert.Len(t, messages, 3)

	// Verify order (oldest first)
	assert.Equal(t, "user", messages[0].Role)
	assert.Equal(t, "assistant", messages[1].Role)
	assert.Equal(t, "user", messages[2].Role)
}

func TestStorage_AddMessage_SessionNotFound(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	msg := &types.ChatMessage{
		SessionID: "nonexistent-session",
		Role:      "user",
		Content:   "Hello",
	}

	err := storage.AddMessage(msg)
	assert.ErrorIs(t, err, types.ErrSessionNotFound)
}

// ==================== Concurrent Tests ====================

func TestStorage_Concurrent_AddMessages(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	// Create session
	session := &types.Session{
		Title: "Concurrent Test",
		Model: "gemini-1.5-pro",
	}
	err := storage.CreateSession(session)
	require.NoError(t, err)

	// Concurrent message adds
	const numMessages = 10
	var wg sync.WaitGroup
	errors := make(chan error, numMessages)

	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msg := &types.ChatMessage{
				SessionID: session.ID,
				Role:      "user",
				Content:   fmt.Sprintf("Message %d", i),
			}
			if err := storage.AddMessage(msg); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("AddMessage failed: %v", err)
	}

	// Verify session statistics
	updated, err := storage.GetSession(session.ID)
	require.NoError(t, err)
	assert.Equal(t, numMessages, updated.MessageCount)
}

func TestStorage_Ping(t *testing.T) {
	storage := newTestStorage(t)
	defer storage.Close()

	err := storage.Ping()
	assert.NoError(t, err)
}
