package llm

import (
	"kuorum/backend/internal/infrastructure/llm/providers"
)

// Re-export core types and interfaces from providers package for convenience and backwards compatibility.
type (
	ProviderName       = providers.ProviderName
	EndpointType       = providers.EndpointType
	ReasoningEffort    = providers.ReasoningEffort
	ThinkingCapability = providers.ThinkingCapability
	LLMProvider        = providers.LLMProvider
	BaseProvider       = providers.BaseProvider
	OpenAIProvider     = providers.OpenAIProvider
	AnthropicProvider  = providers.AnthropicProvider
	RateLimiter        = providers.RateLimiter

	ChatMessage    = providers.ChatMessage
	ResponseFormat = providers.ResponseFormat
	ChatRequest    = providers.ChatRequest
	ChatResponse   = providers.ChatResponse
	Usage          = providers.Usage
	ModelSpec      = providers.ModelSpec
	ProviderSpec   = providers.ProviderSpec
	ProviderPreset = providers.ProviderPreset
	ProviderConfig = providers.ProviderConfig
	Catalog        = providers.Catalog
)

const (
	EndpointOpenAI    = providers.EndpointOpenAI
	EndpointAnthropic = providers.EndpointAnthropic

	ProviderOpenAI     = providers.ProviderOpenAI
	ProviderAnthropic  = providers.ProviderAnthropic
	ProviderOpenRouter = providers.ProviderOpenRouter
	ProviderGroq       = providers.ProviderGroq
	ProviderDeepSeek   = providers.ProviderDeepSeek
	ProviderOllama     = providers.ProviderOllama
	ProviderTogether   = providers.ProviderTogether
	ProviderMistral    = providers.ProviderMistral
	ProviderPerplexity = providers.ProviderPerplexity

	ReasoningEffortNone   = providers.ReasoningEffortNone
	ReasoningEffortLow    = providers.ReasoningEffortLow
	ReasoningEffortMedium = providers.ReasoningEffortMedium
	ReasoningEffortHigh   = providers.ReasoningEffortHigh

	ThinkingNotSupported = providers.ThinkingNotSupported
	ThinkingOptional     = providers.ThinkingOptional
	ThinkingRequired     = providers.ThinkingRequired

	RoleSystem    = providers.RoleSystem
	RoleUser      = providers.RoleUser
	RoleAssistant = providers.RoleAssistant
)

var (
	AllStandardProviders  = providers.AllStandardProviders
	AllEndpointTypes      = providers.AllEndpointTypes
	NormalizeEndpointType = providers.NormalizeEndpointType
	NormalizeProviderName = providers.NormalizeProviderName

	NewRateLimiter        = providers.NewRateLimiter
	NewBaseProvider       = providers.NewBaseProvider
	NewOpenAIProvider     = providers.NewOpenAIProvider
	NewAnthropicProvider  = providers.NewAnthropicProvider
	NewProviderFromConfig = providers.NewProviderFromConfig
	NewProviderFromPreset = providers.NewProviderFromPreset
	GetPreset             = providers.GetPreset
	AvailablePresets      = providers.AvailablePresets

	DefaultCatalog  = providers.DefaultCatalog
	GetModelSpec    = providers.GetModelSpec
	FindModelSpec   = providers.FindModelSpec
	ResolveAPIKey   = providers.ResolveAPIKey
	FetchLiveModels = providers.FetchLiveModels
)
