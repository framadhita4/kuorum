package get_models

import (
	"context"
	"testing"

	"kuorum/backend/internal/application/common/interfaces"
)

type mockCatalog struct{}

func (m *mockCatalog) GetAvailableModels() ([]interfaces.ModelCatalogInfo, []interfaces.ProviderCatalogInfo, string) {
	return []interfaces.ModelCatalogInfo{
		{
			ID:                  "gemini-2.0-flash",
			DisplayName:         "Gemini 2.0 Flash",
			Provider:            "gemini",
			ProviderDisplayName: "Google Gemini",
			Description:         "Ultra-fast multimodal model",
			ContextWindow:       1048576,
			MaxOutputTokens:     8192,
			ThinkingCapability:  "not_supported",
			AllowsTemperature:   true,
			IsDefault:           true,
		},
		{
			ID:                  "claude-3-7-sonnet-20250219",
			DisplayName:         "Claude 3.7 Sonnet",
			Provider:            "anthropic",
			ProviderDisplayName: "Anthropic",
			Description:         "Hybrid reasoning model",
			ContextWindow:       200000,
			MaxOutputTokens:     64000,
			ThinkingCapability:  "optional",
			SupportedEfforts:    []string{"low", "medium", "high"},
			AllowsTemperature:   true,
			IsDefault:           false,
		},
	}, []interfaces.ProviderCatalogInfo{
		{
			ID:           "gemini",
			DisplayName:  "Google Gemini",
			Description:  "Google official Gemini endpoint",
			DefaultModel: "gemini-2.0-flash",
			ModelIDs:     []string{"gemini-2.0-flash"},
		},
	}, "gemini-2.0-flash"
}

func TestGetModelsQueryHandler_Success(t *testing.T) {
	handler := NewGetModelsQueryHandler(&mockCatalog{})
	res, err := handler.Handle(context.Background(), GetModelsQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.DefaultModelID != "gemini-2.0-flash" {
		t.Errorf("expected default model gemini-2.0-flash, got %s", res.DefaultModelID)
	}

	if len(res.Models) != 2 {
		t.Fatalf("expected 2 models, got %d", len(res.Models))
	}

	if res.Models[0].DisplayName != "Gemini 2.0 Flash" {
		t.Errorf("expected first model 'Gemini 2.0 Flash', got '%s'", res.Models[0].DisplayName)
	}

	if len(res.Providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(res.Providers))
	}
}

func TestGetModelsQueryHandler_NilCatalog(t *testing.T) {
	handler := NewGetModelsQueryHandler(nil)
	res, err := handler.Handle(context.Background(), GetModelsQuery{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res == nil {
		t.Fatal("expected non-nil result")
	}

	if res.DefaultModelID != "gpt-4o-mini" {
		t.Errorf("expected default gpt-4o-mini, got %s", res.DefaultModelID)
	}
}
