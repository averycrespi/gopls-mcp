package server

import (
	"context"
	"fmt"
	"log"

	"github.com/averycrespi/gopls-mcp/internal/lsp"
	"github.com/averycrespi/gopls-mcp/internal/tools"
	"github.com/averycrespi/gopls-mcp/pkg/project"
	"github.com/averycrespi/gopls-mcp/pkg/types"

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
	mcpServer := server.NewMCPServer(project.Name, project.Version)
	lspManager := lsp.NewManager(config.GoplsPath)

	return &Server{
		mcpServer:  mcpServer,
		lspManager: lspManager,
		config:     config,
	}
}

// RegisterTools registers all MCP tools
func (s *Server) RegisterTools() error {
	return s.registerTools()
}

// Start starts the MCP server
func (s *Server) Start() error {
	log.Printf("Starting %s v%s with config: %+v", project.Name, project.Version, s.config)

	// Initialize LSP manager in background (don't block MCP server startup)
	go func() {
		ctx := context.Background()
		if err := s.lspManager.Initialize(ctx, s.config.WorkspaceRoot); err != nil {
			log.Printf("Failed to initialize LSP manager: %v", err)
			// Continue without LSP - tools will return appropriate errors
		} else {
			s.initialized = true
			log.Printf("LSP manager initialized successfully")
		}
	}()

	log.Printf("Running MCP server on stdio")
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
func (s *Server) wrapHandler(handler func(context.Context, types.Client, *types.Config, mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		lspClient := s.lspManager.GetClient()
		if lspClient == nil && s.lspManager.IsInitialized() {
			// LSP manager is initialized but client is nil, something is wrong
			return mcp.NewToolResultError("LSP client is unavailable"), nil
		}
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
