package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kuorum/backend/internal/domain/retrieval"
)

const (
	s2DefaultFields = "paperId,title,abstract,authors,year,citationCount,influentialCitationCount,tldr,fieldsOfStudy,venue,externalIds,openAccessPdf"
)

type SemanticScholarProvider struct {
	*BaseProvider
	apiKey string
}

// NewSemanticScholarProvider creates a Semantic Scholar API provider.
// If apiKey is provided, allows higher throughput; otherwise strictly throttles to 1 req/second.
func NewSemanticScholarProvider(baseURL, apiKey string, httpClient *http.Client) *SemanticScholarProvider {
	if baseURL == "" {
		baseURL = "https://api.semanticscholar.org/graph/v1"
	}

	interval := 1100 * time.Millisecond
	if apiKey != "" {
		interval = 20 * time.Millisecond
	}

	base := NewBaseProvider(
		ProviderSemanticScholar,
		baseURL,
		"",
		interval,
		httpClient,
	)

	return &SemanticScholarProvider{
		BaseProvider: base,
		apiKey:       apiKey,
	}
}

type s2SearchResponse struct {
	Total int       `json:"total"`
	Data  []s2Paper `json:"data"`
}

type s2Paper struct {
	PaperID                  string            `json:"paperId"`
	Title                    string            `json:"title"`
	Abstract                 string            `json:"abstract"`
	Authors                  []s2Author        `json:"authors"`
	Year                     int               `json:"year"`
	CitationCount            int               `json:"citationCount"`
	InfluentialCitationCount int               `json:"influentialCitationCount"`
	TLDR                     *s2TLDR           `json:"tldr"`
	Venue                    string            `json:"venue"`
	FieldsOfStudy            []string          `json:"fieldsOfStudy"`
	ExternalIds              map[string]string `json:"externalIds"`
	OpenAccessPdf            *s2Pdf            `json:"openAccessPdf"`
}

type s2Author struct {
	AuthorID string `json:"authorId"`
	Name     string `json:"name"`
}

type s2TLDR struct {
	Model string `json:"model"`
	Text  string `json:"text"`
}

type s2Pdf struct {
	URL    string `json:"url"`
	Status string `json:"status"`
}

type s2ReferenceWrapper struct {
	CitedPaper s2Paper `json:"citedPaper"`
}

type s2CitationWrapper struct {
	CitingPaper s2Paper `json:"citingPaper"`
}

type s2ReferencesResponse struct {
	Data []s2ReferenceWrapper `json:"data"`
}

type s2CitationsResponse struct {
	Data []s2CitationWrapper `json:"data"`
}

func (p *SemanticScholarProvider) newS2Request(ctx context.Context, method, fullURL string) (*http.Request, error) {
	req, err := p.NewRequest(ctx, method, fullURL, nil)
	if err != nil {
		return nil, err
	}
	if p.apiKey != "" {
		req.Header.Set("x-api-key", p.apiKey)
	}
	return req, nil
}

func (p *SemanticScholarProvider) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	if strings.TrimSpace(query) == "" {
		return nil, retrieval.ErrEmptyQuery
	}
	if limit <= 0 {
		limit = 10
	}

	searchURL := fmt.Sprintf("%s/paper/search?query=%s&limit=%d&fields=%s",
		p.baseURL,
		url.QueryEscape(query),
		limit,
		s2DefaultFields,
	)

	req, err := p.newS2Request(ctx, http.MethodGet, searchURL)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload s2SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode JSON response: %w", p.name, err)
	}

	papers := make([]retrieval.PaperMetadata, 0, len(payload.Data))
	for _, item := range payload.Data {
		papers = append(papers, p.toDomainPaper(item))
	}

	return papers, nil
}

func (p *SemanticScholarProvider) GetPaperDetails(ctx context.Context, id string) (*retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%s: paper id cannot be empty", p.name)
	}

	detailURL := fmt.Sprintf("%s/paper/%s?fields=%s",
		p.baseURL,
		url.PathEscape(cleanID),
		s2DefaultFields,
	)

	req, err := p.newS2Request(ctx, http.MethodGet, detailURL)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var item s2Paper
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("%s: failed to decode paper details: %w", p.name, err)
	}

	paper := p.toDomainPaper(item)
	return &paper, nil
}

func (p *SemanticScholarProvider) GetReferences(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%s: paper id cannot be empty", p.name)
	}

	refURL := fmt.Sprintf("%s/paper/%s/references?fields=%s&limit=20",
		p.baseURL,
		url.PathEscape(cleanID),
		s2DefaultFields,
	)

	req, err := p.newS2Request(ctx, http.MethodGet, refURL)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload s2ReferencesResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode references: %w", p.name, err)
	}

	refs := make([]retrieval.PaperMetadata, 0, len(payload.Data))
	for _, item := range payload.Data {
		if item.CitedPaper.PaperID != "" {
			refs = append(refs, p.toDomainPaper(item.CitedPaper))
		}
	}

	return refs, nil
}

func (p *SemanticScholarProvider) GetCitations(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%s: paper id cannot be empty", p.name)
	}

	citURL := fmt.Sprintf("%s/paper/%s/citations?fields=%s&limit=20",
		p.baseURL,
		url.PathEscape(cleanID),
		s2DefaultFields,
	)

	req, err := p.newS2Request(ctx, http.MethodGet, citURL)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload s2CitationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode citations: %w", p.name, err)
	}

	cits := make([]retrieval.PaperMetadata, 0, len(payload.Data))
	for _, item := range payload.Data {
		if item.CitingPaper.PaperID != "" {
			cits = append(cits, p.toDomainPaper(item.CitingPaper))
		}
	}

	return cits, nil
}

func (p *SemanticScholarProvider) toDomainPaper(item s2Paper) retrieval.PaperMetadata {
	authors := make([]retrieval.Author, 0, len(item.Authors))
	for _, a := range item.Authors {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			authors = append(authors, retrieval.Author{
				ID:   a.AuthorID,
				Name: name,
			})
		}
	}

	var tldr string
	if item.TLDR != nil {
		tldr = p.SanitizeText(item.TLDR.Text)
	}

	var pdfURL string
	if item.OpenAccessPdf != nil {
		pdfURL = item.OpenAccessPdf.URL
	}

	doi := item.ExternalIds["DOI"]
	if doi == "" {
		doi = item.ExternalIds["doi"]
	}

	return retrieval.PaperMetadata{
		ID:                       item.PaperID,
		Title:                    p.SanitizeText(item.Title),
		Abstract:                 p.SanitizeText(item.Abstract),
		Authors:                  authors,
		Year:                     item.Year,
		CitationCount:            item.CitationCount,
		InfluentialCitationCount: item.InfluentialCitationCount,
		TLDR:                     tldr,
		FieldsOfStudy:            item.FieldsOfStudy,
		OpenAccessPdfURL:         pdfURL,
		Doi:                      p.CleanDOI(doi),
		Venue:                    item.Venue,
		FetchedAt:                time.Now().UTC(),
		Sources:                  []string{string(p.name)},
	}
}
