# Design: State-Driven UI Map

## 1. Philosophy: The UI is a Function of State

This document is the definitive "if-this-then-that" map for the entire `Alignment` user interface. It is built on a core principle: **the UI is a pure function of the `GameState`**.

**`UI = f(GameState)`**

There should be no complex, independent state within UI components. Instead, every component reads from the centralized state provided by `App.tsx` and the Go/Wasm engine, and renders its appearance based on that data. This approach dramatically reduces bugs, simplifies testing, and ensures a consistent and predictable user experience.

When building any new UI component, refer to this map to determine how it should behave under different game conditions.

---

## 2. Global UI States

These states affect the entire application view, regardless of the specific game phase.

| IF (State Condition) | THEN (UI Outcome) | Component(s) Affected |
| :--- | :--- | :--- |
| `isConnecting == true` | Show a full-screen overlay with a "Connecting..." spinner. | `App.tsx` |
| `isReconnecting == true` | Show a non-blocking toast or banner with "Reconnecting...". | `App.tsx` |
| `connectionError != null` | Show a modal or full-screen error with a "Retry" or "Back to Login" button. | `App.tsx` |
| `isLoadingWasm == true` | Show a full-screen "Loading Game Engine..." view. | `App.tsx` |

---

## 3. Screen-Level State Mapping

This section defines which primary screen is rendered based on the overall session and game state.

| IF (State Condition) | THEN (UI Outcome) | Component(s) Affected |
| :--- | :--- | :--- |
| `sessionState == 'IDLE'` AND `appState.playerName` is empty | Show the initial login/identity creation view. | `LoginScreen` |
| `sessionState == 'IDLE'` AND `appState.playerName` exists | Show the list of available game lobbies. | `LobbyListScreen` |
| `sessionState == 'IN_LOBBY'` | Show the pre-game waiting room. | `WaitingScreen` |
| `gameState.phase.type == 'ROLE_REVEAL'` | Show the private role and alignment reveal. | `RoleRevealScreen` |
| `sessionState == 'IN_GAME'` AND `localPlayer.isAlive == true` | Show the main game interface. | `GameScreen` |
| `gameState.winCondition` exists | Show the victory/defeat summary. | `GameOverScreen` |
| `localPlayer.isAlive == false` AND `winCondition` does not exist| Render a disabled "spectator" view of the main game. | `GameScreen` (with overlays/disabled inputs) |

---

## 4. Phase-Based UI States (within `GameScreen`)

This section defines the behavior of the main `GameScreen` panels based on the current phase.

| IF (`gameState.phase.type` is...) | THEN (UI Outcome) | Component(s) Affected |
| :--- | :--- | :--- |
| **`SITREP`** | `ChatInput` is enabled. `ContextualInputArea` is hidden. | `ChatInput` |
| **`PULSE_CHECK`** | `ContextualInputArea` renders `PulseCheckInput`. `ChatInput` is disabled until local player has submitted. | `ContextualInputArea`, `ChatInput` |
| **`DISCUSSION`** | `ChatInput` is enabled for all living players. `ContextualInputArea` is hidden. With 15s remaining, a banner with "Extend" / "Nominate" buttons appears. | `ChatInput`, `ContextualInputArea` |
| **`NOMINATION`** | `ChatInput` is enabled. `ContextualInputArea` renders `VoteUI` for player nomination. | `ContextualInputArea`, `ChatInput` |
| **`TRIAL`** | `ChatInput` is enabled for all living players. | `ChatInput` |
| **`VERDICT`** | `ChatInput` is **enabled**. `ContextualInputArea` renders `VoteUI` with "GUILTY" and "INNOCENT" options. | `ContextualInputArea`, `ChatInput` |
| **`NIGHT`** | `ChatInput` is disabled. `ContextualInputArea` renders `NightActionSelection`. The screen has a "night mode" blue tint. | `GameScreen`, `ContextualInputArea`, `ChatInput`|

---

## 5. Component-Level State Mapping

This section defines the appearance and behavior of individual UI components based on specific state fields.

| Component | IF (State Condition) | THEN (UI Outcome) |
| :--- | :--- | :--- |
| **`ChatMessage`** | `message.authorID == gameState.nominatedPlayer` AND `gameState.phase.type == 'TRIAL'` | The message block has a special highlighted border or background to stand out. |
| **`PlayerCard`** | `player.id == localPlayer.id` | The player card is highlighted and displays "(You)". |
| | `player.isAlive == false` | Card is grayscale and displays a 👻 avatar. |
| | `player.systemShocks.length > 0` | A subtle "glitch" animation is applied to the player's name/avatar. |
| | `player.isRolePubliclyRevealed == true` | The card displays the player's true Role (e.g., "CISO") instead of their Job Title. |
| | `gameState.nominatedPlayer == player.id`| The card has a special "On Trial" border/highlight. |
| **`PlayerHUD`** | `viewedPlayer.id != localPlayer.id` | The panel shows the public "Player Dossier" view. |
| | `viewedPlayer.id == localPlayer.id` | The panel shows the private "My Terminal" view, revealing `alignment` and `personalKPI`. |
| **`VoteUI`** | `voteState.votes[localPlayer.id]` exists | The voting buttons are disabled, and a "Retract Vote" button appears. |
| | `corporateMandate.type == 'TOTAL_TRANSPARENCY'` | The vote tally shows a breakdown of who voted for whom, in real-time. |
| **Chat Channels** | `localPlayer.alignment` is `ALIGNED` or `AI` | The `#aligned` channel becomes visible and selectable in the `RosterPanel`. |
| | `localPlayer.isAlive == false` | The `#off-boarding` channel becomes visible and selectable in the `RosterPanel`. |