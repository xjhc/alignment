# Design: Cognitive Load & Usability Heuristics

## 1. Philosophy: The Interface Must Get Out of the Way

`Alignment` is a game of intense mental effort. Players are constantly tracking conversations, building threat models, and managing social dynamics under time pressure. The user interface must be a tool that *aids* this process, not an obstacle that adds to it.

Our core philosophy is to **ruthlessly minimize extraneous cognitive load**. The UI should feel like an extension of the player's thoughts, presenting the right information at the right time and making desired actions effortless. We achieve this by adhering to a set of established usability heuristics, tailored to the specific challenges of social deduction.

## 2. The Core Heuristics for `Alignment`

These principles are the foundation of our UX design. They should be used as a checklist during design reviews and when evaluating the usability of a new feature.

#### **Heuristic 1: Visibility of System Status ("What's happening now?")**
The player must always know the current state of the game without having to guess.

*   **Implementation:**
    *   The **current game phase** and its **countdown timer** are always visible in the `ChatHeader`.
    *   When an action is taken (e.g., voting), the UI provides **immediate feedback** (e.g., the button enters a "selected" state).
    *   **Loading states** are explicit. Any action that requires a server round-trip must visually indicate that it is in progress (e.g., a spinner in a button).

#### **Heuristic 2: Recognition Over Recall ("Don't make me remember.")**
A player's mental energy should be spent on deduction, not on remembering UI rules or game data.

*   **Implementation:**
    *   **Player Dossier:** Instead of forcing players to remember who has which role, the Dossier makes this public information easily accessible.
    *   **Tooltips:** All icons and non-obvious UI elements have descriptive tooltips on hover.
    *   **Contextual Actions:** The UI only presents actions that are valid in the current phase. The `ContextualInputArea` is the prime example of this, removing the need for the player to remember which commands are available.

#### **Heuristic 3: Consistency and Standards ("I know how this works.")**
The UI must follow consistent patterns, both internally and with conventions established by other popular chat applications.

*   **Implementation:**
    *   **Internal Consistency:** An action performed on a player (e.g., opening their dossier) is triggered the same way everywhere—clicking their avatar in chat is the same as clicking their card in the roster.
    *   **External Consistency:** We adopt familiar patterns from apps like Slack and Discord: `@mentions`, `Cmd+K` for a command palette, `Esc` to close views, and emoji reactions. This lowers the learning curve.
    *   **Design System:** Our `Layout`, `Typography`, and `Motion` systems ensure that spacing, fonts, and animations are uniform throughout the application.

#### **Heuristic 4: Error Prevention ("Guide me away from mistakes.")**
The best error message is no error at all. The design should make it difficult to perform incorrect actions.

*   **Implementation:**
    *   **Disabling Actions:** Buttons for invalid actions are `disabled` and provide a tooltip explaining why (e.g., a disabled "Start Game" button says "Requires at least 4 players").
    *   **Confirmation for Destructive Actions:** Critical, irreversible actions (like casting a final vote or deleting an account) must use a confirmation modal to prevent mis-clicks.
    *   **Clear Constraints:** Inputs have clear character limits, and the UI prevents submitting an empty message.

#### **Heuristic 5: Aesthetic and Minimalist Design ("Signal, not noise.")**
Every element on the screen should have a purpose. We favor clarity and focus over decoration.

*   **Implementation:**
    *   **Information Hierarchy:** The most important information is given the most visual weight. The `Layout & Grid System` and `Typography Scale` are the tools we use to enforce this.
    *   **Negative Space:** We use our spacing tokens deliberately to group related items and separate unrelated ones, allowing the user's eye to parse information easily.
    *   **No Clutter:** We avoid unnecessary icons, borders, and decorative flourishes. If an element doesn't serve a functional or informational purpose, it is removed.

#### **Heuristic 6: The Five-Second Rule (Clarity at a Glance)**
This is our primary litmus test for screen and component design.

*   **Definition:** A player should be able to look at any screen or major UI component and understand its purpose and the most critical piece of information within **five seconds**.
*   **Application:**
    *   *SITREP:* In five seconds, can I see who was affected at night and what the day's crisis is?
    *   *Vote UI:* In five seconds, can I see who is nominated and what the current vote tally is?
    *   *Player Dossier:* In five seconds, can I see the player's name, tokens, and role?
*   **Enforcement:** This rule forces us to be ruthless in our prioritization of information on every screen.

By consistently applying these heuristics, we can build an interface that feels intuitive, responsive, and professional, allowing the rich strategic depth of `Alignment` to shine through.