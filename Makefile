
.PHONY: help dev build clean test test-ci build-backend build-frontend build-simulator test-backend test-frontend test-simulation generate-types vendor

# ==============================================================================
# HELP
# ==============================================================================

help:
	@echo "Usage: make <command>"
	@echo ""
	@echo "Available commands:"
	@echo "  dev                 Run development servers for backend and frontend"
	@echo "  build               Build production binaries and assets"
	@echo "  clean               Remove build artifacts"
	@echo "  test                Run all tests (backend and frontend)"
	@echo "  test-ci             Run all tests in CI mode"
	@echo "  build-backend       Build the backend server binary"
	@echo "  build-frontend      Build the frontend production assets"
	@echo "  build-simulator     Build the game balance simulator"
	@echo "  test-backend        Run backend tests with race detection and coverage"
	@echo "  test-frontend       Run frontend tests"
	@echo "  test-simulation     Run game balance simulation tests"
	@echo "  generate-types      Generate TypeScript types from Go core package"
	@echo "  vendor              Vendor Go dependencies for the server"

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

dev: generate-types
	@echo ">>> Starting development servers..."
	@npm run dev

# ==============================================================================
# BUILD
# ==============================================================================

build: build-backend build-frontend build-simulator
	@echo ">>> All builds complete."

build-backend:
	@echo ">>> Building backend server..."
	@cd server && go build -o ../alignment-server ./cmd/server/

build-frontend:
	@echo ">>> Building frontend assets..."
	@cd client && npm install && npm run build

build-simulator:
	@echo ">>> Building game simulator..."
	@cd cmd/simulator && go build -o ../../simulator-bin .

# ==============================================================================
# CLEAN
# ==============================================================================

clean:
	@echo ">>> Cleaning build artifacts..."
	@rm -f alignment-server simulator-bin
	@rm -rf client/dist
	@rm -f server/coverage.out client/coverage.json
	@echo ">>> Clean complete."

# ==============================================================================
# TESTING
# ==============================================================================

test: test-backend test-frontend
	@echo ">>> All tests passed."

test-ci:
	@echo ">>> Running tests in CI mode..."
	@make test-backend
	@make test-frontend
	@make test-simulation
	@echo ">>> CI tests complete."

test-backend:
	@echo ">>> Running backend tests..."
	@cd server && go test -race -cover -coverprofile=coverage.out ./...

test-frontend:
	@echo ">>> Running frontend tests..."
	@cd client && npm install && npm test

test-simulation: build-simulator
	@echo ">>> Running game balance simulations..."
	@./simulator-bin -runs=100 -output=simulation-results.json -ci

# ==============================================================================
# UTILITIES
# ==============================================================================

generate-types:
	@echo ">>> Generating TypeScript types from Go core..."
	@cd tools/generate-types && go run main.go

vendor:
	@echo ">>> Vendoring Go dependencies..."
	@if [ -f go.work ]; then \
		echo "Detected Go workspace, using 'go work vendor'"; \
		go work vendor; \
	else \
		echo "Using standard go mod vendor"; \
		cd server && go mod tidy && go mod vendor; \
	fi

# Background services for E2E testing
bg-start:
	@echo ">>> Starting background services..."
	@docker-compose -f docker-compose.dev.yml up -d --build

bg-stop:
	@echo ">>> Stopping background services..."
	@docker-compose -f docker-compose.dev.yml down