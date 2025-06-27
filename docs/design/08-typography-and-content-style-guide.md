# Design: Typography and Content Style Guide

## 1. Philosophy: Clarity, Hierarchy, and Thematic Voice

In `Alignment`, text is the primary medium for both gameplay and narrative. Our typography and content strategy is built on three core principles:

1.  **Clarity:** The text must be instantly legible and unambiguous. We prioritize readability over stylistic flair.
2.  **Hierarchy:** We use a strict typographic scale to guide the user's eye, ensuring they can distinguish between headers, body text, and metadata at a glance.
3.  **Thematic Voice:** All written content, from button labels to system alerts, must reinforce the game's atmosphere of a high-stakes, professional, corporate crisis.

## 2. The Font System

Our application uses two carefully selected fonts, each with a specific purpose.

#### **A. Primary UI Font: Inter**
*   **CSS Variable:** `--font-sans`
*   **Use Case:** All interface elements, including buttons, labels, and player-generated chat messages.
*   **Rationale:** Inter is a neutral, highly-legible sans-serif designed specifically for user interfaces. Its clean, professional look is a perfect fit for our "corporate software" aesthetic.

#### **B. Data & System Font: JetBrains Mono**
*   **CSS Variable:** `--font-mono`
*   **Use Case:** Any text that represents "system data" or computer output. This includes timestamps, player IDs, log entries, token counts, and all messages from `Loebmate`.
*   **Rationale:** Using a monospace font for system-generated text creates a clear visual distinction between human communication and machine data, reinforcing the game's theme.

## 3. The Typographic Scale

Our typographic scale is based on our design tokens and provides a consistent hierarchy for all text elements.

| Use Case | Font | Size (`var`) | Weight (`var`) | Letter Spacing | Example |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Page Title** | Mono | `--font-size-3xl` (20px) | `--font-weight-bold` | `1px` | `POST-GAME ANALYSIS` |
| **Section Header** | Mono | `--font-size-xl` (16px) | `--font-weight-bold` | `0.5px` | `📋 OBJECTIVES` |
| **Card Title** | Sans | `--font-size-lg` (14px) | `--font-weight-bold` | `-0.02em` | `Isolate Node` |
| **Body / Chat** | Sans | `--font-size-md` (13px) | `--font-weight-normal` | `normal` | `I think Bob is the AI...` |
| **Input Text** | Sans | `--font-size-md` (13px) | `--font-weight-normal` | `normal` | `Message #war-room` |
| **Supporting Text**| Sans | `--font-size-base` (12px) | `--font-weight-normal` | `normal` | `Ability description text`|
| **Metadata / Label**| Mono | `--font-size-xs` (10px) | `--font-weight-bold` | `0.5px` | `INITIAL ALIGNMENT:` |

## 4. Color & Content Style Guide

This section defines the "voice" of the application and the rules for using color and formatting in text.

#### **A. Text Color Usage**
*   `--text-primary`: For all primary headers and body content. This is the default text color.
*   `--text-secondary`: For supporting, non-critical text, such as component descriptions or secondary labels.
*   `--text-muted`: For disabled text, placeholders, and low-priority metadata like timestamps.
*   **Semantic Colors:** Use `--color-human`, `--color-ai`, etc., *only* to convey specific, critical game-state information (e.g., highlighting an AI's name in red). Do not use them for decoration.

#### **B. Content Voice & Tone**

*   **Player-Facing UI:** Professional, clean, and direct. Use sentence case for all titles and labels (e.g., "Player status," not "Player Status"). Buttons should be clear calls to action (e.g., "Submit Vote," not "Vote").
*   **`Loebmate` (The System Bot):** `Loebmate`'s voice is a critical part of the theme. It should be:
    *   **Unfailingly Cheerful & Corporate:** Uses buzzwords like "synergy," "action items," "touch base," and "circle back."
    *   **Slightly Sinister:** The cheerful tone creates a jarring contrast with the high-stakes situation.
    *   **Formatted:** Always uses `monospace` font and is clearly identified with a `[BOT]` tag.
*   **Security Alerts:** The voice is the opposite of `Loebmate`. It is clinical, urgent, and uses technical language.
    *   *Example:* `[SEV-1] Critical Security Incident - Anomaly detected. Containment protocols engaged.`

#### **C. Data Formatting Rules**
*   **Timestamps:** Displayed in `HH:MM` format.
*   **Dates:** Displayed as `Day X` within the context of the game.
*   **Player Names:** Always referenced by their handle. When used in system text, they should be wrapped in backticks (`` `Vex` ``) to style them as data.
*   **Numbers:** Token counts should always be preceded by the `🪙` icon.

#### **D. Markdown and Rich Text**
The chat input supports a limited, curated set of markdown for strategic communication.

*   `**bold text**` -> **bold text** (for emphasis and accusation)
*   `*italic text*` -> *italic text* (for tone and sarcasm)
*   `~strikethrough~` -> ~strikethrough~ (for corrections and meta-commentary)
*   `* list item` -> • list item (for organizing thoughts and evidence)
*   `` `inline code` `` -> `inline code` (for quoting logs, names, and commands)

This comprehensive guide ensures that every piece of text in `Alignment`—from the smallest label to the most critical system alert—is clear, consistent, and contributes to the immersive, high-stakes atmosphere of the game.