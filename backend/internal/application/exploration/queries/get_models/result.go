package get_models

type ModelDTO struct {
	ID                  string   `json:"id"`
	DisplayName         string   `json:"displayName"`
	Provider            string   `json:"provider"`
	ProviderDisplayName string   `json:"providerDisplayName"`
	Description         string   `json:"description"`
	ContextWindow       int      `json:"contextWindow"`
	MaxOutputTokens     int      `json:"maxOutputTokens"`
	ThinkingCapability  string   `json:"thinkingCapability"`
	SupportedEfforts    []string `json:"supportedEfforts,omitempty"`
	AllowsTemperature   bool     `json:"allowsTemperature"`
	IsDefault           bool     `json:"isDefault"`
}

type ProviderDTO struct {
	ID           string   `json:"id"`
	DisplayName  string   `json:"displayName"`
	Description  string   `json:"description"`
	DefaultModel string   `json:"defaultModel"`
	ModelIDs     []string `json:"modelIds"`
}

type GetModelsResult struct {
	DefaultModelID string        `json:"defaultModelId"`
	Models         []ModelDTO    `json:"models"`
	Providers      []ProviderDTO `json:"providers"`
}
