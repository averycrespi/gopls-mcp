package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"gopls-mcp/pkg/types"
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
		goplsPath = "gopls"
	}

	return &Client{
		goplsPath: goplsPath,
		responses: make(map[int64]chan json.RawMessage),
		done:      make(chan struct{}),
	}
}

// Start starts the gopls process
func (c *Client) Start(ctx context.Context, goplsPath string) error {
	cmd := exec.CommandContext(ctx, goplsPath, "-mode=stdio")

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

	return nil
}

// readResponses reads responses from gopls stdout
func (c *Client) readResponses() {
	defer close(c.done)

	scanner := bufio.NewScanner(c.stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "Content-Length:") {
			continue
		}

		lengthStr := strings.TrimSpace(strings.TrimPrefix(line, "Content-Length:"))
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			continue
		}

		// Read empty line
		if !scanner.Scan() {
			break
		}

		// Read content
		content := make([]byte, length)
		if _, err := io.ReadFull(c.stdout, content); err != nil {
			break
		}

		c.handleResponse(content)
	}
}

// handleResponse handles a JSON-RPC response
func (c *Client) handleResponse(content []byte) {
	var response struct {
		ID     json.RawMessage `json:"id"`
		Result json.RawMessage `json:"result"`
		Error  json.RawMessage `json:"error"`
	}

	if err := json.Unmarshal(content, &response); err != nil {
		return
	}

	if response.ID == nil {
		return // notification
	}

	var id int64
	if err := json.Unmarshal(response.ID, &id); err != nil {
		return
	}

	c.mu.RLock()
	ch, ok := c.responses[id]
	c.mu.RUnlock()

	if ok {
		if response.Error != nil {
			ch <- response.Error
		} else {
			ch <- response.Result
		}
	}
}

// sendRequest sends a JSON-RPC request
func (c *Client) sendRequest(method string, params interface{}) (json.RawMessage, error) {
	id := atomic.AddInt64(&c.requestID, 1)

	request := map[string]interface{}{
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

	response := <-ch
	return response, nil
}

// Initialize initializes the LSP client
func (c *Client) Initialize(ctx context.Context, rootURI string) error {
	params := map[string]interface{}{
		"processId": nil,
		"rootUri":   rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"definition": map[string]interface{}{
					"linkSupport": true,
				},
				"references": map[string]interface{}{},
				"hover":      map[string]interface{}{},
				"completion": map[string]interface{}{},
				"formatting": map[string]interface{}{},
				"rename":     map[string]interface{}{},
			},
		},
	}

	_, err := c.sendRequest("initialize", params)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Send initialized notification
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]interface{}{},
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return fmt.Errorf("failed to write initialized header: %w", err)
	}

	if _, err := c.stdin.Write(data); err != nil {
		return fmt.Errorf("failed to write initialized data: %w", err)
	}

	return nil
}

// GoToDefinition implements types.LSPClient
func (c *Client) GoToDefinition(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.sendRequest("textDocument/definition", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	var locations []types.Location
	if err := json.Unmarshal(response, &locations); err != nil {
		return nil, fmt.Errorf("failed to unmarshal definition response: %w", err)
	}

	return locations, nil
}

// FindReferences implements types.LSPClient
func (c *Client) FindReferences(ctx context.Context, uri string, position types.Position) ([]types.Location, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": position,
		"context": map[string]interface{}{
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

// Hover implements types.LSPClient
func (c *Client) Hover(ctx context.Context, uri string, position types.Position) (string, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"position": position,
	}

	response, err := c.sendRequest("textDocument/hover", params)
	if err != nil {
		return "", fmt.Errorf("failed to get hover: %w", err)
	}

	var hover struct {
		Contents interface{} `json:"contents"`
	}
	if err := json.Unmarshal(response, &hover); err != nil {
		return "", fmt.Errorf("failed to unmarshal hover response: %w", err)
	}

	// Handle different content formats
	switch v := hover.Contents.(type) {
	case string:
		return v, nil
	case map[string]interface{}:
		if value, ok := v["value"]; ok {
			return fmt.Sprintf("%v", value), nil
		}
	}

	return fmt.Sprintf("%v", hover.Contents), nil
}

// GetDiagnostics implements types.LSPClient
func (c *Client) GetDiagnostics(ctx context.Context, uri string) ([]types.Diagnostic, error) {
	// Note: Diagnostics are typically sent as notifications, not requests
	// This is a simplified implementation
	return []types.Diagnostic{}, nil
}

// GetCompletion implements types.LSPClient
func (c *Client) GetCompletion(ctx context.Context, uri string, position types.Position) ([]types.CompletionItem, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
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

// FormatDocument implements types.LSPClient
func (c *Client) FormatDocument(ctx context.Context, uri string) ([]json.RawMessage, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri": uri,
		},
		"options": map[string]interface{}{
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

// RenameSymbol implements types.LSPClient
func (c *Client) RenameSymbol(ctx context.Context, uri string, position types.Position, newName string) (map[string][]json.RawMessage, error) {
	params := map[string]interface{}{
		"textDocument": map[string]interface{}{
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

// Shutdown implements types.LSPClient
func (c *Client) Shutdown(ctx context.Context) error {
	_, err := c.sendRequest("shutdown", nil)
	if err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}

	// Send exit notification
	notification := map[string]interface{}{
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
