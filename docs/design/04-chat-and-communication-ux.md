# Design: Chat & Communication UX

## 1. Philosophy: The Chat is the Game Board

In `Alignment`, the chat interface is not a secondary feature; it is the primary game board. All strategic interactions, deductions, and deceptions happen here, in a single, chronological stream of evidence.

Our design philosophy is to create a **high-signal, low-friction, professional-grade communication tool** that empowers players to focus on strategy. The UX should feel less like a casual game chat and more like a high-stakes Slack or Discord channel during a corporate crisis.

**Core Principle:** All public communication must remain in a single, flat channel (`#war-room`). This forces all arguments to happen in the open and preserves the integrity of the event timeline, making it a "single source of truth" for player deduction.

## 2. Core Components of the Chat Experience

#### **A. The Message Log (`#war-room`)**

This is the central, chronologically-ordered log of all public communication.

*   **Structure:** A compact, modern chat layout. Messages from the same author sent within a 2-minute window are grouped to reduce visual noise. A "date marker" (e.g., `--- DAY 2 ---`) clearly separates discussion from different game days.
*   **Message Anatomy:**
    *   `Avatar`: A simple, consistent visual identifier for the player.
    *   `Author Name`: Clearly legible, with a distinct color or style for the local player and AI/System messages.
    *   `Timestamp`: A subtle, low-contrast timestamp (`HH:MM`) appears on hover.
    *   `Content`: The message text, including any rendered "quoted reply" content.
*   **System Messages:** Messages from `Loebmate` (SITREPs, phase changes) must be visually distinct from player messages. They use a different background color, a brand icon, and a more structured, formal layout.

#### **B. The Typing Indicator**

To enhance real-time presence and psychological tension, players must be able to see who is currently typing.

*   **UI:** A small, unobtrusive indicator will appear at the bottom of the chat log: `Eve is typing...` or `Alice, Bob, and 2 others are typing...`.
*   **Implementation:** The client will emit a `START_TYPING` event to the server when the user starts typing. It will emit a `STOP_TYPING` event after a 3-second pause or when the message is sent.

#### **C. The Chat Input Area**

This is the player's primary tool for action. Its state must adapt to the game phase.

*   **Standard State (Discussion Phase):** A clean, single-line input field with a contextual placeholder: `Message #war-room`. `Enter` sends; `Shift+Enter` creates a new line.
*   **Locked State (Non-Discussion Phases):** The input field is `disabled` with placeholder text explaining why: `Channel locked during Night Phase`.
*   **Contextual State (Voting/Action Phases):** The entire input area is replaced by the relevant UI (e.g., the `VoteUI` component).

#### **D. Message Interactions: Reply & React**

*   **UI:** On hover, each message reveals a toolbar with two primary actions: `Reply` and `React with Emoji`.
*   **The "Shallow Reply" Flow:** Clicking `Reply` focuses the input with a "Replying to [Author]" indicator. The sent message is then rendered with a visual quote of the parent message. This keeps the chat flat while providing context.
*   **The Strategic Emoji Reaction System:** Clicking `React` opens a curated emoji picker (`👍`, `👎`, `🤔`, `👀`, `😂`, `🔥`). Reactions are aggregated on the message and are part of the permanent, auditable game log, allowing players to analyze "who reacted to what."

## 3. Formatting & Rich Content

To support complex arguments and evidence-sharing, the chat must support a limited, curated set of rich formatting.

*   **Markdown:** `**bold**`, `*italic*`, `~strike~`, and `* list item`.
*   **Code Blocks:** `` `inline code` `` for mentioning player names, roles, or commands.
*   **@Mentions:** Typing `@PlayerName` highlights the mention for that user.

## 4. Command Palette & Shortcuts (`Cmd/Ctrl + K`)

This is the power-user interface for players who prefer keyboard-driven actions.

*   **UI Design:** A simple, searchable modal that appears centered on the screen.
*   **Core Commands:**
    *   `/vote [player]`
    *   `/status [message]`
    *   `/whisper [player] [message]`
    *   `/dossier [player]`
    *   `/help`

## 5. Notification & "Unread" Logic

The UI must intelligently handle new information to prevent players from missing critical events.

*   **Unread Channel Indicator:** The `#war-room` channel displays a badge for unread messages.
*   **"Unread Messages" Marker:** A horizontal line appears in the chat log to separate read from unread messages.
*   **Browser Tab Notification:** The page `<title>` is updated (e.g., `(1) Alignment`) when the window is out of focus.