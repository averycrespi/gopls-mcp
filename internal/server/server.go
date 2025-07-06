package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/averycrespi/gopls-mcp/internal/client"
	"github.com/averycrespi/gopls-mcp/internal/tools"
	"github.com/averycrespi/gopls-mcp/pkg/project"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/server"
)

var _ types.Server = &GoplsServer{}

// GoplsServer represents the Gopls MCP server
type GoplsServer struct {
	mcpServer   *server.MCPServer
	goplsClient types.Client
	config      types.Config
}

// NewGoplsServer creates a new Gopls MCP server
func NewGoplsServer(config types.Config) *GoplsServer {
	slog.Debug("Creating new Gopls MCP server",
		"project_name", project.Name,
		"project_version", project.Version,
		"gopls_path", config.GoplsPath,
		"workspace_root", config.WorkspaceRoot)

	mcpServer := server.NewMCPServer(project.Name, project.Version)
	goplsClient := client.NewGoplsClient(config.GoplsPath)

	return &GoplsServer{
		mcpServer:   mcpServer,
		goplsClient: goplsClient,
		config:      config,
	}
}

func (s *GoplsServer) Serve(ctx context.Context) error {
	slog.Debug("Starting Gopls MCP server", "workspace_root", s.config.WorkspaceRoot)

	if err := s.goplsClient.Start(ctx, s.config.WorkspaceRoot); err != nil {
		slog.Error("Failed to start Gopls client", "error", err, "workspace_root", s.config.WorkspaceRoot)
		return fmt.Errorf("failed to start Gopls client: %w", err)
	}
	slog.Debug("Gopls client started successfully")

	s.registerTools()

	slog.Debug("Starting MCP server on stdio")
	if err := server.ServeStdio(s.mcpServer); err != nil {
		slog.Error("Failed to serve MCP server on stdio", "error", err)
		return fmt.Errorf("failed to serve on stdio: %w", err)
	}

	return nil
}

func (s *GoplsServer) registerTools() {
	slog.Debug("Registering MCP tools")

	findSymbolDefinitionsByNameTool := tools.NewFindSymbolDefinitionsByNameTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(findSymbolDefinitionsByNameTool.GetTool(), findSymbolDefinitionsByNameTool.Handle)
	slog.Debug("Registered tool", "name", "find_symbol_definitions_by_name")

	findSymbolReferencesByAnchorTool := tools.NewFindSymbolReferencesByAnchorTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(findSymbolReferencesByAnchorTool.GetTool(), findSymbolReferencesByAnchorTool.Handle)
	slog.Debug("Registered tool", "name", "find_symbol_references_by_anchor")

	listSymbolsInFileTool := tools.NewListSymbolsInFileTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(listSymbolsInFileTool.GetTool(), listSymbolsInFileTool.Handle)
	slog.Debug("Registered tool", "name", "list_symbols_in_file")

	renameSymbolByAnchorTool := tools.NewRenameSymbolByAnchorTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(renameSymbolByAnchorTool.GetTool(), renameSymbolByAnchorTool.Handle)
	slog.Debug("Registered tool", "name", "rename_symbol_by_anchor")

	slog.Debug("Registered all MCP tools")
}
