package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"

	"github.com/averycrespi/gopls-mcp/internal/transport"
	"github.com/averycrespi/gopls-mcp/pkg/project"
	"github.com/averycrespi/gopls-mcp/pkg/types"
)

const (
	defaultGoplsPath = "gopls"
)

var _ types.Client = &GoplsClient{}

// GoplsClient implements the Client interface for the Gopls LSP server
type GoplsClient struct {
	goplsPath string
	cmd       *exec.Cmd
	stderr    io.ReadCloser
	transport types.Transport
}

// NewGoplsClient creates a new Gopls client
func NewGoplsClient(goplsPath string) *GoplsClient {
	if goplsPath == "" {
		goplsPath = defaultGoplsPath
	}

	slog.Debug("Creating new Gopls client", "gopls_path", goplsPath)

	return &GoplsClient{
		goplsPath: goplsPath,
	}
}

// Start starts the Gopls client
func (c *GoplsClient) Start(ctx context.Context, workspaceRoot string) error {
	slog.Debug("Starting Gopls client", "gopls_path", c.goplsPath, "workspace_root", workspaceRoot)

	c.cmd = exec.CommandContext(ctx, c.goplsPath, "serve")

	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	c.stderr = stderr
	c.transport = transport.NewJsonRpcTransport(stdin, stdout)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gopls command: %w", err)
	}
	slog.Debug("Gopls process started successfully", "pid", c.cmd.Process.Pid)

	if err := c.transport.Start(); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}
	slog.Debug("JSON-RPC transport started successfully")

	rootURI := "file://" + workspaceRoot
	slog.Debug("Initializing Gopls client", "root_uri", rootURI)
	if err := c.initialize(rootURI); err != nil {
		return fmt.Errorf("failed to initialize Gopls client: %w", err)
	}
	slog.Debug("Gopls client initialized successfully")

	return nil
}

func (c *GoplsClient) initialize(rootURI string) error {
	params := map[string]any{
		"processId": nil,
		"clientInfo": map[string]any{
			"name":    project.Name,
			"version": project.Version,
		},
		"rootUri": rootURI,
		"capabilities": map[string]any{
			"textDocument": map[string]any{
				"documentSymbol": map[string]any{
					"hierarchicalDocumentSymbolSupport": true,
				},
			},
		},
	}

	_, err := c.transport.SendRequest("initialize", params)
	if err != nil {
		return fmt.Errorf("failed to send initialization request: %w", err)
	}

	if err := c.transport.SendNotification("initialized", map[string]any{}); err != nil {
		return fmt.Errorf("failed to send initialization notification: %w", err)
	}

	return nil
}

func (c *GoplsClient) Stop(ctx context.Context) error {
	_, err := c.transport.SendRequest("shutdown", nil)
	if err != nil {
		return fmt.Errorf("failed to send JSON-RPC shutdown request: %w", err)
	}

	if err := c.transport.SendNotification("exit", nil); err != nil {
		return fmt.Errorf("failed to send JSON-RPC exit notification: %w", err)
	}

	if err := c.transport.Stop(); err != nil {
		return fmt.Errorf("failed to stop transport: %w", err)
	}

	if c.cmd != nil && c.cmd.Process != nil {
		if err := c.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill gopls process: %w", err)
		}
		if _, err := c.cmd.Process.Wait(); err != nil {
			return fmt.Errorf("failed to wait for gopls process: %w", err)
		}
	}

	return nil
}

func (c *GoplsClient) GoToDefinition(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
	slog.Debug("Getting symbol definition", "uri", uri, "line", position.Line, "character", position.Character)

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.transport.SendRequest("textDocument/definition", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	// LSP definition response can be null, Location, or Location[]
	var rawResponse json.RawMessage
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal definition response: %w", err)
	}

	// Handle null response
	if string(rawResponse) == "null" {
		slog.Debug("No definitions found", "uri", uri)
		return []types.Location{}, nil
	}

	// Try to unmarshal as array first
	var locations []types.Location
	if err := json.Unmarshal(rawResponse, &locations); err != nil {
		// If that fails, try to unmarshal as single location
		var location types.Location
		if err := json.Unmarshal(rawResponse, &location); err != nil {
			return nil, fmt.Errorf("failed to unmarshal definition response: %w", err)
		}
		locations = []types.Location{location}
	}

	slog.Debug("Found symbol definitions", "count", len(locations), "uri", uri)
	return locations, nil
}

func (c *GoplsClient) FindReferences(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
	slog.Debug("Finding symbol references", "uri", uri, "line", position.Line, "character", position.Character)

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
		"context": map[string]any{
			"includeDeclaration": true,
		},
	}

	response, err := c.transport.SendRequest("textDocument/references", params)
	if err != nil {
		return nil, fmt.Errorf("failed to find references: %w", err)
	}

	// LSP references response can be null or Location[]
	var rawResponse json.RawMessage
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal references response: %w", err)
	}

	// Handle null response
	if string(rawResponse) == "null" {
		slog.Debug("No references found", "uri", uri)
		return []types.Location{}, nil
	}

	var locations []types.Location
	if err := json.Unmarshal(rawResponse, &locations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal references response: %w", err)
	}

	slog.Debug("Found symbol references", "count", len(locations), "uri", uri)
	return locations, nil
}

func (c *GoplsClient) GetHoverInfo(ctx context.Context, uri string, position types.Position) (string, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.transport.SendRequest("textDocument/hover", params)
	if err != nil {
		return "", fmt.Errorf("failed to get hover: %w", err)
	}

	var hover struct {
		Contents any `json:"contents"`
	}
	if err := json.Unmarshal(response, &hover); err != nil {
		return "", fmt.Errorf("failed to unmarshal hover response: %w", err)
	}

	// Handle different content formats
	switch v := hover.Contents.(type) {
	case string:
		return v, nil
	case map[string]any:
		if value, ok := v["value"]; ok {
			return fmt.Sprintf("%v", value), nil
		}
	}

	return fmt.Sprintf("%v", hover.Contents), nil
}

func (c *GoplsClient) FuzzyFindSymbol(ctx context.Context, query string) ([]types.SymbolInformation, error) {
	slog.Debug("Fuzzy finding symbols", "query", query)

	params := map[string]any{
		"query": query,
	}

	response, err := c.transport.SendRequest("workspace/symbol", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace symbols: %w", err)
	}

	// LSP workspace/symbol response can be null or SymbolInformation[]
	var rawResponse json.RawMessage
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspace symbol response: %w", err)
	}

	// Handle null response
	if string(rawResponse) == "null" {
		slog.Debug("No symbols found", "query", query)
		return []types.SymbolInformation{}, nil
	}

	var symbols []types.SymbolInformation
	if err := json.Unmarshal(rawResponse, &symbols); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspace symbol response: %w", err)
	}

	slog.Debug("Found symbols", "count", len(symbols), "query", query)
	return symbols, nil
}

func (c *GoplsClient) FormatDocument(ctx context.Context, uri string) ([]json.RawMessage, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"options": map[string]any{
			"tabSize":      4,
			"insertSpaces": false,
		},
	}

	response, err := c.transport.SendRequest("textDocument/formatting", params)
	if err != nil {
		return nil, fmt.Errorf("failed to format document: %w", err)
	}

	var edits []json.RawMessage
	if err := json.Unmarshal(response, &edits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal formatting response: %w", err)
	}

	return edits, nil
}

func (c *GoplsClient) PrepareRename(ctx context.Context, uri string, position types.Position) (*types.PrepareRenameResult, error) {
	slog.Debug("Preparing rename", "uri", uri, "line", position.Line, "character", position.Character)

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.transport.SendRequest("textDocument/prepareRename", params)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare rename: %w", err)
	}

	// LSP prepareRename response can be null, Range, or {range, placeholder}
	var rawResponse json.RawMessage
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prepareRename response: %w", err)
	}

	// Handle null response (rename not allowed)
	if string(rawResponse) == "null" {
		return nil, fmt.Errorf("rename not allowed at this position")
	}

	// Try to unmarshal as PrepareRenameResult (with placeholder)
	var result types.PrepareRenameResult
	if err := json.Unmarshal(rawResponse, &result); err != nil {
		// If that fails, try to unmarshal as just Range
		var rangeOnly types.Range
		if err := json.Unmarshal(rawResponse, &rangeOnly); err != nil {
			return nil, fmt.Errorf("failed to unmarshal prepareRename response: %w", err)
		}
		result = types.PrepareRenameResult{
			Range: rangeOnly,
		}
	}

	slog.Debug("Rename prepared", "uri", uri, "range", result.Range, "placeholder", result.Placeholder)
	return &result, nil
}

func (c *GoplsClient) RenameSymbol(ctx context.Context, uri string, position types.Position, newName string) (*types.WorkspaceEdit, error) {
	slog.Debug("Renaming symbol", "uri", uri, "line", position.Line, "character", position.Character, "new_name", newName)

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
		"newName":  newName,
	}

	response, err := c.transport.SendRequest("textDocument/rename", params)
	if err != nil {
		return nil, fmt.Errorf("failed to rename symbol: %w", err)
	}

	// LSP rename response can be null or WorkspaceEdit
	var rawResponse json.RawMessage
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rename response: %w", err)
	}

	// Handle null response
	if string(rawResponse) == "null" {
		slog.Debug("No rename performed", "uri", uri)
		return &types.WorkspaceEdit{Changes: make(map[string][]types.TextEdit)}, nil
	}

	var workspaceEdit types.WorkspaceEdit
	if err := json.Unmarshal(rawResponse, &workspaceEdit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rename response: %w", err)
	}

	editCount := 0
	for _, edits := range workspaceEdit.Changes {
		editCount += len(edits)
	}
	slog.Debug("Symbol renamed", "uri", uri, "file_count", len(workspaceEdit.Changes), "edit_count", editCount)

	return &workspaceEdit, nil
}

func (c *GoplsClient) GetDocumentSymbols(ctx context.Context, uri string) ([]types.DocumentSymbol, error) {
	slog.Debug("Getting document symbols", "uri", uri)

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}

	response, err := c.transport.SendRequest("textDocument/documentSymbol", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get document symbols: %w", err)
	}

	// LSP documentSymbol response can be null, DocumentSymbol[], or SymbolInformation[]
	var rawResponse json.RawMessage
	if err := json.Unmarshal(response, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document symbols response: %w", err)
	}

	// Handle null response
	if string(rawResponse) == "null" {
		slog.Debug("No document symbols found", "uri", uri)
		return []types.DocumentSymbol{}, nil
	}

	// Try to unmarshal as DocumentSymbol[] (hierarchical)
	var symbols []types.DocumentSymbol
	if err := json.Unmarshal(rawResponse, &symbols); err != nil {
		// If that fails, try to unmarshal as SymbolInformation[] (flat)
		var symbolInfos []types.SymbolInformation
		if err := json.Unmarshal(rawResponse, &symbolInfos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document symbols response: %w", err)
		}

		// Convert SymbolInformation to DocumentSymbol
		symbols = make([]types.DocumentSymbol, len(symbolInfos))
		for i, info := range symbolInfos {
			symbols[i] = types.DocumentSymbol{
				Name:           info.Name,
				Kind:           info.Kind,
				Range:          info.Location.Range,
				SelectionRange: info.Location.Range,
			}
		}
		slog.Debug("Found document symbols (flat format)", "count", len(symbols), "uri", uri)
	} else {
		slog.Debug("Found document symbols (hierarchical format)", "count", len(symbols), "uri", uri)
	}

	return symbols, nil
}
