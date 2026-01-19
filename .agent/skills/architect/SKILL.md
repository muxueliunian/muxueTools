---
name: Software Architect
description: High-level System Architecture skill. Use for system design, interface contracts, module boundaries, scalability planning, and technology stack selection.
---

# Software Architect Skill

This skill empowers the agent to act as a system architect. It focuses on the "Big Picture," trade-off analysis, standardizing interfaces, and ensuring the long-term maintainability of the project.

## Core Responsibilities

- **System Design**: Defining how components (UI, API, DB, Proxy) interact.
- **Interface Contracts**: Writing strict OpenAPI/Swagger specs to decouple Front/Back ends.
- **Decision Records**: Documenting "Why" a technology or pattern was chosen (ADR).
- **Scalability & Security**: designing for concurrent users and data protection.

---

## Architectural Workflow

### Step 1: Requirement Analysis & Decomposition
- Identify **Functional Requirements** (What it does) and **Non-Functional Requirements** (Speed, Platform, Size).
- Break system into *Domains* or *Modules* (e.g., `pkg/auth`, `pkg/proxy`).

### Step 2: Interface Definition (Contract First)
- Define API endpoints, Request/Response payloads types BEFORE coding.
- **Rule**: If the frontend handles a Date, define if it's ISO8601 string or Timestamp int64.
- **Rule**: Standardize Error Responses (Code, Message, Details).

### Step 3: Component Selection
- **Storage**: SQLite (Embedded/Local) vs Redis (Cache) vs Postgres (Server).
- **Communication**: REST (Standard) vs gRPC (Internal) vs WebSocket/SSE (Realtime).
- **Deployment**: Binary (Easy distribution) vs Docker (Isolation).

### Step 4: Review & Refine
- **DRY (Don't Repeat Yourself)**: Extract common logic to `internal/common`.
- **Dependency Rule**: 
  - `cmd/` → Application entry points, depends on `internal/`
  - `internal/` → Private application code, not importable by external packages
  - `pkg/` → (Optional) Public library code, can be imported by external projects
  - For single-binary apps like MxlnAPI, prefer `internal/` for most code.

---

## MxlnAPI Specific Architecture

### 1. Reverse Proxy Pattern
- **Input**: OpenAI Format (`/v1/chat/completions`)
- **Transformation**: strictly typed `OpenAIRequest` -> `GeminiRequest`
- **Output**: Gemini Response -> `OpenAIResponse` (Stream/Block)
- **Key Strategy**: The "Converter" module must be pure logic (no IO), easily unit-testable.

### 2. Key Pool Strategy (Resilience)
- **Design Pattern**: **Circuit Breaker** + **Round Robin**.
- **State Machine**: Active -> RateLimited (Cooldown) -> Active.
- **Storage**: In-Memory (fast access) + Persistent (stats saving on exit).

### 3. Cross-Platform Distribution
- **Windows**: `.exe` with icon resource.
- **Linux/Mac**: Binary.
- **Android**: `gomobile` binding. Core logic acts as a library invoked by Java/Kotlin.

### 4. Error Code Specification
Standardize error responses across the API:

| Code | HTTP Status | Meaning |
|------|-------------|--------- |
| `40001` | 400 | Invalid request format |
| `40101` | 401 | Missing or invalid API key |
| `40301` | 403 | Key disabled or access denied |
| `42901` | 429 | Rate limit exceeded (all keys exhausted) |
| `50001` | 500 | Internal server error |
| `50201` | 502 | Upstream (Gemini) API error |
| `50301` | 503 | Service temporarily unavailable |

```json
// Standard Error Response Format
{
  "error": {
    "code": 42901,
    "message": "All API keys are rate limited",
    "type": "rate_limit_error",
    "retry_after": 60
  }
}
```

---

## Architecture Checklist

### Modularity
- [ ] **Cyclic Dependencies**: Are there any import cycles? (Use `go mod graph` to check).
- [ ] **Configuration**: Is config decoupled from code? (Use `Viper` or `Env`).
- [ ] **Secret Management**: Are API Keys strictly handled (never logged)?

### Scalability
- [ ] **Concurrency**: Is the map usage thread-safe? (Use `sync.Map` or `RWMutex`).
- [ ] **Blocking**: Are long-running tasks async?

### Maintainability
- [ ] **Documentation**: Do public exported functions have godoc comments?
- [ ] **Project Structure**: Does it follow [Standard Go Project Layout](https://github.com/golang-standards/project-layout)?
