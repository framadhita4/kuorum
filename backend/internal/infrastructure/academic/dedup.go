package academic

import (
	"math"
	"sort"
	"strings"

	"kuorum/backend/internal/domain/retrieval"
)

// DeduplicateAndMerge consolidates papers without a query, sorting by citations then recency.
func DeduplicateAndMerge(papers []retrieval.PaperMetadata) []retrieval.PaperMetadata {
	return DeduplicateAndRank(papers, "")
}

// DeduplicateAndRank consolidates papers and ranks them using multi-factor relevance scoring
// (query term coverage, title matching, abstract matching, and citation log tie-breaker).
func DeduplicateAndRank(papers []retrieval.PaperMetadata, query string) []retrieval.PaperMetadata {
	if len(papers) <= 1 {
		return papers
	}

	merged := make([]retrieval.PaperMetadata, 0, len(papers))
	doiIndex := make(map[string]int)
	titleIndex := make(map[string]int)

	for _, paper := range papers {
		cleanDOI := strings.ToLower(strings.TrimSpace(paper.Doi))
		cleanTitle := normalizeTitleKey(paper.Title)

		targetIdx := -1
		if cleanDOI != "" {
			if idx, found := doiIndex[cleanDOI]; found {
				targetIdx = idx
			}
		}

		if targetIdx == -1 && len(cleanTitle) >= 8 {
			if idx, found := titleIndex[cleanTitle]; found {
				targetIdx = idx
			}
		}

		if targetIdx >= 0 {
			// Merge into existing paper
			existing := merged[targetIdx]
			merged[targetIdx] = mergePapers(existing, paper)

			// Update indices in case new keys emerged
			if merged[targetIdx].Doi != "" {
				doiIndex[strings.ToLower(merged[targetIdx].Doi)] = targetIdx
			}
			newTitleKey := normalizeTitleKey(merged[targetIdx].Title)
			if len(newTitleKey) >= 8 {
				titleIndex[newTitleKey] = targetIdx
			}
		} else {
			// Add new entry
			idx := len(merged)
			merged = append(merged, paper)
			if cleanDOI != "" {
				doiIndex[cleanDOI] = idx
			}
			if len(cleanTitle) >= 8 {
				titleIndex[cleanTitle] = idx
			}
		}
	}

	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		// Default fallback: sort by citation count descending, then year
		sort.SliceStable(merged, func(i, j int) bool {
			if merged[i].CitationCount != merged[j].CitationCount {
				return merged[i].CitationCount > merged[j].CitationCount
			}
			return merged[i].Year > merged[j].Year
		})
		return merged
	}

	// Calculate semantic & term relevance score for each merged paper
	stems := extractQueryStems(trimmedQuery)
	cleanQueryLower := strings.ToLower(trimmedQuery)
	scores := make(map[string]float64, len(merged))

	for i := range merged {
		scores[merged[i].ID] = calculateRelevanceScore(merged[i], cleanQueryLower, stems)
	}

	sort.SliceStable(merged, func(i, j int) bool {
		scoreI := scores[merged[i].ID]
		scoreJ := scores[merged[j].ID]

		// If scores differ significantly, rank by relevance score
		if math.Abs(scoreI-scoreJ) > 0.05 {
			return scoreI > scoreJ
		}

		// Tie-breaking: citations, then year
		if merged[i].CitationCount != merged[j].CitationCount {
			return merged[i].CitationCount > merged[j].CitationCount
		}
		return merged[i].Year > merged[j].Year
	})

	return merged
}

// calculateRelevanceScore evaluates how closely a paper matches the search query.
func calculateRelevanceScore(paper retrieval.PaperMetadata, cleanQuery string, stems []string) float64 {
	titleLower := strings.ToLower(paper.Title)
	tldrLower := strings.ToLower(paper.TLDR)
	abstractLower := strings.ToLower(paper.Abstract)

	var titleScore float64
	var abstractScore float64
	var matchedStems int

	// 1. Exact phrase match bonus
	if len(cleanQuery) >= 4 {
		if strings.Contains(titleLower, cleanQuery) {
			titleScore += 30.0
		}
		if strings.Contains(abstractLower, cleanQuery) || strings.Contains(tldrLower, cleanQuery) {
			abstractScore += 12.0
		}
	}

	// 2. Individual query term matching & coverage
	for _, stem := range stems {
		inTitle := isStemInText(titleLower, stem)
		inTLDR := isStemInText(tldrLower, stem)
		inAbstract := isStemInText(abstractLower, stem)

		if inTitle || inTLDR || inAbstract {
			matchedStems++
		}

		if inTitle {
			titleScore += 8.0
		}
		if inTLDR {
			abstractScore += 4.0
		}
		if inAbstract {
			cnt := strings.Count(abstractLower, stem)
			if cnt > 3 {
				cnt = 3
			}
			abstractScore += float64(cnt) * 2.0
		}
	}

	// 3. Query Term Coverage penalty (steep penalty if core concepts are missing)
	coverage := 1.0
	if len(stems) > 0 {
		coverage = float64(matchedStems) / float64(len(stems))
	}
	coverageMultiplier := math.Pow(coverage, 2.5)

	// 4. Logarithmic citation bonus (serves as credibility booster, capped so it never overrides text relevance)
	citationScore := math.Log10(float64(paper.CitationCount+1)) * 1.5
	if citationScore > 8.0 {
		citationScore = 8.0
	}

	// 5. Gentle recency bonus for recent work
	recencyScore := 0.0
	if paper.Year >= 2020 {
		recencyScore = float64(paper.Year-2019) * 0.4
	}

	rawScore := titleScore + abstractScore + citationScore + recencyScore
	return rawScore * coverageMultiplier
}

// extractQueryStems decomposes a search query into normalized semantic roots.
func extractQueryStems(query string) []string {
	clean := strings.ToLower(query)
	words := NonAlphaNumRegex.Split(clean, -1)
	stems := make([]string, 0)
	seen := make(map[string]bool)

	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) < 2 || IsStopWord(w) {
			continue
		}
		stem := stemWord(w)
		if !seen[stem] {
			seen[stem] = true
			stems = append(stems, stem)
		}
	}
	return stems
}

// stemWord applies simple English suffix stripping to capture word variants (e.g. caching/cached/cache).
func stemWord(w string) string {
	if len(w) > 5 && strings.HasSuffix(w, "ing") {
		return strings.TrimSuffix(w, "ing") // caching -> cach
	}
	if len(w) > 4 && strings.HasSuffix(w, "ies") {
		return strings.TrimSuffix(w, "ies") + "y"
	}
	if len(w) > 4 && strings.HasSuffix(w, "es") {
		return strings.TrimSuffix(w, "es") // caches -> cach
	}
	if len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss") {
		return strings.TrimSuffix(w, "s") // models -> model
	}
	if len(w) > 4 && strings.HasSuffix(w, "ed") {
		return strings.TrimSuffix(w, "ed") // cached -> cach
	}
	return w
}

func isStemInText(text, stem string) bool {
	if strings.Contains(text, stem) {
		return true
	}
	// Common domain synonyms
	if stem == "yolo" && strings.Contains(text, "you only look once") {
		return true
	}
	return false
}

func normalizeTitleKey(title string) string {
	lower := strings.ToLower(strings.TrimSpace(title))
	return NonAlphaNumRegex.ReplaceAllString(lower, "")
}

func mergePapers(base, incoming retrieval.PaperMetadata) retrieval.PaperMetadata {
	res := base

	// Title: choose the one with better capitalization or length
	if len(incoming.Title) > len(res.Title) || res.Title == "" {
		res.Title = incoming.Title
	}

	// Abstract: choose the longer, more complete abstract
	if len(incoming.Abstract) > len(res.Abstract) {
		res.Abstract = incoming.Abstract
	}

	// TLDR: take if base doesn't have it
	if res.TLDR == "" && incoming.TLDR != "" {
		res.TLDR = incoming.TLDR
	}

	// Year: take non-zero
	if res.Year == 0 && incoming.Year > 0 {
		res.Year = incoming.Year
	}

	// Citation counts: take the maximum observed
	if incoming.CitationCount > res.CitationCount {
		res.CitationCount = incoming.CitationCount
	}
	if incoming.InfluentialCitationCount > res.InfluentialCitationCount {
		res.InfluentialCitationCount = incoming.InfluentialCitationCount
	}

	// DOI: keep non-empty
	if res.Doi == "" && incoming.Doi != "" {
		res.Doi = incoming.Doi
	}

	// Open Access PDF URL: take if not present
	if res.OpenAccessPdfURL == "" && incoming.OpenAccessPdfURL != "" {
		res.OpenAccessPdfURL = incoming.OpenAccessPdfURL
	}

	// Venue: take if not present or more descriptive
	if res.Venue == "" || res.Venue == "arXiv preprint" {
		if incoming.Venue != "" && incoming.Venue != "arXiv preprint" {
			res.Venue = incoming.Venue
		} else if res.Venue == "" {
			res.Venue = incoming.Venue
		}
	}

	// Authors: merge without duplicates
	authorMap := make(map[string]bool)
	for _, a := range res.Authors {
		authorMap[strings.ToLower(strings.TrimSpace(a.Name))] = true
	}
	for _, a := range incoming.Authors {
		nameKey := strings.ToLower(strings.TrimSpace(a.Name))
		if !authorMap[nameKey] && a.Name != "" {
			authorMap[nameKey] = true
			res.Authors = append(res.Authors, a)
		}
	}

	// Fields of study: merge union
	fieldMap := make(map[string]bool)
	for _, f := range res.FieldsOfStudy {
		fieldMap[strings.ToLower(strings.TrimSpace(f))] = true
	}
	for _, f := range incoming.FieldsOfStudy {
		key := strings.ToLower(strings.TrimSpace(f))
		if !fieldMap[key] && f != "" {
			fieldMap[key] = true
			res.FieldsOfStudy = append(res.FieldsOfStudy, f)
		}
	}

	// Sources: union
	sourceMap := make(map[string]bool)
	for _, s := range res.Sources {
		sourceMap[s] = true
	}
	for _, s := range incoming.Sources {
		if !sourceMap[s] && s != "" {
			sourceMap[s] = true
			res.Sources = append(res.Sources, s)
		}
	}

	return res
}
