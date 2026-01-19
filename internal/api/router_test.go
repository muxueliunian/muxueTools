// Package api provides HTTP API handlers and routing for MxlnAPI.
package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mxlnapi/internal/keypool"
	"mxlnapi/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ==================== Test Setup ====================

func init() {
	gin.SetMode(gin.TestMode)
}

// createTestRouter creates a router for testing with mock dependencies.
func createTestRouter() (*gin.Engine, *keypool.Pool) {
	// Create test keys
	testKeys := []types.KeyConfig{
		{Key: "AIzaSyTestKey1XXXXXXXXXXXXXXXXX", Name: "Test Key 1", Enabled: true, Tags: []string{"test"}},
		{Key: "AIzaSyTestKey2XXXXXXXXXXXXXXXXX", Name: "Test Key 2", Enabled: true, Tags: []string{"test"}},
	}

	// Create pool
	pool := keypool.NewPool(testKeys)

	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs during tests

	// Create config
	cfg := &types.Config{
		Server:  types.DefaultServerConfig(),
		Pool:    types.DefaultPoolConfig(),
		Logging: types.DefaultLoggingConfig(),
	}

	// Create router config
	routerConfig := &RouterConfig{
		Config:  cfg,
		Pool:    pool,
		Client:  nil, // We'll skip the client for router tests
		Logger:  logger,
		Version: "test",
	}

	// Build engine manually for testing (without client-dependent routes)
	engine := gin.New()
	engine.Use(RequestIDMiddleware())
	engine.Use(CORSMiddleware())
	engine.Use(RecoveryMiddleware(logger))

	// Health handler
	healthHandler := NewHealthHandler(pool, routerConfig.Version)
	engine.GET("/health", healthHandler.Health)
	engine.GET("/ping", Ping)

	// Admin handler
	adminHandler := NewAdminHandler(pool, logger)
	api := engine.Group("/api")
	{
		keys := api.Group("/keys")
		{
			keys.GET("", adminHandler.ListKeys)
			keys.POST("", adminHandler.AddKey)
			keys.DELETE("/:id", adminHandler.DeleteKey)
			keys.POST("/:id/test", adminHandler.TestKey)
			keys.POST("/import", adminHandler.ImportKeys)
			keys.GET("/export", adminHandler.ExportKeys)
		}
		api.GET("/stats", adminHandler.GetStats)
		api.GET("/stats/keys", adminHandler.GetKeyStats)
		api.GET("/config", adminHandler.GetConfig)
		api.PUT("/config", adminHandler.UpdateConfig)
		api.GET("/update/check", adminHandler.CheckUpdate)
	}

	// Root route
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"name":    "MxlnAPI",
			"version": "test",
		})
	})

	return engine, pool
}

// ==================== Health Check Tests ====================

func TestHealthCheck_Returns200(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp types.HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp.Status != "ok" && resp.Status != "degraded" {
		t.Errorf("Expected status 'ok' or 'degraded', got '%s'", resp.Status)
	}

	if resp.Version != "test" {
		t.Errorf("Expected version 'test', got '%s'", resp.Version)
	}
}

func TestPing_Returns200(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp["message"] != "pong" {
		t.Errorf("Expected message 'pong', got '%v'", resp["message"])
	}
}

// ==================== Root Route Tests ====================

func TestRootRoute_ReturnsInfo(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp["name"] != "MxlnAPI" {
		t.Errorf("Expected name 'MxlnAPI', got '%v'", resp["name"])
	}
}

// ==================== CORS Tests ====================

func TestCORS_PresenceOfHeaders(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	engine.ServeHTTP(w, req)

	// Check CORS headers are present
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected Access-Control-Allow-Origin header")
	}
}

// ==================== Request ID Tests ====================

func TestRequestID_GeneratedAndReturned(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	engine.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set")
	}
}

func TestRequestID_UseProvidedID(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	engine.ServeHTTP(w, req)

	requestID := w.Header().Get("X-Request-ID")
	if requestID != "my-custom-id" {
		t.Errorf("Expected X-Request-ID 'my-custom-id', got '%s'", requestID)
	}
}

// ==================== Admin API Tests ====================

func TestListKeys_ReturnsKeys(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/keys", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp types.KeyListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Total != 2 {
		t.Errorf("Expected 2 keys, got %d", resp.Total)
	}

	// Verify keys are masked
	for _, key := range resp.Data {
		if len(key.MaskedKey) > 15 {
			t.Errorf("Key should be masked, got '%s'", key.MaskedKey)
		}
	}
}

func TestAddKey_ValidKey(t *testing.T) {
	engine, _ := createTestRouter()

	body := types.CreateKeyRequest{
		Key:  "AIzaSyNewTestKeyXXXXXXXXXXXXXX",
		Name: "New Test Key",
		Tags: []string{"new", "test"},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var resp types.CreateKeyResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Data.Name != "New Test Key" {
		t.Errorf("Expected name 'New Test Key', got '%s'", resp.Data.Name)
	}
}

func TestAddKey_InvalidKey(t *testing.T) {
	engine, _ := createTestRouter()

	body := types.CreateKeyRequest{
		Key:  "short",
		Name: "Invalid Key",
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/keys", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDeleteKey_NonExistent(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/keys/non-existent-id", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestGetStats_Returns200(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stats", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp types.StatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

func TestGetKeyStats_Returns200(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/stats/keys", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp types.KeyStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 key stats, got %d", len(resp.Data))
	}
}

func TestCheckUpdate_Returns200(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/update/check", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp UpdateCheckResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}
}

// ==================== Import Keys Tests ====================

// ==================== Import Keys Tests ====================

func TestImportKeys_ValidKeys(t *testing.T) {
	engine, _ := createTestRouter()

	body := types.ImportKeysRequest{
		Keys: []types.ImportKeyItem{
			{Key: "AIzaSyKey1XXXXXXXXXXXXXXXXXXXXXXXXX", Tags: []string{"imported"}},
			{Key: "AIzaSyKey2XXXXXXXXXXXXXXXXXXXXXXXXX", Tags: []string{"imported"}},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/keys/import", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp types.ImportKeysResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.Data.Imported != 2 {
		t.Errorf("Expected 2 imported, got %d", resp.Data.Imported)
	}
}

func TestImportKeys_MixedValidity(t *testing.T) {
	engine, _ := createTestRouter()

	body := types.ImportKeysRequest{
		Keys: []types.ImportKeyItem{
			{Key: "AIzaSyValidKeyXXXXXXXXXXXXXXXXXX"},
			{Key: "short"},
			{Key: "AIzaSyAnotherValidKeyXXXXXXXXXX"},
		},
	}
	jsonBody, _ := json.Marshal(body)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/keys/import", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp types.ImportKeysResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if resp.Data.Imported != 2 {
		t.Errorf("Expected 2 imported, got %d", resp.Data.Imported)
	}

	if len(resp.Data.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(resp.Data.Errors))
	}
}

// ==================== Export Keys Tests ====================

func TestExportKeys_ReturnsText(t *testing.T) {
	engine, _ := createTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/keys/export", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected Content-Type 'text/plain', got '%s'", contentType)
	}

	disposition := w.Header().Get("Content-Disposition")
	if disposition == "" {
		t.Error("Expected Content-Disposition header")
	}
}
