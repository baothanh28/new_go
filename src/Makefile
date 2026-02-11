.PHONY: help build run test clean migrate docker-up docker-down

# Detect OS
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    RM := cmd /C del /Q /F
    RMDIR := cmd /C rmdir /S /Q
    SLEEP := timeout /T 3 /NOBREAK >nul
    INIT_MASTER_DB := ..\deployment\init-master-db.bat
else
    DETECTED_OS := $(shell uname -s)
    RM := rm -f
    RMDIR := rm -rf
    SLEEP := sleep 3
    INIT_MASTER_DB := ../deployment/init-master-db.sh
endif

# Variables
APP_NAME=myapp
BUILD_DIR=bin
MAIN_PATH=cmd/api/main.go
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

help: ## Display this help message
	@echo "Available commands:"
	@echo "  build           - Build the application binary"
	@echo "  run             - Run the application (serve command)"
	@echo "  dev             - Run the application in development mode with hot reload"
	@echo "  product         - Run product service with hot reload"
	@echo "  health          - Run health service with hot reload"
	@echo "  test            - Run tests"
	@echo "  migrate         - Run database migrations"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download dependencies"
	@echo "  docker-up       - Start PostgreSQL databases with Docker Compose"
	@echo "  docker-down     - Stop Docker containers"
	@echo "  docker-logs     - View Docker container logs"
	@echo "  init-master-db  - Initialize master database with tenants table and seed data"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  version         - Display version information"

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

product: ## Run product service with hot reload (requires air)
	@echo "Starting product service in development mode..."
	@air -c .air.product.toml

health: ## Run health service with hot reload (requires air)
	@echo "Starting health service in development mode..."
	@air -c .air.health.toml

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
ifeq ($(OS),Windows_NT)
	@if exist $(BUILD_DIR) $(RMDIR) $(BUILD_DIR) 2>nul || echo Build dir clean
	@if exist coverage.out $(RM) coverage.out 2>nul || echo Coverage files clean
	@if exist coverage.html $(RM) coverage.html 2>nul || echo Coverage files clean
else
	@$(RMDIR) $(BUILD_DIR) 2>/dev/null || true
	@$(RM) coverage.out coverage.html 2>/dev/null || true
endif
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
	@$(SLEEP)
	@echo "Docker containers started"

docker-down: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	@docker-compose -f ../deployment/docker-compose.yml down
	@echo "Docker containers stopped"

docker-logs: ## View Docker container logs
	@docker-compose -f ../deployment/docker-compose.yml logs -f

init-master-db: ## Initialize master database with tenants table and seed data
	@echo "Initializing master database..."
	@$(INIT_MASTER_DB)

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
