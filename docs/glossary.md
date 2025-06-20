# Glossary of Terms

This document defines the core concepts and terminology used throughout the Alignment project documentation.

---

## Game Concepts

### **AI Equity**
A hidden server-side value tracked for each human player representing their susceptibility to AI conversion. Increases when targeted by AI conversion attempts.

### **AI Faction**
The minority faction of players aligned with the AI's goals. Includes the Original AI and any converted Aligned humans. Victory condition: Control 51% or more of total Tokens in play.

### **Aligned**
The alignment status of a human player who has been successfully converted by the AI faction. Aligned players gain access to the secret `#aligned` channel and work toward the AI's victory condition.

### **Deactivation**
The process of removing a player from active gameplay through the daily voting mechanism. Deactivated players move to the `#off-boarding` channel and can influence the game through the Whistleblower Protocol.

### **Human Faction**
The majority faction of players representing human employees. Victory condition: Identify and deactivate the Original AI before it achieves The Singularity.

### **LIAISON Protocol**
A critical emergency procedure that provides humans with extra Tokens and information when triggered by specific crisis events or game conditions.

### **Liquidity Pool**
The shared pool of Tokens available for distribution through mining, crisis events, and other game mechanics.

### **Personal KPI**
A secret individual objective assigned to each human player at game start. Successful completion provides bonuses or alternate win conditions.

### **System Shock**
A temporary negative status effect applied to a human player when they successfully resist an AI conversion attempt. Serves as both punishment and proof of their humanity.

### **The Singularity**
The AI faction's victory condition, achieved when they control 51% or more of all Tokens in play at the end of any Day Phase.

### **Tokens**
The primary resource and voting currency in the game. Used for weighted voting during deactivation proceedings and represents a player's influence within the company.

### **Whistleblower Protocol**
A mechanic allowing deactivated players in `#off-boarding` to vote on crisis events that will affect future gameplay, maintaining their influence on the narrative.

---

## Architectural Terms

### **Actor**
A goroutine that owns and manages the complete state for a single game. Processes all actions and events serially through a Go channel to ensure consistency without locks.

### **Dispatcher**
The central routing component that receives all incoming WebSocket messages and forwards them to the appropriate Game Actor's mailbox channel.

### **Language Model**
The AI component responsible for generating human-like chat messages and social interaction. Uses the MCP interface to access game context and runs separately from game logic.

### **MCP (Model Context Protocol)**
The secure, read-only API interface between the backend server and the Language Model, providing structured game state access without allowing direct game manipulation.

### **Rules Engine**
The deterministic AI component that makes all strategic game decisions (voting, targeting, special abilities). Runs in a dedicated sidecar goroutine alongside the main Game Actor.

### **Scheduler**
A single goroutine managing all time-based events for the entire server using a Timing Wheel algorithm. Handles phase timers, AI delays, and other scheduled actions.

### **Supervisor**
The top-level goroutine that launches and monitors all Game Actors. Provides fault isolation by catching panics from individual games without crashing the entire server.

### **WAL (Write-Ahead Log)**
The persistence pattern using Redis Streams to durably record all game events before applying them to in-memory state, enabling crash recovery and audit trails.

---

## Technical Implementation Terms

### **Bridge Layer**
The JavaScript interface connecting the React UI with the Go/Wasm game engine, enabling bidirectional communication between the frontend components.

### **Circuit Breaker**
A fault tolerance pattern applied to external API calls (like Language Model requests) to prevent cascading failures when services become unavailable.

### **GameState**
The complete in-memory representation of a game's current state, including all players, votes, tokens, phase information, and historical data.

### **Health Monitor**
A background goroutine that periodically checks system resources (memory, request rates) and updates the server's health status for admission control.

### **Sidecar Goroutine**
A concurrent goroutine running alongside the main Game Actor that handles specific responsibilities (like AI decision-making) without blocking the primary game loop.

### **Timing Wheel**
An efficient algorithm for managing thousands of concurrent timers with minimal overhead, used by the Scheduler for all time-based game events.

---

## Game Channels

### **#aligned**
Secret channel visible only to AI faction members (Original AI and Aligned humans). Used for coordination and strategic planning.

### **#off-boarding**
Spectator channel for deactivated players. Can view `#war-room` but cannot participate in active discussions. Home of the Whistleblower Protocol.

### **#war-room**
Main public channel where all active players conduct game discussions, debates, and social deduction activities.

---

## Player Roles & Abilities

### **CISO (Chief Information Security Officer)**
Human role with advanced security abilities for detecting and blocking AI activities.

### **VP, Ethics & Alignment**
Human role with specialized abilities related to organizational ethics and detecting alignment issues.

### **CEO (Chief Executive Officer)**
Human role with the ability to issue a `Performance Review`, forcing a target player to perform a specific action, overriding their choice.

### **CFO (Chief Financial Officer)**
Human role with the ability to `Reallocate Budget`, transferring a Token from one player to another.

### **COO (Chief Operating Officer)**
Human role with the ability to `Pivot`, choosing the next day's Crisis Event from a limited set of options.

### **CTO (Chief Technology Officer)**
Human role with the ability to `Overclock Servers`, guaranteeing a successful mine for themselves and another player, but with a hidden risk.

### **VP, Platforms**
Human role with the ability to `Deploy Hotfix`, redacting a portion of the next day's public information report (SITREP).

---

## Development & Testing Terms

### **ApplyEvent**
The core pure function that takes a current GameState and an Event, returning the new GameState. Central to the deterministic game logic.

### **Event Sourcing**
The architectural pattern of storing all state changes as a sequence of immutable events rather than updating state directly.

### **Table-Driven Tests**
The Go testing pattern used extensively for testing pure functions like ApplyEvent with comprehensive input/output scenarios.

---

*This glossary is maintained as a living document. When adding new concepts to the codebase, ensure they are defined here and linked appropriately throughout the documentation.*