# Feature: Spectator Mode

This document describes the implementation design for the Spectator Mode, allowing users to watch live games.

## 1. Feature Overview

Spectator Mode provides a read-only view of an in-progress game. Spectators can chat in a dedicated channel but cannot interact with the game itself. This feature is crucial for community building, learning the game, and streaming.

## 2. Architectural Approach

The system will extend the existing **Player-Centric Actor Model**. Each spectator will have their own `PlayerActor` and a persistent WebSocket connection, just like a regular player.

*   **Spectator State:** The `PlayerActor`'s state machine will be extended with a new state: `SPECTATING`.
*   **Game Actor Awareness:** The `GameActor` will maintain a separate list of spectator `PlayerActor`s.
*   **Information Filtering:** The server will generate a special, heavily filtered `PublicGameState` snapshot for spectators, removing all secret information.

## 3. System Flow

1.  **Discovery (REST):**
    *   The client's `LobbyListScreen` fetches `GET /api/games`. The server's response now includes games with `status: "IN_PROGRESS"`.
    *   The UI displays a "Spectate" button for these games.
2.  **Joining (REST):**
    *   User clicks "Spectate". Client calls `POST /api/games/{id}/spectate`.
    *   The server validates that the game exists and is running, then generates a temporary, spectator-scoped `session_token` and a unique spectator ID (e.g., `spectator-xyz`).
3.  **Connection & State Sync (WebSocket):**
    *   The client receives the credentials and establishes a WebSocket connection.
    *   The `WebSocketManager` identifies this as a spectator session and creates a `PlayerActor` in the `SPECTATING` state.
    *   The `GameLifecycleManager` routes this `PlayerActor` to the correct `GameActor`.
    *   The `GameActor` adds the actor to its internal list of spectators.
    *   The `GameActor` sends a private `SPECTATOR_STATE_SNAPSHOT` to the new spectator. This contains the public game state and a list of current spectators.
    *   The `GameActor` broadcasts a `SPECTATOR_JOINED` event to all other spectators.

## 4. Key Implementation Details

*   **`GameActor` Modifications:**
    *   Will now have two maps of actors: `players` and `spectators`.
    *   Broadcasting logic will be updated:
        *   Public events (`PHASE_CHANGED`, etc.) are sent to both lists.
        *   Private player-specific events (`ROLE_ASSIGNED`) are sent only to the relevant player.
        *   Spectator-specific events (`SPECTATOR_JOINED`) are sent only to spectators.
*   **Data Filtering:** A new server-side function, `FilterGameStateForSpectator`, will be created. It will be even more strict than the AI filter, ensuring no sensitive data (roles, alignments, KPIs, private chats) is ever sent to a spectator client.
*   **Chat Channels:** The client-side `CommsPanel` will be updated to handle the `#spectators` channel, which will only be visible and accessible if `appState.isSpectating` is true.

## 5. REST API Changes

### New Endpoint

*   **`POST /api/games/{id}/spectate`**
    *   **Request Body:** `{ "spectator_name": string }`
    *   **Response:** `200 OK` with `{ "game_id": string, "player_id": string, "session_token": string }`
    *   **Validation:**
        *   Game must exist and be in `IN_PROGRESS` status
        *   Spectator name must be unique within the game
        *   Game must not exceed maximum spectator limit (10)

### Modified Endpoint

*   **`GET /api/games`** response will now include games with `status: "IN_PROGRESS"` to enable spectator discovery

## 6. WebSocket Protocol Changes

### New Actions (Client → Server)

*   **`POST_SPECTATOR_MESSAGE`**: Send chat message to `#spectators` channel

### New Events (Server → Client)

*   **`SPECTATOR_STATE_SNAPSHOT`**: Private event sent to new spectators with filtered game state
*   **`SPECTATOR_JOINED`**: Broadcast to spectators when a new spectator joins
*   **`SPECTATOR_LEFT`**: Broadcast to spectators when a spectator leaves
*   **`SPECTATOR_CHAT_MESSAGE`**: Chat message in `#spectators` channel

## 7. Data Structures

### Spectator Object

```go
type Spectator struct {
    ID         string    `json:"id"`
    Name       string    `json:"name"`
    JoinedAt   time.Time `json:"joined_at"`
}
```

### PublicGameState

```go
type PublicGameState struct {
    GameID        string                 `json:"game_id"`
    Phase         string                 `json:"phase"`
    DayNumber     int                    `json:"day_number"`
    Players       []PublicPlayerInfo     `json:"players"`
    TokenCounts   map[string]int         `json:"token_counts"`
    PhaseEndTime  time.Time              `json:"phase_end_time"`
    CrisisEvent   *CrisisEvent           `json:"crisis_event,omitempty"`
}

type PublicPlayerInfo struct {
    ID              string `json:"id"`
    Name            string `json:"name"`
    IsActive        bool   `json:"is_active"`
    StatusMessage   string `json:"status_message"`
    TokenCount      int    `json:"token_count"`
    // Note: No role, alignment, or KPI information
}
```

## 8. Frontend Changes

### LobbyListScreen Updates

*   Display "Spectate" button for games with `status: "IN_PROGRESS"`
*   Handle spectator session creation and navigation

### GameScreen Updates

*   Conditional rendering based on `appState.isSpectating`
*   Hide interactive elements (voting, night actions) for spectators
*   Show spectator-specific UI elements (spectator list, leave button)

### CommsPanel Updates

*   Support for `#spectators` channel
*   Channel visibility based on spectator status

## 9. Security Considerations

*   **Information Isolation:** Spectators must never receive any private game information
*   **Action Validation:** Server must reject any game actions from spectator sessions
*   **Rate Limiting:** Spectator chat may need rate limiting to prevent spam
*   **Session Management:** Spectator sessions should have appropriate timeouts

## 10. Testing Strategy

### Unit Tests

*   `FilterGameStateForSpectator` function must strip all sensitive data
*   `GameActor` spectator management methods
*   Spectator-specific event routing

### Integration Tests

*   Spectator join/leave flow
*   Message broadcasting to correct audiences
*   Data filtering validation

### E2E Tests

*   Complete spectator workflow from lobby list to game observation
*   Multiple spectators in same game
*   Spectator interaction with game events

## 11. Performance Considerations

*   **Memory Usage:** Each spectator adds minimal overhead (filtered state snapshots)
*   **Network Traffic:** Spectators receive fewer events than players (no private events)
*   **CPU Usage:** Additional filtering operations per spectator connection

## 12. Future Enhancements

*   **Spectator Analytics:** Track popular games for spectating
*   **Streaming Integration:** API endpoints for streaming platforms
*   **Replay Mode:** Allow spectating of completed games
*   **Spectator Reactions:** Non-disruptive ways for spectators to react to game events