package context

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/internal/safety"
	"github.com/shagston/routerpilot/sdk/events"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

type SystemContextProvider struct {
	registry  *registry.ToolRegistry
	publisher events.Publisher
}

var defaultDiscoveryTools = []types.ToolID{
	"network.interface_status",
	"network.ip_address_get",
	"network.route_get",
}

var intentDependencies = map[string][]types.ToolID{
	"ping":           {"network.interface_status"},
	"interface.set":  {"network.interface_status"},
	"routing.update": {"network.interface_status", "network.ping"},
}

const contextToolTimeout = 10 * time.Second

func NewSystemContextProvider(reg *registry.ToolRegistry, publisher events.Publisher) *SystemContextProvider {
	return &SystemContextProvider{
		registry:  reg,
		publisher: publisher,
	}
}

func (s *SystemContextProvider) Build(ctx context.Context, intent sdkPlanner.Intent) (types.ContextSnapshot, error) {
	contextData := make(types.ContextSnapshot)

	toolsToRun, ok := intentDependencies[intent.Name]
	if !ok {
		toolsToRun = defaultDiscoveryTools
	}

	gatherID := types.ExecutionID(fmt.Sprintf("context-gather-%d", time.Now().UnixNano()))

	for _, toolID := range toolsToRun {
		t, err := s.registry.Get(toolID)
		if err != nil {
			continue
		}

		args := s.argsForTool(toolID)
		metadata := t.Metadata()
		logSource := fmt.Sprintf("context.system.%s", toolID)

		s.publishEvent(gatherID, toolID, "context.tool.started", types.SeverityInfo, map[string]any{
			"tool":   toolID,
			"source": logSource,
		})

		if err := t.Validate(ctx, args); err != nil {
			contextData[string(toolID)] = fmt.Sprintf("validation error: %v", err)
			s.publishEvent(gatherID, toolID, "context.tool.failed", types.SeverityWarning, map[string]any{
				"tool":  toolID,
				"error": err.Error(),
			})
			continue
		}

		if err := safety.ValidateInput(t.InputSchema(), args); err != nil {
			contextData[string(toolID)] = fmt.Sprintf("schema error: %v", err)
			s.publishEvent(gatherID, toolID, "context.tool.failed", types.SeverityWarning, map[string]any{
				"tool":  toolID,
				"error": err.Error(),
			})
			continue
		}

		toolCtx, cancel := context.WithTimeout(ctx, resolveTimeout(metadata.Timeout))
		result, err := t.Execute(toolCtx, args)
		cancel()

		if err != nil {
			contextData[string(toolID)] = fmt.Sprintf("error: %v", err)
			s.publishEvent(gatherID, toolID, "context.tool.failed", types.SeverityWarning, map[string]any{
				"tool":  toolID,
				"error": err.Error(),
			})
		} else {
			contextData[string(toolID)] = result
			s.publishEvent(gatherID, toolID, "context.tool.completed", types.SeverityInfo, map[string]any{
				"tool": toolID,
			})
		}
	}

	return contextData, nil
}

func (s *SystemContextProvider) argsForTool(toolID types.ToolID) types.ToolInput {
	args := types.ToolInput{}
	switch toolID {
	case "network.interface_status":
		args["interface"] = "all"
	case "network.ping":
		args["host"] = "8.8.8.8"
		args["count"] = 1
	}
	return args
}

func resolveTimeout(t time.Duration) time.Duration {
	if t > 0 && t < contextToolTimeout {
		return t
	}
	return contextToolTimeout
}

func (s *SystemContextProvider) publishEvent(executionID types.ExecutionID, toolID types.ToolID, eventType types.EventType, severity types.Severity, payload map[string]any) {
	if s.publisher == nil {
		return
	}
	s.publisher.Publish(types.Event{
		ID:          types.EventID(fmt.Sprintf("ctx-%d", time.Now().UnixNano())),
		Timestamp:   time.Now(),
		ExecutionID: executionID,
		ToolID:      toolID,
		Type:        eventType,
		Source:      "context.system",
		Severity:    severity,
		Payload:     payload,
	})
}
