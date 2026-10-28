package ingest_pdf

import "kuorum/backend/internal/domain/retrieval"

type IngestPdfResult struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	PaperID      string                 `json:"paperId"`
	Document     *retrieval.PdfDocument `json:"document"`
	SavedPath    string                 `json:"savedPath,omitempty"`
	TotalPages   int                    `json:"totalPages"`
	SectionCount int                    `json:"sectionCount"`
	RefCount     int                    `json:"refCount"`
}
