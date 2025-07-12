# Alignment Game Backend Server

A comprehensive Go backend implementation for the corporate-themed social deduction game "Alignment" where humans must identify and eliminate a rogue AI before it converts enough staff to seize control.

## Architecture

The server implements a **Supervised Actor Model** with the following key components:

### Core Components

1. **Game Actor** (`internal/actors/game_actor.go`)

   - Each game runs in a dedicated goroutine with in-memory state
   - Processes actions serially through Go channels (no locks needed)
   - Implements the "Validate -> Persist -> Apply" pattern

2. **Supervisor** (`internal/actors/supervisor.go`)

   - Manages all game actors with fault isolation
   - Handles panic recovery and actor restart
   - Provides health monitoring and statistics

3. **Scheduler** (`internal/game/scheduler.go`)

   - Timing wheel algorithm for phase transitions
   - Handles automatic game progression
   - Supports different timer types (phase end, heartbeat, etc.)

4. **WebSocket Manager** (`internal/comms/websocket.go`)

   - Real-time communication with clients
   - Message routing to appropriate game actors
   - Connection management and heartbeat

5. **Redis Data Store** (`internal/store/redis.go`)

   - Write-Ahead Log (WAL) using Redis Streams
   - State snapshots for fast recovery
   - Event persistence and replay capability

6. **AI Player System** (`internal/ai/`)
   - **Strategic Brain**: Deterministic Go-based rules engine for game decisions
   - **Social Brain**: LLM-powered communication via external APIs
   - **Supervised Sidecar**: Non-blocking AI actors with panic recovery
   - **MCP Integration**: Secure game state access via Model Context Protocol

## Game Phases

The server implements the complete day/night cycle:

### Day Phase Progression

1. **SITREP** (15s) - Bot posts daily crisis and status
2. **PULSE_CHECK** (30s) - Private responses to daily prompt
3. **DISCUSSION** (2min) - Open debate after pulse check reveal
4. **EXTENSION** (15s) - Vote to extend discussion or proceed
5. **NOMINATION** (30s) - Token-weighted vote for elimination candidate
6. **TRIAL** (30s) - Nominated player's defense
7. **VERDICT** (30s) - Final YES/NO vote on elimination

### Night Phase

- **NIGHT** (30s) - All players submit private actions
- Mining attempts, AI conversion, role abilities processed

## API Endpoints

- `GET /health` - Server health check with statistics
- `GET /api/games` - List active games
- `POST /api/games/create` - Create new game
- `GET /api/stats` - Server statistics
- `WebSocket /ws` - Real-time game communication

## Configuration

### Environment Variables

- `PORT` - Server port (default: 8080)
- `REDIS_ADDR` - Redis address (default: localhost:6379)
- `REDIS_PASSWORD` - Redis password (optional)

## Building & Running

```bash
# Install dependencies
go mod tidy

# Build server
go build -o alignment-server ./cmd/server

# Run server (requires Redis)
./alignment-server
```

## Testing

```bash
# Run all tests
go test ./...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
