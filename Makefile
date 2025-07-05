.PHONY: build test clean install help run

# Default target
all: build

# Build the binary
build:
	go build -o bin/gopls-mcp ./cmd/gopls-mcp

# Run tests
test:
	go test ./...

# Build and run basic integration tests
test-integration: build
	@echo "Testing binary exists and is executable..."
	@test -x bin/gopls-mcp || (echo "ERROR: Binary not found or not executable" && exit 1)
	@echo "Running unit tests..."
	@go test ./...
	@echo "Checking gopls is available..."
	@command -v gopls >/dev/null || echo "WARNING: gopls not found in PATH. Install with: go install golang.org/x/tools/gopls@latest"
	@echo "All tests passed! ✅"

# Run full integration tests including MCP server testing
test-integration-full: build
	@echo "Testing binary exists and is executable..."
	@test -x bin/gopls-mcp || (echo "ERROR: Binary not found or not executable" && exit 1)
	@echo "Running unit tests..."
	@go test ./...
	@echo "Checking gopls is available..."
	@command -v gopls >/dev/null || (echo "ERROR: gopls not found in PATH. Install with: go install golang.org/x/tools/gopls@latest" && exit 1)
	@echo "Running full integration tests..."
	@go test -v . -run TestMCPServer
	@echo "All integration tests passed! ✅"

# Clean build artifacts
clean:
	rm -rf bin/
	go clean -cache -testcache

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install gopls if not present
install-gopls:
	go install golang.org/x/tools/gopls@latest

# Run the server (requires workspace-root parameter)
run:
	@if [ -z "$(WORKSPACE)" ]; then \
		echo "Usage: make run WORKSPACE=/path/to/go/project"; \
		exit 1; \
	fi
	./bin/gopls-mcp -workspace-root $(WORKSPACE)

# Development: build and run with current directory
dev: build
	./bin/gopls-mcp -workspace-root .

# Show help
help:
	@echo "Available targets:"
	@echo "  build                 Build the gopls-mcp binary"
	@echo "  test                  Run unit tests"
	@echo "  test-integration      Run basic integration tests"
	@echo "  test-integration-full Run full integration tests including MCP server"
	@echo "  clean                 Clean build artifacts"
	@echo "  deps                  Download and tidy dependencies"
	@echo "  install-gopls         Install gopls language server"
	@echo "  run                   Run server (requires WORKSPACE=/path/to/project)"
	@echo "  dev                   Build and run with current directory"
	@echo "  help                  Show this help message"