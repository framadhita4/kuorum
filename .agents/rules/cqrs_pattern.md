---
trigger: model_decision
description: This guide is **MANDATORY** for all AI Agents implementing features, use cases, and data flows within **`internal/application/`**.
---

# Rule: CQRS Pattern (Application Layer)

This guide is **MANDATORY** for all AI Agents implementing features, use cases, and data flows within **`internal/application/`**.

---

## 1. CQRS Philosophy & Location

CQRS **RESIDES EXCLUSIVELY WITHIN THE APPLICATION LAYER** (`internal/application/`). It is not fragmented across modules with independent layers; all application use cases are managed here.

```
internal/application/
├── common/                          # Shared CQRS Infrastructure & Primitives
│   ├── mediator/                    # In-memory Mediator / Dispatcher Engine
│   ├── behaviors/                   # Pipeline Behaviors (Validation, Logging)
│   └── interfaces/                  # Port Interfaces (ILLMClient, ICache, IRepository)
│
├── exploration/                     # Feature: Module 1
│   ├── commands/                    # MUTATIONS / ACTIONS / LLM PROCESSING
│   │   ├── decompose_concept/
│   │   │   ├── command.go           # Input Command Data Struct
│   │   │   ├── validator.go         # Input Validation Logic
│   │   │   ├── handler.go           # Business Logic & Orchestration
│   │   │   └── result.go            # Output Result DTO
│   │   └── reformulate_query/
│   │       ├── command.go
│   │       ├── validator.go
│   │       ├── handler.go
│   │       └── result.go
│   └── queries/                     # READ-ONLY OPERATIONS
│       └── get_angles/
│           ├── query.go             # Query Request Struct
│           ├── handler.go           # Query Execution Logic
│           └── result.go            # Query Output DTO
│
├── retrieval/                       # Feature: Module 2 (Commands & Queries)
├── graph/                           # Feature: Module 3 (Commands & Queries)
└── extraction/                      # Feature: Module 4 (Commands & Queries)
```

---

## 2. Strict Separation: Command vs Query

### A. Commands (`commands/`)
- Used for operations that:
  - Mutate application or persistence state.
  - Invoke cognitive LLM processing (decomposition, extraction, query reformulation).
  - Perform file ingestion (PDF parsing) or graph modifications.
- Slice structure per command:
  - `command.go`: Input parameters struct.
  - `validator.go`: Validation rules (similar to FluentValidation in .NET).
  - `handler.go`: Executes logic, interacts with ports/repositories, and returns `result.go`.
  - `result.go`: Output DTO.

### B. Queries (`queries/`)
- Used for operations that are **READ-ONLY and IDEMPOTENT**:
  - Fetching records from cache or SQLite.
  - Reading the state of an already-traversed citation graph.
  - Retrieving pre-generated concept angles.
- **STRICT RULE:** Queries **MUST NEVER** alter application state or trigger expensive/paid external LLM calls.

---

## 3. CQRS Core Interfaces (`internal/application/common/mediator/`)

```go
package mediator

import "context"

type Command[TResult any] interface {
    IsCommand()
}

type Query[TResult any] interface {
    IsQuery()
}

type CommandHandler[TCommand Command[TResult], TResult any] interface {
    Handle(ctx context.Context, cmd TCommand) (TResult, error)
}

type QueryHandler[TQuery Query[TResult], TResult any] interface {
    Handle(ctx context.Context, q TQuery) (TResult, error)
}

type Mediator interface {
    Send(ctx context.Context, cmd any) (any, error)
    Query(ctx context.Context, query any) (any, error)
}
```

---

## 4. Pipeline Behaviors (Pre-Execution Interceptors)

Every dispatch via `mediator.Send()` or `mediator.Query()` traverses the pipeline behavior chain:

1. **Validation Behavior:**
   - Checks if the Command/Query struct implements `Validate() error`.
   - If invalid, returns a `ValidationError` immediately without invoking the handler.
2. **Logging Behavior:**
   - Logs start timestamp, end timestamp, latency (ms), and any errors returned during execution.

---

## 5. Handler Implementation Guidelines
1. **Single Responsibility:** Each handler file processes exactly one Command or one Query.
2. **Injected Dependencies:** Handlers accept only interfaces (`ports`), never direct concrete clients.
3. **Domain & Port Purity:** Handlers coordinate domain rules and ports; they know nothing of HTTP controllers or transport protocols.
