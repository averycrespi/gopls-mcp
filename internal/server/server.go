package server

import (
	"context"
	"fmt"

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
	mcpServer := server.NewMCPServer(project.Name, project.Version)
	goplsClient := client.NewGoplsClient(config.GoplsPath)

	return &GoplsServer{
		mcpServer:   mcpServer,
		goplsClient: goplsClient,
		config:      config,
	}
}

func (s *GoplsServer) Serve(ctx context.Context) error {
	if err := s.goplsClient.Start(ctx, s.config.WorkspaceRoot); err != nil {
		return fmt.Errorf("failed to start Gopls client: %w", err)
	}

	s.registerTools()

	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("failed to serve on stdio: %w", err)
	}

	return nil
}

func (s *GoplsServer) registerTools() {
	symbolDefTool := tools.NewSymbolDefinitionTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(symbolDefTool.GetTool(), symbolDefTool.Handle)

	findRefsTool := tools.NewFindReferencesTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(findRefsTool.GetTool(), findRefsTool.Handle)

	hoverTool := tools.NewHoverInfoTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(hoverTool.GetTool(), hoverTool.Handle)

	completionTool := tools.NewGetCompletionTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(completionTool.GetTool(), completionTool.Handle)

	symbolSearchTool := tools.NewSymbolSearchTool(s.goplsClient, s.config)
	s.mcpServer.AddTool(symbolSearchTool.GetTool(), symbolSearchTool.Handle)

}
