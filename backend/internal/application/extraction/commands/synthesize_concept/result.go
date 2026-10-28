package synthesize_concept

import (
	"kuorum/backend/internal/domain/extraction"
)

type AuditSummary struct {
	TotalClaims      int     `json:"totalClaims"`
	VerifiedClaims   int     `json:"verifiedClaims"`
	VerificationRate float64 `json:"verificationRate"`
	AuditPassed      bool    `json:"auditPassed"`
}

type SynthesizeConceptResult struct {
	Synthesis    *extraction.GroundedConceptSynthesis `json:"synthesis"`
	Audit        AuditSummary                         `json:"audit"`
	FromCache    bool                                 `json:"fromCache"`
	SectionCount int                                  `json:"sectionCount"`
}
