# ADR-008: Backend Concurrency & Resiliency Playbook

---

## 1. Context

Our Actor Model provides a strong foundation for concurrency. However, subtle bugs—such as goroutine leaks, race conditions, and inconsistent state—have emerged in production. These are rooted in the absence of **formal, enforced patterns** for managing the complexity of a highly concurrent system.

This document defines the **canonical playbook** for writing **robust, maintainable, and safe** concurrent code in the `Alignment` backend. While primarily focused on concurrency control, it also lays groundwork for system resiliency.

---

## 2. Core Tenets

Every concurrent component must follow these principles:

1. **Isolation:** Components must be self-contained and interact via message-passing or methods, not by accessing shared memory.
2. **Controlled Concurrency:** Goroutines are not fire-and-forget. Their creation, lifecycle, and termination must be explicitly managed.
3. **Predictable State:** All mutable state must be owned by a single actor. Mutations must be serialized and traceable.
4. **Graceful Failure:** Components must contain failure—whether panics, errors, or timeouts—so one failure does not cascade across the system.
5. **Safe Observability:** No internal state should be leaked or shared in a way that compromises consistency or safety.

---

## 3. Patterns & Checklists

---

### Pattern I: `context.Context` Lifecycle Mandate

**Risk Mitigated:** Goroutine Leaks

**The Rule:** Every function that spawns a long-running goroutine **must accept a `context.Context`**. Goroutines must monitor this context and exit when it is canceled.

**Example:**

```go
func (a *MyActor) Start(ctx context.Context) {
    actorCtx, cancel := context.WithCancel(ctx)
    a.ctx = actorCtx
    a.cancel = cancel

    go a.processLoop()
}

func (a *MyActor) processLoop() {
    defer log.Println("Loop stopped.")
    for {
        select {
        case <-a.ctx.Done():
            return
        case msg := <-a.mailbox:
            // ...
        }
    }
}

func (a *MyActor) Stop() {
    a.cancel()
}
```

**Checklist:**

- [ ] Is `context.Context` accepted at the boundary where concurrency begins?
- [ ] Does the goroutine check `ctx.Done()` inside a `select` loop?
- [ ] Is a corresponding `cancel()` always called on shutdown?

---

### Pattern II: “Actor Owns State” Integrity Mandate

**Risks Mitigated:** Data Races, Inconsistent State, Deadlocks

**The Rule:** An actor’s state must only be read or written within its single processing goroutine. External callers must use message-passing.

**Correct Pattern:**

```go
type getStateRequest struct {
    responseChan chan *core.GameState
}

func (ga *GameActor) GetGameState() *core.GameState {
    respChan := make(chan *core.GameState)
    ga.mailbox <- getStateRequest{responseChan: respChan}
    return <-respChan
}

func (ga *GameActor) processLoop() {
    for msg := range ga.mailbox {
        switch req := msg.(type) {
        case getStateRequest:
            copy := deepCopy(ga.state)
            req.responseChan <- copy
        }
    }
}
```

**Checklist:**

- [ ] Are state fields unexported or read-only externally?
- [ ] Are all reads/writes serialized in the actor’s loop?
- [ ] Are methods returning a **copy**, never a pointer, to internal state?

---

### Pattern III: Panic Containment in Goroutines

**Risks Mitigated:** Unrecoverable Crashes, Hidden Failures

**The Rule:** Any long-running goroutine must defer a `recover()` to prevent panics from crashing the process or silently terminating the goroutine.

**Example:**

```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered in goroutine: %v", r)
        }
    }()
    a.processLoop()
}()
```

**Checklist:**

- [ ] Does every long-running goroutine have a panic recovery guard?
- [ ] Is the panic logged with context?
- [ ] Are side-effects (e.g., metrics, error channels) used if appropriate?

---

### Pattern IV: Channel Usage and Shutdown Contracts

**Risks Mitigated:** Stuck Goroutines, Sends to Closed Channels, Message Loss

**The Rule:** Channels are a concurrency contract. Each actor must clearly define:

- Who owns the channel?
- Who closes it (if ever)?
- What happens on shutdown?

**Guidelines:**

- Prefer unbuffered channels for strict coordination.
- Use buffered channels only when necessary—and size them conservatively.
- Never close a channel unless you are **the sole sender**.
- On shutdown, either:

  - Drain and discard remaining messages (if stateless), or
  - Drain and process final messages (if order matters).

**Checklist:**

- [ ] Are channel ownership and closure rules clearly defined?
- [ ] Are all sends/selects guarded against closed channels?
- [ ] Are goroutines guaranteed to exit when the actor shuts down?

---

### Pattern V: Asynchronous Test Synchronization

**Risks Mitigated:** Flaky Tests, Race Conditions in CI

**The Rule:** All asynchronous behavior in tests must be synchronized using `sync.WaitGroup` or equivalent. `time.Sleep()` is **banned**.

**Example:**

````go
func TestManagerReceivesSignal(t *testing.T) {
    var wg sync.WaitGroup
    wg.Add(1)
### New Best Practices Document

Here is a new architectural document that codifies the lessons learned from debugging this concurrent, stateful system. This document is crucial for maintaining the stability and quality of the backend as it grows.

```markdown

````

    mock := &MockHandler{
        onSignal: func() {
            defer wg.Done()
        },
    }

    manager := NewManager(mock)
    manager.TriggerAsyncSignal()

    waitWithTimeout(&wg, time.Second)
    assert.True(t, mock.Called)

}

```

**Checklist:**

- [ ] Is `wg.Add(1)` called **before** the async action?
- [ ] Is `defer wg.Done()` used inside the mock handler?
- [ ] Is a `waitWithTimeout` used instead of `time.Sleep`?

---

## 4. Optional Resiliency Enhancements (Future Patterns)

To deepen resilience, consider defining additional patterns for:

- **Timeout Handling:** All blocking operations (e.g., channel reads, RPCs) should have upper bounds.
- **Retry Policies:** Use backoff and circuit breakers for retryable operations.
- **Error Classification:** Distinguish between transient vs fatal errors in actor messages.

These are not enforced today, but may evolve into future ADRs.

## 5. Glossary

- **Actor:** A long-lived, single-threaded object that receives messages and owns state.
- **Mailbox:** The channel through which messages are sent to the actor.
- **Deep Copy:** A complete, independent clone of a struct or map to prevent shared mutation.
- **Graceful Shutdown:** A termination process that ensures all goroutines exit and resources are released.

---

## 6. Compliance & Review

- New code must follow all patterns and pass checklist items.
- Violations may be flagged in code review or subject to future static analysis tooling.
- Teams are encouraged to self-audit for legacy violations and raise follow-ups if systemic cleanup is needed.
```
