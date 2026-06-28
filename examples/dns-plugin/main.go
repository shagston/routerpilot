package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type manifest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
}

type toolResult struct {
	Success bool                   `json:"success"`
	Output  map[string]any         `json:"output,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type input struct {
	Host string `json:"host"`
}

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	switch os.Args[1] {
	case "plugin-manifest":
		m := manifest{
			ID:      "dns.lookup",
			Name:    "DNS Lookup",
			Version: "0.1.0",
			Type:    "subprocess",
		}
		json.NewEncoder(os.Stdout).Encode(m)

	case "execute":
		var in input
		if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
			json.NewEncoder(os.Stdout).Encode(toolResult{Success: false, Error: fmt.Sprintf("invalid input: %v", err)})
			return
		}

		host := strings.TrimSpace(in.Host)
		if host == "" {
			json.NewEncoder(os.Stdout).Encode(toolResult{Success: false, Error: "host is required"})
			return
		}

		ips, err := net.LookupHost(host)
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(toolResult{Success: false, Error: err.Error()})
			return
		}

		json.NewEncoder(os.Stdout).Encode(toolResult{
			Success: true,
			Output: map[string]any{
				"host":      host,
				"addresses": ips,
				"timestamp": time.Now().Format(time.RFC3339),
			},
		})

	default:
		os.Exit(1)
	}
}
