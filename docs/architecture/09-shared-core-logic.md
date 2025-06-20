# Architecture: Shared Core Logic (Backend & Frontend)

This document outlines the strategy for sharing critical game logic between the Go backend and the Go/Wasm frontend. The primary goal is to **adhere to the DRY (Don't Repeat Yourself) principle**, ensuring that the rules of the game are defined in a single place and behave identically on both the server and the client.

## 1. The Core Problem

Our architecture is event-driven. The server processes actions and broadcasts events, and the client receives these events to build a local "read model" of the game state.

The function that processes these events—our `applyEvent(state, event)` function—is the heart of the game's rules. Implementing this logic separately on both the backend and frontend would lead to:
*   Duplicated effort.
*   The high risk of subtle bugs where the client and server state diverge.
*   A maintenance nightmare, as every rule change would need to be implemented in two places.

## 2. The Solution: A Shared `core` Package

We will create a new, dedicated module within our monorepo that contains all the domain-level logic and data structures for the game. This package will have **zero dependencies on backend- or frontend-specific code**. It will be pure, portable Go.

This package will be located at a top-level directory, for example, `/core`.

#### **Project Structure:**

```
/
├── server/             # Go backend (imports /core)
│   └── ...
├── client/             # Frontend (imports /core)
│   ├── main.go         # Go/Wasm entrypoint
│   └── ...
└── core/               # NEW: The shared logic package
    ├── game_state.go   # Defines the GameState struct and the ApplyEvent function
    ├── types.go        # Defines shared types like Player, Event, Action, etc.
    └── rules.go        # (Optional) Contains pure rule-checking functions
```

#### **Contents of the `core` Package:**

*   **`types.go`:** This file will define all the fundamental data structures that are shared between the client and server. This includes the `Player`, `ChatMessage`, `Event`, and `Action` structs, complete with their `json` tags for serialization.
*   **`game_state.go`:** This file will contain two key items:
    1.  The `GameState` struct, which represents the full state of a game.
    2.  The `ApplyEvent(currentState GameState, event Event) GameState` function. This is the **most critical piece of shared logic**. It is a pure function that takes a state and an event and returns the new state.

## 3. How It's Used

#### **On the Backend (`/server`):**

The `GameActor` will import the `/core` package.
*   When processing an action, the actor will validate it, persist the resulting `core.Event`, and then update its in-memory state by calling:
    ```go
    // In the GameActor...
    import "alignment/core"

    // ...
    newState := core.ApplyEvent(a.currentState, newEvent)
    a.currentState = newState
    ```

#### **On the Frontend (`/client`):**

The main Go/Wasm module will also import the `/core` package.
*   When the WebSocket connection receives an event from the server, the Wasm module will call the exact same function to update its local copy of the game state.
    ```go
    // In the Wasm client's main loop...
    import "alignment/core"

    // ...
    // eventFromServer is a core.Event received over WebSocket
    localGameState = core.ApplyEvent(localGameState, eventFromServer)
    // Now, call a JS function to tell React to re-render with the new state.
    updateReactUI(localGameState)
    ```

## 4. Benefits of This Architecture

*   **Guaranteed Consistency:** It is now **impossible** for the client and server logic to diverge. They are compiled from the exact same source code. The `ApplyEvent` function will always produce the same output given the same input, both on the server and in the browser.
*   **Single Source of Truth for Rules:** Game rules are now implemented only once. If we need to change how tokens are awarded, we change it in `core/game_state.go`, and the change is instantly reflected on both the backend and frontend after the next compile.
*   **Type Safety Across the Wire:** Since both client and server use the same `core.Event` and `core.Action` structs, we have compile-time safety for our communication protocol. This drastically reduces the chance of serialization or deserialization errors.
*   **Simplified Testing:** We can write a single, exhaustive set of unit tests for the `core` package, and by doing so, we are simultaneously validating the logic for both the server and the client.

This shared `core` package is the definitive solution to code duplication in a Go + Go/Wasm project and is a cornerstone of our development strategy.