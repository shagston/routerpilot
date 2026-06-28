package types

import "time"

type EventType string
type Severity string

const (
	SeverityDebug    Severity = "debug"
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

type Event struct {
	ID          EventID        `json:"id"`
	Timestamp   time.Time      `json:"timestamp"`
	ExecutionID ExecutionID    `json:"execution_id,omitempty"`
	TaskID      TaskID         `json:"task_id,omitempty"`
	ToolID      ToolID         `json:"tool_id,omitempty"`
	Type        EventType      `json:"type"`
	Source      string         `json:"source"`
	Severity    Severity       `json:"severity"`
	Payload     map[string]any `json:"payload,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
