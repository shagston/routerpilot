package types

import "time"

type EventType string
type Severity string
type Priority int

const (
	SeverityDebug    Severity = "debug"
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

const (
	PriorityLow    Priority = 10
	PriorityNormal Priority = 50
	PriorityHigh   Priority = 90
	PriorityCritical Priority = 100
)

type Event struct {
	ID            EventID        `json:"id"`
	Timestamp     time.Time      `json:"timestamp"`
	AgentID       AgentID        `json:"agent_id,omitempty"`
	ExecutionID   ExecutionID    `json:"execution_id,omitempty"`
	TaskID        TaskID         `json:"task_id,omitempty"`
	ToolID        ToolID         `json:"tool_id,omitempty"`
	CorrelationID string         `json:"correlation_id,omitempty"`
	Type          EventType      `json:"type"`
	Source        string         `json:"source"`
	Severity      Severity       `json:"severity"`
	Priority      Priority       `json:"priority,omitempty"`
	TTL           time.Duration  `json:"ttl,omitempty"`
	Payload       map[string]any `json:"payload,omitempty"`
	Metadata      map[string]any `json:"metadata,omitempty"`
}

func (e Event) Expired() bool {
	if e.TTL <= 0 {
		return false
	}
	return time.Since(e.Timestamp) > e.TTL
}
