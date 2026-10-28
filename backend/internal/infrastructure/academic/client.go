package academic

import (
	"kuorum/backend/internal/infrastructure/academic/providers"
	"kuorum/backend/internal/infrastructure/config"
)

// Client is a type alias to FederatedClient for clean, unified external usage.
type Client = FederatedClient

// NewClient initializes a default federated academic client with Semantic Scholar API key.
// It sets up all 4 academic providers (OpenAlex, Semantic Scholar, arXiv, Crossref).
func NewClient(apiKey string) *FederatedClient {
	return NewFederatedFromConfig(config.AcademicConfig{
		SemanticScholarAPIKey: apiKey,
		PoliteEmail:           "kuorum@example.com",
	})
}

// NewFederatedFromConfig sets up all 4 academic providers from system configuration.
func NewFederatedFromConfig(cfg config.AcademicConfig) *FederatedClient {
	politeEmail := cfg.PoliteEmail
	if politeEmail == "" {
		politeEmail = "kuorum@example.com"
	}

	openAlex := providers.NewOpenAlexProvider(cfg.OpenAlexBaseURL, politeEmail, nil)
	s2 := providers.NewSemanticScholarProvider(cfg.SemanticScholarBaseURL, cfg.SemanticScholarAPIKey, nil)
	arxiv := providers.NewArxivProvider(cfg.ArxivBaseURL, politeEmail, nil)
	crossref := providers.NewCrossrefProvider(cfg.CrossrefBaseURL, politeEmail, nil)

	return NewFederatedClient(openAlex, s2, arxiv, crossref)
}
