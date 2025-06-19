
# Feature: Crisis & Agenda System

This document describes the implementation of the daily **Crisis Events** and **Pulse Check** prompts, which introduce variety and focus to each Day Phase.

## 1. Feature Overview

At the beginning of each Day Phase, two things happen:
1.  A **Crisis Event** is announced, introducing a temporary rule change for the next 24-hour game cycle (one Day/Night phase).
2.  A **Pulse Check Prompt** is posted, requiring each player to submit a one-sentence response to fuel discussion.

This system is managed entirely by the server and announced to clients via events.

## 2. System Flow

1.  **Trigger:** The system is triggered when the `Game Actor` transitions from the `NIGHT` phase to the `DAY` phase.
2.  **Selection:**
    *   The `Game Actor` maintains a list of potential `CrisisEvent` objects. It randomly selects one that has not been used yet in the current game.
    *   *Exception:* If the **Whistleblower Protocol** is active, the `CrisisEvent` is not chosen randomly but is instead determined by the vote of eliminated players from the previous night.
3.  **Announcement Event:** The chosen `CrisisEventObject` (containing its title, effect description, and the Pulse Check prompt) is included in the payload of the `PHASE_CHANGED` event that signals the start of the Day Phase.
4.  **Client-Side:**
    *   The client UI receives the `PHASE_CHANGED` event and displays the Crisis title and effect prominently.
    *   It also displays the Pulse Check prompt and an input box for the player's response.
5.  **Pulse Check Submission:**
    *   **Action:** Players have 30 seconds to submit their response. The client sends a `SUBMIT_PULSE_CHECK` action with the `response` string.
    *   **Logic:** The `Game Actor` collects these responses.
6.  **Public Reveal:** After the 30-second timer expires, the `Game Actor` iterates through all collected responses.
    *   **Event:** For each response, the server broadcasts a `PULSE_CHECK_SUBMITTED` event containing the `player_id` and their `response`.
    *   **Client-Side:** The UI displays all the submitted responses with attribution, kicking off the open discussion period.

## 3. Implementation Details

*   **Crisis Effects:** The logic for enforcing the Crisis Event's rule change resides entirely on the server. The `Game Actor` will check the `currentCrisis` on its `GameState` before performing certain actions. For example, for the "Hostile Takeover Bid" crisis, the vote-tallying logic will check for a 60% threshold instead of the usual 51%.
*   **Data Structure:** The `CrisisEvent` objects will be defined as a static list or loaded from a configuration file at server startup.
