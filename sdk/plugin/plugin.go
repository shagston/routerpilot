package plugin

import (
	"context"

	"github.com/shagston/routerpilot/sdk/tool"
)

type Manifest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    Type   `json:"type"`
}

type Type string

const (
	TypeBuiltin    Type = "builtin"
	TypeExternal   Type = "external"
	TypeSubprocess Type = "subprocess"
)

type Plugin interface {
	Manifest() Manifest
	Tool() tool.Tool
	Init(context.Context) error
	Close(context.Context) error
}

type Host interface {
	Load(ctx context.Context, path string) (Plugin, error)
	List() []Plugin
}
