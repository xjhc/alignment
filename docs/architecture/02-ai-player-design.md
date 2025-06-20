# AI Player Design: A Pragmatic Hybrid Model

## 1. Design Philosophy

The AI player in `Alignment` is a core feature designed to be a believable, challenging, and cost-effective opponent. A purely Language Model-driven approach proved to be slow, expensive, and strategically unreliable during prototyping.

Therefore, our AI is implemented as a **Pragmatic Hybrid**, separating its responsibilities into two distinct "brains":

1.  **The Rules Engine:** A deterministic, high-performance rules engine that handles all game-critical decisions.
2.  **The Language Model:** A Large Language Model that is used exclusively for social interaction and maintaining a human-like persona.

This separation allows us to leverage the strengths of each technology: the raw, logical power of code for strategy, and the nuanced, creative power of a Language Model for communication.

---

## 2. Component Architecture

The AI's logic is managed by the main `Game Actor`. **To ensure AI decision-making doesn't block the primary event processing loop, the `Game Actor` spawns a "sidecar" goroutine that runs concurrently.** This sidecar receives triggers from the main actor, decides which brain to use, and injects the resulting action back into the main actor's mailbox for processing.

```ascii
+---------------------------------------------------------------------------------+
|                                   Game Actor                                    |
|                                                                                 |
|  +---------------------------------------------------------------------------+  |
|  |                             Main Logic Loop (Hot Loop)                    |  |
|  |  ... Processes actions from players and its own internal cores ...        |  |
|  +-------------------------------------^-------------------------------------+  |
|                                        | (Injects resulting Actions)         |
|                                        |                                     |
|  +-------------------------------------+-------------------------------------+  |
|  |               Sidecar "Brain" Goroutine (Handles AI Decision-Making)      |  |
|  |                                                                           |  |
|  |  trigger = <- triggerChan                                                 |  |
|  |                                                                           |  |
|  |  IF trigger is for a GAME ACTION (Vote, Convert, Mine):                   |  |
|  |      action = StrategicBrain.DecideAction(state)                          |  |
|  |      mainMailbox <- action                                                |  |
|  |                                                                           |  |
|  |  ELSE IF trigger is for COMMUNICATION (Chat):                             |  |
|  |      action = LanguageBrain.GenerateChat(state)                           |  |
|  |      mainMailbox <- action                                                |  |
|  |                                                                           |  |
|  +---------------------------------------------------------------------------+  |
|                                                                                 |
+---------------------------------------------------------------------------------+
```

---

### 2.1 The Rules Engine

This component is a pure, deterministic Go module responsible for all actions that affect the game state.

*   **Responsibilities:**
    *   Deciding who to **Vote** for during the Day Phase.
    *   Deciding who to **Target** for conversion during the Night Phase.
    *   Deciding whether to **Mine for Tokens** or use another night action.
    *   Using any role-specific special abilities.

*   **Logic:** The engine operates on a set of heuristics, primarily `calculateThreat()` and `calculateSuspicionScore()`. These functions analyze the public `GameState` (voting records, token counts, etc.) to determine the optimal move.

*   **Benefits:**
    *   **Fast:** Decisions are made in microseconds, with zero external latency.
    *   **Free:** Runs entirely on our infrastructure with no API costs.
    *   **Reliable:** The logic is predictable, testable, and not subject to the whims of a non-deterministic model.

---

### 2.2 The Language Model

This component is responsible for making the AI *feel* human. Its only job is to generate text.

*   **Responsibilities:**
    *   Generating chat messages in response to being mentioned or to advance a social strategy.
    *   Maintaining a consistent, assigned persona (e.g., "Disaffected Millennial," "Corporate Overachiever").
    *   Deciding when to remain silent, a key strategic move.

*   **Logic:** When triggered, this brain uses the **Model Context Protocol (MCP)** to read the current `GameState`. This state is formatted into a compact, specialized prompt and sent to an external Language Model provider (e.g., Azure OpenAI). The Language Model's text response is then injected back into the game as a chat message.

*   **Interface:** The interaction is strictly controlled via our **[MCP Interface](./03-mcp-interface.md)**. The Language Model has read-only access to the game state and cannot call any tools or take any actions other than producing text.

*   **Benefits:**
    *   **Believable:** Capable of generating nuanced, creative, and deceptive language that a simple rules engine cannot.
    *   **Immersive:** Creates a compelling and unpredictable social opponent.

---

## 3. The Illusion of Humanity

To further enhance believability, we implement several key features:

*   **Dynamic Typing Simulation:** When the Language Brain is generating a response, the UI will show a "typing..." indicator. The duration of this indicator is dynamically calculated based on the word count of the generated response, simulating a realistic typing speed.
*   **Scheduled Actions:** All AI actions (both chat and strategic moves) are sent to the central **Scheduler** with a small, randomized delay. This prevents the AI from acting with inhuman speed and precision the moment a phase begins.

By combining a powerful, deterministic core for strategy with a creative, unpredictable Language Model for communication, we create an AI that is both a formidable gameplay opponent and a believable social actor.