package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/averycrespi/gopls-mcp/internal/server"
	"github.com/averycrespi/gopls-mcp/pkg/types"
)

var (
	goplsPath     string
	workspaceRoot string
	logLevel      string
)

var rootCmd = &cobra.Command{
	Use:   "gopls-mcp",
	Short: "MCP server that exposes LSP functionality of the gopls language server",
	Long: `gopls-mcp is an MCP (Model Context Protocol) server that bridges LLMs with 
the Go language server (gopls), enabling semantic Go code analysis and navigation.

The server provides tools for:
- Finding symbol definitions by name with fuzzy search
- Listing all symbols in a Go file with hierarchical structure  
- Finding all references to a specific symbol by its anchor

All tools return structured JSON responses with precise location information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := types.Config{
			GoplsPath:     goplsPath,
			WorkspaceRoot: workspaceRoot,
			LogLevel:      logLevel,
		}

		// Ensure the workspace root is a valid directory
		if stat, err := os.Stat(config.WorkspaceRoot); err != nil || !stat.IsDir() {
			log.Fatalf("Invalid workspace root: %s", config.WorkspaceRoot)
		}

		// Ensure the workspace root is an absolute path
		if absPath, err := filepath.Abs(config.WorkspaceRoot); err == nil {
			config.WorkspaceRoot = absPath
		}

		srv := server.NewGoplsServer(config)
		log.Printf("Serving Gopls MCP server with config: %+v", config)
		if err := srv.Serve(context.Background()); err != nil {
			log.Fatalf("Failed to serve Gopls MCP server: %v", err)
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVar(&goplsPath, "gopls-path", "gopls", "Path to the gopls binary")
	rootCmd.Flags().StringVar(&workspaceRoot, "workspace-root", ".", "Root directory of the Go workspace")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}
