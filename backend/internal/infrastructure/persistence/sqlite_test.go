package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kuorum/backend/internal/domain/exploration"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/graph"
	"kuorum/backend/internal/domain/retrieval"
)

func setupTestDB(t *testing.T) (*SqlitePaperRepository, *SqliteGraphRepository, *SqliteExplorationRepository, *SqliteExtractionRepository, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "rh-test-db-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")
	database, err := NewSqliteDB(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to init test db: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(tempDir)
	}

	paperRepo := NewSqlitePaperRepository(database).(*SqlitePaperRepository)
	graphRepo := NewSqliteGraphRepository(database).(*SqliteGraphRepository)
	exploreRepo := NewSqliteExplorationRepository(database).(*SqliteExplorationRepository)
	extractRepo := NewSqliteExtractionRepository(database).(*SqliteExtractionRepository)

	return paperRepo, graphRepo, exploreRepo, extractRepo, cleanup
}

func TestExplorationRepository(t *testing.T) {
	ctx := context.Background()
	_, _, exploreRepo, _, cleanup := setupTestDB(t)
	defer cleanup()

	decomp, err := exploration.NewConceptDecomposition("Diffusion Models", "Denoising Diffusion Probabilistic Models", []exploration.ResearchAngle{
		{
			ID:               "angle-1",
			Title:            "Mathematical Foundations",
			Description:      "SDEs and score-based matching",
			SuggestedQueries: []string{"score matching diffusion", "langevin dynamics"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error creating decomposition: %v", err)
	}

	if err := exploreRepo.SaveDecomposition(ctx, decomp); err != nil {
		t.Fatalf("failed to save decomposition: %v", err)
	}

	loaded, err := exploreRepo.GetDecomposition(ctx, "Diffusion Models")
	if err != nil {
		t.Fatalf("failed to get decomposition: %v", err)
	}

	if loaded.NormalizedTerm != decomp.NormalizedTerm {
		t.Errorf("expected normalized term %q, got %q", decomp.NormalizedTerm, loaded.NormalizedTerm)
	}
	if len(loaded.Angles) != 1 {
		t.Fatalf("expected 1 angle, got %d", len(loaded.Angles))
	}
	if loaded.Angles[0].Title != "Mathematical Foundations" {
		t.Errorf("expected angle title %q, got %q", "Mathematical Foundations", loaded.Angles[0].Title)
	}
}

func TestPaperRepository(t *testing.T) {
	ctx := context.Background()
	paperRepo, _, _, _, cleanup := setupTestDB(t)
	defer cleanup()

	paper := &retrieval.PaperMetadata{
		ID:       "paper-transformer-2017",
		Title:    "Attention Is All You Need",
		Abstract: "The dominant sequence transduction models are based on complex recurrent or convolutional neural networks...",
		Authors: []retrieval.Author{
			{ID: "author-1", Name: "Ashish Vaswani"},
			{ID: "author-2", Name: "Noam Shazeer"},
		},
		Year:                     2017,
		CitationCount:            125000,
		InfluentialCitationCount: 15000,
		TLDR:                     "We propose the Transformer, a model architecture eschewing recurrence and entirely relying on an attention mechanism.",
		FieldsOfStudy:            []string{"Computer Science", "Artificial Intelligence"},
		OpenAccessPdfURL:         "https://arxiv.org/pdf/1706.03762.pdf",
		Doi:                      "10.48550/arXiv.1706.03762",
		Venue:                    "NeurIPS 2017",
		FetchedAt:                time.Now().UTC().Truncate(time.Second),
	}

	if err := paperRepo.SavePaper(ctx, paper); err != nil {
		t.Fatalf("failed to save paper: %v", err)
	}

	loaded, err := paperRepo.GetPaper(ctx, "paper-transformer-2017")
	if err != nil {
		t.Fatalf("failed to get paper: %v", err)
	}

	if loaded.Title != paper.Title {
		t.Errorf("expected title %q, got %q", paper.Title, loaded.Title)
	}
	if len(loaded.Authors) != 2 {
		t.Fatalf("expected 2 authors, got %d", len(loaded.Authors))
	}
	if loaded.Authors[0].Name != "Ashish Vaswani" {
		t.Errorf("expected first author %q, got %q", "Ashish Vaswani", loaded.Authors[0].Name)
	}

	// Search test
	searchRes, err := paperRepo.SearchLocalPapers(ctx, "Attention", 5)
	if err != nil {
		t.Fatalf("failed to search papers: %v", err)
	}
	if len(searchRes) == 0 {
		t.Fatalf("expected search to return paper, got 0 results")
	}
	if searchRes[0].ID != paper.ID {
		t.Errorf("expected found paper ID %q, got %q", paper.ID, searchRes[0].ID)
	}
}

func TestGraphRepository(t *testing.T) {
	ctx := context.Background()
	paperRepo, graphRepo, _, _, cleanup := setupTestDB(t)
	defer cleanup()

	// Anchor paper must exist for foreign key
	anchor := &retrieval.PaperMetadata{
		ID:    "paper-anchor-1",
		Title: "Anchor Paper",
		Year:  2023,
	}
	if err := paperRepo.SavePaper(ctx, anchor); err != nil {
		t.Fatalf("failed to save anchor paper: %v", err)
	}

	g := &graph.CitationGraph{
		AnchorPaperID: "paper-anchor-1",
		Nodes: []graph.GraphNode{
			{PaperID: "paper-anchor-1", Title: "Anchor Paper", Role: graph.NodeAnchor},
			{PaperID: "paper-ref-1", Title: "Foundational Baseline", Role: graph.NodeFoundationalReference},
		},
		Edges: []graph.GraphEdge{
			{SourcePaperID: "paper-anchor-1", TargetPaperID: "paper-ref-1", Intent: "methodology"},
		},
		TraversedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := graphRepo.SaveGraph(ctx, g); err != nil {
		t.Fatalf("failed to save citation graph: %v", err)
	}

	loaded, err := graphRepo.GetGraph(ctx, "paper-anchor-1")
	if err != nil {
		t.Fatalf("failed to get graph: %v", err)
	}

	if len(loaded.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(loaded.Nodes))
	}
	if len(loaded.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(loaded.Edges))
	}
}

func TestExtractionRepository(t *testing.T) {
	ctx := context.Background()
	_, _, _, extractRepo, cleanup := setupTestDB(t)
	defer cleanup()

	synthesis := &extraction.GroundedConceptSynthesis{
		AnchorPaperID: "paper-transformer-2017",
		Concept:       "Multi-Head Attention",
		Definitions: []extraction.GroundedDefinition{
			{
				Term:       "Multi-Head Attention",
				Definition: "Multi-head attention allows the model to jointly attend to information from different representation subspaces.",
				Citation: extraction.CitationProof{
					PaperID:      "paper-transformer-2017",
					VerbatimText: "Multi-head attention allows the model to jointly attend...",
					Section:      "Section 3.2.2",
					PageNumber:   4,
				},
			},
		},
		Mechanisms: []extraction.GroundedMechanism{
			{
				StepNumber:  1,
				Description: "Linear projections of queries, keys, and values into dk dimensions.",
				Citation: extraction.CitationProof{
					PaperID:      "paper-transformer-2017",
					VerbatimText: "Instead of performing a single attention function...",
					Section:      "Section 3.2.2",
				},
			},
		},
		GeneratedAt: time.Now().UTC().Truncate(time.Second),
	}

	if err := extractRepo.SaveSynthesis(ctx, synthesis); err != nil {
		t.Fatalf("failed to save synthesis: %v", err)
	}

	loaded, err := extractRepo.GetSynthesis(ctx, "paper-transformer-2017", "Multi-Head Attention")
	if err != nil {
		t.Fatalf("failed to get synthesis: %v", err)
	}

	if len(loaded.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(loaded.Definitions))
	}
	if loaded.Definitions[0].Citation.Section != "Section 3.2.2" {
		t.Errorf("expected section 'Section 3.2.2', got %q", loaded.Definitions[0].Citation.Section)
	}
}
