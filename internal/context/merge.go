package context

import (
	"fmt"

	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/types"
)

func CloneSnapshot(snapshot types.ContextSnapshot) types.ContextSnapshot {
	if snapshot == nil {
		return types.ContextSnapshot{}
	}
	cloned := make(types.ContextSnapshot, len(snapshot))
	for key, value := range snapshot {
		cloned[key] = value
	}
	return cloned
}

func MergeToolResult(snapshot types.ContextSnapshot, task types.Task, result types.ToolResult) {
	if snapshot == nil {
		return
	}
	snapshot[fmt.Sprintf("task.%s", task.ID)] = result
	if task.Tool != "" {
		snapshot[string(task.Tool)] = result
	}
}

func ValidateContextTask(reg *registry.ToolRegistry, task types.Task) error {
	t, err := reg.Get(task.Tool)
	if err != nil {
		return err
	}

	metadata := t.Metadata()
	if metadata.Risk != types.RiskLow {
		return fmt.Errorf("%w: context task %s must use a low-risk tool", types.ErrInvalidInput, task.ID)
	}
	for _, permission := range metadata.Permissions {
		if permission != types.PermissionRead {
			return fmt.Errorf("%w: context task %s must be read-only", types.ErrInvalidInput, task.ID)
		}
	}
	return nil
}
