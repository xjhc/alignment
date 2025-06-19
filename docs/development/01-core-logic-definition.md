# Core Logic Definition

In the `Alignment` codebase, "Core Logic" refers to the deterministic, pure-functional heart of the game. This is the code that encodes the *rules* of the game, not the infrastructure that runs it. It must be completely isolated from side effects like database writes, network calls, or concurrency.

Our two most critical pieces of core logic are:

1.  **`ApplyEvent(state, event)` function:**
    *   **Signature:** `func ApplyEvent(currentState GameState, event Event) GameState`
    *   **Description:** This pure function takes the current state of a game and a single event, and returns the new state of the game. It is the single source of truth for state transitions. For example, it defines that a `PLAYER_VOTED` event adds a vote to the tally, or a `MINING_SUCCESSFUL` event increments a player's token count.
    *   **Location:** `internal/game/state.go`

2.  **`RulesEngine.DecideAction(state)` methods:**
    *   **Signature:** `func (re *RulesEngine) DecideVote(currentState GameState) Action`
    *   **Description:** This is a collection of pure functions that encapsulates the AI's strategic decision-making. Given a `GameState`, it deterministically calculates the optimal move (e.g., who to vote for, who to target). It contains the `calculateThreat` and `calculateSuspicionScore` heuristics.
    *   **Location:** `internal/ai/rules.go`

By keeping this logic pure, we can test it exhaustively and be 100% confident in its correctness, separate from the complexities of the surrounding actor system.
