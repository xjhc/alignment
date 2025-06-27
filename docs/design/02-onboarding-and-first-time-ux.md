# Design: Onboarding & First-Time User Experience (FTUE)

## 1. Philosophy: Learn by Doing, Guided by the System

The core philosophy of our onboarding process is **"Show, Don't Just Tell."** A new player's first game is their tutorial. We will avoid lengthy, mandatory "how-to-play" walls of text. Instead, we will use a system of contextual hints and progressive disclosure to teach the mechanics as they become relevant.

The primary tool for this is the **[`Loebmate` Assistant](./01-notification-and-feedback-system.md)**, which is enabled by default for all players and can be turned off in settings. This ensures that help is always available but unobtrusive.

Our goal is to get a new player from the landing page into their first game in **under 60 seconds**, feeling equipped to participate, even if they haven't mastered every nuance.

## 2. The New Player Journey

This journey map outlines the step-by-step experience for a first-time player.

#### **Step 1: The Landing & Login**
*   **Goal:** Get the player to create their identity with zero friction.
*   **UI:** The `LoginScreen` is clean and focused, asking only for a handle and an avatar.
*   **Onboarding Element:** A single, unobtrusive line of text below the "Browse Lobbies" button: `New here? Our in-game assistant will guide you.` This sets expectations and reduces anxiety.

#### **Step 2: The Lobby Browser**
*   **Goal:** Guide the player to a suitable first game.
*   **UI:** The `LobbyListScreen` is presented.
*   **Onboarding Element:** Lobbies suitable for new players are highlighted with a special "First-Timers Welcome" tag. These lobbies might have slightly longer phase timers or enhanced in-game guidance enabled. The "Create New Game" button is also highlighted as a primary action.

#### **Step 3: The First Game with `Loebmate`**

The `Loebmate` Assistant is the core of the FTUE. It provides contextual help via private messages.

**A. The Role Reveal Screen**
*   **Goal:** Ensure the player understands their immediate objective.
*   **Onboarding Element:** A `Loebmate` "whisper" appears alongside the role card:
    > `[PRIVATE from Loebmate]` As a **HUMAN**, your goal is simple: listen to the discussion, identify the player who seems least trustworthy, and vote with the group to deactivate them. Your advanced abilities will become clear as you play.

**B. Contextual `Loebmate` Whispers**
*   **Goal:** Use the existing system bot to provide just-in-time instructions at the start of each new phase.
*   **Onboarding Element:**
    *   **Start of Nomination Phase:**
        > `[PRIVATE from Loebmate]` **New Action Unlocked: Nominate!** The discussion is over. It's time to choose who to put on trial. Click the "Nominate" button next to a player's name in the roster to cast your vote.
    *   **Start of Night Phase:**
        > `[PRIVATE from Loebmate]` **Welcome to the Night Phase!** The main channel is locked. You have 30 seconds to secretly choose an action from the menu below. "Mine for Tokens" helps your team, while "Project Milestones" helps you unlock your ability. Choose wisely!

**C. The First Elimination**
*   **Goal:** To clarify the consequence of being deactivated without being punitive.
*   **Onboarding Element:** If a player is deactivated, their `Exit Interview` screen will include a `Loebmate` whisper:
    > **You have been deactivated.** Don't worry, the game isn't over! As a "Consultant" in the `#off-boarding` channel, you can still observe the game and even influence future events through the **Whistleblower Protocol**.

## 3. Progressive Disclosure Strategy

We teach concepts as they become relevant, primarily through `Loebmate`.

| Game Day | Concepts Introduced | Onboarding Method |
| :--- | :--- | :--- |
| **Day 1** | Chat, Tokens, Voting, Phases | `Loebmate` Whispers at the start of each phase. |
| **Night 1** | Night Actions (Mining, Projects) | `Loebmate` Whisper explaining the choice. |
| **Day 2** | SITREP, Crisis Events, System Shocks| The SITREP itself is the teaching tool. A `Loebmate` whisper might add: `Pay attention to the Night Activity Log. It may contain clues.`|
| **Night 2**| Role Abilities (if unlocked) | A "New Ability Unlocked!" toast notification, followed by a `Loebmate` whisper explaining how to use it. |

This strategy ensures that the player is given a small, digestible chunk of information at each stage of their first game, allowing them to build a mental model of the rules organically and contextually. Veterans can simply turn `Loebmate` off.