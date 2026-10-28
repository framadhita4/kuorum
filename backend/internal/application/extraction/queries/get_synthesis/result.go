package get_synthesis

import (
	"kuorum/backend/internal/domain/extraction"
)

type GetSynthesisResult struct {
	Synthesis *extraction.GroundedConceptSynthesis `json:"synthesis"`
}
