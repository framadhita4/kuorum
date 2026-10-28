package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OpenAIProvider implements the LLMProvider interface for all OpenAI-compatible endpoints
// (OpenAI, OpenRouter, Groq, DeepSeek, Ollama, Together, vLLM, Perplexity, Mistral, etc.).
type OpenAIProvider struct {
	*BaseProvider
}

// NewOpenAIProvider creates an OpenAI-compatible provider.
func NewOpenAIProvider(
	name ProviderName,
	baseURL string,
	apiKey string,
	defaultModel string,
	httpClient *http.Client,
) *OpenAIProvider {
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}
	base := NewBaseProvider(name, EndpointOpenAI, baseURL, apiKey, defaultModel, 0, httpClient)
	return &OpenAIProvider{
		BaseProvider: base,
	}
}

// Request and response wire types for OpenAI protocol
type openAIChatRequest struct {
	Model               string          `json:"model"`
	Messages            []ChatMessage   `json:"messages"`
	Temperature         *float32        `json:"temperature,omitempty"`
	MaxTokens           *int            `json:"max_tokens,omitempty"`
	MaxCompletionTokens *int            `json:"max_completion_tokens,omitempty"`
	Stream              bool            `json:"stream"`
	ResponseFormat      *ResponseFormat `json:"response_format,omitempty"`
	ReasoningEffort     string          `json:"reasoning_effort,omitempty"`
}

type openAIChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	Delta        ChatMessage `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type openAIChatResponse struct {
	ID      string         `json:"id"`
	Model   string         `json:"model"`
	Choices []openAIChoice `json:"choices"`
	Usage   Usage          `json:"usage"`
}

func (p *OpenAIProvider) endpointURL() string {
	clean := strings.TrimRight(p.baseURL, "/")
	if strings.HasSuffix(clean, "/chat/completions") {
		return clean
	}
	return clean + "/chat/completions"
}

func (p *OpenAIProvider) buildPayload(req ChatRequest, stream bool) openAIChatRequest {
	model := req.Model
	if model == "" {
		model = p.defaultModel
	}

	effort := req.ReasoningEffort
	if effort == "" {
		effort = p.defaultReasoningEffort
	}

	// Consult model spec from catalog if registered to sanitize parameters
	modelSpec, hasSpec := GetModelSpec(string(p.name), model)
	if !hasSpec {
		modelSpec, _, hasSpec = FindModelSpec(model)
	}

	// If model does not support thinking, suppress reasoning_effort to avoid API 400 errors
	if hasSpec && modelSpec.ThinkingCapability == ThinkingNotSupported {
		effort = ReasoningEffortNone
	} else if hasSpec && modelSpec.ThinkingCapability == ThinkingRequired && effort == "" {
		effort = modelSpec.DefaultEffort
	}

	messages := make([]ChatMessage, len(req.Messages))
	copy(messages, req.Messages)

	// OpenAI requires that 'messages' contains the word 'json' when response_format is json_object
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_object" {
		hasJSON := false
		for _, m := range messages {
			if strings.Contains(strings.ToLower(m.Content), "json") {
				hasJSON = true
				break
			}
		}
		if !hasJSON {
			appended := false
			for i := range messages {
				if messages[i].Role == RoleSystem {
					messages[i].Content += "\n\nCRITICAL: Respond strictly with a valid JSON object."
					appended = true
					break
				}
			}
			if !appended {
				if len(messages) > 0 {
					messages[len(messages)-1].Content += "\n\nCRITICAL: Respond strictly with a valid JSON object."
				} else {
					messages = append(messages, ChatMessage{
						Role:    RoleSystem,
						Content: "CRITICAL: Respond strictly with a valid JSON object.",
					})
				}
			}
		}
	}

	payload := openAIChatRequest{
		Model:          model,
		Messages:       messages,
		Stream:         stream,
		ResponseFormat: req.ResponseFormat,
	}

	if effort != "" {
		payload.ReasoningEffort = string(effort)
		// Reasoning models forbid custom temperature
	} else if req.Temperature != nil {
		if !hasSpec || modelSpec.AllowsTemperature {
			payload.Temperature = req.Temperature
		}
	}

	var tokenLimit int
	if req.MaxTokens > 0 {
		tokenLimit = req.MaxTokens
	} else if hasSpec && modelSpec.DefaultMaxTokens > 0 {
		tokenLimit = modelSpec.DefaultMaxTokens
	}

	if tokenLimit > 0 {
		if p.useMaxCompletionTokens(model) {
			payload.MaxCompletionTokens = &tokenLimit
		} else {
			payload.MaxTokens = &tokenLimit
		}
	}

	return payload
}

func (p *OpenAIProvider) useMaxCompletionTokens(model string) bool {
	lowerModel := strings.ToLower(model)

	// 1. OpenAI o-series reasoning models (o1, o3, o4 series) always require max_completion_tokens
	if strings.HasPrefix(lowerModel, "o1") || strings.HasPrefix(lowerModel, "o3") || strings.HasPrefix(lowerModel, "o4") ||
		strings.Contains(lowerModel, "/o1") || strings.Contains(lowerModel, "/o3") || strings.Contains(lowerModel, "/o4") {
		return true
	}

	// 2. Official OpenAI endpoint or OpenAI provider name uses max_completion_tokens for all modern models
	if p.name == ProviderOpenAI || strings.Contains(strings.ToLower(string(p.name)), "openai") || strings.Contains(p.baseURL, "openai.com") {
		return true
	}

	// 3. All other OpenAI-compatible endpoints default to max_tokens
	return false
}

// Complete executes a non-streaming chat completion request.
func (p *OpenAIProvider) Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	payload := p.buildPayload(req, false)
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%s: marshal chat request failed: %w", p.name, err)
	}

	httpReq, err := p.NewRequest(ctx, http.MethodPost, p.endpointURL(), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	resp, err := p.Execute(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response body failed: %w", p.name, err)
	}

	var parsed openAIChatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response failed: %w (body: %s)", p.name, err, string(respBody))
	}

	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("%s: no choices returned in completion", p.name)
	}

	choice := parsed.Choices[0]
	return &ChatResponse{
		ID:           parsed.ID,
		Model:        parsed.Model,
		Role:         choice.Message.Role,
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		Usage:        parsed.Usage,
	}, nil
}

// Stream executes a streaming chat completion request, returning token and error channels.
func (p *OpenAIProvider) Stream(ctx context.Context, req ChatRequest) (<-chan string, <-chan error) {
	tokenChan := make(chan string, 100)
	errChan := make(chan error, 1)

	payload := p.buildPayload(req, true)
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		errChan <- fmt.Errorf("%s: marshal streaming request failed: %w", p.name, err)
		close(tokenChan)
		close(errChan)
		return tokenChan, errChan
	}

	httpReq, err := p.NewRequest(ctx, http.MethodPost, p.endpointURL(), bytes.NewReader(bodyBytes))
	if err != nil {
		errChan <- err
		close(tokenChan)
		close(errChan)
		return tokenChan, errChan
	}

	httpReq.Header.Set("Accept", "text/event-stream")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.apiKey))
	}

	resp, err := p.Execute(ctx, httpReq)
	if err != nil {
		errChan <- err
		close(tokenChan)
		close(errChan)
		return tokenChan, errChan
	}

	go p.readSSE(ctx, resp.Body, tokenChan, errChan)

	return tokenChan, errChan
}

func (p *OpenAIProvider) readSSE(
	ctx context.Context,
	body io.ReadCloser,
	tokenChan chan<- string,
	errChan chan<- error,
) {
	defer func() {
		_ = body.Close()
		close(tokenChan)
		close(errChan)
	}()

	reader := bufio.NewReader(body)

	for {
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				errChan <- fmt.Errorf("%s: sse stream read failed: %w", p.name, err)
			}
			return
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		lineStr := string(line)
		if !strings.HasPrefix(lineStr, "data:") {
			continue
		}

		dataPayload := strings.TrimSpace(strings.TrimPrefix(lineStr, "data:"))
		if dataPayload == "[DONE]" {
			return
		}

		var chunk openAIChatResponse
		if err := json.Unmarshal([]byte(dataPayload), &chunk); err != nil {
			continue
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				select {
				case <-ctx.Done():
					errChan <- ctx.Err()
					return
				case tokenChan <- choice.Delta.Content:
				}
			}
		}
	}
}
