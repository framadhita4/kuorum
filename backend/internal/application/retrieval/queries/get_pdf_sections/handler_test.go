package get_pdf_sections

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"kuorum/backend/internal/domain/retrieval"
)

type mockCache struct {
	data map[string][]byte
}

func (m *mockCache) Get(ctx context.Context, key string) ([]byte, error) {
	if val, ok := m.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("not found")
}

func (m *mockCache) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = val
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestGetPdfSectionsQueryHandler(t *testing.T) {
	doc := &retrieval.PdfDocument{
		ID:    "attention-paper",
		Title: "Attention Is All You Need",
		Sections: []retrieval.ParsedSection{
			{Type: retrieval.SectionAbstract, Title: "Abstract", Content: "Abstract text"},
			{Type: retrieval.SectionResults, Title: "4. Results", Content: "Results text"},
		},
	}
	docBytes, _ := json.Marshal(doc)

	cache := &mockCache{
		data: map[string][]byte{
			"pdf:doc:attention-paper": docBytes,
		},
	}

	handler := NewGetPdfSectionsQueryHandler(cache, nil, nil)

	q := GetPdfSectionsQuery{
		PaperID: "attention-paper",
		Intent:  "results",
	}

	if err := q.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	res, err := handler.Handle(context.Background(), q)
	if err != nil {
		t.Fatalf("handler failed: %v", err)
	}

	if res.AnchorSection == nil || res.AnchorSection.Type != retrieval.SectionAbstract {
		t.Errorf("expected Abstract as anchor section")
	}
	if len(res.TargetSections) != 1 || res.TargetSections[0].Type != retrieval.SectionResults {
		t.Errorf("expected Results as target section")
	}
}
