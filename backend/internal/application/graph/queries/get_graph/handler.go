package get_graph

import (
	"context"
	"fmt"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetCitationGraphQueryHandler struct {
	repo interfaces.IGraphRepository
}

func NewGetCitationGraphQueryHandler(repo interfaces.IGraphRepository) *GetCitationGraphQueryHandler {
	return &GetCitationGraphQueryHandler{repo: repo}
}

func (h *GetCitationGraphQueryHandler) Handle(ctx context.Context, q GetCitationGraphQuery) (*GetCitationGraphResult, error) {
	g, err := h.repo.GetGraph(ctx, q.AnchorPaperID)
	if err != nil {
		return nil, fmt.Errorf("fetching citation graph failed: %w", err)
	}
	return &GetCitationGraphResult{Graph: g}, nil
}
