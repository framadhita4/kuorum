package get_citations

type GetCitationsQuery struct {
	ID string `json:"id"`
}

func (q GetCitationsQuery) IsQuery() {}
