# Community: Feedback and Game Balance Loop

## 1. Philosophy: Our Players are Co-Designers

`Alignment` is a complex, evolving game. The best ideas for improving it will come from our most passionate players. Our philosophy is to create a transparent and structured **feedback loop** where player input and game data are systematically collected, analyzed, and used to inform our development priorities.

This document outlines the process by which we listen to our community and use data to make the game more balanced, fun, and engaging for everyone.

## 2. The Feedback Channels: Where We Listen

We will centralize our feedback collection to ensure no good idea gets lost.

*   **Primary Channel: In-Game Feedback Tool**
    *   A simple, non-intrusive "Feedback" button will be present in the main menu and post-game screen.
    *   This will open a form with a few clear categories:
        *   `Bug Report`
        *   `Balance Suggestion` (e.g., "The CISO role feels too weak.")
        *   `Feature Request`
        *   `General Feedback`
    *   Submitting feedback through this tool automatically includes valuable context like the `gameID` and `playerID`, helping us debug issues faster.

*   **Secondary Channel: Official Discord Server**
    *   We will maintain dedicated channels like `#bug-reports` and `#suggestions`.
    *   Our community team will monitor these channels and formally log actionable ideas into our internal tracking system.

## 3. The Analytics Loop: What We Measure

In addition to direct feedback, we will use anonymous, aggregated gameplay data to understand the health and balance of the game. Our `simulation-tests` provide a baseline, but real-player data is the ultimate source of truth.

#### **A. Key Balance Metrics We Track:**

*   **Faction Win Rate:** The overall win percentage for the Human vs. AI factions. Our goal is to keep this as close to 50/50 as possible over thousands of games.
*   **Role Performance:** The win rate for players when they are assigned a specific role (e.g., CISO, CTO). This helps us identify under- or over-powered roles.
*   **Crisis Event Impact:** How does a specific Crisis Event affect the faction win rate for that day? Does "Press Leak" disproportionately help the AI?
*   **Game Length:** The average number of days a game lasts. If games are ending too quickly or dragging on too long, we need to adjust.

#### **B. The Process:**

1.  **Data Collection:** Our server logs key, anonymous game events (`GAME_ENDED`, `PLAYER_ELIMINATED`, `ROLE_ASSIGNED`) to an analytics database.
2.  **Dashboarding:** We will maintain an internal dashboard (in Grafana) that visualizes these key balance metrics over time.
3.  **The Balance Council:** A cross-functional team (e.g., a designer, an engineer, a community manager) will meet on a regular cadence (e.g., bi-weekly) to review these dashboards and the qualitative feedback from the community.
4.  **Actionable Insights:** This meeting produces a prioritized list of potential balance changes, bug fixes, and feature ideas. For example: "The AI faction is winning 58% of the time. The 'Tainted Data' crisis seems to be a major factor. Let's run a simulation where we tune its AI Equity bonus down from +2 to +1."

## 4. Closing the Loop: Communicating with Players

Transparency is key to building trust. Players are more invested when they feel heard and can see the impact of their feedback.

*   **Public Bug & Suggestion Tracker:** We will consider using a public Trello board or similar tool where players can see the status of popular suggestions and major bugs.
*   **Developer Updates:** Regular blog posts or Discord announcements that explain *why* we are making certain changes. Instead of just saying "Nerfed the CISO," we explain: "Our data showed the CISO role had a 60% win rate, significantly higher than other roles. To bring it into balance, we are adjusting the `Isolate Node` ability..."
*   **Patch Notes:** Every single game update, no matter how small, will be accompanied by clear, concise patch notes.

By combining direct player feedback with hard data, we can move beyond guessing and make informed, evidence-based decisions to continuously improve `Alignment`.