# Development: The Design Token System

This document outlines the formal system for managing and distributing design properties in the `Alignment` project. This system is the evolution of our CSS variables into a platform-agnostic, single source of truth.

## 1. Philosophy & Goal

Our goal is to create a design system that is **scalable, consistent, and automated**. While CSS variables are a good starting point, they are an implementation detail specific to the web. A true design system is platform-agnostic.

**Design Tokens** are the solution. They are the abstract, named entities that represent our core design decisions (e.g., "the primary brand color," "the standard body font size"). We define these decisions once, in a structured format, and then use tooling to automatically compile them into whatever format our target platforms need.

For V1, this means automatically generating our `global.css` file. In the future, it could mean generating styles for an iOS or Android app from the same source.

## 2. The Chosen Tool: Style Dictionary

We will use [Style Dictionary](https://github.com/style-dictionary/style-dictionary), the industry-standard, open-source tool for building and distributing design tokens.

**Why Style Dictionary?**
*   **Platform Agnostic:** It can compile tokens into dozens of formats (CSS variables, JavaScript objects, Swift/Kotlin code, etc.).
*   **Structured:** It enforces a clear, hierarchical structure for defining tokens.
*   **Powerful Transforms:** It can automatically transform token values (e.g., convert a hex color to RGBA, or convert a pixel value to `rem`).

## 3. The Implementation Workflow

The entire system revolves around a simple, three-step, automated process.

#### **Step A: Define Tokens (The Source of Truth)**

All design tokens will live in a new top-level directory: `/tokens`. They will be defined in structured JSON files.

**Example Token Structure (`/tokens/color/base.json`):**
```json
{
  "color": {
    "brand": {
      "primary": { "value": "#f59e0b" },
      "secondary": { "value": "#06b6d4" }
    },
    "background": {
      "primary": {
        "light": { "value": "#ffffff" },
        "dark": { "value": "#0f172a" }
      },
      "secondary": {
        "light": { "value": "#f8fafc" },
        "dark": { "value": "#1e293b" }
      }
    }
  }
}