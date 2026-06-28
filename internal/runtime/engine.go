package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	ctxengine "github.com/shagston/routerpilot/internal/context"
	"github.com/shagston/routerpilot/sdk/events"
	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
)

type Engine struct {
	registry  tool.Registry
	publisher events.Publisher
	validator Validator
	eventSeq  uint64
	now       func() time.Time
}

type Option func(*Engine)

type Validator interface {
	Validate(context.Context, types.Plan) error
}

func NewEngine(registry tool.Registry, publisher events.Publisher, options ...Option) *Engine {
	engine := &Engine{
		registry:  registry,
		publisher: publisher,
		now:       time.Now,
	}
	for _, option := range options {
		option(engine)
	}
	return engine
}

func WithClock(now func() time.Time) Option {
	return func(engine *Engine) {
		engine.now = now
	}
}

func WithValidator(validator Validator) Option {
	return func(engine *Engine) {
		engine.validator = validator
	}
}

func (e *Engine) Execute(ctx context.Context, plan types.Plan, snapshot types.ContextSnapshot) (types.Execution, error) {
	execution := types.Execution{
		ID:        types.ExecutionID(fmt.Sprintf("exec-%d", e.now().UnixNano())),
		CreatedAt: e.now(),
		State:     types.ExecutionNew,
		Plan:      plan,
		Context:   snapshot,
	}
	e.publish(execution, "", "", "execution.created", types.SeverityInfo, nil)

	if err := e.validatePlan(ctx, plan); err != nil {
		execution.State = types.ExecutionFailed
		finishedAt := e.now()
		execution.FinishedAt = &finishedAt
		e.publish(execution, "", "", "execution.failed", types.SeverityError, map[string]any{"error": err.Error()})
		return execution, err
	}
	if e.validator != nil {
		e.publish(execution, "", "", "safety.validation.started", types.SeverityInfo, nil)
		if err := e.validator.Validate(ctx, plan); err != nil {
			execution.State = types.ExecutionFailed
			finishedAt := e.now()
			execution.FinishedAt = &finishedAt
			e.publish(execution, "", "", "safety.validation.failed", types.SeverityError, map[string]any{"error": err.Error()})
			e.publish(execution, "", "", "execution.failed", types.SeverityError, map[string]any{"error": err.Error()})
			return execution, err
		}
		e.publish(execution, "", "", "execution.approved", types.SeverityInfo, nil)
	}

	startedAt := e.now()
	execution.StartedAt = &startedAt
	execution.State = types.ExecutionRunning
	e.publish(execution, "", "", "execution.started", types.SeverityInfo, nil)

	liveContext := ctxengine.CloneSnapshot(snapshot)
	execution.Context = liveContext

	completed := make(map[types.TaskID]bool, len(plan.Steps))
	for len(completed) < len(plan.Steps) {
		progressed := false

		for _, task := range plan.Steps {
			if completed[task.ID] || !dependenciesMet(task, completed) {
				continue
			}

			result, err := e.executeTask(ctx, execution, task, plan.DryRun)
			if err != nil {
				execution.State = classifyFailureState(err)
				execution.Result = &result
				finishedAt := e.now()
				execution.FinishedAt = &finishedAt
				e.publish(execution, task.ID, task.Tool, "execution.failed", types.SeverityError, map[string]any{"error": err.Error()})
				return execution, err
			}

			completed[task.ID] = true
			execution.Result = &result
			if isContextTask(task) {
				ctxengine.MergeToolResult(liveContext, task, result)
				execution.Context = ctxengine.CloneSnapshot(liveContext)
				e.publish(execution, task.ID, task.Tool, "context.updated", types.SeverityInfo, map[string]any{
					"task_id": task.ID,
					"tool":    task.Tool,
				})
			}
			progressed = true
		}

		if !progressed {
			err := fmt.Errorf("%w: unresolved task dependencies", types.ErrInvalidInput)
			execution.State = types.ExecutionFailed
			finishedAt := e.now()
			execution.FinishedAt = &finishedAt
			e.publish(execution, "", "", "execution.failed", types.SeverityError, map[string]any{"error": err.Error()})
			return execution, err
		}
	}

	finishedAt := e.now()
	execution.FinishedAt = &finishedAt
	execution.State = types.ExecutionCompleted
	e.publish(execution, "", "", "execution.completed", types.SeverityInfo, nil)
	return execution, nil
}

func isContextTask(task types.Task) bool {
	return task.Purpose == types.TaskPurposeContext
}

func (e *Engine) validatePlan(ctx context.Context, plan types.Plan) error {
	if plan.ID == "" || plan.Intent == "" || len(plan.Steps) == 0 {
		return types.ErrInvalidInput
	}

	seen := make(map[types.TaskID]types.Task, len(plan.Steps))
	for _, task := range plan.Steps {
		if task.ID == "" || task.Tool == "" {
			return types.ErrInvalidInput
		}
		if _, exists := seen[task.ID]; exists {
			return fmt.Errorf("%w: duplicate task %s", types.ErrInvalidInput, task.ID)
		}
		if _, err := e.registry.Get(task.Tool); err != nil {
			return fmt.Errorf("%w: tool %s", err, task.Tool)
		}
		seen[task.ID] = task
	}

	for _, task := range plan.Steps {
		for _, dependency := range task.Dependencies {
			if _, exists := seen[dependency]; !exists {
				return fmt.Errorf("%w: missing dependency %s", types.ErrInvalidInput, dependency)
			}
		}
	}

	if types.HasDependencyCycle(plan.Steps) {
		return fmt.Errorf("%w: dependency cycle", types.ErrInvalidInput)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (e *Engine) executeTask(ctx context.Context, execution types.Execution, task types.Task, dryRun bool) (types.ToolResult, error) {
	t, err := e.registry.Get(task.Tool)
	if err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	e.publish(execution, task.ID, task.Tool, "task.started", types.SeverityInfo, nil)
	if err := t.Validate(ctx, task.Arguments); err != nil {
		e.publish(execution, task.ID, task.Tool, "task.failed", types.SeverityError, map[string]any{"error": err.Error()})
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	if dryRun && t.Metadata().SupportsDryRun {
		e.publish(execution, task.ID, task.Tool, "tool.dry_run", types.SeverityInfo, map[string]any{
			"tool":   task.Tool,
			"inputs": task.Arguments,
		})
		e.publish(execution, task.ID, task.Tool, "task.completed", types.SeverityInfo, nil)
		return types.ToolResult{
			Success: true,
			Output:  types.ToolOutput{"dry_run": true, "tool": string(task.Tool)},
		}, nil
	}

	attempts := task.Retry.Attempts
	if attempts < 1 {
		attempts = 1
	}

	var lastResult types.ToolResult
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		runCtx := ctx
		cancel := func() {}
		timeout := task.Timeout
		if timeout == 0 {
			timeout = t.Metadata().Timeout
		}
		if timeout > 0 {
			runCtx, cancel = context.WithTimeout(ctx, timeout)
		}

		e.publish(execution, task.ID, task.Tool, "tool.started", types.SeverityInfo, map[string]any{"attempt": attempt})
		lastResult, lastErr = t.Execute(runCtx, task.Arguments)
		cancel()

		if lastErr == nil && lastResult.Success {
			e.publish(execution, task.ID, task.Tool, "tool.completed", types.SeverityInfo, lastResult.Output)
			e.publish(execution, task.ID, task.Tool, "task.completed", types.SeverityInfo, nil)
			return lastResult, nil
		}

		if runCtx.Err() != nil {
			lastErr = runCtx.Err()
			lastResult = types.ToolResult{Success: false, Error: lastErr.Error(), Retryable: true}
		}
		if lastErr == nil {
			lastErr = errors.New(lastResult.Error)
		}
		e.publish(execution, task.ID, task.Tool, "tool.failed", types.SeverityError, map[string]any{"attempt": attempt, "error": lastErr.Error()})

		if !lastResult.Retryable || attempt == attempts {
			break
		}
		if task.Retry.Delay > 0 {
			timer := time.NewTimer(task.Retry.Delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return types.ToolResult{Success: false, Error: ctx.Err().Error()}, ctx.Err()
			case <-timer.C:
			}
		}
	}

	e.publish(execution, task.ID, task.Tool, "task.failed", types.SeverityError, map[string]any{"error": lastErr.Error()})
	return lastResult, lastErr
}

func dependenciesMet(task types.Task, completed map[types.TaskID]bool) bool {
	for _, dependency := range task.Dependencies {
		if !completed[dependency] {
			return false
		}
	}
	return true
}

func classifyFailureState(err error) types.ExecutionState {
	switch {
	case errors.Is(err, context.Canceled):
		return types.ExecutionCancelled
	case errors.Is(err, context.DeadlineExceeded):
		return types.ExecutionTimeout
	default:
		return types.ExecutionFailed
	}
}

func (e *Engine) publish(execution types.Execution, taskID types.TaskID, toolID types.ToolID, eventType types.EventType, severity types.Severity, payload map[string]any) {
	if e.publisher == nil {
		return
	}

	_ = e.publisher.Publish(types.Event{
		ID:          types.EventID(fmt.Sprintf("%s-event-%d", execution.ID, atomic.AddUint64(&e.eventSeq, 1))),
		Timestamp:   e.now(),
		ExecutionID: execution.ID,
		TaskID:      taskID,
		ToolID:      toolID,
		Type:        eventType,
		Source:      "runtime.engine",
		Severity:    severity,
		Payload:     payload,
	})
}
