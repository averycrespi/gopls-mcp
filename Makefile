.PHONY: build test test-integration clean install help run

# Default target
all: build

# Build the binary
build:
	go build -o bin/gopls-mcp ./cmd/gopls-mcp

# Run unit tests
test:
	go test ./...

# Run integration tests
test-integration:
	go test -tags=integration ./...

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

# Run the server
run:
	go run ./cmd/gopls-mcp

# Show help
help:
	@echo "Available targets:"
	@echo "  build                 Build the gopls-mcp binary"
	@echo "  test                  Run unit tests"
	@echo "  test-integration      Run integration tests"
	@echo "  clean                 Clean build artifacts"
	@echo "  deps                  Download and tidy dependencies"
	@echo "  install-gopls         Install gopls language server"
	@echo "  run                   Run server"
	@echo "  help                  Show this help message"