package providers

import (
	"context"
	"strings"
)

// EndpointType denotes the API protocol used to communicate with the LLM provider.
type EndpointType string

const (
	EndpointOpenAI    EndpointType = "openai"
	EndpointAnthropic EndpointType = "anthropic"
)

// ProviderName represents the unique identifier for a registered LLM provider.
type ProviderName string

const (
	ProviderOpenAI     ProviderName = "openai"
	ProviderAnthropic  ProviderName = "anthropic"
	ProviderOpenRouter ProviderName = "openrouter"
	ProviderGroq       ProviderName = "groq"
	ProviderDeepSeek   ProviderName = "deepseek"
	ProviderOllama     ProviderName = "ollama"
	ProviderTogether   ProviderName = "together"
	ProviderMistral    ProviderName = "mistral"
	ProviderPerplexity ProviderName = "perplexity"
	ProviderGemini     ProviderName = "gemini"
)

// ReasoningEffort specifies reasoning / thinking intensity for supported models.
type ReasoningEffort string

const (
	ReasoningEffortNone   ReasoningEffort = ""
	ReasoningEffortLow    ReasoningEffort = "low"
	ReasoningEffortMedium ReasoningEffort = "medium"
	ReasoningEffortHigh   ReasoningEffort = "high"
)

// ThinkingCapability categorizes whether and how a model supports reasoning/thinking tokens.
type ThinkingCapability string

const (
	// ThinkingNotSupported: Standard models with no thinking tokens (e.g., gpt-4o, claude-3-5-sonnet, llama-3.3).
	ThinkingNotSupported ThinkingCapability = "not_supported"

	// ThinkingOptional: Hybrid models that can run either in standard mode or with thinking enabled (e.g., claude-3-7-sonnet).
	ThinkingOptional ThinkingCapability = "optional"

	// ThinkingRequired: Dedicated reasoning models where thinking is always enabled (e.g., o1, o3-mini, deepseek-r1).
	ThinkingRequired ThinkingCapability = "required"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ResponseFormat specifies structured output constraints.
type ResponseFormat struct {
	Type string `json:"type"` // e.g. "json_object"
}

// ChatRequest encapsulates all parameters needed to execute a completion or stream across providers.
type ChatRequest struct {
	Model           string          `json:"model"`
	Messages        []ChatMessage   `json:"messages"`
	Temperature     *float32        `json:"temperature,omitempty"`
	MaxTokens       int             `json:"max_tokens,omitempty"`
	Stream          bool            `json:"stream,omitempty"`
	ResponseFormat  *ResponseFormat `json:"response_format,omitempty"`
	ReasoningEffort ReasoningEffort `json:"reasoning_effort,omitempty"`
	ThinkingBudget  int             `json:"thinking_budget,omitempty"` // Explicit token budget for thinking models
}

// Usage captures token consumption metrics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse contains the normalized result from an LLM completion.
type ChatResponse struct {
	ID           string `json:"id"`
	Model        string `json:"model"`
	Role         string `json:"role"`
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	Usage        Usage  `json:"usage"`
}

// LLMProvider defines the contract that each LLM provider adapter must implement.
// Mirrored after AcademicProvider in internal/infrastructure/academic/providers.
type LLMProvider interface {
	Name() ProviderName
	EndpointType() EndpointType
	DefaultModel() string
	Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Stream(ctx context.Context, req ChatRequest) (<-chan string, <-chan error)
}

// AllStandardProviders returns a slice of well-known standard provider names.
func AllStandardProviders() []ProviderName {
	return []ProviderName{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderOpenRouter,
		ProviderGroq,
		ProviderDeepSeek,
		ProviderOllama,
		ProviderTogether,
		ProviderMistral,
		ProviderPerplexity,
		ProviderGemini,
	}
}

// AllEndpointTypes returns supported endpoint protocol types.
func AllEndpointTypes() []EndpointType {
	return []EndpointType{
		EndpointOpenAI,
		EndpointAnthropic,
	}
}

// NormalizeEndpointType normalizes an input string to a supported EndpointType.
func NormalizeEndpointType(raw string) (EndpointType, bool) {
	clean := strings.ToLower(strings.TrimSpace(raw))
	switch clean {
	case "openai", "openai-compatible", "open_ai":
		return EndpointOpenAI, true
	case "anthropic", "claude", "anthropic-compatible":
		return EndpointAnthropic, true
	default:
		return EndpointType(clean), false
	}
}

// NormalizeProviderName normalizes a string into a known ProviderName if recognized.
func NormalizeProviderName(name string) (ProviderName, bool) {
	clean := strings.ToLower(strings.TrimSpace(name))
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, "_", "")
	clean = strings.ReplaceAll(clean, " ", "")

	switch clean {
	case "openai":
		return ProviderOpenAI, true
	case "anthropic", "claude":
		return ProviderAnthropic, true
	case "openrouter":
		return ProviderOpenRouter, true
	case "groq":
		return ProviderGroq, true
	case "deepseek":
		return ProviderDeepSeek, true
	case "ollama":
		return ProviderOllama, true
	case "together", "togetherai":
		return ProviderTogether, true
	case "mistral", "mistralai":
		return ProviderMistral, true
	case "perplexity":
		return ProviderPerplexity, true
	case "gemini", "google", "googlegemini":
		return ProviderGemini, true
	default:
		return ProviderName(strings.ToLower(strings.TrimSpace(name))), false
	}
}
