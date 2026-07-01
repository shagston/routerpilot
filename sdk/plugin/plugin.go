package plugin

import (
	"context"

	"github.com/shagston/routerpilot/sdk/capability"
	"github.com/shagston/routerpilot/sdk/events"
	"github.com/shagston/routerpilot/sdk/tool"
)

type Category string

const (
	CategoryCapability Category = "capability"
	CategoryTransport  Category = "transport"
	CategoryMemory     Category = "memory"
	CategoryPolicy     Category = "policy"
	CategoryPlanner    Category = "planner"
	CategoryScheduler  Category = "scheduler"
	CategoryObservability Category = "observability"
	CategoryAuth       Category = "authentication"
	CategoryStorage    Category = "storage"
	CategoryTool       Category = "tool"
)

type Type string

const (
	TypeBuiltin    Type = "builtin"
	TypeExternal   Type = "external"
	TypeSubprocess Type = "subprocess"
)

type Manifest struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	Type         Type     `json:"type"`
	Category     Category `json:"category"`
	SDKVersion   string   `json:"sdk_version"`
	Capabilities []string `json:"capabilities,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
	Description  string   `json:"description,omitempty"`
}

type RuntimeContext interface {
	Logger() interface {
		Debug(string, ...any)
		Info(string, ...any)
		Warn(string, ...any)
		Error(string, ...any)
	}
	Events() events.Publisher
	Capabilities() capability.Registry
	Config(key string) (string, bool)
	InstanceID() string
}

type Plugin interface {
	Manifest() Manifest
	Tool() tool.Tool
	Init(ctx context.Context) error
	Close(ctx context.Context) error
	Initialize(ctx context.Context, rt RuntimeContext) error
	Shutdown(ctx context.Context) error
}

type basePlugin struct {
	manifest Manifest
}

func (b *basePlugin) Manifest() Manifest { return b.manifest }
func (b *basePlugin) Tool() tool.Tool    { return nil }
func (b *basePlugin) Init(ctx context.Context) error { return nil }
func (b *basePlugin) Close(ctx context.Context) error { return nil }
func (b *basePlugin) Initialize(ctx context.Context, rt RuntimeContext) error { return nil }
func (b *basePlugin) Shutdown(ctx context.Context) error { return nil }

func NewBase(manifest Manifest) *basePlugin {
	return &basePlugin{manifest: manifest}
}

type Host interface {
	Load(ctx context.Context, path string) (Plugin, error)
	List() []Plugin
}
