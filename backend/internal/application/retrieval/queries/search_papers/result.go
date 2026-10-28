package search_papers

import "kuorum/backend/internal/domain/retrieval"

type SearchPapersResult struct {
	Query     string                    `json:"query"`
	Total     int                       `json:"total"`
	Providers []string                  `json:"providers,omitempty"`
	Papers    []retrieval.PaperMetadata `json:"papers"`
}
