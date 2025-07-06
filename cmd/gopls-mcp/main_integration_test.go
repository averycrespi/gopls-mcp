//go:build integration

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/averycrespi/gopls-mcp/internal/results"
	"github.com/stretchr/testify/assert"
)

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// MCPServerProcess manages the MCP server process for testing
type MCPServerProcess struct {
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	scanner *bufio.Scanner
}

// startMCPServer starts the MCP server process
func startMCPServer(t *testing.T, workspaceRoot string) *MCPServerProcess {
	cmd := exec.Command("go", "run", "main.go", "-workspace-root", workspaceRoot, "-log-level", "debug")

	stdin, err := cmd.StdinPipe()
	assert.NoError(t, err, "Failed to create stdin pipe")

	stdout, err := cmd.StdoutPipe()
	assert.NoError(t, err, "Failed to create stdout pipe")

	stderr, err := cmd.StderrPipe()
	assert.NoError(t, err, "Failed to create stderr pipe")

	err = cmd.Start()
	assert.NoError(t, err, "Failed to start MCP server")

	go func() {
		stderrScanner := bufio.NewScanner(stderr)
		for stderrScanner.Scan() {
			t.Logf("Server stderr: %s", stderrScanner.Text())
		}
	}()

	scanner := bufio.NewScanner(stdout)

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	return &MCPServerProcess{
		cmd:     cmd,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
		scanner: scanner,
	}
}

// stop terminates the MCP server process
func (s *MCPServerProcess) stop() error {
	s.stdin.Close()
	s.stdout.Close()
	s.stderr.Close()
	return s.cmd.Process.Kill()
}

// sendRequest sends a JSON-RPC request to the server
func (s *MCPServerProcess) sendRequest(t *testing.T, req MCPRequest) MCPResponse {
	reqJSON, err := json.Marshal(req)
	assert.NoError(t, err, "Failed to marshal request")

	_, err = s.stdin.Write(append(reqJSON, '\n'))
	assert.NoError(t, err, "Failed to write request")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan MCPResponse, 1)
	errChan := make(chan error, 1)

	go func() {
		if s.scanner.Scan() {
			line := s.scanner.Text()
			var resp MCPResponse
			if err := json.Unmarshal([]byte(line), &resp); err != nil {
				errChan <- fmt.Errorf("failed to unmarshal response: %v", err)
				return
			}
			done <- resp
		} else {
			if err := s.scanner.Err(); err != nil {
				errChan <- fmt.Errorf("scanner error: %v", err)
			} else {
				errChan <- fmt.Errorf("scanner returned false but no error")
			}
		}
	}()

	select {
	case resp := <-done:
		return resp
	case err := <-errChan:
		assert.Fail(t, "Error reading response", err.Error())
	case <-ctx.Done():
		assert.Fail(t, "Timeout waiting for response")
	}

	return MCPResponse{} // unreachable
}

// parseToolResult parses the JSON content from a tool result
func parseToolResult(t *testing.T, result map[string]any) string {
	content, ok := result["content"]
	assert.True(t, ok, "Expected content in tool result")

	// Handle both string and array format
	if contentStr, ok := content.(string); ok {
		return contentStr
	}

	// Handle array format (MCP content can be an array of content items)
	if contentArray, ok := content.([]interface{}); ok {
		assert.NotEmpty(t, contentArray, "Content array should not be empty")

		// Get first content item
		firstContent := contentArray[0]
		if contentMap, ok := firstContent.(map[string]interface{}); ok {
			if text, ok := contentMap["text"].(string); ok {
				return text
			}
		}
	}

	assert.Fail(t, "Unexpected content format", "Expected string or array, got %T", content)
	return ""
}

// validateSymbolDefinitionResult validates the structure of a symbol definition result
func validateSymbolDefinitionResult(t *testing.T, jsonContent string, expectedSymbol string) {
	var result results.SymbolDefinitionResult
	err := json.Unmarshal([]byte(jsonContent), &result)
	assert.NoError(t, err, "Should be able to unmarshal symbol definition result")

	// Validate basic structure
	assert.Equal(t, expectedSymbol, result.Query, "Query should match expected symbol")
	assert.Greater(t, result.Count, 0, "Should have found at least one symbol")
	assert.Len(t, result.Symbols, result.Count, "Symbol count should match actual symbols")

	// Validate first symbol
	assert.NotEmpty(t, result.Symbols, "Should have at least one symbol")
	firstSymbol := result.Symbols[0]
	assert.Equal(t, expectedSymbol, firstSymbol.Name, "First symbol name should match")
	assert.NotEmpty(t, firstSymbol.Kind, "Symbol kind should not be empty")
	assert.NotEmpty(t, firstSymbol.Location.File, "Symbol file should not be empty")
	assert.Greater(t, firstSymbol.Location.Line, 0, "Symbol line should be positive")
	assert.NotEmpty(t, firstSymbol.Definitions, "Should have at least one definition")

	// Validate first definition
	firstDef := firstSymbol.Definitions[0]
	assert.NotEmpty(t, firstDef.Location.File, "Definition file should not be empty")
	assert.Greater(t, firstDef.Location.Line, 0, "Definition line should be positive")
	if firstDef.Source != nil {
		assert.NotEmpty(t, firstDef.Source.Lines, "Source context should have lines")
	}
}

// validateSymbolSearchResult validates the structure of a symbol search result
func validateSymbolSearchResult(t *testing.T, jsonContent string, expectedSymbol string) {
	var result results.SymbolSearchResult
	err := json.Unmarshal([]byte(jsonContent), &result)
	assert.NoError(t, err, "Should be able to unmarshal symbol search result")

	// Validate basic structure
	assert.Equal(t, expectedSymbol, result.Query, "Query should match expected symbol")
	assert.GreaterOrEqual(t, result.Count, 0, "Count should be non-negative")
	assert.Len(t, result.Symbols, result.Count, "Symbol count should match actual symbols")

	if result.Count > 0 {
		// Validate first symbol
		firstSymbol := result.Symbols[0]
		assert.Contains(t, firstSymbol.Name, expectedSymbol, "First symbol name should contain expected symbol")
		assert.NotEmpty(t, firstSymbol.Kind, "Symbol kind should not be empty")
		assert.NotEmpty(t, firstSymbol.Location.File, "Symbol file should not be empty")
		assert.Greater(t, firstSymbol.Location.Line, 0, "Symbol line should be positive")
		if firstSymbol.Source != nil {
			assert.NotEmpty(t, firstSymbol.Source.Lines, "Source context should have lines")
		}
	}
}

// initialize sends the MCP initialize request
func (s *MCPServerProcess) initialize(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"clientInfo": map[string]any{
				"name":    "integration-test",
				"version": "1.0.0",
			},
		},
	}

	resp := s.sendRequest(t, req)
	assert.Nil(t, resp.Error, "MCP initialize should not return an error")
}

// TestMCPServerIntegration tests the MCP server integration using testdata/example
func TestMCPServerIntegration(t *testing.T) {
	// Use testdata/example as the workspace
	workspaceRoot, err := filepath.Abs("../../testdata/example")
	assert.NoError(t, err, "Failed to get testdata/example directory")

	_, err = os.Stat(filepath.Join(workspaceRoot, "go.mod"))
	assert.NoError(t, err, "testdata/example should be a Go module directory with go.mod")

	server := startMCPServer(t, workspaceRoot)
	defer server.stop()

	server.initialize(t)

	t.Run("ListTools", func(t *testing.T) {
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/list",
		}

		resp := server.sendRequest(t, req)
		assert.Nil(t, resp.Error, "List tools should not return an error")

		var result map[string]any
		err := json.Unmarshal(resp.Result, &result)
		assert.NoError(t, err, "Should be able to unmarshal tools list")

		tools, ok := result["tools"].([]any)
		assert.True(t, ok, "Expected tools array, got %T", result["tools"])

		expectedTools := []string{
			"symbol_definition",
			"find_references",
			"symbol_search",
		}

		assert.Len(t, tools, len(expectedTools), "Should have exactly %d tools", len(expectedTools))

		// Verify all expected tools are present
		foundTools := make(map[string]bool)
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]any)
			assert.True(t, ok, "Expected tool to be map, got %T", tool)
			if !ok {
				continue
			}

			name, ok := toolMap["name"].(string)
			assert.True(t, ok, "Expected tool name to be string, got %T", toolMap["name"])
			if ok {
				foundTools[name] = true
			}
		}

		// Check that all expected tools were found
		for _, expectedTool := range expectedTools {
			assert.True(t, foundTools[expectedTool], "Expected tool %s not found", expectedTool)
		}
	})

	t.Run("SymbolDefinition", func(t *testing.T) {
		// Test symbol definition by searching for "NewCalculator" symbol
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "symbol_definition",
				"arguments": map[string]any{
					"symbol": "NewCalculator",
				},
			},
		}

		resp := server.sendRequest(t, req)
		assert.Nil(t, resp.Error, "Symbol definition should not return an error")

		// Validate that we got a definition result
		var result map[string]any
		err := json.Unmarshal(resp.Result, &result)
		assert.NoError(t, err, "Should be able to unmarshal symbol definition result")

		// Parse and validate the JSON response structure
		contentStr := parseToolResult(t, result)
		validateSymbolDefinitionResult(t, contentStr, "NewCalculator")

		t.Logf("Symbol definition content: %v", contentStr)
	})

	t.Run("SymbolSearch", func(t *testing.T) {
		// Test searching for symbols by searching for "Calculator"
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      3.1,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "symbol_search",
				"arguments": map[string]any{
					"symbol": "Calculator",
				},
			},
		}

		resp := server.sendRequest(t, req)
		assert.Nil(t, resp.Error, "Symbol search should not return an error")

		// Validate that we got symbol search results
		var result map[string]any
		err := json.Unmarshal(resp.Result, &result)
		assert.NoError(t, err, "Should be able to unmarshal symbol search result")

		// Parse and validate the JSON response structure
		contentStr := parseToolResult(t, result)
		validateSymbolSearchResult(t, contentStr, "Calculator")

		t.Logf("Symbol search content: %v", contentStr)
	})

	t.Run("FindReferences", func(t *testing.T) {
		// Test find references on Calculator type in calculator.go
		// calculator.go line 6: type Calculator struct {
		//                            ^
		//                        char 5 (0-based)
		calcFile := filepath.Join(workspaceRoot, "calculator.go")

		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      5,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "find_references",
				"arguments": map[string]any{
					"file_path": calcFile,
					"line":      5, // Zero-based: line 6 in editor
					"character": 5, // Zero-based: position of "Calculator"
				},
			},
		}

		resp := server.sendRequest(t, req)
		assert.Nil(t, resp.Error, "Find references should not return an error")

		// Validate that we got find references content
		var result map[string]any
		err := json.Unmarshal(resp.Result, &result)
		assert.NoError(t, err, "Should be able to unmarshal find references result")

		content, ok := result["content"]
		assert.True(t, ok, "Expected content in find references result")

		// Should contain reference information
		contentStr := fmt.Sprintf("%v", content)
		assert.Contains(t, contentStr, "reference", "Response should contain reference information")
		assert.Contains(t, contentStr, "calculator.go", "Response should reference calculator.go file")
		t.Logf("Find references content: %v", content)
	})

}
