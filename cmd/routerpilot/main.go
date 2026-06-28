package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/shagston/routerpilot/internal/api"
	"github.com/shagston/routerpilot/internal/app"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "ping":
		if err := runPing(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "plan":
		if err := runPlan(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "tools":
		if err := listTools(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(2)
	}
}

func runPing(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("host is required")
	}

	showEvents := false
	filtered := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--events" {
			showEvents = true
			continue
		}
		filtered = append(filtered, arg)
	}
	args = filtered
	if len(args) < 1 {
		return fmt.Errorf("host is required")
	}

	host := args[0]
	count := 4
	if len(args) > 1 {
		parsed, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("count must be an integer: %w", err)
		}
		count = parsed
	}

	instance, err := app.New()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	execution, err := instance.Runtime.Execute(ctx, types.Plan{
		ID:     "plan-cli-ping",
		Intent: "network.ping",
		Steps: []types.Task{
			{
				ID:        "ping",
				Tool:      "network.ping",
				Arguments: types.ToolInput{"host": host, "count": count},
				Retry:     types.RetryPolicy{Attempts: 1},
			},
		},
		Risk: types.RiskLow,
	}, types.ContextSnapshot{"source": "cli"})
	if err != nil {
		return err
	}

	payload := map[string]any{"result": execution.Result}
	if showEvents {
		payload["events"] = instance.Events.Events()
	}

	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func runPlan(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("intent name is required (e.g., 'ping')")
	}

	intentName := args[0]
	intentArgs := make(map[string]any)
	if len(args) > 1 {
		intentArgs["target"] = args[1]
	}

	intent := sdkPlanner.Intent{
		Name:      intentName,
		Arguments: intentArgs,
	}

	instance, err := app.New()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	execution, err := instance.ExecuteIntent(ctx, intent, true)
	if err != nil {
		var safetyErr *app.SafetyError
		if errors.As(err, &safetyErr) {
			fmt.Printf("\n⚠️  WARNING: This plan is marked as %s risk. Do you want to proceed? (y/N): ", safetyErr.Plan.Risk)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				return fmt.Errorf("execution aborted by user due to safety risk")
			}
			fmt.Println("Proceeding with caution...")

			// Выполняем план напрямую через Runtime
			execResult, execErr := instance.Runtime.Execute(ctx, safetyErr.Plan, safetyErr.Snapshot)
			if execErr != nil {
				return fmt.Errorf("execution failed after confirmation: %w", execErr)
			}
			execution = &execResult
		} else {
			return err
		}
	}

	fmt.Println("Plan executed successfully!")
	payload := map[string]any{
		"result": execution.Result,
	}
	encoded, _ := json.MarshalIndent(payload, "", "  ")
	fmt.Println(string(encoded))

	return nil
}

func listTools() error {
	instance, err := app.New()
	if err != nil {
		return err
	}

	encoded, err := json.MarshalIndent(instance.Registry.List(), "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

func runServe() error {
	instance, err := app.New()
	if err != nil {
		return fmt.Errorf("failed to initialize app: %w", err)
	}

	server := api.NewServer(instance)
	port := ":8080"

	fmt.Printf("RouterPilot API server starting on %s...\n", port)
	fmt.Println("Endpoints:")
	fmt.Println("  POST /intent         - Execute an intent")
	fmt.Println("  GET  /tools          - List available tools")
	fmt.Println("  GET  /status         - Check server status")
	fmt.Println("  GET  /events         - List execution events")
	fmt.Println("  GET  /events/stream  - Stream execution events (SSE)")

	return http.ListenAndServe(port, server.Routes())
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  routerpilot tools")
	fmt.Println("  routerpilot ping <host> [count] [--events]")
	fmt.Println("  routerpilot plan <intent> [args...]")
	fmt.Println("  routerpilot serve")
}
