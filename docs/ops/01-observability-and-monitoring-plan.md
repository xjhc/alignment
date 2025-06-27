# Ops: Observability & Monitoring Plan

## 1. Philosophy: Proactive Insight, Not Reactive Panic

Our approach to observability is proactive. We will not wait for players to report problems. We will build a system that gives us a real-time, comprehensive view of the application's health, performance, and behavior, allowing us to identify and address issues before they impact the user experience.

This plan is guided by the "Three Pillars of Observability":
1.  **Logs:** Detailed, structured records of specific events. *("What happened?")*
2.  **Metrics:** Aggregated, numerical data about the system over time. *("How is it performing?")*
3.  **Traces:** A view of a single request's journey through our entire system. *("Where did it slow down or fail?")*

Our goal is to implement a simple but effective observability stack for V1 that can be expanded upon as the system grows.

## 2. The V1 Observability Stack

For our initial single-VM deployment, we will use a lightweight, powerful, and cost-effective stack.

*   **Metrics & Dashboarding:** **Prometheus** for collecting time-series metrics, and **Grafana** for creating dashboards and alerts. Both are open-source and can run in Docker containers on the same VM.
*   **Logging:** We will use **structured logging** (JSON format) directed to `stdout`. A log-shipping agent like **Promtail** will collect these logs and send them to **Loki**, which integrates seamlessly with Grafana.
*   **Tracing (Future):** For V1, we will defer distributed tracing, as our architecture is a single process. When we scale to multiple nodes, we will integrate OpenTelemetry.

## 3. Key Performance Indicators (KPIs) to Monitor

Our primary Grafana dashboard will be a single pane of glass showing the health of the entire system. It will be built from the following key metrics, exposed by our Go application via a `/metrics` endpoint for Prometheus to scrape.

#### **A. Technical & System Metrics (The "Server Health" Dashboard)**

| Metric Name | Type | Description & Thresholds |
| :--- | :--- | :--- |
| `go_goroutines` | Gauge | **Number of active goroutines.** *Alert if > 20,000 for 5 mins.* |
| `go_memstats_alloc_bytes` | Gauge | **Heap memory allocated.** *Alert if > 80% of system memory.* |
| `process_cpu_seconds_total` | Counter | **CPU time consumed.** *Alert on sustained high rate of change.* |
| `http_requests_total` | Counter | **Total HTTP requests by path & status code.** *Alert on high rate of 5xx errors.* |
| `websocket_connections_active`| Gauge | **Current number of active WebSocket connections.** |

#### **B. Application & Game Metrics (The "Game Performance" Dashboard)**

| Metric Name | Type | Description & Labels |
| :--- | :--- | :--- |
| `alignment_games_active` | Gauge | The current number of running `GameActor`s. |
| `alignment_lobbies_active` | Gauge | The current number of active `Lobby`s. |
| `alignment_players_online` | Gauge | Total number of connected `PlayerActor`s. |
| `alignment_action_processed_total`| Counter| Total actions processed, labeled by `action_type`. Helps identify most used actions. |
| `alignment_action_processing_duration_seconds`| Histogram | Latency of the "Validate -> Persist -> Apply" loop. Critical for performance. |
| `alignment_event_broadcast_total` | Counter | Total events broadcast, labeled by `event_type`. |

#### **C. Business & External Service Metrics (The "API & Cost" Dashboard)**

| Metric Name | Type | Description & Labels |
| :--- | :--- | :--- |
| `alignment_llm_api_requests_total`| Counter | Total calls to the Language Model API, labeled by `status` (success/fail). |
| `alignment_llm_api_request_duration_seconds`| Histogram | Latency of the LLM API calls. |
| `alignment_llm_circuit_breaker_state` | Gauge | The current state of the LLM circuit breaker (0=Closed, 1=HalfOpen, 2=Open). *Alert if state > 0.* |
| `alignment_redis_commands_total`| Counter | Total Redis commands, labeled by `command` (XADD, GET, etc.) and `status`. |
| `alignment_redis_command_duration_seconds`| Histogram | Latency of Redis commands. |

## 4. Structured Logging Strategy

To make logs useful for debugging, all log output from the Go application will be structured JSON written to `stdout`.

*   **Log Library:** We will use a library like `log/slog` (standard in Go 1.21+) or `zerolog`.
*   **Standard Fields:** Every single log line must include:
    *   `timestamp`: The time of the event.
    *   `level`: `DEBUG`, `INFO`, `WARN`, `ERROR`.
    *   `message`: The human-readable log message.
    *   `service`: The component that generated the log (e.g., `GameActor`, `Supervisor`, `LobbyManager`).
*   **Contextual Fields:** Logs should include relevant context.
    *   *Example Error Log:*
        ```json
        {
          "level": "error",
          "timestamp": "2023-10-27T10:00:00Z",
          "service": "GameActor",
          "message": "Failed to process action",
          "game_id": "g-f4b1",
          "player_id": "p-123",
          "action_type": "SUBMIT_VOTE",
          "error": "player is not alive"
        }
        ```
*   **Log Levels:**
    *   `DEBUG`: Verbose information for development (e.g., "Received action from player"). Disabled in production.
    *   `INFO`: Standard operational messages (e.g., "Supervisor started", "New game created").
    *   `WARN`: Non-critical issues that should be investigated (e.g., "LLM API latency is high," "Player submitted action after deadline").
    *   `ERROR`: Critical errors that require attention (e.g., "Failed to persist event to Redis," "GameActor panicked").

## 5. Alerting Strategy

Alerts will be configured in Grafana based on our defined metrics. They will be routed to a dedicated channel (e.g., a "Ops Alerts" channel in Discord/Slack).

**Critical Alerts (Page an on-call engineer):**

*   `High 5xx Error Rate`: The HTTP API is failing.
*   `LLM Circuit Breaker Open`: The AI's language brain is completely down.
*   `High Memory Usage`: The server is at risk of crashing.
*   `Redis Unreachable`: The persistence layer is down.

**Warning Alerts (Post to a channel, no page):**

*   `High LLM API Latency`: The AI might feel sluggish to players.
*   `High Action Processing Latency`: The core game loop is slowing down.
*   `GameActor Panicked`: The supervisor caught a crash (logged via an `alignment_actor_panics_total` counter).

This plan provides a comprehensive but achievable foundation for V1 observability, ensuring we have the necessary tools to operate the game reliably and professionally from day one.