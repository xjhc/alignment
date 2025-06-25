# Design: Chat & Communication UX

## 1. Philosophy: The Chat is the Game Board

In `Alignment`, the chat interface is not a secondary feature; it is the primary game board. All strategic interactions, deductions, and deceptions happen here, in a single, chronological stream of evidence.

Our design philosophy is to create a **high-signal, low-friction, professional-grade communication tool** that empowers players to focus on strategy. The UX should feel less like a casual game chat and more like a high-stakes Slack or Discord channel during a corporate crisis.

**Core Principle:** All communication must remain in a single, flat channel. There will be **no threaded discussions**. This forces all arguments to happen in the open and preserves the integrity of the event timeline, making it a "single source of truth" for player deduction.

## 2. Core Components of the Chat Experience

#### **A. The Message Log (`#war-room`)**

This is the central, chronologically-ordered log of all public communication.

*   **Structure:** A compact, modern chat layout. Messages from the same author sent within a 2-minute window are grouped to reduce visual noise. A "date marker" (e.g., `--- DAY 2 ---`) clearly separates discussion from different game days.
*   **Message Anatomy:**
    *   `Avatar`: A simple, consistent visual identifier for the player.
    *   `Author Name`: Clearly legible, with a distinct color or style for the local player (`is-me`) and AI/System messages (`Loebmate`).
    *   `Timestamp`: A subtle, low-contrast timestamp (`HH:MM`) appears on hover.
    *   `Content`: The message text, including any rendered "quoted reply" content.
*   **System Messages:** Messages from `Loebmate` (SITREPs, phase changes) must be visually distinct from player messages. They should use a different background color, a brand icon, and a more structured, formal layout.

#### **B. The Typing Indicator**

To enhance real-time presence and psychological tension, players must be able to see who is currently typing.

*   **UI:** A small, unobtrusive indicator will appear at the bottom of the chat log: `Eve is typing...` or `Alice, Bob, and 2 others are typing...`.
*   **Implementation:** The client will emit a `START_TYPING` event to the server when the user starts typing. It will emit a `STOP_TYPING` event after a 3-second pause or when the message is sent. The server aggregates these states and broadcasts a single, consolidated `TYPING_STATE_UPDATE` event to all clients.

#### **C. The Chat Input Area**

This is the player's primary tool for action. Its state must adapt to the game phase.

*   **Standard State (Discussion Phase):** A clean, single-line input field with a contextual placeholder: `Message #war-room`. Pressing `Enter` sends the message; `Shift+Enter` creates a new line.
*   **Locked State (Non-Discussion Phases):** The input field is `disabled` with placeholder text explaining why: `Channel locked during Night Phase`.
*   **Contextual State (Voting/Action Phases):** The entire input area is replaced by the relevant UI (e.g., the `VoteUI` component).

#### **D. Message Interactions: The "Shallow Reply" & Emoji System**

To facilitate direct responses and non-verbal communication, we use a unified interaction model.

*   **UI:** On hover, each message reveals a toolbar with two primary actions: `Reply` and `React with Emoji`.
*   **The "Shallow Reply" Flow:**
    1.  A player clicks the `Reply` button on a message from "Alice."
    2.  The Chat Input Area gains focus, with a "Replying to Alice" indicator above it.
    3.  The player types their response and hits `Enter`.
    *   **Message Format (Backend):** The client constructs a message using a BBCode-style `[quote]` tag. Example: `[quote=Alice]That's exactly what a System Shock looks like.[/quote] I disagree, I think you're misinterpreting the log.`
    *   **Rendered Output (Frontend):** The UI parses this and renders a visual reply, with the quoted text truncated and styled distinctly above the new message.
    *   **Reply Chaining:** A player *can* reply to a message that is itself a reply. The system will quote the immediate parent message's text, but will not render the grandparent quote. This keeps replies one level deep visually, maintaining the flat structure of the chat.

*   **The Strategic Emoji Reaction System:**
    *   **UI:** Clicking the `React` button opens a small, curated emoji picker (e.g., `👍`, `👎`, `🤔`, `👀`, `😂`, `🔥`).
    *   **Action Flow:** When a player reacts to a message, this is treated as a special type of reply.
    *   **Message Format (Backend):** The client sends a `POST_CHAT_MESSAGE` action, but with a special format: `[react to=msg-id]thinking_face[/react]`.
    *   **Rendered Output (Frontend):** The client receives this event, finds the message with `msg-id`, and visually appends the emoji reaction to it. It aggregates counts (e.g., `🤔 3`). The client also stores who reacted with what, visible on hover over the reaction bubble.
    *   **Strategic Importance:** This makes reactions part of the permanent, auditable game log. They are not just metadata; they are recordable actions. This allows players (and the AI) to analyze "who reacted to what" as a core deduction mechanic.

## 3. Formatting & Rich Content

To support complex arguments and evidence-sharing, the chat must support a limited, curated set of rich formatting.

*   **Markdown:**
    *   `**bold**` for emphasis.
    *   `*italic*` for tone.
    *   `~strike~` for corrections or ironic effect.
    *   `* list item` for bullet points.
*   **Code Blocks:** For sharing "evidence" or logs.
    *   ` `inline code` ` for mentioning player names, roles, or commands. This is critical for clear communication.
*   **@Mentions:** Typing `@PlayerName` highlights the mention for the mentioned user with a distinct background color, ensuring they see it.
*   **System-Only Tags:** The `[quote]` and `[react]` tags are generated by the client UI and parsed for rendering; they are not intended for users to type directly.

## 4. Command Palette & Shortcuts (`Cmd/Ctrl + K`)

This is the power-user interface for players who prefer keyboard-driven actions.

*   **UI Design:** A simple, searchable modal that appears centered on the screen.
*   **Core Commands:**
    *   `/vote [player]`
    *   `/status [message]`
    *   `/dossier [player]`
    *   `/help`

## 5. Notification & "Unread" Logic

The UI must intelligently handle new information to prevent players from missing critical events.

*   **Unread Channel Indicator:** The `#war-room` channel displays a badge for unread messages.
*   **"Unread Messages" Marker:** A horizontal line appears in the chat log to separate read from unread messages.
*   **Browser Tab Notification:** The page `<title>` is updated (e.g., `(1) Alignment`) when the window is out of focus.

By implementing this specific, linear communication model, we maintain a clear, auditable timeline of events, forcing all strategic conversations—including non-verbal reactions—into a single, high-stakes arena.