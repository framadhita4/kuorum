package get_angles

type AngleSummaryDTO struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GetAnglesResult struct {
	Concept string            `json:"concept"`
	Angles  []AngleSummaryDTO `json:"angles"`
}
