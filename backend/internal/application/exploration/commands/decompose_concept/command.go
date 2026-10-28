package decompose_concept

type DecomposeConceptCommand struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
}

func (c DecomposeConceptCommand) IsCommand() {}
