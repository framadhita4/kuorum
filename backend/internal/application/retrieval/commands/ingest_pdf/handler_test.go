package ingest_pdf

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

type mockPdfParser struct {
	returnDoc *retrieval.PdfDocument
	err       error
}

func (m *mockPdfParser) Parse(ctx context.Context, reader io.Reader, filename string) (*retrieval.PdfDocument, error) {
	return m.returnDoc, m.err
}

func (m *mockPdfParser) ParseFile(ctx context.Context, filePath string) (*retrieval.PdfDocument, error) {
	return m.returnDoc, m.err
}

type mockPdfStorage struct {
	saved map[string]string
}

func (m *mockPdfStorage) Save(ctx context.Context, paperID string, reader io.Reader) (string, error) {
	if m.saved == nil {
		m.saved = make(map[string]string)
	}
	path := "/mock/" + paperID + ".pdf"
	m.saved[paperID] = path
	return path, nil
}

func (m *mockPdfStorage) Open(ctx context.Context, paperID string) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("pdf content")), nil
}

func (m *mockPdfStorage) Exists(ctx context.Context, paperID string) bool {
	_, ok := m.saved[paperID]
	return ok
}

func (m *mockPdfStorage) GetFilePath(paperID string) string {
	return "/mock/" + paperID + ".pdf"
}

func TestIngestPdfCommandHandler_Success(t *testing.T) {
	mockDoc := &retrieval.PdfDocument{
		ID:         "test-transformer",
		FileName:   "transformer.pdf",
		Title:      "Attention Is All You Need",
		Authors:    []string{"Ashish Vaswani"},
		Abstract:   "Dominant sequence transduction models...",
		TotalPages: 11,
		Sections: []retrieval.ParsedSection{
			{ID: "sec-1", Type: retrieval.SectionAbstract, Title: "Abstract", Content: "Abstract content"},
			{ID: "sec-2", Type: retrieval.SectionResults, Title: "4. Results", Content: "BLEU score 28.4"},
		},
		References: []retrieval.ExtractedReference{
			{Index: 1, Title: "Layer Norm", Year: 2016},
		},
		ParsedAt: time.Now(),
	}

	parser := &mockPdfParser{returnDoc: mockDoc}
	storage := &mockPdfStorage{}
	handler := NewIngestPdfCommandHandler(parser, storage, nil, nil)

	cmd := IngestPdfCommand{
		FileName:  "transformer.pdf",
		FileBytes: []byte("%PDF-1.4 mock content"),
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	result, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	if !result.Success {
		t.Errorf("expected success true")
	}
	if result.PaperID != "test-transformer" {
		t.Errorf("expected paper ID 'test-transformer', got %q", result.PaperID)
	}
	if result.SectionCount != 2 {
		t.Errorf("expected 2 sections, got %d", result.SectionCount)
	}
	if result.RefCount != 1 {
		t.Errorf("expected 1 reference, got %d", result.RefCount)
	}
}

type mockDownloader struct {
	returnPath string
	err        error
}

func (m *mockDownloader) DownloadPdf(ctx context.Context, paperID string, downloadURL string) (string, error) {
	return m.returnPath, m.err
}

func (m *mockDownloader) DownloadBatch(ctx context.Context, papers []retrieval.PaperMetadata, maxConcurrency int) ([]interfaces.DownloadResult, error) {
	return nil, nil
}

func TestIngestPdfCommandHandler_DownloadURL_Success(t *testing.T) {
	mockDoc := &retrieval.PdfDocument{
		ID:         "arxiv_2301_0001",
		FileName:   "2301_0001.pdf",
		Title:      "Solid State Battery",
		TotalPages: 8,
		Sections: []retrieval.ParsedSection{
			{ID: "sec-1", Type: retrieval.SectionAbstract, Title: "Abstract", Content: "Battery abstract"},
		},
		ParsedAt: time.Now(),
	}

	parser := &mockPdfParser{returnDoc: mockDoc}
	storage := &mockPdfStorage{}
	downloader := &mockDownloader{returnPath: "/tmp/downloaded.pdf"}

	handler := NewIngestPdfCommandHandler(parser, storage, nil, nil, downloader)

	cmd := IngestPdfCommand{
		PaperID:     "arxiv_2301_0001",
		FileName:    "2301_0001.pdf",
		DownloadURL: "https://arxiv.org/pdf/2301.0001.pdf",
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	res, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}

	if res.PaperID != "arxiv_2301_0001" {
		t.Errorf("expected ID arxiv_2301_0001, got %s", res.PaperID)
	}
	if res.SectionCount != 1 {
		t.Errorf("expected 1 section, got %d", res.SectionCount)
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

func TestIngestPdfCommandHandler_DownloadFailed_FallbackToMetadata(t *testing.T) {
	downloader := &mockDownloader{err: retrieval.ErrParsingFailed}
	storage := &mockPdfStorage{}
	repo := &mockPaperRepo{}

	handler := NewIngestPdfCommandHandler(nil, storage, repo, nil, downloader)

	cmd := IngestPdfCommand{
		PaperID:     "https://openalex.org/W4386509670",
		FileName:    "w4386509670.pdf",
		DownloadURL: "https://file.sciopen.com/failed.pdf",
		Metadata: &retrieval.PaperMetadata{
			ID:       "https://openalex.org/W4386509670",
			Title:    "Weakly solvating electrolytes",
			Abstract: "Lithium batteries play a dominant role in modern power systems.",
		},
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	res, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected fallback success, got error: %v", err)
	}

	if res.PaperID != "https://openalex.org/W4386509670" {
		t.Errorf("expected ID https://openalex.org/W4386509670, got %s", res.PaperID)
	}
	if res.SectionCount != 1 {
		t.Errorf("expected 1 abstract section, got %d", res.SectionCount)
	}
	if res.Document == nil || res.Document.Sections[0].Type != retrieval.SectionAbstract {
		t.Errorf("expected SectionAbstract in fallback document")
	}
}

func TestIngestPdfCommandHandler_NoDownloadURL_FallbackToRepo(t *testing.T) {
	repo := &mockPaperRepo{
		papers: map[string]*retrieval.PaperMetadata{
			"https://openalex.org/W4386509670": {
				ID:       "https://openalex.org/W4386509670",
				Title:    "Weakly solvating electrolytes",
				Abstract: "Lithium batteries play a dominant role in modern power systems.",
			},
		},
	}

	handler := NewIngestPdfCommandHandler(nil, nil, repo, nil)

	cmd := IngestPdfCommand{
		PaperID:  "https://openalex.org/W4386509670",
		FileName: "w4386509670.pdf",
	}

	if err := cmd.Validate(); err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	res, err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected fallback success from repo, got error: %v", err)
	}

	if res.PaperID != "https://openalex.org/W4386509670" {
		t.Errorf("expected ID https://openalex.org/W4386509670, got %s", res.PaperID)
	}
	if res.SectionCount != 1 {
		t.Errorf("expected 1 abstract section, got %d", res.SectionCount)
	}
}
