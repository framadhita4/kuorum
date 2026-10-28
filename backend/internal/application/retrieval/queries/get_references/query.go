package get_references

type GetReferencesQuery struct {
	ID string `json:"id"`
}

func (q GetReferencesQuery) IsQuery() {}
