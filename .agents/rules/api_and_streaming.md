---
trigger: model_decision
description: This guide is **MANDATORY** for all AI Agents implementing or modifying code in **`internal/api/`** (Presentation Layer).
---

# Rule: API Presentation, Response Envelope & Streaming (go-chi)

This guide is **MANDATORY** for all AI Agents implementing or modifying code in **`internal/api/`** (Presentation Layer).

---

## 1. HTTP Controller Responsibility (Thin Controller)

Controllers in `internal/api/controllers/` **MUST NEVER** contain domain business logic. Their responsibilities are strictly:
1. Decode and syntactically validate HTTP Request bodies/query parameters into a Command or Query struct.
2. Dispatch to `mediator.Send(ctx, cmd)` or `mediator.Query(ctx, query)`.
3. Return the response using standard helpers `response.Success` or `response.Error`.

Example controller implementation:
```go
func (c *ExplorationController) Decompose(w http.ResponseWriter, r *http.Request) {
    var cmd decompose_concept.DecomposeConceptCommand
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        response.Error(w, http.StatusBadRequest, "Invalid JSON payload", err.Error())
        return
    }

    result, err := c.mediator.Send(r.Context(), cmd)
    if err != nil {
        response.FromAppError(w, err)
        return
    }

    response.Success(w, http.StatusOK, "Concept decomposed successfully", result)
}
```

---

## 2. Standard Response Envelope (Non-Streaming)

All non-streaming REST responses **MUST** adhere to the uniform envelope format:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": { ... },
  "error": null
}
```

When an error occurs:
```json
{
  "success": false,
  "message": "User-friendly error explanation",
  "data": null,
  "error": "Detailed technical error message"
}
```

### JSON Tag Conventions:
- All Go data structs serialized to JSON **MUST use `camelCase` tags**:
  ```go
  type ExplorationAngleDTO struct {
      AngleID          string   `json:"angleId"`
      Title            string   `json:"title"`
      Description      string   `json:"description"`
      SuggestedQueries []string `json:"suggestedQueries"`
  }
  ```

---

## 3. Server-Sent Events (SSE) Streaming Protocol

For operations requiring real-time token streaming (such as LLM decomposition or citation graph traversal updates):

1. **Mandatory SSE Headers:**
   ```http
   Content-Type: text/event-stream
   Cache-Control: no-cache
   Connection: keep-alive
   X-Accel-Buffering: no
   ```
2. **Immediate Flushing:**
   Handlers must cast the writer to `http.Flusher` to ensure tokens are dispatched immediately without buffer delays:
   ```go
   flusher, ok := w.(http.Flusher)
   if !ok {
       http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
       return
   }
   flusher.Flush()
   ```
3. **Standardized SSE Events:**
   - `event: status` $\rightarrow$ Lifecycle stage update (e.g., `{"status": "decomposing_facets"}`).
   - `event: token` $\rightarrow$ Real-time text token from the LLM (e.g., `{"token": " Deep"}`).
   - `event: error` $\rightarrow$ Error payload if the stream terminates prematurely.
   - `event: done` $\rightarrow$ Final completion signal with the complete structured JSON payload.

---

## 4. SPA Fallback for TanStack Router
- The `go-chi` router mounts `/api/v1/*` for backend endpoints.
- All non-API routes (e.g., `/explore`, `/graph/*`) **MUST** serve `dist/index.html` (SPA Fallback) to allow TanStack Router in the browser to handle client-side routing seamlessly.
