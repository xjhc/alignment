# Alignment Project: TODO List & Known Issues

This file tracks the current state of development, including tasks to be completed and known bugs.

## 🔴 Critical Bugs

- when player is in the game and refreshes the page (or disconnect and reconnect) they should join the game immediately (not role reveal etc). currently seems like trying to go to lobby and do lobby_sync which fails.
- abandon game still doesn't work
- sometimes user is in a bad state or backend in bad state and they cant even make a game [BACKEND] {"time":"2025-07-09T18:28:14.10577537-07:00","level":"INFO","msg":"Creating lobby","service":"alignment-server","endpoint":"createLobby","user_id":"guest:ee336c9d-d21b-4109-95c5-92c55261be15","player_name":"1","lobby_name":"1 Game","is_private":false} but no game

- when "project milestone" is selected, it should show.
- playercard should have abbreviated role names like CEO instead of Chief Executive Officer.
- voting doesn't seem to work? vote should tally up by num tokens.
- skip should update with 0/n, 1/n, etc. as people press skip.

- reconnecting doesn't work.
- in lobby: host can run game in "Play as AI" mode where a random player will be assigned the AI player.
- in lobby: host can run game with "Custom Game" mode where they can change the number of Aligned players at start (0 by default)
- in lobby: there shouldn't be "Ready" button. everyone is ready by default.

- reactions don't show up on the msg.
- verdict phase should still let people talk
- vote should update with the "blockchain ui" in the storybook.
- mine for player night action should show self as target
- there should be Loebmate messages for announcements, crisis events, etc as well as the initial event. see docs.
- every role has a role ability. it should still display what it is, and how many research more needed to unlock.

- reactions/reply no longer show
- send textbox should support markdown and emojis
- it shouldn't go to /waiting when reconnecting
- lobby list should refresh automatically
- it spams [BACKEND] {"time":"2025-07-06T07:06:04.310315533-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59294","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.312977082-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59302","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.317574543-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59314","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.321270125-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59316","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"}
  [BACKEND] {"time":"2025-07-06T07:06:04.323964159-07:00","level":"WARN","msg":"Rate limit exceeded","service":"alignment-server","ip":"127.0.0.1:59322","endpoint":"/api/games/a52e197c-7175-4e59-b653-666cecc24497/join"} when trying to join a lobby using a link
- if player is in bad state, it should just log out.
- a player who was in a game before can't rejoin the game and can't do anything.
- players need to start w 1 token. even though there's a const and everything, it doesn't work.
- during trial: show who is nominated
- in lobby list etc before game join, show user profile etc. who is logged in, etc.
- skip still doesn't work
- player ability should ALWAYS show even if it's locked. right now it says No Active Ability
  This role has no special abilities.
- pulse check should also send with "Return" like all other msgs. currently its newline. newline is always alt + return
- remove msg render animation, just show grey (pending) -> not grey (sent) instantly
- when reconnecting, sometimes give sNo credentials available for reconnection
- when reconnecting to existing game
  Attempting to reconnect...
  websocket.ts:69 WebSocket connected
  websocket.ts:94 WebSocket closed: 1006
  websocket.ts:334 Attempting to reconnect...
  websocket.ts:69 WebSocket connected
  websocket.ts:94 WebSocket closed: 1006
  websocket.ts:334 Attempting to reconnect...
  websocket.ts:69 WebSocket connected
  websocket.ts:94 WebSocket closed: 1006
  websocket.ts:334 Attempting to reconnect...
  websocket.ts:69 WebSocket connected
  websocket.ts:94 WebSocket closed: 1006
  websocket.ts:334 Attempting to reconnect...
  websocket.ts:69 WebSocket connected
  websocket.ts:94 WebSocket closed: 1006
  [BACKEND] {"time":"2025-07-06T12:28:21.547472447-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547477674-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547489525-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:e4813ffb-fb23-4d6f-bc76-075b97f33295","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547502206-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:ba30e873-57d9-4c7d-b422-d2c71416e945: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547507112-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547510787-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547522337-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547542891-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:ba30e873-57d9-4c7d-b422-d2c71416e945","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547567797-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:21.547579869-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551228817-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551271796-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:e4813ffb-fb23-4d6f-bc76-075b97f33295: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.55127718-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551285572-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551293205-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551308241-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:ba30e873-57d9-4c7d-b422-d2c71416e945: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.55131261-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551316404-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551331232-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:e4813ffb-fb23-4d6f-bc76-075b97f33295","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551342512-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:ba30e873-57d9-4c7d-b422-d2c71416e945","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551355227-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:23.551367712-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554680287-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554719751-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:e4813ffb-fb23-4d6f-bc76-075b97f33295: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554725424-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554732543-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554746742-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:e4813ffb-fb23-4d6f-bc76-075b97f33295","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554758304-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554770743-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554789183-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:ba30e873-57d9-4c7d-b422-d2c71416e945: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.55479241-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554807037-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.554827801-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:ba30e873-57d9-4c7d-b422-d2c71416e945","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:25.55484551-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558230841-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.55826609-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:e4813ffb-fb23-4d6f-bc76-075b97f33295: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558272207-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558283374-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558304852-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558326171-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:ba30e873-57d9-4c7d-b422-d2c71416e945: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558333544-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Stopping","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558336142-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558355175-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:e4813ffb-fb23-4d6f-bc76-075b97f33295","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558362155-07:00","level":"INFO","msg":"GameLifecycleManager: Handling player disconnection: guest:ba30e873-57d9-4c7d-b422-d2c71416e945","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.55836488-07:00","level":"INFO","msg":"EventBus: Published event player_disconnected to 1 subscribers","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:27.558428397-07:00","level":"INFO","msg":"[PlayerActor/guest:ba30e873-57d9-4c7d-b422-d2c71416e945] Disconnecting in state Idle","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:29.561616007-07:00","level":"INFO","msg":"[PlayerActor/guest:e4813ffb-fb23-4d6f-bc76-075b97f33295] Starting","service":"alignment-server"}
  [BACKEND] {"time":"2025-07-06T12:28:29.561649293-07:00","level":"INFO","msg":"WebSocketManager: Failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c for player guest:e4813ffb-fb23-4d6f-bc76-075b97f33295: failed to auto-join lobby a8361668-957b-457b-a0f8-f3276d05f57c: lobby not found: a8361668-957b-457b-a0f8-f3276d05f57c","service":"alignment-server"}
- 🚨 Active Session Detected
  You're already in an active game session. To join a new game, you need to leave your current session first.
  Clear Current Session DOESNT WORK. there should also be a "REJOIN GAME" option.
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
