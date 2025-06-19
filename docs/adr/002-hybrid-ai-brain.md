
# ADR-002: Implement a Hybrid AI Brain

*   **Status:** Accepted
*   **Date:** 2023-10-26

### Context

A core feature of `Alignment` is a believable AI opponent. The initial design goal was to use a Large Language Model (LLM) for all AI behavior, including strategic decision-making (voting, targeting) and communication (chat).

Prototyping this "pure LLM" approach revealed several critical flaws:
1.  **High Latency:** Every single decision, no matter how small, required a round-trip to an external API, making the AI feel sluggish.
2.  **High Cost:** LLM API calls are metered. A chatty, active AI would be prohibitively expensive to operate at scale.
3.  **Unreliability:** LLMs are non-deterministic. They would frequently fail to follow the strict output formatting required for game actions (e.g., `{"action": "VOTE", "target": "p-123"}`) or would make strategically nonsensical moves.

### Decision

We will implement a **hybrid AI model** that separates the AI's responsibilities into two distinct components:

1.  **The Strategic Brain (Rules Engine):** A deterministic, pure Go module will be responsible for all concrete, state-changing game actions. This includes deciding who to vote for, who to target at night, and when to use abilities. Its decisions will be based on hardcoded heuristics that analyze the game state.

2.  **The Language Brain (LLM):** A Large Language Model will be used **exclusively for communication**. Its only responsibility is to generate human-like chat messages to maintain its assigned persona and engage in social deduction. It will be given read-only access to the game state via a secure MCP interface and will have no ability to execute game actions directly.

### Consequences

*   **Pros:**
    *   **Best of Both Worlds:** We get the creative, nuanced, and deceptive communication of an LLM, combined with the fast, reliable, and cost-free strategic optimization of a traditional game AI.
    *   **Drastically Reduced Cost & Latency:** The vast majority of AI decisions are now handled locally and instantly, with LLM API calls reserved only for generating chat.
    *   **Reliability & Testability:** The game-critical strategic logic is now part of our deterministic, testable Go codebase.

*   **Cons:**
    *   **Increased Implementation Complexity:** We must build and maintain two separate AI systems and a means for them to be coordinated by the main Game Actor.
    *   **Heuristic-Dependent Strategy:** The strategic competence of the AI is now entirely dependent on the quality of the heuristics we design in the Rules Engine. A poorly designed engine will result in a weak AI opponent.
