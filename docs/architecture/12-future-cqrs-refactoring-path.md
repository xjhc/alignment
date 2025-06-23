# Architecture: Future Path to a CQRS-Based System

## 1. Context and Current State (V1)

The V1 architecture for `Alignment` is a stateful, in-memory actor model with unified managers (`GameLifecycleManager`). This design is optimized for simplicity, low cost, and rapid development on a single-machine deployment. It uses direct method calls between components (e.g., `PlayerActor` -> `GameLifecycleManager`), which is clear and performant in a single-process environment.

However, as the application scales and its reliability requirements increase, this tightly-coupled model will present challenges in scalability, testability, and resilience.

This document outlines the formal, planned evolution of the architecture to a fully decoupled system based on **Command-Query Responsibility Segregation (CQRS)** and **Event Sourcing**. This is the designated "V2" architecture.

## 2. The Trigger for Refactoring

The transition to this architecture will be considered when one or more of the following become true:
*   The business need arises to scale the application across multiple server nodes.
*   The complexity of a long-running process (like Night Phase Resolution) becomes too difficult to manage within a single manager, and requires orchestration.
*   We require an absolute, "zero-loss" guarantee for event processing, even in the case of a server crash during a write operation.
*   The tight coupling between components begins to significantly slow down development and increase the rate of bugs.

## 3. The "V2" Target Architecture: CQRS and Event Sourcing

The target architecture fully decouples all major components using a command bus and an event bus.

```mermaid
graph TD
    subgraph Client
        UI -- sends --> Commands[Client Commands]
    end

    subgraph Server
        PA[PlayerActor] -- forwards --> Bus((Command Bus))
        Bus --> CH[Command Handler]

        subgraph Orchestration
            CH -->|publishes| EBus((Event Bus))
            PM[Process Manager] -- subscribes --> EBus
            PM -- issues --> Bus
        end

        subgraph State
            GA[GameActor] -- subscribes to commands --> Bus
            GLM[GameLifecycleManager<br>(Read Model)] -- subscribes to events --> EBus
        end

        EBus --> GA
        EBus --> GLM

        subgraph Broadcasting
            WSM[WebSocketManager] -- subscribes --> EBus
            WSM -->|sends Event| Client
        end
    end

    Client -- WebSocket --> PA
```

## 4. Phased Implementation Plan

The migration can be performed in distinct, manageable phases.

#### **Phase 1: Introduce Commands and a Command Bus**
*   **Action:** Formalize all state-changing actions as `Command` structs (e.g., `SubmitVoteCommand`). Create a central `CommandBus`.
*   **Result:** Actors no longer call manager methods directly. They dispatch commands to the bus. The `GameLifecycleManager` and `GameActor` become the primary command handlers. This begins the decoupling process.

#### **Phase 2: Introduce an Event Bus and Decouple Read Models**
*   **Action:** Create the `EventBus`. Command Handlers will now publish `Event`s instead of directly changing state. Components like the `GameLifecycleManager` (for its lobby list) and `WebSocketManager` will subscribe to these events to build their own state (their "read models").
*   **Result:** The "write" side (handling commands) is now fully separate from the "read" side (reacting to events).

#### **Phase 3: Isolate Complex Logic with Process Managers (Sagas)**
*   **Action:** Identify complex, multi-step processes like `StartGame` or `ResolveNightPhase`. Create dedicated `ProcessManager` structs for each. These process managers listen for a sequence of events and issue new commands to orchestrate the flow.
*   **Result:** The main `CommandHandler` becomes much simpler, only handling atomic actions. Complex orchestration is isolated into its own testable unit.

#### **Phase 4: Implement the Outbox Pattern for Atomic Persistence**
*   **Action:** For maximum data integrity, implement the Outbox Pattern. When handling a command, a `CommandHandler` will, in a single Redis transaction, write the resulting event(s) to the event stream AND add a reference to a special `outbox` list. A separate, single-threaded "Event Dispatcher" will safely read from this outbox and publish the events to the in-memory `EventBus`.
*   **Result:** This provides an ironclad guarantee that any event successfully persisted to the database will eventually be published to the rest of the system, even if the server crashes immediately after the database write.

## 5. Conclusion

By documenting this path now, we acknowledge the trade-offs made for V1 while providing a clear and professional blueprint for the future. This ensures that current development can proceed with speed and focus, without sacrificing long-term architectural integrity.