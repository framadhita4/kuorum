package providers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAlexProvider_SearchAndDetails(t *testing.T) {
	mockJSON := `{
		"results": [
			{
				"id": "https://openalex.org/W123456789",
				"doi": "https://doi.org/10.1000/182",
				"title": "Deep Learning Frameworks",
				"publication_year": 2022,
				"cited_by_count": 350,
				"abstract_inverted_index": {
					"Deep": [0],
					"learning": [1],
					"has": [2],
					"revolutionized": [3],
					"AI": [4]
				},
				"authorships": [
					{
						"author": {
							"id": "https://openalex.org/A1",
							"display_name": "Alice Turing"
						}
					}
				],
				"best_oa_location": {
					"pdf_url": "https://example.com/paper.pdf"
				},
				"primary_location": {
					"source": {
						"display_name": "Nature AI"
					}
				},
				"concepts": [
					{ "display_name": "Machine Learning", "score": 0.95 },
					{ "display_name": "Neural Networks", "score": 0.88 }
				]
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Errorf("expected User-Agent header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	provider := NewOpenAlexProvider(server.URL, "test@example.com", server.Client())

	ctx := context.Background()
	papers, err := provider.SearchPapers(ctx, "deep learning", 5)
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}

	p := papers[0]
	if p.Title != "Deep Learning Frameworks" {
		t.Errorf("expected title 'Deep Learning Frameworks', got '%s'", p.Title)
	}
	expectedAbstract := "Deep learning has revolutionized AI"
	if p.Abstract != expectedAbstract {
		t.Errorf("expected reconstructed abstract '%s', got '%s'", expectedAbstract, p.Abstract)
	}
	if p.OpenAccessPdfURL != "https://example.com/paper.pdf" {
		t.Errorf("unexpected PDF url: %s", p.OpenAccessPdfURL)
	}
	if p.Doi != "10.1000/182" {
		t.Errorf("unexpected cleaned DOI: %s", p.Doi)
	}
	if p.Venue != "Nature AI" {
		t.Errorf("unexpected venue: %s", p.Venue)
	}
	if len(p.Authors) != 1 || p.Authors[0].Name != "Alice Turing" {
		t.Errorf("unexpected authors: %+v", p.Authors)
	}
}

func TestSemanticScholarProvider_SearchPapers(t *testing.T) {
	mockJSON := `{
		"total": 1,
		"data": [
			{
				"paperId": "s2-paper-abc",
				"title": "Attention Is All You Need",
				"abstract": "The dominant sequence transduction models are based on complex recurrent or convolutional neural networks.",
				"year": 2017,
				"citationCount": 95000,
				"influentialCitationCount": 12000,
				"venue": "NeurIPS 2017",
				"tldr": {
					"text": "We propose the Transformer, a model architecture eschewing recurrence and relying entirely on an attention mechanism."
				},
				"fieldsOfStudy": ["Computer Science", "Artificial Intelligence"],
				"authors": [
					{ "authorId": "auth-1", "name": "Ashish Vaswani" }
				],
				"externalIds": {
					"DOI": "10.5555/3295222.3295349"
				},
				"openAccessPdf": {
					"url": "https://arxiv.org/pdf/1706.03762.pdf"
				}
			}
		]
	}`

	var receivedAPIKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAPIKey = r.Header.Get("x-api-key")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	provider := NewSemanticScholarProvider(server.URL, "secret-test-key", server.Client())

	papers, err := provider.SearchPapers(context.Background(), "transformer", 1)
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if receivedAPIKey != "secret-test-key" {
		t.Errorf("expected x-api-key header 'secret-test-key', got '%s'", receivedAPIKey)
	}

	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}

	p := papers[0]
	if p.CitationCount != 95000 || p.InfluentialCitationCount != 12000 {
		t.Errorf("unexpected citation counts: %d, %d", p.CitationCount, p.InfluentialCitationCount)
	}
	if p.TLDR == "" {
		t.Errorf("expected TLDR to be populated")
	}
	if p.Doi != "10.5555/3295222.3295349" {
		t.Errorf("unexpected DOI: %s", p.Doi)
	}
}

func TestArxivProvider_SearchPapers(t *testing.T) {
	mockXML := `<?xml version="1.0" encoding="UTF-8"?>
	<feed xmlns="http://www.w3.org/2005/Atom" xmlns:arxiv="http://arxiv.org/schemas/atom">
		<entry>
			<id>http://arxiv.org/abs/2301.00001v1</id>
			<published>2023-01-05T12:00:00Z</published>
			<title>
				Scalable Graph Neural Networks
				for Large Graphs
			</title>
			<summary>
				We introduce an efficient sampling method
				for scaling GNNs.
			</summary>
			<author>
				<name>Bob Smith</name>
			</author>
			<link title="pdf" href="http://arxiv.org/pdf/2301.00001v1.pdf" rel="related" type="application/pdf"/>
			<arxiv:primary_category term="cs.LG"/>
			<arxiv:doi>10.48550/arXiv.2301.00001</arxiv:doi>
		</entry>
	</feed>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/atom+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockXML))
	}))
	defer server.Close()

	// Use short interval for testing rate limiter
	provider := NewArxivProvider(server.URL, "test@example.com", server.Client())
	provider.rateLimiter = NewRateLimiter(0) // disable wait for fast test

	papers, err := provider.SearchPapers(context.Background(), "graph neural networks", 5)
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}

	p := papers[0]
	if p.Title != "Scalable Graph Neural Networks for Large Graphs" {
		t.Errorf("whitespace normalization failed on title: '%s'", p.Title)
	}
	if p.Abstract != "We introduce an efficient sampling method for scaling GNNs." {
		t.Errorf("whitespace normalization failed on abstract: '%s'", p.Abstract)
	}
	if p.Year != 2023 {
		t.Errorf("expected year 2023, got %d", p.Year)
	}
	if p.OpenAccessPdfURL != "http://arxiv.org/pdf/2301.00001v1.pdf" {
		t.Errorf("unexpected PDF url: %s", p.OpenAccessPdfURL)
	}
	if p.Doi != "10.48550/arXiv.2301.00001" {
		t.Errorf("unexpected DOI: %s", p.Doi)
	}
	if len(p.FieldsOfStudy) == 0 || p.FieldsOfStudy[0] != "cs.LG" {
		t.Errorf("unexpected fields of study: %+v", p.FieldsOfStudy)
	}
}

func TestCrossrefProvider_SearchPapers(t *testing.T) {
	mockJSON := `{
		"status": "ok",
		"message": {
			"total-results": 1,
			"items": [
				{
					"DOI": "10.1038/nature14539",
					"title": ["Deep Learning in Scientific Discovery"],
					"author": [
						{ "given": "Yann", "family": "LeCun" },
						{ "given": "Yoshua", "family": "Bengio" }
					],
					"abstract": "<jats:p>Deep learning enables computational models to learn representations of data.</jats:p>",
					"is-referenced-by-count": 42000,
					"container-title": ["Nature"],
					"published-print": {
						"date-parts": [[2015, 5, 28]]
					},
					"subject": ["Multidisciplinary"],
					"link": [
						{ "URL": "https://www.nature.com/articles/nature14539.pdf", "content-type": "application/pdf" }
					]
				}
			]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	provider := NewCrossrefProvider(server.URL, "test@example.com", server.Client())
	provider.rateLimiter = NewRateLimiter(0)

	papers, err := provider.SearchPapers(context.Background(), "deep learning nature", 5)
	if err != nil {
		t.Fatalf("unexpected search error: %v", err)
	}

	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}

	p := papers[0]
	if p.Title != "Deep Learning in Scientific Discovery" {
		t.Errorf("unexpected title: %s", p.Title)
	}
	if p.Abstract != "Deep learning enables computational models to learn representations of data." {
		t.Errorf("failed to strip JATS XML tags from abstract: '%s'", p.Abstract)
	}
	if p.Year != 2015 {
		t.Errorf("expected year 2015, got %d", p.Year)
	}
	if p.CitationCount != 42000 {
		t.Errorf("expected 42000 citations, got %d", p.CitationCount)
	}
	if p.Venue != "Nature" {
		t.Errorf("expected venue 'Nature', got '%s'", p.Venue)
	}
	if len(p.Authors) != 2 || p.Authors[0].Name != "Yann LeCun" {
		t.Errorf("unexpected authors: %+v", p.Authors)
	}
}

func TestFormatArxivQuery(t *testing.T) {
	singleTerm := formatArxivQuery("transformers")
	if singleTerm != "all:transformers" {
		t.Errorf("expected all:transformers, got %s", singleTerm)
	}

	multiTerm := formatArxivQuery("YOLO model caching")
	expected := "all:\"YOLO model caching\" OR (all:yolo AND all:model AND all:caching)"
	if multiTerm != expected {
		t.Errorf("expected '%s', got '%s'", expected, multiTerm)
	}
}
