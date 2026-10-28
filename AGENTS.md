# AI Agent Directives & Architecture Rules Index

This document is the **Master Router & Directive Index** for the **Research Helper** project.
To optimize context efficiency, **AI Agents DO NOT need to read all rules**. Instead, agents **MUST read the specific rule file mapped below based on the task being performed**:

---

## Directive Routing by Task Domain

| Task Scope | Mandatory Rule File to Read | Brief Description |
| :--- | :--- | :--- |
| **Frontend Development (React, Vite, TanStack)** | [`.agents/rules/frontend_architecture.md`](file:///.agents/rules/frontend_architecture.md) | `@emotion/react` styling engine, 1 main theme file, component visual vs layout separation, `-components/`, `-schema/`, path alias `#/`, TanStack Router & Query. |
| **Backend Core Architecture (Golang)** | [`.agents/rules/backend_clean_architecture.md`](file:///.agents/rules/backend_clean_architecture.md) | 4-Layer Clean Architecture (`domain`, `application`, `infrastructure`, `api`), inward dependency flow, Go idioms. |
| **Business Logic, Commands & Queries (CQRS)** | [`.agents/rules/cqrs_pattern.md`](file:///.agents/rules/cqrs_pattern.md) | CQRS centralized in Application layer (.NET style), Command vs Query separation, Handlers, Validators, Mediator & Pipelines. |
| **HTTP API, Controllers & SSE Streaming** | [`.agents/rules/api_and_streaming.md`](file:///.agents/rules/api_and_streaming.md) | `go-chi` routing, Thin Controllers, Envelope Response (`success`, `data`, `error`), JSON `camelCase`, and SSE protocol. |
| **LLM Integration & Custom HTTP Client** | [`.agents/rules/llm_infrastructure.md`](file:///.agents/rules/llm_infrastructure.md) | Custom `net/http.Client` (no third-party SDKs), OpenAI-compatible API, line-by-line SSE parsing, Structured Outputs. |

---

## Strict Directives for All AI Agents
1. **Always Read the Relevant Rule First**: Before writing or modifying any file, check the table above and inspect the corresponding rule file.
2. **Zero Architecture Violation**: The 4-Layer Clean Architecture and CQRS pattern are enterprise requirements and must never be bypassed (e.g., never put business logic in controllers, never import infrastructure packages from the domain layer).
3. **No Shortcuts**: Do not bypass mediator pipelines, launch unbounded goroutines, or use `any` types in TypeScript.
4. **Styling & Emotion Rules**: All UI styling is governed by `@emotion/react` with `#/-components/ui/theme.ts` as the single source of truth. Component internal styling (colors, font-size, interactive states) MUST be controlled by the component; layout (flex, grid, position, etc.) is adjusted directly by the parent container.
