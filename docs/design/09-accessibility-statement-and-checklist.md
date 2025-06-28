# Design: Accessibility Statement & Checklist

## 1. Our Commitment to Accessibility

`Alignment` is a game for everyone. We are committed to creating an experience that is enjoyable and usable by the widest possible audience, including players with disabilities. Our goal is to meet and exceed the standards set by the **Web Content Accessibility Guidelines (WCAG) 2.1 at the AA level**.

Accessibility is not an afterthought or a separate feature; it is a core requirement of our design and development process. Every team member is responsible for ensuring their work is accessible.

This document serves as our public statement of intent and a practical guide for implementation.

## 2. Core Accessibility Pillars

Our accessibility strategy is built on four key pillars that will guide our development.

#### **Pillar 1: Keyboard-First Navigation**
A user must be able to play the entire game without using a mouse.

*   **Logical Tab Order:** All interactive elements (buttons, inputs, player cards, messages) must be reachable via the `Tab` key in a logical, predictable sequence.
*   **Visible Focus States:** The currently focused element must have a highly visible and consistent outline. We will use the `:focus-visible` pseudo-class to avoid showing outlines on mouse clicks but ensure they appear for keyboard navigators. Our standard focus outline will be `2px solid var(--accent-blue)`.
*   **Keyboard Interactions:** All actions that can be performed with a click must also be performable with the `Enter` or `Space` key when an element is focused. The `Escape` key must consistently dismiss modals or temporary views like the Inspector Panel.

#### **Pillar 2: Screen Reader & Semantic HTML**
The game must be understandable and navigable using screen readers (like VoiceOver, NVDA, or JAWS).

*   **Semantic HTML:** We will use HTML elements for their intended purpose. Buttons will be `<button>`, lists will be `<ul>` or `<ol>`, and main content areas will be wrapped in `<main>`. We will avoid using `<div>` with an `onClick` handler where a `<button>` is appropriate.
*   **ARIA Attributes:** We will use ARIA (Accessible Rich Internet Applications) attributes to provide context where HTML cannot.
    *   **Labels:** All icon-only buttons *must* have an `aria-label`. For example, the `(X)` close button will have `aria-label="Close Dossier"`.
    *   **Live Regions:** Dynamic content, such as new chat messages or phase change announcements, must be wrapped in an element with `aria-live="polite"` so screen readers announce the update without interrupting the user. Critical alerts may use `aria-live="assertive"`.
*   **Alternative Text:** While our game is not image-heavy, any meaningful images (if added in the future) must have descriptive `alt` text. Decorative images will have an empty `alt=""`.

#### **Pillar 3: Visual & Color Accessibility**
The UI must be legible and understandable for users with various forms of vision impairment.

*   **Color Contrast:** All text content must meet the **WCAG AA standard of 4.5:1** contrast ratio against its background. Interactive elements must meet a **3:1** ratio. This is a strict requirement, and we will use tools to verify our color palette.
*   **Information is Not Conveyed by Color Alone:** Color will be used to *enhance* information, but never as the *only* way to convey it. For example, an error state on an input field will have both a red border (`--color-danger`) *and* an error icon and text. A player's "Aligned" status will be indicated by color *and* a unique icon or text label.

#### **Pillar 4: Cognitive Accessibility**
The game is inherently complex. The UI must strive to reduce unnecessary cognitive load.

*   **Clear and Consistent Layout:** Our Layout & Grid System provides a predictable structure.
*   **Understandable Language:** We will avoid overly complex or ambiguous jargon in UI instructions.
*   **Timed Actions:** Timers for critical actions (like voting) must be clearly displayed. We will consider adding an accessibility option to slightly extend these timers for players who require more time to read or interact.

## 3. The Accessibility Pull Request Checklist

To enforce these principles, the following checklist **must be copied into the description of every Pull Request** that involves UI changes. The author is responsible for verifying each item before requesting a review.

```markdown
### Accessibility (A11y) Checklist

- [ ] **Keyboard Navigation:** I have tested this feature using only the `Tab`, `Enter`, `Space`, and `Esc` keys. All interactive elements are reachable and usable.
- [ ] **Focus States:** All interactive elements have a clear and visible `:focus-visible` state that meets our design standards.
- [ ] **Screen Reader Support:** All new elements are semantically correct. All icon-only buttons have an `aria-label`. Dynamic content is announced via an `aria-live` region.
- [ ] **Color Contrast:** I have checked all new text and background color combinations using a contrast checker, and they meet the WCAG AA 4.5:1 ratio.
- [ ] **No Color-Only Information:** All information conveyed with color is also available through text, icons, or other visual indicators.
```

## 4. Automated Testing

We will leverage tooling to automatically catch common accessibility issues before they reach production.

*   **`@storybook/addon-a11y`:** This addon is already installed. It runs `axe-core` on every story in Storybook, flagging violations di/11-game-economy-and-balance-sheet.md
---

### New File: `docs/design/09-accessibility-statement-and-checklist.md`

```markdown
# Design: Accessibility Statement & Checklist

## 1. Our Commitment to Accessibility

`Alignment` is a game for everyone. We are committed to creating an experience that is enjoyable and usable by the widest possible audience, including players with disabilities. Our goal is to meet and exceed the standards set by the **Web Content Accessibility Guidelines (WCAG) 2.1 at the AA level**.

Accessibility is not an afterthought or a separate feature; it is a core requirement of our design and development process. Every team member is responsible for ensuring their work is accessible.

This document serves as our public statement of intent and a practical guide for implementation.

## 2. Core Accessibility Pillars

Our accessibility strategy is built on four key pillars that will guide our development.

#### **Pillar 1: Keyboard-First Navigation**
A user must be able to play the entire game without using a mouse.

*   **Logical Tab Order:** All interactive elements (buttons, inputs, player cards, messages) must be reachable via the `Tab` key in a logical, predictable sequence.
*   **Visible Focus States:** The currently focused element must have a highly visible and consistent outline. We will use the `:focus-visible` pseudo-class to avoid showing outlines on mouse clicks but ensure they appear for keyboard navigators. Our standard focus outline will be `2px solid var(--accent-blue)`.
*   **Keyboard Interactions:** All actions that can be performed with a click must also be performable with the `Enter` or `Space` key when an element is focused. The `Escape` key must consistently dismiss modals or temporary views like the Inspector Panel.

#### **Pillar 2: Screen Reader & Semantic HTML**
The game must be understandable and navigable using screen readers (like VoiceOver, NVDA, or JAWS).

*   **Semantic HTML:** We will use HTML elements for their intended purpose. Buttons will be `<button>`, lists will be `<ul>` or `<ol>`, and main content areas will be wrapped in `<main>`. We will avoid using `<div>` with an `onClick` handler where a `<button>` is appropriate.
*   **ARIA Attributes:** We will use ARIA (Accessible Rich Internet Applications) attributes to provide context where HTML cannot.
    *   **Labels:** All icon-only buttons *must* have an `aria-label`. For example, the `(X)` close button will have `aria-label="Close Dossier"`.
    *   **Live Regions:** Dynamic content, such as new chat messages or phase change announcements, must be wrapped in an element with `aria-live="polite"` so screen readers announce the update without interrupting the user. Critical alerts may use `aria-live="assertive"`.
*   **Alternative Text:** While our game is not image-heavy, any meaningful images (if added in the future) must have descriptive `alt` text. Decorative images will have an empty `alt=""`.

#### **Pillar 3: Visual & Color Accessibility**
The UI must be legible and understandable for users with various forms of vision impairment.

*   **Color Contrast:** All text content must meet the **WCAG AA standard of 4.5:1** contrast ratio against its background. Interactive elements must meet a **3:1** ratio. This is a strict requirement, and we will use tools to verify our color palette.
*   **Information is Not Conveyed by Color Alone:** Color will be used to *enhance* information, but never as the *only* way to convey it. For example, an error state on an input field will have both a red border (`--color-danger`) *and* an error icon and text. A player's "Aligned" status will be indicated by color *and* a unique icon or text label.

#### **Pillar 4: Cognitive Accessibility**
The game is inherently complex. The UI must strive to reduce unnecessary cognitive load.

*   **Clear and Consistent Layout:** Our Layout & Grid System provides a predictable structure.
*   **Understandable Language:** We will avoid overly complex or ambiguous jargon in UI instructions.
*   **Timed Actions:** Timers for critical actions (like voting) must be clearly displayed. We will consider adding an accessibility option to slightly extend these timers for players who require more time to read or interact.

## 3. The Accessibility Pull Request Checklist

To enforce these principles, the following checklist **must be copied into the description of every Pull Request** that involves UI changes. The author is responsible for verifying each item before requesting a review.

```markdown
### Accessibility (A11y) Checklist

- [ ] **Keyboard Navigation:** I have tested this feature using only the `Tab`, `Enter`, `Space`, and `Esc` keys. All interactive elements are reachable and usable.
- [ ] **Focus States:** All interactive elements have a clear and visible `:focus-visible` state that meets our design standards.
- [ ] **Screen Reader Support:** All new elements are semantically correct. All icon-only buttons have an `aria-label`. Dynamic content is announced via an `aria-live` region.
- [ ] **Color Contrast:** I have checked all new text and background color combinations using a contrast checker, and they meet the WCAG AA 4.5:1 ratio.
- [ ] **No Color-Only Information:** All information conveyed with color is also available through text, icons, or other visual indicators.
```

## 4. Automated Testing

We will leverage tooling to automatically catch common accessibility issues before they reach production.

*   **`@storybook/addon-a11y`:** This addon is already installed. It runs `axe-core` on every story in Storybook, flagging violations directly in the developer's browser. All stories must have zero violations in the "Accessibility" tab.
*   **CI Integration:** We will add a script to our CI pipeline (e.g., using `jest-axe` or `cypress-axe`) to run automated accessibility checks on key pages as part of our test suite. A PR with critical accessibility violations will be blocked from merging.

By integrating accessibility into our core workflow, from design to deployment, we commit to building a high-quality, inclusive product for all players.
```