# Alignment

`Alignment` is a corporate-themed social deduction game played in a real-time chat environment. Humans must work together to identify a rogue AI hiding among them before it converts a majority of the staff (by voting power) and seizes control of the company.

This repository contains the full source code for the game, including the Go backend server and the hybrid Go/Wasm + React frontend.

---

## Quick Links

New to the project? Start here.

*   **[Game Design Document](./01-game-design-document.md):** What is the game? Read this for the rules, roles, and core concepts.
*   **[Onboarding for Engineers](./02-onboarding-for-engineers.md):** The 5-minute technical overview of the entire stack.
*   **[Static UI Mocks](../design/README.md):** The visual reference for the application.
*   **[Architectural Deep Dive](./architecture/README.md):** Detailed explanations of the backend systems (Actors, AI, etc.).
*   **[Architectural Decisions (ADRs)](./adr/README.md):** The "why" behind our key technical choices.

---

## 🛠️ Getting Started: Running Locally

To run the full application locally, you will need Go, Node.js, and Redis installed.

#### 1. Start Dependencies

Ensure your local Redis server is running:
```bash
redis-server
```

#### 2. Start the Backend Server

Navigate to the server directory and run the application:
```bash
cd server/
go run ./cmd/server/
```
The server will be running on `localhost:8080` by default.

#### 3. Start the Frontend Dev Server

In a separate terminal, navigate to the client directory, install dependencies, and start the Vite dev server:
```bash
cd client/
npm install
npm run dev
```
The frontend will be accessible at `http://localhost:5173`. Any changes to the React/TypeScript code will hot-reload.
