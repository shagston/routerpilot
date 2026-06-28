package app

import (
	"context"
	"fmt"
	"time"

	ctxengine "github.com/shagston/routerpilot/internal/context"
	eventbus "github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/internal/network"
	"github.com/shagston/routerpilot/internal/planner"
	"github.com/shagston/routerpilot/internal/registry"
	runtimeengine "github.com/shagston/routerpilot/internal/runtime"
	"github.com/shagston/routerpilot/internal/safety"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
	networktools "github.com/shagston/routerpilot/tools/network"
)

type App struct {
	Registry *registry.ToolRegistry
	Events   *eventbus.Bus
	Runtime  *runtimeengine.Engine
}

type SafetyError struct {
	Plan     types.Plan
	Snapshot types.ContextSnapshot
}

func (e *SafetyError) Error() string {
	return fmt.Sprintf("safety confirmation required for plan %s", e.Plan.ID)
}

func New() (*App, error) {
	reg := registry.NewToolRegistry()

	netProv := network.NewLinuxProvider()

	for _, t := range []tool.Tool{
		networktools.PingTool{},
		networktools.InterfaceStatusTool{Provider: netProv},
		networktools.InterfaceSetStateTool{Provider: netProv},
		networktools.IPAddressGetTool{Provider: netProv},
		networktools.IPAddressSetTool{Provider: netProv},
		networktools.RouteGetTool{Provider: netProv},
		networktools.RouteAddTool{Provider: netProv},
	} {
		if err := reg.Register(t); err != nil {
			return nil, err
		}
	}

	bus := eventbus.NewBus()
	engine := runtimeengine.NewEngine(reg, bus, runtimeengine.WithValidator(safety.NewValidator(reg, safety.Config{
		Permissions: []types.Permission{types.PermissionRead},
	})))

	return &App{
		Registry: reg,
		Events:   bus,
		Runtime:  engine,
	}, nil
}

func (a *App) ExecuteIntent(ctx context.Context, intent sdkPlanner.Intent, interactive bool) (*types.Execution, error) {
	a.publishEvent("intent.received", types.SeverityInfo, map[string]any{
		"intent": intent.Name,
		"args":   intent.Arguments,
	})

	ctxProvider := ctxengine.NewSystemContextProvider(a.Registry, a.Events)
	guard := safety.NewSimpleSafetyGuard(types.RiskLow)
	planGen := planner.SelectPlanner(a.Registry)

	const maxAttempts = 3
	var lastExecution types.Execution
	var hasLastExecution bool
	var lastError error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		snapshot, err := ctxProvider.Build(ctx, intent)
		if err != nil {
			a.publishEvent("context.failed", types.SeverityError, map[string]any{"error": err.Error()})
			return nil, fmt.Errorf("context build failed: %w", err)
		}
		a.publishEvent("context.built", types.SeverityInfo, map[string]any{
			"attempt": attempt,
			"sources": snapshotKeys(snapshot),
		})

		if attempt > 1 && lastError != nil {
			snapshot["execution_error"] = fmt.Sprintf("Previous attempt failed: %v", lastError)
			if hasLastExecution && lastExecution.Result != nil {
				snapshot["last_result"] = *lastExecution.Result
			}
		}

		plan, err := planGen.Plan(ctx, intent, snapshot)
		if err != nil {
			a.publishEvent("planning.failed", types.SeverityError, map[string]any{
				"attempt": attempt,
				"error":   err.Error(),
			})
			return nil, fmt.Errorf("planning failed: %w", err)
		}
		a.publishEvent("plan.created", types.SeverityInfo, map[string]any{
			"attempt": attempt,
			"plan_id": plan.ID,
			"risk":    plan.Risk,
			"steps":   len(plan.Steps),
		})

		safe, err := guard.Validate(plan)
		if err != nil {
			return nil, fmt.Errorf("safety validation error: %w", err)
		}

		if !safe {
			a.publishEvent("safety.confirmation_required", types.SeverityWarning, map[string]any{
				"plan_id": plan.ID,
				"risk":    plan.Risk,
			})
			return nil, &SafetyError{Plan: plan, Snapshot: snapshot}
		}

		execution, err := a.executeAdaptivePlan(ctx, planGen, guard, intent, snapshot, plan)
		lastExecution = execution
		hasLastExecution = true
		lastError = err

		if err == nil {
			return &execution, nil
		}
	}

	if hasLastExecution {
		return &lastExecution, fmt.Errorf("failed to achieve intent after %d attempts. Last error: %v", maxAttempts, lastError)
	}
	return nil, fmt.Errorf("failed to achieve intent after %d attempts. Last error: %v", maxAttempts, lastError)
}

func (a *App) publishEvent(eventType types.EventType, severity types.Severity, payload map[string]any) {
	_ = a.Events.Publish(types.Event{
		ID:        types.EventID(fmt.Sprintf("app-event-%d", time.Now().UnixNano())),
		Timestamp: time.Now(),
		Type:      eventType,
		Source:    "app",
		Severity:  severity,
		Payload:   payload,
	})
}

func snapshotKeys(snapshot types.ContextSnapshot) []string {
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	return keys
}
