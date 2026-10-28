---
trigger: model_decision
description: This guide is **MANDATORY** for all AI Agents implementing or modifying the LLM integration layer in **`internal/infrastructure/llm/`**.
---

# Rule: LLM Client & Infrastructure (Custom HTTP Client)

This guide is **MANDATORY** for all AI Agents implementing or modifying the LLM integration layer in **`internal/infrastructure/llm/`**.

---

## 1. Technology Philosophy & Constraints
- **STRICTLY FORBIDDEN to use third-party SDKs** (such as `go-openai` or `langchaingo`).
- All communication with LLMs **MUST use a pure Custom HTTP Client built on Go's `net/http.Client`**.
- This guarantees total control over connection management, SSE stream parsing, custom headers, and seamless compatibility across all **OpenAI-Compatible APIs** (OpenAI, Ollama, OpenRouter, vLLM, Groq).

---

## 2. LLM Package Structure

```
internal/infrastructure/llm/
├── client.go        # Custom HTTP Client (Timeout, Transport, BaseURL, APIKey)
├── request.go       # OpenAI-Compatible payload structs (Model, Messages, ResponseFormat, Stream)
├── response.go      # OpenAI-Compatible response structs
├── sse_scanner.go   # Line-by-line SSE parser (parses 'data: {...}' and '[DONE]')
└── service.go       # Implementation of Application Layer ILLMClient port
```

---

## 3. Port Interface (Application Layer)

The interface is declared in `internal/application/common/interfaces/llm.go`:

```go
package interfaces

import "context"

type ILLMClient interface {
    // CompleteStructured requests the LLM to return structured JSON adhering to a specified schema
    CompleteStructured(ctx context.Context, systemPrompt, userPrompt string, target any) error

    // Stream streams tokens in real-time through a Go channel
    Stream(ctx context.Context, systemPrompt, userPrompt string) (<-chan string, <-chan error)
}
```

---

## 4. Custom HTTP Client Implementation Rules

1. **Connection Pooling & Timeouts:**
   - The client must configure an `http.Transport` optimized for long-lived streaming connections:
     ```go
     transport := &http.Transport{
         MaxIdleConns:        100,
         MaxIdleConnsPerHost: 20,
         IdleConnTimeout:     90 * time.Second,
     }
     ```
2. **Context Propagation & Early Cancellation:**
   - Every outgoing HTTP request **MUST be constructed with `http.NewRequestWithContext(ctx, ...)`**.
   - If the user disconnects or cancels from the frontend, the outgoing HTTP connection to the LLM provider must terminate immediately to conserve compute and tokens.
3. **SSE Line Parsing:**
   - Use `bufio.Scanner` or `bufio.Reader` to read responses line by line.
   - Detect the `data: ` prefix cleanly.
   - When encountering `data: [DONE]`, gracefully close the streaming channel.
4. **Structured Output Enforcement:**
   - For non-streaming completions, pass JSON schema / response format parameters:
     ```json
     "response_format": {
       "type": "json_object"
     }
     ```
   - Unmarshal the response payload directly into the `target any` struct supplied by the Application Layer.
