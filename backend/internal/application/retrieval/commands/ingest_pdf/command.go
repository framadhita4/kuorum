package ingest_pdf

import "kuorum/backend/internal/domain/retrieval"

// IngestPdfCommand encapsulates input parameters to ingest and parse an academic PDF.
type IngestPdfCommand struct {
	PaperID     string                   `json:"paperId,omitempty"`
	FileName    string                   `json:"fileName"`
	FilePath    string                   `json:"filePath,omitempty"`
	DownloadURL string                   `json:"downloadUrl,omitempty"`
	FileBytes   []byte                   `json:"-"`
	Metadata    *retrieval.PaperMetadata `json:"metadata,omitempty"`
}

func (c IngestPdfCommand) IsCommand() {}
