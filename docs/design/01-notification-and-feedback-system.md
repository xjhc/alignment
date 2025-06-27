# Design: Notification and Feedback System

## 1. Philosophy: The Right Information, at the Right Time, in the Right Place

In a game as information-dense as `Alignment`, how we communicate state changes to the player is as important as the changes themselves. A poorly designed notification system can lead to confusion, frustration, and players missing critical game events.

Our philosophy is to use a **tiered system of feedback**. The visual prominence and intrusiveness of a notification will be directly proportional to its urgency and importance to the player's immediate strategy.

## 2. The Notification Tiers

Every piece of feedback in the UI falls into one of four distinct tiers.

---

#### **Tier 1: Toasts (High Priority, Ephemeral)**

*   **Purpose:** To deliver urgent, time-sensitive, and **personal** information that requires the player's immediate attention. Toasts are for events that have just happened *to you*.
*   **UI:** A small, self-dismissing banner that appears in a consistent screen location (e.g., top-right corner). They use color and icons to convey their intent at a glance.
*   **Interaction:** Toasts automatically fade out after 5-8 seconds. High-priority toasts may require a manual dismissal (`X` button). They should never stack more than three high.

**Use Cases & Content:**

| Event Trigger | Icon | Color | Toast Title | Toast Message |
| :--- | :--- | :--- | :--- | :--- |
| **Your Night Action is Blocked** | 🚫 | `danger` | Action Blocked | Your night action failed. Another operative interfered with your work. |
| **You are targeted by Conversion** | ⚡ | `magenta` | System Shock | **[ALERT]** Unauthorized intrusion repelled. Cognitive artifacts detected. |
| **Your Role Ability Unlocks** | ✨ | `info` | Ability Unlocked | You have unlocked **IsolateNode**. You can now use it during the Night Phase. |
| **Your KPI is Completed** | 🏆 | `success` | KPI Complete | You have completed your "Guardian" objective. Reward unlocked. |

---

#### **Tier 2: In-Line System Messages (Medium Priority, Permanent)**

*   **Purpose:** To communicate major, public game-state changes that affect all players. These messages become a permanent part of the official game log in the `#war-room`.
*   **UI:** A visually distinct message block within the chat log, clearly attributed to a system entity. They are not ephemeral and can be scrolled back to for review.
*   **Interaction:** They are part of the chat log and can be replied to or reacted to.

**Use Cases & Content:**

| Event Trigger | Component | Key Information Conveyed |
| :--- | :--- | :--- |
| **Start of Day Phase** | `SitrepMessage` | Announces the day number, night action summary, and the daily Crisis Event. |
| **End of a Vote** | `VoteResultMessage` | Displays the final token-weighted tally, the outcome, and the target's revealed role/alignment. |
| **Reveal of Pulse Check** | `PulseCheckMessage` | Displays all player responses publicly, with attribution. |
| **Game Ends** | `Victory/Defeat Message` | Announces the winning faction and the reason for their victory. |

---

#### **Tier 3: `Loebmate` Assistant Whispers (Contextual Help, Private)**

*   **Purpose:** To provide players with contextual, just-in-time help and guidance about the current game phase. This is the primary onboarding tool.
*   **UI:** A private "whisper" message from `Loebmate`, visually distinct and marked with a 🔒 icon. These appear only to the individual player in their chat log and are not part of the public record.
*   **Interaction:** These messages are dismissible. This entire feature can be disabled in the settings menu (`"Loebmate Assistant: On/Off"`).

**Use Cases & Content:**

| Event Trigger | `Loebmate` Whisper Content |
| :--- | :--- |
| **Start of Nomination Phase**| `[PRIVATE from Loebmate]` **New Action Unlocked: Nominate!** The discussion is over. It's time to choose who to put on trial. Click the "Nominate" button next to a player's name in the roster to cast your vote. |
| **Start of Night Phase**| `[PRIVATE from Loebmate]` **Welcome to the Night Phase!** The main channel is locked. You have 30 seconds to secretly choose an action from the menu below. "Mine for Tokens" helps your team, while "Project Milestones" helps you unlock your ability. Choose wisely! |
| **First time receiving Tokens**| `[PRIVATE from Loebmate]` **You've received new Tokens!** Tokens increase your voting power. The more you have, the more your vote counts. They are also your defense against AI conversion. |

---

#### **Tier 4: Subtle Indicators (Low Priority, Ambient)**

*   **Purpose:** To provide passive, non-intrusive feedback about the state of the UI or game without interrupting the player's flow.
*   **UI:** Small, contextual changes to existing UI elements, often using color, small icons, or glowing effects.
*   **Interaction:** Mostly non-interactive, though some might have a tooltip on hover for more context.

**Use Cases & Content:**

| Element | State Change | Indicator Style | Tooltip on Hover |
| :--- | :--- | :--- | :--- |
| **Channel List** | New unread message in `#war-room`. | Channel name becomes bold and brighter. A small white dot appears to its left. | `Unread messages` |
| **Vote Button** | Player has cast their vote. | The button becomes disabled and its style changes to a "selected" state. | `You have voted for [Player/Option].` |
| **Typing Indicator** | Other players are typing. | Text appears at the bottom of the chat: `Eve is typing...` | N/A |
| **Browser Tab** | New message while window is out of focus. | The page `<title>` prepends with `(1)`. | N/A |