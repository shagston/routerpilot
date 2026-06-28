package types

import "time"

type Capability string

type ToolMetadata struct {
	ID               ToolID        `json:"id"`
	Version          string        `json:"version"`
	Category         string        `json:"category"`
	Description      string        `json:"description"`
	Permissions      []Permission  `json:"permissions,omitempty"`
	Capabilities     []Capability  `json:"capabilities,omitempty"`
	Timeout          time.Duration `json:"timeout"`
	Risk             RiskLevel     `json:"risk"`
	SupportsDryRun   bool          `json:"supports_dry_run"`
	SupportsRollback bool          `json:"supports_rollback"`
}

type ToolInput map[string]any
type ToolOutput map[string]any

type ToolResult struct {
	Success   bool       `json:"success"`
	Output    ToolOutput `json:"output,omitempty"`
	Error     string     `json:"error,omitempty"`
	Retryable bool       `json:"retryable,omitempty"`
}
