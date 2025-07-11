# WebSocket: Events & Actions

This document provides a comprehensive list of all messages exchanged between the client and server over the WebSocket connection.

## Important: The Hybrid REST + WebSocket Model

This project uses a hybrid communication model for session management.

1.  **Session Initiation (REST API):** Creating a new lobby or joining an existing one is handled via a standard REST API (`/api/games`). These endpoints return the necessary credentials (`game_id`, `player_id`, `session_token`).
2.  **Real-time Communication (WebSocket):** Once credentials are acquired, the client establishes a single, persistent WebSocket connection. All subsequent real-time game actions are sent as JSON messages over this connection.

The actions listed below are only those sent over the **WebSocket**.


## I. Client → Server Actions

These are the commands a client can send to the server via WebSocket. The server will validate each action and, if valid, generate one or more corresponding events.

**Note:** Game creation, lobby joining, and spectating are handled via REST endpoints:
- `POST /api/games` - Create a new game lobby
- `POST /api/games/{id}/join` - Join an existing lobby
- `POST /api/games/{id}/spectate` - Join a running game as a spectator

| Action Name | Payload | Description |
| :--- | :--- | :--- |
| **`RECONNECT`** | `{ "game_id": string, "player_id": string, "session_token": string }` | Sent immediately upon connection to rejoin an active game. The server will respond with a `GAME_STATE_SNAPSHOT` to bring the client up-to-date instantly. |
| **`START_GAME`** | `{}` | Sent by the lobby host to begin the game, assigning roles and starting Day 1. |
| **`POST_CHAT_MESSAGE`**| `{ "content": string }` | Sends a single chat message to be broadcast to other players. <br> **Quote Format:** Use BBCode-style `[quote=player_name]original message[/quote]` to reply to specific messages. |
| **`UPDATE_STATUS`**| `{ "status": string }` | Updates the player's public Player Status message (max 20 chars). |
| **`SUBMIT_NIGHT_ACTION`**| `{ "type": string, "data": object }` | Submits the player's choice for the night. The `data` payload is specific to the action `type`. <br> **Examples:** <br> `MINE`: `{ "target_player_id": "p-xyz" }` <br> `REALLOCATE_BUDGET`: `{ "source_player_id": "p-abc", "destination_player_id": "p-def" }` |
| **`SUBMIT_VOTE`** | `{ "vote_target_id"?: string, "verdict"?: string }` | Casts a vote. During nomination, `vote_target_id` is used. During the verdict, `verdict` (`YES` or `NO`) is used. |
| **`SUBMIT_PULSE_CHECK`**| `{ "response": string }` | Submits the player's one-sentence response to the daily Pulse Check prompt. |
| **`SUBMIT_EXIT_INTERVIEW`**| `{ "action": string, "target_player_id"?: string, "final_status": string }` | Sent by a just-deactivated player. `action` can be `HANDOFF`, `CONFIDENTIAL_FEEDBACK`, or `BURN_BRIDGES`. |
| **`POST_SPECTATOR_MESSAGE`** | `{ "content": string }` | Sends a chat message to the `#spectators` channel. Only available to spectators. |

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
| **`SPECTATOR_STATE_SNAPSHOT`** | `{ "game_state": PublicGameState, "spectators": Spectator[] }` | **Sent privately** to a new spectator. Contains a filtered, public-only view of the game state and the current list of spectators. |
| **`SPECTATOR_JOINED`** | `{ "spectator": SpectatorObject }` | A new spectator has joined. Broadcast only to other spectators. |
| **`SPECTATOR_LEFT`** | `{ "player_id": string }` | A spectator has left. Broadcast only to other spectators. |
| **`SPECTATOR_CHAT_MESSAGE`**| `{ "message": ChatMessageObject }` | A new chat message for the `#spectators` channel. Broadcast only to spectators. |

---

## III. Single Authoritative Event Principle

**As of ADR-006**, the codebase follows the **"Single Authoritative Event"** principle:

> For any given player action, the server generates exactly one event that fully describes the resulting state change. The client applies this event to its local state and re-renders.

This eliminates race conditions and ensures deterministic client behavior.

### Voting Events

| Event Type | Payload | Description |
| :--- | :--- | :--- |
| **`VOTE_STARTED`** | `{ "vote_type": string, "phase": string }` | A new voting session has begun. |
| **`VOTE_TALLY_UPDATED`** | **Authoritative voting event** | Contains complete voting state including vote tallies, token weights, and voter information (when transparency mandate is active). Replaces the deprecated dual-event pattern of `VOTE_CAST` + separate tally updates. |

**VOTE_TALLY_UPDATED Payload:**
```json
{
  "vote_type": "NOMINATION|VERDICT|EXTENSION",
  "results": { "player_id": vote_count },
  "token_weights": { "voter_id": token_count },
  "is_complete": boolean,
  "voter_id": "string",        // Player who cast this vote
  "target_id": "string",       // Target of this vote
  "public_voting": boolean,    // True if transparency mandate active
  "voter_choices": {           // Only included if public_voting: true
    "voter_id": "target_id"
  }
}
```

### Night Action Resolution

| Event Type | Payload | Description |
| :--- | :--- | :--- |
| **`NIGHT_ACTIONS_RESOLVED`** | **Comprehensive night resolution** | Single authoritative event containing all night action outcomes in structured format. Replaces individual granular events. |

**NIGHT_ACTIONS_RESOLVED Payload:**
```json
{
  "night_number": number,
  "total_actions": number,
  "blocked_players": [
    {
      "player_id": "string",
      "player_name": "string", 
      "blocker_id": "string",
      "blocker_name": "string",
      "block_type": "BLOCK|ISOLATE_NODE"
    }
  ],
  "converted_players": [
    {
      "player_id": "string",
      "player_name": "string",
      "converter_id": "string", 
      "converter_name": "string",
      "previous_equity": number,
      "new_equity": number
    }
  ],
  "mining_results": [
    {
      "miner_id": "string",
      "miner_name": "string",
      "target_id": "string", 
      "target_name": "string",
      "tokens_mined": number,
      "success": boolean
    }
  ],
  "player_state_changes": {
    "player_id": {
      "tokens_gained": number,
      "status_message": "string",
      "alignment": "HUMAN|ALIGNED", 
      "ai_equity": number,
      "project_milestones": number,
      "has_used_ability": boolean,
      "role_unlocked": boolean
    }
  },
  "summary_message": "string"
}
```

### Specific Semantic Events

These events replace the deprecated generic `SYSTEM_MESSAGE` pattern:

| Event Type | Payload | Description |
| :--- | :--- | :--- |
| **`CLIENT_ERROR`** | `{ "error_code": string, "message": string, "retry_allowed": boolean }` | Client-side error notifications with structured error handling. |
| **`LIAISON_PROTOCOL_ACTIVATED`** | `{ "ai_percentage": number, "trigger_threshold": number, "mining_bonus_slots": number, "protocol_duration": string }` | LIAISON Protocol activation with specific trigger information. |
| **`LIAISON_INTEL_REVEALED`** | `{ "revealed_player_id": string, "revealed_player_name": string, "revealed_action_type": string, "action_description": string, "night_number": number }` | Intelligence revelation from LIAISON Protocol. |
| **`AI_CONVERSION_BLOCKED`** | `{ "converter_id": string, "target_id": string, "blocking_reason": "CRISIS|MANDATE|PROTECTION", "blocking_source": string, "blocking_details": object }` | AI conversion attempt blocked by game mechanics. |
| **`SITREP_PUBLISHED`** | `{ "day_number": number, "crisis_event": object, "sitrep_content": string, "affected_mechanics": string[], "redacted_sections": string[] }` | Daily situation report with crisis information. |
| **`GAME_RULE_MODIFIED`** | `{ "rule_category": string, "modification_type": string, "source": string, "source_name": string, "affected_phases": string[], "duration": string, "rule_text": string }` | Dynamic game rule changes from crisis events or mandates. |

### Deprecated Events

⚠️ **These events are deprecated and should not be used in new development:**

- `SYSTEM_MESSAGE` - Replaced by specific semantic events above
- `VOTE_CAST` - Replaced by `VOTE_TALLY_UPDATED`
- Individual night action events (`PLAYER_BLOCKED`, `MINING_SUCCESSFUL`, etc.) - Replaced by `NIGHT_ACTIONS_RESOLVED`

---