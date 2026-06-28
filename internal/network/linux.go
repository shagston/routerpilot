package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

type LinuxProvider struct{}

func NewLinuxProvider() *LinuxProvider {
	return &LinuxProvider{}
}

func (p *LinuxProvider) GetInterfaceStatus() ([]InterfaceStatus, error) {
	out, err := exec.Command("ip", "-j", "link", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse interfaces: %w", err)
	}

	var statuses []InterfaceStatus
	for _, iface := range raw {
		name, _ := iface["ifname"].(string)
		flags, _ := iface["flags"].([]any)

		isUp := false
		for _, f := range flags {
			if f == "UP" {
				isUp = true
				break
			}
		}

		statuses = append(statuses, InterfaceStatus{
			Name:   name,
			Up:     isUp,
			Active: isUp,
		})
	}
	return statuses, nil
}

func (p *LinuxProvider) SetInterfaceState(name string, up bool) error {
	state := "down"
	if up {
		state = "up"
	}
	return exec.Command("ip", "link", "set", name, state).Run()
}

func (p *LinuxProvider) GetAddresses() ([]Address, error) {
	out, err := exec.Command("ip", "-j", "addr", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse addresses: %w", err)
	}

	var addresses []Address
	for _, iface := range raw {
		name, _ := iface["ifname"].(string)
		addrInfos, _ := iface["addr_info"].([]any)

		for _, info := range addrInfos {
			if addrMap, ok := info.(map[string]any); ok {
				if local, ok := addrMap["local"].(string); ok {
					addresses = append(addresses, Address{
						Interface: name,
						Address:   local,
					})
				}
			}
		}
	}
	return addresses, nil
}

func (p *LinuxProvider) AddAddress(iface, addr string) error {
	return exec.Command("ip", "addr", "add", addr, "dev", iface).Run()
}

func (p *LinuxProvider) GetRoutes() ([]Route, error) {
	out, err := exec.Command("ip", "-j", "route", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse routes: %w", err)
	}

	var routes []Route
	for _, r := range raw {
		dest, _ := r["dst"].(string)
		gw, _ := r["gateway"].(string)
		dev, _ := r["dev"].(string)

		routes = append(routes, Route{
			Destination: dest,
			Gateway:     gw,
			Interface:   dev,
		})
	}
	return routes, nil
}

func (p *LinuxProvider) AddRoute(dest, gateway, iface string) error {
	args := []string{"route", "add", dest}
	if gateway != "" {
		args = append(args, "via", gateway)
	}
	if iface != "" {
		args = append(args, "dev", iface)
	}
	return exec.Command("ip", args...).Run()
}
