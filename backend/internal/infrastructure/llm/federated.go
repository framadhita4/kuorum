package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"kuorum/backend/internal/application/common/interfaces"
)

// FederatedClient coordinates multiple LLM providers (OpenAI-compatible and Anthropic-compatible),
// supporting active provider selection, automatic fallback failover, and structured JSON output unmarshaling.
// Mirrored after FederatedClient in internal/infrastructure/academic.
type FederatedClient struct {
	mu                sync.RWMutex
	providers         map[ProviderName]LLMProvider
	activeProvider    ProviderName
	fallbackProviders []ProviderName
}

// NewFederatedClient instantiates the central federated LLM manager with the given providers.
func NewFederatedClient(llmProviders ...LLMProvider) *FederatedClient {
	fc := &FederatedClient{
		providers:         make(map[ProviderName]LLMProvider),
		fallbackProviders: make([]ProviderName, 0),
	}

	for _, p := range llmProviders {
		if p != nil {
			fc.providers[p.Name()] = p
			if fc.activeProvider == "" {
				fc.activeProvider = p.Name()
			}
		}
	}

	return fc
}

// RegisterProvider dynamically adds or overrides an LLM provider.
func (fc *FederatedClient) RegisterProvider(provider LLMProvider) {
	if provider == nil {
		return
	}
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.providers[provider.Name()] = provider
	if fc.activeProvider == "" {
		fc.activeProvider = provider.Name()
	}
}

// SetActiveProvider sets the primary provider to use for completions and streams.
func (fc *FederatedClient) SetActiveProvider(name ProviderName) error {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	norm, _ := NormalizeProviderName(string(name))
	if _, exists := fc.providers[norm]; exists {
		fc.activeProvider = norm
		return nil
	}
	if _, exists := fc.providers[name]; exists {
		fc.activeProvider = name
		return nil
	}

	return fmt.Errorf("provider %s is not registered", name)
}

// GetActiveProvider returns the currently designated primary provider.
func (fc *FederatedClient) GetActiveProvider() (LLMProvider, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	p, ok := fc.providers[fc.activeProvider]
	return p, ok
}

// GetProvider retrieves an active provider by name.
func (fc *FederatedClient) GetProvider(name ProviderName) (LLMProvider, bool) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	norm, _ := NormalizeProviderName(string(name))
	if p, ok := fc.providers[norm]; ok {
		return p, true
	}
	p, ok := fc.providers[name]
	return p, ok
}

// AvailableProviders returns the names of all currently registered providers.
func (fc *FederatedClient) AvailableProviders() []string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	names := make([]string, 0, len(fc.providers))
	for name := range fc.providers {
		names = append(names, string(name))
	}
	return names
}

// SetFallbackOrder configures the fallback sequence to try when the primary provider encounters an error.
func (fc *FederatedClient) SetFallbackOrder(providerNames ...ProviderName) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.fallbackProviders = make([]ProviderName, 0, len(providerNames))
	for _, raw := range providerNames {
		norm, ok := NormalizeProviderName(string(raw))
		if !ok {
			norm = raw
		}
		fc.fallbackProviders = append(fc.fallbackProviders, norm)
	}
}

// FallbackOrder returns the configured fallback provider names.
func (fc *FederatedClient) FallbackOrder() []string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	res := make([]string, len(fc.fallbackProviders))
	for i, name := range fc.fallbackProviders {
		res[i] = string(name)
	}
	return res
}

// Complete executes a chat completion using the active provider (or dynamically routing to the model's provider),
// with automatic failover to fallback providers.
func (fc *FederatedClient) Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	fc.mu.RLock()
	targetName := fc.activeProvider

	// If request specifies a model, dynamically route to the registered provider that owns it
	if req.Model != "" {
		if _, pSpec, ok := FindModelSpec(req.Model); ok {
			if _, exists := fc.providers[pSpec.Name]; exists {
				targetName = pSpec.Name
			}
		}
	}

	active, hasActive := fc.providers[targetName]
	fallbacks := make([]LLMProvider, 0, len(fc.fallbackProviders))
	for _, fbName := range fc.fallbackProviders {
		if fb, ok := fc.providers[fbName]; ok && fbName != targetName {
			fallbacks = append(fallbacks, fb)
		}
	}
	fc.mu.RUnlock()

	if !hasActive {
		return nil, fmt.Errorf("no active LLM provider configured")
	}

	resp, err := active.Complete(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Active provider failed; attempt fallbacks in order
	log.Printf("[FederatedLLM] Active provider %s failed: %v. Attempting %d fallbacks...", targetName, err, len(fallbacks))
	for _, fb := range fallbacks {
		log.Printf("[FederatedLLM] Trying fallback provider %s...", fb.Name())
		fbResp, fbErr := fb.Complete(ctx, req)
		if fbErr == nil {
			return fbResp, nil
		}
		log.Printf("[FederatedLLM] Fallback provider %s failed: %v", fb.Name(), fbErr)
	}

	return nil, fmt.Errorf("all providers failed: primary (%s): %w", targetName, err)
}

// CompleteWithProvider executes a chat completion directly with a specific provider.
func (fc *FederatedClient) CompleteWithProvider(
	ctx context.Context,
	providerName ProviderName,
	req ChatRequest,
) (*ChatResponse, error) {
	p, ok := fc.GetProvider(providerName)
	if !ok {
		return nil, fmt.Errorf("provider %s is not registered", providerName)
	}
	return p.Complete(ctx, req)
}

// StreamChat starts a token stream with the active provider (or dynamically routing to the model's provider).
func (fc *FederatedClient) StreamChat(ctx context.Context, req ChatRequest) (<-chan string, <-chan error) {
	fc.mu.RLock()
	targetName := fc.activeProvider

	if req.Model != "" {
		if _, pSpec, ok := FindModelSpec(req.Model); ok {
			if _, exists := fc.providers[pSpec.Name]; exists {
				targetName = pSpec.Name
			}
		}
	}

	active, hasActive := fc.providers[targetName]
	activeName := targetName
	fc.mu.RUnlock()

	if !hasActive {
		tokenChan := make(chan string)
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("no active LLM provider configured (tried %s)", activeName)
		close(tokenChan)
		close(errChan)
		return tokenChan, errChan
	}

	return active.Stream(ctx, req)
}

// StreamWithProvider starts a token stream with a designated provider.
func (fc *FederatedClient) StreamWithProvider(
	ctx context.Context,
	providerName ProviderName,
	req ChatRequest,
) (<-chan string, <-chan error) {
	p, ok := fc.GetProvider(providerName)
	if !ok {
		tokenChan := make(chan string)
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("provider %s is not registered", providerName)
		close(tokenChan)
		close(errChan)
		return tokenChan, errChan
	}
	return p.Stream(ctx, req)
}

// CompleteStructured implements interfaces.ILLMClient by requesting structured JSON adhering to a target schema.
func (fc *FederatedClient) CompleteStructured(ctx context.Context, systemPrompt, userPrompt string, target any) error {
	sys := systemPrompt
	if !strings.Contains(strings.ToLower(sys), "json") && !strings.Contains(strings.ToLower(userPrompt), "json") {
		if sys != "" {
			sys += "\n\nCRITICAL: Respond strictly with a valid JSON object."
		} else {
			sys = "CRITICAL: Respond strictly with a valid JSON object."
		}
	}

	req := ChatRequest{
		Model: interfaces.ModelFromContext(ctx),
		Messages: []ChatMessage{
			{Role: RoleSystem, Content: sys},
			{Role: RoleUser, Content: userPrompt},
		},
		ResponseFormat: &ResponseFormat{
			Type: "json_object",
		},
	}

	resp, err := fc.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("llm complete structured failed: %w", err)
	}

	cleanJSON := cleanJSONResponse(resp.Content)
	if err := json.Unmarshal([]byte(cleanJSON), target); err != nil {
		return fmt.Errorf("unmarshal structured target failed: %w (raw response: %s)", err, resp.Content)
	}

	return nil
}

// Stream implements interfaces.ILLMClient by streaming tokens in real-time through a Go channel.
func (fc *FederatedClient) Stream(ctx context.Context, systemPrompt, userPrompt string) (<-chan string, <-chan error) {
	req := ChatRequest{
		Model: interfaces.ModelFromContext(ctx),
		Messages: []ChatMessage{
			{Role: RoleSystem, Content: systemPrompt},
			{Role: RoleUser, Content: userPrompt},
		},
	}
	return fc.StreamChat(ctx, req)
}

// cleanJSONResponse extracts clean JSON text from LLM outputs, stripping markdown code block fences if present.
func cleanJSONResponse(raw string) string {
	trimmed := strings.TrimSpace(raw)

	if strings.HasPrefix(trimmed, "```") {
		lines := strings.Split(trimmed, "\n")
		if len(lines) >= 2 {
			if strings.HasPrefix(lines[0], "```") {
				lines = lines[1:]
			}
			if len(lines) > 0 && strings.HasPrefix(lines[len(lines)-1], "```") {
				lines = lines[:len(lines)-1]
			}
			trimmed = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}

	// Find the bounds of the outermost JSON object or array
	start := strings.IndexAny(trimmed, "{[")
	end := strings.LastIndexAny(trimmed, "}]")
	if start != -1 && end != -1 && end >= start {
		return trimmed[start : end+1]
	}

	return trimmed
}
