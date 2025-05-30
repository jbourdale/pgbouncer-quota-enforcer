# Variables
BINARY_NAME=pgbouncer-quota-enforcer
MAIN_PATH=./cmd
BUILD_DIR=./bin

# Default target
.DEFAULT_GOAL := build

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run unit tests only (exclude integration tests)
test:
	@echo "Running unit tests..."
	go test -v $(shell go list ./... | grep -v /test/)
	@echo "Unit tests complete"

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./test/...
	@echo "Integration tests complete"

# Run all tests (unit + integration)
test-all: test test-integration
	@echo "All tests complete"

# Run linter
lint:
	golangci-lint run ./...
	@echo "Linting complete"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# Run the application (show help)
run: build
	@echo "Running $(BINARY_NAME)..."
	$(BUILD_DIR)/$(BINARY_NAME) --help

# Run the TCP server
server: build
	@echo "Starting TCP server on port 8080..."
	$(BUILD_DIR)/$(BINARY_NAME) server --address :8080

# Run integration test demo
demo:
	@echo "Running integration test demo..."
	./scripts/demo-integration.sh

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Full check (fmt, vet, lint, unit tests)
check: fmt vet lint test

# Full check including integration tests
check-all: fmt vet lint test-all

# Help
help:
	@echo "Available targets:"
	@echo "  build            - Build the application"
	@echo "  test             - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-all         - Run all tests (unit + integration)"
	@echo "  lint             - Run linter"
	@echo "  clean            - Clean build artifacts"
	@echo "  run              - Build and show help"
	@echo "  server           - Build and run the TCP server on port 8080"
	@echo "  demo             - Run integration test demonstration"
	@echo "  deps             - Install dependencies"
	@echo "  fmt              - Format code"
	@echo "  vet              - Vet code"
	@echo "  check            - Run fmt, vet, lint, and unit tests"
	@echo "  check-all        - Run fmt, vet, lint, and all tests"
	@echo "  help             - Show this help message"

.PHONY: build test test-integration test-all lint clean run server demo deps fmt vet check check-all help 