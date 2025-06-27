# Architectural Linter

This tool implements static analysis to enforce ADR-006's "Single Authoritative Event" pattern by detecting usage of deprecated event types in the frontend codebase.

## What it does

1. **Parses `core/types.go`** to identify event constants marked with `// Deprecated:` comments
2. **Scans `client/src/services/websocket.ts`** for `case` statements handling deprecated events
3. **Fails the build** if deprecated events are used in client-side switch statements

## Usage

From the project root:

```bash
go run tools/arch-lint/main.go
```

## CI Integration

The linter runs automatically in the `test-frontend` job of the CI pipeline and will fail the build if architectural violations are detected.

## Adding New Deprecated Events

To mark an event as deprecated:

1. Add a comment in `core/types.go`:
   ```go
   // Deprecated: see ADR-006. Use EventNewPattern with comprehensive payload instead
   EventOldPattern EventType = "OLD_PATTERN"
   ```

2. The linter will automatically detect it and check for violations in the frontend code.

## Example Output

```
Found 6 deprecated events
  - EventMiningSuccessful (MINING_SUCCESSFUL)
  - EventPlayerBlocked (PLAYER_BLOCKED)
  - EventSystemMessage (SYSTEM_MESSAGE)
  ...

❌ ARCHITECTURAL VIOLATIONS DETECTED:
  Line 249: Usage of deprecated event EventSystemMessage

See ADR-006 for guidance on using single authoritative events instead.
```