package planner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shagston/routerpilot/internal/config"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
	networktools "github.com/shagston/routerpilot/tools/network"
)

func TestLLMPlannerPlanParsesAndValidatesJSONResponse(t *testing.T) {
	reg := registry.NewToolRegistry()
	if err := reg.Register(networktools.PingTool{}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header: %q", r.Header.Get("Authorization"))
		}

		response := chatResponse{}
		response.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: `{"plan_id":"plan-test","intent":"Ping host","risk":"low","steps":[{"id":"task-1","tool":"network.ping","arguments":{"host":"127.0.0.1","count":1}}]}`,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := config.Default()
	p := NewLLMPlanner(reg, cfg)
	p.apiKey = "test-key"
	p.endpoint = server.URL

	plan, err := p.Plan(context.Background(), planner.Intent{
		Name: "ping",
		Arguments: map[string]any{
			"target": "127.0.0.1",
		},
	}, types.ContextSnapshot{})
	if err != nil {
		t.Fatalf("Plan() error: %v", err)
	}

	if plan.ID != "plan-test" {
		t.Fatalf("expected plan-test, got %q", plan.ID)
	}
	if len(plan.Steps) != 1 || plan.Steps[0].Tool != "network.ping" {
		t.Fatalf("unexpected steps: %+v", plan.Steps)
	}
}

func TestLLMPlannerRejectsInvalidPlanFromModel(t *testing.T) {
	reg := registry.NewToolRegistry()
	if err := reg.Register(networktools.PingTool{}); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := chatResponse{}
		response.Choices = []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: `{"plan_id":"plan-bad","intent":"Ping host","risk":"low","steps":[{"id":"task-1","tool":"network.unknown","arguments":{"host":"127.0.0.1"}}]}`,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg2 := config.Default()
	p := NewLLMPlanner(reg, cfg2)
	p.apiKey = "test-key"
	p.endpoint = server.URL

	_, err := p.Plan(context.Background(), planner.Intent{Name: "ping"}, types.ContextSnapshot{})
	if err == nil {
		t.Fatal("expected validation error for unknown tool")
	}
}
