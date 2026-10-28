package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"kuorum/backend/internal/domain/retrieval"
)

var (
	xmlTagRegex        = regexp.MustCompile(`<[^>]*>`)
	whitespaceReplacer = regexp.MustCompile(`\s+`)
)

// RateLimiter manages request intervals to avoid IP blocks or HTTP 429s.
type RateLimiter struct {
	mu          sync.Mutex
	minInterval time.Duration
	lastRequest time.Time
}

// NewRateLimiter creates a new RateLimiter with a given minimum interval.
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
	}
}

// Wait blocks until the required interval has passed since the last request, or ctx is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	if rl == nil || rl.minInterval <= 0 {
		return nil
	}

	rl.mu.Lock()
	now := time.Now()
	elapsed := now.Sub(rl.lastRequest)

	var waitDuration time.Duration
	if elapsed < rl.minInterval {
		waitDuration = rl.minInterval - elapsed
		rl.lastRequest = now.Add(waitDuration)
	} else {
		rl.lastRequest = now
	}
	rl.mu.Unlock()

	if waitDuration > 0 {
		select {
		case <-time.After(waitDuration):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

// BaseProvider serves as the reusable base template struct that each academic
// provider embeds (inherits from) to share standard HTTP client mechanisms,
// polite pool headers, rate limiting, and response helpers.
type BaseProvider struct {
	name        ProviderName
	baseURL     string
	politeEmail string
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// NewBaseProvider instantiates the common provider base struct.
func NewBaseProvider(
	name ProviderName,
	baseURL string,
	politeEmail string,
	minInterval time.Duration,
	httpClient *http.Client,
) *BaseProvider {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	if politeEmail == "" {
		politeEmail = "kuorum@example.com"
	}

	cleanBaseURL := strings.TrimRight(baseURL, "/")

	return &BaseProvider{
		name:        name,
		baseURL:     cleanBaseURL,
		politeEmail: politeEmail,
		httpClient:  httpClient,
		rateLimiter: NewRateLimiter(minInterval),
	}
}

func (p *BaseProvider) Name() ProviderName {
	return p.name
}

func (p *BaseProvider) BaseURL() string {
	return p.baseURL
}

func (p *BaseProvider) PoliteEmail() string {
	return p.politeEmail
}

func (p *BaseProvider) HTTPClient() *http.Client {
	return p.httpClient
}

// SetRateLimiter allows overriding or resetting the rate limiter (e.g. for testing).
func (p *BaseProvider) SetRateLimiter(rl *RateLimiter) {
	p.rateLimiter = rl
}

// UserAgentHeader generates polite pool compliant User-Agent headers.
func (p *BaseProvider) UserAgentHeader() string {
	return fmt.Sprintf("kuorum/1.0 (mailto:%s)", p.politeEmail)
}

// NewRequest creates an http.Request with context and standard polite headers.
func (p *BaseProvider) NewRequest(ctx context.Context, method, fullURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", p.name, err)
	}

	req.Header.Set("User-Agent", p.UserAgentHeader())
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// Execute enforces rate limiting and executes the HTTP request.
func (p *BaseProvider) Execute(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := p.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%s: rate limiter cancelled: %w", p.name, err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: HTTP request failed: %w", p.name, err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: %w", p.name, retrieval.ErrRateLimited)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: %w", p.name, retrieval.ErrPaperNotFound)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodySnippet, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: unexpected status code %d: %s", p.name, resp.StatusCode, string(bodySnippet))
	}

	return resp, nil
}

// SanitizeText cleans excess spaces, tabs, and newlines from titles and abstracts.
func (p *BaseProvider) SanitizeText(text string) string {
	clean := whitespaceReplacer.ReplaceAllString(text, " ")
	return strings.TrimSpace(clean)
}

// StripXML removes XML / JATS formatting tags often found in abstracts.
func (p *BaseProvider) StripXML(text string) string {
	withoutTags := xmlTagRegex.ReplaceAllString(text, "")
	return p.SanitizeText(withoutTags)
}

// CleanDOI standardizes DOI strings by removing URL and prefix wrappers.
func (p *BaseProvider) CleanDOI(rawDOI string) string {
	doi := strings.TrimSpace(rawDOI)
	doi = strings.TrimPrefix(doi, "https://doi.org/")
	doi = strings.TrimPrefix(doi, "http://doi.org/")
	doi = strings.TrimPrefix(doi, "https://dx.doi.org/")
	doi = strings.TrimPrefix(doi, "http://dx.doi.org/")
	doi = strings.TrimPrefix(doi, "doi:")
	return strings.TrimSpace(doi)
}
