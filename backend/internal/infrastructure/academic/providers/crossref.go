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

type CrossrefProvider struct {
	*BaseProvider
}

// NewCrossrefProvider creates a Crossref REST API provider.
// Crossref provides a Polite Pool for requests including a valid contact email.
func NewCrossrefProvider(baseURL, politeEmail string, httpClient *http.Client) *CrossrefProvider {
	if baseURL == "" {
		baseURL = "https://api.crossref.org"
	}

	base := NewBaseProvider(
		ProviderCrossref,
		baseURL,
		politeEmail,
		100*time.Millisecond,
		httpClient,
	)

	return &CrossrefProvider{
		BaseProvider: base,
	}
}

type crossrefSearchResponse struct {
	Status  string          `json:"status"`
	Message crossrefMessage `json:"message"`
}

type crossrefMessage struct {
	TotalResults int            `json:"total-results"`
	Items        []crossrefWork `json:"items"`
}

type crossrefDetailResponse struct {
	Status  string       `json:"status"`
	Message crossrefWork `json:"message"`
}

type crossrefWork struct {
	DOI                 string              `json:"DOI"`
	Title               []string            `json:"title"`
	Author              []crossrefAuthor    `json:"author"`
	Abstract            string              `json:"abstract"`
	IsReferencedByCount int                 `json:"is-referenced-by-count"`
	ContainerTitle      []string            `json:"container-title"`
	Issued              *crossrefDate       `json:"issued"`
	PublishedPrint      *crossrefDate       `json:"published-print"`
	Subject             []string            `json:"subject"`
	Link                []crossrefLink      `json:"link"`
	Reference           []crossrefReference `json:"reference"`
}

type crossrefAuthor struct {
	Given  string `json:"given"`
	Family string `json:"family"`
	Name   string `json:"name"`
}

type crossrefDate struct {
	DateParts [][]int `json:"date-parts"`
}

type crossrefLink struct {
	URL         string `json:"URL"`
	ContentType string `json:"content-type"`
}

type crossrefReference struct {
	DOI          string `json:"DOI"`
	ArticleTitle string `json:"article-title"`
	Year         string `json:"year"`
	JournalTitle string `json:"journal-title"`
	Unstructured string `json:"unstructured"`
}

func (p *CrossrefProvider) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	if strings.TrimSpace(query) == "" {
		return nil, retrieval.ErrEmptyQuery
	}
	if limit <= 0 {
		limit = 10
	}

	// Use query.bibliographic instead of broad query to target title, abstract, and authors (avoiding footnote citations)
	searchURL := fmt.Sprintf("%s/works?query.bibliographic=%s&rows=%d&mailto=%s",
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

	var payload crossrefSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode JSON response: %w", p.name, err)
	}

	papers := make([]retrieval.PaperMetadata, 0, len(payload.Message.Items))
	for _, item := range payload.Message.Items {
		papers = append(papers, p.toDomainPaper(item))
	}

	return papers, nil
}

func (p *CrossrefProvider) GetPaperDetails(ctx context.Context, id string) (*retrieval.PaperMetadata, error) {
	cleanDOI := p.CleanDOI(id)
	if cleanDOI == "" {
		return nil, fmt.Errorf("%s: valid DOI is required for Crossref lookup", p.name)
	}

	detailURL := fmt.Sprintf("%s/works/%s?mailto=%s",
		p.baseURL,
		url.PathEscape(cleanDOI),
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

	var payload crossrefDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode paper details: %w", p.name, err)
	}

	paper := p.toDomainPaper(payload.Message)
	return &paper, nil
}

func (p *CrossrefProvider) GetReferences(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	cleanDOI := p.CleanDOI(id)
	if cleanDOI == "" {
		return nil, fmt.Errorf("%s: valid DOI is required for Crossref lookup", p.name)
	}

	detailURL := fmt.Sprintf("%s/works/%s?mailto=%s",
		p.baseURL,
		url.PathEscape(cleanDOI),
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

	var payload crossrefDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%s: failed to decode paper details: %w", p.name, err)
	}

	refs := make([]retrieval.PaperMetadata, 0, len(payload.Message.Reference))
	for _, ref := range payload.Message.Reference {
		title := ref.ArticleTitle
		if title == "" {
			title = ref.Unstructured
		}
		if title == "" && ref.DOI != "" {
			title = fmt.Sprintf("Referenced Paper (%s)", ref.DOI)
		}

		if title != "" || ref.DOI != "" {
			var year int
			if ref.Year != "" {
				_, _ = fmt.Sscanf(ref.Year, "%d", &year)
			}

			refs = append(refs, retrieval.PaperMetadata{
				ID:        p.CleanDOI(ref.DOI),
				Title:     p.SanitizeText(title),
				Year:      year,
				Doi:       p.CleanDOI(ref.DOI),
				Venue:     ref.JournalTitle,
				FetchedAt: time.Now().UTC(),
				Sources:   []string{string(p.name)},
			})
		}
	}

	return refs, nil
}

func (p *CrossrefProvider) GetCitations(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	// Crossref provides is-referenced-by-count but does not provide an open endpoint listing forward citation papers
	return []retrieval.PaperMetadata{}, nil
}

func (p *CrossrefProvider) toDomainPaper(item crossrefWork) retrieval.PaperMetadata {
	var title string
	if len(item.Title) > 0 {
		title = item.Title[0]
	}

	authors := make([]retrieval.Author, 0, len(item.Author))
	for _, a := range item.Author {
		var name string
		if a.Name != "" {
			name = a.Name
		} else {
			name = strings.TrimSpace(fmt.Sprintf("%s %s", a.Given, a.Family))
		}
		if name != "" {
			authors = append(authors, retrieval.Author{
				ID:   fmt.Sprintf("crossref-auth-%s", strings.ToLower(strings.ReplaceAll(name, " ", "-"))),
				Name: name,
			})
		}
	}

	var year int
	if item.PublishedPrint != nil && len(item.PublishedPrint.DateParts) > 0 && len(item.PublishedPrint.DateParts[0]) > 0 {
		year = item.PublishedPrint.DateParts[0][0]
	} else if item.Issued != nil && len(item.Issued.DateParts) > 0 && len(item.Issued.DateParts[0]) > 0 {
		year = item.Issued.DateParts[0][0]
	}

	var venue string
	if len(item.ContainerTitle) > 0 {
		venue = item.ContainerTitle[0]
	}

	var pdfURL string
	for _, link := range item.Link {
		if strings.Contains(link.ContentType, "pdf") || strings.HasSuffix(link.URL, ".pdf") {
			pdfURL = link.URL
			break
		}
	}

	abstract := p.StripXML(item.Abstract)

	doi := p.CleanDOI(item.DOI)
	id := doi
	if id == "" {
		id = fmt.Sprintf("crossref-%d", time.Now().UnixNano())
	}

	return retrieval.PaperMetadata{
		ID:               id,
		Title:            p.SanitizeText(title),
		Abstract:         abstract,
		Authors:          authors,
		Year:             year,
		CitationCount:    item.IsReferencedByCount,
		FieldsOfStudy:    item.Subject,
		OpenAccessPdfURL: pdfURL,
		Doi:              doi,
		Venue:            venue,
		FetchedAt:        time.Now().UTC(),
		Sources:          []string{string(p.name)},
	}
}
