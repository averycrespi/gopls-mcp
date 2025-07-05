package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/averycrespi/gopls-mcp/pkg/types"
)

// Manager manages the LSP client lifecycle
type Manager struct {
	client      *GoplsClient
	goplsPath   string
	initialized bool
	mu          sync.RWMutex
}

// NewManager creates a new LSP manager
func NewManager(goplsPath string) *Manager {
	return &Manager{
		goplsPath: goplsPath,
	}
}

// Initialize initializes the LSP client with the given workspace root
func (m *Manager) Initialize(ctx context.Context, workspaceRoot string) error {
	log.Printf("Initializing LSP manager with workspace: %s", workspaceRoot)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	m.client = NewGoplsClient(m.goplsPath)

	if err := m.client.Start(ctx, m.goplsPath); err != nil {
		return fmt.Errorf("failed to start LSP client: %w", err)
	}

	rootURI := "file://" + workspaceRoot
	if err := m.client.Initialize(ctx, rootURI); err != nil {
		return fmt.Errorf("failed to initialize LSP client: %w", err)
	}

	m.initialized = true
	return nil
}

// GetClient returns the LSP client
func (m *Manager) GetClient() types.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.initialized {
		return nil
	}

	return m.client
}

// Shutdown shuts down the LSP client
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return nil
	}

	if err := m.client.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown LSP client: %w", err)
	}

	m.initialized = false
	m.client = nil

	return nil
}

// IsInitialized returns whether the LSP client is initialized
func (m *Manager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.initialized
}
