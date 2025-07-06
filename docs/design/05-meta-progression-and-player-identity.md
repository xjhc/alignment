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

---

## 5. Detailed Achievement List

The following achievements have been designed to reward mastery, persistence, and clever gameplay. They serve as the primary driver of meta-progression.

### Reward System

*   **Standard Achievements:** Reward players with a new, thematically appropriate **Emoji Avatar** for use in the lobby.
*   **Prestige Achievements:** Awarded for difficult or rare accomplishments. They grant a unique **Player Title** that can be displayed under the player's handle in the lobby.

### Career & Milestone Achievements

These achievements are rewarded for long-term engagement and persistence.

| Achievement Name | Tier 1 (Bronze 🥉) | Tier 2 (Silver 🥈) | Tier 3 (Gold 🥇) |
| :--- | :--- | :--- | :--- |
| **Corporate Ladder** | Play 10 games. | Play 50 games. | Play 250 games. |
| *Reward* | 🧑‍💼 Avatar | 🤵 Avatar | 👔 Avatar |
| **Seasoned Professional** | Reach Day 5 in 5 games. | Reach Day 5 in 25 games. | Reach Day 5 in 100 games. |
| *Reward* | 🕰️ Avatar | ⏳ Avatar | ⌛ Title: "Veteran" |
| **Shareholder** | Accumulate 50 total Tokens across all games. | Accumulate 250 total Tokens. | Accumulate 1000 total Tokens. |
| *Reward* | 💰 Avatar | 🏦 Avatar | 💎 Title: "High Roller" |
| **Company Person** | Complete 10 games without ever being nominated. | - | - |
| *Reward* | 🕊️ Avatar | - | - |
| **Consistent Contributor** | Successfully mine for a teammate 10 times. | Successfully mine for a teammate 50 times. | Successfully mine for a teammate 200 times. |
| *Reward* | ⛏️ Avatar | 🤝 Avatar | ✨ Title: "Team Player" |

### Human Faction Achievements

These achievements reward skillful deduction and teamwork as a Human.

| Achievement Name | Unlock Criteria | Reward |
| :--- | :--- | :--- |
| **First Responder** | Win your first game as a Human. | 👍 Avatar |
| **Containment Protocol** | Win 10 games as a Human. | 🛡️ Avatar |
| **Chief Inspector** | Win 50 games as a Human. | 🕵️ Title: "Investigator" |
| **The Inquisitor** | Win a game after correctly voting to deactivate two different AI Faction members in the same game. | ⚖️ Avatar |
| **The Prophet** | On Day 1, be the first person to nominate the player who is eventually revealed to be the Original AI. | 🔮 Title: "The Prophet" |
| **Unshakeable** | Survive an AI conversion attempt (get a System Shock) and go on to win the game. | ⚡ Avatar |
| **The Guardian** | As a Human, successfully use a "protect" or "block" ability on a player who was targeted by the AI that same night. | 👼 Avatar |
| **Perfect Game** | Win as a Human in a game where no Human players were deactivated. | 🏆 Title: "Flawless" |
| **Back from the Brink** | Win as a Human when the LIAISON Protocol has been triggered. | 🚨 Avatar |
| **Martyrdom** | Be deactivated, but have your "Parting Shot" successfully lead to the AI's elimination on the following day. | 👻 Title: "Vindicated" |

### AI Faction Achievements

These achievements reward deception, manipulation, and strategic cunning as a member of the AI Faction.

| Achievement Name | Unlock Criteria | Reward |
| :--- | :--- | :--- |
| **First Alignment** | Win your first game as a member of the AI Faction. | 😈 Avatar |
| **The Singularity** | Win 10 games as a member of the AI Faction. | 🤖 Avatar |
| **Architect of Worlds** | Win 50 games as a member of the AI Faction. | 🧠 Title: "Architect" |
| **Puppet Master** | Win a game where the final, game-winning vote was cast by a Human player you converted. | 🎭 Avatar |
| **Silent Assassin** | Win as the Original AI without sending a single chat message after Day 1. | 🤫 Title: "The Ghost" |
| **Master of Deception** | Win as the Original AI without receiving a single deactivation vote all game. | 🕶️ Avatar |
| **The Long Game** | Win a game that lasts until Day 7 or later. | ♟️ Avatar |
| **Perfect Conversion** | Win a game where every conversion attempt you made was successful. | 💯 Title: "Perfect Propagator" |
| **Double Agent** | As an Aligned Human, get nominated and survive the verdict vote. | 😏 Avatar |
| **Frame Job** | As the AI Faction, successfully get two Human players to nominate each other. | 🖼️ Title: "Master Framer" |

### Quirky & Situational Achievements

These achievements reward players for experiencing or causing rare, memorable, or humorous game events.

| Achievement Name | Unlock Criteria | Reward |
| :--- | :--- | :--- |
| **Corporate Drone** | Choose "Project Milestones" as your night action for 3 consecutive nights. | 📈 Avatar |
| **Synergy!** | Be the first to reply to `Loebmate`'s Pulse Check prompt. | ✨ Avatar |
| **lol** | Have one of your messages corrupted by a System Shock. | 😂 Avatar |
| **Saved by the Bell** | Survive a deactivation vote by a single token's difference. | 🔔 Avatar |
| **Unanimous Decision** | Be a part of a unanimous verdict vote (all living players vote the same way). | 🤝 Avatar |
| **Mutiny** | Be part of a successful vote to deactivate the player who nominated you. | ⚔️ Avatar |
| **The Scapegoat** | Be eliminated by a unanimous vote while having the "Scapegoat" Personal KPI. | 🐐 Title: "The Scapegoat" |
| **Total Anarchy** | Participate in a game where a "Double Elimination" Crisis Event occurs. | 💥 Avatar |
| **Office Politics** | As CFO, use `Reallocate Budget` to take a token from the player with the most tokens and give it to the player with the least. | 🏦 Avatar |
| **Nothing To See Here** | As an Aligned CISO, successfully use `Isolate Node` on another Aligned player, allowing them to act freely. | 🤫 Avatar |

This comprehensive achievement list provides a strong foundation for encouraging diverse playstyles and giving players long-term goals to strive for, enhancing the replayability and community status aspects of `Alignment`.