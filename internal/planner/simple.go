package planner

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

type SimplePlanner struct{}

func NewSimplePlanner() *SimplePlanner {
	return &SimplePlanner{}
}

func (p *SimplePlanner) Plan(ctx context.Context, intent planner.Intent, snapshot types.ContextSnapshot) (types.Plan, error) {
	switch intent.Name {
	case "ping":
		return p.planPing(intent)
	case "interface.status":
		return p.planInterfaceStatus(intent)
	case "interface.set":
		return p.planInterfaceSet(intent)
	case "ip.show":
		return p.planIPShow(intent)
	case "ip.set":
		return p.planIPSet(intent)
	case "route.show":
		return p.planRouteShow(intent)
	case "route.add":
		return p.planRouteAdd(intent)
	case "diagnose":
		return p.planDiagnose(intent)
	default:
		return types.Plan{}, fmt.Errorf("unsupported intent: %s", intent.Name)
	}
}

func (p *SimplePlanner) planPing(intent planner.Intent) (types.Plan, error) {
	target, ok := intent.Arguments["target"].(string)
	if !ok || target == "" {
		return types.Plan{}, fmt.Errorf("ping intent requires 'target' argument")
	}

	return types.Plan{
		ID:     types.PlanID(fmt.Sprintf("plan-ping-%d", time.Now().UnixNano())),
		Intent: fmt.Sprintf("Ping host %s", target),
		Steps: []types.Task{
			{
				ID:   types.TaskID("ping"),
				Tool: "network.ping",
				Arguments: types.ToolInput{
					"host":  target,
					"count": 4,
				},
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planInterfaceStatus(intent planner.Intent) (types.Plan, error) {
	iface, _ := intent.Arguments["interface"].(string)
	if iface == "" {
		iface = "all"
	}

	return types.Plan{
		ID:     types.PlanID("plan-if-status"),
		Intent: fmt.Sprintf("Get interface status for %s", iface),
		Steps: []types.Task{
			{
				ID:   types.TaskID("if-status"),
				Tool: "network.interface_status",
				Arguments: types.ToolInput{
					"interface": iface,
				},
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planInterfaceSet(intent planner.Intent) (types.Plan, error) {
	iface, ok := intent.Arguments["interface"].(string)
	if !ok || iface == "" {
		return types.Plan{}, fmt.Errorf("interface.set intent requires 'interface' argument")
	}
	state, ok := intent.Arguments["state"].(string)
	if !ok || (state != "up" && state != "down") {
		return types.Plan{}, fmt.Errorf("interface.set intent requires 'state' argument (up/down)")
	}

	return types.Plan{
		ID:     types.PlanID("plan-if-set"),
		Intent: fmt.Sprintf("Set interface %s to %s", iface, state),
		Steps: []types.Task{
			{
				ID:      types.TaskID("if-before"),
				Tool:    "network.interface_status",
				Purpose: types.TaskPurposeContext,
				Arguments: types.ToolInput{
					"interface": iface,
				},
			},
			{
				ID:      types.TaskID("if-set"),
				Tool:    "network.interface_set_state",
				Purpose: types.TaskPurposeAction,
				Arguments: types.ToolInput{
					"interface": iface,
					"state":     state,
				},
				Dependencies: []types.TaskID{"if-before"},
			},
		},
		Risk: types.RiskMedium,
	}, nil
}

func (p *SimplePlanner) planIPShow(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-ip-show"),
		Intent: "Show IP addresses",
		Steps: []types.Task{
			{
				ID:   types.TaskID("ip-show"),
				Tool: "network.ip_address_get",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planIPSet(intent planner.Intent) (types.Plan, error) {
	iface, ok := intent.Arguments["interface"].(string)
	if !ok || iface == "" {
		return types.Plan{}, fmt.Errorf("ip.set intent requires 'interface' argument")
	}
	address, ok := intent.Arguments["address"].(string)
	if !ok || address == "" {
		return types.Plan{}, fmt.Errorf("ip.set intent requires 'address' argument (CIDR)")
	}

	return types.Plan{
		ID:     types.PlanID("plan-ip-set"),
		Intent: fmt.Sprintf("Set IP %s on %s", address, iface),
		Steps: []types.Task{
			{
				ID:      types.TaskID("ip-before"),
				Tool:    "network.ip_address_get",
				Purpose: types.TaskPurposeContext,
			},
			{
				ID:   types.TaskID("ip-set"),
				Tool: "network.ip_address_set",
				Arguments: types.ToolInput{
					"interface": iface,
					"address":   address,
				},
				Dependencies: []types.TaskID{"ip-before"},
			},
		},
		Risk: types.RiskMedium,
	}, nil
}

func (p *SimplePlanner) planRouteShow(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-route-show"),
		Intent: "Show routing table",
		Steps: []types.Task{
			{
				ID:   types.TaskID("route-show"),
				Tool: "network.route_get",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planRouteAdd(intent planner.Intent) (types.Plan, error) {
	dest, ok := intent.Arguments["destination"].(string)
	if !ok || dest == "" {
		return types.Plan{}, fmt.Errorf("route.add intent requires 'destination' argument")
	}
	gateway, ok := intent.Arguments["gateway"].(string)
	if !ok || gateway == "" {
		return types.Plan{}, fmt.Errorf("route.add intent requires 'gateway' argument")
	}
	iface, _ := intent.Arguments["interface"].(string)

	return types.Plan{
		ID:     types.PlanID("plan-route-add"),
		Intent: fmt.Sprintf("Add route to %s via %s", dest, gateway),
		Steps: []types.Task{
			{
				ID:      types.TaskID("route-before"),
				Tool:    "network.route_get",
				Purpose: types.TaskPurposeContext,
			},
			{
				ID:   types.TaskID("route-add"),
				Tool: "network.route_add",
				Arguments: types.ToolInput{
					"destination": dest,
					"gateway":     gateway,
					"interface":   iface,
				},
				Dependencies: []types.TaskID{"route-before"},
			},
		},
		Risk: types.RiskMedium,
	}, nil
}

func (p *SimplePlanner) planDiagnose(intent planner.Intent) (types.Plan, error) {
	target, _ := intent.Arguments["target"].(string)
	if target == "" {
		target = "8.8.8.8"
	}

	return types.Plan{
		ID:     types.PlanID("plan-diagnose"),
		Intent: fmt.Sprintf("Network diagnostics (target: %s)", target),
		Steps: []types.Task{
			{
				ID:      types.TaskID("diag-if"),
				Tool:    "network.interface_status",
				Purpose: types.TaskPurposeContext,
				Arguments: types.ToolInput{
					"interface": "all",
				},
			},
			{
				ID:      types.TaskID("diag-ip"),
				Tool:    "network.ip_address_get",
				Purpose: types.TaskPurposeContext,
			},
			{
				ID:      types.TaskID("diag-route"),
				Tool:    "network.route_get",
				Purpose: types.TaskPurposeContext,
			},
			{
				ID:   types.TaskID("diag-ping"),
				Tool: "network.ping",
				Arguments: types.ToolInput{
					"host":  target,
					"count": 4,
				},
			},
		},
		Risk: types.RiskLow,
	}, nil
}
