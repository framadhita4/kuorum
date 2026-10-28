package pdf_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"kuorum/backend/internal/domain/retrieval"
	"kuorum/backend/internal/infrastructure/pdf"
)

func TestPaperDownloader_ResolveDownloadURL(t *testing.T) {
	storage, err := pdf.NewLocalPdfStorage(t.TempDir())
	if err != nil {
		t.Fatalf("creating storage failed: %v", err)
	}
	downloader := pdf.NewPaperDownloader(storage, http.DefaultClient)

	tests := []struct {
		name     string
		paper    retrieval.PaperMetadata
		expected string
	}{
		{
			name: "Direct OpenAccess URL",
			paper: retrieval.PaperMetadata{
				OpenAccessPdfURL: "https://example.com/paper.pdf",
			},
			expected: "https://example.com/paper.pdf",
		},
		{
			name: "arXiv DOI resolution",
			paper: retrieval.PaperMetadata{
				Doi: "10.48550/arXiv.2301.12345",
			},
			expected: "https://arxiv.org/pdf/2301.12345.pdf",
		},
		{
			name: "arXiv source tag resolution",
			paper: retrieval.PaperMetadata{
				Sources: []string{"arxiv:2305.67890"},
			},
			expected: "https://arxiv.org/pdf/2305.67890.pdf",
		},
		{
			name: "arXiv abs URL in ID",
			paper: retrieval.PaperMetadata{
				ID: "http://arxiv.org/abs/2509.02013v1",
			},
			expected: "https://arxiv.org/pdf/2509.02013v1.pdf",
		},
		{
			name:     "No downloadable URL",
			paper:    retrieval.PaperMetadata{},
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := downloader.ResolveDownloadURL(tc.paper)
			if got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestPaperDownloader_DownloadPdf_Success(t *testing.T) {
	// Mock HTTP server returning valid PDF content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("%PDF-1.5 \n%binary data stream here\n%%EOF"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	storage, err := pdf.NewLocalPdfStorage(tempDir)
	if err != nil {
		t.Fatalf("creating storage failed: %v", err)
	}

	downloader := pdf.NewPaperDownloader(storage, server.Client())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	savedPath, err := downloader.DownloadPdf(ctx, "paper-test-1", server.URL)
	if err != nil {
		t.Fatalf("download failed unexpectedly: %v", err)
	}

	if _, err := os.Stat(savedPath); os.IsNotExist(err) {
		t.Fatalf("expected saved file at %q, but not found", savedPath)
	}

	// Re-download should skip and return existing path
	rePath, err := downloader.DownloadPdf(ctx, "paper-test-1", server.URL)
	if err != nil {
		t.Fatalf("re-download failed: %v", err)
	}
	if rePath != savedPath {
		t.Errorf("expected cached path %q, got %q", savedPath, rePath)
	}
}

func TestPaperDownloader_DownloadPdf_InvalidMagicBytes(t *testing.T) {
	// Mock HTTP server returning HTML paywall instead of PDF
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Please log in to purchase access to this article.</body></html>"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	storage, err := pdf.NewLocalPdfStorage(tempDir)
	if err != nil {
		t.Fatalf("creating storage failed: %v", err)
	}

	downloader := pdf.NewPaperDownloader(storage, server.Client())

	ctx := context.Background()
	_, err = downloader.DownloadPdf(ctx, "paper-paywall", server.URL)
	if err == nil {
		t.Fatal("expected error due to missing PDF magic bytes, got nil")
	}
}

func TestPaperDownloader_DownloadBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Base(r.URL.Path) == "fail.pdf" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("%PDF-1.4 mock content"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	storage, err := pdf.NewLocalPdfStorage(tempDir)
	if err != nil {
		t.Fatalf("creating storage failed: %v", err)
	}

	downloader := pdf.NewPaperDownloader(storage, server.Client())

	papers := []retrieval.PaperMetadata{
		{
			ID:               "paper-ok-1",
			OpenAccessPdfURL: server.URL + "/ok1.pdf",
		},
		{
			ID:               "paper-fail",
			OpenAccessPdfURL: server.URL + "/fail.pdf",
		},
		{
			ID:               "paper-no-url",
			OpenAccessPdfURL: "",
		},
	}

	ctx := context.Background()
	results, err := downloader.DownloadBatch(ctx, papers, 2)
	if err != nil {
		t.Fatalf("batch download returned unexpected fatal error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// First should succeed
	if results[0].Error != nil || results[0].FilePath == "" {
		t.Errorf("expected paper-ok-1 to succeed, got err=%v", results[0].Error)
	}

	// Second should fail with HTTP error
	if results[1].Error == nil {
		t.Errorf("expected paper-fail to return an error, got nil")
	}

	// Third should fail with no download URL
	if results[2].Error != pdf.ErrNoDownloadURL {
		t.Errorf("expected ErrNoDownloadURL for paper-no-url, got %v", results[2].Error)
	}
}
