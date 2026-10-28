package get_paper_details

import (
	"context"
	"fmt"
	"strings"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

type GetPaperDetailsQueryHandler struct {
	paperApi interfaces.IPaperApi
	repo     interfaces.IPaperRepository
}

func NewGetPaperDetailsQueryHandler(paperApi interfaces.IPaperApi, repo interfaces.IPaperRepository) *GetPaperDetailsQueryHandler {
	return &GetPaperDetailsQueryHandler{
		paperApi: paperApi,
		repo:     repo,
	}
}

func (h *GetPaperDetailsQueryHandler) Handle(ctx context.Context, q GetPaperDetailsQuery) (*retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(q.ID)

	// Check local repository cache first if available
	if h.repo != nil {
		if cached, err := h.repo.GetPaper(ctx, cleanID); err == nil && cached != nil {
			return cached, nil
		}
	}

	paper, err := h.paperApi.GetPaperDetails(ctx, cleanID)
	if err != nil {
		return nil, fmt.Errorf("get paper details failed: %w", err)
	}

	if h.repo != nil && paper != nil {
		_ = h.repo.SavePaper(ctx, paper)
	}

	return paper, nil
}
