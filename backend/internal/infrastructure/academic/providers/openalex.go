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

type OpenAlexProvider struct {
	*BaseProvider
}

// NewOpenAlexProvider creates an OpenAlex API provider.
// OpenAlex rate limit is up to 10 requests/second under the Polite Pool.
func NewOpenAlexProvider(baseURL, politeEmail string, httpClient *http.Client) *OpenAlexProvider {
	if baseURL == "" {
		baseURL = "https://api.openalex.org"
	}
	base := NewBaseProvider(
		ProviderOpenAlex,
		baseURL,
		politeEmail,
		100*time.Millisecond,
		httpClient,
	)
	return &OpenAlexProvider{
		BaseProvider: base,
	}
}

type openAlexWorkResponse struct {
	Results []openAlexWork `json:"results"`
}

type openAlexWork struct {
	ID                    string            `json:"id"`
	DOI                   string            `json:"doi"`
	Title                 string            `json:"title"`
	PublicationYear       int               `json:"publication_year"`
	CitedByCount          int               `json:"cited_by_count"`
	AbstractInvertedIndex map[string][]int  `json:"abstract_inverted_index"`
	Authorships           []openAlexAuthor  `json:"authorships"`
	PrimaryLocation       *openAlexLocation `json:"primary_location"`
	BestOaLocation        *openAlexLocation `json:"best_oa_location"`
	Concepts              []openAlexConcept `json:"concepts"`
	ReferencedWorks       []string          `json:"referenced_works"`
}

type openAlexAuthor struct {
	Author struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"author"`
}

type openAlexLocation struct {
	Source *struct {
		DisplayName string `json:"display_name"`
	} `json:"source"`
	PdfURL         *string `json:"pdf_url"`
	LandingPageURL *string `json:"landing_page_url"`
}

type openAlexConcept struct {
	DisplayName string  `json:"display_name"`
	Score       float64 `json:"score"`
}

func (p *OpenAlexProvider) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	if strings.TrimSpace(query) == "" {
		return nil, retrieval.ErrEmptyQuery
	}
	if limit <= 0 {
		limit = 10
	}

	searchURL := fmt.Sprintf("%s/works?search=%s&per-page=%d&mailto=%s",
		p.baseURL,
		url.QueryEscape(query),
		limit,
		url.QueryEscape(p.politeEmail),
	)

	req, err := p.NewRequest(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload openAlexWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode JSON response: %w", p.name, err)
	}

	papers := make([]retrieval.PaperMetadata, 0, len(payload.Results))
	for _, w := range payload.Results {
		papers = append(papers, p.toDomainPaper(w))
	}

	return papers, nil
}

func (p *OpenAlexProvider) GetPaperDetails(ctx context.Context, id string) (*retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%s: paper id cannot be empty", p.name)
	}

	// OpenAlex supports IDs formatted as W..., https://openalex.org/W..., or DOIs
	endpointID := cleanID
	if strings.HasPrefix(cleanID, "https://openalex.org/") {
		endpointID = strings.TrimPrefix(cleanID, "https://openalex.org/")
	}

	detailURL := fmt.Sprintf("%s/works/%s?mailto=%s",
		p.baseURL,
		url.PathEscape(endpointID),
		url.QueryEscape(p.politeEmail),
	)

	req, err := p.NewRequest(ctx, http.MethodGet, detailURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var w openAlexWork
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("%s: failed to decode paper details: %w", p.name, err)
	}

	paper := p.toDomainPaper(w)
	return &paper, nil
}

func (p *OpenAlexProvider) GetReferences(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	// For references in OpenAlex, we first fetch paper details to get referenced_works
	paper, err := p.GetPaperDetails(ctx, id)
	if err != nil {
		return nil, err
	}

	if paper.Doi != "" {
		// Or search works cited by this paper
		filterURL := fmt.Sprintf("%s/works?filter=cites:%s&per-page=10&mailto=%s",
			p.baseURL,
			url.QueryEscape(paper.ID),
			url.QueryEscape(p.politeEmail),
		)
		req, err := p.NewRequest(ctx, http.MethodGet, filterURL, nil)
		if err == nil {
			resp, err := p.Execute(ctx, req)
			if err == nil {
				defer resp.Body.Close()
				var payload openAlexWorkResponse
				if err := json.NewDecoder(resp.Body).Decode(&payload); err == nil && len(payload.Results) > 0 {
					refs := make([]retrieval.PaperMetadata, 0, len(payload.Results))
					for _, w := range payload.Results {
						refs = append(refs, p.toDomainPaper(w))
					}
					return refs, nil
				}
			}
		}
	}

	return []retrieval.PaperMetadata{}, nil
}

func (p *OpenAlexProvider) GetCitations(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%s: id cannot be empty", p.name)
	}

	filterURL := fmt.Sprintf("%s/works?filter=cites:%s&per-page=10&mailto=%s",
		p.baseURL,
		url.QueryEscape(cleanID),
		url.QueryEscape(p.politeEmail),
	)

	req, err := p.NewRequest(ctx, http.MethodGet, filterURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var payload openAlexWorkResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode citations: %w", p.name, err)
	}

	citations := make([]retrieval.PaperMetadata, 0, len(payload.Results))
	for _, w := range payload.Results {
		citations = append(citations, p.toDomainPaper(w))
	}

	return citations, nil
}

// reconstructAbstract transforms OpenAlex's abstract_inverted_index into a coherent plain text string.
func (p *OpenAlexProvider) reconstructAbstract(invertedIndex map[string][]int) string {
	if len(invertedIndex) == 0 {
		return ""
	}

	maxPos := -1
	for _, positions := range invertedIndex {
		for _, pos := range positions {
			if pos > maxPos {
				maxPos = pos
			}
		}
	}

	if maxPos < 0 {
		return ""
	}

	words := make([]string, maxPos+1)
	for word, positions := range invertedIndex {
		for _, pos := range positions {
			if pos >= 0 && pos <= maxPos {
				words[pos] = word
			}
		}
	}

	return p.SanitizeText(strings.Join(words, " "))
}

func (p *OpenAlexProvider) toDomainPaper(w openAlexWork) retrieval.PaperMetadata {
	authors := make([]retrieval.Author, 0, len(w.Authorships))
	for _, a := range w.Authorships {
		name := strings.TrimSpace(a.Author.DisplayName)
		if name != "" {
			authors = append(authors, retrieval.Author{
				ID:   a.Author.ID,
				Name: name,
			})
		}
	}

	fields := make([]string, 0, len(w.Concepts))
	for i, c := range w.Concepts {
		if i >= 5 {
			break
		}
		if c.DisplayName != "" {
			fields = append(fields, c.DisplayName)
		}
	}

	var venue string
	if w.PrimaryLocation != nil && w.PrimaryLocation.Source != nil {
		venue = w.PrimaryLocation.Source.DisplayName
	}

	var openAccessURL string
	if w.BestOaLocation != nil && w.BestOaLocation.PdfURL != nil {
		openAccessURL = *w.BestOaLocation.PdfURL
	} else if w.PrimaryLocation != nil && w.PrimaryLocation.PdfURL != nil {
		openAccessURL = *w.PrimaryLocation.PdfURL
	}

	return retrieval.PaperMetadata{
		ID:               w.ID,
		Title:            p.SanitizeText(w.Title),
		Abstract:         p.reconstructAbstract(w.AbstractInvertedIndex),
		Authors:          authors,
		Year:             w.PublicationYear,
		CitationCount:    w.CitedByCount,
		FieldsOfStudy:    fields,
		OpenAccessPdfURL: openAccessURL,
		Doi:              p.CleanDOI(w.DOI),
		Venue:            venue,
		FetchedAt:        time.Now().UTC(),
		Sources:          []string{string(p.name)},
	}
}
