package get_graph

// GetCitationGraphQuery retrieves an existing citation graph by anchor paper ID.
type GetCitationGraphQuery struct {
	AnchorPaperID string `json:"anchorPaperId"`
}

func (q GetCitationGraphQuery) IsQuery() {}

