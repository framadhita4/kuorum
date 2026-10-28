package interfaces

import "context"

type modelContextKeyType struct{}

var modelContextKey = modelContextKeyType{}

// WithModel returns a new context containing the specified model ID.
func WithModel(ctx context.Context, model string) context.Context {
	if model == "" {
		return ctx
	}
	return context.WithValue(ctx, modelContextKey, model)
}

// ModelFromContext extracts the model ID from context if present.
func ModelFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(modelContextKey).(string); ok {
		return val
	}
	return ""
}

type ILLMClient interface {
	// CompleteStructured requests the LLM to return structured JSON adhering to a specified schema
	CompleteStructured(ctx context.Context, systemPrompt, userPrompt string, target any) error

	// Stream streams tokens in real-time through a Go channel
	Stream(ctx context.Context, systemPrompt, userPrompt string) (<-chan string, <-chan error)
}

// ModelCatalogInfo carries metadata describing an LLM model for catalog queries.
type ModelCatalogInfo struct {
	ID                  string
	DisplayName         string
	Provider            string
	ProviderDisplayName string
	Description         string
	ContextWindow       int
	MaxOutputTokens     int
	ThinkingCapability  string
	SupportedEfforts    []string
	AllowsTemperature   bool
	IsDefault           bool
}

// ProviderCatalogInfo carries metadata describing an LLM provider and its models.
type ProviderCatalogInfo struct {
	ID           string
	DisplayName  string
	Description  string
	DefaultModel string
	ModelIDs     []string
}

// IModelCatalog provides model metadata query capabilities across registered LLM providers.
type IModelCatalog interface {
	GetAvailableModels() ([]ModelCatalogInfo, []ProviderCatalogInfo, string)
}
