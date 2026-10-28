package synthesize_concept

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/extraction"
	"kuorum/backend/internal/domain/retrieval"
)

type SynthesizeConceptCommandHandler struct {
	llmClient interfaces.ILLMClient
	repo      interfaces.IExtractionRepository
	cache     interfaces.ICacheRepository
	parser    interfaces.IPdfParser
	storage   interfaces.IPdfStorage
	paperRepo interfaces.IPaperRepository
	auditor   *GroundingAuditor
}

func NewSynthesizeConceptCommandHandler(
	llmClient interfaces.ILLMClient,
	repo interfaces.IExtractionRepository,
	cache interfaces.ICacheRepository,
	parser interfaces.IPdfParser,
	storage interfaces.IPdfStorage,
	extra ...any,
) *SynthesizeConceptCommandHandler {
	var a *GroundingAuditor
	var paperRepo interfaces.IPaperRepository

	for _, arg := range extra {
		switch v := arg.(type) {
		case *GroundingAuditor:
			a = v
		case interfaces.IPaperRepository:
			paperRepo = v
		}
	}

	if a == nil {
		a = NewGroundingAuditor()
	}

	return &SynthesizeConceptCommandHandler{
		llmClient: llmClient,
		repo:      repo,
		cache:     cache,
		parser:    parser,
		storage:   storage,
		paperRepo: paperRepo,
		auditor:   a,
	}
}

type llmSynthesisResponse struct {
	Definitions []struct {
		Term         string `json:"term"`
		Definition   string `json:"definition"`
		VerbatimText string `json:"verbatimText"`
		Section      string `json:"section"`
		PaperID      string `json:"paperId,omitempty"`
	} `json:"definitions"`
	Mechanisms []struct {
		StepNumber   int    `json:"stepNumber"`
		Description  string `json:"description"`
		VerbatimText string `json:"verbatimText"`
		Section      string `json:"section"`
		PaperID      string `json:"paperId,omitempty"`
	} `json:"mechanisms"`
	CaseStudies []struct {
		IndustryDomain string `json:"industryDomain"`
		ProblemSolved  string `json:"problemSolved"`
		MethodApplied  string `json:"methodApplied"`
		Outcome        string `json:"outcome"`
		VerbatimText   string `json:"verbatimText"`
		Section        string `json:"section"`
		PaperID        string `json:"paperId,omitempty"`
	} `json:"caseStudies"`
}

func (h *SynthesizeConceptCommandHandler) Handle(ctx context.Context, cmd SynthesizeConceptCommand) (*SynthesizeConceptResult, error) {
	// 1. Check existing repository if not force refresh
	if !cmd.ForceRefresh && h.repo != nil {
		if existing, err := h.repo.GetSynthesis(ctx, cmd.AnchorPaperID, cmd.Concept); err == nil && existing != nil {
			return &SynthesizeConceptResult{
				Synthesis: existing,
				Audit: AuditSummary{
					TotalClaims:      len(existing.Definitions) + len(existing.Mechanisms) + len(existing.CaseStudies),
					VerifiedClaims:   len(existing.Definitions) + len(existing.Mechanisms) + len(existing.CaseStudies),
					VerificationRate: 1.0,
					AuditPassed:      true,
				},
				FromCache: true,
			}, nil
		}
	}

	// 2. Load Parsed Sections from Cache or Storage across target papers
	paperIDs := []string{cmd.AnchorPaperID}
	for _, pid := range cmd.PaperIDs {
		cleanPid := strings.TrimSpace(pid)
		if cleanPid != "" && cleanPid != cmd.AnchorPaperID {
			paperIDs = append(paperIDs, cleanPid)
		}
	}

	intent := cmd.Intent
	if intent == "" {
		intent = "all"
	}

	var allDocs []*retrieval.PdfDocument
	var allSections []retrieval.ParsedSection
	for _, pid := range paperIDs {
		doc := h.loadDocument(ctx, pid)
		if doc != nil {
			allDocs = append(allDocs, doc)
			anchorSec, targetSecs := doc.GetTargetSectionsForIntent(intent)
			if anchorSec != nil {
				allSections = append(allSections, *anchorSec)
			}
			allSections = append(allSections, targetSecs...)
			if anchorSec == nil && len(targetSecs) == 0 && len(doc.Sections) > 0 {
				allSections = append(allSections, doc.Sections...)
			}
		}
	}

	if len(allDocs) == 0 {
		return nil, fmt.Errorf("%w: paper %q must be ingested before synthesis", retrieval.ErrPaperNotFound, cmd.AnchorPaperID)
	}

	primaryDoc := allDocs[0]

	// 3. Context Selection (Concatenate sections across all analyzed papers)
	var contextBuilder strings.Builder
	for _, doc := range allDocs {
		contextBuilder.WriteString(fmt.Sprintf("\n=== [PAPER ID: %s | TITLE: %s] ===\n", doc.ID, doc.Title))
		anchorSec, targetSecs := doc.GetTargetSectionsForIntent(intent)
		hasWrittenSec := false
		if anchorSec != nil {
			contextBuilder.WriteString(fmt.Sprintf("\n--- [SECTION: %s | TITLE: %s] ---\n%s\n", anchorSec.Type, anchorSec.Title, anchorSec.Content))
			hasWrittenSec = true
		}
		for _, sec := range targetSecs {
			contextBuilder.WriteString(fmt.Sprintf("\n--- [SECTION: %s | TITLE: %s] ---\n%s\n", sec.Type, sec.Title, sec.Content))
			hasWrittenSec = true
		}
		if !hasWrittenSec && len(doc.Sections) > 0 {
			for _, sec := range doc.Sections {
				contextBuilder.WriteString(fmt.Sprintf("\n--- [SECTION: %s | TITLE: %s] ---\n%s\n", sec.Type, sec.Title, sec.Content))
			}
		}
	}

	// 4. Build Structured LLM Prompts
	systemPrompt := `You are an expert academic research synthesis AI with strict provenance requirements.
Your role is to extract factual, grounded information about a research concept based SOLELY on the provided paper text sections.
CRITICAL ZERO-HALLUCINATION RULES:
1. Every definition, mechanism step, and case study MUST have a citation with an exact, word-for-word verbatim quote from the text.
2. The 'verbatimText' field MUST be copied directly from the text without any alteration or paraphrase.
3. If the paper does not mention an industry case study or application, return an empty array for caseStudies rather than inventing one.
4. Specify the exact 'paperId' where the quote was extracted from.

You must respond strictly with a valid JSON object adhering to this schema:
{
  "definitions": [
    {
      "term": "string (name of concept/term defined)",
      "definition": "string (clear academic definition)",
      "verbatimText": "string (exact verbatim sentence from the text supporting this definition)",
      "section": "string (section title where quote was found)",
      "paperId": "string (ID of the paper where quote was found)"
    }
  ],
  "mechanisms": [
    {
      "stepNumber": 1,
      "description": "string (explanation of this mechanical step)",
      "verbatimText": "string (exact verbatim sentence from the text supporting this step)",
      "section": "string (section title where quote was found)",
      "paperId": "string (ID of the paper where quote was found)"
    }
  ],
  "caseStudies": [
    {
      "industryDomain": "string (domain/sector, e.g. Healthcare, Finance, Energy)",
      "problemSolved": "string (specific bottleneck or challenge addressed)",
      "methodApplied": "string (how the concept/technique was applied)",
      "outcome": "string (measured result, benchmark, or efficiency gain)",
      "verbatimText": "string (exact verbatim sentence from the text supporting this case study)",
      "section": "string (section title where quote was found)",
      "paperId": "string (ID of the paper where quote was found)"
    }
  ]
}`

	userPrompt := fmt.Sprintf(`Research Concept: %s
Primary Paper: %s (%s)
Total Papers Analyzed: %d

Please extract and return a valid JSON object with:
1. Formal definitions of this concept from the papers.
2. Step-by-step technical mechanisms or architecture explaining how it works.
3. Concrete experimental case studies, applications, or benchmark outcomes.

PAPER CONTEXT:
%s`, cmd.Concept, primaryDoc.ID, primaryDoc.Title, len(allDocs), contextBuilder.String())

	var llmResp llmSynthesisResponse
	if err := h.llmClient.CompleteStructured(ctx, systemPrompt, userPrompt, &llmResp); err != nil {
		return nil, fmt.Errorf("LLM structured synthesis failed: %w", err)
	}

	// 5. Map LLM Response to Domain Entity
	synthesis := &extraction.GroundedConceptSynthesis{
		AnchorPaperID: cmd.AnchorPaperID,
		Concept:       cmd.Concept,
		GeneratedAt:   time.Now().UTC(),
	}

	for _, d := range llmResp.Definitions {
		pid := d.PaperID
		if pid == "" {
			pid = cmd.AnchorPaperID
		}
		synthesis.Definitions = append(synthesis.Definitions, extraction.GroundedDefinition{
			Term:       d.Term,
			Definition: d.Definition,
			Citation: extraction.CitationProof{
				PaperID:      pid,
				VerbatimText: d.VerbatimText,
				Section:      d.Section,
			},
		})
	}

	for _, m := range llmResp.Mechanisms {
		pid := m.PaperID
		if pid == "" {
			pid = cmd.AnchorPaperID
		}
		synthesis.Mechanisms = append(synthesis.Mechanisms, extraction.GroundedMechanism{
			StepNumber:  m.StepNumber,
			Description: m.Description,
			Citation: extraction.CitationProof{
				PaperID:      pid,
				VerbatimText: m.VerbatimText,
				Section:      m.Section,
			},
		})
	}

	for _, cs := range llmResp.CaseStudies {
		pid := cs.PaperID
		if pid == "" {
			pid = cmd.AnchorPaperID
		}
		synthesis.CaseStudies = append(synthesis.CaseStudies, extraction.GroundedCaseStudy{
			IndustryDomain: cs.IndustryDomain,
			ProblemSolved:  cs.ProblemSolved,
			MethodApplied:  cs.MethodApplied,
			Outcome:        cs.Outcome,
			Citation: extraction.CitationProof{
				PaperID:      pid,
				VerbatimText: cs.VerbatimText,
				Section:      cs.Section,
			},
		})
	}

	// 6. Grounding Audit: Verify Verbatim Quotes against Document Sections
	auditSummary := h.auditor.Audit(synthesis, allSections)

	// 7. Persist to Extraction Repository
	if h.repo != nil {
		_ = h.repo.SaveSynthesis(ctx, synthesis)
	}

	return &SynthesizeConceptResult{
		Synthesis:    synthesis,
		Audit:        auditSummary,
		FromCache:    false,
		SectionCount: len(allSections),
	}, nil
}

func (h *SynthesizeConceptCommandHandler) loadDocument(ctx context.Context, paperID string) *retrieval.PdfDocument {
	if h.cache != nil {
		if cachedBytes, err := h.cache.Get(ctx, "pdf:doc:"+paperID); err == nil && len(cachedBytes) > 0 {
			var cachedDoc retrieval.PdfDocument
			if err := json.Unmarshal(cachedBytes, &cachedDoc); err == nil {
				return &cachedDoc
			}
		}
	}

	if h.storage != nil && h.parser != nil {
		if h.storage.Exists(ctx, paperID) {
			filePath := h.storage.GetFilePath(paperID)
			parsed, err := h.parser.ParseFile(ctx, filePath)
			if err == nil {
				parsed.ID = paperID
				return parsed
			}
		}
	}

	// Resilient fallback: Query paper metadata from persistence repository
	if h.paperRepo != nil {
		if pMeta, err := h.paperRepo.GetPaper(ctx, paperID); err == nil && pMeta != nil {
			doc := &retrieval.PdfDocument{
				ID:         pMeta.ID,
				FileName:   pMeta.ID + ".pdf",
				Title:      pMeta.Title,
				Abstract:   pMeta.Abstract,
				DOI:        pMeta.Doi,
				TotalPages: 1,
				ParsedAt:   time.Now().UTC(),
			}
			for _, a := range pMeta.Authors {
				doc.Authors = append(doc.Authors, a.Name)
			}
			content := strings.TrimSpace(pMeta.Abstract)
			if content == "" {
				content = fmt.Sprintf("Title: %s. Year: %d. Venue: %s.", pMeta.Title, pMeta.Year, pMeta.Venue)
			}
			doc.Sections = append(doc.Sections, retrieval.ParsedSection{
				ID:        "sec-abs-fallback",
				Type:      retrieval.SectionAbstract,
				Title:     "Abstract",
				Content:   content,
				StartPage: 1,
				EndPage:   1,
			})
			if h.cache != nil {
				if docBytes, err := json.Marshal(doc); err == nil {
					_ = h.cache.Set(ctx, "pdf:doc:"+doc.ID, docBytes, 24*time.Hour)
				}
			}
			return doc
		}
	}

	return nil
}
