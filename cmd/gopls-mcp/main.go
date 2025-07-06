package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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
		// Configure structured logging first
		configureLogging(logLevel)

		config := types.Config{
			GoplsPath:     goplsPath,
			WorkspaceRoot: workspaceRoot,
			LogLevel:      logLevel,
		}

		// Ensure the workspace root is a valid directory
		if stat, err := os.Stat(config.WorkspaceRoot); err != nil || !stat.IsDir() {
			slog.Error("Invalid workspace root", "path", config.WorkspaceRoot, "error", err)
			os.Exit(1)
		}

		// Ensure the workspace root is an absolute path
		if absPath, err := filepath.Abs(config.WorkspaceRoot); err == nil {
			config.WorkspaceRoot = absPath
		}

		srv := server.NewGoplsServer(config)
		slog.Info("Starting Gopls MCP server",
			"gopls_path", config.GoplsPath,
			"workspace_root", config.WorkspaceRoot,
			"log_level", config.LogLevel)

		if err := srv.Serve(context.Background()); err != nil {
			slog.Error("Failed to serve Gopls MCP server", "error", err)
			os.Exit(1)
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVar(&goplsPath, "gopls-path", "gopls", "Path to the gopls binary")
	rootCmd.Flags().StringVar(&workspaceRoot, "workspace-root", ".", "Root directory of the Go workspace")
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}

// configureLogging sets up structured logging with the specified log level
func configureLogging(level string) {
	var logLevelVar slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevelVar = slog.LevelDebug
	case "info":
		logLevelVar = slog.LevelInfo
	case "warn", "warning":
		logLevelVar = slog.LevelWarn
	case "error":
		logLevelVar = slog.LevelError
	default:
		logLevelVar = slog.LevelInfo
	}

	// Create a JSON handler for structured logging
	opts := &slog.HandlerOptions{
		Level:     logLevelVar,
		AddSource: logLevelVar == slog.LevelDebug, // Add source info for debug level
	}

	handler := slog.NewJSONHandler(os.Stderr, opts)
	logger := slog.New(handler)

	// Set as the default logger
	slog.SetDefault(logger)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Error executing command", "error", err)
		os.Exit(1)
	}
}
