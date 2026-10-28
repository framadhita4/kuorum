package llm

import (
	"kuorum/backend/internal/infrastructure/llm/providers"
)

// ChatCompletionChoice provides OpenAI wire compatibility.
type ChatCompletionChoice struct {
	Index        int                   `json:"index"`
	Message      providers.ChatMessage `json:"message"`
	Delta        providers.ChatMessage `json:"delta"`
	FinishReason string                `json:"finish_reason"`
}

// ChatCompletionResponse provides OpenAI wire compatibility.
type ChatCompletionResponse struct {
	ID      string                 `json:"id"`
	Choices []ChatCompletionChoice `json:"choices"`
}
