package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kuorum/backend/internal/api/response"
	"kuorum/backend/internal/application/common/interfaces"
	"kuorum/backend/internal/application/common/mediator"
	"kuorum/backend/internal/application/retrieval/commands/ingest_pdf"
	"kuorum/backend/internal/application/retrieval/queries/get_citations"
	"kuorum/backend/internal/application/retrieval/queries/get_paper_details"
	"kuorum/backend/internal/application/retrieval/queries/get_pdf_sections"
	"kuorum/backend/internal/application/retrieval/queries/get_providers"
	"kuorum/backend/internal/application/retrieval/queries/get_references"
	"kuorum/backend/internal/application/retrieval/queries/search_papers"
	"kuorum/backend/internal/domain/retrieval"

	"github.com/go-chi/chi/v5"
)

type PaperController struct {
	mediator mediator.Mediator
	storage  interfaces.IPdfStorage
}

func NewPaperController(mediator mediator.Mediator, storage ...interfaces.IPdfStorage) *PaperController {
	ctrl := &PaperController{
		mediator: mediator,
	}
	if len(storage) > 0 {
		ctrl.storage = storage[0]
	}
	return ctrl
}

// Search godoc
// @Summary Search papers across academic providers
// @Description Concurrently queries OpenAlex, Semantic Scholar, arXiv, and Crossref with automatic deduplication.
// @Tags Papers
// @Accept json
// @Produce json
// @Param query query string false "Search query keywords or concept"
// @Param q query string false "Alias for query"
// @Param limit query int false "Max papers to return (default: 10)"
// @Param providers query string false "Comma-separated provider names (openalex, semanticscholar, arxiv, crossref)"
// @Success 200 {object} response.Envelope{data=search_papers.SearchPapersResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/papers/search [get]
// @Router /api/v1/papers/search [post]
func (c *PaperController) Search(w http.ResponseWriter, r *http.Request) {
	var q search_papers.SearchPapersQuery

	if r.Method == http.MethodPost && r.Header.Get("Content-Type") == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&q); err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
			return
		}
	} else {
		queryText := r.URL.Query().Get("query")
		if queryText == "" {
			queryText = r.URL.Query().Get("q")
		}

		limit := 10
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		var providers []string
		if provs := r.URL.Query().Get("providers"); provs != "" {
			for _, p := range strings.Split(provs, ",") {
				if clean := strings.TrimSpace(p); clean != "" {
					providers = append(providers, clean)
				}
			}
		}

		q = search_papers.SearchPapersQuery{
			Query:     queryText,
			Limit:     limit,
			Providers: providers,
		}
	}

	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Papers retrieved successfully", result)
}

// GetDetails godoc
// @Summary Get paper metadata details
// @Description Fetches normalized paper metadata from academic providers by ID, DOI, or canonical URL.
// @Tags Papers
// @Produce json
// @Param id query string false "Paper ID, DOI, or URL"
// @Param id path string false "Paper ID (when using path parameter)"
// @Success 200 {object} response.Envelope{data=retrieval.PaperMetadata}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/papers/details [get]
// @Router /api/v1/papers/{id} [get]
func (c *PaperController) GetDetails(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing paper ID", "id parameter is required")
		return
	}

	q := get_paper_details.GetPaperDetailsQuery{ID: id}
	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Paper details retrieved successfully", result)
}

// GetReferences godoc
// @Summary Get foundational reference papers
// @Description Retrieves references cited by the specified paper ID.
// @Tags Papers
// @Produce json
// @Param id query string false "Paper ID, DOI, or URL"
// @Param id path string false "Paper ID"
// @Success 200 {object} response.Envelope{data=get_references.GetReferencesResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/papers/references [get]
// @Router /api/v1/papers/{id}/references [get]
func (c *PaperController) GetReferences(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing paper ID", "id parameter is required")
		return
	}

	q := get_references.GetReferencesQuery{ID: id}
	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Paper references retrieved successfully", result)
}

// GetCitations godoc
// @Summary Get forward citations
// @Description Retrieves papers that cite the specified paper ID.
// @Tags Papers
// @Produce json
// @Param id query string false "Paper ID, DOI, or URL"
// @Param id path string false "Paper ID"
// @Success 200 {object} response.Envelope{data=get_citations.GetCitationsResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/papers/citations [get]
// @Router /api/v1/papers/{id}/citations [get]
func (c *PaperController) GetCitations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}

	id = strings.TrimSpace(id)
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing paper ID", "id parameter is required")
		return
	}

	q := get_citations.GetCitationsQuery{ID: id}
	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Paper citations retrieved successfully", result)
}

// GetProviders godoc
// @Summary Get active academic providers
// @Description Returns the list of registered academic literature providers in the aggregator.
// @Tags Papers
// @Produce json
// @Success 200 {object} response.Envelope{data=get_providers.GetProvidersResult}
// @Router /api/v1/papers/providers [get]
func (c *PaperController) GetProviders(w http.ResponseWriter, r *http.Request) {
	q := get_providers.GetProvidersQuery{}
	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Available providers retrieved successfully", result)
}

// UploadPdf godoc
// @Summary Upload and parse academic PDF
// @Description Ingests an academic PDF, extracts title, authors, canonical sections, and references.
// @Tags Papers
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Academic PDF document"
// @Success 200 {object} response.Envelope{data=ingest_pdf.IngestPdfResult}
// @Failure 400 {object} response.Envelope
// @Router /api/v1/papers/upload-pdf [post]
func (c *PaperController) UploadPdf(w http.ResponseWriter, r *http.Request) {
	// 50MB maximum upload limit
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		response.Error(w, http.StatusBadRequest, "File upload too large", "Maximum allowed PDF size is 50MB")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Missing file in upload", "Form field 'file' is required")
		return
	}
	defer file.Close()

	paperID, err := retrieval.ParsePaperID(header.Filename)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid file name", err.Error())
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to read uploaded file", err.Error())
		return
	}

	cmd := ingest_pdf.IngestPdfCommand{
		PaperID:   paperID,
		FileName:  header.Filename,
		FileBytes: fileBytes,
	}

	result, err := c.mediator.Send(r.Context(), cmd)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "PDF uploaded and parsed successfully", result)
}

// GetSections godoc
// @Summary Get academic paper sections
// @Description Returns structured sections of an ingested paper with optional intent routing (e.g. ?intent=results).
// @Tags Papers
// @Produce json
// @Param id path string true "Paper ID"
// @Param intent query string false "Research query intent (results, definitions, mechanisms, prior_work)"
// @Success 200 {object} response.Envelope{data=get_pdf_sections.GetPdfSectionsResult}
// @Failure 400 {object} response.Envelope
// @Failure 404 {object} response.Envelope
// @Router /api/v1/papers/{id}/sections [get]
func (c *PaperController) GetSections(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing paper ID", "id parameter is required")
		return
	}

	intent := r.URL.Query().Get("intent")

	q := get_pdf_sections.GetPdfSectionsQuery{
		PaperID: id,
		Intent:  intent,
	}

	result, err := c.mediator.Query(r.Context(), q)
	if err != nil {
		response.FromAppError(w, err)
		return
	}

	response.Success(w, http.StatusOK, "Paper sections retrieved successfully", result)
}

// StreamPdf godoc
// @Summary Stream raw PDF document
// @Description Serves raw PDF file for inline browser rendering in the frontend reader.
// @Tags Papers
// @Produce application/pdf
// @Param id path string true "Paper ID"
// @Success 200 {file} binary
// @Failure 404 {object} response.Envelope
// @Router /api/v1/papers/{id}/pdf [get]
func (c *PaperController) StreamPdf(w http.ResponseWriter, r *http.Request) {
	if c.storage == nil {
		response.Error(w, http.StatusNotImplemented, "PDF storage not enabled", "Local PDF storage is not configured")
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("id")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		response.Error(w, http.StatusBadRequest, "Missing paper ID", "id parameter is required")
		return
	}

	if !c.storage.Exists(r.Context(), id) {
		response.Error(w, http.StatusNotFound, "PDF file not found", fmt.Sprintf("No PDF stored for paper ID %q", id))
		return
	}

	file, err := c.storage.Open(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Failed to open PDF file", err.Error())
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.pdf\"", id))

	if seeker, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(w, r, fmt.Sprintf("%s.pdf", id), time.Now(), seeker)
		return
	}

	_, _ = io.Copy(w, file)
}
