# ADR-006: Enforce "Single Authoritative Event" Pattern Across All Game Actions

*   **Status:** Accepted
*   **Supersedes:** Dual-event emission pattern with granular and aggregate events

## Context

During development, we discovered a systematic fragility in our event handling architecture that manifested in two critical bugs:

1. **Pulse Check Bug:** The system emitted duplicate pulse check messages because multiple code paths generated different events in response to a single player action.
2. **System Message Spam Bug:** Generic `SYSTEM_MESSAGE` events flooded the client with inconsistent payload structures, making client-side handling fragile and error-prone.

The root cause was a "Dual-Event Emitter" anti-pattern where multiple, uncoordinated events were generated for single player actions:

- **Voting System:** Emitted both `VOTE_CAST` (granular) and `VOTE_TALLY_UPDATED` (aggregate) events
- **Night Resolution:** Emitted individual action events (`EventPlayerBlocked`, `EventMiningSuccessful`) AND a comprehensive summary (`EventNightActionsResolved`)
- **System Notifications:** Used generic `EventSystemMessage` with varied payload structures instead of semantic event types

This pattern created race conditions, made debugging difficult, and established tight coupling between server logic and client UI implementation.

## Decision

We will formally adopt and enforce the **"Single, Authoritative Event"** principle across the entire codebase:

**Core Principle:** For any given player action, the server must generate one, and only one, event that fully describes the resulting state change. The client's sole responsibility is to apply this event to its local state and re-render.

### Specific Implementation:

#### 1. Voting System Refactoring
- **Deprecated:** `VOTE_CAST` event for individual votes
- **Enhanced:** `VOTE_TALLY_UPDATED` event with comprehensive payload including:
  - Complete vote tallies and token weights
  - Individual voter information (when transparency mandate is active)
  - Vote completion status
  - The specific voter and target that triggered the update

#### 2. Night Action Resolution Refactoring
- **Eliminated:** Granular events (`EventPlayerBlocked`, `EventMiningSuccessful`, etc.)
- **Enhanced:** `EventNightActionsResolved` with structured payload containing:
  - Categorized results (blocked players, mining results, conversions, etc.)
  - Player state changes organized by player ID
  - Structured data instead of descriptive strings
  - Complete outcome information for deterministic client rendering

#### 3. System Message Replacement
- **Deprecated:** Generic `EventSystemMessage` 
- **Implemented:** Specific semantic events:
  - `EventClientError` - Client-side error notifications
  - `EventLiaisonProtocolActivated` - LIAISON protocol activation
  - `EventLiaisonIntelRevealed` - Intelligence revelations
  - `EventAIConversionBlocked` - AI conversion blocking notifications
  - `EventSitrepPublished` - Daily situation reports
  - `EventGameRuleModified` - Dynamic rule changes

#### 4. Shared Contract Updates
- All new event types added to `core/events.go` for contract validation
- Updated `EventTypeValues` array for automated tooling
- Maintained backward compatibility during transition

## Consequences

### Pros:
- **Eliminates Race Conditions:** Single events prevent timing issues between multiple event emissions
- **Improved Debugging:** Clear cause-and-effect relationship between actions and state changes
- **Reduced Client Complexity:** Clients only need to handle one authoritative event per action
- **Type Safety:** Specific event types enable proper TypeScript interfaces and validation
- **Better UX:** Structured payloads allow deterministic UI rendering without string parsing
- **Easier Testing:** Mock single events instead of coordinating multiple event sequences

### Cons:
- **Migration Effort:** Required refactoring existing dual-event patterns
- **Larger Event Payloads:** Single events may contain more data than granular events
- **Learning Curve:** Developers must understand the single-event constraint

### Implementation Benefits Achieved:

1. **Voting System:** Client state is now updated atomically with complete vote information
2. **Night Resolution:** Deterministic rendering of night outcomes without coordination issues  
3. **Error Handling:** Type-safe error events with structured error codes and retry logic
4. **Protocol Events:** Semantic events for liaison protocol with proper state tracking

## Enforcement

This principle is now enforced through:

1. **Code Review Guidelines:** PR template includes checklist for single authoritative events
2. **Documentation Updates:** `CONTRIBUTING_AI.md` and API documentation reflect this standard
3. **Architectural Decision:** Future development must adhere to this pattern
4. **Type System:** Specific event types prevent accidental dual-event patterns

This decision significantly improves system robustness, reduces client-side complexity, eliminates an entire class of race condition bugs, and makes the codebase easier to reason about and maintain.