package synthesize_concept

// SynthesizeConceptCommand triggers verbatim-grounded extraction of academic definitions, mechanisms, and case studies.
type SynthesizeConceptCommand struct {
	AnchorPaperID string   `json:"anchorPaperId"`
	PaperIDs      []string `json:"paperIds,omitempty"`
	Concept       string   `json:"concept"`
	Intent        string   `json:"intent,omitempty"`
	ForceRefresh  bool     `json:"forceRefresh,omitempty"`
}

func (c SynthesizeConceptCommand) IsCommand() {}

