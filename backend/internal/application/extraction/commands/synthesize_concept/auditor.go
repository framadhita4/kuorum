package synthesize_concept

import (
	"strings"

	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/retrieval"
)

// GroundingAuditor validates that citations in the synthesis contain authentic verbatim quotes from the paper sections.
type GroundingAuditor struct{}

func NewGroundingAuditor() *GroundingAuditor {
	return &GroundingAuditor{}
}

// Audit validates all definitions, mechanisms, and case studies against the paper's parsed sections.
func (a *GroundingAuditor) Audit(synthesis *extraction.GroundedConceptSynthesis, sections []retrieval.ParsedSection) AuditSummary {
	if synthesis == nil {
		return AuditSummary{AuditPassed: false}
	}

	// Build a normalized search corpus from all sections
	var corpusBuilder strings.Builder
	for _, sec := range sections {
		corpusBuilder.WriteString(sec.Title)
		corpusBuilder.WriteString(" ")
		corpusBuilder.WriteString(sec.Content)
		corpusBuilder.WriteString(" ")
	}
	normalizedCorpus := normalizeText(corpusBuilder.String())

	totalClaims := 0
	verifiedClaims := 0

	// 1. Audit Definitions
	for i := range synthesis.Definitions {
		totalClaims++
		if a.verifyQuote(synthesis.Definitions[i].Citation.VerbatimText, normalizedCorpus) {
			verifiedClaims++
		}
	}

	// 2. Audit Mechanisms
	for i := range synthesis.Mechanisms {
		totalClaims++
		if a.verifyQuote(synthesis.Mechanisms[i].Citation.VerbatimText, normalizedCorpus) {
			verifiedClaims++
		}
	}

	// 3. Audit Case Studies
	for i := range synthesis.CaseStudies {
		totalClaims++
		if a.verifyQuote(synthesis.CaseStudies[i].Citation.VerbatimText, normalizedCorpus) {
			verifiedClaims++
		}
	}

	rate := 1.0
	if totalClaims > 0 {
		rate = float64(verifiedClaims) / float64(totalClaims)
	}

	// Audit passes if at least 70% of quotes are verifiably grounded (allowing for minor token/OCR variance)
	auditPassed := totalClaims > 0 && rate >= 0.70

	return AuditSummary{
		TotalClaims:      totalClaims,
		VerifiedClaims:   verifiedClaims,
		VerificationRate: rate,
		AuditPassed:      auditPassed,
	}
}

func (a *GroundingAuditor) verifyQuote(verbatimQuote, normalizedCorpus string) bool {
	normQuote := normalizeText(verbatimQuote)
	if len(normQuote) < 10 {
		return false
	}
	return strings.Contains(normalizedCorpus, normQuote)
}

func normalizeText(t string) string {
	fields := strings.Fields(strings.ToLower(t))
	return strings.Join(fields, " ")
}
