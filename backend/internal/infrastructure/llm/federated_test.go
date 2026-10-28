package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kuorum/backend/internal/infrastructure/config"
	"kuorum/backend/internal/infrastructure/llm/providers"
)

type sampleStructuredOutput struct {
	NormalizedTerm string   `json:"normalizedTerm"`
	Angles         []string `json:"angles"`
}

func TestFederatedClient_CompleteStructured(t *testing.T) {
	// Server returns markdown-wrapped JSON to test fence stripping
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		rawJSON := "```json\n{\"normalizedTerm\": \"Graph Neural Networks\", \"angles\": [\"theory\", \"application\"]}\n```"
		resp := map[string]any{
			"id": "chatcmpl-test",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": rawJSON,
					},
					"finish_reason": "stop",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := providers.NewOpenAIProvider(ProviderOpenAI, server.URL, "key", "gpt-4o-mini", server.Client())
	client := NewFederatedClient(p)

	var output sampleStructuredOutput
	err := client.CompleteStructured(context.Background(), "System prompt", "User prompt", &output)
	if err != nil {
		t.Fatalf("unexpected CompleteStructured error: %v", err)
	}

	if output.NormalizedTerm != "Graph Neural Networks" {
		t.Errorf("expected NormalizedTerm to be 'Graph Neural Networks', got '%s'", output.NormalizedTerm)
	}
	if len(output.Angles) != 2 || output.Angles[0] != "theory" {
		t.Errorf("expected 2 angles, got %+v", output.Angles)
	}
}

func TestFederatedClient_Fallback(t *testing.T) {
	// Server 1 (Primary): returns HTTP 429 Rate Limit
	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": "rate_limit_exceeded"}`, http.StatusTooManyRequests)
	}))
	defer s1.Close()

	// Server 2 (Fallback Anthropic): returns success
	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]any{
			"id":   "msg-fallback",
			"role": "assistant",
			"content": []map[string]any{
				{"type": "text", "text": "Anthropic fallback recovered successfully"},
			},
			"stop_reason": "end_turn",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer s2.Close()

	primary := providers.NewOpenAIProvider(ProviderOpenAI, s1.URL, "key", "gpt-4o", s1.Client())
	fallback := providers.NewAnthropicProvider(ProviderAnthropic, s2.URL, "key", "claude-3-7-sonnet", s2.Client())

	client := NewFederatedClient(primary, fallback)
	client.SetFallbackOrder(ProviderAnthropic)

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{{Role: RoleUser, Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("expected fallback to succeed, got error: %v", err)
	}

	if resp.Content != "Anthropic fallback recovered successfully" {
		t.Errorf("expected fallback content, got '%s'", resp.Content)
	}
}

func TestFederatedClient_ProviderManagement(t *testing.T) {
	p1 := providers.NewOpenAIProvider(ProviderOpenAI, "http://localhost:8080", "k1", "gpt-4o", nil)
	p2 := providers.NewAnthropicProvider(ProviderAnthropic, "http://localhost:8081", "k2", "claude-3", nil)

	client := NewFederatedClient(p1)
	if len(client.AvailableProviders()) != 1 {
		t.Errorf("expected 1 provider, got %d", len(client.AvailableProviders()))
	}

	client.RegisterProvider(p2)
	if len(client.AvailableProviders()) != 2 {
		t.Errorf("expected 2 providers, got %d", len(client.AvailableProviders()))
	}

	if err := client.SetActiveProvider(ProviderAnthropic); err != nil {
		t.Fatalf("failed to set active provider: %v", err)
	}

	active, ok := client.GetActiveProvider()
	if !ok || active.Name() != ProviderAnthropic {
		t.Errorf("expected active provider to be anthropic, got %v", active)
	}
}

func TestNewFederatedFromConfig(t *testing.T) {
	// Anthropic configuration
	cfgAnthropic := config.LLMConfig{
		EndpointType:    "anthropic",
		BaseURL:         "https://api.anthropic.com/v1",
		APIKey:          "sk-ant-test",
		Model:           "claude-3-7-sonnet-20250219",
		ReasoningEffort: "medium",
	}

	client := NewFederatedFromConfig(cfgAnthropic)
	active, ok := client.GetActiveProvider()
	if !ok {
		t.Fatal("expected active provider from config")
	}

	if active.EndpointType() != EndpointAnthropic {
		t.Errorf("expected endpoint type anthropic, got %s", active.EndpointType())
	}
	if active.DefaultModel() != "claude-3-7-sonnet-20250219" {
		t.Errorf("expected model claude-3-7-sonnet-20250219, got %s", active.DefaultModel())
	}

	// OpenAI configuration
	cfgOpenAI := config.LLMConfig{
		EndpointType: "openai",
		BaseURL:      "https://api.openai.com/v1",
		APIKey:       "sk-test",
		Model:        "gpt-4o-mini",
	}
	clientOpenAI := NewFederatedFromConfig(cfgOpenAI)
	activeOAI, ok := clientOpenAI.GetActiveProvider()
	if !ok || activeOAI.EndpointType() != EndpointOpenAI {
		t.Errorf("expected openai provider, got %v", activeOAI)
	}
}

func TestFederatedClient_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Federated"}}]}`,
			`data: {"choices":[{"delta":{"content":" Stream"}}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			_, _ = w.Write([]byte(chunk + "\n\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	p := providers.NewOpenAIProvider(ProviderOpenAI, server.URL, "key", "gpt-4o", server.Client())
	client := NewFederatedClient(p)

	tokenChan, errChan := client.Stream(context.Background(), "System prompt", "User prompt")

	var tokens []string
	for token := range tokenChan {
		tokens = append(tokens, token)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("stream error: %v", err)
	}

	if strings.Join(tokens, "") != "Federated Stream" {
		t.Errorf("expected 'Federated Stream', got '%s'", strings.Join(tokens, ""))
	}
}
