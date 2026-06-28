package network

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type LinuxProvider struct{}

func NewLinuxProvider() *LinuxProvider {
	return &LinuxProvider{}
}

func (p *LinuxProvider) GetInterfaceStatus() ([]InterfaceStatus, error) {
	statuses, err := p.getInterfaceStatusJSON()
	if err == nil {
		return statuses, nil
	}
	return p.getInterfaceStatusText()
}

func (p *LinuxProvider) getInterfaceStatusJSON() ([]InterfaceStatus, error) {
	out, err := exec.Command("ip", "-j", "link", "show").Output()
	if err != nil {
		return nil, err
	}

	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
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

func (p *LinuxProvider) getInterfaceStatusText() ([]InterfaceStatus, error) {
	out, err := exec.Command("ip", "link", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get interfaces: %w", err)
	}

	re := regexp.MustCompile(`^(\d+):\s+(\S+):\s+<([^>]+)>`)
	var statuses []InterfaceStatus

	for _, line := range strings.Split(string(out), "\n") {
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		isUp := false
		for _, flag := range strings.Split(matches[3], ",") {
			if strings.TrimSpace(flag) == "UP" {
				isUp = true
				break
			}
		}

		statuses = append(statuses, InterfaceStatus{
			Name:   matches[2],
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
	addrs, err := p.getAddressesJSON()
	if err == nil {
		return addrs, nil
	}
	return p.getAddressesText()
}

func (p *LinuxProvider) getAddressesJSON() ([]Address, error) {
	out, err := exec.Command("ip", "-j", "addr", "show").Output()
	if err != nil {
		return nil, err
	}

	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
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

func (p *LinuxProvider) getAddressesText() ([]Address, error) {
	out, err := exec.Command("ip", "addr", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	ifaceRe := regexp.MustCompile(`^(\d+):\s+(\S+):`)
	addrRe := regexp.MustCompile(`inet\s+(\S+)/(\d+)`)

	var addresses []Address
	var currentIface string

	for _, line := range strings.Split(string(out), "\n") {
		if ifaceMatch := ifaceRe.FindStringSubmatch(line); ifaceMatch != nil {
			currentIface = ifaceMatch[2]
		}
		if addrMatch := addrRe.FindStringSubmatch(line); addrMatch != nil && currentIface != "" {
			addresses = append(addresses, Address{
				Interface: currentIface,
				Address:   addrMatch[1],
			})
		}
	}
	return addresses, nil
}

func (p *LinuxProvider) AddAddress(iface, addr string) error {
	return exec.Command("ip", "addr", "add", addr, "dev", iface).Run()
}

func (p *LinuxProvider) GetRoutes() ([]Route, error) {
	routes, err := p.getRoutesJSON()
	if err == nil {
		return routes, nil
	}
	return p.getRoutesText()
}

func (p *LinuxProvider) getRoutesJSON() ([]Route, error) {
	out, err := exec.Command("ip", "-j", "route", "show").Output()
	if err != nil {
		return nil, err
	}

	var raw []map[string]any
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
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

func (p *LinuxProvider) getRoutesText() ([]Route, error) {
	out, err := exec.Command("ip", "route", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get routes: %w", err)
	}

	re := regexp.MustCompile(`^(\S+)(?:\s+via\s+(\S+))?(?:\s+dev\s+(\S+))?`)
	var routes []Route

	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		routes = append(routes, Route{
			Destination: matches[1],
			Gateway:     matches[2],
			Interface:   matches[3],
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
