# Feature: Internal Admin Tool

This document describes the design for an internal web-based tool for developers and administrators to monitor the health, status, and logs of the `Alignment` server in real-time.

## 1. Feature Overview & Goals

The Internal Admin Tool is a **separate, password-protected web interface** served by the main Go backend on a different port or path (e.g., `/admin`). It is **not** part of the main game client.

Its primary goals are:
*   **Operational Visibility:** Provide an at-a-glance view of server health and resource usage.
*   **Live Debugging:** Allow developers to inspect the state of active games and view live log streams.
*   **Cost Monitoring:** Track key metrics related to expensive resources, like LLM API calls.
*   **Manual Intervention:** Provide basic controls to manage the server or specific games.

## 2. Architectural Approach

The admin tool will be a simple, server-rendered HTML application or a lightweight single-page app (built with a minimal framework like Preact or just vanilla JS).

*   **Data Source:** The admin UI will get its data from a new set of secure, internal-only API endpoints exposed by the Go backend.
*   **Real-time Updates:** It will use a separate WebSocket connection to receive a stream of live metrics and logs from the server.
*   **Security:** Access to the `/admin` path and its API endpoints will be protected by a simple, hardcoded username/password (or an environment variable-based secret) to prevent public access.

## 3. Key UI Panels & Features

The tool will be a dashboard composed of several key panels:

#### **Panel 1: Server Health & Status**

*   **Purpose:** A high-level overview of the server's current state.
*   **Metrics Displayed:**
    *   **Server Status:** `HEALTHY` / `OVERLOADED` (from the Health Monitor).
    *   **CPU & Memory Usage:** Real-time graphs of the Go process's resource consumption.
    *   **Active Goroutines:** A count of total running goroutines.
    *   **Uptime:** How long the server has been running since the last restart.
    *   **Waitlist:** The current number of users on the waitlist.

#### **Panel 2: Active Games & Actors**

*   **Purpose:** A list of all currently active `GameActors` and `LobbyActors`.
*   **Features:**
    *   A table showing `GameID`, `Status` (Lobby/In-Game), `Player Count`, and `Age`.
    *   A "View State" button next to each game that allows an admin to dump the current, in-memory `GameState` object as a formatted JSON blob for inspection.
    *   A "Kill Actor" button (a "big red button") that allows an admin to manually terminate a misbehaving or stuck game actor.

#### **Panel 3: LLM & API Metrics**

*   **Purpose:** Monitor usage and cost of external services.
*   **Metrics Displayed:**
    *   **Total LLM Calls (24h):** A counter for API requests to the LLM provider.
    *   **Average LLM Latency:** A running average of the response time from the LLM API.
    *   **LLM Circuit Breaker Status:** Shows if the circuit breaker for the LLM API is currently `OPEN`, `HALF_OPEN`, or `CLOSED`.

#### **Panel 4: Live Log Stream**

*   **Purpose:** A real-time, "tail -f" view of the server's structured logs.
*   **Features:**
    *   Displays logs as they are generated (e.g., `INFO`, `WARN`, `ERROR`).
    *   A filter input to show logs only for a specific `gameID`.
    *   Color-coding for different log levels (e.g., errors in red).
