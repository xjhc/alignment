# ADR-007: Implement a Progressive Identity Model

## Context

A core requirement for the game is preventing a single user from joining a lobby multiple times, as this would break the game's balance. The initial consideration was to require all users to log in via a third-party OAuth provider (e.g., Discord) to establish a unique, persistent identity.

However, a mandatory registration process introduces significant friction for new players who may simply want to try the game quickly. This friction could negatively impact user acquisition and initial engagement. We need a solution that provides reasonable duplicate-join prevention for casual players while offering robust, cheat-proof identity for engaged users.

## Decision

We will implement a **hybrid, progressive identity model** with two distinct tiers of user identity. The system will default to a low-friction "Guest" mode and offer an upgrade path to a "Registered" account.

1.  **Tier 1: Guest Identity (Anonymous)**

    - **Mechanism:** When a user first visits the site, the client application will generate a V4 UUID and store it in the browser's `localStorage`. This ID will be prefixed with `guest:` (e.g., `guest:123e4567-...`).
    - **Scope:** This identity is tied to the specific browser profile. It prevents a user from joining the same game from two different tabs but does **not** prevent them from joining via an incognito window or a different browser.
    - **Data Persistence:** Game history and stats for guest users will be ephemeral, held in server memory only for the duration of a single game and then discarded.

2.  **Tier 2: Registered Identity (Authenticated)**

    - **Mechanism:** Users will have the option to "Login with Discord" (or another OAuth provider). Upon successful authentication, the server will issue a secure JWT session cookie. The user's identity will be their persistent, provider-issued ID, prefixed accordingly (e.g., `discord:123456789...`).
    - **Scope:** This identity is tied to the user's account and is persistent across all devices, browsers, and sessions. It provides robust protection against all forms of duplicate joining.
    - **Data Persistence:** Game history and stats for registered users will be saved to a database, linked to their permanent user ID.

3.  **Server-Side Enforcement:**

    - The server will maintain a server-wide, in-memory map of `UserID -> GameID` for all active players (both guests and registered).
    - Before allowing any user to create or join a lobby, the server will check this map. If the user's ID is already present in another active game, the request will be rejected. This prevents a single user, regardless of type, from being in more than one game at a time.

4.  **Upgrade Path:**
    - Guest users will be prompted to register at the end of a match to save their game stats.
    - A successful registration will trigger a one-time "account linking" process where the stats from the just-finished game (keyed by the guest ID) are migrated to the new, permanent user ID.

## Consequences

- **Pros:**

  - **Reduced User Friction:** Maximizes the number of new players who can try the game instantly without the hurdle of creating an account.
  - **Effective "Good Enough" Security:** The `localStorage` ID solves the most common, low-effort "multi-tabbing" scenario.
  - **Robust Anti-Cheat for Engaged Players:** The most dedicated players, who are most likely to care about stats and ranking, are incentivized to register, at which point they are subject to stricter, cheat-proof identity validation.
  - **Clear Upgrade Path:** Provides a clear value proposition for registering ("Save your progress!") which can drive conversion from casual players to community members.

- **Cons:**
  - **Guest-Mode is Exploitable:** A determined user can still "cheat" by using multiple browsers or incognito windows. This is a conscious and accepted trade-off for V1 to prioritize accessibility.
  - **Increased Implementation Complexity:** The server must now handle two types of user identities and manage the active user tracking system. The client needs logic to handle both guest and authenticated states.
  - **Potential for Orphaned Guest Data:** If the server crashes, the link between a guest's browser ID and their active game might be lost, although this is a minor issue as their session was ephemeral anyway.

This hybrid approach provides an optimal balance between accessibility for new users and integrity for the core gameplay experience. It allows us to "gate" the most robust anti-cheat measures behind an action (registration) that engaged players are naturally motivated to take.
