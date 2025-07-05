package server

import (
	"context"
	"fmt"
	"log"

	"github.com/averycrespi/gopls-mcp/internal/client"
	"github.com/averycrespi/gopls-mcp/internal/tools"
	"github.com/averycrespi/gopls-mcp/pkg/project"
	"github.com/averycrespi/gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
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

	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	if err := server.ServeStdio(s.mcpServer); err != nil {
		return fmt.Errorf("failed to serve MCP server: %w", err)
	}

	return nil
}

func (s *GoplsServer) registerTools() error {
	goToDefTool := tools.NewGoToDefinitionTool(s.goplsClient, s.config)
	findRefsTool := tools.NewFindReferencesTool(s.goplsClient, s.config)
	hoverTool := tools.NewHoverInfoTool(s.goplsClient, s.config)
	completionTool := tools.NewGetCompletionTool(s.goplsClient, s.config)
	formatTool := tools.NewFormatCodeTool(s.goplsClient, s.config)
	renameTool := tools.NewRenameSymbolTool(s.goplsClient, s.config)

	// Register tools with custom handlers that inject the LSP client
	// TODO: remove these
	s.mcpServer.AddTool(*goToDefTool.GetTool(), s.wrapHandler(goToDefTool.Handle))
	s.mcpServer.AddTool(*findRefsTool.GetTool(), s.wrapHandler(findRefsTool.Handle))
	s.mcpServer.AddTool(*hoverTool.GetTool(), s.wrapHandler(hoverTool.Handle))
	s.mcpServer.AddTool(*completionTool.GetTool(), s.wrapHandler(completionTool.Handle))
	s.mcpServer.AddTool(*formatTool.GetTool(), s.wrapHandler(formatTool.Handle))
	s.mcpServer.AddTool(*renameTool.GetTool(), s.wrapHandler(renameTool.Handle))

	return nil
}

// wrapHandler wraps tool handlers to inject the LSP client dynamically
func (s *GoplsServer) wrapHandler(handler func(context.Context, types.Client, *types.Config, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handler(ctx, s.goplsClient, s.config, req)
	}
}

// Shutdown gracefully shuts down the server
func (s *GoplsServer) Shutdown(ctx context.Context) error {
	if err := s.goplsClient.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown Gopls client: %w", err)
	}

	return nil
}
