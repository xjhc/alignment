
# Feature: Roles & Abilities

This document describes the implementation of role-specific special abilities, which players must unlock before they can be used.

## 1. Feature Overview

Each player is assigned a unique role (CEO, CTO, etc.) with a powerful ability. To use this ability, they must first unlock it by spending time on "project work" during the Night Phase.

## 2. System Flow

1.  **Unlocking the Ability:**
    *   Players start with their ability locked and a `ProjectMilestones` count of 0.
    *   To make progress, a player must choose the `PROJECT_MILESTONES` night action.
    *   **Action:** `SUBMIT_NIGHT_ACTION` with `payload: { "type": "PROJECT_MILESTONES" }`.
    *   **Logic:** The `Game Actor` receives this action. During `Night Phase Resolution`, if the player was not blocked, their `ProjectMilestones` count is incremented by 1.
    *   When a player's `ProjectMilestones` reaches 3, their ability is permanently unlocked for the rest of the game. The client UI should update to show the ability is now usable.

2.  **Using the Ability:**
    *   Once unlocked, a player can use their ability as a night action.
    *   **Action:** `SUBMIT_NIGHT_ACTION` with a role-specific payload, e.g., `payload: { "type": "ISOLATE_NODE", "target_player_id": "p-sam" }`.
    *   **Logic:** The `Game Actor` resolves this action according to its specific rules during `Night Phase Resolution`.

3.  **Public & Covert Effects:**
    *   Many abilities have two effects, as defined in the GDD.
    *   **Public Effect:** After the night resolves, the `NIGHT_ACTIONS_RESOLVED` event will contain a public summary of the action (e.g., "The CISO isolated a node."). This is visible to all players.
    *   **Covert Effect:** Some abilities reveal secret information only to the AI faction. For example, the VP of Ethics's `Run Audit` ability will send a private message to the `#aligned` channel with the true result. This is handled by the `Game Actor` sending a special `CHAT_MESSAGE_POSTED` event that is only routed to members of the AI faction.

## 3. Key Implementation Details

*   **State Tracking:** The `Player` object in the `GameState` tracks the `ProjectMilestones` count.
*   **Ability Cooldowns:** The GDD does not specify cooldowns, but if they were added, the `Game Actor` would track `last_ability_use_day` on the `Player` object and validate against it.
*   **Action Validation:** The server will reject any attempt to use an ability if the player's `ProjectMilestones` count is less than 3.
