package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"

	"github.com/shagston/routerpilot/sdk/plugin"
	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/types"
)

type subprocessPlugin struct {
	path     string
	manifest plugin.Manifest
	mu       sync.Mutex
}

func loadSubprocess(ctx context.Context, path string) (plugin.Plugin, error) {
	manifest, err := getManifest(ctx, path)
	if err != nil {
		return nil, err
	}

	return &subprocessPlugin{
		path:     path,
		manifest: manifest,
	}, nil
}

func (p *subprocessPlugin) Manifest() plugin.Manifest {
	return p.manifest
}

func (p *subprocessPlugin) Tool() tool.Tool {
	return &subprocessTool{plugin: p}
}

func (p *subprocessPlugin) Init(ctx context.Context) error {
	return nil
}

func (p *subprocessPlugin) Close(ctx context.Context) error {
	return nil
}

type subprocessTool struct {
	plugin *subprocessPlugin
}

func (t *subprocessTool) Metadata() types.ToolMetadata {
	return types.ToolMetadata{
		ID:          types.ToolID(t.plugin.manifest.ID),
		Version:     t.plugin.manifest.Version,
		Description: t.plugin.manifest.Name,
	}
}

func (t *subprocessTool) InputSchema() types.Schema {
	return types.Schema{}
}

func (t *subprocessTool) OutputSchema() types.Schema {
	return types.Schema{}
}

func (t *subprocessTool) Validate(ctx context.Context, input types.ToolInput) error {
	return nil
}

func (t *subprocessTool) Execute(ctx context.Context, input types.ToolInput) (types.ToolResult, error) {
	t.plugin.mu.Lock()
	defer t.plugin.mu.Unlock()

	payload, err := json.Marshal(input)
	if err != nil {
		return types.ToolResult{Success: false, Error: err.Error()}, err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.CommandContext(ctx, t.plugin.path, "execute")
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return types.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("plugin failed: %v, stderr: %s", err, stderr.String()),
		}, err
	}

	var result types.ToolResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return types.ToolResult{Success: false, Error: fmt.Sprintf("invalid plugin output: %v", err)}, err
	}
	return result, nil
}

func getManifest(ctx context.Context, path string) (plugin.Manifest, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, path, "plugin-manifest")
	cmd.Stdout = &stdout
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return plugin.Manifest{}, fmt.Errorf("plugin manifest failed: %w", err)
	}

	var m plugin.Manifest
	if err := json.Unmarshal(stdout.Bytes(), &m); err != nil {
		return plugin.Manifest{}, fmt.Errorf("invalid manifest: %w", err)
	}
	return m, nil
}
