# Feature: Player Safety & Moderation

This document outlines the design for the systems that ensure a safe and positive community environment. These features give players control over their experience and provide administrators with the tools to enforce a code of conduct.

## 1. Feature: Player Reporting

*   **Overview:** A system for players to report others for negative behavior like abusive chat, griefing, or spam.
*   **User Flow:**
    1.  A player clicks a "Report" button, accessible from another player's profile (in the lobby or post-game screen) or via a context menu in the chat.
    2.  A modal appears, asking the player to select a reason for the report (e.g., "Abusive Language," "Gameplay Sabotage") and to provide a brief description.
    3.  The client includes a snapshot of the recent chat log with the report for context.
*   **System Flow:**
    1.  The client sends a `REPORT_PLAYER` action to the server.
    2.  The server validates the report and stores it in a dedicated `reports` collection/table in the database, linked to both the reporter and the reported player.
    3.  The report is flagged for review in the **Internal Admin Tool**.
*   **API & Payloads:**
    *   **Action:** `REPORT_PLAYER`
    *   **Payload:** `{ "target_player_id": "p-toxic", "reason": "Abusive Language", "description": "...", "chat_log_snapshot": "[...]" }`

## 2. Feature: Player Blocking

*   **Overview:** Allows a player to prevent future interaction with another player.
*   **User Flow:**
    1.  A player clicks a "Block" button on another player's profile.
    2.  A confirmation modal appears: "Are you sure you want to block this player? You will not be matched with them in future games."
*   **System Flow:**
    1.  The client sends a `BLOCK_PLAYER` action.
    2.  The server adds the target's `player_id` to a `blocked_players` array on the requester's user profile in the database.
    3.  **Effect:** The matchmaking system will now check this list. If a player tries to join a lobby containing someone on their block list (or vice-versa), the join will be rejected.
    *   *(Optional V2 Feature)* The client could also use this list to locally mute the blocked player's chat messages for the remainder of the current game.
*   **API & Payloads:**
    *   **Action:** `BLOCK_PLAYER`
    *   **Payload:** `{ "target_player_id": "p-annoying" }`

## 3. Feature: Reputation System (Internal)

*   **Overview:** A hidden, server-side score for each player that helps identify community members who are consistently positive or negative. This is **not** visible to players.
*   **Implementation:**
    *   Each player profile in the database has a `reputation_score` field, starting at a neutral value.
    *   **Score Decreases:** The score is lowered when credible reports are filed against the player (as verified by an admin). Multiple reports for the same incident have diminishing impact.
    *   **Score Increases:** The score is raised when the player receives `Kudos` from other players post-game.
*   **Usage:**
    *   **Matchmaking:** The system can use this score to create healthier matches, potentially creating a separate queue for players with very low reputation scores to keep them from disrupting the general player base.
    *   **Moderation:** A sudden drop in a player's reputation score can automatically flag them for administrative review in the **Internal Admin Tool**.

## 4. Feature: Administrative Moderation Tools

*   **Overview:** A set of controls available to administrators via the **Internal Admin Tool** to enforce the code of conduct.
*   **Tool UI:** A "Player Management" tab in the admin tool allows searching for players by handle. Viewing a player's profile shows their report history and reputation score.
*   **Available Admin Actions:**
    *   **View Reports:** See all reports filed against a player, including the chat logs.
    *   **Issue Warning:** Sends a formal `PRIVATE_NOTIFICATION` to the player with a canned or custom warning message.
    *   **Force Mute:** An API action that flags a player's account. The `GameActor` will reject any `POST_CHAT_MESSAGE` actions from a muted player for a specified duration.
    *   **Suspend Account:** Flags a player's account as suspended, preventing them from logging in for a specified duration.
    *   **Ban Account:** Permanently deactivates the player's account.