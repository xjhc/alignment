# Architecture: Actor and Event Processing

This document details the core processing loop of a single **Game Actor**. This loop is the heart of our system, responsible for handling all player actions and state changes for a given game in a consistent, performant, and durable manner.

## 1. Architectural Context

A Game Actor is a self-contained goroutine that "owns" a single game. It receives all incoming player actions for its game from the **Dispatcher** via a dedicated Go channel (its "mailbox"). The actor's primary responsibility is to process these actions sequentially, ensuring a strict, ordered transformation of the game's state.

## 2. The "Validate -> Persist -> Apply" Pattern

Every action received by the actor is processed through a critical, three-stage pattern. This ensures that our state is always consistent and that no action is confirmed until it has been durably stored.

```ascii
             +------------------------------------------------------+
             |                    Game Actor Loop                   |
             |                                                      |
[Action]---->| 1. Validate Action (against in-memory GameState)     |
             |       |                                              |
             |       | (valid)                                      |
             |       v                                              |
             | 2. Persist Event (Write to Redis Stream - The WAL)   |
             |       |                                              |
             |       | (success)                                    |
             |       v                                              |
             | 3. Apply Event (Update in-memory GameState)          |
             |       |                                              |
             |       v                                              |
             | 4. Broadcast Event (to all clients in the game)------> [Event]
             |                                                      |
             +------------------------------------------------------+
```

1.  **Validate:** When an action arrives, the actor first validates it against its current, in-memory `GameState`. This is a near-instantaneous, CPU-bound check. This step acts as our primary security and anti-cheat layer, ensuring a player cannot perform an illegal move (e.g., voting in the wrong phase).

2.  **Persist (Write-Ahead Log):** If the action is valid, the actor creates a corresponding, immutable `Event` object. This event is immediately written to the **Redis Stream** for that game. This is the "Write-Ahead Log" (WAL) pattern and is the only blocking I/O in the hot path. By persisting *before* applying, we guarantee that if the server crashes at any moment, we have a record of the event and can recover the state perfectly.

3.  **Apply:** Once the event is successfully persisted in Redis, the actor applies it to its in-memory `GameState` by calling the pure `applyEvent` function. This function takes the current state and the new event and returns the updated state.

4.  **Broadcast:** Finally, the actor sends the new event to the Dispatcher, which broadcasts it to all connected clients in that game, informing their UIs of the state change.

## 3. The Role of `applyEvent`

The `applyEvent(state, event)` function is the deterministic core of our game's rules. It is a **pure function**, meaning it has no side effects and its output depends only on its inputs. This isolation is critical for testability. We can unit test every single game rule and state transition with 100% confidence, entirely separate from the complexities of the surrounding concurrent actor system.
