# Feature: Post-Game Analysis Screen

This document describes the implementation design for the Post-Game Analysis screen, which is displayed after the Victory/Defeat announcement. Its purpose is to provide a detailed, data-driven "after-action report" of the game to all players.

## 1. Feature Overview

The Post-Game Analysis screen is a multi-tabbed interface that allows players to explore the key events, player statistics, and memorable moments from the match they just completed. This screen is designed to be the primary driver of post-game discussion, learning, and community engagement, encouraging players to immediately play another round.

## 2. System Flow

1.  **Data Aggregation (Server-Side):**
    *   Throughout the game, the `Game Actor` is not just processing events but also **aggregating statistics** in a separate, non-game-critical data structure. This includes tracking votes cast, mining actions, ability uses, message counts, emoji reactions, etc.
    *   When the game ends, the `Game Actor` performs a final analysis pass to calculate derived stats (like Most Valuable Personnel), identify key events (like "Turning Points" based on heuristics), and select "Quote Highlights."
2.  **Data Transmission:**
    *   This complete `GameAnalysis` object is included as the payload of the `GAME_ENDED` event, alongside the `winning_faction`. This event is sent once to all clients.
3.  **Client-Side Rendering:**
    *   The client receives the `GameAnalysis` object and stores it. After the player dismisses the main Victory/Defeat screen, the client uses this static object to render the various tabs and statistics on the Post-Game Analysis screen. The screen is a read-only report.

## 3. Screen Tabs & Content Breakdown

The interface will be broken down into the following tabs, each drawing from the `GameAnalysis` payload.

#### Tab 1: Summary

*   **Purpose:** An at-a-glance overview of the game's highlights.
*   **Content:**
    *   **Most Valuable Personnel (MVP):** Highlights a player based on a server-calculated score. The formula can be tuned, but an example is: `Score = (Tokens Mined for others * 2) + (Correct Deactivation Votes * 3) - (Incorrect Deactivation Votes * 1)`.
    *   **Key Turning Point:** A server-identified event that significantly shifted the game's outcome. The heuristic could identify events like "The first correct deactivation of an AI faction member" or "A successful block by the CISO that prevented a game-winning conversion."
    *   **Parting Shots:** A display of all deactivated players' final, permanent status messages.

#### Tab 2: Event Timeline

*   **Purpose:** A chronological, visual log of the game's most critical moments.
*   **Content:** A scrollable list of key events, tagged by the day they occurred. The server will filter the full event log to only include:
    *   Player deactivations (including their revealed alignment).
    *   AI conversions.
    *   Critical ability uses (e.g., a successful CISO block, a pivotal CFO token transfer).
    *   Daily Crisis Event announcements.

#### Tab 3: Personnel Analytics

*   **Purpose:** A detailed statistical breakdown for every player in the game.
*   **Content:** A grid of "player cards," each showing:
    *   **Influence Stats:** Tokens Mined, Tokens Received, Final Token Count.
    *   **Deduction Stats:** Correct Votes for Deactivation, Incorrect Votes for Deactivation.
    *   **Social Stats:** Times Nominated by Others, Total Messages Sent.
    *   **Survival:** Days Survived.

#### Tab 4: Communication Highlights

*   **Purpose:** A showcase of the most memorable social moments from the game's chat log.
*   **Content:**
    *   **Most Reacted-To Message:** The single chat message that received the highest number of total emoji reactions.
    *   **Notable Quotes:** A server-selected collection of 2-3 other messages. Heuristics for selection could include high reaction counts, messages that directly mentioned an AI player who was later deactivated, or the first message that accused the correct AI.

By providing this rich, data-driven summary, we give players the tools to tell the story of their game, compare strategies, and build a lasting engagement with `Alignment`.