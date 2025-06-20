# API & Communication Protocols

This directory contains the definitive reference for all communication between the **`Alignment` server** and the **client application**. Our communication is built exclusively on a persistent **WebSocket** connection.

---

## 1. Core Principles

*   **Single Connection:** A client establishes a single WebSocket connection upon joining a game and maintains it for the entire session.
*   **Event-Driven Architecture:** The server does not expose a traditional REST or RPC API. Instead, it operates on an event-driven model. The client sends **Actions** to the server, and the server broadcasts **Events** to all clients.
*   **JSON Payload:** All messages sent over the WebSocket, in either direction, are serialized as JSON objects.
*   **Source of Truth:** The server is the absolute source of truth. The client's state is a local projection of the event stream it receives from the server. Clients should never assume an action is successful until they receive a corresponding event from the server.

---

## 2. Message Envelope

All messages adhere to a simple envelope structure to distinguish their type.

#### **Client → Server (Actions)**

A client sends an `Action` to declare its intent to do something.

```json
{
  "action": "ACTION_NAME",
  "payload": { ... }
}
```

*   **`action`**: A string identifying the type of action (e.g., `SUBMIT_VOTE`, `POST_CHAT_MESSAGE`).
*   **`payload`**: A JSON object containing the data required for that action.

#### **Server → Client (Events)**

The server broadcasts an `Event` to notify clients that something has officially happened.

```json
{
  "event_id": "1678886400000-0",
  "event_type": "EVENT_NAME",
  "payload": { ... }
}
```

*   **`event_id`**: The unique ID of the event, taken directly from the Redis Stream. This is used by clients for reconnecting and catching up.
*   **`event_type`**: A string identifying the type of event (e.g., `PLAYER_JOINED`, `PHASE_CHANGED`).
*   **`payload`**: A JSON object containing the data describing what happened.

---

## 3. Detailed Documentation

For a complete and detailed list of all possible messages and their data structures, please refer to the following documents:

*   **[WebSocket Events & Actions](./01-websocket-events-and-actions.md):** A comprehensive list of every `action` and `event` type, including their specific `payload` schemas. This is the primary reference for client-side developers.
*   **[Core Data Structures](./02-data-structures.md):** Detailed definitions of the primary Go structs (`GameState`, `Player`, etc.) that are serialized within the message payloads.
*   **[Information Visibility Model](./03-information-visibility-model.md):** A critical guide defining what data can be seen by whom. This is a must-read for security and game integrity.