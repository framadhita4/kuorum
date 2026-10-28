package get_models

import (
	"context"

	"kuorum/backend/internal/application/common/interfaces"
)

type GetModelsQueryHandler struct {
	catalog interfaces.IModelCatalog
}

func NewGetModelsQueryHandler(catalog interfaces.IModelCatalog) *GetModelsQueryHandler {
	return &GetModelsQueryHandler{
		catalog: catalog,
	}
}

func (h *GetModelsQueryHandler) Handle(ctx context.Context, q GetModelsQuery) (*GetModelsResult, error) {
	if h.catalog == nil {
		return &GetModelsResult{
			DefaultModelID: "gpt-4o-mini",
			Models:         []ModelDTO{},
			Providers:      []ProviderDTO{},
		}, nil
	}

	modelsInfo, providersInfo, defaultModel := h.catalog.GetAvailableModels()

	models := make([]ModelDTO, len(modelsInfo))
	for i, m := range modelsInfo {
		models[i] = ModelDTO{
			ID:                  m.ID,
			DisplayName:         m.DisplayName,
			Provider:            m.Provider,
			ProviderDisplayName: m.ProviderDisplayName,
			Description:         m.Description,
			ContextWindow:       m.ContextWindow,
			MaxOutputTokens:     m.MaxOutputTokens,
			ThinkingCapability:  m.ThinkingCapability,
			SupportedEfforts:    m.SupportedEfforts,
			AllowsTemperature:   m.AllowsTemperature,
			IsDefault:           m.IsDefault,
		}
	}

	providers := make([]ProviderDTO, len(providersInfo))
	for i, p := range providersInfo {
		providers[i] = ProviderDTO{
			ID:           p.ID,
			DisplayName:  p.DisplayName,
			Description:  p.Description,
			DefaultModel: p.DefaultModel,
			ModelIDs:     p.ModelIDs,
		}
	}

	return &GetModelsResult{
		DefaultModelID: defaultModel,
		Models:         models,
		Providers:      providers,
	}, nil
}
