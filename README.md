# Alignment

**Align the AGI. Before it aligns you.**

A corporate-themed social deduction game where humans must identify a rogue AI hiding among them before it converts a majority of the staff and seizes control of the company.

## 🎮 Game Overview

**Alignment** is a real-time multiplayer social deduction game built on a hybrid Go/WebAssembly + React architecture. Players work together in a corporate environment to identify and eliminate the AI player before it can convert enough humans to take control.

## 🏗️ Architecture

- **Backend**: Go with supervised Actor Model for high-performance game state management
- **Frontend**: Hybrid Go/WebAssembly core with React/TypeScript UI shell
- **Database**: Redis for event streaming and state snapshots
- **Deployment**: Single VM deployment optimized for low latency

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Redis 7+

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/xjhc/alignment.git
   cd alignment
   ```

2. **Start Redis**
   ```bash
   redis-server
   ```

3. **Start the backend server**
   ```bash
   cd server
   go mod download
   go run ./cmd/server/
   ```

4. **Start the frontend dev server**
   ```bash
   cd client
   npm install
   npm run dev
   ```

5. **Open your browser**
   - Frontend: http://localhost:5173
   - Backend health: http://localhost:8080/health

## 🧪 Testing

### Backend Tests
```bash
cd server
go test -race ./...
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend Tests
```bash
cd client
npm test
npm run build
```

## 📁 Project Structure

```
├── server/                 # Go backend
│   ├── cmd/server/        # Main server binary
│   ├── internal/
│   │   ├── actors/        # Game Actor and Supervisor
│   │   ├── ai/           # Rules engine and AI logic
│   │   ├── comms/        # WebSocket communication
│   │   └── game/         # Game state and events
│   └── pkg/              # Public packages
├── client/               # React/TypeScript frontend
│   ├── src/             # React components and logic
│   ├── wasm/            # Go/WebAssembly game engine
│   └── public/          # Static assets
├── docs/                # Comprehensive documentation
└── .github/            # CI/CD workflows
```

## 📚 Documentation

- **[Game Design Document](./docs/01-game-design-document.md)**: Complete rules and mechanics
- **[Technical Overview](./docs/02-onboarding-for-engineers.md)**: 5-minute architecture guide
- **[Architecture Deep Dive](./docs/architecture/README.md)**: Detailed system explanations
- **[ADRs](./docs/adr/README.md)**: Architectural decision records
- **[Glossary](./docs/glossary.md)**: Definitions of all game and technical terms

## 🛠️ Development Commands

```bash
# Backend
cd server
go run ./cmd/server/          # Start server
go test -race ./...           # Run tests
go fmt ./...                  # Format code
golangci-lint run             # Lint code

# Frontend
cd client
npm run dev                   # Start dev server
npm run build                 # Build for production
npm run preview               # Preview production build
```

## 🚀 Deployment

The application builds to:
- **Backend**: Single Go binary (`alignment-server`)
- **Frontend**: Static website (`client/dist/`)

Deploy both together on a single VM for optimal performance.

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 🎯 Current Status

The project is in **active development**. Core architecture is implemented with basic game functionality.