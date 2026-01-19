// Package gemini provides format conversion between OpenAI and Gemini API formats.
package gemini

import (
	"encoding/json"
	"testing"
	"time"

	"muxueTools/internal/types"
)

// ==================== Test Helpers ====================

// newFloat64 creates a pointer to a float64 value.
func newFloat64(v float64) *float64 { return &v }

// newInt creates a pointer to an int value.
func newInt(v int) *int { return &v }

// makeTextMessage creates a Message with plain text content.
func makeTextMessage(role, text string) types.Message {
	content, _ := json.Marshal(text)
	return types.Message{Role: role, Content: content}
}

// makeMultimodalMessage creates a Message with multimodal content parts.
func makeMultimodalMessage(role string, parts []types.ContentPart) types.Message {
	content, _ := json.Marshal(parts)
	return types.Message{Role: role, Content: content}
}

// ==================== Basic Conversion Tests ====================

func TestConvertMessages_SimpleText(t *testing.T) {
	messages := []types.Message{
		makeTextMessage("user", "Hello, how are you?"),
	}

	contents, systemInstruction, err := ConvertMessages(messages)
	if err != nil {
		t.Fatalf("ConvertMessages failed: %v", err)
	}

	if systemInstruction != nil {
		t.Error("Expected no system instruction for user-only messages")
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content, got %d", len(contents))
	}

	if contents[0].Role != "user" {
		t.Errorf("Expected role 'user', got %q", contents[0].Role)
	}

	if len(contents[0].Parts) != 1 || contents[0].Parts[0].Text != "Hello, how are you?" {
		t.Error("Text content mismatch")
	}
}

func TestConvertMessages_MultiTurn(t *testing.T) {
	messages := []types.Message{
		makeTextMessage("user", "What is 2+2?"),
		makeTextMessage("assistant", "2+2 equals 4."),
		makeTextMessage("user", "And 3+3?"),
	}

	contents, systemInstruction, err := ConvertMessages(messages)
	if err != nil {
		t.Fatalf("ConvertMessages failed: %v", err)
	}

	if systemInstruction != nil {
		t.Error("Expected no system instruction")
	}

	if len(contents) != 3 {
		t.Fatalf("Expected 3 contents, got %d", len(contents))
	}

	// Verify roles are correctly mapped
	expectedRoles := []string{"user", "model", "user"}
	for i, expected := range expectedRoles {
		if contents[i].Role != expected {
			t.Errorf("Content[%d]: expected role %q, got %q", i, expected, contents[i].Role)
		}
	}
}

func TestConvertMessages_WithSystemMessage(t *testing.T) {
	messages := []types.Message{
		makeTextMessage("system", "You are a helpful assistant."),
		makeTextMessage("user", "Hello!"),
	}

	contents, systemInstruction, err := ConvertMessages(messages)
	if err != nil {
		t.Fatalf("ConvertMessages failed: %v", err)
	}

	// System message should be extracted
	if systemInstruction == nil {
		t.Fatal("Expected system instruction to be extracted")
	}

	if len(systemInstruction.Parts) != 1 || systemInstruction.Parts[0].Text != "You are a helpful assistant." {
		t.Error("System instruction content mismatch")
	}

	// Only user message should remain in contents
	if len(contents) != 1 {
		t.Fatalf("Expected 1 content (user only), got %d", len(contents))
	}

	if contents[0].Role != "user" {
		t.Errorf("Expected role 'user', got %q", contents[0].Role)
	}
}

// ==================== Multimodal Tests ====================

func TestConvertMessages_Base64Image(t *testing.T) {
	parts := []types.ContentPart{
		{Type: "text", Text: "What's in this image?"},
		{
			Type: "image_url",
			ImageURL: &types.ImageURL{
				URL:    "data:image/jpeg;base64,/9j/4AAQSkZJRg==",
				Detail: "high",
			},
		},
	}
	messages := []types.Message{
		makeMultimodalMessage("user", parts),
	}

	contents, _, err := ConvertMessages(messages)
	if err != nil {
		t.Fatalf("ConvertMessages failed: %v", err)
	}

	if len(contents) != 1 {
		t.Fatalf("Expected 1 content, got %d", len(contents))
	}

	geminiParts := contents[0].Parts
	if len(geminiParts) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(geminiParts))
	}

	// First part should be text
	if geminiParts[0].Text != "What's in this image?" {
		t.Errorf("Text mismatch: %q", geminiParts[0].Text)
	}

	// Second part should be inline image data
	if geminiParts[1].InlineData == nil {
		t.Fatal("Expected InlineData for image")
	}
	if geminiParts[1].InlineData.MimeType != "image/jpeg" {
		t.Errorf("Expected mimeType 'image/jpeg', got %q", geminiParts[1].InlineData.MimeType)
	}
	if geminiParts[1].InlineData.Data != "/9j/4AAQSkZJRg==" {
		t.Error("Base64 data mismatch")
	}
}

func TestConvertMessages_URLImage(t *testing.T) {
	parts := []types.ContentPart{
		{Type: "text", Text: "Describe this image:"},
		{
			Type: "image_url",
			ImageURL: &types.ImageURL{
				URL: "https://example.com/image.png",
			},
		},
	}
	messages := []types.Message{
		makeMultimodalMessage("user", parts),
	}

	contents, _, err := ConvertMessages(messages)
	if err != nil {
		t.Fatalf("ConvertMessages failed: %v", err)
	}

	geminiParts := contents[0].Parts
	if len(geminiParts) != 2 {
		t.Fatalf("Expected 2 parts, got %d", len(geminiParts))
	}

	// Second part should be file data (URL reference)
	if geminiParts[1].FileData == nil {
		t.Fatal("Expected FileData for URL image")
	}
	if geminiParts[1].FileData.FileURI != "https://example.com/image.png" {
		t.Errorf("FileURI mismatch: %q", geminiParts[1].FileData.FileURI)
	}
}

func TestConvertMessages_MultipleImages(t *testing.T) {
	parts := []types.ContentPart{
		{Type: "text", Text: "Compare these images:"},
		{Type: "image_url", ImageURL: &types.ImageURL{URL: "data:image/png;base64,abc123"}},
		{Type: "image_url", ImageURL: &types.ImageURL{URL: "data:image/png;base64,def456"}},
	}
	messages := []types.Message{
		makeMultimodalMessage("user", parts),
	}

	contents, _, err := ConvertMessages(messages)
	if err != nil {
		t.Fatalf("ConvertMessages failed: %v", err)
	}

	geminiParts := contents[0].Parts
	if len(geminiParts) != 3 {
		t.Fatalf("Expected 3 parts (1 text + 2 images), got %d", len(geminiParts))
	}

	// Verify both images are converted
	if geminiParts[1].InlineData == nil || geminiParts[2].InlineData == nil {
		t.Error("Expected both images to have InlineData")
	}
}

// ==================== Parameter Conversion Tests ====================

func TestConvertOpenAIRequest_Parameters(t *testing.T) {
	req := &types.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []types.Message{
			makeTextMessage("user", "Test"),
		},
		Temperature: newFloat64(0.7),
		TopP:        newFloat64(0.9),
		MaxTokens:   newInt(1000),
		Stop:        types.StopSequence{"END", "STOP"},
	}

	geminiReq, err := ConvertOpenAIRequest(req)
	if err != nil {
		t.Fatalf("ConvertOpenAIRequest failed: %v", err)
	}

	if geminiReq.GenerationConfig == nil {
		t.Fatal("Expected GenerationConfig to be set")
	}

	cfg := geminiReq.GenerationConfig

	if cfg.Temperature == nil || *cfg.Temperature != 0.7 {
		t.Errorf("Temperature mismatch: got %v", cfg.Temperature)
	}
	if cfg.TopP == nil || *cfg.TopP != 0.9 {
		t.Errorf("TopP mismatch: got %v", cfg.TopP)
	}
	if cfg.MaxOutputTokens == nil || *cfg.MaxOutputTokens != 1000 {
		t.Errorf("MaxOutputTokens mismatch: got %v", cfg.MaxOutputTokens)
	}
	if len(cfg.StopSequences) != 2 || cfg.StopSequences[0] != "END" || cfg.StopSequences[1] != "STOP" {
		t.Errorf("StopSequences mismatch: got %v", cfg.StopSequences)
	}
}

func TestConvertOpenAIRequest_StopSingleString(t *testing.T) {
	req := &types.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []types.Message{
			makeTextMessage("user", "Test"),
		},
		Stop: types.StopSequence{"END"},
	}

	geminiReq, err := ConvertOpenAIRequest(req)
	if err != nil {
		t.Fatalf("ConvertOpenAIRequest failed: %v", err)
	}

	if geminiReq.GenerationConfig == nil {
		t.Fatal("Expected GenerationConfig")
	}

	if len(geminiReq.GenerationConfig.StopSequences) != 1 || geminiReq.GenerationConfig.StopSequences[0] != "END" {
		t.Error("Single stop sequence conversion failed")
	}
}

// ==================== Response Conversion Tests ====================

func TestConvertGeminiResponse_Normal(t *testing.T) {
	geminiResp := &types.GeminiResponse{
		Candidates: []types.GeminiCandidate{
			{
				Content: &types.GeminiContent{
					Parts: []types.GeminiPart{
						{Text: "Hello! How can I help you?"},
					},
					Role: "model",
				},
				FinishReason: types.GeminiFinishReasonStop,
				Index:        0,
			},
		},
		UsageMetadata: &types.GeminiUsageMetadata{
			PromptTokenCount:     10,
			CandidatesTokenCount: 8,
			TotalTokenCount:      18,
		},
	}

	resp, err := ConvertGeminiResponse(geminiResp, "gpt-4")
	if err != nil {
		t.Fatalf("ConvertGeminiResponse failed: %v", err)
	}

	if resp.Object != "chat.completion" {
		t.Errorf("Expected object 'chat.completion', got %q", resp.Object)
	}

	if resp.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got %q", resp.Model)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(resp.Choices))
	}

	choice := resp.Choices[0]
	if choice.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %q", choice.Message.Role)
	}
	if choice.Message.Content != "Hello! How can I help you?" {
		t.Errorf("Content mismatch: %q", choice.Message.Content)
	}
	if choice.FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got %q", choice.FinishReason)
	}

	// Verify usage
	if resp.Usage.PromptTokens != 10 {
		t.Errorf("PromptTokens mismatch: got %d", resp.Usage.PromptTokens)
	}
	if resp.Usage.CompletionTokens != 8 {
		t.Errorf("CompletionTokens mismatch: got %d", resp.Usage.CompletionTokens)
	}
	if resp.Usage.TotalTokens != 18 {
		t.Errorf("TotalTokens mismatch: got %d", resp.Usage.TotalTokens)
	}
}

func TestConvertGeminiStreamChunk(t *testing.T) {
	chunk := &types.GeminiResponse{
		Candidates: []types.GeminiCandidate{
			{
				Content: &types.GeminiContent{
					Parts: []types.GeminiPart{
						{Text: "Hello"},
					},
					Role: "model",
				},
				Index: 0,
			},
		},
	}

	streamChunk, err := ConvertGeminiStreamChunk(chunk, "gpt-4", 0)
	if err != nil {
		t.Fatalf("ConvertGeminiStreamChunk failed: %v", err)
	}

	if streamChunk.Object != "chat.completion.chunk" {
		t.Errorf("Expected object 'chat.completion.chunk', got %q", streamChunk.Object)
	}

	if len(streamChunk.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(streamChunk.Choices))
	}

	if streamChunk.Choices[0].Delta.Content != "Hello" {
		t.Errorf("Delta content mismatch: %q", streamChunk.Choices[0].Delta.Content)
	}
}

func TestConvertGeminiStreamChunk_WithFinishReason(t *testing.T) {
	chunk := &types.GeminiResponse{
		Candidates: []types.GeminiCandidate{
			{
				Content: &types.GeminiContent{
					Parts: []types.GeminiPart{
						{Text: ""},
					},
					Role: "model",
				},
				FinishReason: types.GeminiFinishReasonStop,
				Index:        0,
			},
		},
	}

	streamChunk, err := ConvertGeminiStreamChunk(chunk, "gpt-4", 5)
	if err != nil {
		t.Fatalf("ConvertGeminiStreamChunk failed: %v", err)
	}

	if streamChunk.Choices[0].FinishReason != "stop" {
		t.Errorf("Expected finish_reason 'stop', got %q", streamChunk.Choices[0].FinishReason)
	}
}

// ==================== Model Mapping Tests ====================

func TestMapModelName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gpt-4", "gemini-1.5-pro-latest"},
		{"gpt-4-turbo", "gemini-1.5-pro-latest"},
		{"gpt-4o", "gemini-1.5-flash-latest"},
		{"gpt-4o-mini", "gemini-1.5-flash-8b-latest"},
		{"gpt-3.5-turbo", "gemini-1.5-flash-latest"},
		// Gemini native names should pass through
		{"gemini-1.5-pro", "gemini-1.5-pro-latest"},
		{"gemini-1.5-flash", "gemini-1.5-flash-latest"},
		{"gemini-2.0-flash", "gemini-2.0-flash"},
		// Unknown models should pass through as-is
		{"unknown-model", "unknown-model"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MapModelName(tt.input)
			if result != tt.expected {
				t.Errorf("MapModelName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== Edge Case Tests ====================

func TestConvertMessages_EmptyList(t *testing.T) {
	_, _, err := ConvertMessages([]types.Message{})
	if err == nil {
		t.Error("Expected error for empty message list")
	}
}

func TestConvertMessages_UnsupportedContentType(t *testing.T) {
	parts := []types.ContentPart{
		{Type: "video", Text: "unsupported"},
	}
	messages := []types.Message{
		makeMultimodalMessage("user", parts),
	}

	_, _, err := ConvertMessages(messages)
	if err == nil {
		t.Error("Expected error for unsupported content type")
	}
}

func TestConvertGeminiResponse_EmptyCandidates(t *testing.T) {
	resp := &types.GeminiResponse{
		Candidates: []types.GeminiCandidate{},
	}

	_, err := ConvertGeminiResponse(resp, "gpt-4")
	if err == nil {
		t.Error("Expected error for empty candidates")
	}
}

func TestConvertGeminiResponse_BlockedContent(t *testing.T) {
	resp := &types.GeminiResponse{
		Candidates: []types.GeminiCandidate{
			{
				FinishReason: types.GeminiFinishReasonSafety,
				Index:        0,
			},
		},
		PromptFeedback: &types.GeminiPromptFeedback{
			BlockReason: "SAFETY",
		},
	}

	result, err := ConvertGeminiResponse(resp, "gpt-4")
	if err != nil {
		t.Fatalf("ConvertGeminiResponse failed: %v", err)
	}

	// Should return with content_filter finish reason
	if result.Choices[0].FinishReason != "content_filter" {
		t.Errorf("Expected finish_reason 'content_filter', got %q", result.Choices[0].FinishReason)
	}
}

// ==================== Finish Reason Mapping Tests ====================

func TestMapFinishReason(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{types.GeminiFinishReasonStop, "stop"},
		{types.GeminiFinishReasonMaxTokens, "length"},
		{types.GeminiFinishReasonSafety, "content_filter"},
		{types.GeminiFinishReasonRecitation, "content_filter"},
		{"", "stop"}, // Empty defaults to stop
		{"UNKNOWN", "stop"}, // Unknown defaults to stop
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := MapFinishReason(tt.input)
			if result != tt.expected {
				t.Errorf("MapFinishReason(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== Benchmark Tests ====================

func BenchmarkConvertOpenAIRequest(b *testing.B) {
	req := &types.ChatCompletionRequest{
		Model: "gpt-4",
		Messages: []types.Message{
			makeTextMessage("system", "You are a helpful assistant."),
			makeTextMessage("user", "Hello!"),
			makeTextMessage("assistant", "Hi there! How can I help you today?"),
			makeTextMessage("user", "Tell me about Go programming."),
		},
		Temperature: newFloat64(0.7),
		TopP:        newFloat64(0.9),
		MaxTokens:   newInt(2000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertOpenAIRequest(req)
	}
}

func BenchmarkConvertGeminiResponse(b *testing.B) {
	resp := &types.GeminiResponse{
		Candidates: []types.GeminiCandidate{
			{
				Content: &types.GeminiContent{
					Parts: []types.GeminiPart{
						{Text: "Go is a statically typed, compiled programming language designed at Google. It's known for simplicity, efficiency, and strong support for concurrent programming."},
					},
					Role: "model",
				},
				FinishReason: types.GeminiFinishReasonStop,
				Index:        0,
			},
		},
		UsageMetadata: &types.GeminiUsageMetadata{
			PromptTokenCount:     50,
			CandidatesTokenCount: 30,
			TotalTokenCount:      80,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ConvertGeminiResponse(resp, "gpt-4")
	}
}

func BenchmarkConvertMessages_Multimodal(b *testing.B) {
	parts := []types.ContentPart{
		{Type: "text", Text: "What's in this image?"},
		{Type: "image_url", ImageURL: &types.ImageURL{URL: "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQABAAD"}},
	}
	messages := []types.Message{
		makeMultimodalMessage("user", parts),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ConvertMessages(messages)
	}
}

// ==================== Integration-like Tests ====================

func TestConvertOpenAIRequest_FullConversion(t *testing.T) {
	req := &types.ChatCompletionRequest{
		Model: "gpt-4-turbo",
		Messages: []types.Message{
			makeTextMessage("system", "You are a coding assistant."),
			makeTextMessage("user", "Write a hello world in Go."),
		},
		Temperature: newFloat64(0.5),
		MaxTokens:   newInt(500),
		Stop:        types.StopSequence{"```"},
	}

	geminiReq, err := ConvertOpenAIRequest(req)
	if err != nil {
		t.Fatalf("ConvertOpenAIRequest failed: %v", err)
	}

	// Verify system instruction
	if geminiReq.SystemInstruction == nil {
		t.Fatal("Expected system instruction")
	}
	if geminiReq.SystemInstruction.Parts[0].Text != "You are a coding assistant." {
		t.Error("System instruction text mismatch")
	}

	// Verify contents (should only have user message)
	if len(geminiReq.Contents) != 1 {
		t.Fatalf("Expected 1 content (user only), got %d", len(geminiReq.Contents))
	}

	// Verify generation config
	if geminiReq.GenerationConfig == nil {
		t.Fatal("Expected generation config")
	}
	if *geminiReq.GenerationConfig.Temperature != 0.5 {
		t.Error("Temperature mismatch")
	}
	if *geminiReq.GenerationConfig.MaxOutputTokens != 500 {
		t.Error("MaxOutputTokens mismatch")
	}
}

func TestGenerateResponseID(t *testing.T) {
	id1 := GenerateResponseID()
	id2 := GenerateResponseID()

	if id1 == id2 {
		t.Error("Generated IDs should be unique")
	}

	// Check prefix
	if len(id1) < 10 || id1[:8] != "chatcmpl" {
		t.Errorf("ID format unexpected: %s", id1)
	}
}

func TestGetCreatedTimestamp(t *testing.T) {
	ts := GetCreatedTimestamp()
	now := time.Now().Unix()

	// Should be within 1 second of now
	if ts < now-1 || ts > now+1 {
		t.Errorf("Timestamp %d not close to now %d", ts, now)
	}
}
