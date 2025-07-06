.PHONY: build test test-integration clean install help run test-symbol-definition test-symbol-references test-file-symbols

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
	@echo "Testing symbol_definition tool..."
	@timeout 10s bash -c '{ \
		tr -d "\n" < ./testdata/mcp_init.json; echo ""; \
		sleep 1; \
		tr -d "\n" < ./testdata/symbol_definition_test.json; echo ""; \
	} | ./bin/gopls-mcp -workspace-root ./testdata/example || true'

# Test symbol references tool
test-symbol-references: build
	@timeout 10s bash -c '{ \
		tr -d "\n" < ./testdata/mcp_init.json; echo ""; \
		sleep 1; \
		tr -d "\n" < ./testdata/symbol_references_test.json; echo ""; \
	} | ./bin/gopls-mcp -workspace-root ./testdata/example || true'

# Test file symbols tool
test-file-symbols: build
	@timeout 10s bash -c '{ \
		tr -d "\n" < ./testdata/mcp_init.json; echo ""; \
		sleep 1; \
		tr -d "\n" < ./testdata/file_symbols_test.json; echo ""; \
	} | ./bin/gopls-mcp -workspace-root ./testdata/example || true'


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
	@echo "  test-file-symbols      Test file_symbols MCP tool"
	@echo "  help                   Show this help message"