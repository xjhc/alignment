# WebSocket: Events & Actions

This document provides a comprehensive list of all messages exchanged between the client and server over the WebSocket connection.

## Important: The Hybrid REST + WebSocket Model

This project uses a hybrid communication model for session management.

1.  **Session Initiation (REST API):** Creating a new lobby or joining an existing one is handled via a standard REST API (`/api/games`). These endpoints return the necessary credentials (`game_id`, `player_id`, `session_token`).
2.  **Real-time Communication (WebSocket):** Once credentials are acquired, the client establishes a single, persistent WebSocket connection. All subsequent real-time game actions are sent as JSON messages over this connection.

The actions listed below are only those sent over the **WebSocket**.


## I. Client → Server Actions

These are the commands a client can send to the server. The server will validate each action and, if valid, generate one or more corresponding events.

| Action Name | Payload | Description |
| :--- | :--- | :--- |
| **`RECONNECT`** | `{ "game_id": string, "player_id": string, "session_token": string }` | Sent immediately upon connection to rejoin an active game. The server will respond with a `GAME_STATE_SNAPSHOT` to bring the client up-to-date instantly. |
| **`CREATE_GAME`** | `{ "player_name": string }` | Asks the server to create a new game lobby and join it as the host. |
| **`JOIN_GAME`** | `{ "game_id": string, "player_name": string }` | Joins an existing game lobby. |
| **`START_GAME`** | `{}` | Sent by the lobby host to begin the game, assigning roles and starting Day 1. |
| **`POST_CHAT_MESSAGE`**| `{ "content": string }` | Sends a single chat message to be broadcast to other players. |
| **`UPDATE_STATUS`**| `{ "status": string }` | Updates the player's public Player Status message (max 20 chars). |
| **`SUBMIT_NIGHT_ACTION`**| `{ "type": string, "data": object }` | Submits the player's choice for the night. The `data` payload is specific to the action `type`. <br> **Examples:** <br> `MINE`: `{ "target_player_id": "p-xyz" }` <br> `REALLOCATE_BUDGET`: `{ "source_player_id": "p-abc", "destination_player_id": "p-def" }` |
| **`SUBMIT_VOTE`** | `{ "vote_target_id"?: string, "verdict"?: string }` | Casts a vote. During nomination, `vote_target_id` is used. During the verdict, `verdict` (`YES` or `NO`) is used. |
| **`SUBMIT_PULSE_CHECK`**| `{ "response": string }` | Submits the player's one-sentence response to the daily Pulse Check prompt. |
| **`SUBMIT_EXIT_INTERVIEW`**| `{ "action": string, "target_player_id"?: string, "final_status": string }` | Sent by a just-deactivated player. `action` can be `HANDOFF`, `CONFIDENTIAL_FEEDBACK`, or `BURN_BRIDGES`. |

---

## II. Server → Client Events

These are the immutable facts the server broadcasts. The client uses these events to construct and update its local `GameState`.


| Event Type | Payload | Description |
| :--- | :--- | :--- |
| **`GAME_STATE_SNAPSHOT`**| `{ "game_state": GameState }` | **Sent privately** upon reconnect. Contains the complete, authoritative core game state (players, phase, etc.) but omits bulky historical data like the chat log. This is used to instantly hydrate the client's UI. |
| **`CHAT_HISTORY_SNAPSHOT`**| `{ "messages": ChatMessage[] }` | **Sent privately** upon reconnect, immediately after the `GAME_STATE_SNAPSHOT`. Contains the full chat history to allow the client to backfill its chat panel. |
| **`PLAYER_JOINED`** | `{ "player": PlayerObject }` | A new player has joined the lobby. |
| **`PLAYER_LEFT`** | `{ "player_id": string }` | A player has disconnected from the lobby or game. |
| **`PLAYER_DEACTIVATED`** | `{ "player_id": string, "revealed_role": string, "revealed_alignment": string }` | A player has been voted out. This event crucially reveals their final role and alignment to all players. |
| **`ROLE_ASSIGNED`** | `{ "your_role": RoleInfo }` | **Sent privately** to each player at the start of the game, revealing their role, alignment, and secret Personal KPI. |
| **`ALIGNMENT_CHANGED`** | `{ "new_alignment": string }` | **Sent privately** to a player when they have been converted by the AI faction. Signals the client to update its state and reveal AI-faction UI elements. |
| **`PHASE_CHANGED`** | `{ "new_phase": string, "duration_sec": int, "day_number": int, "crisis_event"?: CrisisEventObject }` | Signals a new game phase (`LOBBY`, `DAY`, `NIGHT`, `END`). The daily crisis event is announced with the `DAY` phase change. |
| **`CHAT_MESSAGE`**| `{ "message": ChatMessageObject }` | A new chat message to be displayed. |
| **`PULSE_CHECK_SUBMITTED`**| `{ "player_id": string, "player_name": string, "response": string }` | A player's response to the daily Pulse Check. The client should display this publicly with attribution. |
| **`NIGHT_ACTIONS_RESOLVED`**| `{ "results": NightResultsObject }` | Summarizes the outcomes of the Night Phase. The full `NightResultsObject` is defined in the [Core Data Structures](./02-data-structures.md) document. This event triggers the start of the next Day Phase. |
| **`GAME_ENDED`** | `{ "winning_faction": string, "reason": string, "player_states": Player[] }` | Announces the end of the game, the winner, and the final state of all players. |
| **`PRIVATE_NOTIFICATION`**| `{ "message": string, "type": string }` | **Sent privately** to a single player to deliver sensitive information that only they should see. The `type` field allows the client to handle different kinds of notifications. <br> **Examples:** <br> • `"type": "SYSTEM_SHOCK_AFFLICTED"` <br> • `"type": "KPI_OBJECTIVE_COMPLETED"`|
---