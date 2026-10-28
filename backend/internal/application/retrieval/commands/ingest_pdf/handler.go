package ingest_pdf

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/domain/retrieval"
)

type IngestPdfCommandHandler struct {
	parser     interfaces.IPdfParser
	storage    interfaces.IPdfStorage
	repo       interfaces.IPaperRepository
	cache      interfaces.ICacheRepository
	downloader interfaces.IPdfDownloader
}

func NewIngestPdfCommandHandler(
	parser interfaces.IPdfParser,
	storage interfaces.IPdfStorage,
	repo interfaces.IPaperRepository,
	cache interfaces.ICacheRepository,
	downloader ...interfaces.IPdfDownloader,
) *IngestPdfCommandHandler {
	h := &IngestPdfCommandHandler{
		parser:  parser,
		storage: storage,
		repo:    repo,
		cache:   cache,
	}
	if len(downloader) > 0 {
		h.downloader = downloader[0]
	}
	return h
}

func (h *IngestPdfCommandHandler) Handle(ctx context.Context, cmd IngestPdfCommand) (*IngestPdfResult, error) {
	var doc *retrieval.PdfDocument
	var savedPath string
	var err error

	if cmd.DownloadURL != "" && len(cmd.FileBytes) == 0 && cmd.FilePath == "" {
		if h.downloader == nil {
			log.Printf("[IngestPdf] Notice: Downloader service not configured for %q, attempting metadata fallback", cmd.DownloadURL)
		} else {
			paperID := cmd.PaperID
			if paperID == "" {
				paperID = "downloaded_" + time.Now().Format("20060102150405")
			}
			savedPath, err = h.downloader.DownloadPdf(ctx, paperID, cmd.DownloadURL)
			if err != nil {
				log.Printf("[IngestPdf] Warning: downloading PDF for paper %q from %q failed: %v", paperID, cmd.DownloadURL, err)
			} else {
				cmd.FilePath = savedPath
			}
		}
	}

	if len(cmd.FileBytes) > 0 {
		reader := bytes.NewReader(cmd.FileBytes)
		doc, err = h.parser.Parse(ctx, reader, cmd.FileName)
		if err != nil {
			return nil, fmt.Errorf("parsing PDF bytes failed: %w", err)
		}

		if cmd.PaperID != "" {
			doc.ID = cmd.PaperID
		} else if doc.ID == "" && cmd.FileName != "" {
			if parsedID, err := retrieval.ParsePaperID(cmd.FileName); err == nil && parsedID != "" {
				doc.ID = parsedID
			}
		}

		if h.storage != nil {
			_, _ = reader.Seek(0, 0)
			savedPath, _ = h.storage.Save(ctx, doc.ID, reader)
		}
	} else if cmd.FilePath != "" {
		doc, err = h.parser.ParseFile(ctx, cmd.FilePath)
		if err != nil {
			return nil, fmt.Errorf("parsing PDF file failed: %w", err)
		}
		if cmd.PaperID != "" {
			doc.ID = cmd.PaperID
		} else if doc.ID == "" && cmd.FileName != "" {
			if parsedID, err := retrieval.ParsePaperID(cmd.FileName); err == nil && parsedID != "" {
				doc.ID = parsedID
			}
		}
		savedPath = cmd.FilePath
	}

	// Resilient fallback: If full PDF ingestion failed or no PDF URL was provided,
	// synthesize a structured document from the paper's metadata and abstract.
	if doc == nil {
		var meta *retrieval.PaperMetadata = cmd.Metadata
		if meta == nil && h.repo != nil && cmd.PaperID != "" {
			meta, _ = h.repo.GetPaper(ctx, cmd.PaperID)
		}

		if meta != nil && (strings.TrimSpace(meta.Abstract) != "" || strings.TrimSpace(meta.Title) != "") {
			doc = h.createFallbackDocument(meta)
			log.Printf("[IngestPdf] Notice: Successfully created abstract-grounded document for paper %q", meta.ID)
		}
	}

	if doc == nil {
		return nil, retrieval.ErrParsingFailed
	}

	// Cache parsed document for fast section routing
	if h.cache != nil {
		if docBytes, err := json.Marshal(doc); err == nil {
			_ = h.cache.Set(ctx, "pdf:doc:"+doc.ID, docBytes, 24*time.Hour)
		}
	}

	var authors []retrieval.Author
	for _, name := range doc.Authors {
		authors = append(authors, retrieval.Author{Name: name})
	}

	if h.repo != nil {
		paperMeta := &retrieval.PaperMetadata{
			ID:        doc.ID,
			Title:     doc.Title,
			Abstract:  doc.Abstract,
			Doi:       doc.DOI,
			Authors:   authors,
			FetchedAt: time.Now().UTC(),
			Sources:   []string{"local_pdf_upload"},
		}
		_ = h.repo.SavePaper(ctx, paperMeta)
	}

	return &IngestPdfResult{
		Success:      true,
		Message:      fmt.Sprintf("Successfully parsed %d pages into %d sections", doc.TotalPages, len(doc.Sections)),
		PaperID:      doc.ID,
		Document:     doc,
		SavedPath:    savedPath,
		TotalPages:   doc.TotalPages,
		SectionCount: len(doc.Sections),
		RefCount:     len(doc.References),
	}, nil
}

func (h *IngestPdfCommandHandler) createFallbackDocument(meta *retrieval.PaperMetadata) *retrieval.PdfDocument {
	doc := &retrieval.PdfDocument{
		ID:         meta.ID,
		FileName:   meta.ID + ".pdf",
		Title:      meta.Title,
		Abstract:   meta.Abstract,
		DOI:        meta.Doi,
		TotalPages: 1,
		ParsedAt:   time.Now().UTC(),
	}

	for _, a := range meta.Authors {
		doc.Authors = append(doc.Authors, a.Name)
	}

	content := strings.TrimSpace(meta.Abstract)
	if content == "" {
		content = fmt.Sprintf("Title: %s. Year: %d. Venue: %s.", meta.Title, meta.Year, meta.Venue)
	}

	doc.Sections = append(doc.Sections, retrieval.ParsedSection{
		ID:        "sec-abs-fallback",
		Type:      retrieval.SectionAbstract,
		Title:     "Abstract",
		Content:   content,
		StartPage: 1,
		EndPage:   1,
	})

	return doc
}
