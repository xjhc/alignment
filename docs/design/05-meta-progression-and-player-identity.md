# Design: Meta-Progression & Player Identity

## 1. Philosophy: Your Reputation is Your Reward

While each game of `Alignment` is a self-contained experience, the meta-progression system is designed to give players a sense of long-term accomplishment, identity, and status within the community.

Our philosophy is **"Recognition over Advantage."** All meta-rewards are purely cosmetic and statistical. They are designed to build a player's reputation and allow for self-expression, but they will **never** provide any in-game gameplay advantage. This preserves the competitive integrity of the core game loop.

## 2. Core Components of Player Identity

A player's permanent identity is composed of three main elements, all accessible from a dedicated "Profile" page outside of an active game.

#### **A. The Player Profile Page**

This is the central hub for a player's history and accomplishments.

*   **UI:** A clean, professional dashboard that feels like a "Loebian Inc. personnel file."
*   **Key Sections:**
    1.  **Header:** Player's chosen `Handle` and `Avatar`, with their current `Title` displayed prominently below. A "Kudos" counter shows positive community feedback.
    2.  **Career Snapshot:** High-level, lifetime statistics (Total Games Played, Win Rate, Preferred Faction).
    3.  **Performance Analytics:** Detailed statistics broken down by role and alignment (e.g., "As CISO: 65% Win Rate," "As AI: 4 Successful Conversions").
    4.  **Achievements Showcase:** A section where players can choose a few of their proudest achievements to display publicly.
    5.  **Match History:** A list of recent games played, with links to their detailed `Post-Game Analysis` reports.

#### **B. Customizable Avatars**

*   **System:** Players start with a small set of basic avatars. New avatars are unlocked by completing Achievements.
*   **Implementation:** Avatars are a library of high-quality, consistently styled SVG icons.
    *   *Examples:* A "Glitching Skull" avatar for winning 10 games as AI, a "Golden Shield" avatar for successfully protecting a key target as CEO.

#### **C. Player Titles**

*   **System:** Titles are prestigious, prefix-style labels that a player can choose to display below their name on their profile and in game lobbies.
*   **Unlocking:** Titles are the primary reward for completing difficult or unique achievements.
*   **Examples:**
    *   `Corporate Analyst` (Default)
    *   `Junior Detective` (Win 5 games as Human)
    *   `Master of Deception` (Win a game as AI without receiving a single vote)
    *   `The Prophet` (Correctly vote to eliminate the AI on Day 1)

## 3. The Achievement System

Achievements are the engine that drives all meta-progression. They are designed to encourage skillful play, experimentation with different strategies, and long-term engagement.

*   **Tracking:** The server will track a wide range of in-game statistics. After each game, a `StatsProcessor` service will parse the final `GameState` and award any achievements earned.
*   **Categories:** Achievements will be organized into several categories:
    *   **Career Milestones:** Based on volume (e.g., Play 100 games).
    *   **Faction Mastery:** Tied to skillful play as Human or AI.
    *   **Role-Specific Feats:** Unique, difficult tasks for each role.
    *   **"Style" Achievements:** For unique or humorous accomplishments.
*   **Rewards:** Every achievement unlocks something tangible (Avatars or Titles).

## 4. Player Safety & Community Tools

*   **Kudos:** A post-game commendation system to promote positive behavior.
*   **Reporting:** A system for reporting players for toxicity or rule-breaking.
*   **Blocking:** A system allowing a player to prevent being matched with a specific other player in the future.

This system provides a compelling "game outside the game," giving players a strong incentive to return, master different roles, and build their unique identity within the `Alignment` community.