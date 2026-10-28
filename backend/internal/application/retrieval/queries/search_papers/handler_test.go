package search_papers

import (
	"context"
	"errors"
	"testing"

	"kuorum/backend/internal/domain/retrieval"
)

type mockPaperApi struct {
	shouldFail bool
}

func (m *mockPaperApi) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	return m.SearchPapersWithProviders(ctx, query, limit, nil)
}

func (m *mockPaperApi) SearchPapersWithProviders(ctx context.Context, query string, limit int, providers []string) ([]retrieval.PaperMetadata, error) {
	if m.shouldFail {
		return nil, errors.New("provider failure")
	}
	return []retrieval.PaperMetadata{
		{ID: "p1", Title: "Paper 1", Year: 2024},
		{ID: "p2", Title: "Paper 2", Year: 2023},
	}, nil
}

func (m *mockPaperApi) GetPaperDetails(ctx context.Context, paperID string) (*retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockPaperApi) GetReferences(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockPaperApi) GetCitations(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockPaperApi) AvailableProviders() []string {
	return []string{"openalex", "arxiv"}
}

func TestSearchPapersQuery_Validate(t *testing.T) {
	valid := SearchPapersQuery{Query: "quantum computing", Limit: 10}
	if err := valid.Validate(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	invalidEmpty := SearchPapersQuery{Query: "   "}
	if err := invalidEmpty.Validate(); err == nil {
		t.Errorf("expected error for empty query")
	}

	invalidLimit := SearchPapersQuery{Query: "valid", Limit: -1}
	if err := invalidLimit.Validate(); err == nil {
		t.Errorf("expected error for negative limit")
	}
}

func TestSearchPapersQueryHandler_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful search", func(t *testing.T) {
		h := NewSearchPapersQueryHandler(&mockPaperApi{}, nil)
		res, err := h.Handle(ctx, SearchPapersQuery{Query: "transformers", Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Total != 2 {
			t.Errorf("expected 2 papers, got %d", res.Total)
		}
		if res.Query != "transformers" {
			t.Errorf("expected query transformers, got %s", res.Query)
		}
	})

	t.Run("provider error propagation", func(t *testing.T) {
		h := NewSearchPapersQueryHandler(&mockPaperApi{shouldFail: true}, nil)
		_, err := h.Handle(ctx, SearchPapersQuery{Query: "error test"})
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})
}
