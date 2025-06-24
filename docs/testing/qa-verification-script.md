# QA Verification Script

This script provides step-by-step verification procedures for each phase to ensure all requirements from the feature checklist are properly implemented.

## Automated Verification Tests

### Core Game Logic Tests

Run these tests to verify backend requirements:

```bash
# Verify core game logic
go test ./core -v

# Verify action processing 
go test ./server/internal/game -v

# Verify phase transitions
go test ./server/internal/actors -v

# Run simulation tests for balance
make test-simulation
```

### Manual UI Verification Procedures

For each phase, follow these exact steps to verify UI requirements:

---

## LOBBY PHASE Verification

### Setup
1. Start development server: `make dev`
2. Open browser to `http://localhost:3000`
3. Create a new game

### Verification Steps

#### Player List Display
- [ ] **VERIFY**: Player list shows your name immediately after joining
- [ ] **TEST**: Open second browser tab, join with different name
- [ ] **VERIFY**: Both players appear in both browser tabs immediately
- [ ] **TEST**: Leave game in one tab
- [ ] **VERIFY**: Player list updates in both tabs immediately

#### Join/Leave Functionality  
- [ ] **TEST**: Click "Leave Game" button
- [ ] **VERIFY**: Button changes to "Join Game"
- [ ] **VERIFY**: Player removed from list
- [ ] **TEST**: Click "Join Game" again
- [ ] **VERIFY**: Button changes back to "Leave Game"
- [ ] **VERIFY**: Player re-added to list

#### Host Controls
- [ ] **TEST**: As game creator, verify "Start Game" button visible
- [ ] **VERIFY**: Button shows current player count: "Start Game (1/6)"
- [ ] **VERIFY**: Button disabled with < 2 players (minimum)
- [ ] **TEST**: Join with second player
- [ ] **VERIFY**: Button becomes enabled, shows "Start Game (2/6)"
- [ ] **TEST**: In non-host browser tab
- [ ] **VERIFY**: "Start Game" button NOT visible to non-host player

#### Game Settings
- [ ] **VERIFY**: Display shows "Max Players: 6" (or configured value)
- [ ] **VERIFY**: Display shows estimated duration
- [ ] **VERIFY**: Game ID displayed and copyable

---

## SITREP PHASE (PULSE CHECK) Verification

### Setup
1. Complete lobby verification above with 2+ players
2. Host clicks "Start Game"
3. Verify phase transitions to SITREP

### Critical UI Verification ⚠️
- [ ] **VERIFY**: Header displays "Pulse Check - Day 1"
- [ ] **VERIFY**: Three buttons displayed:
  - [ ] "Nominal" button (green color)
  - [ ] "Elevated" button (yellow color)  
  - [ ] "Critical" button (red color)
- [ ] **VERIFY**: Main chat input field is disabled/hidden
- [ ] **VERIFY**: No other action buttons available
- [ ] **VERIFY**: Timer shows countdown for phase duration

### Button Functionality
- [ ] **TEST**: Click "Nominal" button
- [ ] **VERIFY**: SUBMIT_PULSE_CHECK action sent (check network tab)
- [ ] **VERIFY**: All three buttons become disabled
- [ ] **VERIFY**: UI shows "Response submitted" or similar confirmation
- [ ] **VERIFY**: Other players see "[PlayerName] has responded"

### Action Payload Verification
- [ ] **TEST**: Check browser network tab after clicking button
- [ ] **VERIFY**: Payload contains: `{ "response": "NOMINAL" }`
- [ ] **TEST**: Try to submit second response (should fail)
- [ ] **VERIFY**: Server rejects duplicate submission

### Phase Transition
- [ ] **TEST**: Have all players submit responses
- [ ] **VERIFY**: Phase automatically transitions to NOMINATION
- [ ] **TEST**: Alternative - wait for timer to expire with some non-responses
- [ ] **VERIFY**: Phase still transitions (handles timeouts)

---

## NOMINATION PHASE Verification

### Setup
Complete SITREP verification above

### Player Selection Interface
- [ ] **VERIFY**: Grid/list shows all alive players except self
- [ ] **VERIFY**: Each player has "Nominate" button
- [ ] **VERIFY**: Player names and relevant info displayed
- [ ] **VERIFY**: Dead players (if any) not shown as options

### Nomination Process
- [ ] **TEST**: Click "Nominate" on another player
- [ ] **VERIFY**: Confirmation dialog appears: "Nominate [PlayerName] for elimination?"
- [ ] **TEST**: Click "Confirm" in dialog
- [ ] **VERIFY**: SUBMIT_VOTE action sent with correct target_player_id
- [ ] **VERIFY**: UI shows "You nominated [PlayerName]"
- [ ] **VERIFY**: Other players see "[PlayerName] has been nominated"

### Discussion Features
- [ ] **VERIFY**: Chat input is enabled during nomination
- [ ] **TEST**: Send chat message
- [ ] **VERIFY**: Message appears in all player views
- [ ] **VERIFY**: Evidence/voting discussion possible

### Phase Progression
- [ ] **TEST**: Submit nomination
- [ ] **VERIFY**: Phase transitions to VERDICT automatically
- [ ] **VERIFY**: Nominated player prominently displayed in verdict phase

---

## VERDICT PHASE Verification

### Setup
Complete NOMINATION verification above

### Voting Interface
- [ ] **VERIFY**: Nominated player clearly displayed at top
- [ ] **VERIFY**: Two voting buttons visible:
  - [ ] "Vote Guilty" (eliminate player)
  - [ ] "Vote Innocent" (save player)
- [ ] **VERIFY**: Vote explanation text visible
- [ ] **VERIFY**: Timer showing time remaining

### Vote Submission
- [ ] **TEST**: Click "Vote Guilty"
- [ ] **VERIFY**: Confirmation dialog (optional but recommended)
- [ ] **VERIFY**: SUBMIT_VOTE action sent
- [ ] **VERIFY**: UI shows "You voted Guilty"
- [ ] **VERIFY**: Voting buttons disabled after submission

### Live Vote Tally
- [ ] **VERIFY**: Real-time vote count displayed
- [ ] **VERIFY**: Token weights applied correctly (player with 3 tokens = 3 votes)
- [ ] **VERIFY**: Progress bar or indicator toward threshold
- [ ] **TEST**: Multiple players vote
- [ ] **VERIFY**: Tally updates in real-time for all players

### Vote Resolution
- [ ] **TEST**: Reach voting threshold before timer
- [ ] **VERIFY**: Phase ends immediately when threshold reached
- [ ] **VERIFY**: Result displayed: "Player eliminated" or "Player saved"
- [ ] **TEST**: Alternative - let timer expire
- [ ] **VERIFY**: Votes tallied and result applied

---

## EXTENSION PHASE Verification

### Setup
Complete VERDICT phase but ensure player is saved (not eliminated)

### Extension Interface
- [ ] **VERIFY**: Clear question displayed: "Extend discussion for [nominated player]?"
- [ ] **VERIFY**: Two buttons: "Yes" and "No"
- [ ] **VERIFY**: Current vote tally: "X Yes, Y No"
- [ ] **VERIFY**: Timer for extension phase

### Voting Process
- [ ] **TEST**: Click "Yes" for extension
- [ ] **VERIFY**: SUBMIT_VOTE action with extension payload
- [ ] **VERIFY**: Vote count updates immediately
- [ ] **VERIFY**: Other players see updated tally

### Result Application
- [ ] **TEST**: Extension passes (majority Yes)
- [ ] **VERIFY**: Phase returns to VERDICT with more time
- [ ] **TEST**: Extension fails (majority No)
- [ ] **VERIFY**: Phase transitions to NIGHT

---

## NIGHT PHASE Verification

### Setup
Complete day phases above to reach NIGHT

### Action Selection Interface
- [ ] **VERIFY**: Tab or dropdown for action selection
- [ ] **VERIFY**: Available actions based on player role/alignment:
  - [ ] Token mining (all players)
  - [ ] Role abilities (if role unlocked)
  - [ ] AI conversion (if AI player)

### Token Mining Interface
- [ ] **VERIFY**: Player selector excludes self
- [ ] **VERIFY**: Only alive players shown as targets
- [ ] **TEST**: Select player and click "Mine for [Player]"
- [ ] **VERIFY**: MINE_TOKENS action sent with correct beneficiary_id
- [ ] **VERIFY**: Confirmation: "Mining for [Player]"

### Role Abilities (if applicable)
- [ ] **VERIFY**: Role-specific UI appears when role unlocked
- [ ] **VERIFY**: Ability description clearly displayed
- [ ] **VERIFY**: Target selection (for targeted abilities)
- [ ] **TEST**: Use ability
- [ ] **VERIFY**: Correct night action submitted
- [ ] **VERIFY**: Ability marked as used (cooldown)

### AI Conversion (for AI players)
- [ ] **VERIFY**: Human players shown as conversion targets
- [ ] **VERIFY**: Success probability displayed
- [ ] **TEST**: Attempt conversion
- [ ] **VERIFY**: ATTEMPT_CONVERSION action sent
- [ ] **VERIFY**: Private result notification

### Night Resolution
- [ ] **TEST**: All players submit night actions
- [ ] **VERIFY**: Phase transitions automatically
- [ ] **VERIFY**: Night action results applied and displayed
- [ ] **VERIFY**: New day begins with updated game state

---

## Error Case Verification

### Network Issues
- [ ] **TEST**: Disconnect internet during action submission
- [ ] **VERIFY**: Error message displayed
- [ ] **TEST**: Reconnect internet
- [ ] **VERIFY**: Graceful reconnection with state sync

### Invalid Actions
- [ ] **TEST**: Attempt action in wrong phase (via console/DevTools)
- [ ] **VERIFY**: Server rejects with appropriate error
- [ ] **TEST**: Submit invalid payload data
- [ ] **VERIFY**: Client validation prevents submission

### Edge Cases
- [ ] **TEST**: Last player leaves during game
- [ ] **VERIFY**: Game handles gracefully (pause/end)
- [ ] **TEST**: Player disconnects during critical vote
- [ ] **VERIFY**: Game continues with remaining players

---

## Cross-Browser Verification

Repeat key verification steps in:
- [ ] **Chrome**: Full verification
- [ ] **Firefox**: Core functionality
- [ ] **Safari**: Core functionality  
- [ ] **Mobile Chrome**: Responsive design
- [ ] **Mobile Safari**: Responsive design

## Performance Verification

- [ ] **TEST**: Maximum player count (6 players)
- [ ] **VERIFY**: No lag in UI updates
- [ ] **VERIFY**: WebSocket messages handled efficiently
- [ ] **TEST**: Rapid action submission
- [ ] **VERIFY**: No race conditions or lost actions

## Accessibility Verification

- [ ] **TEST**: Tab navigation through all interactive elements
- [ ] **TEST**: Screen reader compatibility (basic)
- [ ] **VERIFY**: Sufficient color contrast for colorblind users
- [ ] **VERIFY**: Buttons have appropriate aria-labels

---

## QA Sign-off Checklist

Before marking any phase as complete:

- [ ] All verification steps completed successfully
- [ ] Error cases tested and handled appropriately  
- [ ] Cross-browser compatibility verified
- [ ] Performance acceptable with max players
- [ ] No console errors during normal usage
- [ ] All requirements from feature checklist verified

**QA Approval**: _[Name and Date]_

**Developer Sign-off**: _[Name and Date]_

---

## Regression Testing

Run this verification script:
- Before each release
- After any UI changes
- After backend logic changes
- When new features are added

This ensures existing functionality remains intact while new features are developed.