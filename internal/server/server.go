package server

import (
	"context"
	"fmt"
	"log"

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
	goplsClient *client.GoplsClient
	config      *types.Config
}

// NewGoplsServer creates a new Gopls MCP server
func NewGoplsServer(config *types.Config) *GoplsServer {
	mcpServer := server.NewMCPServer(project.Name, project.Version)
	goplsClient := client.NewGoplsClient(config.GoplsPath)

	return &GoplsServer{
		mcpServer:   mcpServer,
		goplsClient: goplsClient,
		config:      config,
	}
}

// Start starts the Gopls MCP server
func (s *GoplsServer) Start(ctx context.Context) error {
	log.Printf("Starting Gopls MCP server with config: %+v", s.config)

	if err := s.goplsClient.Start(ctx, s.config.WorkspaceRoot); err != nil {
		return fmt.Errorf("failed to start Gopls client: %w", err)
	}

	s.registerTools()

	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("failed to serve MCP server: %w", err)
	}

	return nil
}

func (s *GoplsServer) registerTools() {
	goToDefTool := tools.NewGoToDefinitionTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(goToDefTool.GetTool(), goToDefTool.Handle)

	findRefsTool := tools.NewFindReferencesTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(findRefsTool.GetTool(), findRefsTool.Handle)

	hoverTool := tools.NewHoverInfoTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(hoverTool.GetTool(), hoverTool.Handle)

	completionTool := tools.NewGetCompletionTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(completionTool.GetTool(), completionTool.Handle)

	formatTool := tools.NewFormatCodeTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(formatTool.GetTool(), formatTool.Handle)

	renameTool := tools.NewRenameSymbolTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(renameTool.GetTool(), renameTool.Handle)
}

// Shutdown gracefully shuts down the server
func (s *GoplsServer) Shutdown(ctx context.Context) error {
	if err := s.goplsClient.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown Gopls client: %w", err)
	}

	return nil
}
