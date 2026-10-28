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

// AnthropicProvider implements the LLMProvider interface for Anthropic-compatible endpoints.
type AnthropicProvider struct {
	*BaseProvider
	anthropicVersion string
}

// NewAnthropicProvider creates an Anthropic-compatible provider.
func NewAnthropicProvider(
	name ProviderName,
	baseURL string,
	apiKey string,
	defaultModel string,
	httpClient *http.Client,
) *AnthropicProvider {
	if defaultModel == "" {
		defaultModel = "claude-3-7-sonnet-20250219"
	}
	base := NewBaseProvider(name, EndpointAnthropic, baseURL, apiKey, defaultModel, 0, httpClient)
	return &AnthropicProvider{
		BaseProvider:     base,
		anthropicVersion: "2023-06-01",
	}
}

// Wire request structures for Anthropic Messages API
type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicThinking struct {
	Type         string `json:"type"` // "enabled"
	BudgetTokens int    `json:"budget_tokens"`
}

type anthropicMessagesRequest struct {
	Model       string             `json:"model"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature *float32           `json:"temperature,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
	Thinking    *anthropicThinking `json:"thinking,omitempty"`
}

// Wire response structures for Anthropic Messages API
type anthropicContentBlock struct {
	Type     string `json:"type"` // "text", "thinking"
	Text     string `json:"text,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicMessagesResponse struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Role       string                  `json:"role"`
	Content    []anthropicContentBlock `json:"content"`
	Model      string                  `json:"model"`
	StopReason string                  `json:"stop_reason"`
	Usage      anthropicUsage          `json:"usage"`
}

// Wire structures for Anthropic SSE streaming
type anthropicDelta struct {
	Type         string `json:"type"` // "text_delta", "thinking_delta"
	Text         string `json:"text,omitempty"`
	Thinking     string `json:"thinking,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

type anthropicStreamEvent struct {
	Type  string         `json:"type"` // "content_block_delta", "message_delta", "message_stop"
	Delta anthropicDelta `json:"delta,omitempty"`
}

func (p *AnthropicProvider) endpointURL() string {
	clean := strings.TrimRight(p.baseURL, "/")
	if strings.HasSuffix(clean, "/messages") || strings.HasSuffix(clean, "/v1/messages") {
		return clean
	}
	if strings.HasSuffix(clean, "/v1") {
		return clean + "/messages"
	}
	return clean + "/v1/messages"
}

func (p *AnthropicProvider) buildPayload(req ChatRequest, stream bool) anthropicMessagesRequest {
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

	// If model does not support thinking, suppress reasoning effort
	if hasSpec && modelSpec.ThinkingCapability == ThinkingNotSupported {
		effort = ReasoningEffortNone
		req.ThinkingBudget = 0
	} else if hasSpec && modelSpec.ThinkingCapability == ThinkingRequired && effort == "" && req.ThinkingBudget <= 0 {
		effort = modelSpec.DefaultEffort
	}

	// 1. Separate system prompt from messages (Anthropic treats system as a top-level field)
	var systemParts []string
	var anthropicMsgs []anthropicMessage

	for _, msg := range req.Messages {
		if strings.EqualFold(msg.Role, RoleSystem) {
			systemParts = append(systemParts, msg.Content)
		} else {
			anthropicMsgs = append(anthropicMsgs, anthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// If structured output requested, reinforce via system prompt
	if req.ResponseFormat != nil && req.ResponseFormat.Type == "json_object" {
		systemParts = append(systemParts, "IMPORTANT: Your response must be raw valid JSON matching the requested schema. Do not output markdown code blocks, backticks, or conversational preamble.")
	}

	systemPrompt := strings.Join(systemParts, "\n\n")

	payload := anthropicMessagesRequest{
		Model:    model,
		System:   systemPrompt,
		Messages: anthropicMsgs,
		Stream:   stream,
	}

	// 2. Handle Reasoning Effort / Thinking Budget for Claude 3.7+
	if effort != "" || req.ThinkingBudget > 0 {
		var budget int
		if req.ThinkingBudget > 0 {
			budget = req.ThinkingBudget
		} else {
			switch effort {
			case ReasoningEffortLow:
				budget = 1024
			case ReasoningEffortHigh:
				budget = 16384
			case ReasoningEffortMedium:
				fallthrough
			default:
				budget = 4096
			}
		}

		maxTokens := req.MaxTokens
		if maxTokens <= budget {
			maxTokens = budget + 2048
		}

		payload.Thinking = &anthropicThinking{
			Type:         "enabled",
			BudgetTokens: budget,
		}
		payload.MaxTokens = maxTokens
		// Anthropic requires temperature to be omitted or 1.0 when thinking is enabled
		payload.Temperature = nil
	} else {
		maxTokens := req.MaxTokens
		if maxTokens <= 0 {
			if hasSpec && modelSpec.DefaultMaxTokens > 0 {
				maxTokens = modelSpec.DefaultMaxTokens
			} else {
				maxTokens = 4096
			}
		}
		payload.MaxTokens = maxTokens
		payload.Temperature = req.Temperature
	}

	return payload
}

// Complete executes a non-streaming chat completion with an Anthropic-compatible endpoint.
func (p *AnthropicProvider) Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
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
		httpReq.Header.Set("x-api-key", p.apiKey)
	}
	httpReq.Header.Set("anthropic-version", p.anthropicVersion)

	resp, err := p.Execute(ctx, httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: read response body failed: %w", p.name, err)
	}

	var parsed anthropicMessagesResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("%s: unmarshal response failed: %w (body: %s)", p.name, err, string(respBody))
	}

	var textBuilder strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			textBuilder.WriteString(block.Text)
		}
	}

	return &ChatResponse{
		ID:           parsed.ID,
		Model:        parsed.Model,
		Role:         parsed.Role,
		Content:      textBuilder.String(),
		FinishReason: parsed.StopReason,
		Usage: Usage{
			PromptTokens:     parsed.Usage.InputTokens,
			CompletionTokens: parsed.Usage.OutputTokens,
			TotalTokens:      parsed.Usage.InputTokens + parsed.Usage.OutputTokens,
		},
	}, nil
}

// Stream executes a streaming chat completion, returning token and error channels.
func (p *AnthropicProvider) Stream(ctx context.Context, req ChatRequest) (<-chan string, <-chan error) {
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
		httpReq.Header.Set("x-api-key", p.apiKey)
	}
	httpReq.Header.Set("anthropic-version", p.anthropicVersion)

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

func (p *AnthropicProvider) readSSE(
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

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(dataPayload), &event); err != nil {
			continue
		}

		if event.Type == "message_stop" {
			return
		}

		if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" && event.Delta.Text != "" {
			select {
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			case tokenChan <- event.Delta.Text:
			}
		}
	}
}
