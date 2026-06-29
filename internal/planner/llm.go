package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/shagston/routerpilot/internal/config"
	"github.com/shagston/routerpilot/internal/registry"
	"github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

type LLMPlanner struct {
	apiKey     string
	endpoint   string
	model      string
	toolSchema string
	registry   *registry.ToolRegistry
}

func NewLLMPlanner(reg *registry.ToolRegistry, cfg *config.Config) *LLMPlanner {
	endpoint := cfg.Planner.Endpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	} else if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = strings.TrimRight(endpoint, "/") + "/chat/completions"
	}

	model := cfg.Planner.Model
	if model == "" {
		model = "gpt-4o"
	}

	return &LLMPlanner{
		apiKey:     cfg.Planner.APIKey,
		endpoint:   endpoint,
		model:      model,
		toolSchema: FormatToolSchemas(reg),
		registry:   reg,
	}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *LLMPlanner) Plan(ctx context.Context, intent planner.Intent, snapshot types.ContextSnapshot) (types.Plan, error) {
	if p.apiKey == "" {
		return types.Plan{}, fmt.Errorf("ROUTERPILOT_API_KEY is not set")
	}

	systemPrompt := fmt.Sprintf(`You are a Network Automation Planner for RouterPilot.
Your goal is to transform user intent into a deterministic execution plan.
If the Current System State contains "execution_error", it means a previous attempt failed. 
Analyze the error and the current state to provide a corrected plan that bypasses the issue or fixes it.

Available Tools:
%s

Respond ONLY with a JSON object matching this structure:
{
  "plan_id": "unique-id",
  "intent": "description of the goal",
  "risk": "low|medium|high",
  "steps": [
    {
      "id": "task-id",
      "tool": "tool.id",
      "purpose": "context|action",
      "arguments": { "arg_name": "value" }
    }
  ]
}

Use purpose "context" for read-only discovery steps that gather live system state.
Use purpose "action" (or omit purpose) for mutating steps.
When you need fresh data before an action, insert context steps first; the runtime will
refresh the plan automatically after those context steps complete.
Do not include any explanations or markdown blocks.`, p.toolSchema)

	snapshotJSON, _ := json.MarshalIndent(snapshot, "", "  ")
	userContent := fmt.Sprintf("Intent: %s\nArguments: %v\n\nCurrent System State:\n%s",
		intent.Name,
		intent.Arguments,
		string(snapshotJSON))

	reqBody := chatRequest{
		Model: p.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return types.Plan{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return types.Plan{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return types.Plan{}, fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return types.Plan{}, fmt.Errorf("LLM API response parse error: %w (endpoint: %s)", err, p.endpoint)
	}

	if len(chatResp.Choices) == 0 {
		return types.Plan{}, fmt.Errorf("LLM returned no choices")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	plan, err := parsePlanResponse(content, intent)
	if err != nil {
		return types.Plan{}, err
	}
	return ValidatePlan(p.registry, plan)
}

func stripMarkdownJSON(content string) string {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "```") {
		return content
	}
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	return strings.TrimSpace(content)
}
