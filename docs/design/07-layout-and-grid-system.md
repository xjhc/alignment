# Design: Layout and Grid System

## 1. Philosophy: Predictable Structure, Clear Hierarchy

The layout of `Alignment` is foundational to its user experience. In a game driven by information, a clear and consistent spatial structure is not a "nice-to-have"—it is a core requirement for managing cognitive load and guiding player focus.

Our philosophy is built on two principles:
1.  **A Consistent Macro-Layout:** The main application panels provide a stable, predictable structure that users can rely on.
2.  **A Rhythmic Micro-Layout:** A strict, mathematical grid (the 8-point system) governs the spacing and sizing of all internal elements, creating a subtle, professional, and visually pleasing harmony.

## 2. The Macro-Layout: The Three-Panel Application Grid

The application's primary structure is a flexible grid that adapts to the user's viewport and context.

#### **A. Desktop Layout (Screens > 768px)**

*   **Default State (Two-Panel):** The default view is `[ Roster/Channels ] | [ Comms Panel ]`.
    *   The `Roster` panel has a fixed width (e.g., `260px`).
    *   The `Comms` panel is fluid, taking up the remaining space. This prioritizes the conversation.
*   **Inspection State (Three-Panel):** When the Inspector Panel is summoned, the layout becomes `[ Roster ] | [ Comms ] | [ Inspector ]`.
    *   The `Roster` and `Inspector` panels have fixed widths (e.g., `260px` and `320px` respectively).
    *   The `Comms` panel fluidly resizes to fit the space between them.

#### **B. Mobile Layout (Screens <= 768px)**

*   **View Stack Model:** The layout collapses into a single-column "view stack" to accommodate smaller screens. Only one panel is visible at a time.
    1.  **Level 1 (Root):** `Roster/Channels Panel`. Tapping a channel navigates to...
    2.  **Level 2 (Primary):** `Comms Panel`. Tapping an avatar navigates to...
    3.  **Level 3 (Detail):** `Inspector Panel`.
*   **Navigation:** Each view must have a clear "Back" button in its header to navigate up the stack (e.g., the Comms panel has a `< Roster` button).

## 3. The Micro-Layout: The 8-Point Grid System

To ensure consistent and harmonious spacing, all UI elements, margins, and padding must use dimensions that are multiples of **8 points (8px)**. This is a non-negotiable rule.

*   **Why 8 points?** It provides a flexible yet constrained set of values that scales well across devices and aligns perfectly with common icon sizes (16, 24, 32px) and base font sizes.

#### **Spacing Tokens in Practice**

Our CSS `--space-*` variables are the implementation of this system. This table defines their intended use.

| Token | Value | Semantic Use Case | Example |
| :--- | :--- | :--- | :--- |
| `--space-1` | `4px` | (Half-step) For tight clustering, like the gap between an icon and its text. | `[icon] 4px [text]` |
| `--space-2` | `8px` | **Small Gaps.** The space between closely related inline elements, like two buttons in a group. | `[Button] 8px [Button]` |
| `--space-3` | `12px`| (1.5x step) For separating distinct UI elements within a single component. | `Card Title` <br> `12px` <br> `Card Body` |
| `--space-4` | `16px`| **Standard Padding.** The default internal padding for most components like cards and panels. | `Card Border -> 16px -> Content` |
| `--space-5` | `20px`| (2.5x step) Used for grouping related components into a section. | `Section A` <br> `20px` <br> `Section B` |
| `--space-6` | `24px`| **Large Gaps.** The space between major, distinct sections on a page. | `Player List Section` <br> `24px` <br> `Objectives Section` |
| `--space-8` | `32px`| **Extra-Large Gaps.** For creating significant visual separation, like between the main header and content. | `Page Header` <br> `32px` <br> `Page Content` |

**Enforcement:** Developers must use these CSS variables (`var(--space-4)`) instead of magic numbers (`padding: 15px`). Code reviews should enforce this strictly.

## 4. Standard Layout Patterns

To increase development speed and consistency, we will define and reuse common layout patterns.

#### **A. The Card Layout**

*   **Purpose:** The standard container for a self-contained piece of content (e.g., `AbilityCard`, `ObjectiveCard`).
*   **Specification:**
    *   `background-color: var(--bg-tertiary);`
    *   `border: 1px solid var(--border);`
    *   `border-radius: var(--radius-lg);` (8px)
    *   `padding: var(--space-4);` (16px)
    *   `box-shadow: var(--shadow);` (Subtle)

#### **B. The Header + Content Layout**

*   **Purpose:** A common pattern for panels and modals.
*   **Specification:**
    *   A `header` element with `padding: var(--space-3)` and a `border-bottom`.
    *   A `content` element with `padding: var(--space-4)`.
    *   (Optional) A `footer` element with `padding: var(--space-3)` and a `border-top`, typically containing action buttons.

#### **C. The Labeled Data Layout**

*   **Purpose:** For displaying key-value information, as seen in the Dossier and SITREP.
*   **Specification:**
    *   The `Label` uses `--font-size-xs`, `font-weight-bold`, `--text-muted`, and `text-transform: uppercase`.
    *   The `Value` uses `--font-size-base` and `--text-primary`.
    *   There is a `var(--space-1)` (4px) gap between the label and its value.

By adhering to this Layout and Grid System, we ensure that the `Alignment` UI is not just a collection of components, but a cohesive, structured, and professional application that is intuitive to navigate and pleasing to the eye.