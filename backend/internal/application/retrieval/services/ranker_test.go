package services_test

import (
	"testing"
	"time"

	"kuorum/backend/internal/application/retrieval/services"
	"kuorum/backend/internal/domain/retrieval"
)

func TestAnchorPaperRanker_EmptyInput(t *testing.T) {
	ranker := services.NewAnchorPaperRanker()
	result := ranker.Rank("transformer attention", nil)

	if result.TotalRaw != 0 || result.TotalPassed != 0 {
		t.Errorf("expected 0 papers, got raw=%d, passed=%d", result.TotalRaw, result.TotalPassed)
	}
	if result.AnchorPaper != nil {
		t.Errorf("expected nil anchor paper for empty input")
	}
}

func TestAnchorPaperRanker_QualityGateFiltering(t *testing.T) {
	opts := services.DefaultRankerOptions()
	opts.MinAbstractLength = 50
	opts.RequireOpenAccess = true
	ranker := services.NewAnchorPaperRanker(opts)

	rawPapers := []retrieval.PaperMetadata{
		{
			ID:               "p1-no-title",
			Title:            "",
			Abstract:         "This is a sufficiently long abstract describing the paper findings in detail.",
			OpenAccessPdfURL: "https://arxiv.org/pdf/1234.pdf",
		},
		{
			ID:               "p2-short-abstract",
			Title:            "Valid Title Here",
			Abstract:         "Too short",
			OpenAccessPdfURL: "https://arxiv.org/pdf/1234.pdf",
		},
		{
			ID:               "p3-no-open-access",
			Title:            "Paywalled Paper Title",
			Abstract:         "This is a sufficiently long abstract describing the paper findings in detail.",
			OpenAccessPdfURL: "",
		},
		{
			ID:               "p4-valid",
			Title:            "Valid Open Access Paper",
			Abstract:         "This is a sufficiently long abstract describing the paper findings in detail.",
			OpenAccessPdfURL: "https://arxiv.org/pdf/1234.pdf",
			Year:             2023,
		},
	}

	result := ranker.Rank("valid open access", rawPapers)

	if result.TotalRaw != 4 {
		t.Fatalf("expected 4 raw papers, got %d", result.TotalRaw)
	}
	if result.TotalPassed != 1 {
		t.Fatalf("expected 1 paper to pass quality gate, got %d", result.TotalPassed)
	}
	if result.AnchorPaper == nil || result.AnchorPaper.Paper.ID != "p4-valid" {
		t.Errorf("expected p4-valid as anchor, got %v", result.AnchorPaper)
	}
}

func TestAnchorPaperRanker_ScoringAndAnchorDesignation(t *testing.T) {
	ranker := services.NewAnchorPaperRanker()

	currentYear := time.Now().Year()
	papers := []retrieval.PaperMetadata{
		{
			ID:            "paper-irrelevant",
			Title:         "Quantum Computing Superconducting Qubits",
			Abstract:      "We present hardware architectures for superconducting flux qubits with low coherence decay.",
			CitationCount: 50,
			Year:          currentYear - 2,
		},
		{
			ID:                       "paper-relevant-anchor",
			Title:                    "All-Solid-State Lithium Batteries: Dendrite Suppression at Subzero Temperatures",
			Abstract:                 "Solid-state lithium metal batteries suffer from rapid dendrite growth at low temperatures. Here we demonstrate in-situ interphase engineering to completely suppress lithium dendrite penetration.",
			CitationCount:            450,
			InfluentialCitationCount: 42,
			OpenAccessPdfURL:         "https://arxiv.org/pdf/2301.0001.pdf",
			Year:                     currentYear - 1,
		},
		{
			ID:               "paper-somewhat-relevant",
			Title:            "Lithium Batteries at Low Temperatures: Electrolyte Review",
			Abstract:         "A broad review on liquid and polymer electrolyte battery performance under freezing climates.",
			CitationCount:    80,
			OpenAccessPdfURL: "https://openalex.org/pdf/abc",
			Year:             currentYear - 4,
		},
	}

	query := "solid-state lithium battery dendrite suppression low temperature"
	result := ranker.Rank(query, papers)

	if result.TotalPassed != 3 {
		t.Fatalf("expected 3 papers to pass, got %d", result.TotalPassed)
	}

	if result.AnchorPaper == nil {
		t.Fatal("expected AnchorPaper to be non-nil")
	}

	if result.AnchorPaper.Paper.ID != "paper-relevant-anchor" {
		t.Errorf("expected 'paper-relevant-anchor' as #1, got %s", result.AnchorPaper.Paper.ID)
	}

	if !result.AnchorPaper.IsAnchor {
		t.Errorf("expected IsAnchor to be true for top-ranked paper")
	}

	if result.AnchorPaper.Rank != 1 {
		t.Errorf("expected rank 1, got %d", result.AnchorPaper.Rank)
	}

	// Ensure other papers do not have IsAnchor == true
	for _, cand := range result.Candidates[1:] {
		if cand.IsAnchor {
			t.Errorf("candidate %s should not be marked as anchor", cand.Paper.ID)
		}
	}

	// Verify composite score order: descending
	for i := 1; i < len(result.Candidates); i++ {
		prev := result.Candidates[i-1].CompositeScore
		curr := result.Candidates[i].CompositeScore
		if prev < curr {
			t.Errorf("rank %d score (%.4f) should be >= rank %d score (%.4f)", i, prev, i+1, curr)
		}
	}
}

func TestAnchorPaperRanker_DeduplicationAndCohortSelection(t *testing.T) {
	ranker := services.NewAnchorPaperRanker()

	papers := []retrieval.PaperMetadata{
		{
			ID:       "dup-1",
			Doi:      "10.1016/j.battery.2023.01",
			Title:    "Solid-State Electrolyte Mechanisms",
			Abstract: "Detailed study on ionic conductivity in solid electrolyte materials.",
		},
		{
			ID:       "dup-2-same-doi",
			Doi:      "10.1016/j.battery.2023.01",
			Title:    "Solid-State Electrolyte Mechanisms (Preprint)",
			Abstract: "Preprint version of the same study on ionic conductivity.",
		},
		{
			ID:       "p3-different",
			Doi:      "10.1016/j.battery.2023.02",
			Title:    "Lithium Anode Interphase Engineering",
			Abstract: "Engineering the interface between solid electrolyte and lithium anode.",
		},
	}

	result := ranker.Rank("solid-state electrolyte lithium anode", papers)

	if result.TotalPassed != 2 {
		t.Fatalf("expected 2 unique papers after deduplication, got %d", result.TotalPassed)
	}

	cohort := result.SelectCohort(2)
	if len(cohort) != 2 {
		t.Fatalf("expected 2 cohort papers selected, got %d", len(cohort))
	}
}
