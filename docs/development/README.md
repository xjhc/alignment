# Development Process & Best Practices

This directory contains guides and standards for developing the `Alignment` codebase. Adhering to these practices ensures our code remains maintainable, testable, and robust.

All contributions, from small bug fixes to major features, are expected to follow these guidelines.

---

## 1. Core Philosophy

*   **Clarity and Simplicity:** Write code that is easy for the next developer to understand. A clever one-liner is worse than five readable lines of code.
*   **Test Everything That Matters:** Our stability comes from a strong foundation of automated tests. Untested code is considered broken.
*   **Documentation is Not Optional:** Changes to code must be reflected in the relevant documentation in the same pull request.

---

## 2. Key Development Guides

Before writing code, please familiarize yourself with our core development patterns and strategies.

*   **[Core Logic Definition](./01-core-logic-definition.md):**
    This document defines what we consider "Core Logic" (e.g., `ApplyEvent`, `RulesEngine`). It explains why this logic **must** be written as pure, deterministic functions, completely isolated from side effects like network or database calls. This is the foundation of our system's predictability.

*   **[Testing Strategy](./02-testing-strategy.md):**
    This is the most important guide for any developer. It outlines our multi-layered testing approach, including:
    *   **Unit Tests** for pure Core Logic.
    *   **Integration Tests** for stateful Game Actors.
    *   **System Tests** for high-level resiliency features.
    It provides concrete examples and explains what is expected for a pull request to be considered "tested."

---

## 3. Code Style and Linting

*   **Formatting:** We use the standard `gofmt` for all Go code.
*   **Linting:** We use `golangci-lint` to enforce a consistent style and catch common programming errors.

Our CI/CD pipeline automatically runs both `gofmt` and `golangci-lint` on every pull request. To avoid CI failures, it is highly recommended to run these tools locally before committing:

```bash
# Auto-format all Go files
gofmt -w .

# Run the linter
golangci-lint run ./...
```