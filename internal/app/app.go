package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ctxengine "github.com/shagston/routerpilot/internal/context"
	eventbus "github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/internal/network"
	"github.com/shagston/routerpilot/internal/planner"
	pluginloader "github.com/shagston/routerpilot/internal/plugin"
	"github.com/shagston/routerpilot/internal/registry"
	runtimeengine "github.com/shagston/routerpilot/internal/runtime"
	"github.com/shagston/routerpilot/internal/safety"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
	bridgeTools "github.com/shagston/routerpilot/tools/bridge"
	dhcptools "github.com/shagston/routerpilot/tools/dhcp"
	dnstools "github.com/shagston/routerpilot/tools/dns"
	firewalltools "github.com/shagston/routerpilot/tools/firewall"
	networktools "github.com/shagston/routerpilot/tools/network"
	packagetools "github.com/shagston/routerpilot/tools/package"
	servicetools "github.com/shagston/routerpilot/tools/service"
	systemtools "github.com/shagston/routerpilot/tools/system"
	vpntools "github.com/shagston/routerpilot/tools/vpn"
	wifitools "github.com/shagston/routerpilot/tools/wifi"
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
		networktools.TracerouteTool{},
		networktools.NeighborsTool{},
		networktools.ConnectionsTool{},
		systemtools.InfoTool{},
		systemtools.UptimeTool{},
		systemtools.RebootTool{},
		systemtools.LogsTool{},
		systemtools.MemoryTool{},
		systemtools.DiskTool{},
		systemtools.ProcessesTool{},
		dnstools.LookupTool{},
		dnstools.StatusTool{},
		dnstools.FlushTool{},
		wifitools.ScanTool{},
		wifitools.StatusTool{},
		dhcptools.LeasesTool{},
		firewalltools.StatusTool{},
		firewalltools.ReloadTool{},
		servicetools.ListTool{},
		servicetools.RestartTool{},
		packagetools.ListTool{},
		vpntools.StatusTool{},
		bridgeTools.StatusTool{},
	} {
		if err := reg.Register(t); err != nil {
			return nil, err
		}
	}

	bus := eventbus.NewBus()
	validateCfg := parsePermissionsConfig()
	engine := runtimeengine.NewEngine(reg, bus, runtimeengine.WithValidator(safety.NewValidator(reg, validateCfg)))

	plugDir := os.Getenv("ROUTERPILOT_PLUGIN_DIR")
	if plugDir == "" {
		plugDir = "plugins"
	}
	plugLoader := pluginloader.NewLoader(plugDir)
	if err := plugLoader.LoadAll(context.Background(), reg); err != nil {
		return nil, fmt.Errorf("plugin loading: %w", err)
	}

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
	guard := safety.NewSimpleSafetyGuard(parseRiskLevel())
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

func (a *App) PreviewPlan(ctx context.Context, intent sdkPlanner.Intent) (types.ContextSnapshot, types.Plan, error) {
	ctxProvider := ctxengine.NewSystemContextProvider(a.Registry, a.Events)
	planGen := planner.SelectPlanner(a.Registry)

	snapshot, err := ctxProvider.Build(ctx, intent)
	if err != nil {
		return nil, types.Plan{}, fmt.Errorf("context build failed: %w", err)
	}

	plan, err := planGen.Plan(ctx, intent, snapshot)
	if err != nil {
		return nil, types.Plan{}, fmt.Errorf("planning failed: %w", err)
	}

	return snapshot, plan, nil
}

func snapshotKeys(snapshot types.ContextSnapshot) []string {
	keys := make([]string, 0, len(snapshot))
	for key := range snapshot {
		keys = append(keys, key)
	}
	return keys
}

func parsePermissionsConfig() safety.Config {
	permMap := map[string]types.Permission{
		"read":  types.PermissionRead,
		"write": types.PermissionWrite,
		"admin": types.PermissionAdmin,
	}

	raw := os.Getenv("ROUTERPILOT_PERMISSIONS")
	if raw == "" {
		return safety.Config{
			Permissions: []types.Permission{types.PermissionRead, types.PermissionWrite},
		}
	}

	var perms []types.Permission
	for _, p := range strings.Split(raw, ",") {
		p = strings.TrimSpace(strings.ToLower(p))
		if perm, ok := permMap[p]; ok {
			perms = append(perms, perm)
		}
	}
	if len(perms) == 0 {
		perms = []types.Permission{types.PermissionRead, types.PermissionWrite}
	}
	return safety.Config{Permissions: perms}
}

func parseRiskLevel() types.RiskLevel {
	switch strings.ToLower(os.Getenv("ROUTERPILOT_RISK")) {
	case "low":
		return types.RiskLow
	case "medium":
		return types.RiskMedium
	case "high":
		return types.RiskHigh
	case "critical":
		return types.RiskCritical
	default:
		return types.RiskLow
	}
}
