package providers

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"kuorum/backend/internal/domain/retrieval"
)

type ArxivProvider struct {
	*BaseProvider
}

// NewArxivProvider creates an arXiv API provider.
// arXiv public guidelines strictly require no more than 1 request per 3 seconds.
func NewArxivProvider(baseURL, politeEmail string, httpClient *http.Client) *ArxivProvider {
	if baseURL == "" {
		baseURL = "https://export.arxiv.org/api"
	}

	base := NewBaseProvider(
		ProviderArxiv,
		baseURL,
		politeEmail,
		3*time.Second,
		httpClient,
	)

	return &ArxivProvider{
		BaseProvider: base,
	}
}

type arxivFeed struct {
	XMLName xml.Name     `xml:"feed"`
	Entries []arxivEntry `xml:"entry"`
}

type arxivEntry struct {
	ID              string        `xml:"id"`
	Published       string        `xml:"published"`
	Title           string        `xml:"title"`
	Summary         string        `xml:"summary"`
	Authors         []arxivAuthor `xml:"author"`
	Links           []arxivLink   `xml:"link"`
	DOI             string        `xml:"doi"`
	PrimaryCategory struct {
		Term string `xml:"term,attr"`
	} `xml:"primary_category"`
	Categories []struct {
		Term string `xml:"term,attr"`
	} `xml:"category"`
	JournalRef string `xml:"journal_ref"`
}

type arxivAuthor struct {
	Name string `xml:"name"`
}

type arxivLink struct {
	Href  string `xml:"href,attr"`
	Rel   string `xml:"rel,attr"`
	Type  string `xml:"type,attr"`
	Title string `xml:"title,attr"`
}

func (p *ArxivProvider) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	if strings.TrimSpace(query) == "" {
		return nil, retrieval.ErrEmptyQuery
	}
	if limit <= 0 {
		limit = 10
	}

	// Format query with targeted boolean conjunction and phrase matching across arXiv
	formattedQuery := formatArxivQuery(query)
	searchURL := fmt.Sprintf("%s/query?search_query=%s&start=0&max_results=%d&sortBy=relevance&sortOrder=descending",
		p.baseURL,
		url.QueryEscape(formattedQuery),
		limit,
	)

	req, err := p.NewRequest(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/atom+xml, application/xml")

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var feed arxivFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("%s: failed to decode XML feed: %w", p.name, err)
	}

	papers := make([]retrieval.PaperMetadata, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		papers = append(papers, p.toDomainPaper(entry))
	}

	return papers, nil
}

func (p *ArxivProvider) GetPaperDetails(ctx context.Context, id string) (*retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(id)
	if cleanID == "" {
		return nil, fmt.Errorf("%s: paper id cannot be empty", p.name)
	}

	arxivID := cleanID
	if strings.Contains(arxivID, "arxiv.org/abs/") {
		parts := strings.Split(arxivID, "arxiv.org/abs/")
		if len(parts) > 1 {
			arxivID = parts[1]
		}
	}
	arxivID = strings.TrimPrefix(arxivID, "arxiv:")

	queryURL := fmt.Sprintf("%s/query?id_list=%s",
		p.baseURL,
		url.QueryEscape(arxivID),
	)

	req, err := p.NewRequest(ctx, http.MethodGet, queryURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/atom+xml, application/xml")

	resp, err := p.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var feed arxivFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("%s: failed to decode XML feed: %w", p.name, err)
	}

	if len(feed.Entries) == 0 {
		return nil, fmt.Errorf("%s: %w: %s", p.name, retrieval.ErrPaperNotFound, cleanID)
	}

	paper := p.toDomainPaper(feed.Entries[0])
	return &paper, nil
}

func (p *ArxivProvider) GetReferences(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	// arXiv Atom API does not directly expose parsed citation references
	return []retrieval.PaperMetadata{}, nil
}

func (p *ArxivProvider) GetCitations(ctx context.Context, id string) ([]retrieval.PaperMetadata, error) {
	// arXiv Atom API does not track forward citations
	return []retrieval.PaperMetadata{}, nil
}

func (p *ArxivProvider) toDomainPaper(entry arxivEntry) retrieval.PaperMetadata {
	authors := make([]retrieval.Author, 0, len(entry.Authors))
	for _, a := range entry.Authors {
		name := strings.TrimSpace(a.Name)
		if name != "" {
			authors = append(authors, retrieval.Author{
				ID:   fmt.Sprintf("arxiv-auth-%s", strings.ToLower(strings.ReplaceAll(name, " ", "-"))),
				Name: name,
			})
		}
	}

	var year int
	if entry.Published != "" {
		if t, err := time.Parse(time.RFC3339, entry.Published); err == nil {
			year = t.Year()
		} else if len(entry.Published) >= 4 {
			_, _ = fmt.Sscanf(entry.Published[:4], "%d", &year)
		}
	}

	var pdfURL string
	for _, link := range entry.Links {
		if link.Type == "application/pdf" || link.Title == "pdf" || strings.Contains(link.Href, "/pdf/") {
			pdfURL = link.Href
			break
		}
	}

	// If pdfURL is still empty and we have an abs link, convert it
	if pdfURL == "" && strings.Contains(entry.ID, "/abs/") {
		pdfURL = strings.Replace(entry.ID, "/abs/", "/pdf/", 1) + ".pdf"
	}

	fields := make([]string, 0)
	if entry.PrimaryCategory.Term != "" {
		fields = append(fields, entry.PrimaryCategory.Term)
	}
	for _, cat := range entry.Categories {
		if cat.Term != "" && cat.Term != entry.PrimaryCategory.Term {
			fields = append(fields, cat.Term)
		}
	}

	venue := entry.JournalRef
	if venue == "" {
		venue = "arXiv preprint"
	}

	return retrieval.PaperMetadata{
		ID:               entry.ID,
		Title:            p.SanitizeText(entry.Title),
		Abstract:         p.SanitizeText(entry.Summary),
		Authors:          authors,
		Year:             year,
		FieldsOfStudy:    fields,
		OpenAccessPdfURL: pdfURL,
		Doi:              p.CleanDOI(entry.DOI),
		Venue:            venue,
		FetchedAt:        time.Now().UTC(),
		Sources:          []string{string(p.name)},
	}
}

// formatArxivQuery constructs targeted arXiv boolean search syntax:
// exact phrase search OR conjunction of all significant terms across all fields.
func formatArxivQuery(query string) string {
	clean := strings.TrimSpace(query)
	words := NonAlphaNumRegex.Split(strings.ToLower(clean), -1)
	terms := make([]string, 0)
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) >= 2 && !IsStopWord(w) {
			terms = append(terms, w)
		}
	}

	if len(terms) <= 1 {
		return fmt.Sprintf("all:%s", clean)
	}

	andClauses := make([]string, len(terms))
	for i, t := range terms {
		andClauses[i] = fmt.Sprintf("all:%s", t)
	}

	return fmt.Sprintf("all:\"%s\" OR (%s)", clean, strings.Join(andClauses, " AND "))
}
