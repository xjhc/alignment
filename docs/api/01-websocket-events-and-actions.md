# WebSocket: Events & Actions

This document provides a comprehensive list of all messages exchanged between the client and server over the WebSocket connection.

## I. Client → Server Actions

These are the commands a client can send to the server. The server will validate each action and, if valid, generate one or more corresponding events.

| Action Name | Payload | Description |
| :--- | :--- | :--- |
| **`RECONNECT`** | `{ "game_id": string, "player_id": string, "session_token": string, "last_event_id": string }` | Sent immediately upon connection to rejoin an active game. The `last_event_id` tells the server which events the client has already seen, allowing for an efficient catch-up. |
| **`CREATE_GAME`** | `{ "player_name": string }` | Asks the server to create a new game lobby and join it as the host. |
| **`JOIN_GAME`** | `{ "game_id": string, "player_name": string }` | Joins an existing game lobby. |
| **`START_GAME`** | `{}` | Sent by the lobby host to begin the game, assigning roles and starting Day 1. |
| **`POST_CHAT_MESSAGE`**| `{ "content": string }` | Sends a single chat message to be broadcast to other players. |
| **`UPDATE_STATUS`**| `{ "status": string }` | Updates the player's public Slack Status message (max 20 chars). |
| **`SUBMIT_NIGHT_ACTION`**| `{ "type": string, "target_player_id"?: string }` | Submits the player's choice for the night. `type` can be `MINE`, `PROJECT_MILESTONES`, or a role-specific ability like `RUN_AUDIT`. |
| **`SUBMIT_VOTE`** | `{ "vote_target_id"?: string, "verdict"?: string }` | Casts a vote. During nomination, `vote_target_id` is used. During the verdict, `verdict` (`YES` or `NO`) is used. |
| **`SUBMIT_PULSE_CHECK`**| `{ "response": string }` | Submits the player's one-sentence response to the daily Pulse Check prompt. |
| **`SUBMIT_EXIT_INTERVIEW`**| `{ "action": string, "target_player_id"?: string, "final_status": string }` | Sent by a just-eliminated player. `action` can be `HANDOFF`, `CONFIDENTIAL_FEEDBACK`, or `BURN_BRIDGES`. |

---

## II. Server → Client Events

These are the immutable facts the server broadcasts. The client uses these events to construct and update its local `GameState`.

| Event Type | Payload | Description |
| :--- | :--- | :--- |
| **`GAME_CREATED`** | `{ "game_id": string, "host_player_id": string }` | Confirms a new game lobby has been created. |
| **`PLAYER_JOINED`** | `{ "player": PlayerObject }` | A new player has joined the lobby. |
| **`PLAYER_LEFT`** | `{ "player_id": string }` | A player has disconnected or been removed. |
| **`ROLES_ASSIGNED`** | `{ "your_role": RoleInfo }` | **Sent privately** to each player at the start of the game, revealing their role, alignment, and secret Personal KPI. |
| **`PHASE_CHANGED`** | `{ "new_phase": string, "duration_sec": int, "day_number": int, "crisis_event"?: CrisisEventObject }` | Signals a new game phase (`LOBBY`, `DAY`, `NIGHT`, `END`). The daily crisis event is announced with the `DAY` phase change. |
| **`CHAT_MESSAGE_POSTED`**| `{ "message": ChatMessageObject }` | A new chat message to be displayed. |
| **`STATUS_UPDATED`** | `{ "player_id": string, "new_status": string }` | A player's public status has changed. |
| **`PULSE_CHECK_SUBMITTED`**| `{ "player_id": string, "response": string }` | Broadcasts a player's response to the Pulse Check *after* the submission window closes. |
| **`NIGHT_ACTIONS_RESOLVED`**| `NightResultsObject` | **Key Event:** A detailed summary of everything that happened during the night (mining, conversions, blocks, etc.). See the `data-structures` doc for details. |
| **`VOTE_TALLY_UPDATED`**| `{ "type": string, "results": map[string]int }` | Announces the results of a token-weighted vote. `type` is `NOMINATION` or `VERDICT`. The map is `option -> total_tokens`. |
| **`PLAYER_ELIMINATED`** | `{ "player_id": string, "role": string, "alignment": string }` | A player was deactivated, revealing their role and alignment to all. |
| **`GAME_ENDED`** | `{ "winning_faction": string, "reason": string, "player_states": Player[] }` | Announces the end of the game, the winner, and the final state of all players. |
| **`SYNC_COMPLETE`** | `{}` | **Sent privately** to a reconnecting client after its batch of catch-up events has been delivered, signaling it's now up-to-date. |
| **`PRIVATE_NOTIFICATION`**| `{ "message": string, "type": string }` |
---

### `docs/api/02-data-structures.md`
