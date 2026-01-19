// Package gemini provides the Gemini API client with format conversion.
package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mxlnapi/internal/types"
)

// ==================== Client Options ====================

// ClientOption is a functional option for configuring the Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithRequestTimeout sets the request timeout.
func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.requestTimeout = timeout
	}
}

// ==================== Key Pool Interface ====================

// KeyPoolInterface defines the interface for key pool operations.
// This allows for easy mocking in tests.
type KeyPoolInterface interface {
	GetKey() (*types.Key, error)
	ReleaseKey(key *types.Key)
	ReportSuccess(key *types.Key, promptTokens, completionTokens int, model string)
	ReportFailure(key *types.Key, err error, model string)
}

// ==================== Stream Event ====================

// StreamEvent represents an event in the streaming response.
type StreamEvent struct {
	Chunk *types.ChatCompletionChunk
	Err   error
	Done  bool
}

// ==================== Client ====================

// Client is the Gemini API client that handles requests and format conversion.
type Client struct {
	httpClient     *http.Client
	pool           KeyPoolInterface
	baseURL        string
	requestTimeout time.Duration
}

// NewClient creates a new Gemini API client.
func NewClient(pool KeyPoolInterface, opts ...ClientOption) *Client {
	client := &Client{
		httpClient:     &http.Client{Timeout: 60 * time.Second},
		pool:           pool,
		baseURL:        types.GeminiBaseURL,
		requestTimeout: 120 * time.Second,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

// ==================== Chat Completion (Blocking) ====================

// ChatCompletion sends a blocking chat completion request.
func (c *Client) ChatCompletion(ctx context.Context, req *types.ChatCompletionRequest) (*types.ChatCompletionResponse, error) {
	// 0. Validate request
	if req == nil {
		return nil, types.NewInvalidRequestError("Request cannot be nil")
	}

	// 1. Get a key from the pool
	key, err := c.pool.GetKey()
	if err != nil {
		return nil, err
	}
	defer c.pool.ReleaseKey(key)

	// 2. Convert OpenAI request to Gemini format
	geminiReq, err := ConvertOpenAIRequest(req)
	if err != nil {
		return nil, err
	}

	// 3. Map model name
	geminiModel := MapModelName(req.Model)

	// 4. Build URL
	url := c.buildURL(geminiModel, key.APIKey, false)

	// 5. Send HTTP request
	geminiResp, err := c.doRequest(ctx, url, geminiReq)
	if err != nil {
		c.pool.ReportFailure(key, err, req.Model)
		return nil, err
	}

	// 6. Convert Gemini response to OpenAI format
	openAIResp, err := ConvertGeminiResponse(geminiResp, req.Model)
	if err != nil {
		c.pool.ReportFailure(key, err, req.Model)
		return nil, err
	}

	// 7. Report success to pool
	promptTokens := 0
	completionTokens := 0
	if geminiResp.UsageMetadata != nil {
		promptTokens = geminiResp.UsageMetadata.PromptTokenCount
		completionTokens = geminiResp.UsageMetadata.CandidatesTokenCount
	}
	c.pool.ReportSuccess(key, promptTokens, completionTokens, req.Model)

	return openAIResp, nil
}

// ==================== Chat Completion (Streaming) ====================

// ChatCompletionStream sends a streaming chat completion request.
// It returns a channel that will receive StreamEvent objects.
func (c *Client) ChatCompletionStream(ctx context.Context, req *types.ChatCompletionRequest) (<-chan StreamEvent, error) {
	// 0. Validate request
	if req == nil {
		return nil, types.NewInvalidRequestError("Request cannot be nil")
	}

	// 1. Get a key from the pool
	key, err := c.pool.GetKey()
	if err != nil {
		return nil, err
	}

	// 2. Convert OpenAI request to Gemini format
	geminiReq, err := ConvertOpenAIRequest(req)
	if err != nil {
		c.pool.ReleaseKey(key)
		return nil, err
	}

	// 3. Map model name
	geminiModel := MapModelName(req.Model)

	// 4. Build URL with streaming endpoint
	url := c.buildURL(geminiModel, key.APIKey, true)

	// 5. Marshal request body
	body, err := json.Marshal(geminiReq)
	if err != nil {
		c.pool.ReleaseKey(key)
		return nil, types.NewInternalError("Failed to marshal request").WithCause(err)
	}

	// 6. Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		c.pool.ReleaseKey(key)
		return nil, types.NewInternalError("Failed to create request").WithCause(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 7. Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		c.pool.ReportFailure(key, err, req.Model) // Report failure BEFORE releasing
		c.pool.ReleaseKey(key)
		return nil, c.wrapHTTPError(err)
	}

	// 8. Check for error status codes
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		appErr := c.parseErrorResponse(resp)
		c.pool.ReportFailure(key, appErr, req.Model) // Report failure BEFORE releasing
		c.pool.ReleaseKey(key)
		return nil, appErr
	}

	// 9. Create output channel and start streaming goroutine
	eventChan := make(chan StreamEvent)
	go c.streamResponse(ctx, resp, key, req.Model, eventChan)

	return eventChan, nil
}

// streamResponse reads SSE events from the response and sends them to the channel.
func (c *Client) streamResponse(ctx context.Context, resp *http.Response, key *types.Key, originalModel string, eventChan chan<- StreamEvent) {
	defer resp.Body.Close()
	defer c.pool.ReleaseKey(key)
	defer close(eventChan)

	reader := bufio.NewReader(resp.Body)
	chunkIndex := 0
	var totalPromptTokens, totalCompletionTokens int

	for {
		select {
		case <-ctx.Done():
			eventChan <- StreamEvent{Err: ctx.Err()}
			c.pool.ReportFailure(key, ctx.Err(), originalModel)
			return
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// Normal end of stream
				c.pool.ReportSuccess(key, totalPromptTokens, totalCompletionTokens, originalModel)
				return
			}
			eventChan <- StreamEvent{Err: types.NewUpstreamError("Stream read error").WithCause(err)}
			c.pool.ReportFailure(key, err, originalModel)
			return
		}

		// Trim whitespace
		line = bytes.TrimSpace(line)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Check for data: prefix
		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}

		// Extract JSON data
		jsonData := line[6:] // Remove "data: " prefix

		// Parse Gemini response
		var geminiResp types.GeminiResponse
		if err := json.Unmarshal(jsonData, &geminiResp); err != nil {
			eventChan <- StreamEvent{Err: types.NewUpstreamError("Failed to parse stream chunk").WithCause(err)}
			c.pool.ReportFailure(key, err, originalModel)
			return
		}

		// Track token usage from final chunk
		if geminiResp.UsageMetadata != nil {
			totalPromptTokens = geminiResp.UsageMetadata.PromptTokenCount
			totalCompletionTokens = geminiResp.UsageMetadata.CandidatesTokenCount
		}

		// Convert to OpenAI chunk format
		openAIChunk, err := ConvertGeminiStreamChunk(&geminiResp, originalModel, chunkIndex)
		if err != nil {
			eventChan <- StreamEvent{Err: err}
			c.pool.ReportFailure(key, err, originalModel)
			return
		}

		chunkIndex++

		// Send chunk to channel with context awareness
		select {
		case eventChan <- StreamEvent{Chunk: openAIChunk}:
			// Successfully sent
		case <-ctx.Done():
			c.pool.ReportFailure(key, ctx.Err(), originalModel)
			return
		}

		// Check if this is the final chunk
		if len(geminiResp.Candidates) > 0 && geminiResp.Candidates[0].FinishReason != "" {
			// Send Done event to signal completion
			select {
			case eventChan <- StreamEvent{Done: true}:
				// Successfully sent Done
			case <-ctx.Done():
				// Context cancelled, but we already sent the content
			}
			c.pool.ReportSuccess(key, totalPromptTokens, totalCompletionTokens, originalModel)
			return
		}
	}
}

// ==================== Internal Helpers ====================

// buildURL constructs the Gemini API URL.
func (c *Client) buildURL(model, apiKey string, stream bool) string {
	endpoint := "generateContent"
	if stream {
		endpoint = "streamGenerateContent"
	}

	url := fmt.Sprintf("%s/models/%s:%s?key=%s", c.baseURL, model, endpoint, apiKey)
	if stream {
		url += "&alt=sse"
	}

	return url
}

// doRequest sends an HTTP request and returns the parsed Gemini response.
func (c *Client) doRequest(ctx context.Context, url string, geminiReq *types.GeminiRequest) (*types.GeminiResponse, error) {
	// Marshal request body
	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, types.NewInternalError("Failed to marshal request").WithCause(err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, types.NewInternalError("Failed to create request").WithCause(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, c.wrapHTTPError(err)
	}
	defer resp.Body.Close()

	// Check for error status codes
	if resp.StatusCode != http.StatusOK {
		return nil, c.parseErrorResponse(resp)
	}

	// Parse response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, types.NewUpstreamError("Failed to read response").WithCause(err)
	}

	var geminiResp types.GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return nil, types.NewUpstreamError("Failed to parse response").WithCause(err)
	}

	return &geminiResp, nil
}

// parseErrorResponse parses an error response from Gemini API.
func (c *Client) parseErrorResponse(resp *http.Response) *types.AppError {
	body, _ := io.ReadAll(resp.Body)

	// Try to parse as Gemini error response
	var geminiErr types.GeminiErrorResponse
	if err := json.Unmarshal(body, &geminiErr); err == nil && geminiErr.Error.Code != 0 {
		return c.mapGeminiError(resp.StatusCode, &geminiErr)
	}

	// Fallback to generic error based on status code
	return c.mapStatusCodeToError(resp.StatusCode, string(body))
}

// mapGeminiError maps a Gemini error response to an AppError.
func (c *Client) mapGeminiError(statusCode int, geminiErr *types.GeminiErrorResponse) *types.AppError {
	switch statusCode {
	case http.StatusTooManyRequests:
		return types.NewRateLimitError(60).WithMessage(geminiErr.Error.Message)
	case http.StatusUnauthorized:
		return types.NewAuthenticationError(geminiErr.Error.Message)
	case http.StatusForbidden:
		return types.NewPermissionError(geminiErr.Error.Message)
	case http.StatusBadRequest:
		return types.NewInvalidRequestError(geminiErr.Error.Message)
	case http.StatusNotFound:
		return types.NewNotFoundError(geminiErr.Error.Message)
	default:
		return types.NewUpstreamError(geminiErr.Error.Message)
	}
}

// mapStatusCodeToError maps an HTTP status code to an AppError.
func (c *Client) mapStatusCodeToError(statusCode int, body string) *types.AppError {
	switch statusCode {
	case http.StatusTooManyRequests:
		return types.NewRateLimitError(60)
	case http.StatusUnauthorized:
		return types.NewAuthenticationError("")
	case http.StatusForbidden:
		return types.NewPermissionError("")
	case http.StatusBadRequest:
		return types.NewInvalidRequestError(body)
	case http.StatusNotFound:
		return types.NewNotFoundError("model")
	case http.StatusServiceUnavailable:
		return types.NewServiceUnavailableError("")
	default:
		return types.NewUpstreamError(fmt.Sprintf("Upstream API returned status %d", statusCode))
	}
}

// wrapHTTPError wraps an HTTP client error in an AppError.
func (c *Client) wrapHTTPError(err error) *types.AppError {
	errStr := err.Error()

	// Check for context cancellation
	if strings.Contains(errStr, "context canceled") || strings.Contains(errStr, "context deadline exceeded") {
		return types.NewInternalError("Request cancelled or timed out").WithCause(err)
	}

	// Check for connection errors
	if strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host") {
		return types.NewServiceUnavailableError("Failed to connect to Gemini API").WithCause(err)
	}

	return types.NewUpstreamError("HTTP request failed").WithCause(err)
}
