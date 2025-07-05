package lsp

import (
	"testing"

	"gopls-mcp/pkg/types"
)

func TestNewClient(t *testing.T) {
	client := NewClient("")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	
	if client.responses == nil {
		t.Error("client.responses not initialized")
	}
	
	if client.done == nil {
		t.Error("client.done channel not initialized")
	}
}

func TestNewManager(t *testing.T) {
	manager := NewManager("gopls")
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	
	if manager.goplsPath != "gopls" {
		t.Errorf("Expected goplsPath to be 'gopls', got %s", manager.goplsPath)
	}
	
	if manager.initialized {
		t.Error("Manager should not be initialized by default")
	}
}

func TestManagerIsInitialized(t *testing.T) {
	manager := NewManager("gopls")
	
	if manager.IsInitialized() {
		t.Error("Manager should not be initialized by default")
	}
}

func TestClientImplementsInterface(t *testing.T) {
	client := NewClient("gopls")
	
	// Verify that Client implements the LSPClient interface
	var _ types.LSPClient = client
}