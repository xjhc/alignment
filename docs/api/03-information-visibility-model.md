# API: Information Visibility Model

## 1. Overview and Security Principle

This document defines the rules for information visibility in `Alignment`. It is a critical part of both the game design and the system's security model.

The core principle is: **The server is the sole arbiter of information visibility. The client is a "dumb" renderer and must never be trusted to filter information.**

It is the server's absolute responsibility to ensure that no player ever receives data they are not privileged to see. This prevents cheating and maintains the integrity of the game's social deduction mechanics.

## 2. Server-Side Enforcement

The `Game Actor` holds the complete, unfiltered `GameState` with all secret information. Before broadcasting any event, the server must determine the correct audience for that event and, if necessary, create different versions of the payload for different recipients.

*   **Example: Private Events.** For an event like `ROLES_ASSIGNED`, the server does not broadcast a single message. Instead, it iterates through each player and sends a unique, private version of the event containing only that specific player's role and KPI.
*   **Example: Factional Events.** For a covert action, the server might send a public `NIGHT_ACTIONS_RESOLVED` event to all players, but also send a special `CHAT_MESSAGE_POSTED` event containing secret results only to the players currently in the AI faction.

## 3. Information Tiers

All data in the game falls into one of three visibility tiers.

| Tier | Description | Examples |
| :--- | :--- | :--- |
| **Public** | Information visible to all living players at all times. This forms the basis of public knowledge and discussion. | • Player token counts <br> • Which players are alive/deactivated <br> • The current Crisis Event <br> • Aggregate vote totals (but not who voted for whom) <br> • Public Player Statuses <br> • A deactivated player's final role and alignment |
| **Private (Per-Player)** | Information known only to a single player. This is the most sensitive data and must be delivered via private, targeted events. | • Your own Role and Alignment <br> • Your secret Personal KPI <br> • Your hidden `AI Equity` score (if human) <br> • The fact that you have a `System Shock` <br> • The contents of a private message (DM) you sent or received |
| **Factional (Hidden)** | Information known only to members of a specific faction (typically the AI faction). This is managed via a separate, secret communication channel. | • The identity of the Original AI and all Aligned players <br> • The contents of the `#aligned` chat channel <br> • The true results of covert abilities (e.g., the `Run Audit` ability) |

Developers must consult this model when implementing any feature that handles or transmits game state to ensure these visibility rules are strictly enforced.