package types

import "context"

// Server defines the MCP server interface
type Server interface {
	Start(ctx context.Context) error
	RegisterTools() error
	Shutdown(ctx context.Context) error
}
