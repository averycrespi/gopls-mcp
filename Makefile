.PHONY: build test test-integration clean install help run test-find-symbol-definitions-by-name test-find-symbol-references-by-anchor test-list-symbols-in-file

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

# Test find symbol definitions by name tool
test-find-symbol-definitions-by-name: build
	@./scripts/test-mcp-tool.sh find_symbol_definitions_by_name

# Test find symbol references by anchor tool
test-find-symbol-references-by-anchor: build
	@./scripts/test-mcp-tool.sh find_symbol_references_by_anchor

# Test list symbols in file tool
test-list-symbols-in-file: build
	@./scripts/test-mcp-tool.sh list_symbols_in_file

# Show help
help:
	@echo "Available targets:"
	@echo "  build                                    Build the gopls-mcp binary"
	@echo "  test                                     Run unit tests"
	@echo "  test-integration                         Run integration tests"
	@echo "  clean                                    Clean build artifacts"
	@echo "  deps                                     Download and tidy dependencies"
	@echo "  install-gopls                            Install gopls language server"
	@echo "  run                                      Run server"
	@echo "  test-find-symbol-definitions-by-name     Test find_symbol_definitions_by_name MCP tool"
	@echo "  test-find-symbol-references-by-anchor    Test find_symbol_references_by_anchor MCP tool"
	@echo "  test-list-symbols-in-file                Test list_symbols_in_file MCP tool"
	@echo "  help                                     Show this help message"
