package types

import "time"

type ExecutionState string

const (
	ExecutionNew        ExecutionState = "new"
	ExecutionPlanning   ExecutionState = "planning"
	ExecutionValidating ExecutionState = "validating"
	ExecutionReady      ExecutionState = "ready"
	ExecutionRunning    ExecutionState = "running"
	ExecutionVerifying  ExecutionState = "verifying"
	ExecutionCompleted  ExecutionState = "completed"
	ExecutionFailed     ExecutionState = "failed"
	ExecutionCancelled  ExecutionState = "cancelled"
	ExecutionTimeout    ExecutionState = "timeout"
	ExecutionRolledBack ExecutionState = "rolled_back"
)

type ContextSnapshot map[string]any

type Execution struct {
	ID         ExecutionID     `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	StartedAt  *time.Time      `json:"started_at,omitempty"`
	FinishedAt *time.Time      `json:"finished_at,omitempty"`
	State      ExecutionState  `json:"state"`
	Plan       Plan            `json:"plan"`
	Context    ContextSnapshot `json:"context,omitempty"`
	Result     *ToolResult     `json:"result,omitempty"`
}
