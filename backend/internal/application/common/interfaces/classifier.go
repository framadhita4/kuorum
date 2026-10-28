package interfaces

import (
	"context"

	"kuorum/backend/internal/domain/retrieval"
)

// ClassificationResult captures rapid model evaluation output for a candidate paper abstract.
type ClassificationResult struct {
	PaperID        string  `json:"paperId"`
	Title          string  `json:"title"`
	RelevanceScore float64 `json:"relevanceScore"` // 0.0 - 1.0 calibrated probability
	IsRelevant     bool    `json:"isRelevant"`
	Confidence     float64 `json:"confidence"`
	Label          string  `json:"label"`     // "directly_relevant", "tangentially_related", "irrelevant"
	ModelUsed      string  `json:"modelUsed"` // e.g. "@cf/cloudflare/clef-flash" or "clef-decision-evaluator"
	Reason         string  `json:"reason,omitempty"`
}

// IRapidClassifier defines the contract for rapid abstract-level relevance classification.
// Designed for high-throughput screening across 100-200 candidate papers without heavy generative LLM overhead.
type IRapidClassifier interface {
	// ClassifyAbstract evaluates a single paper abstract against the research topic.
	ClassifyAbstract(ctx context.Context, topic string, paper retrieval.PaperMetadata) (*ClassificationResult, error)

	// BatchClassify evaluates a collection of candidate papers concurrently.
	BatchClassify(ctx context.Context, topic string, papers []retrieval.PaperMetadata, concurrency int) ([]ClassificationResult, error)
}
