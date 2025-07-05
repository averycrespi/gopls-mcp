package server

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopls-mcp/internal/lsp"
	"gopls-mcp/pkg/types"
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
	mcpServer := server.NewMCPServer("gopls-mcp", "1.0.0")
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
	// Go to definition tool
	goToDefTool := mcp.NewTool("gopls.go_to_definition",
		mcp.WithDescription("Find the definition of a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	s.mcpServer.AddTool(goToDefTool, s.handleGoToDefinition)
	
	// Find references tool
	findRefsTool := mcp.NewTool("gopls.find_references",
		mcp.WithDescription("Find all references to a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	s.mcpServer.AddTool(findRefsTool, s.handleFindReferences)
	
	// Hover info tool
	hoverTool := mcp.NewTool("gopls.hover_info",
		mcp.WithDescription("Get hover information for a symbol in Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	s.mcpServer.AddTool(hoverTool, s.handleHoverInfo)
	
	// Get completion tool
	completionTool := mcp.NewTool("gopls.get_completion",
		mcp.WithDescription("Get code completion suggestions for Go code"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
	)
	s.mcpServer.AddTool(completionTool, s.handleGetCompletion)
	
	// Format code tool
	formatTool := mcp.NewTool("gopls.format_code",
		mcp.WithDescription("Format Go code using gofmt"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
	)
	s.mcpServer.AddTool(formatTool, s.handleFormatCode)
	
	// Rename symbol tool
	renameTool := mcp.NewTool("gopls.rename_symbol",
		mcp.WithDescription("Rename a symbol across the Go project"),
		mcp.WithString("file_path", mcp.Required(), mcp.Description("Path to the Go file")),
		mcp.WithNumber("line", mcp.Required(), mcp.Description("Line number (0-based)")),
		mcp.WithNumber("character", mcp.Required(), mcp.Description("Character position (0-based)")),
		mcp.WithString("new_name", mcp.Required(), mcp.Description("New name for the symbol")),
	)
	s.mcpServer.AddTool(renameTool, s.handleRenameSymbol)
	
	return nil
}

// Helper function to get file URI from path
func (s *Server) getFileURI(filePath string) string {
	if strings.HasPrefix(filePath, "file://") {
		return filePath
	}
	
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(s.config.WorkspaceRoot, filePath)
	}
	
	return "file://" + filePath
}

// Helper function to get position from request
func (s *Server) getPosition(req mcp.CallToolRequest) (types.Position, error) {
	line := mcp.ParseFloat64(req, "line", 0)
	character := mcp.ParseFloat64(req, "character", 0)
	
	return types.Position{
		Line:      int(line),
		Character: int(character),
	}, nil
}

// Tool handlers
func (s *Server) handleGoToDefinition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := s.lspManager.GetClient()
	if client == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}
	
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}
	
	position, err := s.getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	uri := s.getFileURI(filePath)
	locations, err := client.GoToDefinition(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get definition: %v", err)), nil
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Found %d definition(s): %+v", len(locations), locations)), nil
}

func (s *Server) handleFindReferences(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := s.lspManager.GetClient()
	if client == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}
	
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}
	
	position, err := s.getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	uri := s.getFileURI(filePath)
	locations, err := client.FindReferences(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to find references: %v", err)), nil
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Found %d reference(s): %+v", len(locations), locations)), nil
}

func (s *Server) handleHoverInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := s.lspManager.GetClient()
	if client == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}
	
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}
	
	position, err := s.getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	uri := s.getFileURI(filePath)
	hover, err := client.Hover(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get hover info: %v", err)), nil
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Hover info: %s", hover)), nil
}

func (s *Server) handleGetCompletion(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := s.lspManager.GetClient()
	if client == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}
	
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}
	
	position, err := s.getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	uri := s.getFileURI(filePath)
	completions, err := client.GetCompletion(ctx, uri, position)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get completions: %v", err)), nil
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Found %d completion(s): %+v", len(completions), completions)), nil
}

func (s *Server) handleFormatCode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := s.lspManager.GetClient()
	if client == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}
	
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}
	
	uri := s.getFileURI(filePath)
	edits, err := client.FormatDocument(ctx, uri)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format document: %v", err)), nil
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Formatting complete. Applied %d edit(s): %+v", len(edits), edits)), nil
}

func (s *Server) handleRenameSymbol(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := s.lspManager.GetClient()
	if client == nil {
		return mcp.NewToolResultError("LSP client not initialized"), nil
	}
	
	filePath := mcp.ParseString(req, "file_path", "")
	if filePath == "" {
		return mcp.NewToolResultError("file_path parameter is required"), nil
	}
	
	newName := mcp.ParseString(req, "new_name", "")
	if newName == "" {
		return mcp.NewToolResultError("new_name parameter is required"), nil
	}
	
	position, err := s.getPosition(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	
	uri := s.getFileURI(filePath)
	changes, err := client.RenameSymbol(ctx, uri, position, newName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to rename symbol: %v", err)), nil
	}
	
	return mcp.NewToolResultText(fmt.Sprintf("Rename complete. Changed %d file(s): %+v", len(changes), changes)), nil
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