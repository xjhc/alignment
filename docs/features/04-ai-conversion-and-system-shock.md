
# Feature: AI Conversion & System Shock

This document describes the AI faction's primary win condition path: converting humans to their cause. It also details the **System Shock** mechanic, which serves as a defense and an information-gathering tool for humans.

## 1. Feature Overview

*   **AI Conversion:** Each night, the Original AI (and any Aligned humans) can target one human player. A successful conversion brings that human into the AI faction.
*   **System Shock:** If a conversion attempt fails because the target has too much influence (Tokens), the target is afflicted with a temporary, negative status effect, but also gains proof that they were targeted by the AI.

## 2. System Flow

1.  **AI Action:** During the Night Phase, the AI player submits its target.
    *   **Action:** The AI (via its Rules Engine) sends `SUBMIT_NIGHT_ACTION` with `payload: { "type": "CONVERT", "target_player_id": "p-dani" }`.
2.  **Server-Side Resolution:** This action is resolved during the `Night Phase Resolution` with high precedence.
    *   **Blocking:** The target player is automatically blocked from performing their own night action.
    *   **Conversion Check:** The server performs the core conversion logic:
        1.  Each human player has a hidden `AI Equity` score tracked by the server-side `Game Actor`.
        2.  Targeting a player increases their `AI Equity`.
        3.  The server compares the target's updated `AI Equity` score against their current `Tokens` count.
        4.  **`IF AI_Equity > Tokens`**: The conversion is **successful**.
        5.  **`ELSE`**: The conversion **fails**.
3.  **Event & Notification:** The outcome is included in the `AITargetResult` object within the `NIGHT_ACTIONS_RESOLVED` event.
    *   **On Success (`WasConverted: true`):**
        *   The converted player's `Alignment` is secretly changed to `ALIGNED` on the server.
        *   The server adds the player to the secret `#aligned` chat channel (a server-side concept).
        *   The player receives a **private** `ALIGNMENT_CHANGED` event, informing them of their new alignment and giving them access to the AI faction's tools.
    *   **On Failure (`HadSystemShock: true`):**
        *   The target player remains human.
        *   The server sends a **private** `PRIVATE_NOTIFICATION` event to the target with a specific type and message, e.g., `payload: { "type": "SYSTEM_SHOCK_AFFLICTED", "message": "[GPT-4o ALERT] Unauthorized intrusion repelled. WARNING: Minor systemic shock detected." }`.
        *   The server assigns a random `SystemShock` effect (e.g., `MessageCorruption`) to the player for the next 24 hours. The logic for this effect (e.g., a 25% chance to corrupt messages) is handled by the `Game Actor` when processing that player's future actions.

## 3. Key Implementation Details

*   **Hidden State:** `AI Equity` is a server-only value. It is never broadcast to any client, not even the AI's.
*   **Dilemma:** The System Shock mechanic creates a strategic dilemma. A player can reveal they have a shock to prove they are likely human, but this also confirms they have low tokens, making them a prime target for future conversion or deactivation.
