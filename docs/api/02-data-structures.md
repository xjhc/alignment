# Core Data Structures

This document defines the primary Go structs that are serialized and sent over the WebSocket. These objects form the building blocks of the `GameState` and event payloads.

*(Note: JSON field names are `snake_case` by convention.)*

---

### 1. `GameState` (Client-Side Read Model)

This is the "read model" the client builds by applying events. The UI renders directly from this object. It is never sent in its entirety except during a full state sync for a new connection.

```go
type GameState struct {
    GameID          string                `json:"game_id"`
    LocalPlayerID   string                `json:"local_player_id"`
    CurrentPhase    string                `json:"current_phase"` // LOBBY, DAY, NIGHT, etc.
    DayNumber       int                   `json:"day_number"`
    Players         map[string]Player     `json:"players"`       // Keyed by PlayerID
    ChatMessages    []ChatMessage         `json:"chat_messages"`
    CrisisEvent     CrisisEvent           `json:"crisis_event"`
    // ... other UI-relevant state like current vote info, timers, etc.
}
```

---

### 2. Core Objects

These are the most common objects used in payloads.

**`Player` Object**
```go
type Player struct {
    ID                string       `json:"id"`
    Name              string       `json:"name"`
    JobTitle          string       `json:"jobTitle"`
    ControlType       string       `json:"controlType"` // "HUMAN" or "AI"
    Status            PlayerStatus `json:"status"`
    IsAlive           bool         `json:"isAlive"`
    Tokens            int          `json:"tokens"`
    ProjectMilestones int          `json:"projectMilestones"`
    StatusMessage     string       `json:"statusMessage"`
    JoinedAt          time.Time    `json:"joinedAt"`

    // Private fields (only visible to the player themselves)
    Alignment              string       `json:"alignment,omitempty"` // "HUMAN" or "AI" or "ALIGNED"
    Role                   *Role        `json:"role,omitempty"`
    PersonalKPI            *PersonalKPI `json:"personalKPI,omitempty"`
    AIEquity               int          `json:"aiEquity,omitempty"` // For alignment conversion
    HasUsedAbility         bool         `json:"hasUsedAbility,omitempty"`
    LastNightAction        *NightAction `json:"lastNightAction,omitempty"`
    HasSubmittedPulseCheck bool         `json:"hasSubmittedPulseCheck,omitempty"`
    LobbyHandle            string       `json:"lobbyHandle,omitempty"` // Original lobby identity for post-game reveal
    BootcampPoints         int          `json:"bootcampPoints,omitempty"` // Intern role resource for shadowing abilities

    // FTUE and Assistance Settings
    SeenHints            map[string]bool `json:"seenHints,omitempty"`             // Phase hints the player has seen
    DisableLoebmateHints bool            `json:"disableLoebmateHints,omitempty"`  // Whether to disable Loebmate assistance

    // Public status and effects
    SlackStatus            string        `json:"slackStatus,omitempty"`
    PartingShot            string        `json:"partingShot,omitempty"`
    SystemShocks           []SystemShock `json:"systemShocks,omitempty"`
}
```

**`ChatMessage` Object**
```go
type ChatMessage struct {
    ID         string    `json:"id"`
    AuthorID   string    `json:"author_id"`
    AuthorName string    `json:"author_name"` // Denormalized for easy display
    Content    string    `json:"content"`
    Timestamp  time.Time `json:"timestamp"`
}
```

**`CrisisEvent` Object**
```go
type CrisisEvent struct {
    Title        string `json:"title"`
    Effect       string `json:"effect"`
    PulseCheckPrompt string `json:"pulse_check_prompt"`
}
```

**`RoleInfo` Object** (Payload for the `ROLE_ASSIGNED` event)
```go
type RoleInfo struct {
    Role        string `json:"role"`
    Alignment   string `json:"alignment"`
    PersonalKPI string `json:"personal_kpi"`
}
```

**`AlignmentChangedPayload` Object** (Payload for the `ALIGNMENT_CHANGED` event)
```go
type AlignmentChangedPayload struct {
    NewAlignment string `json:"new_alignment"` // e.g., "ALIGNED"
}
```

---

### 3. Event-Specific Payloads

This is the payload for the `NIGHT_ACTIONS_RESOLVED` event, summarizing the night's outcomes.

```go
type NightResultsObject struct {
    SuccessfulMines   map[string]string `json:"successful_mines"`   // MinerID -> TargetID
    FailedMineCount   int               `json:"failed_mine_count"`
    BlockResult       *BlockInfo        `json:"block_result,omitempty"` // Info about who was blocked
    AITargetResult    *AITargetInfo     `json:"ai_target_result,omitempty"` // Info about the AI's action
    ProgressUpdates   []string          `json:"progress_updates"`   // List of PlayerIDs who progressed
    // ... any other role-specific results like audit results, etc.
}

type BlockInfo struct {
    BlockerID string `json:"blocker_id"`
    TargetID  string `json:"target_id"`
}

type AITargetInfo struct {
    TargetID          string `json:"target_id"`
    WasConverted      bool   `json:"was_converted"`
    HadSystemShock    bool   `json:"had_system_shock"`
}
```