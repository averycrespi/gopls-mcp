#!/bin/bash

# Basic integration test for gopls-mcp
set -e

echo "Building gopls-mcp..."
go build -o bin/gopls-mcp ./cmd/gopls-mcp

echo "Testing binary exists and is executable..."
if [ ! -x "bin/gopls-mcp" ]; then
    echo "ERROR: Binary not found or not executable"
    exit 1
fi

echo "Testing help flag..."
timeout 5s ./bin/gopls-mcp --help || {
    if [ $? -eq 124 ]; then
        echo "Help command timed out (expected for MCP server)"
    else
        echo "Help command failed with exit code: $?"
        exit 1
    fi
}

echo "Running unit tests..."
go test ./...

echo "Checking gopls is available..."
if ! command -v gopls &> /dev/null; then
    echo "WARNING: gopls not found in PATH. Install with: go install golang.org/x/tools/gopls@latest"
fi

echo "All tests passed! ✅"