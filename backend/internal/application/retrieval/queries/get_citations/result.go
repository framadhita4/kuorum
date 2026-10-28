package get_citations

import "kuorum/backend/internal/domain/retrieval"

type GetCitationsResult struct {
	PaperID   string                    `json:"paperId"`
	Total     int                       `json:"total"`
	Citations []retrieval.PaperMetadata `json:"citations"`
}
