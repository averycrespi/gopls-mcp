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

// validateFindSymbolDefinitionsByNameResult validates the structure of a find symbol definitions by name result
func validateFindSymbolDefinitionsByNameResult(t *testing.T, jsonContent string, expectedSymbol string) {
	var result results.FindSymbolDefinitionsByNameToolResult
	err := json.Unmarshal([]byte(jsonContent), &result)
	assert.NoError(t, err, "Should be able to unmarshal find symbol definitions by name tool result")

	// Validate basic structure
	assert.Equal(t, expectedSymbol, result.SymbolName, "Searched symbol name should match")
	assert.NotEmpty(t, result.Message, "Message should not be empty")
	assert.Greater(t, len(result.Definitions), 0, "Should have found at least one symbol")

	// Validate first symbol
	firstSymbol := result.Definitions[0]
	assert.NotEmpty(t, firstSymbol.Name, "Symbol name should not be empty")
	assert.NotEmpty(t, firstSymbol.Kind, "Symbol kind should not be empty")
	assert.NotEmpty(t, firstSymbol.Location.File, "Symbol file should not be empty")
	assert.Greater(t, firstSymbol.Location.Line, 0, "Symbol line should be positive")
	assert.NotEmpty(t, firstSymbol.Anchor, "Symbol anchor should not be empty")
	assert.True(t, firstSymbol.Anchor.IsValid(), "Symbol anchor should be valid")
}

// validateSymbolReferenceResult validates the structure of a symbol reference result
func validateSymbolReferenceResult(t *testing.T, jsonContent string, expectedAnchor string) {
	var result results.SymbolReferenceResult
	err := json.Unmarshal([]byte(jsonContent), &result)
	assert.NoError(t, err, "Should be able to unmarshal symbol reference result")

	// Validate basic structure
	assert.NotEmpty(t, result.Name, "Symbol name should not be empty")
	assert.NotEmpty(t, result.Kind, "Symbol kind should not be empty")
	assert.NotEmpty(t, result.Location.File, "Symbol file should not be empty")
	assert.Greater(t, result.Location.Line, 0, "Symbol line should be positive")
	assert.NotEmpty(t, result.Anchor, "Symbol anchor should not be empty")
	assert.True(t, result.Anchor.IsValid(), "Symbol anchor should be valid")
	assert.NotNil(t, result.References, "References should not be nil")

	// Validate that the anchor matches expected format
	if expectedAnchor != "" {
		assert.Equal(t, expectedAnchor, result.Anchor.String(), "Anchor should match expected value")
	}
}

// validateListSymbolsInFileToolResult validates the structure of a list symbols in file result
func validateListSymbolsInFileToolResult(t *testing.T, jsonContent string) {
	var result results.ListSymbolsInFileToolResult
	err := json.Unmarshal([]byte(jsonContent), &result)
	assert.NoError(t, err, "Should be able to unmarshal list symbols in file tool result")

	// Validate basic structure
	assert.NotEmpty(t, result.FilePath, "File path should not be empty")
	assert.NotEmpty(t, result.Message, "Message should not be empty")
	assert.Greater(t, len(result.FileSymbols), 0, "Should have found at least one symbol")

	// Validate first symbol
	firstSymbol := result.FileSymbols[0]
	assert.NotEmpty(t, firstSymbol.Name, "Symbol name should not be empty")
	assert.NotEmpty(t, firstSymbol.Kind, "Symbol kind should not be empty")
	assert.NotEmpty(t, firstSymbol.Location.File, "Symbol file should not be empty")
	assert.Greater(t, firstSymbol.Location.Line, 0, "Symbol line should be positive")
	assert.NotEmpty(t, firstSymbol.Anchor, "Symbol anchor should not be empty")
	assert.True(t, firstSymbol.Anchor.IsValid(), "Symbol anchor should be valid")

	// Look for a struct symbol to verify hierarchical structure
	var structSymbol *results.FileSymbol
	for _, symbol := range result.FileSymbols {
		if symbol.Kind == "struct" {
			structSymbol = &symbol
			break
		}
	}

	// If we found a struct symbol, it should potentially have children (fields)
	if structSymbol != nil {
		t.Logf("Found struct symbol '%s' with %d children", structSymbol.Name, len(structSymbol.Children))
		// Note: Children may be empty if the struct has no fields, which is valid
		if len(structSymbol.Children) > 0 {
			// Validate first child if present
			firstChild := structSymbol.Children[0]
			assert.NotEmpty(t, firstChild.Name, "Child symbol name should not be empty")
			assert.NotEmpty(t, firstChild.Kind, "Child symbol kind should not be empty")
			assert.Greater(t, firstChild.Location.Line, 0, "Child symbol line should be positive")
			assert.NotEmpty(t, firstChild.Anchor, "Child symbol anchor should not be empty")
			assert.True(t, firstChild.Anchor.IsValid(), "Child symbol anchor should be valid")
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
			"find_symbol_definitions_by_name",
			"find_symbol_references_by_anchor",
			"list_symbols_in_file",
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

	t.Run("FindSymbolDefinitionsByName", func(t *testing.T) {
		// Test symbol definition by searching for "NewCalculator" symbol
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "find_symbol_definitions_by_name",
				"arguments": map[string]any{
					"symbol_name": "NewCalculator",
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
		validateFindSymbolDefinitionsByNameResult(t, contentStr, "NewCalculator")

		t.Logf("Symbol definition content: %v", contentStr)
	})

	t.Run("FindSymbolReferencesByAnchor", func(t *testing.T) {
		// Test find symbol references by anchor using Calculator struct anchor
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      5,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "find_symbol_references_by_anchor",
				"arguments": map[string]any{
					"anchor": "anchor://calculator.go#6:6", // Calculator struct definition (1-indexed)
				},
			},
		}

		resp := server.sendRequest(t, req)
		assert.Nil(t, resp.Error, "Find symbol references by anchor should not return an error")

		// Validate that we got a references result
		var result map[string]any
		err := json.Unmarshal(resp.Result, &result)
		assert.NoError(t, err, "Should be able to unmarshal symbol references result")

		// Parse and validate the JSON response structure
		contentStr := parseToolResult(t, result)
		validateSymbolReferenceResult(t, contentStr, "anchor://calculator.go#6:6")

		t.Logf("Find symbol references by anchor content: %v", contentStr)
	})

	t.Run("FileSymbols", func(t *testing.T) {
		// Test file symbols by analyzing calculator.go file
		calcFile := filepath.Join(workspaceRoot, "calculator.go")

		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      6,
			Method:  "tools/call",
			Params: map[string]any{
				"name": "list_symbols_in_file",
				"arguments": map[string]any{
					"file_path": calcFile,
				},
			},
		}

		resp := server.sendRequest(t, req)
		assert.Nil(t, resp.Error, "File symbols should not return an error")

		// Validate that we got file symbols result
		var result map[string]any
		err := json.Unmarshal(resp.Result, &result)
		assert.NoError(t, err, "Should be able to unmarshal file symbols result")

		// Parse and validate the JSON response structure
		contentStr := parseToolResult(t, result)
		validateListSymbolsInFileToolResult(t, contentStr)

		t.Logf("File symbols content: %v", contentStr)
	})

}
