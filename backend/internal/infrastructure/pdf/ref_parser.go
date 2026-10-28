package pdf

import (
	"regexp"
	"strconv"
	"strings"

	"kuorum/backend/internal/domain/retrieval"
)

var (
	// Numbered reference markers: [1], 1., (1)
	numberedRefRegex = regexp.MustCompile(`(?m)^(?:\[(\d+)\]|(\d+)\.|\((\d+)\))\s*`)

	// Publication year: (2023), 2023, 1998
	yearRegex = regexp.MustCompile(`\b(19\d{2}|20\d{2})[a-z]?\b`)

	// DOI regex: 10.1000/182
	doiRegex = regexp.MustCompile(`(?i)\b(10\.\d{4,9}/[-._;()/:A-Za-z0-9]+)`)

	// arXiv ID: arXiv:1706.03762 or 1706.03762
	arxivRegex = regexp.MustCompile(`(?i)(?:arxiv:\s*|arxiv\.org/abs/)(\d{4}\.\d{4,5})`)

	// Quoted title: "Attention is all you need" or “Attention is all you need”
	quotedTitleRegex = regexp.MustCompile(`["“]([^"”]+)["”]`)
)

// ParseReferences parses a raw bibliography text block into structured citation items.
func ParseReferences(rawReferences string) []retrieval.ExtractedReference {
	rawItems := splitReferenceItems(rawReferences)
	results := make([]retrieval.ExtractedReference, 0, len(rawItems))

	for idx, item := range rawItems {
		cleanItem := strings.TrimSpace(item)
		if cleanItem == "" || len(cleanItem) < 15 {
			continue
		}

		ref := parseSingleReference(idx+1, cleanItem)
		results = append(results, ref)
	}

	return results
}

func splitReferenceItems(text string) []string {
	// First check if text uses standard numbered brackets [1], [2]
	locs := numberedRefRegex.FindAllStringIndex(text, -1)
	if len(locs) >= 2 {
		items := make([]string, 0, len(locs))
		for i := 0; i < len(locs); i++ {
			start := locs[i][0]
			end := len(text)
			if i+1 < len(locs) {
				end = locs[i+1][0]
			}
			items = append(items, strings.TrimSpace(text[start:end]))
		}
		return items
	}

	// Fallback: Split by double newline or single newline if lines are substantive
	lines := strings.Split(text, "\n\n")
	if len(lines) >= 2 {
		return lines
	}

	// Last resort: line-by-line
	return strings.Split(text, "\n")
}

func parseSingleReference(index int, raw string) retrieval.ExtractedReference {
	ref := retrieval.ExtractedReference{
		Index:   index,
		RawText: raw,
	}

	// 1. Extract DOI if present
	if doiMatches := doiRegex.FindStringSubmatch(raw); len(doiMatches) > 1 {
		ref.DOI = strings.TrimRight(doiMatches[1], ".,;)")
	}

	// 2. Extract arXiv ID if present
	if arxivMatches := arxivRegex.FindStringSubmatch(raw); len(arxivMatches) > 1 {
		ref.ArXivID = arxivMatches[1]
	}

	// 3. Extract Year
	if yearMatches := yearRegex.FindAllStringSubmatch(raw, -1); len(yearMatches) > 0 {
		// Take the last or first valid 4-digit year
		for _, ym := range yearMatches {
			if len(ym) > 1 {
				if y, err := strconv.Atoi(ym[1]); err == nil && y >= 1900 && y <= 2030 {
					ref.Year = y
					break
				}
			}
		}
	}

	// 4. Extract Title if quoted
	if qMatches := quotedTitleRegex.FindStringSubmatch(raw); len(qMatches) > 1 {
		ref.Title = strings.TrimSpace(qMatches[1])
	}

	// 5. Clean up author guessing (text before title or year)
	cleaned := numberedRefRegex.ReplaceAllString(raw, "")
	parts := strings.Split(cleaned, ".")
	if len(parts) > 1 && ref.Title == "" {
		// First sentence or clause is usually authors
		authorsPart := strings.TrimSpace(parts[0])
		if len(authorsPart) > 0 && len(authorsPart) < 120 {
			ref.Authors = splitAuthors(authorsPart)
		}
		if len(parts) > 2 {
			ref.Title = strings.TrimSpace(parts[1])
		}
	} else if len(parts) > 0 {
		ref.Authors = splitAuthors(strings.TrimSpace(parts[0]))
	}

	return ref
}

func splitAuthors(authorStr string) []string {
	// Normalize "and" separator
	authorStr = strings.ReplaceAll(authorStr, " and ", ", ")
	rawAuthors := strings.Split(authorStr, ",")

	var authors []string
	for _, a := range rawAuthors {
		trimmed := strings.TrimSpace(a)
		if trimmed != "" && len(trimmed) > 1 && !strings.HasPrefix(trimmed, "[") {
			authors = append(authors, trimmed)
		}
	}
	return authors
}
