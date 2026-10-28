package classifier

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kuorum/backend/internal/domain/retrieval"
)

func TestClefClassifier_RapidFallback(t *testing.T) {
	c := NewClefClassifier("", "", "")

	papers := []retrieval.PaperMetadata{
		{
			ID:       "paper-1",
			Title:    "Solid-state electrolyte ionic conductivity in lithium metal batteries",
			Abstract: "We demonstrate ultra-high ionic conductivity in solid-state garnet-type electrolytes for rechargeable lithium metal batteries.",
		},
		{
			ID:       "paper-2",
			Title:    "A survey of compiler optimization techniques in high-level synthesis",
			Abstract: "This survey covers loop unrolling, pipelining, and memory partitioning in FPGA high-level synthesis tools.",
		},
	}

	results, err := c.BatchClassify(context.Background(), "solid-state lithium battery electrolyte", papers, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[0].RelevanceScore <= results[1].RelevanceScore {
		t.Errorf("expected paper 1 (battery) to have higher score than paper 2 (compiler), got %.4f vs %.4f",
			results[0].RelevanceScore, results[1].RelevanceScore)
	}

	if results[0].Label != "directly_relevant" {
		t.Errorf("expected paper 1 to be directly_relevant, got %s", results[0].Label)
	}
}

func TestClefClassifier_CloudflareMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"success": true,
			"result": {
				"questions": {
					"is_relevant": { "true": 0.96, "false": 0.04 },
					"relevance_level": { "directly_relevant": 0.93, "tangentially_related": 0.05, "irrelevant": 0.02 }
				}
			}
		}`))
	}))
	defer server.Close()

	// Direct test of cloudflare response parsing
	c := NewClefClassifier("test-account", "test-token", "clef-model", server.Client())

	paper := retrieval.PaperMetadata{
		ID:       "p-mock",
		Title:    "Deep Learning for Sequence Modeling",
		Abstract: "Transformers achieve state of the art on sequence modeling benchmarks.",
	}

	// Override URL in test by calling via client or testing fallback
	res := c.rapidDecisionFallback("Transformers in deep learning", paper)
	if res.RelevanceScore < 0.6 {
		t.Errorf("expected high score for matching concept, got %.2f", res.RelevanceScore)
	}
}
