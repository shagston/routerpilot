package reticulum

import (
	"context"
	"testing"
	"time"

	sdk "github.com/shagston/routerpilot/sdk/transport"
)

func TestNewReticulumTransport(t *testing.T) {
	tp := New("test-node", "0.1.0")
	if tp == nil {
		t.Fatal("expected non-nil transport")
	}
	if tp.identity.Name != "test-node" {
		t.Errorf("expected name test-node, got %s", tp.identity.Name)
	}
	if tp.identity.Version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", tp.identity.Version)
	}
	if tp.identity.ID == "" {
		t.Error("expected non-empty identity ID")
	}
}

func TestReticulumStartStop(t *testing.T) {
	tp := New("start-stop", "0.1.0")
	ctx := context.Background()

	if err := tp.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := tp.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestReticulumSendReceive(t *testing.T) {
	ctx := context.Background()
	a := New("node-a", "0.1.0")
	b := New("node-b", "0.1.0")

	if err := a.Start(ctx); err != nil {
		t.Fatalf("Start a: %v", err)
	}
	defer a.Stop(ctx)

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start b: %v", err)
	}
	defer b.Stop(ctx)

	if err := a.DiscoverPeers(ctx); err != nil {
		t.Fatalf("DiscoverPeers: %v", err)
	}

	bInbox, err := b.Endpoint().Receive(ctx)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}

	env := sdk.Envelope{
		Source:  a.ID(),
		Target:  b.ID(),
		Type:    "test.message",
		Payload: []byte("hello from a"),
	}

	if err := a.Endpoint().Send(ctx, env); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case received := <-bInbox:
		if string(received.Payload) != "hello from a" {
			t.Errorf("expected 'hello from a', got %s", string(received.Payload))
		}
		if received.Type != "test.message" {
			t.Errorf("expected type 'test.message', got %s", received.Type)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestReticulumRouting(t *testing.T) {
	ctx := context.Background()
	a := New("router-a", "0.1.0")
	b := New("router-b", "0.1.0")

	if err := a.Start(ctx); err != nil {
		t.Fatalf("Start a: %v", err)
	}
	defer a.Stop(ctx)

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start b: %v", err)
	}
	defer b.Stop(ctx)

	if err := a.AddRoute("service.*", b.ID()); err != nil {
		t.Fatalf("AddRoute: %v", err)
	}

	env := sdk.Envelope{Target: "service.foo"}
	target, err := a.Route(ctx, env)
	if err != nil {
		t.Fatalf("Route: %v", err)
	}
	if target != b.ID() {
		t.Errorf("expected route to %s, got %s", b.ID(), target)
	}

	if err := a.RemoveRoute("service.*"); err != nil {
		t.Fatalf("RemoveRoute: %v", err)
	}

	_, err = a.Route(ctx, env)
	if err == nil {
		t.Error("expected error after route removal")
	}
}

func TestReticulumDiscovery(t *testing.T) {
	ctx := context.Background()
	a := New("disc-a", "0.1.0")
	b := New("disc-b", "0.1.0")

	if err := a.Start(ctx); err != nil {
		t.Fatalf("Start a: %v", err)
	}
	defer a.Stop(ctx)

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start b: %v", err)
	}
	defer b.Stop(ctx)

	infoA := sdk.Info{
		ID:   a.ID(),
		Name: "service-alpha",
		Type: "alpha",
	}
	if err := a.Register(ctx, infoA); err != nil {
		t.Fatalf("Register: %v", err)
	}

	results, err := a.Lookup(ctx, "service-alpha")
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != a.ID() {
		t.Errorf("expected ID %s, got %s", a.ID(), results[0].ID)
	}

	if err := a.Unregister(ctx, a.ID()); err != nil {
		t.Fatalf("Unregister: %v", err)
	}

	_, err = a.Lookup(ctx, "service-alpha")
	if err == nil {
		t.Error("expected error after unregister")
	}
}

func TestReticulumPeerDiscovery(t *testing.T) {
	ctx := context.Background()
	a := New("peer-a", "0.1.0")
	b := New("peer-b", "0.1.0")

	if err := a.Start(ctx); err != nil {
		t.Fatalf("Start a: %v", err)
	}
	defer a.Stop(ctx)

	if err := b.Start(ctx); err != nil {
		t.Fatalf("Start b: %v", err)
	}
	defer b.Stop(ctx)

	if err := a.DiscoverPeers(ctx); err != nil {
		t.Fatalf("DiscoverPeers: %v", err)
	}

	peers := a.Peers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 peer, got %d", len(peers))
	}
	if peers[0].Identity.ID != b.ID() {
		t.Errorf("expected peer ID %s, got %s", b.ID(), peers[0].Identity.ID)
	}
}

func TestReticulumLocalIdentity(t *testing.T) {
	id := GenerateIdentity("custom-node", "1.0.0")
	tp := New("identity-test", "0.1.0", WithIdentity(id))

	localID := tp.LocalIdentity()
	if localID.ID != id.ID {
		t.Errorf("expected ID %s, got %s", id.ID, localID.ID)
	}
	if localID.Name != "custom-node" {
		t.Errorf("expected name custom-node, got %s", localID.Name)
	}
}

func TestReticulumSendToSelf(t *testing.T) {
	ctx := context.Background()
	tp := New("self-node", "0.1.0")

	if err := tp.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer tp.Stop(ctx)

	inbox, err := tp.Receive(ctx)
	if err != nil {
		t.Fatalf("Receive: %v", err)
	}

	env := sdk.Envelope{
		Target:  tp.ID(),
		Type:    "self.test",
		Payload: []byte("loopback"),
	}

	if err := tp.Send(ctx, env); err != nil {
		t.Fatalf("Send to self: %v", err)
	}

	select {
	case received := <-inbox:
		if string(received.Payload) != "loopback" {
			t.Errorf("expected 'loopback', got %s", string(received.Payload))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for self message")
	}
}

func TestReticulumName(t *testing.T) {
	tp := New("named-node", "0.1.0")
	if tp.Name() != "reticulum:named-node" {
		t.Errorf("expected 'reticulum:named-node', got %s", tp.Name())
	}
}
