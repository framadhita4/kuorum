package reformulate_query

type ReformulateQueryResult struct {
	OriginalPrompt string   `json:"originalPrompt"`
	Keywords       []string `json:"keywords"`
	Synonyms       []string `json:"synonyms"`
	BooleanQuery   string   `json:"booleanQuery"`
	DenseQuery     string   `json:"denseQuery"`
}
