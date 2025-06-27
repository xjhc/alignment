# Feature: Communication Mechanics

This document describes the implementation design for specialized communication features that go beyond standard chat messages. These tools provide players with more nuanced ways to signal their status, share information privately, and leave a final mark on the game.

## 1. Feature: Player Status (`/status`)

*   **Overview:** A public, short-form message (max 20 characters) that is displayed next to a player's name in the daily SITREP. It's a tool for public declarations, subtle signaling, or sowing misinformation.
*   **User Flow:**
    1.  During any Day Phase, a player types `/status [message]` into the chat input (e.g., `/status Mining for Alice`).
    2.  The client intercepts this command.
    3.  A `UPDATE_STATUS` action is sent to the server.
*   **System Flow:**
    1.  The `GameActor` receives the `UPDATE_STATUS` action.
    2.  It validates the message length and that it's a Day Phase.
    3.  It updates the `StatusMessage` field on the corresponding `Player` object in its in-memory `GameState`.
    4.  No event is immediately broadcast. The updated status is revealed to all players in the *next* day's `SITREP` event.
*   **API & Payloads:**
    *   **Action:** `UPDATE_STATUS`
    *   **Payload:** `{ "status": "Mining for Alice" }`

## 2. Feature: Whisper

*   **Overview:** A once-per-day private message. The action of whispering is public, but the content is secret, creating a powerful social deduction tool.
*   **User Flow:**
    1.  During any Day Phase, a player clicks a "Whisper" icon on another player's card or uses the `/whisper [player] [message]` command.
    2.  A private input modal appears. The player types their short message and clicks "Send."
*   **System Flow:**
    1.  The client sends a `WHISPER` action to the server.
    2.  The `GameActor` validates:
        *   Is it a Day Phase?
        *   Has this player already used their whisper for the day?
        *   Is the target player alive?
    3.  If valid, the server generates **two** events:
        *   **Public Event:** A `CHAT_MESSAGE_POSTED` event is broadcast to **all players** with a system-style message: `Vex whispers to Astra.` This makes the act of whispering public knowledge.
        *   **Private Event:** A `PRIVATE_NOTIFICATION` event is sent **only to the target player** containing the content of the whisper.
    4.  The server updates the sender's `Player` state to mark their whisper as used for the day.
*   **API & Payloads:**
    *   **Action:** `WHISPER`
    *   **Payload:** `{ "target_player_id": "p-astra", "message": "I think Bob is lying." }`

## 3. Feature: Parting Shot

*   **Overview:** Upon deactivation, a player can set one final, permanent status message that will be visible to all players for the remainder of the game.
*   **User Flow:**
    1.  A player is eliminated via a deactivation vote.
    2.  Their client UI transitions to an "Exit Interview" screen.
    3.  This screen presents an input field for their "Parting Shot" (max 20 characters).
    4.  The player types their message and clicks "Submit."
*   **System Flow:**
    1.  The client sends a `SUBMIT_EXIT_INTERVIEW` action.
    2.  The `GameActor` receives this action. Since the player is already deactivated, this is their final allowed action.
    3.  The server updates the `PartingShot` field on the (now deactivated) `Player` object in the `GameState`.
    4.  The server broadcasts a `PLAYER_STATUS_CHANGED` event to all clients, which includes the new `parting_shot` value.
    5.  Client UIs (like the roster panel) update to display this permanent message next to the deactivated player's name.
*   **API & Payloads:**
    *   **Action:** `SUBMIT_EXIT_INTERVIEW`
    *   **Payload:** `{ "parting_shot": "Vex is the AI!" }`