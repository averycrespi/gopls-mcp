package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"

	"gopls-mcp/pkg/project"
	"gopls-mcp/pkg/types"
)

var _ types.LSPClient = &Client{}

const (
	defaultGoplsPath = "gopls"
	goplsStartDelay  = 100 * time.Millisecond
)

// Client implements the LSP client interface
type Client struct {
	goplsPath string
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	requestID int64
	responses map[int64]chan json.RawMessage
	mu        sync.RWMutex
	done      chan struct{}
}

// NewClient creates a new LSP client
func NewClient(goplsPath string) *Client {
	if goplsPath == "" {
		goplsPath = defaultGoplsPath
	}

	return &Client{
		goplsPath: goplsPath,
		responses: make(map[int64]chan json.RawMessage),
		done:      make(chan struct{}),
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
	c.stdin = stdin
	c.stdout = stdout
	c.stderr = stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start gopls: %w", err)
	}

	go c.readResponses()

	// Give gopls a moment to start
	time.Sleep(goplsStartDelay)

	return nil
}

// readResponses reads responses from gopls stdout
func (c *Client) readResponses() {
	defer close(c.done)

	for {
		// Read Content-Length header byte by byte until we find \r\n\r\n
		var contentLength int
		var header []byte

		for {
			b := make([]byte, 1)
			if _, err := c.stdout.Read(b); err != nil {
				return
			}
			header = append(header, b[0])

			if len(header) >= 4 && string(header[len(header)-4:]) == "\r\n\r\n" {
				headerStr := string(header)
				if _, err := fmt.Sscanf(headerStr, "Content-Length: %d\r\n\r\n", &contentLength); err != nil {
					continue
				}
				break
			}
		}

		// Read the JSON response body
		body := make([]byte, contentLength)
		if _, err := io.ReadFull(c.stdout, body); err != nil {
			return
		}

		c.handleResponse(body)
	}
}

type lspResponse struct {
	ID     json.RawMessage `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  json.RawMessage `json:"error"`
}

// handleResponse handles a JSON-RPC response
func (c *Client) handleResponse(content []byte) {
	var resp lspResponse
	if err := json.Unmarshal(content, &resp); err != nil {
		return
	}

	if resp.ID == nil {
		return // notification
	}

	var id int64
	if err := json.Unmarshal(resp.ID, &id); err != nil {
		return
	}

	c.mu.RLock()
	ch, ok := c.responses[id]
	c.mu.RUnlock()

	if ok {
		if resp.Error != nil {
			ch <- resp.Error
		} else {
			ch <- resp.Result
		}
	}
}

// sendRequest sends a JSON-RPC request
func (c *Client) sendRequest(method string, params interface{}) (json.RawMessage, error) {
	id := atomic.AddInt64(&c.requestID, 1)

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ch := make(chan json.RawMessage, 1)
	c.mu.Lock()
	c.responses[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.responses, id)
		c.mu.Unlock()
	}()

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	// Wait for response with timeout
	select {
	case response := <-ch:
		return response, nil
	case <-time.After(10 * time.Second):
		return nil, fmt.Errorf("timeout waiting for response to method %s", method)
	}
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

	_, err := c.sendRequest("initialize", params)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Send initialized notification using sendNotification helper
	return c.sendNotification("initialized", map[string]any{})
}

// sendNotification sends a JSON-RPC notification
func (c *Client) sendNotification(method string, params any) error {
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write notification header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write notification data: %w", err)
	}

	return nil
}

func (c *Client) GoToDefinition(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.sendRequest("textDocument/definition", params)
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

	response, err := c.sendRequest("textDocument/references", params)
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

	response, err := c.sendRequest("textDocument/hover", params)
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

	response, err := c.sendRequest("textDocument/completion", params)
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

	response, err := c.sendRequest("textDocument/formatting", params)
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

	response, err := c.sendRequest("textDocument/rename", params)
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
	_, err := c.sendRequest("shutdown", nil)
	if err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	// Send exit notification
	notification := map[string]any{
		"jsonrpc": "2.0",
		"method":  "exit",
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal exit notification: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write exit header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write exit data: %w", err)
	}

	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_ = c.cmd.Wait()
	}

	return nil
}
