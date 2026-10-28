package providers

import (
	"context"
	"strings"

	"kuorum/backend/internal/domain/retrieval"
)

type ProviderName string

const (
	ProviderOpenAlex        ProviderName = "openalex"
	ProviderSemanticScholar ProviderName = "semanticscholar"
	ProviderArxiv           ProviderName = "arxiv"
	ProviderCrossref        ProviderName = "crossref"
)

// AcademicProvider defines the contract that each academic literature provider must implement.
type AcademicProvider interface {
	Name() ProviderName
	SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error)
	GetPaperDetails(ctx context.Context, id string) (*retrieval.PaperMetadata, error)
	GetReferences(ctx context.Context, id string) ([]retrieval.PaperMetadata, error)
	GetCitations(ctx context.Context, id string) ([]retrieval.PaperMetadata, error)
}

// AllProviders returns the list of all supported academic literature providers.
func AllProviders() []ProviderName {
	return []ProviderName{
		ProviderOpenAlex,
		ProviderSemanticScholar,
		ProviderArxiv,
		ProviderCrossref,
	}
}

// NormalizeProviderName normalizes a string into a known ProviderName if recognized.
func NormalizeProviderName(name string) (ProviderName, bool) {
	clean := strings.ToLower(strings.TrimSpace(name))
	clean = strings.ReplaceAll(clean, "-", "")
	clean = strings.ReplaceAll(clean, "_", "")
	clean = strings.ReplaceAll(clean, " ", "")

	switch clean {
	case "openalex":
		return ProviderOpenAlex, true
	case "semanticscholar", "s2":
		return ProviderSemanticScholar, true
	case "arxiv":
		return ProviderArxiv, true
	case "crossref":
		return ProviderCrossref, true
	default:
		return ProviderName(clean), false
	}
}
