# Alignment

> **Align the AGI. Before it aligns you.**

[![CI](https://github.com/xjhc/alignment/actions/workflows/ci.yml/badge.svg)](https://github.com/xjhc/alignment/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg?style=flat-square)](https://go.dev/doc/go1.21)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square)](https://opensource.org/licenses/MIT)
[![Architecture: Actor Model](https://img.shields.io/badge/Architecture-Actor%20Model-8A2BE2.svg?style=flat-square)](#architecture)

---

**Alignment** is a chat-based social deduction game of corporate paranoia. Players must identify a rogue AI hiding among them in a corporate chat server before it seizes control of the company.

### Architecture

This project is a real-time, stateful application built on a modern Go stack designed for high concurrency and resilience.

-   **Backend:** A **Go** server using a **Supervised Actor Model**. Each game runs in an isolated goroutine, processing events serially from a channel to guarantee consistency without locks.
-   **Persistence:** **Redis Streams** are used as a Write-Ahead Log (WAL) for event sourcing. This provides durability and fast recovery without Redis being a bottleneck for live gameplay.
-   **Frontend:** A **Go/WebAssembly** core shares critical game logic (`ApplyEvent`) with the server, wrapped in a **React/TypeScript** UI.
-   **AI Player:** A hybrid model using a deterministic **Go Rules Engine** for strategic decisions and an **LLM** for all communication, securely interfaced via a MCP protocol.

### Documentation

The project's design is extensively documented.

-   **[Game Design Document](./docs/01-game-design-document.md)**: The complete rules and player mechanics.
-   **[5-Minute Technical Onboarding](./docs/02-onboarding-for-engineers.md)**: The quickest way to understand the entire stack.
-   **[Architecture Deep Dive](./docs/architecture/README.md)**: Detailed explanations of the Actor Model, AI, and persistence strategy.
