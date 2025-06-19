# Architecture: Night Phase Resolution

This document details the logic for resolving the Night Phase. Because all players submit their night actions "simultaneously," we need a strict, deterministic order of operations to resolve their effects fairly and prevent race conditions. This process is one of the most complex pieces of business logic in the game.

## 1. The Principle of Precedence

The resolution logic is structured as a series of **precedence passes**. We process actions in order of their ability to influence or cancel other actions. Actions with higher precedence are resolved first, and their effects are factored into the resolution of subsequent passes.

The order of operations is as follows:

1.  **Pass 1: Blocking Actions**
2.  **Pass 2: AI Conversion**
3.  **Pass 3: Standard Actions**

## 2. The Resolution Order

#### Pass 1: Blocking Actions

*   **What:** Actions that explicitly prevent another player from acting are resolved first. This includes role-specific abilities like the CISO's `Isolate Node` or the CEO's `Performance Review`.
*   **Why:** These actions must be resolved first because their entire purpose is to nullify the actions of other players. We must know who is blocked before we can determine which other actions succeed.
*   **Outcome:** A `Set` of `player_ids` who have been blocked for the night is created.

#### Pass 2: AI Conversion

*   **What:** The AI faction's attempt to convert a human player is resolved next.
*   **Why:** The AI's targeting action also functions as a block, preventing the target from performing their own action, regardless of whether the conversion is successful. It has a lower precedence than an explicit block from a role like the CISO.
*   **Outcome:**
    *   The target player's ID is added to the "blocked" set.
    *   The system checks if the AI's hidden `AI Equity` score for the target has surpassed the target's `Tokens`.
    *   The outcome (successful conversion or a `SystemShock`) is determined.

#### Pass 3: Standard Actions

*   **What:** All remaining actions are now resolved, but only for players whose IDs are **not** in the "blocked" set. This includes actions like `Mine for Tokens` and `Project Milestones`.
*   **Why:** This pass resolves all actions that do not interfere with others.
*   **Outcome:**
    *   For `Mine for Tokens` actions, the system checks against the available "Liquidity Pool" slots for the night to determine success or failure.
    *   Other actions, like `Project Milestones`, are marked as successful.

## 3. Finalization

Once all passes are complete, the system aggregates all the outcomes into a single, large `NIGHT_ACTIONS_RESOLVED` event. This event contains the full summary of what happened (who was blocked, who was converted, who successfully mined, etc.) and is broadcast to all players, triggering the start of the next Day Phase.
