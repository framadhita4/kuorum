package exploration

import (
	"errors"
	"time"
)

var (
	ErrEmptyConceptPrompt    = errors.New("concept prompt cannot be empty")
	ErrPromptTooShort        = errors.New("concept prompt must be at least 3 characters")
	ErrAngleNotFound         = errors.New("research angle not found")
	ErrEmptyDecomposedAngles = errors.New("no research angles could be decomposed")
)

// ResearchAngle represents an academic dimension or perspective of a research concept.
type ResearchAngle struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	SuggestedQueries []string `json:"suggestedQueries"`
}

// AcademicQuery represents an optimized search query targeting literature repositories.
type AcademicQuery struct {
	Keywords []string `json:"keywords"`
	Synonyms []string `json:"synonyms"`
	Boolean  string   `json:"boolean"`
	Dense    string   `json:"dense"`
}

// ConceptDecomposition represents the domain aggregate of decomposed research angles.
type ConceptDecomposition struct {
	OriginalPrompt string          `json:"originalPrompt"`
	NormalizedTerm string          `json:"normalizedTerm"`
	Angles         []ResearchAngle `json:"angles"`
	CreatedAt      time.Time       `json:"createdAt"`
}

// NewConceptDecomposition creates a new ConceptDecomposition aggregate.
func NewConceptDecomposition(prompt, normalized string, angles []ResearchAngle) (*ConceptDecomposition, error) {
	if len(prompt) < 3 {
		return nil, ErrPromptTooShort
	}
	if len(angles) == 0 {
		return nil, ErrEmptyDecomposedAngles
	}

	return &ConceptDecomposition{
		OriginalPrompt: prompt,
		NormalizedTerm: normalized,
		Angles:         angles,
		CreatedAt:      time.Now().UTC(),
	}, nil
}
