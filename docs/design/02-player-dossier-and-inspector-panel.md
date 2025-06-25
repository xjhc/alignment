# Design: The Inspector Panel & Player Dossier

## 1. Philosophy: Context is King

The `Alignment` UI is designed around a core principle: **the primary user focus is the conversation**. All other views are secondary and contextual. The Inspector Panel is the main tool for detailed analysis, providing a "deep dive" on a selected player. It must be designed to be accessed and dismissed with minimal friction, ensuring players can quickly reference information and return to the main chat flow.

This document defines the behavior of this panel across different viewports.

## 2. Desktop UI/UX (Screens > 768px)

On larger screens, the design prioritizes showing both the conversation and the inspection details simultaneously.

#### **A. The Default Two-Panel Layout**

*   By default, the screen is a **two-panel layout**: `[ Roster/Channels ] | [ Comms Panel ]`.
*   The `Comms` panel, containing the chat log and input area, takes up the majority of the screen width. This is the primary "game board."

#### **B. Summoning the Inspector Panel**

*   When a player's details are requested, the Inspector Panel appears as a **third panel on the right-hand side**.
*   **Behavior:** The `Comms` panel **resizes** to make space for the Inspector Panel. The layout becomes `[ Roster ] | [ Comms (narrower) ] | [ Inspector ]`. There is no mandatory "slide-in" animation; the resizing should be fast and responsive, though a gentle `transition` on the `width` property can smooth the change.
*   **Triggering Actions:**
    1.  **Clicking** on a player's `PlayerCard` in the Roster.
    2.  **Clicking** on a player's `Avatar` in a chat message.
    3.  Using the `/dossier [player_name]` command.
    4.  Using the `D` keyboard shortcut when a player is in focus.

#### **C. Dismissing the Inspector Panel**

*   When the panel is dismissed, the layout reverts to the default two-panel view, with the `Comms` panel expanding to its original width.
*   **Dismissal Actions:**
    1.  Clicking the `(X)` close button in the Inspector Panel's header.
    2.  Pressing the `Escape` key.
    3.  Clicking on the currently inspected player's card or avatar again.

## 3. Mobile UI/UX (Screens <= 768px)

On mobile, the design prioritizes a clean, focused view, mimicking the navigation patterns of modern chat applications.

#### **A. The View Stack**

*   The mobile experience is a **stack of views**, not a persistent grid. Only one view is visible at a time.
*   **Default View:** The `Comms` panel is the default, primary view. It features a "Back" button (e.g., `< Roster`) in its header to return to the channel list.
*   **Level 1 (Root):** The `Roster/Channels` panel. Clicking a channel navigates to Level 2.
*   **Level 2 (Primary):** The `Comms` panel for the selected channel.
*   **Level 3 (Detail):** The `Inspector Panel` (Dossier view).

#### **B. Summoning and Dismissing the Inspector Panel**

*   **Behavior:** When a player's details are requested, the Inspector Panel **overlays and completely replaces** the `Comms` panel. It becomes the active view.
*   **Triggering Actions:**
    1.  Tapping on a player's `Avatar` in a chat message.
    2.  Tapping on a player's `PlayerCard` in the Roster (which would be accessed by navigating back to the Roster view first).
*   **Dismissal:** The Inspector Panel has its own header with a "Back" button (e.g., `< Chat`) that returns the user to the `Comms` panel.

## 4. Content of the Inspector Panel

The *content* of the Inspector Panel is identical across desktop and mobile and follows the "dual-mode" logic.

#### **Mode 1: "My Terminal" (Inspecting Yourself)**
*   **Trigger:** Inspecting your own player entity.
*   **Content:** A private HUD containing secret information: your `Alignment`, `Personal KPI`, and `AI Equity`.

#### **Mode 2: "Player Dossier" (Inspecting Others)**
*   **Trigger:** Inspecting any other player.
*   **Content:** A public "personnel file" with all secret information redacted.

### Information Architecture

The content is structured consistently, with conditional rendering for private fields.

| Section | Content | "My Terminal" View (You) | "Player Dossier" View (Other) |
| :--- | :--- | :--- | :--- |
| **1. Identity** | Avatar, Name, Job Title, Status Message | ✅ Displayed | ✅ Displayed |
| **2. Metrics** | Token Count, Project Progress | ✅ Displayed | ✅ Displayed |
| **3. Role** | Role Name, Description, Ability Status | ✅ Displayed | ✅ Displayed |
| **4. Alignment** | Your Alignment (Human/Aligned/AI) | ✅ **Visible** | ❌ **Hidden** |
| **5. Objectives** | Personal KPI and description | ✅ **Visible** | ❌ **Hidden** |
| **6. Game History** | Voting & Action Log | ❓ **Conditionally Visible** | ❓ **Conditionally Visible** |

*Note on Game History: This section is hidden by default and only appears for all players if a specific game rule (e.g., "Total Transparency Initiative" Mandate) makes it public.*

---

This refined model provides a clear, professional, and platform-aware specification for the Dossier/Inspector panel, ensuring it functions as an intuitive and powerful tool for players on any device.