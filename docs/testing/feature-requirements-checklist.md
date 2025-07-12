# Feature Requirements Checklist

This document provides a comprehensive checklist for each game phase and feature to prevent implementation gaps. Each requirement must be explicitly implemented, tested, and verified.

**The Pulse Check Bug Example:** An outdated version of this checklist specified a three-button UI, which was not built. This matrix ensures all features, as implemented, are tracked and verified.

## Phase-by-Phase Requirements

### 1. LOBBY PHASE

#### UI Requirements

- [ ] **Player List Display**: Show all joined players with names and ready status
- [ ] **Join/Leave Buttons**:
  - [ ] "Join Game" button when not in game
  - [ ] "Leave Game" button when in game
  - [ ] Button states update immediately when clicked
- [ ] **Host Controls**:
  - [ ] "Start Game" button visible only to host
  - [ ] Button disabled when < minimum players
  - [ ] Button shows player count: "Start Game (4/6)"
- [ ] **Game Settings Display**: Show max players, estimated duration
- [ ] **Real-time Updates**: Player joins/leaves update UI immediately

#### Action Requirements

- [ ] **JOIN_GAME Action**: Creates player entry, broadcasts update
- [ ] **LEAVE_GAME Action**: Removes player, broadcasts update
- [ ] **START_GAME Action**: Validates min players, transitions to SITREP

#### State Requirements

- [ ] **Player Management**: Track players, host designation, ready states
- [ ] **Validation**: Enforce min/max player limits
- [ ] **Persistence**: Lobby state survives brief disconnections

---

### 2. SITREP PHASE (PULSE CHECK)

#### UI Requirements ⚠️ **CRITICAL - Freeform Text Input**

- [ ] **Freeform Text Input Must Display**:
  - [ ] A multi-line `<textarea>` element with a placeholder (e.g., "Share your thoughts...").
  - [ ] A character limit is enforced (e.g., 280 characters).
  - [ ] A visible character counter shows `current/max` characters.
  - [ ] A "Submit" button is present.
- [ ] **Input Behavior**:
  - [ ] The "Submit" button is disabled when the textarea is empty or exceeds the character limit.
  - [ ] The "Submit" button becomes enabled when valid text is entered.
  - [ ] Pressing `Enter` (without `Shift`) submits the form.
  - [ ] After submission, the input area and button become disabled, showing a "Response submitted" confirmation.
- [ ] **Input Restrictions**:
  - [ ] The main chat input field MUST be disabled during the pulse check phase until a response is submitted.
  - [ ] No other game actions are allowed except submitting the pulse check response.
- [ ] **Phase Display**:
  - [ ] A clear header displays "Pulse Check - Day X".
  - [ ] A timer shows the time remaining in the phase.

#### Action Requirements

- [ ] **SUBMIT_PULSE_CHECK Action**:
  - [ ] The action payload must be `{ "response": "[user's freeform text string]" }`.
  - [ ] The server must validate that the `response` is a non-empty string and respects the character limit.
  - [ ] The server must not allow the same player to submit a response twice in the same phase.

#### State Requirements

- [ ] **Response Tracking**: The server must store the submitted free-form response string for each player ID.
- [ ] **Phase Completion**: The server must track which players have submitted their response.
- [ ] **Auto-progression**: The phase must end automatically when either the timer expires OR all living players have submitted their responses.

#### Integration Requirements

- [ ] **Server Processing**: The `PulseCheckManager` (or equivalent) validates and stores responses.
- [ ] **Broadcast Updates**: Other players see a generic notification (e.g., "[PlayerName] has responded") to show progress without revealing the content.
- [ ] **Analytics**: Pulse check responses feed into AI threat assessment and game analysis.

---

### 3. NOMINATION PHASE

#### UI Requirements

- [ ] **Player Nomination Interface**:
  - [ ] Grid/list of all alive players (excluding self)
  - [ ] "Nominate" button for each player
  - [ ] Confirmation dialog: "Nominate [PlayerName] for elimination?"
- [ ] **Nomination Display**:
  - [ ] Clear indication when nomination submitted: "You nominated [Player]"
  - [ ] Show if others have nominated: "Waiting for nominations..."
  - [ ] Display nominated player prominently when nomination exists
- [ ] **Discussion Interface**:
  - [ ] Chat enabled for discussion
  - [ ] Evidence sharing tools (if applicable)

#### Action Requirements

- [ ] **SUBMIT_VOTE Action** (nomination):
  - [ ] Payload: `{ "target_player_id": "playerID" }`
  - [ ] Validates target is alive and not self
  - [ ] Only one nomination per phase

#### State Requirements

- [ ] **Nomination Tracking**: Store nominated player ID
- [ ] **Voting State**: Initialize vote structure when nomination made
- [ ] **Timer Management**: Phase duration and auto-progression

---

### 4. VERDICT PHASE

#### UI Requirements

- [ ] **Voting Interface**:
  - [ ] Prominent display of nominated player
  - [ ] Two clear voting options:
    - [ ] "Vote Guilty" (eliminate player)
    - [ ] "Vote Innocent" (save player)
  - [ ] Vote confirmation: "You voted [Guilty/Innocent]"
- [ ] **Live Vote Tally**:
  - [ ] Real-time vote count display
  - [ ] Token-weighted voting visualization
  - [ ] Progress toward required threshold
- [ ] **Timer Display**: Clear countdown to voting deadline

#### Action Requirements

- [ ] **SUBMIT_VOTE Action** (verdict):
  - [ ] Payload: `{ "target_player_id": "nominatedPlayer" }` (guilty) or `{ "target_player_id": "" }` (innocent)
  - [ ] Token-weight calculation applied automatically
  - [ ] Cannot change vote once cast

#### State Requirements

- [ ] **Vote Aggregation**: Track votes with token weights
- [ ] **Threshold Calculation**: Dynamic threshold based on total tokens
- [ ] **Result Determination**: Auto-resolve when threshold reached or timer expires

---

### 5. EXTENSION PHASE

#### UI Requirements

- [ ] **Extension Voting**:
  - [ ] Clear question: "Extend discussion for [nominated player]?"
  - [ ] "Yes" and "No" voting buttons
  - [ ] Shows current vote status: "3 Yes, 2 No"
- [ ] **Timer Display**: Shows extension phase duration
- [ ] **Result Preview**: "Extension will pass/fail with current votes"

#### Action Requirements

- [ ] **SUBMIT_VOTE Action** (extension):
  - [ ] Payload: `{ "vote_type": "EXTENSION", "in_favor": true/false }`
  - [ ] Simple majority required (not token-weighted)

#### State Requirements

- [ ] **Binary Vote Tracking**: Count Yes/No votes
- [ ] **Result Application**: Return to VERDICT if passed, proceed to NIGHT if failed

---

### 6. NIGHT PHASE

#### UI Requirements

- [ ] **Action Selection Interface**:
  - [ ] Tab-based or dropdown selection of available actions
  - [ ] **Token Mining**:
    - [ ] Player selector (excluding self)
    - [ ] "Mine for [Player]" button
  - [ ] **Role Abilities** (if unlocked):
    - [ ] Role-specific UI elements
    - [ ] Target selection for targeted abilities
    - [ ] Clear description of ability effect
  - [ ] **AI Conversion** (for AI players):
    - [ ] Target selection from human players
    - [ ] Success probability display

#### Action Requirements

- [ ] **MINE_TOKENS Action**:
  - [ ] Payload: `{ "beneficiary_id": "playerID" }`
  - [ ] Enforces selfless mining rule
  - [ ] Validates beneficiary is alive
- [ ] **SUBMIT_NIGHT_ACTION Action**:
  - [ ] Role-specific payload validation
  - [ ] Ability cooldown enforcement
  - [ ] Target validation for targeted abilities
- [ ] **ATTEMPT_CONVERSION Action** (AI only):
  - [ ] Payload: `{ "target_id": "playerID" }`
  - [ ] Validates target is human and alive

#### State Requirements

- [ ] **Action Storage**: Track all submitted night actions
- [ ] **Resolution Order**: Process actions in correct sequence
- [ ] **Conflict Resolution**: Handle blocks, protections, competing actions

---

## Cross-Phase Requirements

### Real-Time Communication

- [ ] **WebSocket Connection**: Maintains stable connection
- [ ] **Reconnection Handling**: Graceful reconnection with state sync
- [ ] **Event Broadcasting**: All state changes broadcast to relevant players
- [ ] **Private Information**: Sensitive data only sent to authorized players

### Game State Management

- [ ] **State Synchronization**: Client state matches server state
- [ ] **Event Sourcing**: All changes represented as events
- [ ] **Rollback Protection**: Invalid actions rejected gracefully
- [ ] **Persistence**: Game state survives server restarts

### User Experience

- [ ] **Loading States**: Show spinners/progress during async operations
- [ ] **Error Handling**: Clear error messages for failed actions
- [ ] **Accessibility**: Keyboard navigation, screen reader support
- [ ] **Mobile Responsiveness**: Functional on mobile devices

### Security & Validation

- [ ] **Input Sanitization**: All user inputs validated and sanitized
- [ ] **Authorization**: Players can only perform allowed actions
- [ ] **Anti-Cheating**: Client cannot manipulate game logic
- [ ] **Rate Limiting**: Prevent spam/DoS attacks

## Feature Implementation Process

For each new feature or phase:

1. **Requirements Definition**: Create specific, testable requirements like above
2. **UI Mockups**: Design exact interface before implementation
3. **Action Specification**: Define precise action payloads and validation
4. **State Management**: Plan how state changes propagate
5. **Test Cases**: Write test cases for each requirement
6. **Implementation**: Build according to requirements
7. **Verification**: Test each checklist item explicitly
8. **Integration Testing**: Test interaction with other phases

## Quality Gates

Before any phase is considered "complete":

- [ ] **All checklist items implemented and tested**
- [ ] **Unit tests cover all action validation logic**
- [ ] **Integration tests verify phase transitions**
- [ ] **UI tested in multiple browsers**
- [ ] **Error cases tested and handled gracefully**
- [ ] **Performance tested with max player count**
