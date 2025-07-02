# ==============================================================================
# Makefile for the Alignment Project
# The single source of truth for all development and operational tasks.
# ==============================================================================

# --- Host Environment ---
# Get the current user's UID and GID to pass into Docker, solving permission issues.
# The `export` command makes these variables available to sub-shells, like docker-compose.
UID := $(shell id -u)
GID := $(shell id -g)
export UID GID

# Default command: Show help message
.DEFAULT_GOAL := help

# --- Go/Wasm Build Dependency ---
WASM_FILE := client/public/core.wasm

## --------------------------------------
## PRIMARY WORKFLOWS
## --------------------------------------

.PHONY: dev
dev: ## ✨ INTERACTIVE: Run servers in the foreground with combined, colorized logs.
	@echo ">>> Starting interactive dev servers... (Press Ctrl+C to stop)"
	@npm run dev

.PHONY: build
build: build-backend build-frontend ## 📦 Build all production artifacts (Backend, Wasm, Frontend)
	@echo "✅ Production build complete."

.PHONY: test
test: test-backend test-frontend test-simulation ## 🧪 Run all backend and frontend tests
	@echo "✅ All tests passed!"

.PHONY: test-ci
test-ci: test-backend test-frontend ## 🤖 Run tests suitable for CI (excludes simulation)
	@echo "✅ CI tests passed!"

.PHONY: docs
docs: ## 📚 Generate OpenAPI documentation from code annotations
	@echo ">>> Generating OpenAPI documentation..."
	@cd server && GOWORK=off go run -mod=mod github.com/swaggo/swag/cmd/swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
	@echo "✅ API documentation generated at server/docs/"

## --------------------------------------
## BACKGROUND DEVELOPMENT / E2E TESTING
## --------------------------------------

.PHONY: bg-start
bg-start: ## 🚀 BACKGROUND: Start all services in the background (detached mode).
	@echo ">>> Starting detached environment (backend, frontend, redis)..."
	@docker compose -f docker-compose.dev.yml up -d --build

.PHONY: bg-stop
bg-stop: ## 🛑 BACKGROUND: Stop and clean up all background services.
	@echo ">>> Stopping and cleaning up detached environment..."
	@docker compose -f docker-compose.dev.yml down -v --remove-orphans

.PHONY: bg-logs
bg-logs: ## 📜 BACKGROUND: View live logs from all background services.
	@echo ">>> Tailing logs (Ctrl+C to exit)..."
	@docker compose -f docker-compose.dev.yml logs -f

## --------------------------------------
## DEPENDENCY MANAGEMENT
## --------------------------------------

.PHONY: vendor
vendor: ## 🤝 Synchronize Go backend dependencies into the server/vendor directory.
	@echo ">>> Tidying and vendoring Go modules for the backend..."
	@cd server && go mod tidy
	@echo ">>> Updating workspace vendor directory..."
	@go work vendor

## --------------------------------------
## INDIVIDUAL COMPONENTS
## --------------------------------------

.PHONY: build-backend
build-backend:
	@echo ">>> Building Go backend binary..."
	@cd server && go build -o ../alignment-server ./cmd/server/

.PHONY: build-frontend
build-frontend: $(WASM_FILE)
	@echo ">>> Building React frontend..."
	@cd client && npm run build

$(WASM_FILE): core/*.go client/wasm/main.go
	@echo ">>> Building Go/Wasm core module..."
	@cd client/wasm && GOOS=js GOARCH=wasm go build -o ../public/core.wasm .

.PHONY: test-backend
test-backend:
	@echo ">>> Running backend tests..."
	@cd server && go test -race -cover ./...

.PHONY: test-frontend
test-frontend: ## 🧪 Run frontend tests once
	@echo ">>> Running frontend tests..."
	@cd client && npm test

.PHONY: test-frontend-watch
test-frontend-watch: ## 👀 Run frontend tests in watch mode for development
	@echo ">>> Running frontend tests in watch mode... (Press Ctrl+C to stop)"
	@cd client && npm run test:watch

.PHONY: test-frontend-coverage
test-frontend-coverage: ## 📊 Run frontend tests with coverage report
	@echo ">>> Running frontend tests with coverage..."
	@cd client && npx vitest run --coverage

.PHONY: test-contract-e2e
test-contract-e2e: ## 🔗 Run contract verification against live server (requires running server)
	@echo ">>> Running contract verification against live server..."
	@cd client && E2E_SERVER_URL=http://localhost:8080 npm test -- contract-verification

.PHONY: test-simulation
test-simulation: build-simulator ## 🎯 Run game balance simulation tests
	@echo ">>> Running balance simulation tests..."
	@./cmd/simulator/simulator -runs=50 -ci

.PHONY: build-simulator
build-simulator:
	@echo ">>> Building simulation runner..."
	@cd cmd/simulator && go build -o simulator .

## --------------------------------------
## HELP
## --------------------------------------

.PHONY: help
help: ## 🙋 Show this help message
	@echo
	@echo "Usage: make <target>"
	@echo
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
	@echo