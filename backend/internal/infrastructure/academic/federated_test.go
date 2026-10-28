package academic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kuorum/backend/internal/domain/retrieval"
	"kuorum/backend/internal/infrastructure/academic/providers"
)

func TestDeduplicateAndMerge(t *testing.T) {
	paperA := retrieval.PaperMetadata{
		ID:               "openalex-1",
		Title:            "Attention Is All You Need",
		Abstract:         "Paper abstract from OpenAlex.",
		Doi:              "10.5555/3295222.3295349",
		Year:             2017,
		CitationCount:    80000,
		OpenAccessPdfURL: "https://openalex.org/pdf/attention.pdf",
		Sources:          []string{"openalex"},
	}

	paperB := retrieval.PaperMetadata{
		ID:                       "s2-paper-xyz",
		Title:                    "Attention Is All You Need",
		Abstract:                 "Detailed abstract from Semantic Scholar with more completeness.",
		TLDR:                     "We present the Transformer architecture.",
		Doi:                      "10.5555/3295222.3295349",
		Year:                     2017,
		CitationCount:            95000,
		InfluentialCitationCount: 12000,
		Venue:                    "NeurIPS 2017",
		Sources:                  []string{"semanticscholar"},
	}

	paperC := retrieval.PaperMetadata{
		ID:            "arxiv-3",
		Title:         "Graph Convolutional Networks",
		Doi:           "10.1234/gcn",
		Year:          2016,
		CitationCount: 15000,
		Sources:       []string{"arxiv"},
	}

	results := DeduplicateAndMerge([]retrieval.PaperMetadata{paperA, paperB, paperC})

	if len(results) != 2 {
		t.Fatalf("expected 2 merged papers, got %d", len(results))
	}

	// First paper should be the one with highest citations (Attention Is All You Need)
	mergedAttention := results[0]
	if mergedAttention.Title != "Attention Is All You Need" {
		t.Errorf("expected first paper to be Attention Is All You Need, got %s", mergedAttention.Title)
	}
	if mergedAttention.CitationCount != 95000 {
		t.Errorf("expected max citation count 95000, got %d", mergedAttention.CitationCount)
	}
	if mergedAttention.TLDR != "We present the Transformer architecture." {
		t.Errorf("TLDR was not preserved during merge: '%s'", mergedAttention.TLDR)
	}
	if mergedAttention.OpenAccessPdfURL != "https://openalex.org/pdf/attention.pdf" {
		t.Errorf("OpenAccessPdfURL was not preserved during merge: '%s'", mergedAttention.OpenAccessPdfURL)
	}
	if len(mergedAttention.Sources) != 2 {
		t.Errorf("expected 2 sources merged, got %+v", mergedAttention.Sources)
	}
}

func TestFederatedClient_SearchWithProviders(t *testing.T) {
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":"oa-1","title":"Federated Learning Survey","publication_year":2021,"cited_by_count":500}]}`))
	}))
	defer s1.Close()

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"paperId":"s2-1","title":"Federated Learning Survey","year":2021,"citationCount":520,"tldr":{"text":"A survey of federated algorithms."}}]}`))
	}))
	defer s2.Close()

	oaProvider := providers.NewOpenAlexProvider(s1.URL, "test@example.com", s1.Client())
	oaProvider.SetRateLimiter(providers.NewRateLimiter(0))

	s2Provider := providers.NewSemanticScholarProvider(s2.URL, "", s2.Client())
	s2Provider.SetRateLimiter(providers.NewRateLimiter(0))

	client := NewFederatedClient(oaProvider, s2Provider)

	ctx := context.Background()

	// Test 1: Query specifically with selected providers
	papers, err := client.SearchPapersWithProviders(ctx, "federated learning", 10, []string{"openalex", "semanticscholar"})
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(papers) != 1 {
		t.Fatalf("expected 1 deduplicated merged paper, got %d", len(papers))
	}

	p := papers[0]
	if p.CitationCount != 520 {
		t.Errorf("expected merged citation count 520, got %d", p.CitationCount)
	}
	if p.TLDR != "A survey of federated algorithms." {
		t.Errorf("expected merged TLDR, got '%s'", p.TLDR)
	}

	// Test 2: Query single provider
	oaOnly, err := client.SearchPapersWithProviders(ctx, "federated learning", 10, []string{"openalex"})
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}
	if len(oaOnly) != 1 || len(oaOnly[0].Sources) != 1 || oaOnly[0].Sources[0] != "openalex" {
		t.Errorf("expected only openalex provider, got %+v", oaOnly[0].Sources)
	}
}

func TestDeduplicateAndRank_RelevanceVsCitations(t *testing.T) {
	// Famous generic YOLO paper without any caching
	genericYolo := retrieval.PaperMetadata{
		ID:            "yolo-orig",
		Title:         "You Only Look Once: Unified, Real-Time Object Detection",
		Abstract:      "We present YOLO, a new approach to object detection. Prior work on object detection repurposes classifiers.",
		Year:          2016,
		CitationCount: 45000,
	}

	// Unrelated paper with the word "model" in agriculture
	unrelatedPaper := retrieval.PaperMetadata{
		ID:            "agri-model",
		Title:         "A Crop Yield Simulation Model for Modern Agriculture",
		Abstract:      "This paper describes a statistical model predicting wheat yields based on weather patterns.",
		Year:          2021,
		CitationCount: 300,
	}

	// Highly relevant paper specifically discussing caching for YOLO models (low citations)
	cachingYolo := retrieval.PaperMetadata{
		ID:            "yolo-cache",
		Title:         "Intermediate Feature Caching for YOLO Inference Acceleration on Edge Devices",
		Abstract:      "Deep neural network caching can eliminate redundant computations. We present a feature caching method for YOLO models.",
		Year:          2024,
		CitationCount: 8,
	}

	query := "YOLO model caching"
	results := DeduplicateAndRank([]retrieval.PaperMetadata{genericYolo, unrelatedPaper, cachingYolo}, query)

	if len(results) != 3 {
		t.Fatalf("expected 3 papers, got %d", len(results))
	}

	// The paper matching all concepts (YOLO + model + caching) must rank #1 despite having only 8 citations!
	if results[0].ID != "yolo-cache" {
		t.Errorf("expected #1 paper to be 'yolo-cache', got '%s' (title: %s)", results[0].ID, results[0].Title)
	}

	// The unrelated paper should be ranked last
	if results[2].ID != "agri-model" {
		t.Errorf("expected last paper to be 'agri-model', got '%s'", results[2].ID)
	}
}
