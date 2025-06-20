
### **Design: The Supervisor & Resiliency Layer**

#### **1. Overview & Goal**

This document outlines a Supervisor pattern for our single-machine architecture. The goal is to make our stateful server resilient to both internal failures (panicking game actors) and external overload (too many new game requests or low system memory) without crashing the entire process.

The system will gracefully degrade service by placing new users on a waitlist when under stress, ensuring stability for existing games.

#### **2. Core Components**

1.  **The Supervisor:** A top-level goroutine responsible for launching, monitoring, and restarting child **Game Actor** goroutines. It acts as a crash-proof boundary.
2.  **The Health Monitor:** A background goroutine that periodically checks system vitals (memory usage, global request rate) and maintains a server-wide `HealthStatus`.
3.  **The Admission Controller:** A middleware that acts as a gatekeeper for new game creation. It consults the `HealthStatus` to decide whether to accept a new game, or add the user to a waitlist.
4.  **Circuit Breakers:** A pattern applied within actors to wrap risky external calls (e.g., to the Language Model API), preventing repeated calls to a failing service.
5.  **The Waitlist:** A simple FIFO queue in Redis for users who tried to create a game when the server was overloaded.

#### **3. Architectural Diagram**

```ascii
+----------------+      +---------------------------+      +---------------------+
| New Game       |----->|   Admission Controller    |----->| The Supervisor      |
| Request        |      | (Consults Health Status)  |      | - Spawns Actors     |
+----------------+      |            |              |      | - Recovers Panics   |
                        | (Reject)   | (Accept)     |      +----------+----------+
                        v            v              |                 | (launches & monitors)
+----------------+      +---------------------------+                 |
| Redis Waitlist |      |  (Updates Health Status)  |                 v
+----------------+      |                           |      +---------------------+
                        |   +-------------------+   |      | Game Actor (Panic!) |
                        +-->|  Health Monitor   |<--+      +---------------------+
                            +-------------------+
```

#### **4. Component Logic & Implementation**

##### **A. The Supervisor Loop**

The Supervisor launches each Game Actor in its own protected goroutine. It uses `defer` and `recover()` to catch panics from any single game, preventing a cascading failure.

```go
func (s *Supervisor) StartNewGame(gameID string) {
    go func() {
        // The panic-proof boundary
        defer func() {
            if r := recover(); r != nil {
                log.Error("GameActor panicked", "gameID", gameID, "error", r)
                // Actor is now dead. We could add logic here to
                // notify players or attempt a clean shutdown for this game.
            }
        }()

        // Launch the actual game actor
        actor := NewGameActor(gameID)
        actor.Run() // This is the actor's main blocking loop
    }()
}
```

##### **B. The Health Monitor**

This runs in a single, dedicated goroutine for the entire server lifetime.

```go
// Shared, concurrent-safe status object
var currentHealthStatus = "HEALTHY"

func RunHealthMonitor() {
    // These values should be configured via environment variables based on deployment specs
    rateLimiter := rate.NewLimiter(rate.Limit(5), 10) // 5 new games/sec, burst of 10 (example values)
    memThresholdMB := 6000 // Memory threshold in MB (should be configured for deployment, e.g., 6GB on an 8GB machine)

    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        // 1. Check Memory
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        usedMemoryMB := m.Alloc / 1024 / 1024

        // 2. Check Rate Limiter (by seeing if we can take a token)
        isRateLimited := !rateLimiter.Allow()

        // 3. Update Global Status
        if usedMemoryMB > memThresholdMB || isRateLimited {
            atomic.Store(&currentHealthStatus, "OVERLOADED")
        } else {
            atomic.Store(&currentHealthStatus, "HEALTHY")
        }
    }
}
```

##### **C. The Admission Controller**

This is a simple function that wraps the creation of a new game.

```go
func HandleCreateGameRequest(userID string) (gameID string, err error) {
    serverStatus := atomic.Load(&currentHealthStatus)

    if serverStatus == "OVERLOADED" {
        // Add user to waitlist and return an error
        redisClient.LPush("game_waitlist", userID)
        return "", errors.New("server busy, you have been added to a waitlist")
    }

    // Server is healthy, proceed to create the game
    gameID = generateNewGameID()
    supervisor.StartNewGame(gameID)
    return gameID, nil
}
```
*A separate "Waitlist Processor" goroutine would periodically check if the server is `HEALTHY` and pop users from the Redis list to create their games.*

##### D. Circuit Breaker Integration (Inside the AI Component)

This pattern is applied to the AI component's failure-prone operations, specifically its external call to the Language Model API. The `GameActor` that manages the AI is responsible for this component's lifecycle. We use a library like `gobreaker`.

```go
// Inside the AI component's initialization (managed by the GameActor)
var llmCircuitBreaker = gobreaker.NewCircuitBreaker(...)

// When the AI's Rules Engine decides to generate chat, it calls the LLM via the breaker.
func (re *RulesEngine) callLanguageModelWithBreaker(context string) (string, error) {
    result, err := llmCircuitBreaker.Execute(func() (interface{}, error) {
        // This code only runs if the circuit is "CLOSED" or "HALF_OPEN"
        return callAzureOpenAI(context) // The actual blocking API call
    })

    if err != nil {
        // Error could be the original error, or gobreaker.ErrOpenState
        // The AI logic should handle this gracefully (e.g., stay silent).
        return "", err
    }
    return result.(string), nil
}
```

#### **5. Summary of Benefits**

*   **Stability:** A crash in one game actor will not bring down the server. Existing games continue unaffected.
*   **Graceful Degradation:** The server automatically protects itself from overload by refusing new games when memory is low or request rates are too high.
*   **Improved User Experience:** Instead of a generic error, users are informed they are on a waitlist, managing expectations.
*   **Fault Isolation:** Circuit breakers prevent the AI from repeatedly hitting a failing external service, saving resources and reducing error noise.
