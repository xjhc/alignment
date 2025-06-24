# Development: The Motion System

This document defines the formal Motion System for the `Alignment` application. Its purpose is to create a consistent, performant, and informative user experience by establishing a clear set of rules for how and why elements move in the UI.

## 1. Philosophy of Motion

All animations in `Alignment` should adhere to these principles:

1.  **Purposeful & Informative:** Motion must have a reason. It should guide the user's eye, show relationships between UI elements, and provide feedback on interactions.
2.  **Crisp & Efficient:** Animations should be quick and precise, reflecting the game's high-stakes, data-driven theme. No slow or whimsical animations.
3.  **Consistent:** The same interaction should always produce the same motion. This creates a predictable and intuitive rhythm for the user.
4.  **Performant:** All animations must only animate `transform` and `opacity` to ensure they are smooth and do not cause layout shifts.

## 2. The Primitives of Motion

Our system is built on a small, well-defined set of "tokens" that are combined to create all animations.

#### **Duration Tokens (The "When")**

| Token Name | CSS Variable | Value | Purpose |
| :--- | :--- | :--- | :--- |
| `duration.fast` | `--duration-fast` | `150ms` | For immediate feedback on user interaction (e.g., button press, hover effect). |
| `duration.medium`| `--duration-medium` | `300ms` | For standard UI element transitions (e.g., an element appearing/disappearing on the screen). |
| `duration.slow` | `--duration-slow` | `500ms` | For large-scale screen or panel transitions. |

#### **Easing Tokens (The "How")**

| Token Name | CSS Variable | Value | Purpose |
| :--- | :--- | :--- | :--- |
| `ease.out` | `--ease-out` | `cubic-bezier(0.25, 0.46, 0.45, 0.94)` | Standard curve for elements entering the screen. |
| `ease.in` | `--ease-in` | `cubic-bezier(0.55, 0.085, 0.68, 0.53)` | Standard curve for elements leaving the screen. |
| `ease.in-out`| `--ease-in-out` | `cubic-bezier(0.445, 0.05, 0.55, 0.95)` | For elements that transform in place (e.g., color change). |
| `ease.feedback`| `--ease-feedback`| `cubic-bezier(0.68, -0.55, 0.265, 1.55)` | "Overshoot" curve for attention-grabbing feedback. |

## 3. Animation Patterns

These patterns codify how the primitives are used for specific UI scenarios.

#### **Pattern 1: Fade In / Fade Out**

*   **Purpose:** For elements that appear or disappear in place (e.g., toasts, error messages).
*   **Tokens Used:** `duration.medium`, `ease.out` (in), `ease.in` (out).
*   **Implementation:**
    ```css
    @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }
    .animation-fade-in { animation: fadeIn var(--duration-medium) var(--ease-out) forwards; }
    ```

#### **Pattern 2: Slide In**

*   **Purpose:** For elements entering the viewport, like a new screen or a side panel.
*   **Tokens Used:** `duration.medium`, `ease.out`.
*   **Implementation:**
    ```css
    @keyframes slideInUp {
      from { opacity: 0; transform: translateY(16px); }
      to { opacity: 1; transform: translateY(0); }
    }
    .animation-slide-in-up { animation: slideInUp var(--duration-medium) var(--ease-out) forwards; }
    ```

#### **Pattern 3: Staggered List Reveal**

*   **Purpose:** To animate lists of items (e.g., the player roster) so they appear sequentially.
*   **Tokens Used:** `duration.medium`, `ease.out`.
*   **Implementation:** A combination of CSS for the animation and JavaScript to apply delays.
    ```css
    .stagger-child {
      opacity: 0;
      animation: slideInUp var(--duration-medium) var(--ease-out) forwards;
    }
    ```
    ```typescript
    // Example React Hook
    useEffect(() => {
      const items = listRef.current.querySelectorAll('.stagger-child');
      items.forEach((item, index) => {
        (item as HTMLElement).style.animationDelay = `${index * 50}ms`; // 50ms stagger delay
      });
    }, [items]);
    ```

## 4. Implementation & Enforcement

1.  **Codify Tokens in CSS:** All `duration` and `easing` variables must be defined in `:root` within `client/src/global.css`.
2.  **Centralize Animation Utilities:** All animation class names and helper functions should be defined and exported from `client/src/utils/animations.ts` to ensure consistency and prevent magic strings.
3.  **Document in Storybook:** Create a `Motion.stories.mdx` file in Storybook. This is the **living documentation** and should visually demonstrate each duration, easing curve, and animation pattern. This is the source of truth for designers and developers.