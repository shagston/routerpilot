package types

import "time"

type RetryPolicy struct {
	Attempts int           `json:"attempts"`
	Delay    time.Duration `json:"delay"`
	Strategy string        `json:"strategy,omitempty"`
}

type TaskPurpose string

const (
	TaskPurposeAction  TaskPurpose = "action"
	TaskPurposeContext TaskPurpose = "context"
)

type Task struct {
	ID           TaskID        `json:"id"`
	Tool         ToolID        `json:"tool"`
	Purpose      TaskPurpose   `json:"purpose,omitempty"`
	Arguments    ToolInput     `json:"arguments,omitempty"`
	Dependencies []TaskID      `json:"dependencies,omitempty"`
	Timeout      time.Duration `json:"timeout,omitempty"`
	Retry        RetryPolicy   `json:"retry,omitempty"`
	Rollback     []Task        `json:"rollback,omitempty"`
}

type Plan struct {
	ID           PlanID         `json:"plan_id"`
	Intent       string         `json:"intent"`
	Steps        []Task         `json:"steps"`
	Rollback     []Task         `json:"rollback,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	Requirements []Capability   `json:"requirements,omitempty"`
	Risk         RiskLevel      `json:"risk"`
	DryRun       bool           `json:"dry_run,omitempty"`
}

func HasDependencyCycle(tasks []Task) bool {
	graph := make(map[TaskID][]TaskID, len(tasks))
	for _, task := range tasks {
		graph[task.ID] = task.Dependencies
	}

	visiting := map[TaskID]bool{}
	visited := map[TaskID]bool{}

	var visit func(TaskID) bool
	visit = func(id TaskID) bool {
		if visiting[id] {
			return true
		}
		if visited[id] {
			return false
		}
		visiting[id] = true
		for _, dependency := range graph[id] {
			if visit(dependency) {
				return true
			}
		}
		visiting[id] = false
		visited[id] = true
		return false
	}

	for _, task := range tasks {
		if visit(task.ID) {
			return true
		}
	}
	return false
}
