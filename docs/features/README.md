# Feature Implementation Designs

This directory contains detailed implementation designs for the major gameplay features and systems described in the main **[Game Design Document](../01-game-design-document.md)**.

While the `/docs/architecture` directory explains the foundational *systems* (like the Actor Model), the documents here explain the implementation of specific *game rules* and mechanics. They bridge the gap between high-level game design and the concrete engineering work required to bring those features to life.

Each document outlines a feature's purpose, the expected user flow, the necessary API calls (Actions & Events), and any key server-side logic.

## Index of Feature Designs

### Core Gameplay Loop
*   **[01: Voting & Deactivation](./01-voting-and-deactivation.md):** The core loop for player elimination.
*   **[02: Crisis & Agenda System](./02-crisis-and-agenda-system.md):** How daily rule changes and discussion prompts are managed.
*   **[03: Tokens & Mining](./03-tokens-and-mining.md):** The mechanics of resource generation and influence.
*   **[04: AI Conversion & System Shock](./04-ai-conversion-and-system-shock.md):** The AI's primary win condition and the human's defense mechanism.
*   **[05: Roles & Abilities](./05-roles-and-abilities.md):** How players unlock and use their special powers.

### Meta-Game & Advanced Systems
*   **[06: Personal KPIs & Mandates](./06-personal-kpis-and-mandates.md):** The implementation of secret objectives and game-wide modifiers.
*   **[07: Post-Game Analysis](./07-post-game-analysis.md):** The after-action report screen showing stats and key moments.
*   **[08: Game Lobby & Matchmaking](./08-game-lobby-and-matchmaking.md):** The system for creating, browsing, and joining games before they start.
*   **[09: Internal Admin Tool](./09-internal-admin-tool.md):** The design for a private dashboard to monitor server health and debug live games.