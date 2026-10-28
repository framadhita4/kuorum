package build_graph

import (
	"kuorum/backend/internal/domain/graph"
)

type BuildCitationGraphResult struct {
	AnchorPaperID     string            `json:"anchorPaperId"`
	Nodes             []graph.GraphNode `json:"nodes"`
	Edges             []graph.GraphEdge `json:"edges"`
	TotalNodes        int               `json:"totalNodes"`
	TotalEdges        int               `json:"totalEdges"`
	FoundationalCount int               `json:"foundationalCount"`
	ApplicationCount  int               `json:"applicationCount"`
	CohortCount       int               `json:"cohortCount"`
	FromCache         bool              `json:"fromCache"`
}
