# MCP Interface: A Read-Only API for the AI

## 1. Role and Philosophy

The **Model Context Protocol (MCP)** serves a single, specific purpose in our architecture: to provide a **secure, structured, and read-only API** for our AI's "Language Brain" (the LLM) to access game state.

We use MCP as a formal contract. It ensures the LLM can get the context it needs to generate believable chat, without giving it the power to directly affect the game state or cheat. The LLM is a consumer of data, not an actor in the system.

Our backend implements the `McpServer` specification, and the Language Brain uses an `McpClient` to interact with it.

---

## 2. Resource Definition: The `GameState`

We expose a single, primary resource to the LLM.

*   **URI Template:** `game://alignment/{game_id}`
*   **Example URI:** `game://alignment/g-f4b1`
*   **Description:** "Provides the real-time public state of a specific game of Alignment. This context is used to generate conversational responses."

#### **Resource Content:**

The content of this resource is a JSON object representing a filtered view of the `GameState`. It includes all information a human player would have access to, such as:

*   `your_player_id`: The ID of the AI player itself.
*   `current_phase`: The current game phase (e.g., `DAY_DISCUSSION`, `NIGHT`).
*   `day_number`: The current day in the game.
*   `players`: A map of all players, including their public information (ID, name, token count, alive status).
*   `chat_log`: The full history of public chat messages.
*   `crisis_event`: The details of the current day's crisis event.
*   *Other public game state information as needed.*

Critically, it **does not** include any secret information like true alignments, other players' roles, or the AI's own `AI Equity` scores.

---

## 3. Capabilities: Read-Only, No Tools

A key aspect of our implementation is its limited scope. During the `initialize` handshake, our `McpServer` declares its capabilities to the client.

```json
// Part of the server's response to the `initialize` request
"capabilities": {
  "resources": {
    "templates": [
      {
        "uri": "game://alignment/{game_id}",
        "description": "Provides the real-time state of a game of Alignment."
      }
    ]
  },
  "tools": {},
  "prompts": {}
}
```

The `tools` object is intentionally empty. This explicitly tells the LLM client: **"You have no functions to call. You cannot perform any actions. You can only read the resources I provide."**

---

## 4. The Interaction Flow

The interaction between the AI's Language Brain and the MCP server is straightforward and event-driven.

1.  **Trigger:** The backend's **Strategic Core** decides the AI should consider speaking.
2.  **Request:** The `McpClient` (on behalf of the Language Brain) sends a `request` message to our `McpServer` for the `game://alignment/{game_id}` resource.
3.  **Response:** The `McpServer` immediately returns a `response` message containing the current `GameState` JSON object.
4.  **Prompt & Generation:** The Language Brain takes this JSON context, injects it into its system prompt, and sends the final package to the external LLM provider to generate a chat message. The LLM's only output is text.
5.  **State Updates:** Whenever the `GameState` changes due to any player's action, our `McpServer` proactively sends a `notifications/resources/updated` message to the `McpClient`. This informs the client that its cached version of the resource is stale, ensuring it always requests fresh data on its next turn.

This one-way flow of information—from the game server to the LLM—is fundamental to our AI's security and stability. It allows the LLM to be an informed social participant without being a direct mechanical actor.