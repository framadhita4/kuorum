package synthesize_concept_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"kuorum/backend/internal/application/extraction/commands/synthesize_concept"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/retrieval"
)

type mockLLMClient struct {
	responseJSON string
	err          error
}

func (m *mockLLMClient) CompleteStructured(ctx context.Context, systemPrompt, userPrompt string, target any) error {
	if m.err != nil {
		return m.err
	}
	return json.Unmarshal([]byte(m.responseJSON), target)
}

func (m *mockLLMClient) Stream(ctx context.Context, systemPrompt, userPrompt string) (<-chan string, <-chan error) {
	return nil, nil
}

type mockExtractionRepo struct {
	saved map[string]*extraction.GroundedConceptSynthesis
}

func (m *mockExtractionRepo) SaveSynthesis(ctx context.Context, s *extraction.GroundedConceptSynthesis) error {
	if m.saved == nil {
		m.saved = make(map[string]*extraction.GroundedConceptSynthesis)
	}
	m.saved[s.AnchorPaperID+":"+s.Concept] = s
	return nil
}

func (m *mockExtractionRepo) GetSynthesis(ctx context.Context, anchorPaperID string, concept string) (*extraction.GroundedConceptSynthesis, error) {
	if m.saved == nil {
		return nil, extraction.ErrUngroundedClaim
	}
	s, ok := m.saved[anchorPaperID+":"+concept]
	if !ok {
		return nil, extraction.ErrUngroundedClaim
	}
	return s, nil
}

type mockCacheRepo struct {
	items map[string][]byte
}

func (m *mockCacheRepo) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := m.items[key]
	if !ok {
		return nil, retrieval.ErrPaperNotFound
	}
	return val, nil
}

func (m *mockCacheRepo) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.items == nil {
		m.items = make(map[string][]byte)
	}
	m.items[key] = value
	return nil
}

func (m *mockCacheRepo) Delete(ctx context.Context, key string) error {
	delete(m.items, key)
	return nil
}

func TestSynthesizeConceptCommandHandler_Success(t *testing.T) {
	exactQuote := "The in-situ formed LiF-rich interphase suppressed dendrite penetration by 84% under -20 °C testing conditions."

	mockDoc := &retrieval.PdfDocument{
		ID:    "paper-battery-101",
		Title: "Solid-State Electrolyte Interphase Engineering",
		Sections: []retrieval.ParsedSection{
			{
				ID:      "sec-abs",
				Type:    retrieval.SectionAbstract,
				Title:   "Abstract",
				Content: "Here we report an artificial interphase for lithium metal batteries.",
			},
			{
				ID:      "sec-meth",
				Type:    retrieval.SectionMethodology,
				Title:   "3. Interphase Synthesis",
				Content: "We spin-coated fluorinated polymer onto lithium metal foils to form the composite.",
			},
			{
				ID:      "sec-res",
				Type:    retrieval.SectionResults,
				Title:   "4. Results",
				Content: exactQuote,
			},
		},
	}

	docBytes, _ := json.Marshal(mockDoc)
	cache := &mockCacheRepo{
		items: map[string][]byte{
			"pdf:doc:paper-battery-101": docBytes,
		},
	}

	llmResponse := `{
		"definitions": [
			{
				"term": "Artificial SEI",
				"definition": "An artificial solid electrolyte interphase engineered onto lithium metal.",
				"verbatimText": "Here we report an artificial interphase for lithium metal batteries.",
				"section": "Abstract"
			}
		],
		"mechanisms": [
			{
				"stepNumber": 1,
				"description": "Spin-coating fluorinated polymer onto lithium metal foils.",
				"verbatimText": "We spin-coated fluorinated polymer onto lithium metal foils to form the composite.",
				"section": "3. Interphase Synthesis"
			}
		],
		"caseStudies": [
			{
				"industryDomain": "Electric Vehicles",
				"problemSolved": "Subzero dendrite short-circuits",
				"methodApplied": "LiF interphase coating",
				"outcome": "84% dendrite suppression",
				"verbatimText": "The in-situ formed LiF-rich interphase suppressed dendrite penetration by 84% under -20 °C testing conditions.",
				"section": "4. Results"
			}
		]
	}`

	llm := &mockLLMClient{responseJSON: llmResponse}
	repo := &mockExtractionRepo{}

	handler := synthesize_concept.NewSynthesizeConceptCommandHandler(llm, repo, cache, nil, nil)

	cmd := synthesize_concept.SynthesizeConceptCommand{
		AnchorPaperID: "paper-battery-101",
		Concept:       "Lithium Dendrite Suppression",
		Intent:        "all",
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	if result.Synthesis == nil {
		t.Fatal("expected non-nil synthesis")
	}

	if len(result.Synthesis.Definitions) != 1 {
		t.Errorf("expected 1 definition, got %d", len(result.Synthesis.Definitions))
	}
	if len(result.Synthesis.Mechanisms) != 1 {
		t.Errorf("expected 1 mechanism, got %d", len(result.Synthesis.Mechanisms))
	}
	if len(result.Synthesis.CaseStudies) != 1 {
		t.Errorf("expected 1 case study, got %d", len(result.Synthesis.CaseStudies))
	}

	// Verify Grounding Audit: all 3 quotes were verbatim from the document
	if !result.Audit.AuditPassed {
		t.Errorf("expected audit to pass, but failed: %+v", result.Audit)
	}
	if result.Audit.VerifiedClaims != 3 {
		t.Errorf("expected 3 verified claims, got %d", result.Audit.VerifiedClaims)
	}
	if result.Audit.VerificationRate != 1.0 {
		t.Errorf("expected verification rate 1.0, got %f", result.Audit.VerificationRate)
	}
}

func TestSynthesizeConceptCommandHandler_HallucinationDetected(t *testing.T) {
	mockDoc := &retrieval.PdfDocument{
		ID:    "paper-hallucinate",
		Title: "Actual Real Paper Content",
		Sections: []retrieval.ParsedSection{
			{
				ID:      "sec-res",
				Type:    retrieval.SectionResults,
				Title:   "Results",
				Content: "We observed a modest 10% increase in ionic conductivity at room temperature.",
			},
		},
	}

	docBytes, _ := json.Marshal(mockDoc)
	cache := &mockCacheRepo{
		items: map[string][]byte{
			"pdf:doc:paper-hallucinate": docBytes,
		},
	}

	// LLM fabricates a completely fictional quote
	llmResponse := `{
		"definitions": [
			{
				"term": "Magical Battery",
				"definition": "A battery that never runs out of energy.",
				"verbatimText": "This fictional phrase does not exist in the real paper text anywhere at all.",
				"section": "Results"
			}
		],
		"mechanisms": [],
		"caseStudies": []
	}`

	llm := &mockLLMClient{responseJSON: llmResponse}
	repo := &mockExtractionRepo{}

	handler := synthesize_concept.NewSynthesizeConceptCommandHandler(llm, repo, cache, nil, nil)

	cmd := synthesize_concept.SynthesizeConceptCommand{
		AnchorPaperID: "paper-hallucinate",
		Concept:       "Perpetual Battery",
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	// The auditor should flag this as unverified
	if result.Audit.AuditPassed {
		t.Errorf("expected audit to FAIL for fabricated quote, but it passed")
	}
	if result.Audit.VerifiedClaims != 0 {
		t.Errorf("expected 0 verified claims, got %d", result.Audit.VerifiedClaims)
	}
}

type mockPaperRepo struct {
	papers map[string]*retrieval.PaperMetadata
}

func (m *mockPaperRepo) SavePaper(ctx context.Context, p *retrieval.PaperMetadata) error {
	if m.papers == nil {
		m.papers = make(map[string]*retrieval.PaperMetadata)
	}
	m.papers[p.ID] = p
	return nil
}

func (m *mockPaperRepo) GetPaper(ctx context.Context, id string) (*retrieval.PaperMetadata, error) {
	if m.papers == nil {
		return nil, retrieval.ErrPaperNotFound
	}
	p, ok := m.papers[id]
	if !ok {
		return nil, retrieval.ErrPaperNotFound
	}
	return p, nil
}

func (m *mockPaperRepo) SearchLocalPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	return nil, nil
}

func TestSynthesizeConceptCommandHandler_RepoFallback_Success(t *testing.T) {
	quote := "Weakly solvating electrolytes offer another solution to these challenges in lithium batteries."

	paperRepo := &mockPaperRepo{
		papers: map[string]*retrieval.PaperMetadata{
			"https://openalex.org/W4386509670": {
				ID:       "https://openalex.org/W4386509670",
				Title:    "Weakly solvating electrolytes for next-generation lithium batteries",
				Abstract: quote,
			},
		},
	}

	llmResponse := fmt.Sprintf(`{
		"definitions": [
			{
				"term": "Weakly Solvating Electrolytes",
				"definition": "Electrolytes that offer another solution for lithium battery challenges.",
				"verbatimText": %q,
				"section": "Abstract",
				"paperId": "https://openalex.org/W4386509670"
			}
		],
		"mechanisms": [],
		"caseStudies": []
	}`, quote)

	llm := &mockLLMClient{responseJSON: llmResponse}
	cache := &mockCacheRepo{items: map[string][]byte{}}

	// Create handler with empty cache and empty storage, but with paperRepo
	handler := synthesize_concept.NewSynthesizeConceptCommandHandler(llm, nil, cache, nil, nil, paperRepo)

	cmd := synthesize_concept.SynthesizeConceptCommand{
		AnchorPaperID: "https://openalex.org/W4386509670",
		Concept:       "Weakly Solvating Electrolytes",
		Intent:        "all",
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected success with repo fallback, got error: %v", err)
	}

	if result.Synthesis == nil {
		t.Fatal("expected non-nil synthesis")
	}
	if len(result.Synthesis.Definitions) != 1 {
		t.Errorf("expected 1 definition, got %d", len(result.Synthesis.Definitions))
	}
	if !result.Audit.AuditPassed {
		t.Errorf("expected audit to pass, but failed: %+v", result.Audit)
	}
}
