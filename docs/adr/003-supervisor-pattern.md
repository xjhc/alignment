
# ADR-003: Isolate Actor Failures with a Supervisor Pattern

*   **Status:** Accepted

### Context

The adoption of the In-Memory Actor Model (ADR-001) introduces a significant operational risk. Because each game runs as a goroutine within a single server process, an unhandled panic within the code for a *single* game (e.g., due to a rare bug or unexpected data) would crash the entire server process. This would abruptly terminate all other healthy, concurrent games, leading to a poor user experience and server-wide instability.

### Decision

We will implement a **Supervisor pattern** to provide fault isolation between Game Actors.

1.  A top-level **Supervisor** goroutine will be responsible for launching all Game Actors.
2.  Each Game Actor will be spawned in its own new goroutine, wrapped with a `defer` block containing a `recover()` statement.
3.  If a Game Actor goroutine panics, the Supervisor's `recover` function will catch the panic.
4.  Upon catching a panic, the Supervisor will:
    *   Log the error in detail, including the `gameID` and the panic payload (stack trace).
    *   Allow the panicking goroutine to terminate cleanly.
    *   **Crucially, the Supervisor itself will not panic.** It will continue running, and all other Game Actors will remain completely unaffected.

### Consequences

*   **Pros:**
    *   **Greatly Improved Server Stability:** The server process becomes resilient to bugs within individual game logic. A single faulty game can no longer cause a server-wide outage.
    *   **Fault Isolation:** Errors are contained to the specific game in which they occur, preserving the experience for all other players on the server.
    *   **Enhanced Debugging:** By logging the specific `gameID` associated with a panic, we can more easily trace the state and sequence of events that led to the failure.

*   **Cons:**
    *   **Terminated Games are Still Lost:** While the server remains stable, the players in the game that crashed are still disconnected. The user experience for *that specific game* is still a hard stop. For our V1, this is an acceptable trade-off.
    *   **Adds a Layer of Abstraction:** Developers must be aware that actors are running under a supervisor and what the recovery semantics are.

This pattern is a critical component of our strategy for running a stable, multi-tenant game server on a single machine.
