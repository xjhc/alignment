# Core Logic Definition

In the `Alignment` codebase, "Core Logic" refers to the deterministic, pure-functional heart of the game. This logic is separated into two categories based on its boundaries.

---

### 1. Universal Core Logic (`/core`)

This is the code that encodes the **universal rules of the game**. It is shared between the server and the Go/Wasm client to ensure they both interpret events identically. It must be completely isolated from all side effects.

*   **Primary Example: `ApplyEvent(state, event)` function**
    *   **Signature:** `func ApplyEvent(currentState GameState, event Event) GameState`
    *   **Description:** This pure function is the single source of truth for state transitions. It takes the current state of a game and a single event, and returns the new state. For example, it defines that a `MINING_SUCCESSFUL` event increments a player's token count.
    *   **Location:** `core/game_state.go`

### 2. Server-Side Core Logic (`/server`)

This is the code that encodes the **authoritative, secret, or infrastructure-dependent rules**. It runs only on the server and is not shared with the client.

*   **Primary Example: `RulesEngine.DecideAction(state)` methods**
    *   **Signature:** `func (re *RulesEngine) DecideVote(currentState GameState) Action`
    *   **Description:** This is a collection of pure functions that encapsulates the AI's strategic decision-making. Given a `GameState`, it deterministically calculates the optimal move (e.g., who to vote for, who to target). It contains the secret `calculateThreat` and `calculateSuspicionScore` heuristics.
    *   **Location:** `server/internal/ai/rules.go`

By keeping both categories of logic pure within their respective boundaries, we can test them exhaustively and be confident in their correctness, separate from the complexities of the surrounding actor system.