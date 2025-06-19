
# Feature: Personal KPIs & Mandates

This document describes the implementation of game-wide modifiers (**Mandates**) and secret personal objectives (**Personal KPIs**).

## 1. Feature: Corporate Mandates

*   **Overview:** A Mandate is a global rule modification that is chosen at the very start of the game and affects all players.
*   **Implementation Flow:**
    1.  **Trigger:** This logic runs once, when the Host sends the `START_GAME` action.
    2.  **Selection:** The `Game Actor` randomly selects one `Mandate` from a predefined list.
    3.  **State Modification:** The `Game Actor` immediately applies the Mandate's effects. This might involve:
        *   Modifying the initial `GameState` (e.g., "Aggressive Growth Quarter" changes every player's starting `Tokens` to 2).
        *   Setting a game-wide boolean flag on the `GameState` that modifies server logic (e.g., "Total Transparency Initiative" sets a `public_voting` flag to true, which changes how `VOTE_TALLY_UPDATED` events are structured).
    4.  **Announcement:** The chosen Mandate is announced to all players, likely via a special `GAME_STARTED` event payload or an initial `CHAT_MESSAGE_POSTED` from a "System" user.

## 2. Feature: Personal KPIs (Secret Objectives)

*   **Overview:** A Personal KPI is a secret objective given to each human player, offering a bonus or an alternate win condition if completed.
*   **Implementation Flow:**
    1.  **Assignment:** When roles are assigned at the start of the game, the `Game Actor` also randomly assigns a unique `PersonalKPI` to each player with the `HUMAN` alignment.
    2.  **Private Notification:** The text description of the KPI is included in the private `ROLES_ASSIGNED` event sent to each player.
    3.  **Server-Side Tracking:** The `Game Actor` tracks the progress of each player's KPI. This is the most complex part of the implementation.
        *   The server must listen for the specific game events that trigger KPI progress.
        *   **Example for "The Inquisitor":** After each `PLAYER_ELIMINATED` event, the server checks if a player's vote matched the eliminated player. If so, it increments a hidden `correct_votes` counter for that player's KPI.
        *   **Example for "The Scapegoat":** When a player is eliminated, the server checks the `VOTE_TALLY_UPDATED` results. If the `NO` vote count was 0, the Scapegoat's alternate win condition is met.
    4.  **Resolution:** When a KPI's condition is met, the `Game Actor` applies its bonus. This could be a one-time event (e.g., sending a `PRIVATE_NOTIFICATION` with secret info) or a flag that modifies their power in the final vote tally. The logic for alternate win-cons is checked alongside the standard faction win conditions at the end of each phase.

## 3. Key Implementation Details

*   **Data-Driven Design:** Both Mandates and KPIs should be defined iOf course. This is a great set of documents that will fully flesh out the `/docs/features` directory, connecting the GDD directly to engineering implementation.
