# Feature Implementation Designs

This document contains detailed implementation designs for the major gameplay features and systems described in the main **[Game Design Document](../01-game-design-document.md)**.

While the `/docs/architecture` directory explains the foundational *systems* (like the Actor Model), the documents here explain the implementation of specific *game rules* and mechanics. They bridge the gap between high-level game design and the concrete engineering work required to bring those features to life. Each document outlines a feature's purpose, the expected user flow, the necessary API calls, and key server-side logic.

---

## Index of Feature Designs

### UI & UX Design
*   **[04: State-Driven UI Map](./../design/04-state-driven-ui-map.md):** The master "if-this-then-that" map that connects every game state to a specific UI representation.
*   **[02: Onboarding & First-Time UX](./../design/02-onboarding-and-first-time-ux.md):** The "learn-by-doing" approach for new players, guided by the `Loebmate` assistant.
*   **[01: Notification & Feedback System](./../design/01-notification-and-feedback-system.md):** The tiered system for delivering information to the player (Toasts, System Messages, Whispers, etc.).
*   **[03: Player Dossier & Inspector Panel](./../design/03-player-dossier-and-inspector-panel.md):** The design for the main panel used to inspect player information.

### Lobby & Pre-Game
*   **[08: Game Lobby & Matchmaking](./08-game-lobby-and-matchmaking.md):** The system for creating, browsing, and joining games. Covers public/private lobbies, invite links, and host controls (kick, transfer).

### Core Gameplay Loop
*   **[15: Day/Night Cycle & Phase Management](./15-day-night-cycle-and-phase-management.md):** The master state machine for the game. Details the sequence, duration, and transition logic for all timed phases.
*   **[01: Voting & Deactivation](./01-voting-and-deactivation.md):** The core loop for player nomination and token-weighted voting. Includes logic for changing votes and abstaining.
*   **[02: Crisis & Agenda System](./02-crisis-and-agenda-system.md):** How daily rule changes (Crisis Events) and discussion prompts (Pulse Checks) are managed.
*   **[03: Tokens & Mining](./03-tokens-and-mining.md):** The mechanics of resource generation via the Liquidity Pool and priority-based mining.
*   **[05: Roles & Abilities](./05-roles-and-abilities.md):** How players unlock and use their special powers. Details the `Project Milestones` mechanic.

### Advanced Gameplay Systems
*   **[04: AI Conversion & System Shock](./04-ai-conversion-and-system-shock.md):** The AI's primary win condition path and the human's defense mechanism.
*   **[12: Communication Mechanics](./12-communication-mechanics.md):** Design for specialized communication tools: `/status` updates, the once-per-day "Whisper" action, and post-elimination "Parting Shots".
*   **[06: Personal KPIs & Corporate Mandates](./06-personal-kpis-and-mandates.md):** Implementation of secret objectives and game-wide rule modifiers.
*   **[13: LIAISON & Whistleblower Protocols](./13-liaison-and-whistleblower-protocols.md):** Design for the human catch-up mechanic (LIAISON) and the post-elimination influence system (Whistleblower).

### Post-Game & Meta Systems
*   **[07: Post-Game Analysis](./07-post-game-analysis.md):** The after-action report screen showing stats, highlights, and key moments.
*   **[14: Meta-Progression & Player Identity](./14-meta-progression-and-player-identity.md):** The "game outside the game". Details the system for unlocking cosmetic Avatars and Titles via Achievements, and the "Kudos" system.

### AI & Administration
*   **[09: Internal Admin Tool](./09-internal-admin-tool.md):** The design for a private dashboard to monitor server health and debug live games.

### Community & Safety
*   **[16: Player Safety & Moderation](./16-player-safety-and-moderation.md):** Design for the player reporting system, block functionality, reputation tracking, and administrative moderation tools.