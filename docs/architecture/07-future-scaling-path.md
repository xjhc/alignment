# Architecture: Future Scaling Path

## 1. Overview

The current V1 architecture is intentionally designed for a **single, powerful virtual machine**. This approach prioritizes simplicity and low operational cost for the initial launch.

However, the core architectural patterns—specifically the stateless application process and the use of Redis as an external state/event store—were chosen because they provide a clear, phased path to a highly available, multi-node architecture if the game's popularity demands it.

This document outlines the strategic, step-by-step plan for scaling `Alignment` beyond a single server.

## 2. The Phased Scaling Plan

The migration from a single VM to a scalable cluster can be achieved in three distinct phases with minimal changes to the core application logic.

#### Phase 1: Separate Infrastructure

The first step is to de-couple the components from the single VM and move them to managed, scalable cloud services.

*   **Go Backend:** Move the application from running as a `systemd` service on a VM to a managed application platform like **Azure App Service** or AWS Elastic Beanstalk.
*   **Redis:** Move the Redis instance from a container on the VM to a managed cache service like **Azure Cache for Redis** or AWS ElastiCache.

**Result:** This phase separates the concerns of compute and storage. We can now scale the application servers and the Redis instance independently. The application code remains virtually unchanged.

#### Phase 2: Horizontal Scaling (Scale Out)

With the application on a managed platform, we can now run multiple instances of the Go backend to handle more concurrent connections.

*   **Load Balancer:** Place an application-aware load balancer, such as an **Azure Application Gateway**, in front of the App Service instances. This gateway will be configured for **WebSocket support and session affinity (sticky sessions)** to ensure a player's connection consistently routes to the same server instance for the duration of their session.
*   **The Broadcasting Problem:** This phase introduces a new challenge. If Player A is connected to Server 1 and Player B is connected to Server 2, how does an action processed by Server 1 result in an event being broadcast to Player B via Server 2?

#### Phase 3: Cross-Instance Communication with Pub/Sub

The final step is to solve the broadcasting problem, enabling all server instances to act as a unified cluster. We will use the Redis instance we already have.

*   **Mechanism: Redis Pub/Sub.**
*   **The Flow:**
    1.  A player action arrives at **Server 1**.
    2.  The Game Actor on Server 1 validates the action and persists the resulting event to the **Redis Stream** (our Write-Ahead Log). This behavior is unchanged.
    3.  After persisting, Server 1 publishes a tiny, lightweight notification message (e.g., `game_updated:g-xyz`) to a global Redis **Pub/Sub channel**.
    4.  **Server 2**, **Server 3**, and all other instances are subscribed to this channel.
    5.  Upon receiving the notification, Server 2 knows that new events are available for game `g-xyz`. It then reads the new event(s) from the Redis Stream and broadcasts them to any clients connected to it for that game.

## 3. Conclusion

This phased approach provides a robust and low-risk path to scaling. Our initial design, centered on a stateless application process that relies on Redis for state recovery, is the key enabler. The most complex application logic within the Game Actor remains untouched throughout this entire infrastructure evolution.
