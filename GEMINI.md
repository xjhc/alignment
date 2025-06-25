# CLAUDE.md

This file provides guidance to Gemini when working with code in this repository.

## Project Overview

**Alignment** is a corporate-themed social deduction game where humans identify a rogue AI among them before it converts staff and seizes company control. This is a **Go monorepo** in active development. The core architecture is implemented, and work is ongoing on the detailed game logic.

## Architecture & Technology Stack

-   **Backend**: Go with a supervised Actor Model architecture.
-   **Frontend**: Hybrid Go/WebAssembly + React/TypeScript application.
-   **Database**: Redis used as a Write-Ahead Log (WAL) and for state snapshots.
-   **Build**: Vite for frontend, standard Go toolchain for backend.
-   **Deployment**: Single VM deployment model with Docker Compose.

## Development Commands

### Primary Workflow (using `Makefile`)

The project is managed via a `Makefile` at the root. This is the preferred way to run all common tasks. To see the full list of commands and their descriptions, run `make help`.

**1. One-Time Setup**
```bash
# Install all Node.js dependencies
npm install

# Create the Go vendor directory for the backend
make vendor

# Ensure Redis is running in a separate terminal
redis-server &
```

**2. Daily Development**
```bash
# Run both servers with hot-reloading for interactive development
make dev
```

**3. Testing**
```bash
# Run all backend and frontend tests
make test
```

**4. Background Services (for E2E tests)**
```bash
# Start services in the background
make bg-start

# Stop and clean up services
make bg-stop
```

### Individual Commands

#### Backend (Go)

```bash
# Navigate to the server directory
cd server/

# Start the backend server (requires Redis to be running)
go run ./cmd/server/

# Run tests with race detection
go test -race ./...

# Linting and formatting
go fmt ./...
golangci-lint run

# Generate a coverage report
go tool cover -html=coverage.out
```

#### Frontend (React + Go/Wasm)

```bash
# Navigate to the client directory
cd client/

# Install dependencies
npm install

# Start the frontend dev server
npm run dev

# Build for production
npm run build
```

### Dependencies

```bash
# Start Redis (required for the backend)
redis-server
```

## Development guidelines

This codebase prioritizes **maintainability and performance** through clean, idiomatic Go and modern frontend practices. When generating or modifying code, adhere strictly to these principles:

-   Simplicity and Conciseness: Write the most straightforward code possible. Avoid overly clever or "magic" solutions. Code should be dense with meaning, not with characters.
-   Single Responsibility Principle (SRP): Every function, struct, and package should have one, and only one, reason to change. Decompose complex logic into smaller, focused units.
-   Don't Repeat Yourself (DRY): Aggressively refactor to eliminate duplicated code. Use functions and shared modules to promote reuse.
-   Testability: Code must be structured to be easily testable. This often means preferring pure functions and using interfaces for dependencies.
-   No Technical Debt: We adhere to the "Boy Scout Rule"—always leave the code cleaner than you found it. Do not defer refactoring or implement temporary hacks. Choose the correct, maintainable solution now, even if it takes longer.
-   Documentation as Code: Documentation must be kept up-to-date. If a code change alters a feature, API, or architectural pattern, the corresponding documentation in the `/docs` directory must be updated within the same commit or pull request. Treat documentation with the same rigor as source code.

## Frontend Styling Guidelines

This project uses **Tailwind CSS** as the primary styling framework, integrated with our design token system to ensure consistency and maintain our design system constraints.

### Styling Approach

-   **Tailwind Utility Classes**: Use Tailwind utility classes directly in component JSX. Avoid writing custom CSS unless absolutely necessary.
-   **Design Token Integration**: Tailwind is configured to use our design tokens as the single source of truth. Colors, spacing, typography, and other design elements are automatically available as Tailwind classes.
-   **Theme Support**: For dynamic theming (light/dark mode), use CSS custom properties through our `background-*`, `text-*`, and `border` color classes.
-   **Component Composition**: Build complex layouts by composing Tailwind utility classes rather than creating custom CSS classes.

### Tailwind Configuration

The `tailwind.config.js` file is configured to:
-   Import design tokens from `src/styles/generated/tokens.js`
-   Map token values to Tailwind's theme configuration
-   Include custom animation utilities that match our existing animations
-   Support CSS custom properties for theme switching

### Available Design Token Classes

Our design tokens are available as Tailwind classes:

**Colors**: `bg-primary`, `text-primary`, `border-primary`, `bg-background-primary`, `text-text-primary`, etc.
**Spacing**: `p-4`, `m-2`, `gap-3` (based on our spacing scale)
**Typography**: `text-xs`, `text-sm`, `font-medium`, `font-mono`
**Border Radius**: `rounded`, `rounded-md`, `rounded-lg`
**Custom Colors**: `bg-human`, `bg-aligned`, `bg-ai`, `bg-success`, `bg-danger`

### Development Workflow

1. **Use Tailwind First**: Always attempt to solve styling needs with Tailwind utility classes
2. **Reference Design Tokens**: Prefer classes that map to our design tokens over arbitrary values
3. **Component Patterns**: For reusable patterns, create TypeScript utility functions that return class strings
4. **Avoid CSS Modules**: Do not create new `.module.css` files; use Tailwind classes instead

### Motion System

This project implements a formal **Motion System** for consistent, purposeful animations. All animations follow the principles defined in `docs/development/05-motion-system.md`.

**Animation Guidelines:**
-   **Use Motion System Classes**: Always use the standardized `animation-*` classes (e.g., `animation-fade-in`, `animation-slide-in-up`) instead of ad-hoc animations
-   **Import from Utils**: Use animation constants from `src/utils/animations.ts` to avoid magic strings (e.g., `FADE_IN`, `SLIDE_IN_UP`)
-   **Respect Timing**: Use the defined duration tokens (`--duration-fast`, `--duration-medium`, `--duration-slow`) and easing curves (`--ease-out`, `--ease-in`, `--ease-feedback`)
-   **Staggered Animations**: Use the `applyStaggeredAnimation()` helper function for sequential list reveals

**Available Animation Classes:**
- **Fade**: `animation-fade-in`, `animation-fade-out`
- **Slide**: `animation-slide-in-up`, `animation-slide-in-down`, `animation-slide-in-left`, `animation-slide-in-right`
- **Scale**: `animation-scale-in`, `animation-scale-in-feedback`
- **Stagger**: `stagger-child` for sequential list animations
- **Game-specific**: `animation-glitch`, `animation-pulse`, `animation-shake`, `animation-elimination-fade`

### Migration Status

The project is currently migrating from CSS Modules to Tailwind CSS. As components are refactored:
-   Remove CSS Module imports and corresponding `.module.css` files
-   Replace CSS Module classes with equivalent Tailwind utility classes
-   Replace old `animate-*` classes with new Motion System `animation-*` classes
-   Maintain visual consistency during the transition

## Key Architectural Patterns

-   **Actor Model:** Each game runs in a dedicated goroutine (`GameActor`) with its own in-memory `GameState`. All actions are processed serially through a Go channel, eliminating locks.
-   **Supervisor Pattern:** A top-level `Supervisor` launches and monitors all `GameActors`, providing fault isolation so a single game crash doesn't take down the server.
-   **Event Sourcing with WAL:** The server is stateful in memory for speed, but all state-changing `Events` are first persisted to a Redis Stream (the Write-Ahead Log) for durability and recovery.
-   **Shared `core` Package:** Critical game types and the pure `ApplyEvent` function are defined in a shared `/core` package, compiled for both the Go backend and the Go/Wasm frontend to guarantee rule consistency.
-   **Hybrid AI Brain:** The AI opponent is split into a deterministic Go `RulesEngine` for strategic game actions and an `LLM` for generating human-like chat.

## Testing Strategy

### Coverage Requirements

-   **Core Logic (`/core`)**: Target 95%+ unit test coverage for `ApplyEvent` and rule functions.
-   **Server Logic (`/server`)**: Target 80%+ integration test coverage. The CI pipeline enforces this.

### Test Patterns

```go
// Table-driven tests for pure functions in /core
func TestApplyEvent(t *testing.T) {
    testCases := []struct {
        name          string
        initialState  core.GameState
        event         core.Event
        expectedState core.GameState
    }{
        // Test cases here
    }
    // ... loop and run tests ...
}

// Actor integration tests in /server with mocked dependencies
func TestActor_PlayerJoinsAndVotes(t *testing.T) {
    mockStore := &MockDataStore{}
    mockBroadcaster := &MockBroadcaster{}
    actor := NewGameActor("test-game", mockStore, mockBroadcaster)
    // Send actions to actor.mailbox and assert on mock calls
}
```

## Project Structure

```
.
├── core/               # SHARED Go logic (types, ApplyEvent)
├── server/             # Go backend
│   ├── cmd/server/     # Main server binary
│   └── internal/
│       ├── actors/     # Game Actor and Supervisor
│       ├── ai/         # Rules engine and LLM integration
│       ├── comms/      # WebSocket communication
│       ├── game/       # Server-side game logic managers
│       └── store/      # Redis persistence logic
├── client/             # React/TypeScript frontend
│   ├── src/            # React components and UI logic
│   └── wasm/           # (Future) Go/Wasm game engine source
└── docs/               # Comprehensive design documentation
```

## Important Documentation

-   **`docs/01-game-design-document.md`**: Complete game rules and mechanics.
-   **`docs/02-onboarding-for-engineers.md`**: Essential 5-minute technical overview.
-   **`docs/development/03-code-logic-boundaries.md`**: The strict rules for what code belongs in `/core`, `/server`, and `/client`. **(Must Read)**
-   **`docs/architecture/README.md`**: Detailed backend architecture explanations.
-   **`docs/adr/README.md`**: Architectural decisions with context and rationale.

## Current Status

This repository is in **active development**. The project has moved past the pure design phase into implementation.

-   **DONE**:
    -   The foundational architecture is implemented. This includes the Go server, the Supervisor/Actor model, WebSocket communication, the Redis WAL store, and the shared `/core` package structure.
    -   Development tooling has been streamlined with `concurrently` for unified dev server management.
    -   React/TypeScript frontend foundation is established with Vite build system.

-   **IN PROGRESS**: The detailed game logic is being built out.
    -   The `core.ApplyEvent` function has a complete structure, but many individual `apply...` handlers are stubs.
    -   The `server.GameActor` is implemented but does not yet fully delegate logic to the specialized managers in `/server/internal/game`.
    -   The AI `RulesEngine` is a placeholder and needs its strategic heuristics implemented.
    -   Frontend React components and game UI are being developed alongside the backend logic.