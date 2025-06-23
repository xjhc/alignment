
# ADR-001: Adopt an In-Memory Actor Model

*   **Status:** Accepted

### Context

The `Alignment` game is a real-time system where multiple players interact concurrently within a single game instance. Low latency is a critical product requirement. An initial architectural consideration was a traditional stateless web server model, where each player action would trigger a request that reads the current game state from a database (like Redis), processes the action, and writes the new state back.

Performance analysis of this stateless model revealed a significant bottleneck: to validate any single action, the server would need to read and process the entire event history of the game from Redis. For a game with hundreds of events, this I/O and computation overhead would lead to unacceptably high latency, especially under load.

### Decision

We will implement a **stateful, in-memory Actor Model** for the backend architecture.

1.  Each active game will be managed by its own dedicated goroutine, referred to as a **Game Actor**.
2.  This Game Actor will hold the entire `GameState` for its game in the server's memory.
3.  All actions and events for a specific game will be funneled through a dedicated Go channel (`chan`) which serves as the actor's "mailbox." This enforces serialized processing of events for a single game, guaranteeing data consistency without the need for mutexes.
4.  Redis will not be used for live reads. Its role will be relegated to a **Write-Ahead Log (WAL)** for persistence and recovery. Events will be appended to a Redis Stream *before* being applied to the in-memory state.

### Consequences

*   **Pros:**
    *   **High Performance:** Action validation and state mutation become near-instantaneous, CPU-bound operations against the in-memory `GameState`, drastically reducing latency.
    *   **High Concurrency:** Different games run in isolated, parallel goroutines, allowing the server to handle many concurrent games efficiently on a single machine.
    *   **Simplified Logic:** The serialized, single-threaded nature of processing within an actor simplifies game logic by eliminating the need to reason about race conditions or lock contention for a given game's state.

*   **Cons:**
    *   **Increased Complexity:** This is a stateful architecture. We are now responsible for managing goroutine lifecycles, memory usage, and a more complex crash recovery strategy.
    *   **Single Point of Failure (Process Level):** While games are isolated from each other, a crash in the main server process will terminate all active games simultaneously. State can be recovered from Redis on restart, but the user experience is a hard stop.
    *   **Memory Footprint:** The server's memory capacity becomes a hard limit on the number of concurrent games it can host. Careful memory management and profiling are required.
