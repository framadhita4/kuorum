package interfaces

import (
	"context"

	"kuorum/backend/internal/domain/exploration"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/graph"
	"kuorum/backend/internal/domain/pipeline"
	"kuorum/backend/internal/domain/retrieval"
)

type IPipelineRepository interface {
	SavePipeline(ctx context.Context, p *pipeline.ResearchPipeline) error
	GetPipeline(ctx context.Context, id string) (*pipeline.ResearchPipeline, error)
	ListPipelines(ctx context.Context, limit int) ([]pipeline.ResearchPipeline, error)
	DeletePipeline(ctx context.Context, id string) error
	SaveMessage(ctx context.Context, msg *pipeline.ChatMessage) error
	GetMessages(ctx context.Context, chatID string) ([]pipeline.ChatMessage, error)
	PinChat(ctx context.Context, id string, pinned bool) error
}

type IExplorationRepository interface {
	SaveDecomposition(ctx context.Context, decomp *exploration.ConceptDecomposition) error
	GetDecomposition(ctx context.Context, prompt string) (*exploration.ConceptDecomposition, error)
}

type IPaperRepository interface {
	SavePaper(ctx context.Context, paper *retrieval.PaperMetadata) error
	GetPaper(ctx context.Context, id string) (*retrieval.PaperMetadata, error)
	SearchLocalPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error)
}

type IGraphRepository interface {
	SaveGraph(ctx context.Context, g *graph.CitationGraph) error
	GetGraph(ctx context.Context, anchorPaperID string) (*graph.CitationGraph, error)
	DeleteGraph(ctx context.Context, anchorPaperID string) error
}

type IExtractionRepository interface {
	SaveSynthesis(ctx context.Context, s *extraction.GroundedConceptSynthesis) error
	GetSynthesis(ctx context.Context, anchorPaperID string, concept string) (*extraction.GroundedConceptSynthesis, error)
}
