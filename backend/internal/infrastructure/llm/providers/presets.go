package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
)

// ModelSpec defines metadata, context limits, and reasoning capabilities for a specific model.
type ModelSpec struct {
	ID                 string             `json:"id"`
	DisplayName        string             `json:"display_name"`
	Description        string             `json:"description"`
	ContextWindow      int                `json:"context_window"`
	MaxOutputTokens    int                `json:"max_output_tokens"`
	ThinkingCapability ThinkingCapability `json:"thinking_capability"`
	SupportedEfforts   []ReasoningEffort  `json:"supported_efforts,omitempty"`
	DefaultEffort      ReasoningEffort    `json:"default_effort,omitempty"`
	AllowsTemperature  bool               `json:"allows_temperature"`
	DefaultTemperature *float32           `json:"default_temperature,omitempty"`
	DefaultMaxTokens   int                `json:"default_max_tokens"`
}

// ProviderSpec defines configuration, endpoint metadata, environment variable keys,
// and registered models for an LLM provider.
type ProviderSpec struct {
	Name           ProviderName         `json:"name"`
	DisplayName    string               `json:"display_name"`
	EndpointType   EndpointType         `json:"endpoint_type"`
	DefaultBaseURL string               `json:"default_base_url"`
	EnvAPIKeyNames []string             `json:"env_api_key_names"` // Prioritized list of env vars (e.g. ["OPENAI_API_KEY", "LLM_API_KEY"])
	DefaultModelID string               `json:"default_model_id"`
	Models         map[string]ModelSpec `json:"models"`
	Description    string               `json:"description"`
}

// Alias for backwards compatibility
type ProviderPreset = ProviderSpec

// Catalog manages registered provider and model specifications.
type Catalog struct {
	mu        sync.RWMutex
	providers map[ProviderName]ProviderSpec
}

var (
	defaultCatalogInstance *Catalog
	onceDefaultCatalog     sync.Once
)

// DefaultCatalog returns the singleton catalog populated with the newest major industry providers and models.
func DefaultCatalog() *Catalog {
	onceDefaultCatalog.Do(func() {
		defaultCatalogInstance = NewCatalog()
		defaultCatalogInstance.initDefaults()
	})
	return defaultCatalogInstance
}

// NewCatalog creates an empty Catalog.
func NewCatalog() *Catalog {
	return &Catalog{
		providers: make(map[ProviderName]ProviderSpec),
	}
}

func (c *Catalog) initDefaults() {
	stdEfforts := []ReasoningEffort{ReasoningEffortLow, ReasoningEffortMedium, ReasoningEffortHigh}

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderGemini,
		DisplayName:    "Google Gemini",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://generativelanguage.googleapis.com/v1beta/openai",
		EnvAPIKeyNames: []string{"GEMINI_API_KEY", "GOOGLE_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "gemini-2.0-flash",
		Description:    "Google Gemini official OpenAI-compatible endpoint",
		Models: map[string]ModelSpec{
			"gemini-2.0-flash": {
				ID:                 "gemini-2.0-flash",
				DisplayName:        "Gemini 2.0 Flash",
				Description:        "Ultra-fast multimodal model with 1M token context",
				ContextWindow:      1048576,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"gemini-2.0-pro-exp-02-05": {
				ID:                 "gemini-2.0-pro-exp-02-05",
				DisplayName:        "Gemini 2.0 Pro Exp",
				Description:        "Frontier complex reasoning and code generation model with 2M context",
				ContextWindow:      2097152,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"gemini-1.5-pro": {
				ID:                 "gemini-1.5-pro",
				DisplayName:        "Gemini 1.5 Pro",
				Description:        "High-capability multimodal reasoning model with 2M context",
				ContextWindow:      2097152,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"gemini-1.5-flash": {
				ID:                 "gemini-1.5-flash",
				DisplayName:        "Gemini 1.5 Flash",
				Description:        "Lightweight high-throughput model with 1M context",
				ContextWindow:      1048576,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderOpenAI,
		DisplayName:    "OpenAI",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://api.openai.com/v1",
		EnvAPIKeyNames: []string{"OPENAI_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "gpt-4o-mini",
		Description:    "OpenAI official API (GPT-4o, GPT-4.5, o1, o3-mini)",
		Models: map[string]ModelSpec{
			"gpt-4o-mini": {
				ID:                 "gpt-4o-mini",
				DisplayName:        "GPT-4o Mini",
				Description:        "Fast, cost-efficient model for lightweight reasoning, multimodal tasks, and vision",
				ContextWindow:      128000,
				MaxOutputTokens:    16384,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"gpt-4o": {
				ID:                 "gpt-4o",
				DisplayName:        "GPT-4o",
				Description:        "High-intelligence flagship multimodal model for advanced text and vision tasks",
				ContextWindow:      128000,
				MaxOutputTokens:    16384,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"gpt-4.5-preview": {
				ID:                 "gpt-4.5-preview",
				DisplayName:        "GPT-4.5 Preview",
				Description:        "Largest knowledge and broad world intuition research model",
				ContextWindow:      128000,
				MaxOutputTokens:    16384,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"chatgpt-4o-latest": {
				ID:                 "chatgpt-4o-latest",
				DisplayName:        "ChatGPT-4o Latest",
				Description:        "Continuously updated dynamic checkpoint tracking ChatGPT web behavior",
				ContextWindow:      128000,
				MaxOutputTokens:    16384,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"o3-mini": {
				ID:                 "o3-mini",
				DisplayName:        "o3-mini",
				Description:        "High-speed STEM, coding, and mathematical reasoning model with selectable reasoning effort",
				ContextWindow:      200000,
				MaxOutputTokens:    100000,
				ThinkingCapability: ThinkingRequired,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  false, // Reasoning models disallow custom temperature
				DefaultMaxTokens:   16384,
			},
			"o1": {
				ID:                 "o1",
				DisplayName:        "o1",
				Description:        "Deep reasoning frontier model trained with reinforcement learning for complex multi-step problems",
				ContextWindow:      200000,
				MaxOutputTokens:    100000,
				ThinkingCapability: ThinkingRequired,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  false,
				DefaultMaxTokens:   16384,
			},
			"o1-mini": {
				ID:                 "o1-mini",
				DisplayName:        "o1-mini",
				Description:        "Cost-effective reasoning model specialized in STEM and coding",
				ContextWindow:      128000,
				MaxOutputTokens:    65536,
				ThinkingCapability: ThinkingRequired,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  false,
				DefaultMaxTokens:   16384,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderAnthropic,
		DisplayName:    "Anthropic Claude",
		EndpointType:   EndpointAnthropic,
		DefaultBaseURL: "https://api.anthropic.com/v1",
		EnvAPIKeyNames: []string{"ANTHROPIC_API_KEY", "CLAUDE_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "claude-3-7-sonnet-20250219",
		Description:    "Anthropic Claude official Messages API (Claude 3.7 Sonnet, Claude 3.5 Sonnet & Haiku)",
		Models: map[string]ModelSpec{
			"claude-3-7-sonnet-20250219": {
				ID:                 "claude-3-7-sonnet-20250219",
				DisplayName:        "Claude 3.7 Sonnet",
				Description:        "Flagship hybrid reasoning model supporting instant responses or deep extended thinking budgets",
				ContextWindow:      200000,
				MaxOutputTokens:    64000,
				ThinkingCapability: ThinkingOptional,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  true,
				DefaultMaxTokens:   8192,
			},
			"claude-3-7-sonnet-latest": {
				ID:                 "claude-3-7-sonnet-latest",
				DisplayName:        "Claude 3.7 Sonnet (Latest Alias)",
				Description:        "Alias always pointing to the most up-to-date Claude 3.7 Sonnet snapshot",
				ContextWindow:      200000,
				MaxOutputTokens:    64000,
				ThinkingCapability: ThinkingOptional,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  true,
				DefaultMaxTokens:   8192,
			},
			"claude-3-5-sonnet-20241022": {
				ID:                 "claude-3-5-sonnet-20241022",
				DisplayName:        "Claude 3.5 Sonnet (v2)",
				Description:        "Industry-standard coding, analysis, and multi-step agentic instruction following",
				ContextWindow:      200000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"claude-3-5-sonnet-latest": {
				ID:                 "claude-3-5-sonnet-latest",
				DisplayName:        "Claude 3.5 Sonnet (Latest Alias)",
				Description:        "Alias pointing to the current Claude 3.5 Sonnet release",
				ContextWindow:      200000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"claude-3-5-haiku-20241022": {
				ID:                 "claude-3-5-haiku-20241022",
				DisplayName:        "Claude 3.5 Haiku",
				Description:        "High-velocity compact Claude model for fast classification and lightweight processing",
				ContextWindow:      200000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"claude-3-5-haiku-latest": {
				ID:                 "claude-3-5-haiku-latest",
				DisplayName:        "Claude 3.5 Haiku (Latest Alias)",
				Description:        "Alias pointing to current Claude 3.5 Haiku release",
				ContextWindow:      200000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"claude-3-opus-20240229": {
				ID:                 "claude-3-opus-20240229",
				DisplayName:        "Claude 3 Opus",
				Description:        "High-depth model for complex writing and philosophical synthesis",
				ContextWindow:      200000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderGroq,
		DisplayName:    "Groq Cloud",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://api.groq.com/openai/v1",
		EnvAPIKeyNames: []string{"GROQ_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "llama-3.3-70b-versatile",
		Description:    "Groq LPU ultra-low latency inference engine",
		Models: map[string]ModelSpec{
			"llama-3.3-70b-versatile": {
				ID:                 "llama-3.3-70b-versatile",
				DisplayName:        "Llama 3.3 70B Versatile",
				Description:        "Meta 70B parameter model operating at hundreds of tokens/sec",
				ContextWindow:      128000,
				MaxOutputTokens:    32768,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"llama-3.1-8b-instant": {
				ID:                 "llama-3.1-8b-instant",
				DisplayName:        "Llama 3.1 8B Instant",
				Description:        "Ultra-low latency edge model operating at ~750 tokens/sec",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"deepseek-r1-distill-llama-70b": {
				ID:                 "deepseek-r1-distill-llama-70b",
				DisplayName:        "DeepSeek R1 Distill Llama 70B",
				Description:        "Open reasoning model distilled from DeepSeek-R1 running on Groq LPU hardware",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"deepseek-r1-distill-qwen-32b": {
				ID:                 "deepseek-r1-distill-qwen-32b",
				DisplayName:        "DeepSeek R1 Distill Qwen 32B",
				Description:        "DeepSeek-R1 reasoning distilled into Qwen 2.5 32B architecture",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"qwen-2.5-32b": {
				ID:                 "qwen-2.5-32b",
				DisplayName:        "Qwen 2.5 32B",
				Description:        "High-capability open multilingual, coding, and mathematical reasoning model",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"mixtral-8x7b-32768": {
				ID:                 "mixtral-8x7b-32768",
				DisplayName:        "Mixtral 8x7B MoE",
				Description:        "Sparse mixture-of-experts model with 32k context window",
				ContextWindow:      32768,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"gemma2-9b-it": {
				ID:                 "gemma2-9b-it",
				DisplayName:        "Gemma 2 9B Instruct",
				Description:        "Google's lightweight instruction-tuned model",
				ContextWindow:      8192,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   2048,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderDeepSeek,
		DisplayName:    "DeepSeek AI",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://api.deepseek.com/v1",
		EnvAPIKeyNames: []string{"DEEPSEEK_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "deepseek-chat",
		Description:    "DeepSeek Official API (DeepSeek-V3, DeepSeek-R1)",
		Models: map[string]ModelSpec{
			"deepseek-chat": {
				ID:                 "deepseek-chat",
				DisplayName:        "DeepSeek-V3",
				Description:        "General flagship conversation and coding model with 671B MoE architecture",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"deepseek-reasoner": {
				ID:                 "deepseek-reasoner",
				DisplayName:        "DeepSeek-R1",
				Description:        "Frontier reasoning model with reinforced multi-step reasoning tokens",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderOpenRouter,
		DisplayName:    "OpenRouter",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://openrouter.ai/api/v1",
		EnvAPIKeyNames: []string{"OPENROUTER_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "anthropic/claude-3.7-sonnet",
		Description:    "OpenRouter Unified Gateway routing to hundreds of frontier models",
		Models: map[string]ModelSpec{
			"anthropic/claude-3.7-sonnet": {
				ID:                 "anthropic/claude-3.7-sonnet",
				DisplayName:        "Claude 3.7 Sonnet (via OpenRouter)",
				Description:        "Anthropic Claude 3.7 Sonnet with hybrid reasoning through OpenRouter",
				ContextWindow:      200000,
				MaxOutputTokens:    64000,
				ThinkingCapability: ThinkingOptional,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  true,
				DefaultMaxTokens:   8192,
			},
			"anthropic/claude-3.5-sonnet": {
				ID:                 "anthropic/claude-3.5-sonnet",
				DisplayName:        "Claude 3.5 Sonnet (via OpenRouter)",
				ContextWindow:      200000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"openai/o3-mini": {
				ID:                 "openai/o3-mini",
				DisplayName:        "o3-mini (via OpenRouter)",
				ContextWindow:      200000,
				MaxOutputTokens:    100000,
				ThinkingCapability: ThinkingRequired,
				SupportedEfforts:   stdEfforts,
				DefaultEffort:      ReasoningEffortMedium,
				AllowsTemperature:  false,
				DefaultMaxTokens:   16384,
			},
			"openai/gpt-4o": {
				ID:                 "openai/gpt-4o",
				DisplayName:        "GPT-4o (via OpenRouter)",
				ContextWindow:      128000,
				MaxOutputTokens:    16384,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"openai/gpt-4.5-preview": {
				ID:                 "openai/gpt-4.5-preview",
				DisplayName:        "GPT-4.5 Preview (via OpenRouter)",
				ContextWindow:      128000,
				MaxOutputTokens:    16384,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"deepseek/deepseek-r1": {
				ID:                 "deepseek/deepseek-r1",
				DisplayName:        "DeepSeek-R1 (via OpenRouter)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"deepseek/deepseek-chat": {
				ID:                 "deepseek/deepseek-chat",
				DisplayName:        "DeepSeek-V3 (via OpenRouter)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"meta-llama/llama-3.3-70b-instruct": {
				ID:                 "meta-llama/llama-3.3-70b-instruct",
				DisplayName:        "Llama 3.3 70B Instruct (via OpenRouter)",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"google/gemini-2.0-flash-001": {
				ID:                 "google/gemini-2.0-flash-001",
				DisplayName:        "Google Gemini 2.0 Flash (via OpenRouter)",
				Description:        "Ultra-fast multimodal model with 1M token context",
				ContextWindow:      1048576,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"google/gemini-2.0-pro-exp-02-05": {
				ID:                 "google/gemini-2.0-pro-exp-02-05",
				DisplayName:        "Google Gemini 2.0 Pro Exp (via OpenRouter)",
				Description:        "Frontier complex reasoning and code generation model with 2M context",
				ContextWindow:      2097152,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderOllama,
		DisplayName:    "Ollama (Local)",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "http://localhost:11434/v1",
		EnvAPIKeyNames: []string{"OLLAMA_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "llama3.2",
		Description:    "Local Ollama Inference Server",
		Models: map[string]ModelSpec{
			"llama3.3": {
				ID:                 "llama3.3",
				DisplayName:        "Llama 3.3 70B (Local)",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"llama3.2": {
				ID:                 "llama3.2",
				DisplayName:        "Llama 3.2 3B (Local)",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   2048,
			},
			"deepseek-r1:70b": {
				ID:                 "deepseek-r1:70b",
				DisplayName:        "DeepSeek-R1 70B (Local)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"deepseek-r1:32b": {
				ID:                 "deepseek-r1:32b",
				DisplayName:        "DeepSeek-R1 32B (Local)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"deepseek-r1:14b": {
				ID:                 "deepseek-r1:14b",
				DisplayName:        "DeepSeek-R1 14B (Local)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"deepseek-r1:8b": {
				ID:                 "deepseek-r1:8b",
				DisplayName:        "DeepSeek-R1 8B (Local)",
				ContextWindow:      32000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   4096,
			},
			"qwen2.5:72b": {
				ID:                 "qwen2.5:72b",
				DisplayName:        "Qwen 2.5 72B (Local)",
				ContextWindow:      32768,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"qwen2.5:32b": {
				ID:                 "qwen2.5:32b",
				DisplayName:        "Qwen 2.5 32B (Local)",
				ContextWindow:      32768,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"qwen2.5-coder:32b": {
				ID:                 "qwen2.5-coder:32b",
				DisplayName:        "Qwen 2.5 Coder 32B (Local)",
				ContextWindow:      32768,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"phi4": {
				ID:                 "phi4",
				DisplayName:        "Phi-4 14B (Local)",
				ContextWindow:      16384,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"mistral-small:24b": {
				ID:                 "mistral-small:24b",
				DisplayName:        "Mistral Small 24B (Local)",
				ContextWindow:      32768,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderTogether,
		DisplayName:    "Together AI",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://api.together.xyz/v1",
		EnvAPIKeyNames: []string{"TOGETHER_API_KEY", "TOGETHERAI_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "meta-llama/Llama-3.3-70B-Instruct-Turbo",
		Description:    "Together AI Open-Source Cloud Inference",
		Models: map[string]ModelSpec{
			"meta-llama/Llama-3.3-70B-Instruct-Turbo": {
				ID:                 "meta-llama/Llama-3.3-70B-Instruct-Turbo",
				DisplayName:        "Llama 3.3 70B Instruct Turbo",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo": {
				ID:                 "meta-llama/Meta-Llama-3.1-405B-Instruct-Turbo",
				DisplayName:        "Llama 3.1 405B Instruct Turbo",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo": {
				ID:                 "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo",
				DisplayName:        "Llama 3.1 70B Instruct Turbo",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo": {
				ID:                 "meta-llama/Meta-Llama-3.1-8B-Instruct-Turbo",
				DisplayName:        "Llama 3.1 8B Instruct Turbo",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"deepseek-ai/DeepSeek-R1": {
				ID:                 "deepseek-ai/DeepSeek-R1",
				DisplayName:        "DeepSeek-R1 Full (Together)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"deepseek-ai/DeepSeek-V3": {
				ID:                 "deepseek-ai/DeepSeek-V3",
				DisplayName:        "DeepSeek-V3 (Together)",
				ContextWindow:      64000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"Qwen/Qwen2.5-72B-Instruct-Turbo": {
				ID:                 "Qwen/Qwen2.5-72B-Instruct-Turbo",
				DisplayName:        "Qwen 2.5 72B Turbo (Together)",
				ContextWindow:      32768,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"Qwen/Qwen2.5-Coder-32B-Instruct": {
				ID:                 "Qwen/Qwen2.5-Coder-32B-Instruct",
				DisplayName:        "Qwen 2.5 Coder 32B (Together)",
				ContextWindow:      32768,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderMistral,
		DisplayName:    "Mistral AI",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://api.mistral.ai/v1",
		EnvAPIKeyNames: []string{"MISTRAL_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "mistral-large-latest",
		Description:    "Mistral AI Official API",
		Models: map[string]ModelSpec{
			"mistral-large-latest": {
				ID:                 "mistral-large-latest",
				DisplayName:        "Mistral Large (2411)",
				Description:        "Frontier 675B parameter sparse MoE model with strong reasoning and agentic capabilities",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"mistral-small-latest": {
				ID:                 "mistral-small-latest",
				DisplayName:        "Mistral Small (24B)",
				Description:        "Efficient enterprise model unifying instruct, reasoning, and coding",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"codestral-latest": {
				ID:                 "codestral-latest",
				DisplayName:        "Codestral (2501)",
				Description:        "Specialized coding, autocomplete, and software development model with 256k context",
				ContextWindow:      256000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"pixtral-large-latest": {
				ID:                 "pixtral-large-latest",
				DisplayName:        "Pixtral Large",
				Description:        "Frontier multimodal vision-language model for image understanding and document OCR",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"ministral-8b-latest": {
				ID:                 "ministral-8b-latest",
				DisplayName:        "Ministral 8B",
				Description:        "Compact high-efficiency model built for edge inference and high-throughput workloads",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   2048,
			},
			"ministral-3b-latest": {
				ID:                 "ministral-3b-latest",
				DisplayName:        "Ministral 3B",
				Description:        "Ultra-fast lightweight edge model for instant responses",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   2048,
			},
		},
	})

	c.RegisterProvider(ProviderSpec{
		Name:           ProviderPerplexity,
		DisplayName:    "Perplexity AI",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: "https://api.perplexity.ai",
		EnvAPIKeyNames: []string{"PERPLEXITY_API_KEY", "PPLX_API_KEY", "LLM_API_KEY"},
		DefaultModelID: "sonar-pro",
		Description:    "Perplexity Online Retrieval & Search Models",
		Models: map[string]ModelSpec{
			"sonar-pro": {
				ID:                 "sonar-pro",
				DisplayName:        "Sonar Pro",
				Description:        "Advanced search model with superior factuality, broader context handling, and deep web summarization",
				ContextWindow:      200000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   4096,
			},
			"sonar": {
				ID:                 "sonar",
				DisplayName:        "Sonar",
				Description:        "Fast, cost-effective search-grounded text generation model",
				ContextWindow:      128000,
				MaxOutputTokens:    4096,
				ThinkingCapability: ThinkingNotSupported,
				AllowsTemperature:  true,
				DefaultMaxTokens:   2048,
			},
			"sonar-reasoning-pro": {
				ID:                 "sonar-reasoning-pro",
				DisplayName:        "Sonar Reasoning Pro",
				Description:        "Heavy reasoning model for complex problem-solving and multi-source synthesis backed by live web data",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"sonar-reasoning": {
				ID:                 "sonar-reasoning",
				DisplayName:        "Sonar Reasoning",
				Description:        "Integrates real-time web search capabilities with explicit step-by-step reasoning processes",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
			"sonar-deep-research": {
				ID:                 "sonar-deep-research",
				DisplayName:        "Sonar Deep Research",
				Description:        "Intensive research-tier model designed for comprehensive multi-step agentic deep dives",
				ContextWindow:      128000,
				MaxOutputTokens:    8192,
				ThinkingCapability: ThinkingRequired,
				AllowsTemperature:  false,
				DefaultMaxTokens:   8192,
			},
		},
	})
}

// RegisterProvider registers or updates a ProviderSpec in the catalog.
func (c *Catalog) RegisterProvider(spec ProviderSpec) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if spec.Models == nil {
		spec.Models = make(map[string]ModelSpec)
	}
	c.providers[spec.Name] = spec
}

// RegisterModel registers a ModelSpec for a specific provider.
func (c *Catalog) RegisterModel(providerName ProviderName, model ModelSpec) {
	c.mu.Lock()
	defer c.mu.Unlock()

	norm, _ := NormalizeProviderName(string(providerName))
	p, ok := c.providers[norm]
	if !ok {
		p, ok = c.providers[providerName]
	}
	if ok {
		if p.Models == nil {
			p.Models = make(map[string]ModelSpec)
		}
		p.Models[model.ID] = model
		c.providers[p.Name] = p
	}
}

// GetProviderSpec retrieves a provider specification by name.
func (c *Catalog) GetProviderSpec(name string) (ProviderSpec, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	norm, _ := NormalizeProviderName(name)
	spec, ok := c.providers[norm]
	if ok {
		return spec, true
	}
	spec, ok = c.providers[ProviderName(strings.ToLower(strings.TrimSpace(name)))]
	return spec, ok
}

// GetModelSpec retrieves a specific model specification for a provider.
func (c *Catalog) GetModelSpec(providerName, modelID string) (ModelSpec, bool) {
	pSpec, ok := c.GetProviderSpec(providerName)
	if !ok {
		return ModelSpec{}, false
	}
	cleanID := strings.TrimSpace(modelID)
	mSpec, found := pSpec.Models[cleanID]
	return mSpec, found
}

// FindModelSpec searches across all registered providers to find a matching model specification.
func (c *Catalog) FindModelSpec(modelID string) (ModelSpec, ProviderSpec, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cleanID := strings.TrimSpace(modelID)
	for _, pSpec := range c.providers {
		if mSpec, found := pSpec.Models[cleanID]; found {
			return mSpec, pSpec, true
		}
	}
	return ModelSpec{}, ProviderSpec{}, false
}

// ListProviders returns all registered providers in the catalog.
func (c *Catalog) ListProviders() []ProviderSpec {
	c.mu.RLock()
	defer c.mu.RUnlock()

	res := make([]ProviderSpec, 0, len(c.providers))
	for _, p := range c.providers {
		res = append(res, p)
	}
	return res
}

// ListModels returns all models registered under a provider.
func (c *Catalog) ListModels(providerName string) []ModelSpec {
	pSpec, ok := c.GetProviderSpec(providerName)
	if !ok {
		return nil
	}
	res := make([]ModelSpec, 0, len(pSpec.Models))
	for _, m := range pSpec.Models {
		res = append(res, m)
	}
	return res
}

// HasConfiguredKey checks whether a valid, non-empty API key is set for the provider.
func (c *Catalog) HasConfiguredKey(providerName string) bool {
	spec, ok := c.GetProviderSpec(providerName)
	if !ok {
		return false
	}

	// 1. Check provider-specific environment variables (excluding general LLM_API_KEY)
	for _, envName := range spec.EnvAPIKeyNames {
		if envName == "LLM_API_KEY" {
			continue
		}
		if val := strings.TrimSpace(os.Getenv(envName)); val != "" {
			return true
		}
	}

	// 2. Check general LLM_API_KEY if configured to map to this provider
	if generalKey := strings.TrimSpace(os.Getenv("LLM_API_KEY")); generalKey != "" {
		endpointType := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_ENDPOINT_TYPE")))
		if endpointType == "" {
			endpointType = "openai"
		}
		if strings.EqualFold(string(spec.Name), endpointType) {
			return true
		}
	}

	return false
}

// GetAvailableModels implements interfaces.IModelCatalog to provide all models and providers that have configured API keys.
func (c *Catalog) GetAvailableModels() ([]interfaces.ModelCatalogInfo, []interfaces.ProviderCatalogInfo, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var models []interfaces.ModelCatalogInfo
	var providerList []interfaces.ProviderCatalogInfo

	priorityProviders := []ProviderName{
		ProviderGemini,
		ProviderAnthropic,
		ProviderOpenAI,
		ProviderDeepSeek,
		ProviderGroq,
		ProviderOpenRouter,
		ProviderMistral,
		ProviderTogether,
		ProviderPerplexity,
	}

	seen := make(map[ProviderName]bool)

	for _, pName := range priorityProviders {
		p, exists := c.providers[pName]
		if !exists {
			continue
		}
		seen[pName] = true

		if !c.HasConfiguredKey(string(pName)) {
			continue
		}

		var modelIDs []string
		for _, m := range p.Models {
			modelIDs = append(modelIDs, m.ID)
			var efforts []string
			for _, eff := range m.SupportedEfforts {
				efforts = append(efforts, string(eff))
			}

			models = append(models, interfaces.ModelCatalogInfo{
				ID:                  m.ID,
				DisplayName:         m.DisplayName,
				Provider:            string(p.Name),
				ProviderDisplayName: p.DisplayName,
				Description:         m.Description,
				ContextWindow:       m.ContextWindow,
				MaxOutputTokens:     m.MaxOutputTokens,
				ThinkingCapability:  string(m.ThinkingCapability),
				SupportedEfforts:    efforts,
				AllowsTemperature:   m.AllowsTemperature,
				IsDefault:           m.ID == p.DefaultModelID,
			})
		}

		providerList = append(providerList, interfaces.ProviderCatalogInfo{
			ID:           string(p.Name),
			DisplayName:  p.DisplayName,
			Description:  p.Description,
			DefaultModel: p.DefaultModelID,
			ModelIDs:     modelIDs,
		})
	}

	for pName, p := range c.providers {
		if seen[pName] {
			continue
		}
		if !c.HasConfiguredKey(string(pName)) {
			continue
		}

		var modelIDs []string
		for _, m := range p.Models {
			modelIDs = append(modelIDs, m.ID)
			var efforts []string
			for _, eff := range m.SupportedEfforts {
				efforts = append(efforts, string(eff))
			}

			models = append(models, interfaces.ModelCatalogInfo{
				ID:                  m.ID,
				DisplayName:         m.DisplayName,
				Provider:            string(p.Name),
				ProviderDisplayName: p.DisplayName,
				Description:         m.Description,
				ContextWindow:       m.ContextWindow,
				MaxOutputTokens:     m.MaxOutputTokens,
				ThinkingCapability:  string(m.ThinkingCapability),
				SupportedEfforts:    efforts,
				AllowsTemperature:   m.AllowsTemperature,
				IsDefault:           m.ID == p.DefaultModelID,
			})
		}

		providerList = append(providerList, interfaces.ProviderCatalogInfo{
			ID:           string(p.Name),
			DisplayName:  p.DisplayName,
			Description:  p.Description,
			DefaultModel: p.DefaultModelID,
			ModelIDs:     modelIDs,
		})
	}

	defaultModel := ""
	if len(models) > 0 {
		for _, m := range models {
			if m.IsDefault {
				defaultModel = m.ID
				break
			}
		}
		if defaultModel == "" {
			defaultModel = models[0].ID
		}
	}

	return models, providerList, defaultModel
}

// ResolveAPIKey resolves the API key for a provider by checking:
// 1. Explicit key argument passed
// 2. Provider-specific environment variable names in priority order
// 3. General fallback environment variable (LLM_API_KEY)
func (c *Catalog) ResolveAPIKey(providerName string, explicitKey string) string {
	if strings.TrimSpace(explicitKey) != "" {
		return strings.TrimSpace(explicitKey)
	}

	pSpec, ok := c.GetProviderSpec(providerName)
	if ok {
		for _, envName := range pSpec.EnvAPIKeyNames {
			if val := os.Getenv(envName); strings.TrimSpace(val) != "" {
				return strings.TrimSpace(val)
			}
		}
	}

	return strings.TrimSpace(os.Getenv("LLM_API_KEY"))
}

// FetchLiveModels connects directly to the provider's /models endpoint to dynamically retrieve available model IDs.
func (c *Catalog) FetchLiveModels(ctx context.Context, providerName string, explicitKey string) ([]string, error) {
	spec, ok := c.GetProviderSpec(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %s is not registered in catalog", providerName)
	}

	apiKey := c.ResolveAPIKey(providerName, explicitKey)
	baseURL := strings.TrimRight(spec.DefaultBaseURL, "/")

	var url string
	if spec.EndpointType == EndpointAnthropic {
		if strings.HasSuffix(baseURL, "/v1") {
			url = baseURL + "/models"
		} else {
			url = baseURL + "/v1/models"
		}
	} else {
		if strings.HasSuffix(baseURL, "/v1") {
			url = baseURL + "/models"
		} else {
			url = baseURL + "/v1/models"
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create /models request failed: %w", err)
	}

	if spec.EndpointType == EndpointAnthropic {
		if apiKey != "" {
			req.Header.Set("x-api-key", apiKey)
		}
		req.Header.Set("anthropic-version", "2023-06-01")
	} else {
		if apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("provider returned status %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("decode /models response failed: %w", err)
	}

	modelIDs := make([]string, 0, len(parsed.Data))
	for _, m := range parsed.Data {
		if m.ID != "" {
			modelIDs = append(modelIDs, m.ID)
		}
	}

	return modelIDs, nil
}

// Package-level functions delegating to DefaultCatalog

func GetPreset(name string) (ProviderPreset, bool) {
	return DefaultCatalog().GetProviderSpec(name)
}

func AvailablePresets() []ProviderPreset {
	return DefaultCatalog().ListProviders()
}

func GetModelSpec(providerName, modelID string) (ModelSpec, bool) {
	return DefaultCatalog().GetModelSpec(providerName, modelID)
}

func FindModelSpec(modelID string) (ModelSpec, ProviderSpec, bool) {
	return DefaultCatalog().FindModelSpec(modelID)
}

func ResolveAPIKey(providerName string, explicitKey string) string {
	return DefaultCatalog().ResolveAPIKey(providerName, explicitKey)
}

func FetchLiveModels(ctx context.Context, providerName string, explicitKey string) ([]string, error) {
	return DefaultCatalog().FetchLiveModels(ctx, providerName, explicitKey)
}

// ProviderConfig specifies the parameters to construct an arbitrary LLM provider.
type ProviderConfig struct {
	Name            string          `json:"name"`
	EndpointType    string          `json:"endpoint_type"` // "openai" or "anthropic"
	BaseURL         string          `json:"base_url"`
	APIKey          string          `json:"api_key"`
	Model           string          `json:"model"`
	ReasoningEffort ReasoningEffort `json:"reasoning_effort,omitempty"`
	HTTPClient      *http.Client    `json:"-"`
}

// NewProviderFromPreset instantiates an LLMProvider using a registered preset.
func NewProviderFromPreset(presetName string, apiKey string, modelOverride ...string) (LLMProvider, error) {
	spec, ok := DefaultCatalog().GetProviderSpec(presetName)
	if !ok {
		return nil, fmt.Errorf("unknown provider preset: %s", presetName)
	}

	resolvedKey := DefaultCatalog().ResolveAPIKey(presetName, apiKey)

	model := spec.DefaultModelID
	if len(modelOverride) > 0 && modelOverride[0] != "" {
		model = modelOverride[0]
	}

	switch spec.EndpointType {
	case EndpointAnthropic:
		return NewAnthropicProvider(spec.Name, spec.DefaultBaseURL, resolvedKey, model, nil), nil
	case EndpointOpenAI:
		fallthrough
	default:
		return NewOpenAIProvider(spec.Name, spec.DefaultBaseURL, resolvedKey, model, nil), nil
	}
}

// NewProviderFromConfig constructs either an OpenAI or Anthropic provider from configuration.
func NewProviderFromConfig(cfg ProviderConfig) (LLMProvider, error) {
	endpointType, _ := NormalizeEndpointType(cfg.EndpointType)
	if endpointType == "" {
		endpointType = EndpointOpenAI
	}

	name := ProviderName(strings.ToLower(strings.TrimSpace(cfg.Name)))
	if name == "" {
		if endpointType == EndpointAnthropic {
			name = ProviderAnthropic
		} else {
			name = ProviderOpenAI
		}
	}

	resolvedKey := DefaultCatalog().ResolveAPIKey(string(name), cfg.APIKey)

	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		if spec, ok := DefaultCatalog().GetProviderSpec(string(name)); ok {
			baseURL = spec.DefaultBaseURL
		} else if endpointType == EndpointAnthropic {
			baseURL = "https://api.anthropic.com/v1"
		} else {
			baseURL = "https://api.openai.com/v1"
		}
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		if spec, ok := DefaultCatalog().GetProviderSpec(string(name)); ok {
			model = spec.DefaultModelID
		} else if endpointType == EndpointAnthropic {
			model = "claude-3-7-sonnet-20250219"
		} else {
			model = "gpt-4o-mini"
		}
	}

	switch endpointType {
	case EndpointAnthropic:
		p := NewAnthropicProvider(name, baseURL, resolvedKey, model, cfg.HTTPClient)
		if cfg.ReasoningEffort != "" {
			p.SetDefaultReasoningEffort(cfg.ReasoningEffort)
		}
		return p, nil
	case EndpointOpenAI:
		fallthrough
	default:
		p := NewOpenAIProvider(name, baseURL, resolvedKey, model, cfg.HTTPClient)
		if cfg.ReasoningEffort != "" {
			p.SetDefaultReasoningEffort(cfg.ReasoningEffort)
		}
		return p, nil
	}
}
