package extraction

import (
	"errors"
	"time"
)

var (
	ErrUngroundedClaim = errors.New("claim lacks verifiable citation evidence")
	ErrEmptySynthesis  = errors.New("synthesis content cannot be empty")
)

type CitationProof struct {
	PaperID      string `json:"paperId"`
	VerbatimText string `json:"verbatimText"`
	Section      string `json:"section,omitempty"`
	PageNumber   int    `json:"pageNumber,omitempty"`
}

type GroundedDefinition struct {
	Term       string        `json:"term"`
	Definition string        `json:"definition"`
	Citation   CitationProof `json:"citation"`
}

type GroundedMechanism struct {
	StepNumber  int           `json:"stepNumber"`
	Description string        `json:"description"`
	Citation    CitationProof `json:"citation"`
}

type GroundedCaseStudy struct {
	IndustryDomain string        `json:"industryDomain"`
	ProblemSolved  string        `json:"problemSolved"`
	MethodApplied  string        `json:"methodApplied"`
	Outcome        string        `json:"outcome"`
	Citation       CitationProof `json:"citation"`
}

type GroundedConceptSynthesis struct {
	AnchorPaperID string               `json:"anchorPaperId"`
	Concept       string               `json:"concept"`
	Definitions   []GroundedDefinition `json:"definitions"`
	Mechanisms    []GroundedMechanism  `json:"mechanisms"`
	CaseStudies   []GroundedCaseStudy  `json:"caseStudies"`
	GeneratedAt   time.Time            `json:"generatedAt"`
}
