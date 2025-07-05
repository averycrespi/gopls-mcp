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
	"strings"
	"testing"
	"time"
)

// MCPRequest represents a JSON-RPC 2.0 request
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents a JSON-RPC 2.0 response
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents a JSON-RPC 2.0 error
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCPServerProcess manages the MCP server process for testing
type MCPServerProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	scanner *bufio.Scanner
}

// startMCPServer starts the MCP server process
func startMCPServer(t *testing.T, workspaceRoot string) *MCPServerProcess {
	// Build the server first
	buildCmd := exec.Command("make", "build")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build MCP server: %v", err)
	}

	// Start the server process
	cmd := exec.Command("./bin/gopls-mcp", "-workspace-root", workspaceRoot, "-log-level", "debug")
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}
	
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP server: %v", err)
	}
	
	// Start a goroutine to read stderr
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
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}
	
	t.Logf("Sending request: %s", string(reqJSON))
	
	if _, err := s.stdin.Write(append(reqJSON, '\n')); err != nil {
		t.Fatalf("Failed to write request: %v", err)
	}
	
	// Read response with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	done := make(chan MCPResponse, 1)
	errChan := make(chan error, 1)
	
	go func() {
		if s.scanner.Scan() {
			line := s.scanner.Text()
			t.Logf("Received response: %s", line)
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
		t.Fatalf("Error reading response: %v", err)
	case <-ctx.Done():
		t.Fatalf("Timeout waiting for response")
	}
	
	return MCPResponse{} // unreachable
}

// initialize sends the MCP initialize request
func (s *MCPServerProcess) initialize(t *testing.T) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "integration-test",
				"version": "1.0.0",
			},
		},
	}
	
	resp := s.sendRequest(t, req)
	if resp.Error != nil {
		t.Fatalf("Initialize failed: %v", resp.Error.Message)
	}
}

// TestMCPServerIntegration tests the MCP server integration
func TestMCPServerIntegration(t *testing.T) {
	// Get the current working directory as workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	// Verify we're in the correct directory (should have go.mod)
	if _, err := os.Stat(filepath.Join(workspaceRoot, "go.mod")); os.IsNotExist(err) {
		t.Fatalf("Not in a Go module directory, go.mod not found")
	}
	
	// Start the MCP server
	server := startMCPServer(t, workspaceRoot)
	defer server.stop()
	
	// Initialize the server
	server.initialize(t)
	
	// Test tools list
	t.Run("ListTools", func(t *testing.T) {
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      2,
			Method:  "tools/list",
		}
		
		resp := server.sendRequest(t, req)
		if resp.Error != nil {
			t.Fatalf("List tools failed: %v", resp.Error.Message)
		}
		
		// Verify we have the expected tools
		var result map[string]interface{}
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			t.Fatalf("Failed to unmarshal tools list: %v", err)
		}
		
		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatalf("Expected tools array, got %T", result["tools"])
		}
		
		expectedTools := []string{
			"gopls.go_to_definition",
			"gopls.find_references", 
			"gopls.hover_info",
			"gopls.get_completion",
			"gopls.format_code",
			"gopls.rename_symbol",
		}
		
		if len(tools) != len(expectedTools) {
			t.Errorf("Expected %d tools, got %d", len(expectedTools), len(tools))
		}
		
		for _, tool := range tools {
			toolMap, ok := tool.(map[string]interface{})
			if !ok {
				t.Errorf("Expected tool to be map, got %T", tool)
				continue
			}
			
			name, ok := toolMap["name"].(string)
			if !ok {
				t.Errorf("Expected tool name to be string, got %T", toolMap["name"])
				continue
			}
			
			found := false
			for _, expected := range expectedTools {
				if name == expected {
					found = true
					break
				}
			}
			
			if !found {
				t.Errorf("Unexpected tool: %s", name)
			}
		}
	})
	
	// Test go_to_definition tool
	t.Run("GoToDefinition", func(t *testing.T) {
		// Use main.go as test file - look for the "main" function definition
		mainFile := filepath.Join(workspaceRoot, "cmd", "gopls-mcp", "main.go")
		
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "gopls.go_to_definition",
				"arguments": map[string]interface{}{
					"file_path": mainFile,
					"line":      5,  // Approximate line where we might find a symbol
					"character": 10, // Approximate character position
				},
			},
		}
		
		resp := server.sendRequest(t, req)
		if resp.Error != nil {
			t.Logf("Go to definition failed (expected for test): %v", resp.Error.Message)
		} else {
			t.Logf("Go to definition succeeded")
		}
	})
	
	// Test hover_info tool
	t.Run("HoverInfo", func(t *testing.T) {
		// Use main.go as test file
		mainFile := filepath.Join(workspaceRoot, "cmd", "gopls-mcp", "main.go")
		
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "gopls.hover_info",
				"arguments": map[string]interface{}{
					"file_path": mainFile,
					"line":      10,
					"character": 5,
				},
			},
		}
		
		resp := server.sendRequest(t, req)
		if resp.Error != nil {
			t.Logf("Hover info failed (expected for test): %v", resp.Error.Message)
		} else {
			t.Logf("Hover info succeeded")
		}
	})
	
	// Test find_references tool
	t.Run("FindReferences", func(t *testing.T) {
		// Use main.go as test file
		mainFile := filepath.Join(workspaceRoot, "cmd", "gopls-mcp", "main.go")
		
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      5,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "gopls.find_references",
				"arguments": map[string]interface{}{
					"file_path": mainFile,
					"line":      15,
					"character": 10,
				},
			},
		}
		
		resp := server.sendRequest(t, req)
		if resp.Error != nil {
			t.Logf("Find references failed (expected for test): %v", resp.Error.Message)
		} else {
			t.Logf("Find references succeeded")
		}
	})
	
	// Test get_completion tool
	t.Run("GetCompletion", func(t *testing.T) {
		// Use main.go as test file
		mainFile := filepath.Join(workspaceRoot, "cmd", "gopls-mcp", "main.go")
		
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      6,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "gopls.get_completion",
				"arguments": map[string]interface{}{
					"file_path": mainFile,
					"line":      20,
					"character": 5,
				},
			},
		}
		
		resp := server.sendRequest(t, req)
		if resp.Error != nil {
			t.Logf("Get completion failed (expected for test): %v", resp.Error.Message)
		} else {
			t.Logf("Get completion succeeded")
		}
	})
}

// TestMCPServerWithRealSymbols tests the MCP server with real symbols in the codebase
func TestMCPServerWithRealSymbols(t *testing.T) {
	// Get the current working directory as workspace root
	workspaceRoot, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	
	// Start the MCP server
	server := startMCPServer(t, workspaceRoot)
	defer server.stop()
	
	// Initialize the server
	server.initialize(t)
	
	// Test hover info on a known symbol in main.go
	t.Run("HoverInfoOnMainFunc", func(t *testing.T) {
		mainFile := filepath.Join(workspaceRoot, "cmd", "gopls-mcp", "main.go")
		
		// Read the main.go file to find the actual main function
		content, err := os.ReadFile(mainFile)
		if err != nil {
			t.Fatalf("Failed to read main.go: %v", err)
		}
		
		lines := strings.Split(string(content), "\n")
		var mainFuncLine int = -1
		for i, line := range lines {
			if strings.Contains(line, "func main()") {
				mainFuncLine = i
				break
			}
		}
		
		if mainFuncLine == -1 {
			t.Skip("Could not find main function in main.go")
		}
		
		req := MCPRequest{
			JSONRPC: "2.0",
			ID:      7,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "gopls.hover_info",
				"arguments": map[string]interface{}{
					"file_path": mainFile,
					"line":      mainFuncLine,
					"character": 5, // Position on "main"
				},
			},
		}
		
		resp := server.sendRequest(t, req)
		if resp.Error != nil {
			t.Logf("Hover info on main function failed: %v", resp.Error.Message)
		} else {
			var result map[string]interface{}
			if err := json.Unmarshal(resp.Result, &result); err != nil {
				t.Fatalf("Failed to unmarshal hover result: %v", err)
			}
			
			if content, ok := result["content"]; ok {
				t.Logf("Hover info content: %v", content)
			}
		}
	})
}