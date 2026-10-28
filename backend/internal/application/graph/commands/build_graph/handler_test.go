package build_graph_test

import (
	"context"
	"testing"

	"kuorum/backend/internal/application/graph/commands/build_graph"
	"kuorum/backend/internal/domain/graph"
	"kuorum/backend/internal/domain/retrieval"
)

type mockPaperApi struct {
	anchorDetails *retrieval.PaperMetadata
	references    []retrieval.PaperMetadata
	citations     []retrieval.PaperMetadata
}

func (m *mockPaperApi) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockPaperApi) SearchPapersWithProviders(ctx context.Context, query string, limit int, providers []string) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func (m *mockPaperApi) GetPaperDetails(ctx context.Context, paperID string) (*retrieval.PaperMetadata, error) {
	return m.anchorDetails, nil
}

func (m *mockPaperApi) GetReferences(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return m.references, nil
}

func (m *mockPaperApi) GetCitations(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return m.citations, nil
}

func (m *mockPaperApi) AvailableProviders() []string {
	return []string{"semanticscholar"}
}

type mockGraphRepo struct {
	graphs map[string]*graph.CitationGraph
}

func (m *mockGraphRepo) SaveGraph(ctx context.Context, g *graph.CitationGraph) error {
	if m.graphs == nil {
		m.graphs = make(map[string]*graph.CitationGraph)
	}
	m.graphs[g.AnchorPaperID] = g
	return nil
}

func (m *mockGraphRepo) GetGraph(ctx context.Context, anchorPaperID string) (*graph.CitationGraph, error) {
	if m.graphs == nil {
		return nil, graph.ErrGraphNotFound
	}
	g, ok := m.graphs[anchorPaperID]
	if !ok {
		return nil, graph.ErrGraphNotFound
	}
	return g, nil
}

func (m *mockGraphRepo) DeleteGraph(ctx context.Context, anchorPaperID string) error {
	delete(m.graphs, anchorPaperID)
	return nil
}

func TestBuildCitationGraphCommandHandler_Success(t *testing.T) {
	anchor := &retrieval.PaperMetadata{
		ID:            "anchor-123",
		Title:         "All-Solid-State Battery Review",
		Year:          2022,
		CitationCount: 150,
	}

	refs := []retrieval.PaperMetadata{
		{
			ID:            "ref-theory-1",
			Title:         "Original Dendrite Nucleation Mechanics",
			Year:          2015,
			CitationCount: 1200,
		},
		{
			ID:            "ref-theory-2",
			Title:         "Electrolyte Interface Equations",
			Year:          2018,
			CitationCount: 800,
		},
	}

	cites := []retrieval.PaperMetadata{
		{
			ID:            "cite-app-1",
			Title:         "EV Battery Pack Application with Solid Electrolyte",
			Year:          2024,
			CitationCount: 25,
		},
	}

	api := &mockPaperApi{
		anchorDetails: anchor,
		references:    refs,
		citations:     cites,
	}
	repo := &mockGraphRepo{}

	handler := build_graph.NewBuildCitationGraphCommandHandler(api, repo, nil)

	cmd := build_graph.BuildCitationGraphCommand{
		AnchorPaperID: "anchor-123",
		MaxReferences: 5,
		MaxCitations:  5,
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	res, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("build graph failed: %v", err)
	}

	if res.TotalNodes != 4 { // 1 anchor + 2 refs + 1 cite
		t.Errorf("expected 4 total nodes, got %d", res.TotalNodes)
	}
	if res.TotalEdges != 3 { // 2 out from anchor + 1 in to anchor
		t.Errorf("expected 3 total edges, got %d", res.TotalEdges)
	}
	if res.FoundationalCount != 2 {
		t.Errorf("expected 2 foundational refs, got %d", res.FoundationalCount)
	}
	if res.ApplicationCount != 1 {
		t.Errorf("expected 1 practical application, got %d", res.ApplicationCount)
	}
	if res.FromCache {
		t.Errorf("first run should not be from cache")
	}

	// Verify caching on second run
	cachedRes, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if !cachedRes.FromCache {
		t.Errorf("second run should be from cache")
	}

	// Verify ForceRefresh
	refreshCmd := cmd
	refreshCmd.ForceRefresh = true
	refreshedRes, err := handler.Handle(context.Background(), refreshCmd)
	if err != nil {
		t.Fatalf("refresh run failed: %v", err)
	}
	if refreshedRes.FromCache {
		t.Errorf("force refresh run should not be from cache")
	}
}

func TestBuildCitationGraphCommandHandler_WithCohort(t *testing.T) {
	anchor := &retrieval.PaperMetadata{
		ID:            "anchor-primary",
		Title:         "Transformer Architecture",
		Year:          2017,
		CitationCount: 50000,
	}

	api := &mockPaperApi{
		anchorDetails: anchor,
		references: []retrieval.PaperMetadata{
			{ID: "ref-attention-baseline", Title: "Neural Machine Translation by Jointly Learning to Align", Year: 2014, CitationCount: 20000},
		},
		citations: []retrieval.PaperMetadata{
			{ID: "cite-bert", Title: "BERT: Pre-training of Deep Bidirectional Transformers", Year: 2018, CitationCount: 40000},
		},
	}
	repo := &mockGraphRepo{}

	handler := build_graph.NewBuildCitationGraphCommandHandler(api, repo, nil)

	cmd := build_graph.BuildCitationGraphCommand{
		AnchorPaperID:  "anchor-primary",
		CohortPaperIDs: []string{"cohort-gpt", "cohort-t5"},
		MaxReferences:  5,
		MaxCitations:   5,
	}

	res, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("build graph with cohort failed: %v", err)
	}

	// 1 anchor + 1 ref + 1 cite + 2 cohort studies = 5 nodes
	if res.TotalNodes != 5 {
		t.Errorf("expected 5 total nodes, got %d", res.TotalNodes)
	}
	if res.CohortCount != 2 {
		t.Errorf("expected 2 cohort nodes, got %d", res.CohortCount)
	}
	// Edges: anchor->ref, cite->anchor, cohort1->anchor, cohort2->anchor = 4 edges
	if res.TotalEdges < 4 {
		t.Errorf("expected at least 4 edges, got %d", res.TotalEdges)
	}
}
