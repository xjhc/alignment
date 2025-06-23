# Backend Architecture Overview

This document provides a high-level overview of the `Alignment` backend architecture. Our design philosophy prioritizes **low latency, high concurrency, and operational resilience** on a single-machine deployment.

To achieve this, we have implemented a **stateful, in-memory, supervised Actor Model**.

---

## Core Components

The backend is composed of several key, concurrent components that work together. Understanding their distinct roles is key to understanding the system.

*   **The [Supervisor](./01-supervisor-and-resiliency.md) (The Guardian):** The top-level goroutine that launches and monitors all active games. Its primary role is to provide fault isolation; if a single `Game Actor` panics, the Supervisor catches the error and terminates that one game without crashing the entire server. It is the core of a wider **resiliency layer** that also includes a **Health Monitor** and **Admission Controller** to protect the server from overload.

*   **The [Game Actor](./04-actor-and-event-processing.md) (The Workhorse):** A dedicated goroutine that "owns" a single game. It holds that game's complete state in memory (`GameState`) and processes all actions and events for that game serially via a private channel. This is the source of our performance and data consistency.

*   **The [Dispatcher](../glossary.md#dispatcher) (The Router):** The central hub for all network traffic. It listens to all incoming WebSocket messages from players, identifies the target game, and routes the message to the correct Game Actor's channel (mailbox). In a single-node deployment, it also handles broadcasting events back to clients; this role shifts to a Redis Pub/Sub model in a [multi-node environment](./07-future-scaling-path.md).

*   **The [Scheduler](../glossary.md#scheduler) (The Metronome):** A single, highly-efficient goroutine that manages all time-based events for the entire server (e.g., phase timers, AI thinking delays). It uses a **[Timing Wheel](../glossary.md#timing-wheel)** algorithm to handle thousands of timers with minimal overhead.

*   **Redis (The Scribe):** Our external persistence layer. It is used exclusively as a **[Write-Ahead Log (WAL)](../glossary.md#wal-write-ahead-log)** to record the event history and for storing **State Snapshots** to enable fast recovery. **It is not read from during normal gameplay.**

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

```mermaid
sequenceDiagram
    participant Client as Player (WebSocket)
    participant Dispatcher as Dispatcher<br/>(Router)
    participant Actor as Game Actor<br/>(Workhorse)
    participant Redis as Redis<br/>(WAL)
    participant Clients as All Game Clients

    Client->>Dispatcher: 1. Action arrives<br/>(WebSocket message)
    Dispatcher->>Actor: 2. Route to mailbox<br/>(Go channel)
    Actor->>Actor: 3. Validate action<br/>(against in-memory state)
    Actor->>Redis: 4. Persist event<br/>(Write-Ahead Log)
    Redis-->>Actor: WAL write confirmed
    Actor->>Actor: 5. Apply event<br/>(update GameState)
    Actor->>Dispatcher: 6. Send event for broadcast
    Dispatcher->>Clients: Broadcast event<br/>(to all game clients)
```

## Deep Dives

This document is a high-level map. For detailed implementations and logic, please refer to the following documents:

*   **[Supervisor & Resiliency](./01-supervisor-and-resiliency.md):** A detailed look at the Supervisor, Health Monitor, and Admission Controller that protect the server from crashes and overload.
*   **[AI Player Design](./02-ai-player-design.md):** An explanation of our Hybrid AI model, separating the Rules Engine from the Language Model.
*   **[MCP Interface](./03-mcp-interface.md):** The formal specification for the read-only API we use to provide game context to the AI's language model.