package reformulate_query

type ReformulateQueryCommand struct {
	AngleID string `json:"angleId"`
	Prompt  string `json:"prompt"`
}

func (c ReformulateQueryCommand) IsCommand() {}
