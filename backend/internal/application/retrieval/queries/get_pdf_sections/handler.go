package get_pdf_sections

import (
	"context"
	"encoding/json"
	"fmt"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

type GetPdfSectionsQueryHandler struct {
	cache   interfaces.ICacheRepository
	parser  interfaces.IPdfParser
	storage interfaces.IPdfStorage
}

func NewGetPdfSectionsQueryHandler(
	cache interfaces.ICacheRepository,
	parser interfaces.IPdfParser,
	storage interfaces.IPdfStorage,
) *GetPdfSectionsQueryHandler {
	return &GetPdfSectionsQueryHandler{
		cache:   cache,
		parser:  parser,
		storage: storage,
	}
}

func (h *GetPdfSectionsQueryHandler) Handle(ctx context.Context, q GetPdfSectionsQuery) (*GetPdfSectionsResult, error) {
	var doc *retrieval.PdfDocument

	// 1. Try fetching parsed document from cache
	if h.cache != nil {
		if cachedBytes, err := h.cache.Get(ctx, "pdf:doc:"+q.PaperID); err == nil && len(cachedBytes) > 0 {
			var cachedDoc retrieval.PdfDocument
			if err := json.Unmarshal(cachedBytes, &cachedDoc); err == nil {
				doc = &cachedDoc
			}
		}
	}

	// 2. If not cached, re-parse from storage
	if doc == nil && h.storage != nil && h.parser != nil {
		if h.storage.Exists(ctx, q.PaperID) {
			filePath := h.storage.GetFilePath(q.PaperID)
			parsed, err := h.parser.ParseFile(ctx, filePath)
			if err == nil {
				doc = parsed
				doc.ID = q.PaperID
			}
		}
	}

	if doc == nil {
		return nil, fmt.Errorf("%w: paper %q not found or not yet parsed", retrieval.ErrPaperNotFound, q.PaperID)
	}

	// 3. Perform intent-based section filtering
	anchor, targets := doc.GetTargetSectionsForIntent(q.Intent)

	return &GetPdfSectionsResult{
		PaperID:        q.PaperID,
		Intent:         q.Intent,
		AnchorSection:  anchor,
		TargetSections: targets,
		AllSections:    doc.Sections,
		TotalSections:  len(doc.Sections),
	}, nil
}
