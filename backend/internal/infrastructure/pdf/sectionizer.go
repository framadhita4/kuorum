package pdf

import (
	"fmt"
	"regexp"
	"strings"

	"kuorum/backend/internal/domain/retrieval"
)

// Regex definitions for academic headings
var (
	// Numbered or Roman numeral prefixes: "1.", "1 ", "1.1 ", "I.", "IV. "
	headingPrefixRegex = `^(?:(?:[0-9]+(?:\.[0-9]+)*\.?|[IVXLCDM]+\.)\s*)?`

	abstractRegex    = regexp.MustCompile(`(?i)^(?:abstract|abstrak)(?:\s*[-—–:])?`)
	articleInfoRegex = regexp.MustCompile(`(?i)^(?:article\s*info|article\s*information|article\s*history)\b`)
	introRegex       = regexp.MustCompile(`(?i)` + headingPrefixRegex + `introduction\b`)
	relatedRegex     = regexp.MustCompile(`(?i)` + headingPrefixRegex + `(?:related work|prior work|literature review|background|preliminaries)\b`)
	methodsRegex     = regexp.MustCompile(`(?i)` + headingPrefixRegex + `(?:methodology|methods?|(?:the\s+)?(?:proposed\s+)?(?:method|approach|architecture|model)|system model|system architecture|model architecture|formulation|algorithm|framework)\b`)
	resultsRegex     = regexp.MustCompile(`(?i)` + headingPrefixRegex + `(?:experiments?|evaluations?|results?(?:\s+and\s+discussion)?|empirical (?:evaluation|results)|benchmarks?)\b`)
	discussionRegex  = regexp.MustCompile(`(?i)` + headingPrefixRegex + `(?:discussions?|ablation (?:study|experiments)|limitations)\b`)
	conclusionRegex  = regexp.MustCompile(`(?i)` + headingPrefixRegex + `(?:conclusions?(?:\s+and\s+future\s+work)?|concluding remarks|summary)\b`)
	refsRegex        = regexp.MustCompile(`(?i)` + headingPrefixRegex + `(?:references|bibliography|works cited)\b`)
	appendixRegex    = regexp.MustCompile(`(?i)` + headingPrefixRegex + `appendix\b`)

	// General heading test: concise line that looks like a title
	generalHeadingRegex = regexp.MustCompile(`^(?:(?:[0-9]+(?:\.[0-9]+)*\.?|[IVXLCDM]+\.)\s+)[A-Z][A-Za-z0-9\s,–—-]+$`)
)

type rawPage struct {
	pageNumber int
	text       string
}

// SectionizePages splits cleaned page texts into structured academic sections.
func SectionizePages(pages []rawPage) []retrieval.ParsedSection {
	var sections []retrieval.ParsedSection
	var currentSection *retrieval.ParsedSection
	orderIndex := 0

	startNewSection := func(st retrieval.SectionType, title string, pageNum int) {
		if currentSection != nil && (len(currentSection.Paragraphs) > 0 || currentSection.Content != "") {
			finalizeSection(currentSection)
			sections = append(sections, *currentSection)
		}
		orderIndex++
		currentSection = &retrieval.ParsedSection{
			ID:         fmt.Sprintf("sec-%d", orderIndex),
			Type:       st,
			Title:      title,
			StartPage:  pageNum,
			EndPage:    pageNum,
			OrderIndex: orderIndex,
			Paragraphs: []string{},
		}
	}

	for _, page := range pages {
		paragraphs := strings.Split(page.text, "\n\n")

		for _, para := range paragraphs {
			trimmedPara := strings.TrimSpace(para)
			if trimmedPara == "" {
				continue
			}

			// Check if paragraph is or begins with a section heading
			headingType, headingTitle, remainingBody := detectHeading(trimmedPara)

			if headingType != "" {
				if headingType == retrieval.SectionAbstract && currentSection != nil && currentSection.Type == retrieval.SectionAbstract {
					// Dual-language abstract (e.g. English Abstract followed by Indonesian Abstrak)
					currentSection.Paragraphs = append(currentSection.Paragraphs, headingTitle)
					if remainingBody != "" {
						currentSection.Paragraphs = append(currentSection.Paragraphs, remainingBody)
					}
				} else {
					startNewSection(headingType, headingTitle, page.pageNumber)
					if remainingBody != "" {
						currentSection.Paragraphs = append(currentSection.Paragraphs, remainingBody)
					}
				}
			} else {
				if currentSection == nil {
					// Preamble before first detected heading (often title / authors / affiliations)
					startNewSection(retrieval.SectionOther, "Title & Preamble", page.pageNumber)
				}
				currentSection.Paragraphs = append(currentSection.Paragraphs, trimmedPara)
				currentSection.EndPage = page.pageNumber
			}
		}
	}

	if currentSection != nil {
		finalizeSection(currentSection)
		sections = append(sections, *currentSection)
	}

	return sections
}

func finalizeSection(s *retrieval.ParsedSection) {
	s.Content = strings.Join(s.Paragraphs, "\n\n")
	words := strings.Fields(s.Content)
	s.WordCount = len(words)
}

// detectHeading checks whether a paragraph contains or is a heading.
// Returns (SectionType, Title, remainingTextIfAny).
func detectHeading(para string) (retrieval.SectionType, string, string) {
	lines := strings.Split(para, "\n")
	firstLine := strings.TrimSpace(lines[0])

	// Unspace wide kerning: e.g. "A B S T R A C T" -> "ABSTRACT", "A B S T R A K" -> "ABSTRAK", "1. I N T R O D U C T I O N" -> "1. INTRODUCTION"
	cleanFirstLine := UnspaceKerning(firstLine)

	// First handle "Abstract—..." or "Abstract: ..." inline
	if abstractRegex.MatchString(cleanFirstLine) {
		title := cleanFirstLine
		if len(title) > 40 {
			title = "Abstract"
		}
		remaining := ""
		loc := abstractRegex.FindStringIndex(cleanFirstLine)
		if loc != nil && len(cleanFirstLine) > loc[1] {
			inlineRest := strings.TrimSpace(cleanFirstLine[loc[1]:])
			inlineRest = strings.TrimLeft(inlineRest, "—–- :")
			if inlineRest != "" {
				if len(lines) > 1 {
					remaining = inlineRest + "\n" + strings.Join(lines[1:], "\n")
				} else {
					remaining = inlineRest
				}
			}
		} else if len(lines) > 1 {
			remaining = strings.Join(lines[1:], "\n")
		}
		return retrieval.SectionAbstract, title, strings.TrimSpace(remaining)
	}

	// For standard headings, check if the first line is short enough to be a heading (<= 90 chars)
	if len(cleanFirstLine) <= 90 {
		st := classifyHeading(cleanFirstLine)
		if st != "" {
			remaining := ""
			if len(lines) > 1 {
				remaining = strings.Join(lines[1:], "\n")
			}
			return st, cleanFirstLine, strings.TrimSpace(remaining)
		}
	}

	return "", "", ""
}

func classifyHeading(line string) retrieval.SectionType {
	clean := strings.TrimRight(line, ":. ")

	switch {
	case abstractRegex.MatchString(clean):
		return retrieval.SectionAbstract
	case articleInfoRegex.MatchString(clean):
		return retrieval.SectionOther
	case introRegex.MatchString(clean):
		return retrieval.SectionIntroduction
	case relatedRegex.MatchString(clean):
		return retrieval.SectionRelatedWork
	case methodsRegex.MatchString(clean):
		return retrieval.SectionMethodology
	case resultsRegex.MatchString(clean):
		return retrieval.SectionResults
	case discussionRegex.MatchString(clean):
		return retrieval.SectionDiscussion
	case conclusionRegex.MatchString(clean):
		return retrieval.SectionConclusion
	case refsRegex.MatchString(clean):
		return retrieval.SectionReferences
	case appendixRegex.MatchString(clean):
		return retrieval.SectionAppendix
	case generalHeadingRegex.MatchString(clean):
		return retrieval.SectionOther
	default:
		return ""
	}
}
