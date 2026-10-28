package search_papers

type SearchPapersQuery struct {
	Query     string   `json:"query"`
	Limit     int      `json:"limit"`
	Providers []string `json:"providers,omitempty"`
}

func (q SearchPapersQuery) IsQuery() {}
