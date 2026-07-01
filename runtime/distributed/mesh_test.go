package distributed

import (
	"context"
	"testing"
	"time"

	"github.com/shagston/routerpilot/internal/events"
	"github.com/shagston/routerpilot/sdk/capability"
	runtimeTp "github.com/shagston/routerpilot/runtime/transport"
	reticulum "github.com/shagston/routerpilot/transports/reticulum"
	sdkTp "github.com/shagston/routerpilot/sdk/transport"
)

func TestNewMesh(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	trans := runtimeTp.NewMemory("test-node")

	m := NewMesh("test-node", trans, pub)
	if m == nil {
		t.Fatal("expected non-nil mesh")
	}
	if m.localID != "test-node" {
		t.Errorf("expected localID test-node, got %s", m.localID)
	}
}

func TestMeshStartStop(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	trans := runtimeTp.NewMemory("test-mesh")
	ctx := context.Background()

	m := NewMesh("test-mesh", trans, pub)
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := m.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestMeshPeerDiscovery(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	ctx := context.Background()

	a := runtimeTp.NewMemory("node-a")
	b := runtimeTp.NewMemory("node-b")

	meshA := NewMesh("node-a", a, pub)
	meshB := NewMesh("node-b", b, pub)

	if err := meshA.Start(ctx); err != nil {
		t.Fatalf("Start meshA: %v", err)
	}
	defer meshA.Stop(ctx)

	if err := meshB.Start(ctx); err != nil {
		t.Fatalf("Start meshB: %v", err)
	}
	defer meshB.Stop(ctx)

	meshA.DiscoverPeer(ctx, sdkTp.Info{
		ID:   "node-b",
		Name: "Node B",
		Type: "memory",
		Metadata: map[string]string{
			"version": "0.1.0",
		},
	})

	peers := meshA.Peers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].RuntimeID != "node-b" {
		t.Errorf("expected RuntimeID node-b, got %s", peers[0].RuntimeID)
	}
	if peers[0].State != PeerDiscovered {
		t.Errorf("expected state discovered, got %s", peers[0].State)
	}
}

func TestMeshLookup(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	trans := runtimeTp.NewMemory("lookup-node")
	ctx := context.Background()

	m := NewMesh("lookup-node", trans, pub)
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer m.Stop(ctx)

	m.DiscoverPeer(ctx, sdkTp.Info{
		ID:   "peer-1",
		Name: "Peer One",
		Type: "memory",
	})

	peer, ok := m.Lookup("peer-1")
	if !ok {
		t.Fatal("expected peer to be found")
	}
	if peer.Name != "Peer One" {
		t.Errorf("expected name 'Peer One', got %s", peer.Name)
	}

	_, ok = m.Lookup("nonexistent")
	if ok {
		t.Error("expected nonexistent peer to not be found")
	}
}

func TestMeshRemovePeer(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	trans := runtimeTp.NewMemory("remove-node")
	ctx := context.Background()

	m := NewMesh("remove-node", trans, pub)
	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer m.Stop(ctx)

	m.DiscoverPeer(ctx, sdkTp.Info{
		ID:   "peer-to-remove",
		Name: "Removable",
		Type: "memory",
	})

	if len(m.Peers()) != 1 {
		t.Fatal("expected 1 peer before removal")
	}

	m.RemovePeer(ctx, "peer-to-remove")

	if len(m.Peers()) != 0 {
		t.Errorf("expected 0 peers after removal, got %d", len(m.Peers()))
	}
}

type testProvider struct {
	id          capability.ID
	name        string
	executeFunc func(ctx context.Context, input map[string]any) (map[string]any, error)
}

func (p *testProvider) ID() capability.ID                                { return p.id }
func (p *testProvider) Info() capability.Info {
	return capability.Info{ID: p.id, Name: p.name, Timeout: 5000}
}
func (p *testProvider) Validate(ctx context.Context, input map[string]any) error { return nil }
func (p *testProvider) Execute(ctx context.Context, input map[string]any) (map[string]any, error) {
	if p.executeFunc != nil {
		return p.executeFunc(ctx, input)
	}
	return map[string]any{"result": "ok"}, nil
}

func TestMeshRemoteExecution(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	ctx := context.Background()

	transA := reticulum.New("node-a", "0.1.0")
	transB := reticulum.New("node-b", "0.1.0")

	if err := transA.Start(ctx); err != nil {
		t.Fatalf("Start transA: %v", err)
	}
	defer transA.Stop(ctx)

	if err := transB.Start(ctx); err != nil {
		t.Fatalf("Start transB: %v", err)
	}
	defer transB.Stop(ctx)

	meshA := NewMesh("node-a", transA, pub)
	meshB := NewMesh("node-b", transB, pub)

	if err := meshA.Start(ctx); err != nil {
		t.Fatalf("Start meshA: %v", err)
	}
	defer meshA.Stop(ctx)

	if err := meshB.Start(ctx); err != nil {
		t.Fatalf("Start meshB: %v", err)
	}
	defer meshB.Stop(ctx)

	meshA.DiscoverPeer(ctx, sdkTp.Info{
		ID:   transB.ID(),
		Name: "node-b",
		Type: "reticulum",
	})

	provider := &testProvider{
		id:   "test.ping",
		name: "Test Ping",
	}

	executed := false
	meshB.RegisterCapabilityHandler("test.ping", func(ctx context.Context, source, providerID string, input map[string]any) (map[string]any, error) {
		executed = true
		return map[string]any{"pong": true, "echo": input["message"]}, nil
	})

	result, err := meshA.ExecuteRemotely(ctx, transB.ID(), provider, map[string]any{"message": "hello"})
	if err != nil {
		t.Fatalf("ExecuteRemotely: %v", err)
	}

	if !executed {
		t.Error("expected remote handler to be executed")
	}

	pong, ok := result["pong"]
	if !ok || pong != true {
		t.Error("expected pong=true in result")
	}
	echo, ok := result["echo"]
	if !ok || echo != "hello" {
		t.Errorf("expected echo=hello, got %v", echo)
	}
}

func TestMeshPeerLifecycle(t *testing.T) {
	pub := events.NewBus()
	defer pub.Close()
	trans := runtimeTp.NewMemory("lifecycle-node")
	ctx := context.Background()

	m := NewMesh("lifecycle-node", trans, pub,
		WithPeerTimeout(200*time.Millisecond),
		WithHeartbeatInterval(50*time.Millisecond),
	)

	if err := m.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer m.Stop(ctx)

	m.DiscoverPeer(ctx, sdkTp.Info{ID: "peer-live", Name: "Live Peer", Type: "memory"})

	time.Sleep(300 * time.Millisecond)

	m.mu.RLock()
	p, ok := m.peers["peer-live"]
	m.mu.RUnlock()

	if !ok {
		t.Fatal("expected peer to still exist")
	}
	if p.State != PeerUnavailable {
		t.Errorf("expected peer to be unavailable after timeout, got %s", p.State)
	}
}
