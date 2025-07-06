# Alignment Project: TODO List & Known Issues

This file tracks the current state of development, including tasks to be completed and known bugs.

## 🔴 Critical Bugs

- it shouldn't go to /waiting when reconnecting
- lobby list should refresh automatically
- it spams [BACKEND] {"time":"2025-07-06T07:06:04.310315533-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59294","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.312977082-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59302","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.317574543-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59314","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.321270125-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59316","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.323964159-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59322","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"} when trying to join a lobby using a link
- if player is in bad state, it should just log out.
- a player who was in a game before can't rejoin the game and can't do anything.

- [ ] **Player Visibility Race Condition**: In lobbies with 4+ players, the last player to join sometimes does not see the other players in the UI. The server state is correct, but the initial `LOBBY_STATE_UPDATE` event seems to be missed or processed incorrectly by the client.
- [ ] **Lobby-to-Game Transition Hang**: The transition from the waiting room to the game screen sometimes hangs, forcing players to refresh. This appears to be a deadlock or race condition in the `LobbyManager` when creating the `GameActor`.
- [ ] **Duplicate Pulse Check Messages**: The system is sometimes emitting two `PULSE_CHECK_SUBMITTED` events for a single response, causing duplicated messages in the chat log. This points to an issue with the event-handling logic, possibly in both the `GameActor` and the client.
- [ ] **System Message Spam**: The generic `SYSTEM_MESSAGE` event is being used for too many different purposes with inconsistent payloads, making client-side handling fragile. This has led to several UI bugs where system announcements are not displayed correctly. This is an architectural issue that needs to be addressed.

## 🟡 High-Priority Tasks

- [ ] **Implement Player Deactivation Flow**: The full voting and deactivation cycle is not implemented.
- [ ] **Flesh out Game Phases**: Implement the logic and UI for all game phases (SITREP, NOMINATION, VERDICT, NIGHT).
- [ ] **Implement AI `RulesEngine`**: The strategic brain of the AI is currently a placeholder. The core heuristics for voting and targeting need to be built.
- [ ] **Implement `MCP Server`**: Build the MCP server to provide game state to the AI's Language Model.
- [ ] **Complete Night Phase Resolution Logic**: The full precedence order for night actions (blocking, conversion, standard actions) needs to be implemented.
- [ ] **Formalize Session Management**: The interaction between `LobbyManager` and `SessionManager` needs to be clarified and hardened to fix the transition hang bug. A single, unified `GameLifecycleManager` may be the solution.
- [ ] **Add "Contract-First" API Tooling**: Implement automated Go-to-TypeScript type generation to prevent client-server contract drift.

## 🟢 Medium-Priority Tasks

- [ ] **Build Post-Game Analysis Screen**: Create the UI and data processing for the after-action report.
- [ ] **Implement Player Identity System**: Add support for guest users (localStorage ID) and authenticated users (Discord OAuth).
- [ ] **Build `Loebmate` Assistant**: Implement the FTUE hints system.
- [ ] **Add "Kudos" and Reporting System**: Implement the post-game social features.
- [ ] **Refine UI/UX**: Polish animations, transitions, and component styles based on the design documents.
- [ ] **Implement `Player Status` and `Whisper`**: Add the specialized communication mechanics.

## 🔵 Low-Priority Tasks

- [ ] **Build Internal Admin Tool**: Create the dashboard for server monitoring and management.
- [ ] **Implement Full Achievement System**: Build out the achievement tracking and reward unlocking system.
- [ ] **Expand AI Personas**: Add more character personas and prompting strategies to the AI system.

## 💡 Architectural Decisions (ADRs) to Document

- [ ] **ADR-004: Defer Full CQRS:** Document the decision to stick with the simpler actor-manager model for V1 instead of a more complex CQRS/event-bus architecture.
- [ ] **ADR-005: Contract-First API:** Document the decision to use automated type generation.
- [ ] **ADR-006: Single Authoritative Event:** Document the principle of emitting one comprehensive event per action to solve the pulse check and system message bugs.
- [ ] **ADR-007: Progressive Identity:** Document the hybrid guest/registered user identity model.
