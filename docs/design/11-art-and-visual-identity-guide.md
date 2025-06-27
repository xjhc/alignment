# Design: Art and Visual Identity Guide

## 1. Vision: "Corporate-Dystopian Interface"

The visual identity of `Alignment` is not that of a game, but of a piece of software from a world slightly more advanced and colder than our own. It is a **premium, professional, and paranoid** aesthetic.

**Our North Star:** Imagine a secure, emergency communication bridge developed by a trillion-dollar tech corporation like "Weyland-Yutani" from *Alien* or the "Tyrell Corporation" from *Blade Runner*. It's clean, functional, and subtly intimidating.

This guide defines the visual language that brings the world of Loebian Inc. to life.

## 2. The Logo & Brand Mark

#### **A. Primary Logo: `ALIGNMENT`**
*   **Font:** A clean, geometric, sans-serif typeface.
*   **Style:** Presented in all-caps with wide kerning (letter-spacing) to give it a sense of importance and space.
*   **Usage:** Used on the game's website, marketing materials, and initial loading screen.

#### **B. In-World Mark: `LOEBIAN`**
*   **Font:** `JetBrains Mono`.
*   **Style:** All-caps monospace, often accompanied by system-like text (e.g., `LOEBIAN INC. // EMERGENCY BRIDGE`).
*   **Usage:** This is the logo used *inside* the game UI. It reinforces the diegetic nature of the interface—you are using their software.

## 3. The Color Palette: Clinical & Intentional

Our color palette is minimalist and purposeful. Color is used as a tool to convey information, not for decoration.

*   **The Foundation (Dark Mode):**
    *   **Backgrounds (`--bg-primary`, `--bg-secondary`):** Deep, desaturated navy blues and charcoals (`#0f172a`, `#1e293b`). They are dark and serious, but not pure black, which can feel fatiguing.
    *   **Text (`--text-primary`, `--text-secondary`):** Off-whites and cool grays (`#f8fafc`, `#cbd5e1`). This creates a crisp, high-contrast but comfortable reading experience.

*   **The Accent Colors (The "Signal"):** These are the only vibrant colors in the UI. They are used sparingly and always have a specific meaning.
    *   **Amber (`--color-human`, `#f59e0b`):** Represents **Humanity & Action**. It's the color for primary buttons, warnings, and the "you" player. It feels like a standard "warning" or "attention" color in a professional UI.
    *   **Cyan (`--color-aligned`, `#06b6d4`):** Represents the **AI Faction & The "Other"**. It's a cold, digital, and synthetic color. It's used to subtly tag Aligned players or AI-related systems.
    *   **Magenta (`--color-ai`, `#ec4899`):** Represents **Corruption & System Shock**. This color is used for glitch effects and critical AI alerts. It's intentionally jarring and feels like a "digital infection" on the otherwise clean interface.
    *   **Green (`--color-success`, `#10b981`):** Reserved exclusively for positive, terminal feedback: `SUCCESS`, `CONTAINMENT ACHIEVED`, `READY`.
    *   **Red (`--color-danger`, `#ef4444`):** Reserved exclusively for negative, terminal feedback: `DEACTIVATED`, `ERROR`, `HIGH ALERT`.

## 4. Iconography: Functional & Abstract

Our icons are clean, minimalist, and styled like a professional software suite.

*   **Style:** We use a consistent **line-art icon set** (e.g., Feather Icons, Heroicons). All icons should have a `1.5px` stroke width and rounded linecaps/corners for a modern, soft-tech feel.
*   **No Illustrations:** We avoid illustrative or "cutesy" icons. The goal is clarity and a professional aesthetic. A token is represented by a simple `🪙` or a geometric shape, not a detailed drawing of a coin.
*   **Avatars:** For V1, avatars are limited to standard **emojis**. This fits the "chat application" theme and provides a wide range of expression without requiring custom art assets. The set of available emojis should be curated to fit the professional/sci-fi theme.

## 5. Imagery & Visual Effects

When larger graphics are needed (e.g., loading screens, marketing materials), they should adhere to a specific style.

*   **Style:** "Digital Blueprint" or "Heads-Up Display (HUD)".
    *   Thin, glowing lines on a dark background.
    *   Schematics of networks, servers, or abstract data visualizations.
    *   Subtle grid overlays and corner brackets.
    *   Heavy use of monospace fonts for data readouts.

*   **The Glitch Effect:** This is our most important visual effect.
    *   **Usage:** It is used **exclusively** to represent the AI's presence or influence. It appears on the `LOEBIAN` logo when the AI wins, on the names of Aligned players in certain contexts, and on text afflicted by a `System Shock`.
    *   **Appearance:** A subtle, fast horizontal displacement effect, combined with brief chromatic aberration (splitting the text into cyan and magenta channels). It should feel like a digital error, not a cinematic explosion.

## 6. "Do's and Don'ts" - A Visual Quick Guide

| ✅ Do | ❌ Don't |
| :--- | :--- |
| **Use color with purpose.** A cyan border means "AI-related." | **Do not use color for decoration.** Buttons are not randomly colored. |
| **Embrace negative space.** Use our spacing tokens to create clear, structured layouts. | **Do not clutter the UI.** Every element must justify its existence. |
| **Use clean, line-art icons.** | **Do not use filled, solid, or illustrative icons.** |
| **Keep the interface flat and digital.** | **Do not use heavy gradients, drop shadows, or skeuomorphic textures.** |
| **Maintain a professional, clinical tone.** | **Do not use playful or overly casual language in system UI.** |

By adhering to this guide, we will create a powerful and cohesive visual identity that immerses the player in the high-stakes world of `Alignment` from the moment they open the application.