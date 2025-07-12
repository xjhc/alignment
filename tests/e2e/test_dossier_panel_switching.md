# Manual Test: Dossier Panel Player Switching

## Test Objective
Verify that the dossier panel correctly changes when selecting different player cards in the roster.

## Prerequisites
- Game running with at least 2 players
- Both players have joined and the game has started (SITREP phase or later)

## Test Steps

### 1. Initial State Verification
1. **Navigate** to the game UI
2. **Verify** the right panel (dossier) shows header "My Terminal"
3. **Verify** you can see private information:
   - Personal KPI objective (if Human)
   - Threat Meter (if Human)
   - Ability details (not "CLASSIFIED")
   - Settings section with "Abandon Game" button

### 2. Select Another Player
1. **Click** on another player's card in the left roster panel
2. **Verify** the dossier panel header changes to "[Other Player's Name]'s Dossier"
3. **Verify** a "🔍 PUBLIC VIEW" badge appears in the header
4. **Verify** private information is hidden:
   - No Personal KPI section visible
   - No Threat Meter visible (if they're Human)
   - Ability shows "🔍 Classified Information" instead of details
   - No Settings section with "Abandon Game" button
5. **Verify** public information is still visible:
   - Player identity card
   - Team objective
   - Corporate mandate (if any)
   - Last night's action (if any, from previous nights)

### 3. Return to Self
1. **Click** on your own player card in the roster (the one marked with "(Me)")
2. **Verify** the dossier panel header returns to "My Terminal"
3. **Verify** the "🔍 PUBLIC VIEW" badge disappears
4. **Verify** all private information is visible again

### 4. Test Multiple Players
1. **Repeat** steps 2-3 for each other player in the game
2. **Verify** each player's public dossier shows correctly
3. **Verify** returning to self always restores private view

## Expected Results
- ✅ Header title changes correctly: "My Terminal" ↔ "[Player Name]'s Dossier"
- ✅ Public view badge appears/disappears appropriately
- ✅ Private information (Personal KPI, Threat Meter, Ability details, Settings) only visible when viewing self
- ✅ Public information always visible regardless of selected player
- ✅ Smooth transitions between different player selections

## Bug Confirmation
If the dossier panel **does not** change when clicking different player cards, then Bug 5 is confirmed and needs fixing.

## Implementation Status
✅ **ALREADY IMPLEMENTED** - Code review shows:
- `PlayerCard` correctly wired with `onSelect={setViewedPlayer}` (RosterPanel.tsx:239)
- `PlayerHUD` correctly calculates `isViewingSelf` (PlayerHUD.tsx:32) 
- `AbilityCard` correctly hides details when `!isViewingSelf` (AbilityCard.tsx:16)
- Personal KPI only shown when `isViewingSelf` (PlayerHUD.tsx:111)
- Header title correctly changes based on `isViewingSelf` (PlayerHUD.tsx:35)