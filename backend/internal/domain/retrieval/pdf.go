package retrieval

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidPdf      = errors.New("invalid or corrupt PDF file")
	ErrEmptyPdf        = errors.New("PDF contains no extractable text")
	ErrEncryptedPdf    = errors.New("PDF is password protected or encrypted")
	ErrParsingFailed   = errors.New("failed to parse PDF document")
	ErrSectionNotFound = errors.New("requested section not found in document")
)

// SectionType categorizes academic sections according to canonical scientific paper taxonomy.
type SectionType string

const (
	SectionAbstract     SectionType = "ABSTRACT"
	SectionIntroduction SectionType = "INTRODUCTION"
	SectionRelatedWork  SectionType = "RELATED_WORK"
	SectionMethodology  SectionType = "METHODOLOGY"
	SectionResults      SectionType = "RESULTS"
	SectionDiscussion   SectionType = "DISCUSSION"
	SectionConclusion   SectionType = "CONCLUSION"
	SectionReferences   SectionType = "REFERENCES"
	SectionAppendix     SectionType = "APPENDIX"
	SectionOther        SectionType = "OTHER"
)

// ParsedSection represents a semantically segmented block of an academic paper.
type ParsedSection struct {
	ID         string      `json:"id"`
	Type       SectionType `json:"type"`
	Title      string      `json:"title"`
	Content    string      `json:"content"`
	Paragraphs []string    `json:"paragraphs"`
	StartPage  int         `json:"startPage"`
	EndPage    int         `json:"endPage"`
	OrderIndex int         `json:"orderIndex"`
	WordCount  int         `json:"wordCount"`
}

// ExtractedReference represents an individual cited work parsed from the bibliography.
type ExtractedReference struct {
	Index   int      `json:"index"`
	RawText string   `json:"rawText"`
	Authors []string `json:"authors,omitempty"`
	Title   string   `json:"title,omitempty"`
	Year    int      `json:"year,omitempty"`
	Venue   string   `json:"venue,omitempty"`
	DOI     string   `json:"doi,omitempty"`
	ArXivID string   `json:"arxivId,omitempty"`
}

// PdfDocument represents a fully parsed academic paper with structured sections and references.
type PdfDocument struct {
	ID            string               `json:"id"`
	FileName      string               `json:"fileName"`
	Title         string               `json:"title"`
	Authors       []string             `json:"authors,omitempty"`
	Abstract      string               `json:"abstract,omitempty"`
	DOI           string               `json:"doi,omitempty"`
	TotalPages    int                  `json:"totalPages"`
	Sections      []ParsedSection      `json:"sections"`
	References    []ExtractedReference `json:"references"`
	RawTextLength int                  `json:"rawTextLength"`
	ParsedAt      time.Time            `json:"parsedAt"`
}

// GetSection returns the first section matching the requested SectionType, or nil if not found.
func (d *PdfDocument) GetSection(st SectionType) *ParsedSection {
	for i := range d.Sections {
		if d.Sections[i].Type == st {
			return &d.Sections[i]
		}
	}
	return nil
}

// GetSections returns all sections matching any of the requested SectionTypes.
func (d *PdfDocument) GetSections(types ...SectionType) []ParsedSection {
	typeMap := make(map[SectionType]bool, len(types))
	for _, t := range types {
		typeMap[t] = true
	}

	var results []ParsedSection
	for _, s := range d.Sections {
		if typeMap[s.Type] {
			results = append(results, s)
		}
	}
	return results
}

// GetTargetSectionsForIntent routes the user's research query intent to the optimal section bundle.
// Returns an anchor section (typically Abstract) for global terminology context, plus the focused target sections.
func (d *PdfDocument) GetTargetSectionsForIntent(intent string) (anchor *ParsedSection, targets []ParsedSection) {
	anchor = d.GetSection(SectionAbstract)

	normalizedIntent := strings.ToLower(strings.TrimSpace(intent))
	switch {
	case strings.Contains(normalizedIntent, "result") ||
		strings.Contains(normalizedIntent, "finding") ||
		strings.Contains(normalizedIntent, "experiment") ||
		strings.Contains(normalizedIntent, "empirical") ||
		strings.Contains(normalizedIntent, "evaluat"):
		targets = d.GetSections(SectionResults, SectionDiscussion)

	case strings.Contains(normalizedIntent, "definit") ||
		strings.Contains(normalizedIntent, "theorem") ||
		strings.Contains(normalizedIntent, "theory") ||
		strings.Contains(normalizedIntent, "preliminar") ||
		strings.Contains(normalizedIntent, "proof"):
		targets = d.GetSections(SectionIntroduction, SectionMethodology)

	case strings.Contains(normalizedIntent, "mechanism") ||
		strings.Contains(normalizedIntent, "architect") ||
		strings.Contains(normalizedIntent, "how") ||
		strings.Contains(normalizedIntent, "algorithm") ||
		strings.Contains(normalizedIntent, "method") ||
		strings.Contains(normalizedIntent, "approach"):
		targets = d.GetSections(SectionMethodology)

	case strings.Contains(normalizedIntent, "prior") ||
		strings.Contains(normalizedIntent, "related") ||
		strings.Contains(normalizedIntent, "baseline") ||
		strings.Contains(normalizedIntent, "reference") ||
		strings.Contains(normalizedIntent, "citation"):
		targets = d.GetSections(SectionRelatedWork, SectionReferences)

	case strings.Contains(normalizedIntent, "conclu") ||
		strings.Contains(normalizedIntent, "future"):
		targets = d.GetSections(SectionConclusion)

	default:
		// Fallback: Return all non-reference sections
		for _, s := range d.Sections {
			if s.Type != SectionReferences && s.Type != SectionAbstract {
				targets = append(targets, s)
			}
		}
	}

	// If targeted sections are empty (e.g. paper had no explicit results heading), return all content
	if len(targets) == 0 {
		for _, s := range d.Sections {
			if s.Type != SectionReferences && s.Type != SectionAbstract {
				targets = append(targets, s)
			}
		}
	}

	return anchor, targets
}
