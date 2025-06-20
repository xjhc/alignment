# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Alignment** is a corporate-themed social deduction game where humans identify a rogue AI among them before it converts staff and seizes company control. This is a **Go monorepo** in the design phase - source code is not yet implemented, only comprehensive documentation exists.

## Architecture & Technology Stack

- **Backend**: Go with supervised Actor Model architecture
- **Frontend**: Hybrid Go/WebAssembly + React/TypeScript application
- **Database**: Redis (Write-Ahead Log and state snapshots)
- **Build**: Vite for frontend, standard Go toolchain for backend
- **Deployment**: Single VM deployment model

## Development Commands

**Note**: The actual source code has not been implemented yet. These are the planned commands once implementation begins:

### Backend (Go)
```bash
# Start backend server (when implemented)
cd server/
go run ./cmd/server/

# Run tests with race detection
go test -race ./...

# Linting and formatting
go fmt ./...
go vet ./...
golangci-lint run

# Generate coverage report
go tool cover -html=coverage.out
```

### Frontend (React + Go/Wasm)
```bash
# Start frontend dev server (when implemented)
cd client/
npm install
npm run dev

# Build for production
npm run build
```

### Dependencies
```bash
# Start Redis (required for backend)
redis-server
```


## Development guidelines

This codebase prioritizes **maintainability and performance** through clean, idiomatic Go and modern frontend practices. When generating or modifying code, adhere strictly to these principles:

* Simplicity and Conciseness: Write the most straightforward code possible. Avoid overly clever or "magic" solutions. Code should be dense with meaning, not with characters.
* Single Responsibility Principle (SRP): Every function, struct, and package should have one, and only one, reason to change. Decompose complex logic into smaller, focused units.
* Don't Repeat Yourself (DRY): Aggressively refactor to eliminate duplicated code. Use functions and shared modules to promote reuse.
* Testability: Code must be structured to be easily testable. This often means preferring pure functions and using interfaces for dependencies.
* No Technical Debt: We adhere to the "Boy Scout Rule"—always leave the code cleaner than you found it. Do not defer refactoring or implement temporary hacks. Choose the correct, maintainable solution now, even if it takes longer.
* Documentation as Code: Documentation must be kept up-to-date. If a code change alters a feature, API, or architectural pattern, the corresponding documentation in the /docs directory must be updated within the same commit or pull request. Treat documentation with the same rigor as source code.

## Key Architectural Patterns

### Actor Model Implementation
- Each game runs in dedicated goroutine with in-memory `GameState`
- All game events processed serially through Go channels (no locks needed)
- Supervisor pattern provides fault isolation - single game crashes don't affect server
- Central Dispatcher routes WebSocket messages to correct actor channels

### State Management
- **In-Memory**: Live games hold complete state in memory for performance
- **Redis WAL**: All events appended to Redis Streams for durability
- **Snapshots**: Periodic full state saves to Redis keys
- **Recovery**: Fast startup by loading snapshot + replaying recent events

### AI Architecture (Hybrid Brain)
- **Rules Engine**: Deterministic Go rules engine for game decisions
- **Language Model**: Large Language Model for human-like communication only
- **MCP Interface**: Model Context Protocol for secure AI-game communication

### Frontend Hybrid Design
- **Go/Wasm Core**: Client-side game logic and WebSocket management
- **React Shell**: UI rendering and user interaction handling
- **Bridge Layer**: JavaScript connects Wasm and React components

## Testing Strategy

### Coverage Requirements
- **Core Logic**: 95% unit test coverage for `ApplyEvent` and `RulesEngine` functions
- **Integration**: Actor black-box testing with mocked dependencies
- **System**: Supervisor resilience and failure scenario testing

### Test Patterns
```go
// Table-driven tests for pure functions
func TestApplyEvent(t *testing.T) {
    testCases := []struct {
        name          string
        initialState  GameState
        event         Event
        expectedState GameState
    }{
        // Test cases here
    }
}

// Actor integration tests with mocks
func TestActor_PlayerJoinsAndVotes(t *testing.T) {
    mockStore := &MockDataStore{}
    actor := NewGameActor("test-game", mockStore)
    // Send actions and assert on mock calls
}
```

## Project Structure (Planned)

```
./server/           # Go backend
  internal/game/    # Game state and ApplyEvent logic
  internal/ai/      # Rules engine and Language Model integration
  internal/comms/   # WebSocket communication
  internal/actors/  # Game Actor and Supervisor
  cmd/server/       # Main server binary

./client/           # React/TypeScript frontend
  src/             # React components and UI
  wasm/            # Go/Wasm game engine source

./docs/             # Comprehensive design documentation
  architecture/     # Backend system details
  adr/             # Architectural Decision Records
  api/             # WebSocket events and data structures
  development/     # Core logic specs and testing strategy
```

## Important Documentation

- **`docs/02-onboarding-for-engineers.md`**: Essential 5-minute technical overview
- **`docs/01-game-design-document.md`**: Complete game rules and mechanics
- **`docs/architecture/README.md`**: Detailed backend architecture explanations
- **`docs/adr/README.md`**: Architectural decisions with context and rationale
- **`docs/development/01-core-logic-definition.md`**: Core function specifications
- **`docs/development/02-testing-strategy.md`**: Testing requirements and patterns

## Current Status

This repository is in **design and documentation phase**. All architecture has been planned and documented, but no Go/React source code exists yet. The next development phase will implement the documented specifications.