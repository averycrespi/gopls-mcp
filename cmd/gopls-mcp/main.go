package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/averycrespi/gopls-mcp/internal/server"
	"github.com/averycrespi/gopls-mcp/pkg/types"
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

	if stat, err := os.Stat(config.WorkspaceRoot); err != nil || !stat.IsDir() {
		log.Fatalf("Invalid workspace root: %s", config.WorkspaceRoot)
	}

	if absPath, err := filepath.Abs(config.WorkspaceRoot); err == nil {
		config.WorkspaceRoot = absPath
	}

	srv := server.NewGoplsServer(config)
	if err := srv.Start(context.Background()); err != nil {
		log.Fatalf("Failed to start Gopls MCP server: %v", err)
	}

	fmt.Println("Gopls MCP server stopped")
	os.Exit(0)
}
