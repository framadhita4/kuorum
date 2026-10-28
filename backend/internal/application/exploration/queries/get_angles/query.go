package get_angles

type GetAnglesQuery struct {
	Concept string `json:"concept"`
}

func (q GetAnglesQuery) IsQuery() {}
