---
trigger: model_decision
description: This guide is **MANDATORY** for all AI Agents implementing or modifying code in the **Backend (Golang)**. The backend strictly follows enterprise **4-Layer Clean Architecture**.
---

# Rule: Backend Clean Architecture (Golang)

This guide is **MANDATORY** for all AI Agents implementing or modifying code in the **Backend (Golang)**. The backend strictly follows enterprise **4-Layer Clean Architecture**.

---

## 1. Four Core Layers

All backend architecture lives within `internal/` with strict layer boundaries:

```
backend/
├── cmd/
│   └── api/
│       └── main.go                  # Bootstrap, Dependency Injection & Server Startup
└── internal/
    ├── domain/                      # LAYER 1: DOMAIN (Enterprise Core)
    ├── application/                 # LAYER 2: APPLICATION (Use Cases & CQRS Engine)
    ├── infrastructure/              # LAYER 3: INFRASTRUCTURE (External Drivers & Adapters)
    └── api/                         # LAYER 4: API / PRESENTATION (HTTP Server & Handlers)
```

---

## 2. Inward Dependency Rule

Dependencies must **strictly point inward**.

```
[ Layer 4: API ] ──────────┐
                           ▼
                 [ Layer 2: Application ] ──► [ Layer 1: Domain ]
                           ▲
[ Layer 3: Infrastructure ] ┘
```

1. **Layer 1: Domain (`internal/domain/`)**
   - **MUST NEVER DEPEND ON ANY OTHER LAYER**. Never import packages from `application`, `infrastructure`, or `api`.
   - Contains only: Entities, Value Objects, Domain Events, Domain Errors, and pure business domain invariants.
   - Strictly no dependencies on external database drivers, HTTP clients, or web frameworks.

2. **Layer 2: Application (`internal/application/`)**
   - Business orchestration hub and the **sole home of the CQRS engine (Commands & Queries)** across all features.
   - Only depends on **Layer 1 (Domain)**.
   - Defines ports/interfaces for external dependencies (e.g., `ILLMClient`, `ICacheRepository`, `IPaperApi`).

3. **Layer 3: Infrastructure (`internal/infrastructure/`)**
   - Implements ports/interfaces defined by the Application Layer.
   - Contains technical implementations: Custom HTTP LLM Client, SQLite persistence, Academic API Clients (Semantic Scholar, OpenAlex), loggers, and config readers.
   - Depends on **Application** and **Domain**.

4. **Layer 4: API (`internal/api/`)**
   - Presentation layer built on `go-chi`.
   - Converts HTTP requests into Commands/Queries, dispatches them through the Mediator, and renders HTTP response envelopes.
   - Depends on **Application** and **Domain**.

---

## 3. Golang Conventions & Quality Standards

### A. Context Propagation
- **EVERY** I/O, database query, HTTP call, and Command/Query handler **MUST** accept `ctx context.Context` as its first parameter:
  ```go
  func (h *DecomposeCommandHandler) Handle(ctx context.Context, cmd DecomposeCommand) (*Result, error)
  ```

### B. Error Handling & Wrapping
- **Always wrap errors** with explicit context using `%w`:
  ```go
  if err != nil {
      return nil, fmt.Errorf("decompose concept failed: %w", err)
  }
  ```
- **STRICTLY FORBIDDEN** to use `panic()` within business flows or handlers. Panics are only permitted during startup initialization in `main.go`.

### C. Dependency Injection & State
- **STRICTLY NO GLOBAL STATE / SINGLETONS**:
  - Never declare package-level mutable variables (`var db *sql.DB`, `var client *HttpClient`).
  - All dependencies **must** be injected via constructors:
    ```go
    func NewExplorationController(mediator mediator.Mediator) *ExplorationController {
        return &ExplorationController{mediator: mediator}
    }
    ```

### D. Concurrency & Goroutines
- **No naked/unmanaged goroutines (`go func()`)**:
  - All asynchronous or parallel tasks must be orchestrated using `golang.org/x/sync/errgroup` or tied to `ctx.Done()` to prevent leaks upon client disconnect.
