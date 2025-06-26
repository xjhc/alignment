# ADR-005: Adopt a "Contract-First" Approach for API and Events

*   **Status:** Accepted
*   **Supersedes:** Implicit, manual synchronization between backend and frontend.

## Context

Our initial development velocity led to a "code-first" approach for our API and WebSocket events. The Go backend's structs and the TypeScript frontend's interfaces were maintained manually. This has resulted in several critical, hard-to-debug bugs, including:

1.  **Event Type Mismatches:** The server sending an event name (e.g., `CHAT_MESSAGE_POSTED`) that the client was not explicitly coded to handle (`CHAT_MESSAGE`), causing the event to be silently dropped.
2.  **Data Mapping Bugs:** The frontend expecting a field name (`chat_messages`) that differed from the Go struct's `json` tag (`chatMessages`), leading to data being ignored.

These issues stem from relying on developer discipline to keep two separate codebases in perfect sync, a practice that is notoriously fragile.

## Decision

We will adopt a formal **"Contract-First"** approach for all communication between the Go backend and the TypeScript frontend. The Go `core` package will become the single source of truth for our API contract. We will implement automated tooling to generate TypeScript types and interfaces directly from the Go source code.

This will be achieved through two primary initiatives:

1.  **Automated Type Generation:** We will integrate a tool (e.g., `ts-generator`) into our build process. This tool will parse the Go structs in the `/core` package and automatically generate corresponding TypeScript interfaces. This ensures that field names, `json` tags, and data types are always perfectly synchronized.

2.  **Shared Event Constant Registry:** We will define all `EventType` strings as exported constants in `/core/events.go`. A build script will parse this file and generate a corresponding TypeScript `enum` or `union type`. The frontend will be refactored to use this generated enum, providing compile-time awareness of all possible events.

## Consequences

*   **Pros:**
    *   **Eliminates an Entire Class of Bugs:** It becomes impossible for backend and frontend data structures to drift out of sync. Discrepancies will now be caught at compile-time on the frontend, not as mysterious bugs at runtime.
    *   **Improved Developer Velocity:** Frontend developers no longer need to manually create or update TypeScript types. The process is automated, reducing boilerplate and human error.
    *   **Single Source of Truth:** The `/core` package becomes the definitive contract. Changes made there are automatically propagated, simplifying the development workflow.
    *   **Enhanced Reliability:** The system becomes more robust and predictable, as the on-the-wire data format is now explicitly defined and enforced by tooling.

*   **Cons:**
    *   **Increased Build Complexity:** We are adding a new step to our build process (`npm run generate:types`). This introduces a new dependency and a small amount of overhead to the dev and CI loops. This is a worthwhile trade-off for the massive gain in reliability.
    *   **Initial Implementation Cost:** There is a one-time cost to research, implement, and integrate the chosen code-generation tools.

This decision moves our project towards a more mature, professional engineering practice that prioritizes automated correctness over manual synchronization.