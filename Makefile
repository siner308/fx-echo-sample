# FX Echo Sample Makefile

.PHONY: help build run test clean dev deps lint fmt vet security setup

# Default target
.DEFAULT_GOAL := help

# Variables
APP_NAME := fx-echo-sample
BUILD_DIR := build
BINARY := $(BUILD_DIR)/$(APP_NAME)
GO_FILES := $(shell find . -name '*.go' -not -path './vendor/*')

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## setup: Set up development environment
setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "Created .env file from template"; fi
	@go mod download
	@echo "✅ Setup complete! Edit .env file with your configuration."

## deps: Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

## build: Build the application
build: deps
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BINARY) main.go
	@echo "✅ Build complete: $(BINARY)"

## run: Run the application
run:
	@echo "Starting $(APP_NAME)..."
	@go run main.go

## dev: Run in development mode with hot reload (requires air)
dev:
	@if command -v air >/dev/null 2>&1; then \
		echo "Starting development server with hot reload..."; \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		make run; \
	fi

## test: Run all tests
test:
	@echo "Running tests..."
	@go test ./...

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./...

## test-cover: Run tests with coverage report
test-cover:
	@echo "Running tests with coverage..."
	@go test -cover ./...
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test ./modules/auth/user/ -run Integration

## benchmark: Run benchmark tests
benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. ./pkg/security/

## lint: Run golangci-lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running golangci-lint..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## fmt: Format Go code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

## security: Run security analysis
security:
	@if command -v gosec >/dev/null 2>&1; then \
		echo "Running security analysis..."; \
		gosec ./...; \
	else \
		echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

## clean: Clean build artifacts and caches
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache
	@go clean -testcache
	@echo "✅ Cleanup complete"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(APP_NAME):latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env-file .env $(APP_NAME):latest

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "✅ Development tools installed"

## check: Run all checks (format, vet, lint, test)
check: fmt vet lint test
	@echo "✅ All checks passed"

## ci: Run CI pipeline locally
ci: deps check security
	@echo "✅ CI pipeline completed successfully"

## generate-secrets: Generate random secrets for .env
generate-secrets:
	@echo "Generating random secrets..."
	@echo "ACCESS_TOKEN_SECRET=$(shell openssl rand -base64 32)"
	@echo "REFRESH_TOKEN_SECRET=$(shell openssl rand -base64 32)"
	@echo "ADMIN_TOKEN_SECRET=$(shell openssl rand -base64 32)"
	@echo "KEYCLOAK_CLIENT_SECRET=$(shell openssl rand -base64 32)"
	@echo ""
	@echo "Copy these values to your .env file"

## docs: Serve documentation locally
docs:
	@echo "Documentation available in docs/ directory"
	@echo "Key files:"
	@echo "  - docs/ARCHITECTURE.md - System architecture"
	@echo "  - docs/API_REFERENCE.md - API documentation"
	@echo "  - docs/FX_CONCEPTS.md - Uber FX concepts"
	@echo "  - .claude/CLAUDE.md - Development notes"

## api-test: Quick API test (requires server to be running)
api-test:
	@echo "Testing API endpoints..."
	@echo "Creating test user..."
	@curl -s -X POST http://localhost:8080/api/v1/users \
		-H "Content-Type: application/json" \
		-d '{"name":"Test User","email":"test@example.com","age":25,"password":"password123"}' \
		| jq '.' || echo "jq not installed - install with: brew install jq"
	@echo "\nGetting item types..."
	@curl -s http://localhost:8080/api/v1/items/types | jq '.' || echo "Raw response above"

## watch: Watch for file changes and rebuild
watch:
	@echo "Watching for changes..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . -e ".*" -i "\\.go$$" | xargs -n1 -I{} make build; \
	else \
		echo "fswatch not found. Install with: brew install fswatch"; \
		echo "Using basic watch instead..."; \
		while true; do make build; sleep 2; done; \
	fi