package types

// Config represents the configuration for the gopls-mcp server
type Config struct {
	GoplsPath     string `json:"gopls_path,omitempty"`
	WorkspaceRoot string `json:"workspace_root"`
	LogLevel      string `json:"log_level,omitempty"`
}
