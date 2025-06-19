# Backend Architecture Overview

This document provides a high-level overview of the `Alignment` backend architecture. Our design philosophy prioritizes **low latency, high concurrency, and operational resilience** on a single-machine deployment.

To achieve this, we have implemented a **stateful, in-memory, supervised Actor Model**.

---

## Core Components

The backend is composed of several key, concurrent components that work together. Understanding their distinct roles is key to understanding the system.

*   **The Supervisor (The Guardian):** The top-level goroutine that launches and monitors all active games. Its primary role is to provide fault isolation; if a single `Game Actor` panics, the Supervisor catches the error and terminates that one game without crashing the entire server.

*   **The Game Actor (The Workhorse):** A dedicated goroutine that "owns" a single game. It holds that game's complete state in memory (`GameState`) and processes all actions and events for that game serially via a private channel. This is the source of our performance and data consistency.

*   **The Dispatcher (The Router):** The central hub for all network traffic. It listens to all incoming WebSocket messages from players, identifies the target game, and routes the message to the correct Game Actor's channel (mailbox).

*   **The Scheduler (The Metronome):** A single, highly-efficient goroutine that manages all time-based events for the entire server (e.g., phase timers, AI thinking delays). It uses a **Timing Wheel** algorithm to handle thousands of timers with minimal overhead.

*   **Redis (The Scribe):** Our external persistence layer. It is used exclusively as a **Write-Ahead Log (WAL)** to record the event history and for storing **State Snapshots** to enable fast recovery. **It is not read from during normal gameplay.**

## System Flow Diagram

```ascii
                               +---------------------------------------------+
                               |        Go Backend Process (Single VM)       |
                               |                                             |
+----------+      +------------+      +------------------+      +-----------+
| WebSocket|----->| Dispatcher |----->|   Supervisor     |----->| Game      |
| Connection      | (Routes msg)      | (Manages Actors) |      | Actor     |
+----------+      +------------+      +------------------+      | (In-Memory|
                               |              ^                 |  State)   |
                               |              | (Time's Up)     |           |
                               |              |                 |           |
                               |      +------------------+      |           |
                               +----->|   Scheduler      |<-----+           |
                                      | (Timing Wheel)   |  (Schedule Timer)
                                      +------------------+
```

## The Lifecycle of a Player Action

1.  **Ingress:** A player action arrives via a WebSocket message.
2.  **Dispatch:** The **Dispatcher** routes the message to the correct **Game Actor's** mailbox channel.
3.  **Process:** The **Game Actor** validates the action against its in-memory state.
4.  **Persist (WAL):** The actor creates a corresponding event and writes it to the **Redis Stream**.
5.  **Apply:** The actor applies the event to its in-memory `GameState`.
6.  **Broadcast:** The actor sends the new event back to the Dispatcher to be broadcast to all clients in the game.

## Deep Dives

This document is a high-level map. For detailed implementations and logic, please refer to the following documents:

*   **[Supervisor & Resiliency](./01-supervisor-and-resiliency.md):** A detailed look at the Supervisor, Health Monitor, and Admission Controller that protect the server from crashes and overload.
*   **[AI Player Design](./02-ai-player-design.md):** An explanation of our Hybrid AI model, separating the "Strategic Brain" (Rules Engine) from the "Language Brain" (LLM).
*   **[MCP Interface](./03-mcp-interface.md):** The formal specification for the read-only API we use to provide game context to the AI's language model.