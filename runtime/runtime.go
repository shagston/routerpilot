package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/shagston/routerpilot/runtime/agent"
	capreg "github.com/shagston/routerpilot/runtime/capability"
	"github.com/shagston/routerpilot/runtime/distributed"
	"github.com/shagston/routerpilot/runtime/engine"
	"github.com/shagston/routerpilot/runtime/policy"
	"github.com/shagston/routerpilot/runtime/scheduler"
	runtimeTp "github.com/shagston/routerpilot/runtime/transport"
	"github.com/shagston/routerpilot/sdk/capability"
	"github.com/shagston/routerpilot/sdk/events"
	"github.com/shagston/routerpilot/sdk/tool"
	"github.com/shagston/routerpilot/sdk/transport"
	"github.com/shagston/routerpilot/sdk/types"
)

type chanSubscriber interface {
	SubscribeChan(int) (<-chan types.Event, func())
}

type Runtime struct {
	engine       *engine.Engine
	registry     tool.Registry
	capabilities *capreg.Registry
	pub          events.Publisher
	agents       *agent.Manager
	policies     *policy.Engine
	transport    transport.Transport
	defaultAgent *agent.Agent
	mesh         *distributed.Mesh
	sched        *scheduler.Scheduler
	schedCancel  func()
}

func New(reg tool.Registry, pub events.Publisher, opts ...engine.Option) *Runtime {
	capReg := capreg.NewRegistry()
	for _, meta := range reg.List() {
		t, err := reg.Get(meta.ID)
		if err != nil {
			continue
		}
		if err := capReg.RegisterFromTool(t); err != nil {
			slog.Warn("failed to register capability", "tool", meta.ID, "error", err)
		}
	}

	policyEngine := policy.NewEngine()
	for _, p := range policy.DefaultPolicies() {
		policyEngine.AddPolicy(p)
	}

	memTp := runtimeTp.NewMemory("default")

	rt := &Runtime{
		engine:       engine.NewEngine(reg, pub, opts...),
		registry:     reg,
		pub:          pub,
		agents:       agent.NewManager(),
		capabilities: capReg,
		policies:     policyEngine,
		transport:    memTp,
	}

	ctx := context.Background()
	defaultAgent, err := rt.agents.Create(ctx, agent.DefaultSpec())
	if err == nil {
		defaultAgent.Start(ctx)
		rt.defaultAgent = defaultAgent
	}

	rt.sched = scheduler.New(func(ctx context.Context, sched scheduler.Schedule) error {
		_, err := rt.ExecuteCapability(ctx, capability.ID(sched.Capability), sched.Input)
		return err
	}, scheduler.WithEventPublisher(pub))

	if sub, ok := pub.(chanSubscriber); ok {
		ch, cancel := sub.SubscribeChan(256)
		rt.schedCancel = cancel
		go func() {
			for evt := range ch {
				rt.sched.OnEvent(context.Background(), evt)
			}
		}()
	}

	slog.Info("runtime initialized",
		slog.Int("tools", len(reg.List())),
		slog.Int("capabilities", len(capReg.List())),
		slog.Int("policies", len(policyEngine.Policies())),
		slog.String("transport", "memory"),
	)
	return rt
}

func (r *Runtime) Transport() transport.Transport {
	return r.transport
}

func (r *Runtime) SetTransport(ctx context.Context, t transport.Transport) error {
	if r.mesh != nil {
		return fmt.Errorf("cannot change transport while mesh is active")
	}
	r.transport = t
	return nil
}

func (r *Runtime) Mesh() *distributed.Mesh {
	return r.mesh
}

func (r *Runtime) StartMesh(ctx context.Context) error {
	if r.mesh != nil {
		return nil
	}

	r.mesh = distributed.NewMesh("runtime-1", r.transport, r.pub)
	r.transport.Start(ctx)
	if err := r.mesh.Start(ctx); err != nil {
		return fmt.Errorf("start mesh: %w", err)
	}

	slog.Info("mesh started", "transport_id", r.transport.Endpoint().ID())
	return nil
}

func (r *Runtime) StopMesh(ctx context.Context) error {
	if r.mesh == nil {
		return nil
	}
	r.mesh.Stop(ctx)
	r.transport.Stop(ctx)
	r.mesh = nil
	slog.Info("mesh stopped")
	return nil
}

func (r *Runtime) Peers() []distributed.Peer {
	if r.mesh == nil {
		return nil
	}
	return r.mesh.Peers()
}

func (r *Runtime) MeshLookup(runtimeID string) (distributed.Peer, bool) {
	if r.mesh == nil {
		return distributed.Peer{}, false
	}
	return r.mesh.Lookup(runtimeID)
}

func (r *Runtime) ExecuteRemotely(ctx context.Context, targetID string, provider capability.Provider, input map[string]any) (map[string]any, error) {
	if r.mesh == nil {
		return nil, fmt.Errorf("mesh not started")
	}
	return r.mesh.ExecuteRemotely(ctx, targetID, provider, input)
}

func (r *Runtime) RegisterCapabilityHandler(capID string, handler distributed.CapabilityHandler) {
	if r.mesh == nil {
		return
	}
	r.mesh.RegisterCapabilityHandler(capID, handler)
}

func (r *Runtime) Execute(ctx context.Context, plan types.Plan, snapshot types.ContextSnapshot) (types.Execution, error) {
	return r.engine.Execute(ctx, plan, snapshot)
}

func (r *Runtime) ExecuteCapability(ctx context.Context, id capability.ID, input map[string]any) (map[string]any, error) {
	provider, err := r.capabilities.Get(id)
	if err != nil {
		return nil, fmt.Errorf("resolve capability %s: %w", id, err)
	}

	if err := provider.Validate(ctx, input); err != nil {
		return nil, fmt.Errorf("validate capability %s: %w", id, err)
	}

	info := provider.Info()
	agentID := types.AgentID("")
	agentPerms := []types.Permission{types.PermissionRead, types.PermissionWrite}
	if r.defaultAgent != nil {
		agentID = r.defaultAgent.ID()
		agentPerms = r.defaultAgent.Info().Permissions
	}

	result := r.policies.Evaluate(ctx, policy.Request{
		Capability:  string(id),
		Risk:        info.Risk,
		Permissions: info.Permissions,
		AgentID:     agentID,
		AgentPerms:  agentPerms,
	})

	if !result.Allowed {
		return nil, fmt.Errorf("policy denied: %s", result.Reason)
	}

	return provider.Execute(ctx, input)
}

func (r *Runtime) Registry() tool.Registry {
	return r.registry
}

func (r *Runtime) Events() events.Publisher {
	return r.pub
}

func (r *Runtime) Engine() *engine.Engine {
	return r.engine
}

func (r *Runtime) Agents() *agent.Manager {
	return r.agents
}

func (r *Runtime) DefaultAgent() *agent.Agent {
	return r.defaultAgent
}

func (r *Runtime) Capabilities() *capreg.Registry {
	return r.capabilities
}

func (r *Runtime) Policies() *policy.Engine {
	return r.policies
}

func (r *Runtime) LoadPolicyFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve policy path: %w", err)
	}
	return r.policies.LoadFile(absPath)
}

func (r *Runtime) Scheduler() *scheduler.Scheduler {
	return r.sched
}

func (r *Runtime) Start(ctx context.Context) error {
	if err := r.transport.Start(ctx); err != nil {
		return fmt.Errorf("start transport: %w", err)
	}
	if err := r.sched.Start(ctx); err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	slog.Info("runtime started")
	return nil
}

func (r *Runtime) Stop(ctx context.Context) error {
	if r.schedCancel != nil {
		r.schedCancel()
	}

	if err := r.sched.Stop(ctx); err != nil {
		slog.Error("stop scheduler", "error", err)
	}

	if r.mesh != nil {
		r.StopMesh(ctx)
	}

	for _, info := range r.agents.List() {
		if a, err := r.agents.Get(info.ID); err == nil {
			a.Stop(ctx)
		}
	}

	r.transport.Stop(ctx)
	slog.Info("runtime stopped")
	return nil
}
