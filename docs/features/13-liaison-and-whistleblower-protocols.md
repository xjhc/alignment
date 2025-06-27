# Feature: LIAISON & Whistleblower Protocols

This document describes the implementation of two advanced game mechanics designed to empower players in specific situations: the **LIAISON Protocol** (a catch-up mechanic for the Human faction) and the **Whistleblower Protocol** (a post-elimination influence system).

## 1. Feature: LIAISON Protocol

*   **Overview:** An automated system that gives the Human faction a boost when they are losing significantly.
*   **Trigger (Server-Side):**
    1.  The LIAISON Protocol check runs automatically at the beginning of every Day Phase (after night actions are resolved, but before the SITREP is generated).
    2.  The `GameActor` calculates the percentage of living players who are in the AI Faction.
    3.  **If `(AI Faction / Total Living Players) >= 0.4` (i.e., 40% or more), the protocol is triggered.**
*   **System Flow:**
    1.  When triggered, the `GameActor` immediately generates a `LIAISON_PROTOCOL_ACTIVATED` event. This event's payload includes the details of the effects.
    2.  **Effect 1 (Information Reveal):** The `GameActor` randomly selects one non-AI player's action from the previous night and includes it in the event payload.
    3.  **Effect 2 (Resource Boost):** The server sets a temporary flag on its internal `GameState` (e.g., `liaison_bonus_active: true`).
    4.  The client receives the `LIAISON_PROTOCOL_ACTIVATED` event and displays a special, high-priority system message in the chat log, announcing the protocol is active and revealing the selected night action.
    5.  During the *next* Night Phase, the server's `MiningManager` checks for the `liaison_bonus_active` flag. If true, it increases the number of available mining slots by 2 for that night only. After the night resolves, the flag is cleared.
*   **API & Payloads:**
    *   **Event:** `LIAISON_PROTOCOL_ACTIVATED`
    *   **Payload:** `{ "ai_percentage": 0.45, "revealed_action": { "player_name": "Alice", "action_description": "Mined for Bob" }, "mining_bonus_slots": 2 }`

## 2. Feature: Whistleblower Protocol

*   **Overview:** A system that allows deactivated players to influence the next day's Crisis Event by voting.
*   **System Flow:**
    1.  **Voting Starts (Server-Side):** At the start of every Night Phase, the `GameActor` generates a new "Whistleblower Poll." It randomly selects three `CrisisEvent` options from the master list of crises that have not yet occurred.
    2.  **Poll Delivery:** The server sends a `WHISTLEBLOWER_POLL_STARTED` event **only to players in the `#off-boarding` channel** (i.e., deactivated players). The payload contains the three crisis options.
    3.  **Voting (Client-Side):** The client UI for deactivated players displays the three crisis options with vote buttons. A player clicks to vote.
        *   **Action:** Client sends `SUBMIT_WHISTLEBLOWER_VOTE`.
    4.  **Tallying (Server-Side):** The `GameActor` receives these votes throughout the Night Phase. At the start of the next Day Phase, it tallies the votes. The crisis with the most votes is selected as the day's event.
    5.  **Selection & Announcement:** The `GameActor` proceeds with the standard Crisis Event flow, but uses the winning crisis from the Whistleblower poll instead of selecting one randomly. The fact that the crisis was chosen by "Consultants" can be mentioned in the SITREP for added flavor.
*   **API & Payloads:**
    *   **Event:** `WHISTLEBLOWER_POLL_STARTED` (private to deactivated players)
    *   **Payload:** `{ "poll_id": "night-3", "options": [CrisisEventObject, CrisisEventObject, CrisisEventObject] }`
    *   **Action:** `SUBMIT_WHISTLEBLOWER_VOTE`
    *   **Payload:** `{ "poll_id": "night-3", "voted_for_crisis_type": "Database Index Corruption" }`