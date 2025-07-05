package server

import (
	"context"
	"fmt"
	"log"

	"gopls-mcp/internal/lsp"
	"gopls-mcp/internal/tools"
	"gopls-mcp/pkg/types"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents the MCP server
type Server struct {
	mcpServer   *server.MCPServer
	lspManager  *lsp.Manager
	config      *types.Config
	initialized bool
}

// NewServer creates a new MCP server
func NewServer(config *types.Config) *Server {
	mcpServer := server.NewMCPServer("gopls-mcp", "0.0.1")
	lspManager := lsp.NewManager(config.GoplsPath)

	return &Server{
		mcpServer:  mcpServer,
		lspManager: lspManager,
		config:     config,
	}
}

// Start starts the MCP server
func (s *Server) Start(ctx context.Context) error {
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	if err := s.lspManager.Initialize(ctx, s.config.WorkspaceRoot); err != nil {
		return fmt.Errorf("failed to initialize LSP manager: %w", err)
	}

	s.initialized = true

	log.Printf("gopls-mcp server started with workspace: %s", s.config.WorkspaceRoot)

	return server.ServeStdio(s.mcpServer)
}

// registerTools registers all MCP tools
func (s *Server) registerTools() error {
	// Create tool instances with nil client (will be set dynamically in handlers)
	goToDefTool := tools.NewGoToDefinitionTool(nil, s.config)
	findRefsTool := tools.NewFindReferencesTool(nil, s.config)
	hoverTool := tools.NewHoverInfoTool(nil, s.config)
	completionTool := tools.NewGetCompletionTool(nil, s.config)
	formatTool := tools.NewFormatCodeTool(nil, s.config)
	renameTool := tools.NewRenameSymbolTool(nil, s.config)

	// Register tools with custom handlers that inject the LSP client
	s.mcpServer.AddTool(*goToDefTool.GetTool(), s.wrapHandler(goToDefTool.Handle))
	s.mcpServer.AddTool(*findRefsTool.GetTool(), s.wrapHandler(findRefsTool.Handle))
	s.mcpServer.AddTool(*hoverTool.GetTool(), s.wrapHandler(hoverTool.Handle))
	s.mcpServer.AddTool(*completionTool.GetTool(), s.wrapHandler(completionTool.Handle))
	s.mcpServer.AddTool(*formatTool.GetTool(), s.wrapHandler(formatTool.Handle))
	s.mcpServer.AddTool(*renameTool.GetTool(), s.wrapHandler(renameTool.Handle))

	return nil
}

// wrapHandler wraps tool handlers to inject the LSP client dynamically
func (s *Server) wrapHandler(handler func(context.Context, types.LSPClient, *types.Config, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lspClient := s.lspManager.GetClient()
		return handler(ctx, lspClient, s.config, req)
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.initialized {
		if err := s.lspManager.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown LSP manager: %w", err)
		}
		s.initialized = false
	}
	return nil
}
