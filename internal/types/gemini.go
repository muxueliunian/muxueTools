// Package types defines all data transfer objects and core types for MxlnAPI.
package types

// ==================== Gemini API Request ====================

// GeminiRequest represents a request to Gemini's generateContent endpoint.
type GeminiRequest struct {
	Contents          []GeminiContent          `json:"contents"`
	SystemInstruction *GeminiContent           `json:"systemInstruction,omitempty"`
	GenerationConfig  *GeminiGenerationConfig  `json:"generationConfig,omitempty"`
	SafetySettings    []GeminiSafetySetting    `json:"safetySettings,omitempty"`
}

// GeminiContent represents a single content block (user/model turn).
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"` // "user" or "model"
}

// GeminiPart represents a single part within a content block.
// Can be text, inline data (image), or file data.
type GeminiPart struct {
	Text       string           `json:"text,omitempty"`
	InlineData *GeminiInlineData `json:"inlineData,omitempty"`
	FileData   *GeminiFileData   `json:"fileData,omitempty"`
}

// GeminiInlineData represents base64-encoded binary data (e.g., images).
type GeminiInlineData struct {
	MimeType string `json:"mimeType"` // e.g., "image/jpeg", "image/png"
	Data     string `json:"data"`     // base64-encoded content
}

// GeminiFileData represents a file reference (for uploaded files).
type GeminiFileData struct {
	MimeType string `json:"mimeType"`
	FileURI  string `json:"fileUri"` // e.g., "gs://..." or uploaded file URI
}

// GeminiGenerationConfig contains generation parameters.
type GeminiGenerationConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
	CandidateCount  *int     `json:"candidateCount,omitempty"`
}

// GeminiSafetySetting configures safety thresholds.
type GeminiSafetySetting struct {
	Category  string `json:"category"`  // e.g., "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	Threshold string `json:"threshold"` // e.g., "BLOCK_NONE", "BLOCK_MEDIUM_AND_ABOVE"
}

// ==================== Gemini API Response ====================

// GeminiResponse represents a response from Gemini's generateContent endpoint.
type GeminiResponse struct {
	Candidates     []GeminiCandidate    `json:"candidates"`
	UsageMetadata  *GeminiUsageMetadata `json:"usageMetadata,omitempty"`
	PromptFeedback *GeminiPromptFeedback `json:"promptFeedback,omitempty"`
}

// GeminiCandidate represents a single generation candidate.
type GeminiCandidate struct {
	Content       *GeminiContent          `json:"content,omitempty"`
	FinishReason  string                  `json:"finishReason,omitempty"` // "STOP", "MAX_TOKENS", "SAFETY", "RECITATION"
	SafetyRatings []GeminiSafetyRating    `json:"safetyRatings,omitempty"`
	Index         int                     `json:"index"`
}

// GeminiSafetyRating provides safety analysis for content.
type GeminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"` // "NEGLIGIBLE", "LOW", "MEDIUM", "HIGH"
	Blocked     bool   `json:"blocked,omitempty"`
}

// GeminiUsageMetadata contains token consumption information.
type GeminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

// GeminiPromptFeedback contains feedback about the prompt.
type GeminiPromptFeedback struct {
	BlockReason   string               `json:"blockReason,omitempty"` // "SAFETY", "OTHER"
	SafetyRatings []GeminiSafetyRating `json:"safetyRatings,omitempty"`
}

// ==================== Gemini Streaming Response ====================

// GeminiStreamChunk represents a single chunk in a streaming response.
// The structure is the same as GeminiResponse but typically contains partial content.
type GeminiStreamChunk = GeminiResponse

// ==================== Gemini Error Response ====================

// GeminiErrorResponse represents an error returned by the Gemini API.
type GeminiErrorResponse struct {
	Error GeminiErrorDetail `json:"error"`
}

// GeminiErrorDetail contains detailed error information.
type GeminiErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"` // e.g., "INVALID_ARGUMENT", "RESOURCE_EXHAUSTED"
}

// ==================== Model List Response ====================

// GeminiModelsResponse represents the response for listing available models.
type GeminiModelsResponse struct {
	Models []GeminiModelInfo `json:"models"`
}

// GeminiModelInfo represents information about a single Gemini model.
type GeminiModelInfo struct {
	Name                       string   `json:"name"`                       // e.g., "models/gemini-pro"
	DisplayName                string   `json:"displayName"`
	Description                string   `json:"description"`
	Version                    string   `json:"version"`
	InputTokenLimit            int      `json:"inputTokenLimit"`
	OutputTokenLimit           int      `json:"outputTokenLimit"`
	SupportedGenerationMethods []string `json:"supportedGenerationMethods"` // ["generateContent", "streamGenerateContent"]
}

// ==================== Constants ====================

// Gemini API endpoints and constants.
const (
	// GeminiBaseURL is the base URL for Google AI Studio (Gemini) API.
	GeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"

	// Gemini finish reasons
	GeminiFinishReasonStop       = "STOP"
	GeminiFinishReasonMaxTokens  = "MAX_TOKENS"
	GeminiFinishReasonSafety     = "SAFETY"
	GeminiFinishReasonRecitation = "RECITATION"

	// Safety categories
	SafetyCategoryHarassment       = "HARM_CATEGORY_HARASSMENT"
	SafetyCategoryHateSpeech       = "HARM_CATEGORY_HATE_SPEECH"
	SafetyCategorySexuallyExplicit = "HARM_CATEGORY_SEXUALLY_EXPLICIT"
	SafetyCategoryDangerousContent = "HARM_CATEGORY_DANGEROUS_CONTENT"

	// Safety thresholds
	SafetyThresholdBlockNone         = "BLOCK_NONE"
	SafetyThresholdBlockLowAndAbove  = "BLOCK_LOW_AND_ABOVE"
	SafetyThresholdBlockMediumAndAbove = "BLOCK_MEDIUM_AND_ABOVE"
	SafetyThresholdBlockHighAndAbove = "BLOCK_ONLY_HIGH"
)

// ==================== Helper Methods ====================

// GetTextContent extracts the text content from a Gemini candidate.
// Returns empty string if no text content is found.
func (c *GeminiCandidate) GetTextContent() string {
	if c.Content == nil {
		return ""
	}
	for _, part := range c.Content.Parts {
		if part.Text != "" {
			return part.Text
		}
	}
	return ""
}

// IsBlocked returns true if the response was blocked for safety reasons.
func (r *GeminiResponse) IsBlocked() bool {
	if r.PromptFeedback != nil && r.PromptFeedback.BlockReason != "" {
		return true
	}
	for _, candidate := range r.Candidates {
		if candidate.FinishReason == GeminiFinishReasonSafety {
			return true
		}
	}
	return false
}

// ToOpenAIUsage converts Gemini usage metadata to OpenAI usage format.
func (u *GeminiUsageMetadata) ToOpenAIUsage() Usage {
	return Usage{
		PromptTokens:     u.PromptTokenCount,
		CompletionTokens: u.CandidatesTokenCount,
		TotalTokens:      u.TotalTokenCount,
	}
}

// NewGeminiTextPart creates a GeminiPart with text content.
func NewGeminiTextPart(text string) GeminiPart {
	return GeminiPart{Text: text}
}

// NewGeminiImagePart creates a GeminiPart with inline image data.
func NewGeminiImagePart(mimeType, base64Data string) GeminiPart {
	return GeminiPart{
		InlineData: &GeminiInlineData{
			MimeType: mimeType,
			Data:     base64Data,
		},
	}
}
