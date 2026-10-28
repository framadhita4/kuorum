package providers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAIProvider_Complete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Errorf("expected /chat/completions endpoint, got %s", r.URL.Path)
		}
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("expected Bearer test-key, got %s", auth)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		var req openAIChatRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to parse json body: %v", err)
		}

		if req.Model != "gpt-4o-mini" {
			t.Errorf("expected model gpt-4o-mini, got %s", req.Model)
		}

		w.Header().Set("Content-Type", "application/json")
		resp := openAIChatResponse{
			ID:    "chatcmpl-test-1",
			Model: "gpt-4o-mini",
			Choices: []openAIChoice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: `{"result": "success"}`,
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewOpenAIProvider("test-openai", server.URL, "test-key", "gpt-4o-mini", server.Client())

	req := ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleSystem, Content: "You are a test assistant"},
			{Role: RoleUser, Content: "Hello"},
		},
	}

	resp, err := p.Complete(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected completion error: %v", err)
	}

	if resp.Content != `{"result": "success"}` {
		t.Errorf("expected content to match, got %s", resp.Content)
	}
	if resp.Usage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestOpenAIProvider_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}

		chunks := []string{
			`data: {"choices":[{"delta":{"content":"Hello"}}]}`,
			`data: {"choices":[{"delta":{"content":" World"}}]}`,
			`data: [DONE]`,
		}

		for _, chunk := range chunks {
			_, _ = w.Write([]byte(chunk + "\n\n"))
			flusher.Flush()
		}
	}))
	defer server.Close()

	p := NewOpenAIProvider("test-openai", server.URL, "test-key", "gpt-4o-mini", server.Client())

	tokenChan, errChan := p.Stream(context.Background(), ChatRequest{
		Messages: []ChatMessage{{Role: RoleUser, Content: "Hi"}},
	})

	var tokens []string
	for token := range tokenChan {
		tokens = append(tokens, token)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("stream returned unexpected error: %v", err)
	}

	joined := strings.Join(tokens, "")
	if joined != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", joined)
	}
}

func TestOpenAIProvider_ReasoningEffort(t *testing.T) {
	var capturedPayload openAIChatRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openAIChatResponse{
			Choices: []openAIChoice{{Message: ChatMessage{Content: "Reasoning result"}}},
		})
	}))
	defer server.Close()

	p := NewOpenAIProvider("test-o1", server.URL, "key", "o3-mini", server.Client())

	_, err := p.Complete(context.Background(), ChatRequest{
		ReasoningEffort: ReasoningEffortHigh,
		Messages:        []ChatMessage{{Role: RoleUser, Content: "Solve this theorem"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPayload.ReasoningEffort != "high" {
		t.Errorf("expected reasoning_effort 'high', got '%s'", capturedPayload.ReasoningEffort)
	}
	if capturedPayload.Temperature != nil {
		t.Errorf("temperature should be omitted for reasoning models")
	}
}

func TestOpenAIProvider_JSONObject_EnforcesJSONInMessages(t *testing.T) {
	var capturedPayload openAIChatRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openAIChatResponse{
			Choices: []openAIChoice{{Message: ChatMessage{Content: `{"ok": true}`}}},
		})
	}))
	defer server.Close()

	p := NewOpenAIProvider("openai", server.URL, "key", "gpt-4o", server.Client())

	// Test 1: Messages without "json" when ResponseFormat is json_object
	_, err := p.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleSystem, Content: "You are an assistant."},
			{Role: RoleUser, Content: "Extract findings."},
		},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasJSON := false
	for _, m := range capturedPayload.Messages {
		if strings.Contains(strings.ToLower(m.Content), "json") {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		t.Fatalf("expected payload messages to contain 'json' for json_object response format")
	}

	// Test 2: Messages with only User role (no System message)
	_, err = p.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Summarize this paper."},
		},
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	hasJSON = false
	for _, m := range capturedPayload.Messages {
		if strings.Contains(strings.ToLower(m.Content), "json") {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		t.Fatalf("expected payload messages without system prompt to still contain 'json'")
	}
}

func TestAnthropicProvider_Complete(t *testing.T) {
	var capturedReq anthropicMessagesRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "claude-secret-key" {
			t.Errorf("expected x-api-key claude-secret-key, got %s", r.Header.Get("x-api-key"))
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version 2023-06-01, got %s", r.Header.Get("anthropic-version"))
		}
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			t.Errorf("expected /v1/messages, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedReq)

		w.Header().Set("Content-Type", "application/json")
		resp := anthropicMessagesResponse{
			ID:   "msg_12345",
			Role: "assistant",
			Content: []anthropicContentBlock{
				{Type: "text", Text: "Claude response here"},
			},
			Model:      "claude-3-7-sonnet-20250219",
			StopReason: "end_turn",
			Usage: anthropicUsage{
				InputTokens:  12,
				OutputTokens: 24,
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := NewAnthropicProvider("test-claude", server.URL, "claude-secret-key", "claude-3-7-sonnet-20250219", server.Client())

	resp, err := p.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleSystem, Content: "System prompt instructions"},
			{Role: RoleUser, Content: "User question"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected completion error: %v", err)
	}

	if capturedReq.System != "System prompt instructions" {
		t.Errorf("expected system prompt to be extracted to top level, got '%s'", capturedReq.System)
	}
	if len(capturedReq.Messages) != 1 || capturedReq.Messages[0].Content != "User question" {
		t.Errorf("expected 1 user message, got %+v", capturedReq.Messages)
	}
	if resp.Content != "Claude response here" {
		t.Errorf("expected content to match, got '%s'", resp.Content)
	}
	if resp.Usage.TotalTokens != 36 {
		t.Errorf("expected 36 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestAnthropicProvider_ReasoningThinking(t *testing.T) {
	var capturedReq anthropicMessagesRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedReq)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(anthropicMessagesResponse{
			Content: []anthropicContentBlock{{Type: "text", Text: "Solved"}},
		})
	}))
	defer server.Close()

	p := NewAnthropicProvider("test-claude-thinking", server.URL, "key", "claude-3-7-sonnet-20250219", server.Client())

	_, err := p.Complete(context.Background(), ChatRequest{
		ReasoningEffort: ReasoningEffortMedium,
		Messages:        []ChatMessage{{Role: RoleUser, Content: "Analyze code"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Thinking == nil || capturedReq.Thinking.BudgetTokens != 4096 {
		t.Errorf("expected thinking budget 4096, got %+v", capturedReq.Thinking)
	}
	if capturedReq.MaxTokens < 4096 {
		t.Errorf("expected max_tokens > budget_tokens, got %d", capturedReq.MaxTokens)
	}
	if capturedReq.Temperature != nil {
		t.Errorf("temperature should be omitted when thinking is enabled")
	}
}

func TestAnthropicProvider_Stream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}

		events := []string{
			"event: message_start\ndata: {\"type\":\"message_start\"}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\"Bonjour\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"text_delta\",\"text\":\" le monde\"}}\n\n",
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		}

		for _, ev := range events {
			_, _ = w.Write([]byte(ev))
			flusher.Flush()
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer server.Close()

	p := NewAnthropicProvider("test-claude-stream", server.URL, "key", "claude-3-7-sonnet-20250219", server.Client())

	tokenChan, errChan := p.Stream(context.Background(), ChatRequest{
		Messages: []ChatMessage{{Role: RoleUser, Content: "Hello"}},
	})

	var tokens []string
	for token := range tokenChan {
		tokens = append(tokens, token)
	}

	if err := <-errChan; err != nil {
		t.Fatalf("stream returned unexpected error: %v", err)
	}

	joined := strings.Join(tokens, "")
	if joined != "Bonjour le monde" {
		t.Errorf("expected 'Bonjour le monde', got '%s'", joined)
	}
}

func TestPresets_Catalog(t *testing.T) {
	groq, ok := GetPreset("groq")
	if !ok {
		t.Fatal("expected groq preset to be found")
	}
	if groq.EndpointType != EndpointOpenAI {
		t.Errorf("expected Groq to use OpenAI endpoint, got %s", groq.EndpointType)
	}

	claude, ok := GetPreset("anthropic")
	if !ok {
		t.Fatal("expected anthropic preset to be found")
	}
	if claude.EndpointType != EndpointAnthropic {
		t.Errorf("expected Anthropic to use Anthropic endpoint, got %s", claude.EndpointType)
	}

	// Test NewProviderFromPreset
	p, err := NewProviderFromPreset("openrouter", "sk-openrouter-key")
	if err != nil {
		t.Fatalf("unexpected error creating provider from preset: %v", err)
	}
	if p.EndpointType() != EndpointOpenAI {
		t.Errorf("expected openrouter to be openai endpoint")
	}

	// Test NewProviderFromConfig
	custom, err := NewProviderFromConfig(ProviderConfig{
		Name:         "my-custom-vllm",
		EndpointType: "openai",
		BaseURL:      "http://localhost:8000/v1",
		Model:        "meta-llama/Llama-3-70b",
	})
	if err != nil {
		t.Fatalf("unexpected error creating custom provider: %v", err)
	}
	if custom.Name() != "my-custom-vllm" {
		t.Errorf("expected name my-custom-vllm, got %s", custom.Name())
	}
}

func TestCatalog_ModelsAndCapabilities(t *testing.T) {
	// 1. OpenAI models
	models := DefaultCatalog().ListModels("openai")
	if len(models) < 3 {
		t.Errorf("expected at least 3 models under OpenAI, got %d", len(models))
	}

	o3, ok := GetModelSpec("openai", "o3-mini")
	if !ok || o3.ThinkingCapability != ThinkingRequired {
		t.Errorf("expected o3-mini to have ThinkingRequired, got %+v", o3)
	}

	gpt4o, ok := GetModelSpec("openai", "gpt-4o")
	if !ok || gpt4o.ThinkingCapability != ThinkingNotSupported {
		t.Errorf("expected gpt-4o to have ThinkingNotSupported, got %+v", gpt4o)
	}

	// 2. Anthropic models
	claude37, ok := GetModelSpec("anthropic", "claude-3-7-sonnet-20250219")
	if !ok || claude37.ThinkingCapability != ThinkingOptional {
		t.Errorf("expected claude-3-7-sonnet to have ThinkingOptional, got %+v", claude37)
	}

	// 3. FindModelSpec cross-lookup
	foundModel, foundProvider, ok := FindModelSpec("deepseek-reasoner")
	if !ok || foundProvider.Name != ProviderDeepSeek || foundModel.ThinkingCapability != ThinkingRequired {
		t.Errorf("failed to find deepseek-reasoner cross-lookup: %+v, %+v", foundModel, foundProvider)
	}
}

func TestCatalog_APIKeyResolution(t *testing.T) {
	// Test environment variable priority resolution
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-from-env")
	t.Setenv("OPENAI_API_KEY", "sk-oai-from-env")
	t.Setenv("GROQ_API_KEY", "gsk-groq-from-env")

	// Resolving without explicit key should pick up provider's specific env var
	antKey := ResolveAPIKey(string(ProviderAnthropic), "")
	if antKey != "sk-ant-from-env" {
		t.Errorf("expected sk-ant-from-env, got '%s'", antKey)
	}

	oaiKey := ResolveAPIKey(string(ProviderOpenAI), "")
	if oaiKey != "sk-oai-from-env" {
		t.Errorf("expected sk-oai-from-env, got '%s'", oaiKey)
	}

	groqKey := ResolveAPIKey(string(ProviderGroq), "")
	if groqKey != "gsk-groq-from-env" {
		t.Errorf("expected gsk-groq-from-env, got '%s'", groqKey)
	}

	// Explicit key should always take precedence
	explicit := ResolveAPIKey(string(ProviderAnthropic), "explicit-key-override")
	if explicit != "explicit-key-override" {
		t.Errorf("expected explicit-key-override, got '%s'", explicit)
	}
}

func TestOpenAIProvider_SanitizeNonThinkingModel(t *testing.T) {
	var capturedPayload openAIChatRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openAIChatResponse{
			Choices: []openAIChoice{{Message: ChatMessage{Content: "Sanitized result"}}},
		})
	}))
	defer server.Close()

	p := NewOpenAIProvider("openai", server.URL, "key", "gpt-4o", server.Client())

	// gpt-4o does NOT support thinking tokens. ReasoningEffort should be stripped to avoid 400 errors.
	temp := float32(0.5)
	_, err := p.Complete(context.Background(), ChatRequest{
		Model:           "gpt-4o",
		ReasoningEffort: ReasoningEffortHigh,
		Temperature:     &temp,
		Messages:        []ChatMessage{{Role: RoleUser, Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedPayload.ReasoningEffort != "" {
		t.Errorf("expected reasoning_effort to be stripped for gpt-4o, got '%s'", capturedPayload.ReasoningEffort)
	}
	if capturedPayload.Temperature == nil || *capturedPayload.Temperature != 0.5 {
		t.Errorf("expected temperature 0.5 preserved for gpt-4o")
	}
}

func TestAnthropicProvider_SanitizeNonThinkingModel(t *testing.T) {
	var capturedReq anthropicMessagesRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedReq)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(anthropicMessagesResponse{
			Content: []anthropicContentBlock{{Type: "text", Text: "Claude sanitized"}},
		})
	}))
	defer server.Close()

	p := NewAnthropicProvider("anthropic", server.URL, "key", "claude-3-5-sonnet-20241022", server.Client())

	// claude-3-5-sonnet does NOT support thinking tokens.
	_, err := p.Complete(context.Background(), ChatRequest{
		Model:           "claude-3-5-sonnet-20241022",
		ReasoningEffort: ReasoningEffortHigh,
		Messages:        []ChatMessage{{Role: RoleUser, Content: "Hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.Thinking != nil {
		t.Errorf("expected thinking block to be omitted for claude-3-5-sonnet, got %+v", capturedReq.Thinking)
	}
}

func TestCatalog_FetchLiveModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/models") {
			t.Errorf("expected /models endpoint, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"gpt-4o"},{"id":"o3-mini"},{"id":"gpt-4.5-preview"}]}`))
	}))
	defer server.Close()

	customCatalog := NewCatalog()
	customCatalog.RegisterProvider(ProviderSpec{
		Name:           "mock-openai",
		EndpointType:   EndpointOpenAI,
		DefaultBaseURL: server.URL,
	})

	models, err := customCatalog.FetchLiveModels(context.Background(), "mock-openai", "mock-key")
	if err != nil {
		t.Fatalf("unexpected error fetching live models: %v", err)
	}

	if len(models) != 3 || models[0] != "gpt-4o" || models[1] != "o3-mini" || models[2] != "gpt-4.5-preview" {
		t.Errorf("unexpected models returned: %+v", models)
	}
}

func TestCatalog_GetAvailableModels_KeyFiltering(t *testing.T) {
	// Clear any keys that might be in env for the test
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	t.Setenv("CLAUDE_API_KEY", "")
	t.Setenv("GROQ_API_KEY", "")
	t.Setenv("DEEPSEEK_API_KEY", "")
	t.Setenv("OPENROUTER_API_KEY", "")
	t.Setenv("TOGETHER_API_KEY", "")
	t.Setenv("MISTRAL_API_KEY", "")
	t.Setenv("PERPLEXITY_API_KEY", "")
	t.Setenv("PPLX_API_KEY", "")
	t.Setenv("LLM_API_KEY", "")

	catalog := DefaultCatalog()

	// With no keys set, no models or providers should be returned
	models, providers, defModel := catalog.GetAvailableModels()
	if len(models) != 0 || len(providers) != 0 {
		t.Errorf("expected 0 models and 0 providers when no keys configured, got %d models, %d providers", len(models), len(providers))
	}
	if defModel != "" {
		t.Errorf("expected empty default model when no keys configured, got %s", defModel)
	}

	// Set only GEMINI_API_KEY
	t.Setenv("GEMINI_API_KEY", "test-gemini-key")
	models, providers, defModel = catalog.GetAvailableModels()
	if len(providers) != 1 {
		t.Fatalf("expected exactly 1 provider, got %d", len(providers))
	}
	if providers[0].ID != "gemini" {
		t.Errorf("expected provider ID 'gemini', got %s", providers[0].ID)
	}
	if len(models) == 0 {
		t.Errorf("expected Gemini models to be returned, got 0")
	}
	for _, m := range models {
		if m.Provider != "gemini" {
			t.Errorf("expected only gemini models, got provider %s for model %s", m.Provider, m.ID)
		}
	}
	if defModel != "gemini-2.0-flash" {
		t.Errorf("expected default model to be gemini-2.0-flash, got %s", defModel)
	}
}

func TestOpenAIProvider_MaxTokensVsMaxCompletionTokens(t *testing.T) {
	var capturedRawBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		freshMap := make(map[string]any)
		_ = json.Unmarshal(body, &freshMap)
		capturedRawBody = freshMap

		// Enforce OpenAI rule: Setting both at the same time is forbidden and returns 400
		_, hasMaxTokens := capturedRawBody["max_tokens"]
		_, hasMaxCompletionTokens := capturedRawBody["max_completion_tokens"]
		if hasMaxTokens && hasMaxCompletionTokens {
			http.Error(w, `{"error":{"message":"Setting 'max_tokens' and 'max_completion_tokens' at the same time is not supported.","type":"invalid_request_error","param":"max_tokens","code":"invalid_parameter_combination"}}`, http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(openAIChatResponse{
			Choices: []openAIChoice{{Message: ChatMessage{Content: "OK"}}},
		})
	}))
	defer server.Close()

	t.Run("OpenAI standard model uses only max_completion_tokens", func(t *testing.T) {
		p := NewOpenAIProvider(ProviderOpenAI, server.URL, "key", "gpt-4o-mini", server.Client())
		_, err := p.Complete(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: RoleUser, Content: "hi"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := capturedRawBody["max_tokens"]; ok {
			t.Errorf("expected max_tokens to be omitted for OpenAI")
		}
		if _, ok := capturedRawBody["max_completion_tokens"]; !ok {
			t.Errorf("expected max_completion_tokens to be present for OpenAI")
		}
	})

	t.Run("OpenAI reasoning model uses only max_completion_tokens", func(t *testing.T) {
		p := NewOpenAIProvider(ProviderOpenAI, server.URL, "key", "o3-mini", server.Client())
		_, err := p.Complete(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: RoleUser, Content: "solve"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := capturedRawBody["max_tokens"]; ok {
			t.Errorf("expected max_tokens to be omitted for OpenAI o3-mini")
		}
		if _, ok := capturedRawBody["max_completion_tokens"]; !ok {
			t.Errorf("expected max_completion_tokens to be present for OpenAI o3-mini")
		}
	})

	t.Run("Third-party provider uses max_tokens", func(t *testing.T) {
		p := NewOpenAIProvider(ProviderGroq, server.URL, "key", "llama-3.3-70b-versatile", server.Client())
		_, err := p.Complete(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: RoleUser, Content: "hello"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := capturedRawBody["max_completion_tokens"]; ok {
			t.Errorf("expected max_completion_tokens to be omitted for Groq")
		}
		if _, ok := capturedRawBody["max_tokens"]; !ok {
			t.Errorf("expected max_tokens to be present for Groq")
		}
	})

	t.Run("OpenRouter routing OpenAI o-series uses max_completion_tokens", func(t *testing.T) {
		p := NewOpenAIProvider(ProviderOpenRouter, server.URL, "key", "openai/o3-mini", server.Client())
		_, err := p.Complete(context.Background(), ChatRequest{
			Messages: []ChatMessage{{Role: RoleUser, Content: "hello"}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := capturedRawBody["max_tokens"]; ok {
			t.Errorf("expected max_tokens to be omitted for OpenRouter o3-mini")
		}
		if _, ok := capturedRawBody["max_completion_tokens"]; !ok {
			t.Errorf("expected max_completion_tokens to be present for OpenRouter o3-mini")
		}
	})
}
