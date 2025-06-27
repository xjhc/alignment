# Feature: Meta-Progression & Player Identity

This document describes the design for the "game outside the game"—the system for long-term player progression, identity customization, and reputation. All rewards are purely cosmetic and statistical to maintain competitive integrity.

## 1. Feature: Player Profile & Statistics

*   **Overview:** A permanent, persistent profile for each player that tracks their career history and achievements. This will be accessible from the main menu, outside of an active game.
*   **Data Schema (`players` collection/table):**
    ```json
    {
      "player_id": "user123",
      "handle": "Vex",
      "stats": {
        "games_played": 10,
        "human_wins": 5,
        "ai_wins": 2,
        "kudos_received": 25
      },
      "unlocked_achievements": ["JUNIOR_DETECTIVE", "THE_PROPHET"],
      "unlocked_avatars": ["default_set", "prophet_icon"],
      "unlocked_titles": ["Corporate Analyst", "The Prophet"],
      "equipped_avatar": "prophet_icon",
      "equipped_title": "The Prophet",
      "blocked_players": ["user456", "user789"]
    }
    ```
*   **System Flow (Post-Game):**
    1.  After a game ends, the `GameActor` broadcasts the final `GAME_ENDED` event, which includes a `GameAnalysis` object.
    2.  A separate, persistent **`StatsProcessor`** service (which could be a serverless function or a dedicated background worker) listens for `GAME_ENDED` events.
    3.  Upon receiving an event, the `StatsProcessor` iterates through each player in the final game state.
    4.  For each player, it updates their lifetime statistics (wins, games played) in the database.
    5.  It checks the `GameAnalysis` data against the rules for all unearned achievements. If an achievement is earned, it's added to the player's database record.
    6.  The server may then send a `PRIVATE_NOTIFICATION` to the player's client (if still connected) to inform them of the unlock.

## 2. Feature: Achievements, Avatars & Titles

*   **Overview:** Achievements are the engine for meta-progression. Completing them unlocks new Avatars and prestigious Titles.
*   **Implementation:**
    *   Achievement definitions (rules, rewards) will be stored in a server-side registry.
    *   The `StatsProcessor` will check against this registry.
    *   **Avatars:** A library of SVG icons. Unlocking an avatar adds its ID to the `unlocked_avatars` array in the player's profile.
    *   **Titles:** A library of strings. Unlocking a title adds it to the `unlocked_titles` array.
    *   **UI:** The Player Profile page will have sections for "Avatars" and "Titles" where a player can browse their unlocked cosmetics and click to equip one. Equipping an item updates the `equipped_avatar` or `equipped_title` field in their database record.

## 3. Feature: Kudos (Commendation System)

*   **Overview:** A post-game system to promote positive sportsmanship.
*   **User Flow:**
    1.  On the `Game Over` screen, a "Give Kudos" 👍 icon appears next to each other player's name.
    2.  A player can click this icon once per game.
*   **System Flow:**
    1.  The client sends a `GIVE_KUDOS` action.
    2.  The server receives the action and validates that the player hasn't already given kudos for this match.
    3.  The server increments the `kudos_received` counter for the target player in the database.
*   **API & Payloads:**
    *   **Action:** `GIVE_KUDOS`
    *   **Payload:** `{ "game_id": "g-xyz", "target_player_id": "p-bob" }`