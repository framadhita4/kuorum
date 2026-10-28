package get_references

import "kuorum/backend/internal/domain/retrieval"

type GetReferencesResult struct {
	PaperID    string                    `json:"paperId"`
	Total      int                       `json:"total"`
	References []retrieval.PaperMetadata `json:"references"`
}
