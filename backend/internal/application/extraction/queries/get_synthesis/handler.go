package get_synthesis

import (
	"context"
	"fmt"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetSynthesisQueryHandler struct {
	repo interfaces.IExtractionRepository
}

func NewGetSynthesisQueryHandler(repo interfaces.IExtractionRepository) *GetSynthesisQueryHandler {
	return &GetSynthesisQueryHandler{repo: repo}
}

func (h *GetSynthesisQueryHandler) Handle(ctx context.Context, q GetSynthesisQuery) (*GetSynthesisResult, error) {
	synthesis, err := h.repo.GetSynthesis(ctx, q.AnchorPaperID, q.Concept)
	if err != nil {
		return nil, fmt.Errorf("fetching grounded synthesis failed: %w", err)
	}
	return &GetSynthesisResult{Synthesis: synthesis}, nil
}
