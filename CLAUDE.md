# Claude Development Guide

This document provides guidance for AI assistants working on the `Alignment` codebase.

## 1. Core Principles

Adherence to these principles is mandatory.

- **Refactor Continuously:** If a file grows too large or complex, break it into smaller, single-responsibility modules. The goal is a clean, modular system where each component is easy to understand and test in isolation.
- **Concise Code:** Keep code short and to the point. Add documentation only where the logic is complex or non-obvious. Clean, self-explanatory code is preferred over heavily commented code.
- **Single Source of Truth (DRY):** Aggressively refactor to eliminate duplication. The shared `/core` package is the definitive source for all game logic and types.
- **Documentation as Code:** If a code change alters a feature, API, or architectural pattern, the corresponding documentation in the `/docs` directory **must** be updated in the same commit.
- **No Technical Debt:** Adhere to the "Boy Scout Rule"—always leave the code cleaner than you found it. Implement the correct, maintainable solution now.

## 2. Development Workflow: Use the Makefile

The project is managed via a `Makefile` at the root. This is the **single source of truth for all common tasks.** Use it for all development, building, and testing.

To see the full list of commands and their descriptions, run `make help`.

**Primary Commands:**

```bash
# Install all Go and npm dependencies (run once)
make install

# Run backend and frontend servers with hot-reloading
make dev

# Run all backend and frontend tests
make test

# Build all production artifacts
make build

# Lint all code
make lint

# Clean all build artifacts
make clean
```

## 3. Key Architectural Patterns

This codebase is built on a specific set of architectural patterns. You must understand and adhere to them.

- **Player-Centric Actor Model:** The backend is a stateful, in-memory system. Each player's WebSocket connection is managed by a dedicated `PlayerActor` goroutine.
- **Supervised Game Actors:** Each game simulation runs in its own isolated `GameActor`, which is monitored by a `Supervisor` to contain crashes.
- **Event Sourcing with Redis WAL:** The `GameActor` holds state in memory for speed. All state-changing events are first persisted to a Redis Stream (Write-Ahead Log) for durability and fast recovery. Redis is **not** read from during normal gameplay.
- **Shared `/core` Package:** All fundamental game logic (`ApplyEvent` function) and data structures (`GameState`, `Player`, etc.) are defined in the `/core` package. This package is compiled for both the Go backend and the Go/Wasm frontend to guarantee rule consistency.
- **Single Authoritative Event (ADR-006):** For any given player action, the server must generate **one, and only one,** event that fully describes the resulting state change. This is a critical pattern to prevent race conditions.

## 4. Testing Strategy

We use a multi-layered testing pyramid. When adding code, you are expected to add corresponding tests at the appropriate level.

1.  **Unit Tests (`/core`):** Pure functions in the `core` package must have near-100% test coverage using table-driven tests.
2.  **Integration Tests (`/server`):** Actors are tested as black boxes with mocked dependencies. Use the `testify/mock` library and the `WaitGroup` pattern for synchronizing asynchronous tests.
3.  **End-to-End Tests (`/tests/e2e`):** The `pytest` suite validates full-stack user flows against a live, containerized application.

Run all tests via `make test`.

## 5. Project Structure

```
.
├── core/               # SHARED Go logic (types, ApplyEvent)
├── server/             # Go backend
│   ├── cmd/server/     # Main server binary
│   └── internal/
│       ├── actors/
│       ├── app/        # Server setup and initialization
│       └── ...
├── client/             # React/TypeScript frontend
│   ├── src/            # React components and UI logic
│   └── wasm/           # Go/Wasm game engine source
└── docs/               # All project documentation
```

## 6. Important Documentation

Before writing code, consult these key documents:

- **`docs/01-game-design-document.md`**: The complete game rules.
- **`docs/architecture/README.md`**: Detailed backend architecture.
- **`docs/adr/README.md`**: The "why" behind our key technical choices.
- **`docs/development/02-testing-strategy.md`**: Our testing patterns and philosophy.
