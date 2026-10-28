package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kuorum/backend/internal/api/response"
	"kuorum/backend/internal/application/common/mediator"
	"kuorum/backend/internal/application/retrieval/commands/ingest_pdf"
	"kuorum/backend/internal/application/retrieval/queries/get_citations"
	"kuorum/backend/internal/application/retrieval/queries/get_paper_details"
	"kuorum/backend/internal/application/retrieval/queries/get_pdf_sections"
	"kuorum/backend/internal/application/retrieval/queries/get_providers"
	"kuorum/backend/internal/application/retrieval/queries/get_references"
	"kuorum/backend/internal/application/retrieval/queries/search_papers"
	"kuorum/backend/internal/domain/retrieval"
	"kuorum/backend/internal/infrastructure/persistence"

	"github.com/go-chi/chi/v5"
)

type mockPaperApi struct{}

func (m *mockPaperApi) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	return []retrieval.PaperMetadata{
		{
			ID:    "W100",
			Title: "Test Paper " + query,
			Year:  2023,
		},
	}, nil
}

func (m *mockPaperApi) SearchPapersWithProviders(ctx context.Context, query string, limit int, providers []string) ([]retrieval.PaperMetadata, error) {
	return []retrieval.PaperMetadata{
		{
			ID:      "W100",
			Title:   "Test Paper " + query,
			Year:    2023,
			Sources: providers,
		},
	}, nil
}

func (m *mockPaperApi) GetPaperDetails(ctx context.Context, paperID string) (*retrieval.PaperMetadata, error) {
	if paperID == "notfound" {
		return nil, retrieval.ErrPaperNotFound
	}
	return &retrieval.PaperMetadata{
		ID:    paperID,
		Title: "Mock Title Details",
		Year:  2024,
	}, nil
}

func (m *mockPaperApi) GetReferences(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return []retrieval.PaperMetadata{
		{ID: "REF1", Title: "Reference 1", Year: 2020},
	}, nil
}

func (m *mockPaperApi) GetCitations(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	return []retrieval.PaperMetadata{
		{ID: "CIT1", Title: "Citation 1", Year: 2025},
	}, nil
}

func (m *mockPaperApi) AvailableProviders() []string {
	return []string{"openalex", "semanticscholar", "arxiv", "crossref"}
}

type mockPdfParserTest struct{}

func (m *mockPdfParserTest) Parse(ctx context.Context, reader io.Reader, filename string) (*retrieval.PdfDocument, error) {
	return &retrieval.PdfDocument{
		ID:         "mock-pdf-id",
		FileName:   filename,
		Title:      "Attention Is All You Need",
		TotalPages: 10,
		Sections: []retrieval.ParsedSection{
			{ID: "sec-1", Type: retrieval.SectionAbstract, Title: "Abstract", Content: "Abstract content"},
			{ID: "sec-2", Type: retrieval.SectionResults, Title: "Results", Content: "Results content"},
		},
	}, nil
}

func (m *mockPdfParserTest) ParseFile(ctx context.Context, filePath string) (*retrieval.PdfDocument, error) {
	return &retrieval.PdfDocument{
		ID:         "mock-pdf-id",
		FileName:   filePath,
		Title:      "Attention Is All You Need",
		TotalPages: 10,
		Sections: []retrieval.ParsedSection{
			{ID: "sec-1", Type: retrieval.SectionAbstract, Title: "Abstract", Content: "Abstract content"},
			{ID: "sec-2", Type: retrieval.SectionResults, Title: "Results", Content: "Results content"},
		},
	}, nil
}

func setupTestMediator() mediator.Mediator {
	med := mediator.NewInMemoryMediator()
	mock := &mockPaperApi{}
	mockParser := &mockPdfParserTest{}
	testCache := persistence.NewMemoryCache()

	med.RegisterQueryHandler(search_papers.SearchPapersQuery{}, search_papers.NewSearchPapersQueryHandler(mock, nil))
	med.RegisterQueryHandler(get_paper_details.GetPaperDetailsQuery{}, get_paper_details.NewGetPaperDetailsQueryHandler(mock, nil))
	med.RegisterQueryHandler(get_references.GetReferencesQuery{}, get_references.NewGetReferencesQueryHandler(mock))
	med.RegisterQueryHandler(get_citations.GetCitationsQuery{}, get_citations.NewGetCitationsQueryHandler(mock))
	med.RegisterQueryHandler(get_providers.GetProvidersQuery{}, get_providers.NewGetProvidersQueryHandler(mock))
	med.RegisterCommandHandler(ingest_pdf.IngestPdfCommand{}, ingest_pdf.NewIngestPdfCommandHandler(mockParser, nil, nil, testCache))
	med.RegisterQueryHandler(get_pdf_sections.GetPdfSectionsQuery{}, get_pdf_sections.NewGetPdfSectionsQueryHandler(testCache, mockParser, nil))

	return med
}

func TestPaperController_Search(t *testing.T) {
	med := setupTestMediator()
	ctrl := NewPaperController(med)

	t.Run("GET Search with query params", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/search?query=neural+networks&limit=5&providers=openalex,arxiv", nil)
		rr := httptest.NewRecorder()

		ctrl.Search(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}

		var env response.Envelope
		if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to decode envelope: %v", err)
		}
		if !env.Success {
			t.Errorf("expected success true")
		}
	})

	t.Run("POST Search with JSON body", func(t *testing.T) {
		body := `{"query":"transformer models","limit":3,"providers":["arxiv"]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/papers/search", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		ctrl.Search(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})
}

func TestPaperController_GetDetails(t *testing.T) {
	med := setupTestMediator()
	ctrl := NewPaperController(med)

	r := chi.NewRouter()
	r.Get("/api/v1/papers/{id}", ctrl.GetDetails)
	r.Get("/api/v1/papers/details", ctrl.GetDetails)

	t.Run("by query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/details?id=W123", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("by path param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/W123", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("missing ID returns 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/details", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rr.Code)
		}
	})
}

func TestPaperController_GetProviders(t *testing.T) {
	med := setupTestMediator()
	ctrl := NewPaperController(med)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/providers", nil)
	rr := httptest.NewRecorder()

	ctrl.GetProviders(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var env response.Envelope
	_ = json.Unmarshal(rr.Body.Bytes(), &env)
	dataMap, ok := env.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map data")
	}
	provs, ok := dataMap["providers"].([]any)
	if !ok || len(provs) != 4 {
		t.Fatalf("expected 4 providers, got %+v", provs)
	}
}

func TestPaperController_GetReferencesAndCitations(t *testing.T) {
	med := setupTestMediator()
	ctrl := NewPaperController(med)

	t.Run("references", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/references?id=W123", nil)
		rr := httptest.NewRecorder()
		ctrl.GetReferences(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})

	t.Run("citations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/citations?id=W123", nil)
		rr := httptest.NewRecorder()
		ctrl.GetCitations(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rr.Code)
		}
	})
}

func TestPaperController_UploadPdfAndSections(t *testing.T) {
	med := setupTestMediator()
	ctrl := NewPaperController(med)

	t.Run("Upload PDF multipart form", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "attention.pdf")
		if err != nil {
			t.Fatalf("create form file failed: %v", err)
		}
		_, _ = part.Write([]byte("%PDF-1.4 test mock content"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/papers/upload-pdf", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		ctrl.UploadPdf(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (body: %s)", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		_ = json.Unmarshal(rr.Body.Bytes(), &env)
		if !env.Success {
			t.Fatalf("expected success true in envelope")
		}
		dataMap, ok := env.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected map data in upload result")
		}
		if dataMap["paperId"] != "attention" {
			t.Errorf("expected paperId 'attention', got %v", dataMap["paperId"])
		}
	})

	t.Run("Get Sections with intent query", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/api/v1/papers/{id}/sections", ctrl.GetSections)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/papers/attention/sections?intent=results", nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (body: %s)", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		_ = json.Unmarshal(rr.Body.Bytes(), &env)
		if !env.Success {
			t.Fatalf("expected success true in envelope")
		}
	})

	t.Run("Upload PDF with unsupported characters and spaces sanitizes paperId", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "Deep Residual Learning (2015): ResNet?.pdf")
		if err != nil {
			t.Fatalf("create form file failed: %v", err)
		}
		_, _ = part.Write([]byte("%PDF-1.4 test mock content"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/papers/upload-pdf", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		ctrl.UploadPdf(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d (body: %s)", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		_ = json.Unmarshal(rr.Body.Bytes(), &env)
		if !env.Success {
			t.Fatalf("expected success true in envelope")
		}
		dataMap, ok := env.Data.(map[string]any)
		if !ok {
			t.Fatalf("expected map data in upload result")
		}
		expectedID := "deep-residual-learning-2015-resnet"
		if dataMap["paperId"] != expectedID {
			t.Errorf("expected paperId %q, got %v", expectedID, dataMap["paperId"])
		}
	})

	t.Run("Upload PDF with invalid unsupported filename returns 400", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "???.pdf")
		if err != nil {
			t.Fatalf("create form file failed: %v", err)
		}
		_, _ = part.Write([]byte("%PDF-1.4 test mock content"))
		_ = writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/papers/upload-pdf", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		ctrl.UploadPdf(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d (body: %s)", rr.Code, rr.Body.String())
		}

		var env response.Envelope
		_ = json.Unmarshal(rr.Body.Bytes(), &env)
		if env.Success {
			t.Errorf("expected success false in envelope")
		}
	})
}
