package build_graph

// BuildCitationGraphCommand initiates bidirectional citation graph construction around an anchor paper and its selected cohort.
type BuildCitationGraphCommand struct {
	AnchorPaperID  string   `json:"anchorPaperId"`
	CohortPaperIDs []string `json:"cohortPaperIds,omitempty"`
	MaxReferences  int      `json:"maxReferences,omitempty"`
	MaxCitations   int      `json:"maxCitations,omitempty"`
	ForceRefresh   bool     `json:"forceRefresh,omitempty"`
}

func (c BuildCitationGraphCommand) IsCommand() {}

