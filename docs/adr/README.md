# Architectural Decision Records (ADRs)

This directory contains the Architectural Decision Records for the `Alignment` project. An ADR is a short document that captures a significant architectural decision, including its context and consequences.

## What is an ADR?

An ADR is a "software design choice that is costly to change." We create them for decisions that have a wide-reaching impact on the codebase, our development process, or our operational strategy.

The purpose of this collection is not to document every single decision, but to create an immutable, historical log of the most critical ones. This helps us:
*   Avoid re-litigating past decisions.
*   Onboard new engineers by explaining the "why" behind our architecture.
*   Understand the trade-offs we have consciously made.

## When to Write a New ADR

You should propose a new ADR when you are considering a change that:
*   Affects the structure of the entire application (e.g., switching from monolith to microservices).
*   Introduces a new, significant technology or dependency (e.g., adding a new database, switching our Language Model provider).
*   Changes a fundamental development or deployment process.
*   Establishes a new, cross-cutting standard (e.g., a new logging or authentication pattern).

ADRs are written *before* the work is implemented, as part of the design and proposal phase.

## ADR Format

Each ADR should be a new markdown file in this directory, following the naming convention `NNN-short-title.md` (e.g., `004-switch-to-protobuf.md`).

The document should contain the following sections:

*   **Title:** A short, descriptive title for the decision.
*   **Status:** The current state of the ADR. One of:
    *   `Proposed`: A new decision under discussion.
    *   `Accepted`: A decision that the team has agreed to implement.
    *   `Deprecated`: A decision that was accepted but is no longer relevant.
    *   `Superseded by ADR-NNN`: A decision that has been replaced by a newer one.
*   **Context:** What is the problem or situation that requires a decision? What are the technical, business, or operational drivers? This section describes the "forces at play."
*   **Decision:** A clear and concise statement of the chosen approach. What exactly are we going to do?
*   **Consequences:** What are the results of this decision? This section should honestly list the pros, the cons, and any trade-offs made. It should also outline any new risks or work that this decision creates.

## Existing Decisions

Here is a list of the key architectural decisions made for this project to date.

*   **[ADR-001: Adopt an In-Memory Actor Model](./001-in-memory-actor-model.md):** The decision to use a stateful, in-memory model for the backend instead of a stateless one.
*   **[ADR-002: Implement a Hybrid AI Brain](./002-hybrid-ai-brain.md):** The decision to split the AI into a deterministic [Rules Engine](../glossary.md#rules-engine) for strategy and a [Language Model](../glossary.md#language-model) for communication.
*   **[ADR-003: Isolate Actor Failures with a Supervisor Pattern](./003-supervisor-pattern.md):** The decision to implement a [Supervisor](../glossary.md#supervisor) to prevent crashes in one game from taking down the entire server.