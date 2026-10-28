package get_references

import (
	"context"
	"fmt"
	"strings"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetReferencesQueryHandler struct {
	paperApi interfaces.IPaperApi
}

func NewGetReferencesQueryHandler(paperApi interfaces.IPaperApi) *GetReferencesQueryHandler {
	return &GetReferencesQueryHandler{
		paperApi: paperApi,
	}
}

func (h *GetReferencesQueryHandler) Handle(ctx context.Context, q GetReferencesQuery) (*GetReferencesResult, error) {
	cleanID := strings.TrimSpace(q.ID)
	refs, err := h.paperApi.GetReferences(ctx, cleanID)
	if err != nil {
		return nil, fmt.Errorf("get references failed: %w", err)
	}

	return &GetReferencesResult{
		PaperID:    cleanID,
		Total:      len(refs),
		References: refs,
	}, nil
}
