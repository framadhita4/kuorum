package academic

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"kuorum/backend/internal/domain/retrieval"

	"golang.org/x/sync/errgroup"
)

// FederatedClient is the main academic literature aggregator service ("Class main provider").
// It manages multiple academic providers (OpenAlex, Semantic Scholar, arXiv, Crossref) and
// allows clients to query one, several, or all providers concurrently with automatic deduplication
// and metadata consolidation.
type FederatedClient struct {
	mu        sync.RWMutex
	providers map[ProviderName]AcademicProvider
}

// NewFederatedClient instantiates the central federated academic manager with given providers.
func NewFederatedClient(providers ...AcademicProvider) *FederatedClient {
	fc := &FederatedClient{
		providers: make(map[ProviderName]AcademicProvider),
	}
	for _, p := range providers {
		if p != nil {
			fc.providers[p.Name()] = p
		}
	}
	return fc
}

// RegisterProvider dynamically adds or overrides an academic provider.
func (fc *FederatedClient) RegisterProvider(provider AcademicProvider) {
	if provider == nil {
		return
	}
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.providers[provider.Name()] = provider
}

// GetProvider retrieves an active provider by name.
func (fc *FederatedClient) GetProvider(name ProviderName) (AcademicProvider, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	p, ok := fc.providers[name]
	return p, ok
}

// AvailableProviders returns the names of all currently registered providers as strings.
func (fc *FederatedClient) AvailableProviders() []string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	names := make([]string, 0, len(fc.providers))
	for name := range fc.providers {
		names = append(names, string(name))
	}
	return names
}

// SearchPapers searches all registered providers concurrently and consolidates results.
func (fc *FederatedClient) SearchPapers(ctx context.Context, query string, limit int) ([]retrieval.PaperMetadata, error) {
	return fc.SearchPapersWithProviders(ctx, query, limit, nil)
}

// SearchPapersWithProviders searches a specified subset of providers (or all if empty/nil).
// Providers run in parallel using errgroup, and results are deduplicated and merged.
func (fc *FederatedClient) SearchPapersWithProviders(
	ctx context.Context,
	query string,
	limit int,
	providerNames []string,
) ([]retrieval.PaperMetadata, error) {
	if strings.TrimSpace(query) == "" {
		return nil, retrieval.ErrEmptyQuery
	}
	if limit <= 0 {
		limit = 10
	}

	fc.mu.RLock()
	selected := make([]AcademicProvider, 0)
	if len(providerNames) == 0 {
		// Default to all registered providers
		for _, p := range fc.providers {
			selected = append(selected, p)
		}
	} else {
		for _, rawName := range providerNames {
			norm, ok := NormalizeProviderName(rawName)
			if !ok {
				norm = ProviderName(strings.ToLower(strings.TrimSpace(rawName)))
			}
			if p, exists := fc.providers[norm]; exists {
				selected = append(selected, p)
			}
		}
	}
	fc.mu.RUnlock()

	if len(selected) == 0 {
		return nil, fmt.Errorf("%w: no matching providers found for request", retrieval.ErrProviderNotAvailable)
	}

	var mu sync.Mutex
	allPapers := make([]retrieval.PaperMetadata, 0)
	errs := make([]error, 0)

	g, gCtx := errgroup.WithContext(ctx)

	// Over-fetch candidate papers per provider to supply a rich candidate pool for re-ranking
	candidateLimit := max(limit, 20)
	if candidateLimit > 100 {
		candidateLimit = 100
	}

	for _, provider := range selected {
		p := provider
		g.Go(func() error {
			papers, err := p.SearchPapers(gCtx, query, candidateLimit)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s provider search failed: %w", p.Name(), err))
				mu.Unlock()
				log.Printf("[FederatedAcademic] Warning: provider %s search returned error: %v", p.Name(), err)
				return nil // Don't abort sibling providers on partial failure
			}

			mu.Lock()
			allPapers = append(allPapers, papers...)
			mu.Unlock()
			return nil
		})
	}

	_ = g.Wait()

	if len(allPapers) == 0 && len(errs) == len(selected) {
		return nil, fmt.Errorf("all selected providers failed: %v", errs)
	}

	merged := DeduplicateAndRank(allPapers, query)
	if len(merged) > limit && limit > 0 {
		merged = merged[:limit]
	}

	return merged, nil
}

// GetPaperDetails attempts to resolve paper details by checking registered providers.
func (fc *FederatedClient) GetPaperDetails(ctx context.Context, paperID string) (*retrieval.PaperMetadata, error) {
	cleanID := strings.TrimSpace(paperID)
	if cleanID == "" {
		return nil, fmt.Errorf("paper ID cannot be empty")
	}

	fc.mu.RLock()
	providersList := make([]AcademicProvider, 0, len(fc.providers))
	for _, p := range fc.providers {
		providersList = append(providersList, p)
	}
	fc.mu.RUnlock()

	// Prioritize provider matching ID prefix if recognizable
	var primaryProvider AcademicProvider
	if strings.HasPrefix(cleanID, "openalex:") || strings.Contains(cleanID, "openalex.org") {
		primaryProvider, _ = fc.GetProvider(ProviderOpenAlex)
	} else if strings.HasPrefix(cleanID, "arxiv:") || strings.Contains(cleanID, "arxiv.org") {
		primaryProvider, _ = fc.GetProvider(ProviderArxiv)
	} else if strings.HasPrefix(cleanID, "10.") || strings.Contains(cleanID, "doi.org") {
		primaryProvider, _ = fc.GetProvider(ProviderCrossref)
	}

	if primaryProvider != nil {
		paper, err := primaryProvider.GetPaperDetails(ctx, cleanID)
		if err == nil && paper != nil {
			return paper, nil
		}
	}

	// Try each provider in order
	for _, p := range providersList {
		paper, err := p.GetPaperDetails(ctx, cleanID)
		if err == nil && paper != nil {
			return paper, nil
		}
	}

	return nil, fmt.Errorf("%w: ID=%s", retrieval.ErrPaperNotFound, cleanID)
}

// GetReferences retrieves foundational reference papers (citations made by paperID).
func (fc *FederatedClient) GetReferences(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	fc.mu.RLock()
	s2, hasS2 := fc.providers[ProviderSemanticScholar]
	openAlex, hasOpenAlex := fc.providers[ProviderOpenAlex]
	crossref, hasCrossref := fc.providers[ProviderCrossref]
	fc.mu.RUnlock()

	// Priority 1: Semantic Scholar (specialized citation graphs)
	if hasS2 {
		refs, err := s2.GetReferences(ctx, paperID)
		if err == nil && len(refs) > 0 {
			return refs, nil
		}
	}

	// Priority 2: OpenAlex
	if hasOpenAlex {
		refs, err := openAlex.GetReferences(ctx, paperID)
		if err == nil && len(refs) > 0 {
			return refs, nil
		}
	}

	// Priority 3: Crossref
	if hasCrossref {
		refs, err := crossref.GetReferences(ctx, paperID)
		if err == nil && len(refs) > 0 {
			return refs, nil
		}
	}

	return []retrieval.PaperMetadata{}, nil
}

// GetCitations retrieves forward citations (papers that cite paperID).
func (fc *FederatedClient) GetCitations(ctx context.Context, paperID string) ([]retrieval.PaperMetadata, error) {
	fc.mu.RLock()
	s2, hasS2 := fc.providers[ProviderSemanticScholar]
	openAlex, hasOpenAlex := fc.providers[ProviderOpenAlex]
	fc.mu.RUnlock()

	// Priority 1: Semantic Scholar (supports citation intent, influential citations)
	if hasS2 {
		cits, err := s2.GetCitations(ctx, paperID)
		if err == nil && len(cits) > 0 {
			return cits, nil
		}
	}

	// Priority 2: OpenAlex
	if hasOpenAlex {
		cits, err := openAlex.GetCitations(ctx, paperID)
		if err == nil && len(cits) > 0 {
			return cits, nil
		}
	}

	return []retrieval.PaperMetadata{}, nil
}
