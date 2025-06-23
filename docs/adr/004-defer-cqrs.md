# **ADR-004: Defer Full CQRS and Event-Sourcing for V1**

*   **Status:** Accepted

### Context

The current architecture uses a manager-based approach (`LobbyManager`, `SessionManager`) to orchestrate the game lifecycle. While functional, it has some tight coupling and potential race conditions, especially during the lobby-to-game transition. A "super robust" architecture using CQRS, an event bus, and process managers was proposed to solve these issues definitively. This would provide maximum resilience and scalability. However, this architecture also introduces significant complexity (more components, asynchronous debugging, eventual consistency) that may be premature for our initial deployment target.


### Decision

We will **not** implement the full CQRS and event-sourcing architecture for the V1 release. We will stick to the existing, simpler actor-and-manager model. The focus for V1 will be on making the *current* architecture as robust as possible within its own paradigm. This includes:
    1.  Merging the `LobbyManager` and `SessionManager` into a unified `GameLifecycleManager` to clarify ownership and simplify handoffs.
    2.  Implementing robust garbage collection for stale lobbies and abandoned games.
    3.  Hardening the existing components with stricter validation, rate limiting, and resource caps.

**Future Path:** The full CQRS architecture is the designated, well-understood path for a future V2 or when scaling beyond a single machine becomes a business requirement. This decision is a **deferral, not a rejection**, of the pattern. We will create a new document, `docs/architecture/12-future-cqrs-refactoring-path.md`, to serve as the blueprint for this future refactoring, ensuring that current development does not build away from this intended goal.

### Consequences

*   **Pros:**
    *   **Reduced Complexity:** The V1 codebase remains smaller and easier to reason about, with a more straightforward, synchronous-style call flow. This accelerates initial development and simplifies onboarding.
    *   **Faster Time to Launch:** We avoid a major, time-consuming refactoring effort before the initial launch, allowing us to focus on implementing core game features.
    *   **Sufficient for V1:** The current architecture, once hardened, is more than sufficient to meet the performance and stability goals of running on a single VM for a limited number of concurrent games. We are not paying a "complexity tax" for scale we don't yet need.

*   **Cons:**
    *   **Technical Debt (Accepted):** We are consciously accepting a small amount of architectural technical debt. The known issues of tight coupling and potential race conditions in the current model will need to be managed through careful coding and testing, rather than being solved by the architecture itself.
    *   **Less Resilient to Crashes:** The system will not have the atomic guarantees of the outbox pattern. A server crash at the wrong moment could still lead to a rare state inconsistency (e.g., an event being persisted but not broadcast). This is an accepted risk for V1.
    *   **More Difficult Future Refactoring:** The longer we build on the current architecture, the more work it will be to refactor to CQRS in the future. Documenting the future path is critical to mitigating this.
