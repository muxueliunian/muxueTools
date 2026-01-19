// Package gemini provides the Gemini API client with format conversion.
package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"muxueTools/internal/types"
)

// ==================== Mock KeyPool ====================

// mockKey is a helper to create a Key for testing.
func mockKey(id, apiKey string) *types.Key {
	return &types.Key{
		ID:        id,
		APIKey:    apiKey,
		MaskedKey: "***" + apiKey[len(apiKey)-4:],
		Name:      "Test Key " + id,
		Status:    types.KeyStatusActive,
		Enabled:   true,
	}
}

// mockPool is a simple mock implementation of key pool for testing.
type mockPool struct {
	mu             sync.Mutex
	keys           []*types.Key
	getKeyFunc     func() (*types.Key, error)
	successReports []successReport
	failureReports []failureReport
}

type successReport struct {
	keyID            string
	promptTokens     int
	completionTokens int
	model            string
}

type failureReport struct {
	keyID string
	err   error
	model string
}

func newMockPool(keys ...*types.Key) *mockPool {
	return &mockPool{
		keys:           keys,
		successReports: make([]successReport, 0),
		failureReports: make([]failureReport, 0),
	}
}

func (p *mockPool) GetKey() (*types.Key, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.getKeyFunc != nil {
		return p.getKeyFunc()
	}
	if len(p.keys) == 0 {
		return nil, types.ErrNoAvailableKeys
	}
	return p.keys[0], nil
}

func (p *mockPool) ReleaseKey(key *types.Key) {
	// No-op for mock
}

func (p *mockPool) ReportSuccess(key *types.Key, promptTokens, completionTokens int, model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.successReports = append(p.successReports, successReport{
		keyID:            key.ID,
		promptTokens:     promptTokens,
		completionTokens: completionTokens,
		model:            model,
	})
}

func (p *mockPool) ReportFailure(key *types.Key, err error, model string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failureReports = append(p.failureReports, failureReport{
		keyID: key.ID,
		err:   err,
		model: model,
	})
}

// ==================== Test Helpers ====================

// newTestClient creates a Client pointing to a test server.
func newTestClient(serverURL string, pool KeyPoolInterface) *Client {
	return &Client{
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		pool:           pool,
		baseURL:        serverURL,
		requestTimeout: 30 * time.Second,
	}
}

// createGeminiResponse helper creates a valid Gemini response JSON.
func createGeminiResponse(text string, finishReason string, promptTokens, completionTokens int) string {
	resp := types.GeminiResponse{
		Candidates: []types.GeminiCandidate{
			{
				Content: &types.GeminiContent{
					Parts: []types.GeminiPart{{Text: text}},
					Role:  "model",
				},
				FinishReason: finishReason,
				Index:        0,
			},
		},
		UsageMetadata: &types.GeminiUsageMetadata{
			PromptTokenCount:     promptTokens,
			CandidatesTokenCount: completionTokens,
			TotalTokenCount:      promptTokens + completionTokens,
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// createGeminiErrorResponse helper creates a Gemini error response JSON.
func createGeminiErrorResponse(code int, message, status string) string {
	resp := types.GeminiErrorResponse{
		Error: types.GeminiErrorDetail{
			Code:    code,
			Message: message,
			Status:  status,
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// ==================== Happy Path Tests ====================

func TestClient_ChatCompletion_SimpleText(t *testing.T) {
	// Arrange: Create mock server that returns a valid response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "generateContent") {
			t.Errorf("Expected path containing 'generateContent', got %s", r.URL.Path)
		}

		// Validate API key in query
		apiKey := r.URL.Query().Get("key")
		if apiKey != "test-api-key-1234" {
			t.Errorf("Expected API key 'test-api-key-1234', got '%s'", apiKey)
		}

		// Return valid response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(createGeminiResponse("Hello, world!", "STOP", 10, 5)))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-api-key-1234"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act
	resp, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", resp.Choices[0].Message.Content)
	}
	if resp.Choices[0].FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", resp.Choices[0].FinishReason)
	}
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("Expected prompt_tokens 10, got %d", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 5 {
		t.Errorf("Expected completion_tokens 5, got %d", resp.Usage.CompletionTokens)
	}

	// Verify pool was notified of success
	if len(pool.successReports) != 1 {
		t.Errorf("Expected 1 success report, got %d", len(pool.successReports))
	}
}

func TestClient_ChatCompletion_WithGenerationConfig(t *testing.T) {
	// Arrange: Capture request body to verify generation config
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(createGeminiResponse("Test response", "STOP", 5, 3)))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	temperature := 0.7
	maxTokens := 100
	req := &types.ChatCompletionRequest{
		Model:       "gpt-4",
		Messages:    []types.Message{types.NewTextContent("user", "Test")},
		Temperature: &temperature,
		MaxTokens:   &maxTokens,
	}

	// Act
	_, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify request body contains generation config
	var geminiReq types.GeminiRequest
	if err := json.Unmarshal(capturedBody, &geminiReq); err != nil {
		t.Fatalf("Failed to parse request body: %v", err)
	}
	if geminiReq.GenerationConfig == nil {
		t.Fatal("Expected generationConfig in request, got nil")
	}
	if geminiReq.GenerationConfig.Temperature == nil || *geminiReq.GenerationConfig.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", geminiReq.GenerationConfig.Temperature)
	}
	if geminiReq.GenerationConfig.MaxOutputTokens == nil || *geminiReq.GenerationConfig.MaxOutputTokens != 100 {
		t.Errorf("Expected maxOutputTokens 100, got %v", geminiReq.GenerationConfig.MaxOutputTokens)
	}
}

// ==================== Error Cases Tests ====================

func TestClient_ChatCompletion_NoAvailableKeys(t *testing.T) {
	// Arrange: Empty pool
	pool := newMockPool()
	client := newTestClient("http://unused", pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act
	resp, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
	if !types.IsAppError(err) {
		t.Errorf("Expected AppError, got %T", err)
	}
}

func TestClient_ChatCompletion_RateLimitError(t *testing.T) {
	// Arrange: Server returns 429
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(createGeminiErrorResponse(429, "Resource exhausted", "RESOURCE_EXHAUSTED")))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act
	_, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for 429, got nil")
	}

	// Verify pool was notified of failure
	if len(pool.failureReports) != 1 {
		t.Errorf("Expected 1 failure report, got %d", len(pool.failureReports))
	}
}

func TestClient_ChatCompletion_UpstreamError(t *testing.T) {
	// Arrange: Server returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(createGeminiErrorResponse(500, "Internal error", "INTERNAL")))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act
	_, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for 500, got nil")
	}
	appErr := types.AsAppError(err)
	if appErr.HTTPStatus != http.StatusBadGateway {
		t.Errorf("Expected HTTPStatus 502, got %d", appErr.HTTPStatus)
	}
}

func TestClient_ChatCompletion_InvalidAPIKey(t *testing.T) {
	// Arrange: Server returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(createGeminiErrorResponse(401, "API key not valid", "UNAUTHENTICATED")))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "invalid-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act
	_, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for 401, got nil")
	}
	appErr := types.AsAppError(err)
	if appErr.Code != types.ErrCodeAuthentication {
		t.Errorf("Expected error code %d, got %d", types.ErrCodeAuthentication, appErr.Code)
	}
}

func TestClient_ChatCompletion_ContextCancellation(t *testing.T) {
	// Arrange: Server that delays
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second) // Simulate slow response
		w.Write([]byte(createGeminiResponse("Late response", "STOP", 1, 1)))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act with cancelled context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.ChatCompletion(ctx, req)

	// Assert
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}
}

// ==================== Streaming Tests ====================

func TestClient_ChatCompletionStream_Success(t *testing.T) {
	// Arrange: Server returns SSE stream
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify alt=sse parameter
		if r.URL.Query().Get("alt") != "sse" {
			t.Errorf("Expected alt=sse query param")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Expected ResponseWriter to be Flusher")
		}

		// Send multiple chunks
		chunks := []string{
			`data: {"candidates":[{"content":{"parts":[{"text":"Hello"}],"role":"model"},"index":0}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":" world"}],"role":"model"},"index":0}]}`,
			`data: {"candidates":[{"content":{"parts":[{"text":"!"}],"role":"model"},"finishReason":"STOP","index":0}],"usageMetadata":{"promptTokenCount":5,"candidatesTokenCount":3,"totalTokenCount":8}}`,
		}

		for _, chunk := range chunks {
			w.Write([]byte(chunk + "\n\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
		Stream:   true,
	}

	// Act
	eventChan, err := client.ChatCompletionStream(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Collect all events
	var contents []string
	var lastFinishReason string
	for event := range eventChan {
		if event.Err != nil {
			t.Fatalf("Unexpected stream error: %v", event.Err)
		}
		if event.Chunk != nil && len(event.Chunk.Choices) > 0 {
			contents = append(contents, event.Chunk.Choices[0].Delta.Content)
			if event.Chunk.Choices[0].FinishReason != "" {
				lastFinishReason = event.Chunk.Choices[0].FinishReason
			}
		}
	}

	// Assert
	fullContent := strings.Join(contents, "")
	if fullContent != "Hello world!" {
		t.Errorf("Expected 'Hello world!', got '%s'", fullContent)
	}
	if lastFinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got '%s'", lastFinishReason)
	}
}

func TestClient_ChatCompletionStream_ContextCancellation(t *testing.T) {
	// Arrange: Server sends stream slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		flusher := w.(http.Flusher)

		// Send one chunk, then wait
		w.Write([]byte(`data: {"candidates":[{"content":{"parts":[{"text":"Start"}],"role":"model"},"index":0}]}` + "\n\n"))
		flusher.Flush()

		time.Sleep(5 * time.Second) // Simulate slow stream
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
		Stream:   true,
	}

	// Act with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	eventChan, err := client.ChatCompletionStream(ctx, req)
	if err != nil {
		t.Fatalf("Unexpected initial error: %v", err)
	}

	// Drain channel and expect context error
	var gotError bool
	for event := range eventChan {
		if event.Err != nil {
			gotError = true
		}
	}

	if !gotError {
		t.Error("Expected context cancellation error in stream")
	}
}

// ==================== Model Mapping Tests ====================

func TestClient_ChatCompletion_ModelMapping(t *testing.T) {
	// Arrange: Capture the URL path to verify model was mapped
	var capturedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(createGeminiResponse("Response", "STOP", 1, 1)))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4", // Should be mapped to gemini-1.5-pro-latest
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
	}

	// Act
	resp, err := client.ChatCompletion(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify model was mapped in URL
	if !strings.Contains(capturedPath, "gemini-1.5-pro-latest") {
		t.Errorf("Expected path to contain mapped model 'gemini-1.5-pro-latest', got '%s'", capturedPath)
	}

	// Verify response model is the original requested model (for OpenAI SDK compatibility)
	if resp.Model != "gpt-4" {
		t.Errorf("Expected response model 'gpt-4', got '%s'", resp.Model)
	}
}

// ==================== Nil Request Tests ====================

func TestClient_ChatCompletion_NilRequest(t *testing.T) {
	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient("http://unused", pool)

	// Act
	resp, err := client.ChatCompletion(context.Background(), nil)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nil request, got nil")
	}
	if resp != nil {
		t.Errorf("Expected nil response, got %v", resp)
	}
	appErr := types.AsAppError(err)
	if appErr.Code != types.ErrCodeInvalidRequest {
		t.Errorf("Expected error code %d, got %d", types.ErrCodeInvalidRequest, appErr.Code)
	}
}

func TestClient_ChatCompletionStream_NilRequest(t *testing.T) {
	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient("http://unused", pool)

	// Act
	eventChan, err := client.ChatCompletionStream(context.Background(), nil)

	// Assert
	if err == nil {
		t.Fatal("Expected error for nil request, got nil")
	}
	if eventChan != nil {
		t.Errorf("Expected nil channel, got %v", eventChan)
	}
	appErr := types.AsAppError(err)
	if appErr.Code != types.ErrCodeInvalidRequest {
		t.Errorf("Expected error code %d, got %d", types.ErrCodeInvalidRequest, appErr.Code)
	}
}

// ==================== Stream Error Order Tests ====================

func TestClient_ChatCompletionStream_HTTPErrorReportsBeforeRelease(t *testing.T) {
	// This test verifies that ReportFailure is called BEFORE ReleaseKey
	// when an HTTP error occurs during streaming request initiation

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(createGeminiErrorResponse(429, "Rate limited", "RESOURCE_EXHAUSTED")))
	}))
	defer server.Close()

	pool := newMockPool(mockKey("key1", "test-key"))
	client := newTestClient(server.URL, pool)

	req := &types.ChatCompletionRequest{
		Model:    "gpt-4",
		Messages: []types.Message{types.NewTextContent("user", "Hello")},
		Stream:   true,
	}

	// Act
	_, err := client.ChatCompletionStream(context.Background(), req)

	// Assert
	if err == nil {
		t.Fatal("Expected error for 429, got nil")
	}

	// Verify failure was reported (the order is correct if ReportFailure was called)
	if len(pool.failureReports) != 1 {
		t.Errorf("Expected 1 failure report, got %d", len(pool.failureReports))
	}
}
