# User Action Graph & UX Flow

This document provides a definitive map of every possible user action at each point in the `Alignment` application's user experience (UX). It serves as a single source of truth for developers, QA testers, and designers to understand the complete scope of user interaction, ensuring feature completeness and consistency.

This graph bridges the [Game Design Document](./01-game-design-document.md) with the [API & Communication Protocols](./api/README.md).

---

## 1. High-Level User Flow

This diagram illustrates the main states a user transitions through from launching the application to finishing a game.

```mermaid
graph TD
    subgraph "Pre-Game & Session"
        A[/"Login Screen"/] -->|Enters Handle| B;
        A -->|Session Token Found| E;
        B["Lobby List Screen"] -->|Clicks 'Create' or 'Join'| C;
        B -->|Clicks 'Logout'| A;
        B -->|Clicks 'Spectate'| E_Spec["In-Game (Spectator)"];
    end

    subgraph "Lobby"
        C["Waiting Screen"] -->|Host Clicks 'Start Game'| D;
        C -->|Leaves or Disconnects| B;
    end

    subgraph "Game"
        D["Role Reveal Screen"] -->|Clicks 'Enter War Room'| E;
        E{In-Game<br>(Player)} -->|Game Ends| F;
        E -->|Abandons Game| B;
    end

    subgraph "Spectate"
      E_Spec -->|Clicks 'Leave'| B;
      E_Spec -->|Game Ends| F;
    end

    subgraph "Post-Game"
        F["Game Over Screen"] -->|Clicks 'View Analysis'| G;
        F -->|Clicks 'Play Again'| B;
        G["Post-Game Analysis"] -->|Clicks 'Back to Results'| F;
        G -->|Clicks 'Play Again'| B;
    end

    style A fill:#f9f,stroke:#333,stroke-width:2px
    style B fill:#f9f,stroke:#333,stroke-width:2px
    style C fill:#f9f,stroke:#333,stroke-width:2px
    style D fill:#f9f,stroke:#333,stroke-width:2px
    style E fill:#ccf,stroke:#333,stroke-width:2px
    style E_Spec fill:#ffc,stroke:#333,stroke-width:2px
    style F fill:#9f9,stroke:#333,stroke-width:2px
    style G fill:#9f9,stroke:#333,stroke-width:2px
```

---

## 2. Detailed Action Breakdown

### A. Pre-Game & Session Management

#### **A.1. Application Load & Login Screen (`/login`)**

*   **State Description:** The initial entry point. Checks for an existing session before showing the login form.
*   **System Actions (on load):**

| Action | User Trigger | Preconditions | Client Effect | Server Action (API) |
| :--- | :--- | :--- | :--- | :--- |
| **Check for Session** | Page load. | A valid `session_token` exists in `localStorage`. | Shows a "Reconnecting..." screen. | Sends a `RECONNECT` request to the server with the token. |
| **Reconnect Success** | Server responds positively to `RECONNECT`. | Session token is valid and game is active. | Client receives `GAME_STATE_SNAPSHOT`, skips login/lobby, and navigates directly to `/game` or `/waiting`. | Server validates token and sends back the current game state. |
| **Reconnect Failure** | Server responds negatively to `RECONNECT`.| Token is invalid or game has ended. | Clears bad token. Shows standard login form. | Server invalidates the token. |

*   **User Actions:**

| Action | User Trigger | Preconditions | Client Effect | Server Action (API) |
| :--- | :--- | :--- | :--- | :--- |
| **Change Avatar** | Clicks on an avatar emoji. | None. | Highlights the selected avatar. | None. |
| **Enter Handle**| Types in the text input. | None. | The input field updates. "Browse Lobbies" button becomes enabled. | None. |
| **Log In** | Clicks "Browse Lobbies". | Handle is non-empty. No valid session found. | Stores `playerName` & `playerAvatar` in app state. Navigates to `/lobby-list`. | None. |

#### **A.2. Lobby List Screen (`/lobby-list`)**

*   **State Description:** The main hub for finding or creating games.
*   **Available Actions:**

| Action | User Trigger | Preconditions | Client Effect | Server Action (API) |
| :--- | :--- | :--- | :--- | :--- |
| **Create New Game** | Clicks "+ Create New Game". | Player has a handle. | Shows a loading state. On success, stores `game_id`, `player_id`, `session_token` and navigates to `/waiting`. | `POST /api/games`. |
| **Join Existing Lobby**| Clicks "Join" on a lobby. | Lobby not full/in-progress. | Shows loading state. On success, stores credentials and navigates to `/waiting`. | `POST /api/games/{id}/join`. |
| **Spectate Game**| Clicks "Spectate" on a running game. | Game is `IN_PROGRESS`. | Shows loading state. On success, stores credentials and navigates to `/game` in spectator view. | `POST /api/games/{id}/spectate` |
| **Logout** | Clicks the "Logout" button. | None. | Clears all session data and navigates to `/login`. | None. |

#### **A.3. Waiting Screen (Lobby) (`/waiting`)**

*   **State Description:** Pre-game staging area. All actions from here are over a persistent WebSocket.
*   **User Actions:**

| Action | User Trigger | Preconditions | Client Effect | Server Action (WebSocket) |
| :--- | :--- | :--- | :--- | :--- |
| **Start Game** | Clicks "[> INITIATE...]" button. | User is the host. Min players present. | A countdown timer appears for all players. | `START_GAME` action. |
| **Leave Lobby** | Clicks "← Leave Lobby". | None. | Disconnects WebSocket, clears session, navigates to `/lobby-list`. | `PlayerActor` disconnect triggers leave logic. |
| **Kick Player** | Clicks "Kick" icon on a player. | User is the host. Target is not the host. | The targeted player is immediately removed from the lobby UI for all clients. The kicked player sees an alert and is returned to the `/lobby-list`. | Sends `KICK_PLAYER` with `target_player_id`. Server validates host status and removes player. |
| **Transfer Host** | Clicks "Make Host" icon on a player. | User is the host. | The host "crown" icon moves to the new host for all players. | Sends `TRANSFER_HOST`. |
| **Set Lobby Privacy** | Clicks a "Public/Private" toggle. | User is the host. | The lobby's status icon changes. | Sends `SET_LOBBY_PRIVACY`. The lobby is removed/added from the public list. |
| **Copy Invite Link** | Clicks "Invite" button. | None. | A unique join URL (e.g., `alignment.gg/join/{id}`) is copied to the clipboard. | None. |

---

### B. In-Game Flow (`/game`)

#### **B.1. General Actions (Available in Most Phases)**

| Action | User Trigger | Preconditions | Client Effect | Server Action (WebSocket) |
| :--- | :--- | :--- | :--- | :--- |
| **Send Chat Message** | Types in chat and presses `Enter`. | Phase allows chat. | Optimistically renders message. | `POST_CHAT_MESSAGE`. |
| **Whisper** | Clicks "Whisper" icon on a player card, types message, sends. | Once per day phase. Target is alive. | A public message appears: `Vex whispers to Astra`. The whispered message is sent privately to the target. | Sends `WHISPER` action with `target_id` and `message`. |
| **Open Settings** | Clicks the gear icon. | None. | The settings modal appears. | None. |
| **Abandon Game** | Clicks "Abandon Game" in settings menu, confirms. | Game is in progress. | Disconnects, clears session, navigates to `/lobby-list`. | Disconnect triggers `PLAYER_LEFT`. Player is marked as "Deactivated (Abandoned)". |
| **Vote to Surrender**| Clicks "Surrender" in settings, confirms. | User is Human. At least one other Human is alive. | A poll appears for all Human players. If all agree, the game ends. | Sends `INITIATE_SURRENDER_VOTE`. |
| **Use `/help`** | Types `/help` in chat. | None. | The `Loebmate` bot sends a private message to the user with contextual help for the current phase. | `POST_CHAT_MESSAGE`. The server intercepts the command and sends a private event back. |

#### **B.2. Day Phase: Voting Cycle**

| Action | User Trigger | Preconditions | Client Effect | Server Action (WebSocket) |
| :--- | :--- | :--- | :--- | :--- |
| **Nominate Player** | Clicks "Nominate" during `NOMINATION`. | Has not already voted. | UI shows vote cast. Tally updates. | `SUBMIT_VOTE` with `vote_target_id`. |
| **Vote Guilty/Yes** | Clicks "✔️ YES" during `VERDICT`. | Has not already voted. | UI shows vote cast. Tally updates. | `SUBMIT_VOTE` with `verdict: "GUILTY"`. |
| **Vote Innocent/No**| Clicks "❌ NO" during `VERDICT`. | Has not already voted. | UI shows vote cast. Tally updates. | `SUBMIT_VOTE` with `verdict: "INNOCENT"`. |
| **Abstain** | Clicks "Abstain" button during a vote. | Has not already voted. | UI shows "You have abstained." The server may show this publicly. | `SUBMIT_VOTE` with `verdict: "ABSTAIN"`. |
| **Retract Vote** | Clicks "Retract Vote" button. | Has already voted in the current phase. Timer has not expired. | UI reverts to the pre-vote state, re-enabling voting buttons. A public log may appear: `Vex has retracted their vote`. | `RETRACT_VOTE`. |

#### **B.3. Night Phase (30s)**

| Action | User Trigger | Preconditions | Client Effect | Server Action (WebSocket) |
| :--- | :--- | :--- | :--- | :--- |
| **Pre-Submit Action**| Interacts with the Night Action panel *during the Day Phase*. | None. | The UI allows selection and queueing of a night action. The choice is saved locally. | None. |
| **Confirm Night Action**| Clicks "Lock In Action" button. | Night phase is active. | UI confirms the action is locked in for the night. | `SUBMIT_NIGHT_ACTION` with the chosen action type and target. |

### B.4. Spectator Mode Actions

*   **State Description:** Read-only view of an in-progress game. Spectators cannot affect the game but can observe and chat with other spectators.
*   **Available Actions:**

| Action | User Trigger | Preconditions | Client Effect | Server Action |
| :--- | :--- | :--- | :--- | :--- |
| **Send Spectator Chat**| Types in `#spectators` chat and sends. | Is a spectator. | Optimistically renders message in `#spectators` channel. | `POST_SPECTATOR_MESSAGE` (WebSocket) |
| **Leave Game**| Clicks "Leave Game". | Is a spectator. | Disconnects WebSocket, clears session, navigates to `/lobby-list`. | Disconnect triggers removal from spectator list. |

---

### C. Post-Game Flow

#### **C.1. Game Over Screen (`/game-over`)**

| Action | User Trigger | Preconditions | Client Effect | Server Action |
| :--- | :--- | :--- | :--- | :--- |
| **Give Kudos** | Clicks a "Kudos" icon next to a player's name. | Can only give one Kudos per game. Cannot give to self. | UI shows confirmation. | `GIVE_KUDOS` action. |
| **Report Player** | Clicks "Report" on a player, fills out a form. | None. | A confirmation modal appears. | `REPORT_PLAYER` action. |
| **Block Player** | Clicks "Block" on a player, confirms. | None. | Confirmation modal appears. | `BLOCK_PLAYER` action. |
| **View Analysis** | Clicks "VIEW ANALYSIS". | Game is over. | Navigates to `/analysis`. | None. |
| **Play Again** | Clicks "PLAY AGAIN". | Game is over. | Clears session, navigates to `/lobby-list`. | None. |