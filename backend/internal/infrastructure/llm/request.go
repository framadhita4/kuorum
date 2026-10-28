package llm

import (
	"kuorum/backend/internal/infrastructure/llm/providers"
)

// ChatCompletionRequest provides OpenAI wire compatibility.
type ChatCompletionRequest struct {
	Model          string                    `json:"model"`
	Messages       []providers.ChatMessage   `json:"messages"`
	ResponseFormat *providers.ResponseFormat `json:"response_format,omitempty"`
	Stream         bool                      `json:"stream,omitempty"`
	Temperature    float32                   `json:"temperature,omitempty"`
}
