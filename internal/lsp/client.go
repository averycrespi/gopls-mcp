package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"

	"github.com/averycrespi/gopls-mcp/pkg/project"
	"github.com/averycrespi/gopls-mcp/pkg/types"
)

const (
	defaultGoplsPath = "gopls"
	goplsStartDelay  = 100 * time.Millisecond
)

var _ types.LSPClient = &Client{}

// Client implements the LSP client interface
type Client struct {
	goplsPath string
	cmd       *exec.Cmd
	stderr    io.ReadCloser
	transport types.Transport
}

// NewClient creates a new LSP client
func NewClient(goplsPath string) *Client {
	if goplsPath == "" {
		goplsPath = defaultGoplsPath
	}

	return &Client{
		goplsPath: goplsPath,
	}
}

// Start starts the gopls process
func (c *Client) Start(ctx context.Context, goplsPath string) error {
	cmd := exec.CommandContext(ctx, goplsPath, "serve")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	c.cmd = cmd
	c.stderr = stderr
	c.transport = NewJsonRpcTransport(stdin, stdout)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gopls: %w", err)
	}

	c.transport.Listen()

	return nil
}

// Initialize initializes the LSP client
func (c *Client) Initialize(ctx context.Context, rootURI string) error {
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
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Send initialized notification
	return c.transport.SendNotification("initialized", map[string]any{})
}

func (c *Client) GoToDefinition(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
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

func (c *Client) FindReferences(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
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

	var locations []types.Location
	if err := json.Unmarshal(response, &locations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal references response: %w", err)
	}

	return locations, nil
}

func (c *Client) Hover(ctx context.Context, uri string, position types.Position) (string, error) {
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

func (c *Client) GetDiagnostics(ctx context.Context, uri string) ([]types.Diagnostic, error) {
	// Note: Diagnostics are typically sent as notifications, not requests
	// This is a simplified implementation
	return []types.Diagnostic{}, nil
}

func (c *Client) GetCompletion(ctx context.Context, uri string, position types.Position) ([]types.CompletionItem, error) {
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

func (c *Client) FormatDocument(ctx context.Context, uri string) ([]json.RawMessage, error) {
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

func (c *Client) RenameSymbol(ctx context.Context, uri string, position types.Position, newName string) (map[string][]json.RawMessage, error) {
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

func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.transport.SendRequest("shutdown", nil)
	if err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	// Send exit notification
	if err := c.transport.SendNotification("exit", nil); err != nil {
		return fmt.Errorf("failed to send exit notification: %w", err)
	}

	// Close the transport
	c.transport.Close()

	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	return nil
}
