// Package gemini provides format conversion between OpenAI and Gemini API formats.
// This is a pure logic module with no IO operations.
package gemini

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"muxueTools/internal/types"
)

// ==================== Model Mappings ====================

// defaultModelMappings maps OpenAI model names to Gemini model names.
var defaultModelMappings = map[string]string{
	"gpt-4":            "gemini-1.5-pro-latest",
	"gpt-4-turbo":      "gemini-1.5-pro-latest",
	"gpt-4o":           "gemini-1.5-flash-latest",
	"gpt-4o-mini":      "gemini-1.5-flash-8b-latest",
	"gpt-3.5-turbo":    "gemini-1.5-flash-latest",
	"gemini-1.5-pro":   "gemini-1.5-pro-latest",
	"gemini-1.5-flash": "gemini-1.5-flash-latest",
	"gemini-2.0-flash": "gemini-2.0-flash",
}

// ==================== Request Conversion ====================

// ConvertOpenAIRequest converts an OpenAI ChatCompletionRequest to a GeminiRequest.
func ConvertOpenAIRequest(req *types.ChatCompletionRequest) (*types.GeminiRequest, error) {
	if len(req.Messages) == 0 {
		return nil, types.ErrEmptyMessages
	}

	contents, systemInstruction, err := ConvertMessages(req.Messages)
	if err != nil {
		return nil, err
	}

	geminiReq := &types.GeminiRequest{
		Contents:          contents,
		SystemInstruction: systemInstruction,
	}

	// Convert generation config
	geminiReq.GenerationConfig = convertGenerationConfig(req)

	return geminiReq, nil
}

// ConvertMessages converts OpenAI messages to Gemini contents format.
// It returns the converted contents, extracted system instruction (if any), and any error.
func ConvertMessages(messages []types.Message) ([]types.GeminiContent, *types.GeminiContent, error) {
	if len(messages) == 0 {
		return nil, nil, types.ErrEmptyMessages
	}

	var contents []types.GeminiContent
	var systemInstruction *types.GeminiContent

	for _, msg := range messages {
		parts, err := convertMessageToParts(msg)
		if err != nil {
			return nil, nil, err
		}

		// Handle system messages - extract to systemInstruction
		if msg.Role == "system" {
			systemInstruction = &types.GeminiContent{
				Parts: parts,
				// Role is omitted for systemInstruction
			}
			continue
		}

		// Map role: assistant -> model
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		contents = append(contents, types.GeminiContent{
			Parts: parts,
			Role:  role,
		})
	}

	return contents, systemInstruction, nil
}

// convertMessageToParts converts a single OpenAI message to Gemini parts.
func convertMessageToParts(msg types.Message) ([]types.GeminiPart, error) {
	// Try to parse as plain string first
	if text, ok := msg.GetContentAsString(); ok {
		return []types.GeminiPart{
			{Text: text},
		}, nil
	}

	// Try to parse as multimodal content parts
	contentParts, ok := msg.GetContentAsParts()
	if !ok {
		return nil, types.NewInvalidMessagesError("Failed to parse message content")
	}

	var parts []types.GeminiPart
	for _, cp := range contentParts {
		part, err := convertContentPart(cp)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	return parts, nil
}

// convertContentPart converts a single OpenAI ContentPart to a GeminiPart.
func convertContentPart(cp types.ContentPart) (types.GeminiPart, error) {
	switch cp.Type {
	case "text":
		return types.GeminiPart{Text: cp.Text}, nil

	case "image_url":
		if cp.ImageURL == nil {
			return types.GeminiPart{}, types.NewInvalidMessagesError("image_url type requires image_url field")
		}
		return convertImageURL(cp.ImageURL)

	default:
		return types.GeminiPart{}, types.NewInvalidMessagesError("Unsupported content type: " + cp.Type)
	}
}

// convertImageURL converts an OpenAI ImageURL to appropriate Gemini format.
func convertImageURL(img *types.ImageURL) (types.GeminiPart, error) {
	url := img.URL

	// Check if it's a base64 data URI
	if strings.HasPrefix(url, "data:") {
		return parseBase64DataURI(url)
	}

	// It's a regular URL - use FileData
	// Infer mime type from URL if possible
	mimeType := inferMimeTypeFromURL(url)
	return types.GeminiPart{
		FileData: &types.GeminiFileData{
			MimeType: mimeType,
			FileURI:  url,
		},
	}, nil
}

// parseBase64DataURI parses a data URI and returns a GeminiPart with InlineData.
// Format: data:[<mediatype>][;base64],<data>
func parseBase64DataURI(dataURI string) (types.GeminiPart, error) {
	// Remove "data:" prefix
	rest := strings.TrimPrefix(dataURI, "data:")

	// Find the comma separator
	commaIdx := strings.Index(rest, ",")
	if commaIdx == -1 {
		return types.GeminiPart{}, types.NewInvalidMessagesError("Invalid data URI format")
	}

	mediaInfo := rest[:commaIdx]
	data := rest[commaIdx+1:]

	// Parse media type (remove ;base64 suffix if present)
	mimeType := strings.TrimSuffix(mediaInfo, ";base64")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return types.GeminiPart{
		InlineData: &types.GeminiInlineData{
			MimeType: mimeType,
			Data:     data,
		},
	}, nil
}

// inferMimeTypeFromURL tries to infer MIME type from URL extension.
func inferMimeTypeFromURL(url string) string {
	lower := strings.ToLower(url)
	switch {
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	default:
		return "image/jpeg" // Default assumption
	}
}

// convertGenerationConfig converts OpenAI request parameters to Gemini GenerationConfig.
func convertGenerationConfig(req *types.ChatCompletionRequest) *types.GeminiGenerationConfig {
	cfg := &types.GeminiGenerationConfig{}
	hasConfig := false

	if req.Temperature != nil {
		cfg.Temperature = req.Temperature
		hasConfig = true
	}

	if req.TopP != nil {
		cfg.TopP = req.TopP
		hasConfig = true
	}

	if req.MaxTokens != nil {
		cfg.MaxOutputTokens = req.MaxTokens
		hasConfig = true
	}

	if len(req.Stop) > 0 {
		cfg.StopSequences = req.Stop
		hasConfig = true
	}

	if req.N != nil {
		cfg.CandidateCount = req.N
		hasConfig = true
	}

	if !hasConfig {
		return nil
	}

	return cfg
}

// ApplyModelSettings applies global model settings to a GeminiRequest.
// It only applies settings if they are not already set in the request.
// Priority: OpenAI request params > Global settings > Gemini defaults
func ApplyModelSettings(req *types.GeminiRequest, settings *types.ModelSettingsConfig) {
	if settings == nil {
		return
	}

	// Apply System Prompt if not already set
	if req.SystemInstruction == nil && settings.SystemPrompt != "" {
		req.SystemInstruction = &types.GeminiContent{
			Parts: []types.GeminiPart{{Text: settings.SystemPrompt}},
		}
	}

	// Initialize GenerationConfig if needed
	if req.GenerationConfig == nil {
		req.GenerationConfig = &types.GeminiGenerationConfig{}
	}

	// Apply generation parameters if not already set
	if req.GenerationConfig.Temperature == nil && settings.Temperature != nil {
		req.GenerationConfig.Temperature = settings.Temperature
	}

	if req.GenerationConfig.TopP == nil && settings.TopP != nil {
		req.GenerationConfig.TopP = settings.TopP
	}

	if req.GenerationConfig.TopK == nil && settings.TopK != nil {
		req.GenerationConfig.TopK = settings.TopK
	}

	if req.GenerationConfig.MaxOutputTokens == nil && settings.MaxOutputTokens != nil {
		req.GenerationConfig.MaxOutputTokens = settings.MaxOutputTokens
	}

	// Apply Gemini 2.5+ features
	if settings.ThinkingLevel != nil && *settings.ThinkingLevel != "" {
		req.GenerationConfig.ThinkingConfig = &types.ThinkingConfig{
			ThinkingLevel: settings.ThinkingLevel,
		}
	}

	if settings.MediaResolution != nil && *settings.MediaResolution != "" {
		req.GenerationConfig.MediaResolution = settings.MediaResolution
	}
}

// ==================== Response Conversion ====================

// ConvertGeminiResponse converts a Gemini response to OpenAI ChatCompletionResponse format.
func ConvertGeminiResponse(resp *types.GeminiResponse, model string) (*types.ChatCompletionResponse, error) {
	if len(resp.Candidates) == 0 {
		return nil, types.NewUpstreamError("No candidates in Gemini response")
	}

	choices := make([]types.Choice, 0, len(resp.Candidates))
	for _, candidate := range resp.Candidates {
		choice := convertCandidate(candidate)
		choices = append(choices, choice)
	}

	// Build usage
	usage := types.Usage{}
	if resp.UsageMetadata != nil {
		usage = resp.UsageMetadata.ToOpenAIUsage()
	}

	return &types.ChatCompletionResponse{
		ID:      GenerateResponseID(),
		Object:  "chat.completion",
		Created: GetCreatedTimestamp(),
		Model:   model,
		Choices: choices,
		Usage:   usage,
	}, nil
}

// convertCandidate converts a single Gemini candidate to an OpenAI Choice.
func convertCandidate(candidate types.GeminiCandidate) types.Choice {
	content := ""
	if candidate.Content != nil {
		content = candidate.GetTextContent()
	}

	finishReason := MapFinishReason(candidate.FinishReason)

	return types.Choice{
		Index: candidate.Index,
		Message: types.ResponseMessage{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: finishReason,
	}
}

// ConvertGeminiStreamChunk converts a Gemini streaming response chunk to OpenAI format.
func ConvertGeminiStreamChunk(chunk *types.GeminiResponse, model string, index int) (*types.ChatCompletionChunk, error) {
	if len(chunk.Candidates) == 0 {
		// Empty chunk - just return empty delta
		return &types.ChatCompletionChunk{
			ID:      GenerateResponseID(),
			Object:  "chat.completion.chunk",
			Created: GetCreatedTimestamp(),
			Model:   model,
			Choices: []types.ChunkChoice{
				{
					Index: index,
					Delta: types.Delta{},
				},
			},
		}, nil
	}

	choices := make([]types.ChunkChoice, 0, len(chunk.Candidates))
	for _, candidate := range chunk.Candidates {
		content := ""
		if candidate.Content != nil {
			content = candidate.GetTextContent()
		}

		finishReason := ""
		if candidate.FinishReason != "" {
			finishReason = MapFinishReason(candidate.FinishReason)
		}

		choices = append(choices, types.ChunkChoice{
			Index: candidate.Index,
			Delta: types.Delta{
				Content: content,
			},
			FinishReason: finishReason,
		})
	}

	return &types.ChatCompletionChunk{
		ID:      GenerateResponseID(),
		Object:  "chat.completion.chunk",
		Created: GetCreatedTimestamp(),
		Model:   model,
		Choices: choices,
	}, nil
}

// ==================== Model Mapping ====================

// MapModelName maps an OpenAI model name to its Gemini equivalent.
// If no mapping exists, the original name is returned (passthrough).
func MapModelName(openaiModel string) string {
	if geminiModel, ok := defaultModelMappings[openaiModel]; ok {
		return geminiModel
	}
	// Return as-is for unknown models (passthrough)
	return openaiModel
}

// ==================== Finish Reason Mapping ====================

// MapFinishReason converts Gemini finish reason to OpenAI format.
func MapFinishReason(geminiReason string) string {
	switch geminiReason {
	case types.GeminiFinishReasonStop:
		return "stop"
	case types.GeminiFinishReasonMaxTokens:
		return "length"
	case types.GeminiFinishReasonSafety, types.GeminiFinishReasonRecitation:
		return "content_filter"
	default:
		return "stop" // Default to stop
	}
}

// ==================== Helper Functions ====================

// GenerateResponseID generates a unique response ID in OpenAI format.
// Format: chatcmpl-{random_hex}
func GenerateResponseID() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes) // Error is safely ignored for rand.Read
	return "chatcmpl-" + hex.EncodeToString(bytes)
}

// GetCreatedTimestamp returns the current Unix timestamp.
func GetCreatedTimestamp() int64 {
	return time.Now().Unix()
}
