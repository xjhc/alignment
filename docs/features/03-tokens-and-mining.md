
# Feature: Tokens & Mining

This document describes the mechanics for **Tokens**, the game's primary resource representing influence and voting power, and **Mining**, the main action for generating them.

## 1. Feature Overview

*   **Tokens:** Each player has a `Tokens` count. This integer determines their voting weight.
*   **Mining:** The primary way to earn tokens is to `Mine for [Player Name]` during the Night Phase. This action is a selfless act; you cannot mine for yourself.

## 2. System Flow

1.  **Action Submission:** During the Night Phase, a player can choose to mine for another player.
    *   **Action:** The client sends `SUBMIT_NIGHT_ACTION` with `payload: { "type": "MINE", "target_player_id": "p-chris" }`.
2.  **Server-Side Resolution:** All `MINE` actions are resolved during the `Night Phase Resolution` process within the `Game Actor`. This occurs after higher-precedence actions (like blocks) are resolved.
3.  **Liquidity Pool Logic:** The server determines if a `MINE` action is successful based on the "Liquidity Pool."
    *   The number of available mining "slots" for the night is calculated as `floor(Number of living Humans / 2)`. This value can be modified by `Crisis Events`.
    *   The server collects all valid, un-blocked `MINE` actions.
    *   If the number of actions is greater than the available slots, a priority system is used to determine who succeeds:
        1.  Priority is given to players who failed to mine on the previous night.
        2.  Any remaining ties are broken by giving priority to players with the fewest `Tokens`.
4.  **Event Broadcast:** The results are included in the `NIGHT_ACTIONS_RESOLVED` event sent at the start of the next Day Phase.
    *   The `SuccessfulMines` map in the event payload lists which miners successfully generated a token for which targets (`miner_id -> target_id`).
    *   The `FailedMineCount` field informs the game of how many players attempted to mine but failed.
5.  **State Update:** The `Game Actor` (and each client) applies this event to its `GameState`. For each entry in `SuccessfulMines`, the target player's `Tokens` count is incremented by 1.

## 3. Key Implementation Details

*   **Selfless Mining:** The server logic will reject any `MINE` action where `player_id` matches `target_player_id`.
*   **Centralized Calculation:** All calculations for the Liquidity Pool and priority happen exclusively on the server to ensure a fair and consistent outcome.
