package llm

import (
	"kuorum/backend/internal/infrastructure/config"
	"kuorum/backend/internal/infrastructure/llm/providers"
)

// Client is a type alias to FederatedClient for clean, unified external usage.
// Mirrored after Client in internal/infrastructure/academic/client.go.
type Client = FederatedClient

// NewClient initializes a default federated LLM client with an OpenAI-compatible endpoint.
func NewClient(baseURL, apiKey, model string) *FederatedClient {
	return NewFederatedFromConfig(config.LLMConfig{
		EndpointType: string(providers.EndpointOpenAI),
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Model:        model,
	})
}

// NewClientWithEndpoint initializes an LLM client with a specified endpoint type ("openai" or "anthropic").
func NewClientWithEndpoint(endpointType, baseURL, apiKey, model string) *FederatedClient {
	return NewFederatedFromConfig(config.LLMConfig{
		EndpointType: endpointType,
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Model:        model,
	})
}

// NewFederatedFromConfig sets up the active LLM provider from system configuration,
// and automatically registers all providers that have API keys configured in the environment.
func NewFederatedFromConfig(cfg config.LLMConfig) *FederatedClient {
	catalog := providers.DefaultCatalog()
	fc := NewFederatedClient()

	// 1. Auto-register all providers with valid API keys found in environment variables
	knownProviders := []string{
		"gemini",
		"anthropic",
		"openai",
		"deepseek",
		"groq",
		"openrouter",
		"mistral",
		"together",
		"perplexity",
	}

	for _, pName := range knownProviders {
		if catalog.HasConfiguredKey(pName) {
			apiKey := catalog.ResolveAPIKey(pName, "")
			if p, err := providers.NewProviderFromPreset(pName, apiKey); err == nil {
				fc.RegisterProvider(p)
			}
		}
	}

	// 2. If explicit config has an API key and wasn't registered yet, construct it
	if cfg.APIKey != "" {
		pName, _ := providers.NormalizeProviderName(cfg.EndpointType)
		if pName == "" {
			pName = providers.ProviderOpenAI
		}
		if _, ok := fc.GetProvider(pName); !ok {
			if p, err := providers.NewProviderFromConfig(providers.ProviderConfig{
				Name:            string(pName),
				EndpointType:    cfg.EndpointType,
				BaseURL:         cfg.BaseURL,
				APIKey:          cfg.APIKey,
				Model:           cfg.Model,
				ReasoningEffort: providers.ReasoningEffort(cfg.ReasoningEffort),
			}); err == nil {
				fc.RegisterProvider(p)
			}
		}
	}

	// 3. Set the active provider from config if provided
	if activeName, ok := providers.NormalizeProviderName(cfg.EndpointType); ok {
		_ = fc.SetActiveProvider(activeName)
	}

	return fc
}
