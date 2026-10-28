package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// RateLimiter coordinates request intervals to comply with provider rate limits.
type RateLimiter struct {
	mu          sync.Mutex
	minInterval time.Duration
	lastRequest time.Time
}

// NewRateLimiter instantiates a new RateLimiter.
func NewRateLimiter(minInterval time.Duration) *RateLimiter {
	return &RateLimiter{
		minInterval: minInterval,
	}
}

// Wait pauses execution until the required interval has passed, or ctx is cancelled.
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

// BaseProvider encapsulates shared configuration, HTTP connection management,
// rate limiting, and request execution rules for all LLM providers.
type BaseProvider struct {
	name                   ProviderName
	endpointType           EndpointType
	baseURL                string
	apiKey                 string
	defaultModel           string
	defaultReasoningEffort ReasoningEffort
	httpClient             *http.Client
	rateLimiter            *RateLimiter
}

// NewBaseProvider creates a standardized base provider struct with connection pooling.
func NewBaseProvider(
	name ProviderName,
	endpointType EndpointType,
	baseURL string,
	apiKey string,
	defaultModel string,
	minInterval time.Duration,
	httpClient *http.Client,
) *BaseProvider {
	if httpClient == nil {
		transport := &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
		}
		httpClient = &http.Client{
			Transport: transport,
			Timeout:   120 * time.Second,
		}
	}

	cleanBaseURL := strings.TrimRight(baseURL, "/")

	return &BaseProvider{
		name:         name,
		endpointType: endpointType,
		baseURL:      cleanBaseURL,
		apiKey:       strings.TrimSpace(apiKey),
		defaultModel: strings.TrimSpace(defaultModel),
		httpClient:   httpClient,
		rateLimiter:  NewRateLimiter(minInterval),
	}
}

func (p *BaseProvider) Name() ProviderName {
	return p.name
}

func (p *BaseProvider) EndpointType() EndpointType {
	return p.endpointType
}

func (p *BaseProvider) BaseURL() string {
	return p.baseURL
}

func (p *BaseProvider) APIKey() string {
	return p.apiKey
}

func (p *BaseProvider) DefaultModel() string {
	return p.defaultModel
}

func (p *BaseProvider) HTTPClient() *http.Client {
	return p.httpClient
}

func (p *BaseProvider) RateLimiter() *RateLimiter {
	return p.rateLimiter
}

// SetRateLimiter allows overriding or disabling the rate limiter (e.g. for testing).
func (p *BaseProvider) SetRateLimiter(rl *RateLimiter) {
	p.rateLimiter = rl
}

// SetDefaultReasoningEffort sets default reasoning effort.
func (p *BaseProvider) SetDefaultReasoningEffort(effort ReasoningEffort) {
	p.defaultReasoningEffort = effort
}

// DefaultReasoningEffort returns default reasoning effort.
func (p *BaseProvider) DefaultReasoningEffort() ReasoningEffort {
	return p.defaultReasoningEffort
}

// NewRequest creates an http.Request with context and standard JSON headers.
func (p *BaseProvider) NewRequest(ctx context.Context, method, fullURL string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create request: %w", p.name, err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

// Execute enforces rate limiting and executes the HTTP request, inspecting status codes.
func (p *BaseProvider) Execute(ctx context.Context, req *http.Request) (*http.Response, error) {
	if p.rateLimiter != nil {
		if err := p.rateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("%s: rate limiter cancelled: %w", p.name, err)
		}
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: HTTP request failed: %w", p.name, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("%s: API error (status %d): %s", p.name, resp.StatusCode, string(snippet))
	}

	return resp, nil
}
