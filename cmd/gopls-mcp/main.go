package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

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
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	mcpServer := server.NewServer(config)
	
	// Start server in a goroutine
	go func() {
		if err := mcpServer.Start(ctx); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")
	
	// Shutdown gracefully
	if err := mcpServer.Shutdown(ctx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	
	fmt.Println("Server stopped")
}