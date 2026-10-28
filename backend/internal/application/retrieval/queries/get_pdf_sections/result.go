package get_pdf_sections

import "kuorum/backend/internal/domain/retrieval"

type GetPdfSectionsResult struct {
	PaperID        string                    `json:"paperId"`
	Intent         string                    `json:"intent,omitempty"`
	AnchorSection  *retrieval.ParsedSection  `json:"anchorSection,omitempty"`
	TargetSections []retrieval.ParsedSection `json:"targetSections"`
	AllSections    []retrieval.ParsedSection `json:"allSections,omitempty"`
	TotalSections  int                       `json:"totalSections"`
}
