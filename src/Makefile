.PHONY: help build run test clean migrate docker-up docker-down

# Variables
APP_NAME=myapp
BUILD_DIR=bin
MAIN_PATH=cmd/api/main.go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	@go build -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

run: ## Run the application (serve command)
	@echo "Starting $(APP_NAME)..."
	@go run $(MAIN_PATH) serve

dev: ## Run the application in development mode with hot reload (requires air)
	@echo "Starting $(APP_NAME) in development mode..."
	@air

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

migrate: ## Run database migrations
	@echo "Running database migrations..."
	@go run $(MAIN_PATH) migrate

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

docker-up: ## Start PostgreSQL databases with Docker Compose
	@echo "Starting Docker containers..."
	@docker-compose -f ../deployment/docker-compose.yml up -d
	@echo "Waiting for databases to be ready..."
	@sleep 3
	@echo "Docker containers started"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose -f ../deployment/docker-compose.yml down
	@echo "Docker containers stopped"

docker-logs: ## View Docker container logs
	@docker-compose -f ../deployment/docker-compose.yml logs -f

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	@go fmt ./...
	@goimports -w .

version: ## Display version information
	@go run $(MAIN_PATH) version

.DEFAULT_GOAL := help
