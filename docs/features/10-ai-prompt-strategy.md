# Feature: AI Prompt Strategy & Persona Design

This document details the design of the master prompts used to control the AI player's Language Model. The prompt is the source code for the AI's behavior, and its quality is paramount.

## 1. Design Goals

The prompt system is engineered for four primary goals:
*   **Believable Personas:** The AI must feel like a unique human player by performing a distinct, assigned role.
*   **Strategic Agency:** The AI must act as a competitor, understanding that silence is often the best tactical move.
*   **Flexibility:** The system must support multiple prompting strategies and allow for rapid iteration without server redeployment.
*   **Performance:** The structure must be optimized for low latency and cost.

## 2. The Persona Framework

Our core strategy is to have the AI **perform a role**. Each AI player is assigned a "character card" for the round (e.g., "The Disaffected Millennial," "The Chronically Online Gen Z"). The prompt provides detailed instructions on the character's speech patterns, vocabulary, and strategic angle. This creates a wide variety of believable AI personalities.

## 3. The Multi-Strategy System

We do not use a single prompting technique. Our system uses an in-code prompt registry that supports multiple reasoning strategies that can be mixed and matched with personas. This allows for A/B testing and greater behavioral variety.

The two primary strategies are:

1.  **Implicit Reasoning:**
    *   **Description:** This prompt gives the Language Model direct behavioral rules ("Read the room," "Agency is paramount") and relies on its advanced zero-shot capabilities.
    *   **Output:** `{ "action": "..." }`
    *   **Use Case:** Fast, cheap, and creates natural behavior. Best for reactive or quiet personas.

2.  **Chain-of-Thought (CoT) Reasoning:**
    *   **Description:** This prompt instructs the Language Model to output its internal reasoning as a "thought" before its action.
    *   **Output:** `{ "thought": "...", "action": "..." }`
    *   **Use Case:** Forces more structured strategic alignment. Best for analytical or "leader" personas.

## 4. Prompt Architecture & Optimization

All prompts are constructed to be highly efficient.

*   **Structure:** Every prompt is built from two parts: a **Static Prefix** (containing all rules and persona instructions) and a **Dynamic Suffix** (containing the immediate game context).
*   **Optimization:** This structure is "cache-friendly." By placing the large, unchanging static instructions first, we enable Language Model providers to cache its tokenized representation, significantly reducing processing latency on subsequent API calls.
*   **Source of Truth:** The entire system is implemented using an "In-Code Prompt Registry" where all prompts are defined as Go structs, as detailed in the **[Prompt Management & Delivery](../architecture/08-prompt-management-and-delivery.md)** architecture document.