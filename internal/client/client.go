package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	return &GoplsClient{
		goplsPath: goplsPath,
	}
}

// Start starts the Gopls client
func (c *GoplsClient) Start(ctx context.Context, workspaceRoot string) error {
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

	if err := c.transport.Start(); err != nil {
		return fmt.Errorf("failed to start transport: %w", err)
	}

	rootURI := "file://" + workspaceRoot
	if err := c.initialize(rootURI); err != nil {
		return fmt.Errorf("failed to initialize Gopls client: %w", err)
	}

	return nil
}

func (c *GoplsClient) initialize(rootURI string) error {
	params := map[string]any{
		"processId": nil,
		"clientInfo": map[string]any{
			"name":    project.Name,
			"version": project.Version,
		},
		"rootUri":      rootURI,
		"capabilities": map[string]any{},
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

	return locations, nil
}

func (c *GoplsClient) FindReferences(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
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
		return []types.Location{}, nil
	}

	var locations []types.Location
	if err := json.Unmarshal(rawResponse, &locations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal references response: %w", err)
	}

	return locations, nil
}

func (c *GoplsClient) Hover(ctx context.Context, uri string, position types.Position) (string, error) {
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

func (c *GoplsClient) GetCompletion(ctx context.Context, uri string, position types.Position) ([]types.CompletionItem, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.transport.SendRequest("textDocument/completion", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get completion: %w", err)
	}

	var completion struct {
		Items []types.CompletionItem `json:"items"`
	}
	if err := json.Unmarshal(response, &completion); err != nil {
		return nil, fmt.Errorf("failed to unmarshal completion response: %w", err)
	}

	return completion.Items, nil
}

func (c *GoplsClient) FuzzyFindSymbol(ctx context.Context, query string) ([]types.SymbolInformation, error) {
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
		return []types.SymbolInformation{}, nil
	}

	var symbols []types.SymbolInformation
	if err := json.Unmarshal(rawResponse, &symbols); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspace symbol response: %w", err)
	}

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

func (c *GoplsClient) RenameSymbol(ctx context.Context, uri string, position types.Position, newName string) (map[string][]json.RawMessage, error) {
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

	var workspaceEdit struct {
		Changes map[string][]json.RawMessage `json:"changes"`
	}
	if err := json.Unmarshal(response, &workspaceEdit); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rename response: %w", err)
	}

	return workspaceEdit.Changes, nil
}
