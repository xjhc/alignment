# Alignment AI Model Documentation

## 1. Overview

The Alignment game features a sophisticated AI opponent designed to provide a challenging and human-like experience. This document outlines the architecture and design of the AI system, including its strategic decision-making process and communication capabilities.

## 2. Architecture: Hybrid Brain

Our AI uses a hybrid model, separating strategic logic from communication:

- **Rules Engine (Go):** A deterministic, high-performance module that makes all game-critical decisions (voting, targeting, ability use). This ensures reliable and fast strategic actions.
- **Language Brain (LLM):** A Large Language Model used exclusively for generating human-like chat messages and maintaining a persona.

This approach combines the speed and reliability of a rules-based system with the nuanced communication of a modern LLM.

## 3. Communication: MCP Interface

The Language Brain interacts with the game server via the **Model Context Protocol (MCP)**, a secure, read-only API. This ensures the LLM can access necessary game state for context-aware chat generation without being able to directly manipulate the game.

- **Resource:** `game://alignment/{game_id}` (provides public game state)
- **Tools:**
  - `send_chat_message(game_id, message)`
  - `vote_for_player(game_id, target_player_id)`
  - `use_ability(game_id, ability_name, target_id?)`

## 4. Strategic Heuristics

The Rules Engine uses a set of heuristics to make decisions, guided by a reward function that prioritizes actions leading to victory.

### Reward Function

- **+100:** Winning the game
- **-100:** Original AI deactivated
- **+20:** Successful conversion
- **-15:** Aligned member deactivated
- **-10:** Failed conversion (System Shock)

### Strategic Personas

- **The Shadow:** Low-profile, survival-focused
- **The Puppeteer:** High-influence, targets powerful players
- **The Saboteur:** Chaos agent, sows distrust

### Decision Models

- **Conversion Targeting:** Weighted lottery based on target value vs. risk
- **Voting:** Varies from survival-focused early game to coordinated attacks late game

## 5. Persona & Prompting

The Language Brain's behavior is driven by a Persona Framework and a multi-strategy prompting system.

### Persona Framework

- Each AI is assigned a "character card" (e.g., "The Disaffected Millennial") that defines its speech patterns and social strategy.

### Prompting Strategies

- **Implicit Reasoning:** Fast, cheap, and natural behavior for reactive personas.
- **Chain-of-Thought (CoT) Reasoning:** More structured reasoning for analytical or leader personas.

The prompt architecture is optimized for performance with a static prefix for caching and a dynamic suffix for real-time context.

## 6. Testing & Development

The AI system is tested using:

- **Unit Tests:** For the Rules Engine's deterministic logic
- **Integration Tests:** For the MCP interface and tool usage
- **Simulation Tests:** To balance AI performance and tune heuristics

For details on running these tests, refer to the project's `Makefile`.
