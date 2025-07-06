.PHONY: build test test-integration clean install help run test-symbol-definition test-symbol-references test-list-symbols-in-file

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

# Test symbol definition tool
test-symbol-definition: build
	@./scripts/test-mcp-tool.sh symbol_definition

# Test symbol references tool
test-symbol-references: build
	@./scripts/test-mcp-tool.sh symbol_references

# Test list symbols in file tool
test-list-symbols-in-file: build
	@./scripts/test-mcp-tool.sh list_symbols_in_file


# Show help
help:
	@echo "Available targets:"
	@echo "  build                  Build the gopls-mcp binary"
	@echo "  test                   Run unit tests"
	@echo "  test-integration       Run integration tests"
	@echo "  clean                  Clean build artifacts"
	@echo "  deps                   Download and tidy dependencies"
	@echo "  install-gopls          Install gopls language server"
	@echo "  run                    Run server"
	@echo "  test-symbol-definition Test symbol_definition MCP tool"
	@echo "  test-symbol-references Test symbol_references MCP tool"
	@echo "  test-list-symbols-in-file Test list_symbols_in_file MCP tool"
	@echo "  help                   Show this help message"