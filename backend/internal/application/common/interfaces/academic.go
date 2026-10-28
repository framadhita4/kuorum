package interfaces

import (
	"context"

	"kuorum/backend/internal/domain/retrieval"
)

type IPaperApi interface {
	SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error)
	SearchPapersWithProviders(ctx context.Context, query string, limit int, providers []string) ([]retrieval.PaperMetadata, error)
	GetPaperDetails(ctx context.Context, paperID string) (*retrieval.PaperMetadata, error)
	GetReferences(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error)
	GetCitations(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error)
	AvailableProviders() []string
}
