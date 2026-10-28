package pdf

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	// Matches standalone page number lines like "1", "Page 2 of 12", "3 / 15", "[12]"
	pageNumberRegex = regexp.MustCompile(`(?i)^(?:page\s+)?(?:\[?\d+\]?)(?:\s*(?:of|/)\s*\d+)?$`)

	// Matches lines ending in a word character and a hyphen (split words)
	trailingHyphenRegex = regexp.MustCompile(`([a-zA-Z])-+$`)

	// Common publication header noise
	pubNoticeRegex = regexp.MustCompile(`(?i)^(preprint|under review|accepted as|proceedings of|conference on|arxiv:\d+\.\d+|ieee|acm).*`)
)

// CleanPageText cleans running headers, footers, standalone page numbers, and joins broken lines.
func CleanPageText(rawPage string, pageNum int) string {
	lines := strings.Split(rawPage, "\n")
	if len(lines) == 0 {
		return ""
	}

	cleanedLines := make([]string, 0, len(lines))

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			cleanedLines = append(cleanedLines, "")
			continue
		}

		// Check if first or last line is a page number or publication stamp
		isTopOrBottom := (i <= 2 || i >= len(lines)-3)
		if isTopOrBottom {
			if pageNumberRegex.MatchString(trimmed) || pubNoticeRegex.MatchString(trimmed) {
				continue
			}
		}

		cleanedLines = append(cleanedLines, trimmed)
	}

	return assembleParagraphs(cleanedLines)
}

// UnspaceKerning normalizes wide letter-spaced headings like "A B S T R A C T" -> "ABSTRACT",
// "A B S T R A K" -> "ABSTRAK", "A R T I C L E   I N F O" -> "ARTICLE INFO",
// "1. I N T R O D U C T I O N" -> "1. INTRODUCTION".
func UnspaceKerning(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return ""
	}

	// Extract optional leading numbering prefix like "1. ", "IV. "
	var prefix string
	prefixRegex := regexp.MustCompile(`^(?:(?:[0-9]+(?:\.[0-9]+)*\.?|[IVXLCDM]+\.)\s*)`)
	if loc := prefixRegex.FindStringIndex(trimmed); loc != nil && loc[0] == 0 {
		prefix = trimmed[:loc[1]]
		trimmed = strings.TrimSpace(trimmed[loc[1]:])
	}

	// Split words on double/multiple spaces or tabs
	parts := regexp.MustCompile(`\s{2,}|\t+`).Split(trimmed, -1)
	var newParts []string
	for _, part := range parts {
		p := strings.TrimSpace(part)
		tokens := strings.Fields(p)
		isAllSingleChar := len(tokens) >= 3
		for _, tok := range tokens {
			if len([]rune(tok)) != 1 {
				isAllSingleChar = false
				break
			}
		}
		if isAllSingleChar {
			newParts = append(newParts, strings.Join(tokens, ""))
		} else {
			newParts = append(newParts, p)
		}
	}

	return prefix + strings.Join(newParts, " ")
}

// assembleParagraphs joins wrapped single lines within a paragraph and de-hyphenates split words.
func assembleParagraphs(lines []string) string {
	var paragraphs []string
	var currentPara strings.Builder

	for _, line := range lines {
		if line == "" {
			if currentPara.Len() > 0 {
				paragraphs = append(paragraphs, strings.TrimSpace(currentPara.String()))
				currentPara.Reset()
			}
			continue
		}

		if currentPara.Len() == 0 {
			currentPara.WriteString(line)
			continue
		}

		currText := currentPara.String()
		if trailingHyphenRegex.MatchString(currText) && len(line) > 0 && unicode.IsLower(rune(line[0])) {
			// De-hyphenate: remove trailing hyphen and concatenate directly
			newText := trailingHyphenRegex.ReplaceAllString(currText, "$1")
			currentPara.Reset()
			currentPara.WriteString(newText)
			currentPara.WriteString(line)
		} else {
			// Normal line continuation within a paragraph: join with single space
			currentPara.WriteString(" ")
			currentPara.WriteString(line)
		}
	}

	if currentPara.Len() > 0 {
		paragraphs = append(paragraphs, strings.TrimSpace(currentPara.String()))
	}

	return strings.Join(paragraphs, "\n\n")
}
