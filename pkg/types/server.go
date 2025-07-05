package types

import "context"

// Server defines the MCP server interface
type Server interface {
	Serve(ctx context.Context) error
}
