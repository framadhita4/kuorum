package get_paper_details

type GetPaperDetailsQuery struct {
	ID string `json:"id"`
}

func (q GetPaperDetailsQuery) IsQuery() {}
