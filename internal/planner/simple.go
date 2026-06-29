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
	case "system.info":
		return p.planSystemInfo(intent)
	case "system.uptime":
		return p.planSystemUptime(intent)
	case "system.reboot":
		return p.planSystemReboot(intent)
	case "wifi.scan":
		return p.planWifiScan(intent)
	case "dns.lookup":
		return p.planDNSLookup(intent)
	case "dns.status":
		return p.planDNSStatus(intent)
	case "dns.flush":
		return p.planDNSFlush(intent)
	case "dhcp.leases":
		return p.planDHCPLeases(intent)
	case "firewall.status":
		return p.planFirewallStatus(intent)
	case "firewall.reload":
		return p.planFirewallReload(intent)
	case "system.logs":
		return p.planSystemLogs(intent)
	case "system.memory":
		return p.planSystemMemory(intent)
	case "system.disk":
		return p.planSystemDisk(intent)
	case "system.processes":
		return p.planSystemProcesses(intent)
	case "wifi.status":
		return p.planWifiStatus(intent)
	case "wifi.connect":
		return p.planWiFiConnect(intent)
	case "network.traceroute":
		return p.planTraceroute(intent)
	case "network.neighbors":
		return p.planNeighbors(intent)
	case "network.connections":
		return p.planConnections(intent)
	case "service.list":
		return p.planServiceList(intent)
	case "service.restart":
		return p.planServiceRestart(intent)
	case "package.list":
		return p.planPackageList(intent)
	case "vpn.status":
		return p.planVPNStatus(intent)
	case "bridge.status":
		return p.planBridgeStatus(intent)
	case "diagnose":
		return p.planDiagnose(intent)
	case "suggest":
		return p.planSuggest(intent)
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

func (p *SimplePlanner) planSystemInfo(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-sys-info"),
		Intent: "Get system information",
		Steps: []types.Task{
			{
				ID:   types.TaskID("sys-info"),
				Tool: "system.info",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planSystemUptime(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-sys-uptime"),
		Intent: "Get system uptime",
		Steps: []types.Task{
			{
				ID:   types.TaskID("sys-uptime"),
				Tool: "system.uptime",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planWifiScan(intent planner.Intent) (types.Plan, error) {
	plan := types.Plan{
		ID:     types.PlanID("plan-wifi-scan"),
		Intent: "Wi-Fi scan",
		Steps: []types.Task{
			{
				ID:   types.TaskID("wifi-scan"),
				Tool: "wifi.scan",
			},
		},
		Risk: types.RiskLow,
	}

	if iface, ok := intent.Arguments["interface"].(string); ok && iface != "" {
		plan.Steps[0].Arguments = types.ToolInput{
			"interface": iface,
		}
	}

	return plan, nil
}

func (p *SimplePlanner) planSystemReboot(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-sys-reboot"),
		Intent: "Reboot system",
		Steps: []types.Task{
			{
				ID:   types.TaskID("sys-reboot"),
				Tool: "system.reboot",
			},
		},
		Risk: types.RiskHigh,
	}, nil
}

func (p *SimplePlanner) planDNSLookup(intent planner.Intent) (types.Plan, error) {
	target, ok := intent.Arguments["target"].(string)
	if !ok || target == "" {
		return types.Plan{}, fmt.Errorf("dns.lookup intent requires 'target' argument")
	}

	return types.Plan{
		ID:     types.PlanID("plan-dns-lookup"),
		Intent: fmt.Sprintf("DNS lookup for %s", target),
		Steps: []types.Task{
			{
				ID:   types.TaskID("dns-lookup"),
				Tool: "dns.lookup",
				Arguments: types.ToolInput{
					"host": target,
				},
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planDNSStatus(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-dns-status"),
		Intent: "Show DNS resolver status",
		Steps: []types.Task{
			{
				ID:   types.TaskID("dns-status"),
				Tool: "dns.status",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planDNSFlush(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-dns-flush"),
		Intent: "Flush DNS cache",
		Steps: []types.Task{
			{
				ID:   types.TaskID("dns-flush"),
				Tool: "dns.flush",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planDHCPLeases(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-dhcp-leases"),
		Intent: "List DHCP leases",
		Steps: []types.Task{
			{
				ID:   types.TaskID("dhcp-leases"),
				Tool: "dhcp.leases",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planFirewallStatus(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-fw-status"),
		Intent: "Show firewall status",
		Steps: []types.Task{
			{
				ID:   types.TaskID("fw-status"),
				Tool: "firewall.status",
			},
		},
		Risk: types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planFirewallReload(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-fw-reload"),
		Intent: "Reload firewall",
		Steps: []types.Task{
			{
				ID:   types.TaskID("fw-reload"),
				Tool: "firewall.reload",
			},
		},
		Risk: types.RiskMedium,
	}, nil
}

func (p *SimplePlanner) planSystemLogs(intent planner.Intent) (types.Plan, error) {
	plan := types.Plan{
		ID:     types.PlanID("plan-sys-logs"),
		Intent: "View system logs",
		Steps: []types.Task{
			{
				ID:   types.TaskID("sys-logs"),
				Tool: "system.logs",
			},
		},
		Risk: types.RiskLow,
	}

	if lines, ok := intent.Arguments["lines"].(int); ok && lines > 0 && lines <= 500 {
		plan.Steps[0].Arguments = types.ToolInput{"lines": lines}
	}

	if filter, ok := intent.Arguments["filter"].(string); ok && filter != "" {
		if plan.Steps[0].Arguments == nil {
			plan.Steps[0].Arguments = types.ToolInput{}
		}
		plan.Steps[0].Arguments["filter"] = filter
	}

	return plan, nil
}

func (p *SimplePlanner) planWifiStatus(intent planner.Intent) (types.Plan, error) {
	plan := types.Plan{
		ID:     types.PlanID("plan-wifi-status"),
		Intent: "Wi-Fi interface status",
		Steps: []types.Task{
			{
				ID:   types.TaskID("wifi-status"),
				Tool: "wifi.status",
			},
		},
		Risk: types.RiskLow,
	}

	if iface, ok := intent.Arguments["interface"].(string); ok && iface != "" {
		plan.Steps[0].Arguments = types.ToolInput{"interface": iface}
	}

	return plan, nil
}

func (p *SimplePlanner) planWiFiConnect(intent planner.Intent) (types.Plan, error) {
	ssid, ok := intent.Arguments["ssid"].(string)
	if !ok || ssid == "" {
		return types.Plan{}, fmt.Errorf("wifi.connect intent requires 'ssid' argument")
	}

	plan := types.Plan{
		ID:     types.PlanID("plan-wifi-connect"),
		Intent: fmt.Sprintf("Connect to Wi-Fi network %s", ssid),
		Steps: []types.Task{
			{
				ID:      types.TaskID("wifi-before"),
				Tool:    "wifi.status",
				Purpose: types.TaskPurposeContext,
			},
			{
				ID:   types.TaskID("wifi-connect"),
				Tool: "wifi.connect",
				Arguments: types.ToolInput{
					"ssid": ssid,
				},
				Dependencies: []types.TaskID{"wifi-before"},
			},
		},
		Risk: types.RiskMedium,
	}

	if password, ok := intent.Arguments["password"].(string); ok && password != "" {
		plan.Steps[1].Arguments["password"] = password
	}
	if iface, ok := intent.Arguments["interface"].(string); ok && iface != "" {
		plan.Steps[1].Arguments["interface"] = iface
	}

	return plan, nil
}

func (p *SimplePlanner) planSystemMemory(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-sys-mem"),
		Intent: "Show memory usage",
		Steps:  []types.Task{{ID: types.TaskID("sys-mem"), Tool: "system.memory"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planSystemDisk(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-sys-disk"),
		Intent: "Show disk usage",
		Steps:  []types.Task{{ID: types.TaskID("sys-disk"), Tool: "system.disk"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planSystemProcesses(intent planner.Intent) (types.Plan, error) {
	plan := types.Plan{
		ID:     types.PlanID("plan-sys-procs"),
		Intent: "List processes",
		Steps:  []types.Task{{ID: types.TaskID("sys-procs"), Tool: "system.processes"}},
		Risk:   types.RiskLow,
	}
	if sort, ok := intent.Arguments["sort"].(string); ok && sort != "" {
		plan.Steps[0].Arguments = types.ToolInput{"sort": sort}
	}
	return plan, nil
}

func (p *SimplePlanner) planTraceroute(intent planner.Intent) (types.Plan, error) {
	target, ok := intent.Arguments["target"].(string)
	if !ok || target == "" {
		return types.Plan{}, fmt.Errorf("traceroute intent requires 'target' argument")
	}
	return types.Plan{
		ID:     types.PlanID("plan-traceroute"),
		Intent: fmt.Sprintf("Trace route to %s", target),
		Steps:  []types.Task{{ID: types.TaskID("traceroute"), Tool: "network.traceroute", Arguments: types.ToolInput{"host": target}}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planNeighbors(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-neighbors"),
		Intent: "Show ARP/neighbor table",
		Steps:  []types.Task{{ID: types.TaskID("neighbors"), Tool: "network.neighbors"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planConnections(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-connections"),
		Intent: "Show active connections",
		Steps:  []types.Task{{ID: types.TaskID("connections"), Tool: "network.connections"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planServiceList(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-svc-list"),
		Intent: "List services",
		Steps:  []types.Task{{ID: types.TaskID("svc-list"), Tool: "service.list"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planServiceRestart(intent planner.Intent) (types.Plan, error) {
	name, ok := intent.Arguments["name"].(string)
	if !ok || name == "" {
		return types.Plan{}, fmt.Errorf("service.restart intent requires 'name' argument")
	}
	return types.Plan{
		ID:     types.PlanID("plan-svc-restart"),
		Intent: fmt.Sprintf("Restart service %s", name),
		Steps:  []types.Task{{ID: types.TaskID("svc-restart"), Tool: "service.restart", Arguments: types.ToolInput{"name": name}}},
		Risk:   types.RiskMedium,
	}, nil
}

func (p *SimplePlanner) planPackageList(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-pkg-list"),
		Intent: "List packages",
		Steps:  []types.Task{{ID: types.TaskID("pkg-list"), Tool: "package.list"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planVPNStatus(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-vpn-status"),
		Intent: "Show VPN status",
		Steps:  []types.Task{{ID: types.TaskID("vpn-status"), Tool: "vpn.status"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planBridgeStatus(intent planner.Intent) (types.Plan, error) {
	return types.Plan{
		ID:     types.PlanID("plan-bridge-status"),
		Intent: "Show bridge status",
		Steps:  []types.Task{{ID: types.TaskID("bridge-status"), Tool: "bridge.status"}},
		Risk:   types.RiskLow,
	}, nil
}

func (p *SimplePlanner) planSuggest(intent planner.Intent) (types.Plan, error) {
	problem, _ := intent.Arguments["problem"].(string)
	if problem == "" {
		if target, ok := intent.Arguments["target"].(string); ok {
			problem = target
		}
	}
	if problem == "" {
		problem = "diagnose"
	}

	return types.Plan{
		ID:     types.PlanID("plan-suggest"),
		Intent: fmt.Sprintf("Suggest solutions for: %s", problem),
		Steps: []types.Task{
			{
				ID:   types.TaskID("suggest"),
				Tool: "suggest",
				Arguments: types.ToolInput{
					"problem": problem,
				},
			},
		},
		Risk: types.RiskLow,
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
