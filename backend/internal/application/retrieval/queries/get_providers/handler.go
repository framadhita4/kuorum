package get_providers

import (
	"context"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetProvidersQueryHandler struct {
	paperApi interfaces.IPaperApi
}

func NewGetProvidersQueryHandler(paperApi interfaces.IPaperApi) *GetProvidersQueryHandler {
	return &GetProvidersQueryHandler{
		paperApi: paperApi,
	}
}

func (h *GetProvidersQueryHandler) Handle(ctx context.Context, q GetProvidersQuery) (*GetProvidersResult, error) {
	return &GetProvidersResult{
		Providers: h.paperApi.AvailableProviders(),
	}, nil
}
