package get_citations

import (
	"context"
	"fmt"
	"strings"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetCitationsQueryHandler struct {
	paperApi interfaces.IPaperApi
}

func NewGetCitationsQueryHandler(paperApi interfaces.IPaperApi) *GetCitationsQueryHandler {
	return &GetCitationsQueryHandler{
		paperApi: paperApi,
	}
}

func (h *GetCitationsQueryHandler) Handle(ctx context.Context, q GetCitationsQuery) (*GetCitationsResult, error) {
	cleanID := strings.TrimSpace(q.ID)
	cits, err := h.paperApi.GetCitations(ctx, cleanID)
	if err != nil {
		return nil, fmt.Errorf("get citations failed: %w", err)
	}

	return &GetCitationsResult{
		PaperID:   cleanID,
		Total:     len(cits),
		Citations: cits,
	}, nil
}
