# Build variables
BINARY_NAME=otp-auth-backend
BUILD_DIR=bin
MAIN_FILE=cmd/main.go

# Docker variables
DOCKER_IMAGE=otp-auth-backend
DOCKER_TAG=latest

# Go variables
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Default target
.PHONY: all
all: clean build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for specific platform
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_FILE)

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_FILE)

# Build for macOS
.PHONY: build-macos
build-macos:
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_DIR)/$(BINARY_NAME)-macos $(MAIN_FILE)

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GO) run $(MAIN_FILE)

# Run with hot reload (requires air)
.PHONY: dev
dev:
	@echo "Running with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# Test the application
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

# Run race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	$(GO) test -race ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Update dependencies
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

# Generate Swagger documentation
.PHONY: swagger
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_FILE) -o docs; \
	else \
		echo "Swag not found. Installing..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
		swag init -g $(MAIN_FILE) -o docs; \
	fi

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Security audit
.PHONY: audit
audit:
	@echo "Running security audit..."
	$(GO) list -json -deps | nancy sleuth

# Docker operations
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping services..."
	docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs:
	@echo "Showing Docker Compose logs..."
	docker-compose logs -f

# Database operations
.PHONY: db-migrate
db-migrate:
	@echo "Running database migrations..."
	@if command -v migrate > /dev/null; then \
		migrate -path migrations -database "postgres://postgres:password@localhost:5432/otp_auth?sslmode=disable" up; \
	else \
		echo "Migrate not found. Please install golang-migrate"; \
	fi

# Performance profiling
.PHONY: profile
profile:
	@echo "Starting performance profiling..."
	$(GO) run $(MAIN_FILE) &
	@echo "Application started. Use Ctrl+C to stop profiling."
	@echo "Profiling data will be saved to profile.out"
	@echo "Press Enter to start profiling..."
	@read
	curl -o /dev/null -s -w "Total time: %{time_total}s\n" http://localhost:8080/health
	curl -o /dev/null -s -w "Total time: %{time_total}s\n" http://localhost:8080/api/v1/auth/request-otp -X POST -H "Content-Type: application/json" -d '{"phone":"+1234567890"}'
	@echo "Profiling complete. Check profile.out for details."

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  dev             - Run with hot reload"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  benchmark       - Run benchmarks"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Install dependencies"
	@echo "  swagger         - Generate Swagger docs"
	@echo "  fmt             - Format code"
	@echo "  lint            - Lint code"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-compose-up   - Start services"
	@echo "  docker-compose-down  - Stop services"
	@echo "  help            - Show this help"
