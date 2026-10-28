package get_synthesis

// GetSynthesisQuery retrieves cached grounded synthesis for a given anchor paper and concept.
type GetSynthesisQuery struct {
	AnchorPaperID string `json:"anchorPaperId"`
	Concept       string `json:"concept"`
}

func (q GetSynthesisQuery) IsQuery() {}

