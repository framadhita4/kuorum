package search_papers

import (
	"context"
	"fmt"

	"kuorum/backend/internal/application/common/interfaces"
)

type SearchPapersQueryHandler struct {
	paperApi interfaces.IPaperApi
	repo     interfaces.IPaperRepository
}

func NewSearchPapersQueryHandler(paperApi interfaces.IPaperApi, repo interfaces.IPaperRepository) *SearchPapersQueryHandler {
	return &SearchPapersQueryHandler{
		paperApi: paperApi,
		repo:     repo,
	}
}

func (h *SearchPapersQueryHandler) Handle(ctx context.Context, q SearchPapersQuery) (*SearchPapersResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 10
	}

	papers, err := h.paperApi.SearchPapersWithProviders(ctx, q.Query, limit, q.Providers)
	if err != nil {
		return nil, fmt.Errorf("search papers failed: %w", err)
	}

	// Persist papers asynchronously or opportunistically in repository if available
	if h.repo != nil {
		for i := range papers {
			_ = h.repo.SavePaper(ctx, &papers[i])
		}
	}

	return &SearchPapersResult{
		Query:     q.Query,
		Total:     len(papers),
		Providers: q.Providers,
		Papers:    papers,
	}, nil
}
