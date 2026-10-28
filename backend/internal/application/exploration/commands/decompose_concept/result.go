package decompose_concept

type AngleDTO struct {
	AngleID          string   `json:"angleId"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	SuggestedQueries []string `json:"suggestedQueries"`
}

type DecomposeConceptResult struct {
	Prompt         string     `json:"prompt"`
	NormalizedTerm string     `json:"normalizedTerm"`
	Angles         []AngleDTO `json:"angles"`
}
