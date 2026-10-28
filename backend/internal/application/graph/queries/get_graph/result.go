package get_graph

import (
	"kuorum/backend/internal/domain/graph"
)

type GetCitationGraphResult struct {
	Graph *graph.CitationGraph `json:"graph"`
}
