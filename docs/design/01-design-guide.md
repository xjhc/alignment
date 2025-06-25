# Design Guide: Alignment

**Vision Statement:** To create a premium, high-stakes social deduction experience that simulates the tense, professional atmosphere of a remote corporate crisis. Our design North Star is **"An Emergency Slack Bridge for AGI Containment."**

This guide defines the principles that will elevate `Alignment` from a web game to an immersive, atmospheric thriller. The user is not playing a game; they are a senior employee fighting for control of their company, and the UI is their only tool.

---

## 1. Core Design Pillars

Every design decision must be measured against these four pillars. They are our "Corporate Mandates" for quality.

#### **Pillar 1: Diegetic Interface ("The Emergency Bridge")**
The UI is not a "game interface"; it is the application—a hastily deployed piece of corporate software designed for a SEV-1 incident. Every pixel must serve this narrative.

*   **Professional, Not Playful:** The aesthetic is "enterprise-grade tool," not "video game." We favor clean lines, structured layouts, and information density over decorative elements. The color palette is muted and professional, punctuated by sharp, clinical accent colors for status (e.g., `#f59e0b` for warnings, `#ef4444` for critical alerts).
*   **Typography as Information:** We use a strict typographic hierarchy.
    *   **UI Text:** A clean, neutral sans-serif (Inter) for all interface elements. It's legible, professional, and unassuming.
    *   **Data & System Text:** A crisp monospace font (JetBrains Mono) for anything that represents raw data: IDs, log entries, timestamps, and system messages from `Loebmate`. This reinforces the feeling of interacting with a computer system.
*   **Authentic Corporate Language:** The copy is critical. We use terms like "SITREP," "Personnel Status," "Operational Metrics," and "After-Action Debrief." The AI is not a "monster"; it's an "uncontained asset." An elimination is a "deactivation." This maintains the atmosphere of detached corporate severity.

#### **Pillar 2: "AAA" Polish & Interaction Fidelity**
Our standard for polish is not other web games, but the seamless, responsive feel of professional-grade software like Figma, Linear, or Slack, combined with the subtle feedback of a high-end game.

*   **Motion with Purpose:** Our **Motion System** is designed for clarity and efficiency.
    *   Animations are fast and precise (`150ms-300ms`). They use `ease-out` curves to feel responsive and decisive.
    *   Motion is used to explain the UI: modals slide up from the bottom, contextual menus fade in, and new items in a list stagger-reveal to draw the eye. There is no decorative or whimsical movement.
*   **Feedback is Information, Not Distraction:** Every interaction must be acknowledged.
    *   **Micro-interactions:** Buttons provide a subtle `scale(0.98)` feedback on click. Interactive elements have a clean, understated hover state.
    *   **Sound Design:** The soundscape is minimal and diegetic. It sounds like software, not a fantasy game.
        *   **UI Sounds:** Crisp, short clicks for buttons and selections.
        *   **Notifications:** A subtle, clean "chime" like a new Slack message.
        *   **Alerts:** A sharp, digital "ping" for critical alerts.
        *   **Ambiance:** A very low, almost imperceptible server hum in the background to create tension.
*   **Keyboard-First Navigation:** Senior employees live on their keyboards. The entire application must be navigable and usable without touching a mouse. This includes hotkeys for common actions (`Cmd+K` to open a command palette, `Tab` to navigate, `Enter` to submit).

#### **Pillar 3: Information Hierarchy ("Signal from Noise")**
The game is an exercise in managing information overload and paranoia. The UI's primary job is to help the player focus on what matters *right now*.

*   **Focus is Paramount:** The UI must guide the user's attention. During a vote, the voting panel should be the most visually prominent element. During discussion, the chat log takes precedence. We use subtle glows, borders, and contextual highlighting to direct focus.
*   **Data, Not Pictures:** We avoid illustrative icons in favor of clear text labels and system-like symbols (`#`, `🔒`, `⚙️`). Player avatars are simple emojis or monograms to keep the interface clean and professional, preventing it from looking like a social media app.
*   **Density is a Feature:** This is a tool for experts under pressure. We are not afraid of information-dense screens, provided the layout is structured and the hierarchy is clear. We use tables, structured lists, and data grids, not large, spacious cards with lots of empty space.

#### **Pillar 4: A Cohesive & Enforced System**
Our design system is a set of rules and tools, not suggestions. Consistency is mandatory.

*   **The Component is the Law:** The single source of truth for any UI element is its React component in Storybook, built with our stateful component philosophy. Static mockups are for initial exploration only and are immediately retired.
*   **Design Tokens as Gospel:** All colors, fonts, spacing, and radii are defined in our Design Token system. There will be no "one-off" or magic-number values in the codebase. If a value is needed, it is added to the token system first.
*   **Automated Verification:** We use **Chromatic** for visual regression testing. Any PR that causes an unintended visual change will be automatically blocked. This makes our design system self-enforcing.

---

## 2. Design Litmus Test

Every new feature or design decision must pass this test:

*   **[ ] Does it feel like a professional, high-stakes tool?**
*   **[ ] Does it enhance clarity or reduce cognitive load?**
*   **[ ] Is the interaction immediate, responsive, and satisfying?**
*   **[ ] Does it adhere to our established tokens and components, or does it justify creating a new, reusable pattern?**

By rigorously applying these principles, we will create an experience that is not just a game, but a believable simulation of corporate espionage against a rogue AI—an experience that is "buttoned-up," tense, and deeply immersive.