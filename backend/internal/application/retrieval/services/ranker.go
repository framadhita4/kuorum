package services

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode"

	"kuorum/backend/internal/domain/retrieval"
)

// RankerWeights defines the relative weights for composite candidate paper ranking.
type RankerWeights struct {
	Relevance  float64
	Authority  float64
	Recency    float64
	OpenAccess float64
}

// DefaultRankerWeights provides standard academic research balance.
var DefaultRankerWeights = RankerWeights{
	Relevance:  0.45,
	Authority:  0.25,
	Recency:    0.15,
	OpenAccess: 0.15,
}

// RankerOptions configures the filtering and scoring parameters.
type RankerOptions struct {
	Weights           RankerWeights
	MinAbstractLength int
	RequireOpenAccess bool
	CurrentYear       int
	MaxBenchmarkCites int
}

// DefaultRankerOptions provides sensible default options.
func DefaultRankerOptions() RankerOptions {
	return RankerOptions{
		Weights:           DefaultRankerWeights,
		MinAbstractLength: 50,
		RequireOpenAccess: false,
		CurrentYear:       time.Now().Year(),
		MaxBenchmarkCites: 1000,
	}
}

// RankedPaper wraps PaperMetadata with composite and dimensional score breakdowns.
type RankedPaper struct {
	Paper           retrieval.PaperMetadata `json:"paper"`
	CompositeScore  float64                 `json:"compositeScore"`
	TotalPoints     float64                 `json:"totalPoints"` // Calculation point on a 0 - 100 scale
	RelevanceScore  float64                 `json:"relevanceScore"`
	AuthorityScore  float64                 `json:"authorityScore"`
	RecencyScore    float64                 `json:"recencyScore"`
	OpenAccessScore float64                 `json:"openAccessScore"`
	Rank            int                     `json:"rank"`
	IsAnchor        bool                    `json:"isAnchor"`
	Formula         string                  `json:"formula,omitempty"`
}

// RankResult holds the designated Anchor Paper and all ranked candidate papers.
type RankResult struct {
	AnchorPaper *RankedPaper  `json:"anchorPaper"`
	Candidates  []RankedPaper `json:"candidates"`
	TotalPassed int           `json:"totalPassed"`
	TotalRaw    int           `json:"totalRaw"`
}

// SelectCohort picks the top K diverse, high-quality candidate papers for deep analysis.
func (res *RankResult) SelectCohort(maxSelect int) []retrieval.PaperMetadata {
	if maxSelect <= 0 {
		maxSelect = 4
	}
	if len(res.Candidates) == 0 {
		return nil
	}
	if maxSelect > len(res.Candidates) {
		maxSelect = len(res.Candidates)
	}

	var selected []retrieval.PaperMetadata
	selectedTitles := make(map[string]bool)

	for _, cand := range res.Candidates {
		normTitle := strings.ToLower(strings.TrimSpace(cand.Paper.Title))
		if selectedTitles[normTitle] {
			continue
		}
		selectedTitles[normTitle] = true
		selected = append(selected, cand.Paper)
		if len(selected) >= maxSelect {
			break
		}
	}

	return selected
}

// AnchorPaperRanker evaluates, filters, and ranks candidate papers.
type AnchorPaperRanker struct {
	options RankerOptions
}

// NewAnchorPaperRanker creates a ranker instance with customized or default options.
func NewAnchorPaperRanker(opts ...RankerOptions) *AnchorPaperRanker {
	options := DefaultRankerOptions()
	if len(opts) > 0 {
		options = opts[0]
	}
	return &AnchorPaperRanker{options: options}
}

// Rank evaluates a slice of candidate papers against the user's research query.
func (r *AnchorPaperRanker) Rank(query string, rawPapers []retrieval.PaperMetadata) RankResult {
	result := RankResult{
		TotalRaw: len(rawPapers),
	}

	if len(rawPapers) == 0 {
		return result
	}

	queryTokens := r.tokenizeQuery(query)

	// Deduplicate papers by DOI and normalized title before scoring
	seenDOIs := make(map[string]bool)
	seenTitles := make(map[string]bool)
	var deduplicated []retrieval.PaperMetadata

	for _, p := range rawPapers {
		cleanDOI := strings.ToLower(strings.TrimSpace(p.Doi))
		cleanTitle := strings.ToLower(strings.TrimSpace(p.Title))
		if cleanDOI != "" && seenDOIs[cleanDOI] {
			continue
		}
		if len(cleanTitle) >= 8 && seenTitles[cleanTitle] {
			continue
		}
		if cleanDOI != "" {
			seenDOIs[cleanDOI] = true
		}
		if len(cleanTitle) >= 8 {
			seenTitles[cleanTitle] = true
		}
		deduplicated = append(deduplicated, p)
	}

	var scoredPapers []RankedPaper
	for _, p := range deduplicated {
		// Tier 1: Quality Gate & Hard Rules
		if !r.passesQualityGate(p) {
			continue
		}

		relScore := r.computeRelevance(queryTokens, p)
		authScore := r.computeAuthority(p)
		recScore := r.computeRecency(p)
		oaScore := r.computeOpenAccess(p)

		w := r.options.Weights
		composite := (w.Relevance * relScore) +
			(w.Authority * authScore) +
			(w.Recency * recScore) +
			(w.OpenAccess * oaScore)

		// Round to 4 decimal places for precision consistency
		composite = math.Round(composite*10000) / 10000
		totalPoints := math.Round(composite*100*10) / 10

		formula := fmt.Sprintf("(0.45 * %.2f) + (0.25 * %.2f) + (0.15 * %.2f) + (0.15 * %.2f) = %.1f pts",
			relScore, authScore, recScore, oaScore, totalPoints)

		ranked := RankedPaper{
			Paper:           p,
			CompositeScore:  composite,
			TotalPoints:     totalPoints,
			RelevanceScore:  math.Round(relScore*10000) / 10000,
			AuthorityScore:  math.Round(authScore*10000) / 10000,
			RecencyScore:    math.Round(recScore*10000) / 10000,
			OpenAccessScore: math.Round(oaScore*10000) / 10000,
			Formula:         formula,
		}
		scoredPapers = append(scoredPapers, ranked)
	}

	if len(scoredPapers) == 0 {
		return result
	}

	// Sort descending by CompositeScore, break ties with CitationCount
	sort.SliceStable(scoredPapers, func(i, j int) bool {
		if scoredPapers[i].CompositeScore == scoredPapers[j].CompositeScore {
			return scoredPapers[i].Paper.CitationCount > scoredPapers[j].Paper.CitationCount
		}
		return scoredPapers[i].CompositeScore > scoredPapers[j].CompositeScore
	})

	// Assign ranks and designate Anchor Paper (Rank 1)
	for i := range scoredPapers {
		scoredPapers[i].Rank = i + 1
		if i == 0 {
			scoredPapers[i].IsAnchor = true
		}
	}

	result.AnchorPaper = &scoredPapers[0]
	result.Candidates = scoredPapers
	result.TotalPassed = len(scoredPapers)

	return result
}

// passesQualityGate checks Tier 1 minimum criteria.
func (r *AnchorPaperRanker) passesQualityGate(p retrieval.PaperMetadata) bool {
	if strings.TrimSpace(p.Title) == "" {
		return false
	}

	// Filter out obvious non-academic entries
	lowerTitle := strings.ToLower(p.Title)
	if strings.HasPrefix(lowerTitle, "table of contents") ||
		strings.HasPrefix(lowerTitle, "author index") ||
		strings.HasPrefix(lowerTitle, "editorial board") ||
		strings.HasPrefix(lowerTitle, "index of") {
		return false
	}

	// Abstract must meet minimum length, or TLDR must be present, or high-citation seminal paper
	abstractLen := len(strings.TrimSpace(p.Abstract))
	tldrLen := len(strings.TrimSpace(p.TLDR))
	if abstractLen < r.options.MinAbstractLength && tldrLen < 20 {
		if p.CitationCount < 50 {
			return false
		}
	}

	if r.options.RequireOpenAccess && !r.hasOpenAccess(p) {
		return false
	}

	return true
}

func (r *AnchorPaperRanker) hasOpenAccess(p retrieval.PaperMetadata) bool {
	if strings.TrimSpace(p.OpenAccessPdfURL) != "" {
		return true
	}
	if strings.Contains(strings.ToLower(p.Doi), "arxiv") {
		return true
	}
	return false
}

// computeRelevance scores lexical and semantic overlap with query tokens and keyphrases.
func (r *AnchorPaperRanker) computeRelevance(queryTokens []string, p retrieval.PaperMetadata) float64 {
	if len(queryTokens) == 0 {
		return 0.5 // Neutral baseline if query has no extractable keywords
	}

	titleTokens := r.tokenizeQuery(p.Title)
	abstractTokens := r.tokenizeQuery(p.Abstract + " " + p.TLDR)

	titleHits := 0
	abstractHits := 0

	for _, qt := range queryTokens {
		if containsToken(titleTokens, qt) {
			titleHits++
		}
		if containsToken(abstractTokens, qt) {
			abstractHits++
		}
	}

	titleScore := float64(titleHits) / float64(len(queryTokens))
	abstractScore := float64(abstractHits) / float64(len(queryTokens))

	// Base overlap: Title is weighted at 70%, Abstract at 30%
	combined := (titleScore * 0.70) + (abstractScore * 0.30)

	// Phrase Matching Bonus (Bigrams in Title & Abstract)
	if len(queryTokens) >= 2 {
		lowerTitle := strings.ToLower(p.Title)
		lowerAbstract := strings.ToLower(p.Abstract + " " + p.TLDR)
		phraseHits := 0
		totalPhrases := len(queryTokens) - 1

		for i := 0; i < len(queryTokens)-1; i++ {
			phrase := queryTokens[i] + " " + queryTokens[i+1]
			if strings.Contains(lowerTitle, phrase) {
				phraseHits += 2
			} else if strings.Contains(lowerAbstract, phrase) {
				phraseHits++
			}
		}

		if totalPhrases > 0 {
			phraseBonus := math.Min(0.20, (float64(phraseHits)/float64(totalPhrases*2))*0.20)
			combined += phraseBonus
		}
	}

	// Facet Richness Bonus (Rewards papers containing methodology, mechanics, and empirical results)
	facetBonus := 0.0
	lowerFull := strings.ToLower(p.Title + " " + p.Abstract + " " + p.TLDR)
	hasMethodology := strings.Contains(lowerFull, "method") || strings.Contains(lowerFull, "architecture") ||
		strings.Contains(lowerFull, "algorithm") || strings.Contains(lowerFull, "model") || strings.Contains(lowerFull, "framework")
	hasEmpirical := strings.Contains(lowerFull, "experiment") || strings.Contains(lowerFull, "benchmark") ||
		strings.Contains(lowerFull, "evaluation") || strings.Contains(lowerFull, "result") || strings.Contains(lowerFull, "accuracy")
	hasTheory := strings.Contains(lowerFull, "theory") || strings.Contains(lowerFull, "foundation") ||
		strings.Contains(lowerFull, "principle") || strings.Contains(lowerFull, "mechanism")

	if hasMethodology {
		facetBonus += 0.03
	}
	if hasEmpirical {
		facetBonus += 0.03
	}
	if hasTheory {
		facetBonus += 0.02
	}
	combined += facetBonus

	if combined > 1.0 {
		combined = 1.0
	}
	return combined
}

// computeAuthority scores log-normalized citation count + influential citation bonus.
func (r *AnchorPaperRanker) computeAuthority(p retrieval.PaperMetadata) float64 {
	cites := math.Max(0, float64(p.CitationCount))
	benchmark := math.Max(100, float64(r.options.MaxBenchmarkCites))

	// Log-scale normalization: ln(1 + cites) / ln(1 + benchmark)
	logScore := math.Log1p(cites) / math.Log1p(benchmark)
	if logScore > 1.0 {
		logScore = 1.0
	}

	// Influential citation boost (up to 0.1 bonus)
	if p.InfluentialCitationCount > 0 {
		infBonus := math.Min(0.1, float64(p.InfluentialCitationCount)*0.02)
		logScore += infBonus
		if logScore > 1.0 {
			logScore = 1.0
		}
	}

	return logScore
}

// computeRecency scores proximity to current year with a 15-year sliding decay.
func (r *AnchorPaperRanker) computeRecency(p retrieval.PaperMetadata) float64 {
	if p.Year <= 0 {
		return 0.3 // Default baseline for papers with unknown year
	}

	currentYear := r.options.CurrentYear
	if p.Year >= currentYear {
		return 1.0
	}

	diff := currentYear - p.Year
	if diff <= 0 {
		return 1.0
	}
	if diff >= 15 {
		return 0.1 // Minimum floor so seminal older papers are not completely zeroed out
	}

	return 1.0 - (float64(diff) / 15.0 * 0.9)
}

// computeOpenAccess checks if the paper has immediate full-text download capability.
func (r *AnchorPaperRanker) computeOpenAccess(p retrieval.PaperMetadata) float64 {
	if r.hasOpenAccess(p) {
		return 1.0
	}
	return 0.0
}

func containsToken(tokens []string, target string) bool {
	for _, t := range tokens {
		if t == target {
			return true
		}
	}
	return false
}

// tokenizeQuery splits text into lowercase alphanumeric tokens and strips common stop words.
func (r *AnchorPaperRanker) tokenizeQuery(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	var filtered []string
	for _, word := range fields {
		if len(word) < 3 || isStopWord(word) {
			continue
		}
		filtered = append(filtered, word)
	}
	return filtered
}

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "that": true, "this": true,
	"with": true, "from": true, "have": true, "were": true, "which": true,
	"what": true, "when": true, "where": true, "how": true, "why": true,
	"about": true, "into": true, "paper": true, "study": true, "research": true,
	"using": true, "based": true, "analysis": true, "towards": true, "review": true,
}

func isStopWord(w string) bool {
	return stopWords[w]
}
