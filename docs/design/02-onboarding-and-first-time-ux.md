# Design: Onboarding & First-Time User Experience (FTUE)

## 1. Philosophy: Learn by Doing, Not by Reading

The core philosophy of our onboarding process is **"Show, Don't Just Tell."** A new player's first game is their tutorial. We will avoid lengthy, mandatory "how-to-play" walls of text. Instead, we will use a system of contextual hints, a guided first game, and progressive disclosure to teach the mechanics as they become relevant.

Our goal is to get a new player from the landing page into their first game in **under 60 seconds**, feeling equipped to participate, even if they haven't mastered every nuance.

## 2. The New Player Journey

This journey map outlines the step-by-step experience for a first-time player.

#### **Step 1: The Landing & Login**
*   **Goal:** Get the player to create their identity with zero friction.
*   **UI:** The `LoginScreen` is clean and focused. It asks for only one thing: a handle. The avatar selection is a simple, visual choice.
*   **Onboarding Element:** A single, unobtrusive line of text below the "Browse Lobbies" button: `New here? Your first game will be a guided experience.` This sets expectations and reduces anxiety.

#### **Step 2: The Lobby Browser**
*   **Goal:** Guide the player to a suitable first game.
*   **UI:** The `LobbyListScreen` is presented.
*   **Onboarding Element:** Lobbies suitable for new players are highlighted with a special "First-Timers Welcome" tag. These lobbies might have slightly longer phase timers or enhanced in-game guidance enabled. The "Create New Game" button is also highlighted as a primary action.

#### **Step 3: The First Game - "Assisted Mode"**
This is the core of the FTUE. A player's first game (tracked via a simple flag on their local profile) activates a special "Assisted Mode." This mode is a layer of contextual help that is disabled for all subsequent games.

**A. The Role Reveal Screen**
*   **Goal:** Ensure the player understands their immediate objective.
*   **Onboarding Element:** The `RoleRevealScreen` includes an extra, highlighted section for new players:
    > **Your First Mission:** As a **HUMAN**, your goal is simple: listen to the discussion, identify the player who seems least trustworthy, and vote with the group to deactivate them. Your advanced abilities will become clear as you play.

**B. The In-Game HUD & Tooltips**
*   **Goal:** Explain UI elements *when the player interacts with them*.
*   **Onboarding Element:** Every key UI element in the game will have a detailed, first-time-only tooltip that explains its purpose. These tooltips are more descriptive than the standard ones.
    *   *Hovering over their own Token count:* `Tokens are your voting power. The more you have, the more your vote counts. Mine for teammates at night to earn more.`
    *   *Hovering over Project Milestones:* `Complete project work at night to unlock your powerful role ability.`
    *   *Hovering over the Phase Timer:* `This is the time remaining for the current discussion phase. Make your case before it runs out!`

**C. `Loebmate` - The Onboarding Assistant**
*   **Goal:** Use the existing system bot to provide contextual, just-in-time instructions.
*   **Onboarding Element:** During a new player's first game, `Loebmate` will send them **private, DMed instructions** at the start of each new phase. These messages do not appear in the public `#war-room`.

    *   **Start of Nomination Phase:**
        > `[PRIVATE from Loebmate]` **New Action Unlocked: Nominate!** The discussion is over. It's time to choose who to put on trial. Click the "Nominate" button next to a player's name in the roster to cast your vote.

    *   **Start of Night Phase:**
        > `[PRIVATE from Loebmate]` **Welcome to the Night Phase!** The main channel is locked. You have 30 seconds to secretly choose an action from the menu below. "Mine for Tokens" helps your team, while "Project Milestones" helps you unlock your ability. Choose wisely!

**D. The First Elimination**
*   **Goal:** To clarify the consequence of being deactivated without being punitive.
*   **Onboarding Element:** If a new player is the first to be deactivated, their `Exit Interview` screen will have an extra message:
    > **You have been deactivated.** Don't worry, the game isn't over! As a "Consultant" in the `#off-boarding` channel, you can still observe the game and even influence future events through the **Whistleblower Protocol**.

## 3. Progressive Disclosure Strategy

We will not explain every mechanic at once. We teach concepts as they become relevant.

| Game Day | Concepts Introduced | Onboarding Method |
| :--- | :--- | :--- |
| **Day 1** | Chat, Tokens, Voting, Phases | `Loebmate` DMs, contextual tooltips. |
| **Night 1** | Night Actions (Mining, Projects) | `Loebmate` DM explaining the choice. |
| **Day 2** | SITREP, Crisis Events, System Shocks | The SITREP itself is the teaching tool. A tooltip on the Crisis and Shock sections provides more detail. |
| **Night 2** | Role Abilities (if unlocked) | A "New Ability Unlocked!" toast notification, followed by a `Loebmate` DM. |

This strategy ensures that the player is given a small, digestible chunk of information at each stage of their first game, allowing them to build a mental model of the rules organically and contextually.