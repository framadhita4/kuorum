package interfaces

import (
	"context"
	"io"

	"kuorum/backend/internal/domain/retrieval"
)

// IPdfParser abstracts high-fidelity academic PDF parsing and layout segmentation.
type IPdfParser interface {
	Parse(ctx context.Context, reader io.Reader, filename string) (*retrieval.PdfDocument, error)
	ParseFile(ctx context.Context, filePath string) (*retrieval.PdfDocument, error)
}

// IPdfStorage handles local persistent storage of uploaded PDF documents.
type IPdfStorage interface {
	Save(ctx context.Context, paperID string, reader io.Reader) (string, error)
	Open(ctx context.Context, paperID string) (io.ReadCloser, error)
	Exists(ctx context.Context, paperID string) bool
	GetFilePath(paperID string) string
}

// DownloadResult captures the status of an automated PDF download attempt.
type DownloadResult struct {
	PaperID   string
	FilePath  string
	BytesRead int64
	Error     error
}

// IPdfDownloader manages robust, concurrent downloading of Open Access academic PDFs.
type IPdfDownloader interface {
	DownloadPdf(ctx context.Context, paperID string, downloadURL string) (string, error)
	DownloadBatch(ctx context.Context, papers []retrieval.PaperMetadata, maxConcurrency int) ([]DownloadResult, error)
}
