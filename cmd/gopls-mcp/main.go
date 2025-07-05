package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopls-mcp/internal/server"
	"gopls-mcp/pkg/types"
)

func main() {
	var (
		goplsPath     = flag.String("gopls-path", "gopls", "Path to the gopls binary")
		workspaceRoot = flag.String("workspace-root", ".", "Root directory of the Go workspace")
		logLevel      = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	)
	flag.Parse()

	config := &types.Config{
		GoplsPath:     *goplsPath,
		WorkspaceRoot: *workspaceRoot,
		LogLevel:      *logLevel,
	}

	// Validate workspace root
	if stat, err := os.Stat(config.WorkspaceRoot); err != nil || !stat.IsDir() {
		log.Fatalf("Invalid workspace root: %s", config.WorkspaceRoot)
	}

	// Convert to absolute path
	if absPath, err := filepath.Abs(config.WorkspaceRoot); err == nil {
		config.WorkspaceRoot = absPath
	}

	// Create and start the server
	mcpServer := server.NewServer(config)

	// Register tools
	if err := mcpServer.RegisterTools(); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Start the server (this blocks until server shuts down)
	if err := mcpServer.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}

	fmt.Println("Server stopped")
}
