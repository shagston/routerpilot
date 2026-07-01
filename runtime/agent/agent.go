package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/shagston/routerpilot/sdk/types"
)

type State string

const (
	StateInit    State = "init"
	StateStart   State = "start"
	StateReady   State = "ready"
	StateSleep   State = "sleep"
	StateWake    State = "wake"
	StateStop    State = "stop"
	StateCrash   State = "crash"
	StateRecover State = "recover"
)

var validTransitions = map[State][]State{
	StateInit:    {StateStart, StateStop},
	StateStart:   {StateReady, StateCrash},
	StateReady:   {StateSleep, StateStop, StateCrash},
	StateSleep:   {StateWake, StateStop, StateCrash},
	StateWake:    {StateReady, StateCrash},
	StateCrash:   {StateRecover, StateStop},
	StateRecover: {StateStart, StateStop},
	StateStop:    {},
}

func ValidTransition(from, to State) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

type Spec struct {
	Name        string
	Permissions []types.Permission
	Capabilities []types.Capability
	Metadata    map[string]any
}

type Update struct {
	Name        *string
	Permissions []types.Permission
	Capabilities []types.Capability
	Metadata    map[string]any
}

type Info struct {
	ID          types.AgentID      `json:"id"`
	Name        string              `json:"name"`
	State       State               `json:"state"`
	Permissions []types.Permission  `json:"permissions"`
	Capabilities []types.Capability `json:"capabilities"`
	Metadata    map[string]any      `json:"metadata,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type Agent struct {
	info  Info
	state State
}

func New(id types.AgentID, spec Spec) *Agent {
	now := time.Now()
	return &Agent{
		info: Info{
			ID:           id,
			Name:         spec.Name,
			State:        StateInit,
			Permissions:  spec.Permissions,
			Capabilities: spec.Capabilities,
			Metadata:     spec.Metadata,
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		state: StateInit,
	}
}

func (a *Agent) ID() types.AgentID          { return a.info.ID }
func (a *Agent) Info() Info                  { return a.info }
func (a *Agent) State() State                { return a.state }

func (a *Agent) setState(ctx context.Context, to State) error {
	if !ValidTransition(a.state, to) {
		return fmt.Errorf("invalid agent state transition: %s -> %s", a.state, to)
	}
	a.state = to
	a.info.State = to
	a.info.UpdatedAt = time.Now()
	return nil
}

func (a *Agent) Start(ctx context.Context) error {
	if err := a.setState(ctx, StateStart); err != nil {
		return err
	}
	return a.setState(ctx, StateReady)
}

func (a *Agent) Stop(ctx context.Context) error {
	return a.setState(ctx, StateStop)
}

func (a *Agent) Sleep(ctx context.Context) error {
	return a.setState(ctx, StateSleep)
}

func (a *Agent) Wake(ctx context.Context) error {
	if err := a.setState(ctx, StateWake); err != nil {
		return err
	}
	return a.setState(ctx, StateReady)
}

func (a *Agent) Crash(ctx context.Context) error {
	return a.setState(ctx, StateCrash)
}

func (a *Agent) Recover(ctx context.Context) error {
	if err := a.setState(ctx, StateRecover); err != nil {
		return err
	}
	return a.setState(ctx, StateReady)
}

func DefaultSpec() Spec {
	return Spec{
		Name:        "default",
		Permissions: []types.Permission{types.PermissionRead, types.PermissionWrite},
		Capabilities: []types.Capability{"*"},
		Metadata: map[string]any{
			"type": "builtin",
		},
	}
}
