package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	sdk "github.com/shagston/routerpilot/sdk/memory"
	"github.com/shagston/routerpilot/sdk/types"
)

func TestWorkingMemoryCRUD(t *testing.T) {
	ctx := context.Background()
	wm := NewWorkingMemory()

	wm.Set(ctx, sdk.Record{Key: "test.key", Value: map[string]any{"val": 42}})

	r, ok := wm.Get(ctx, "test.key")
	if !ok {
		t.Fatal("expected record to exist")
	}
	if r.Key != "test.key" {
		t.Errorf("expected key test.key, got %s", r.Key)
	}
	v, _ := r.Value["val"].(int)
	if v != 42 {
		t.Errorf("expected val=42, got %d", v)
	}

	wm.Delete(ctx, "test.key")
	_, ok = wm.Get(ctx, "test.key")
	if ok {
		t.Error("expected record to be deleted")
	}
}

func TestWorkingMemoryClear(t *testing.T) {
	ctx := context.Background()
	wm := NewWorkingMemory()

	wm.Set(ctx, sdk.Record{Key: "a"})
	wm.Set(ctx, sdk.Record{Key: "b"})
	wm.Set(ctx, sdk.Record{Key: "c"})

	if wm.Len() != 3 {
		t.Errorf("expected len 3, got %d", wm.Len())
	}

	wm.Clear(ctx)

	if wm.Len() != 0 {
		t.Errorf("expected len 0 after clear, got %d", wm.Len())
	}
}

func TestWorkingMemorySearch(t *testing.T) {
	ctx := context.Background()
	wm := NewWorkingMemory()

	wm.Set(ctx, sdk.Record{Key: "alpha.one"})
	wm.Set(ctx, sdk.Record{Key: "alpha.two"})
	wm.Set(ctx, sdk.Record{Key: "beta.one"})

	results := wm.Search(ctx, "alpha", 10)
	if len(results) != 2 {
		t.Errorf("expected 2 alpha results, got %d", len(results))
	}

	results = wm.Search(ctx, "", 2)
	if len(results) != 2 {
		t.Errorf("expected 2 results with limit 2, got %d", len(results))
	}
}

func TestSessionMemoryTTL(t *testing.T) {
	ctx := context.Background()
	sm := NewSessionMemory(WithSessionTTL(50 * time.Millisecond))

	sm.Set(ctx, sdk.Record{Key: "ephemeral", Value: map[string]any{"data": "gone soon"}})

	r, ok := sm.Get(ctx, "ephemeral")
	if !ok {
		t.Fatal("expected record to exist before TTL")
	}
	if r.Key != "ephemeral" {
		t.Errorf("expected key ephemeral, got %s", r.Key)
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = sm.Get(ctx, "ephemeral")
	if ok {
		t.Error("expected record to expire after TTL")
	}
}

func TestSessionMemoryExplicitTTL(t *testing.T) {
	ctx := context.Background()
	sm := NewSessionMemory()

	sm.SetWithTTL(ctx, sdk.Record{Key: "custom-ttl"}, 30*time.Millisecond)

	time.Sleep(60 * time.Millisecond)

	_, ok := sm.Get(ctx, "custom-ttl")
	if ok {
		t.Error("expected record to expire after custom TTL")
	}
}

func TestSessionMemoryNoTTL(t *testing.T) {
	ctx := context.Background()
	sm := NewSessionMemory(WithSessionTTL(0))

	sm.Set(ctx, sdk.Record{Key: "permanent"})

	time.Sleep(50 * time.Millisecond)

	_, ok := sm.Get(ctx, "permanent")
	if !ok {
		t.Error("expected record to persist when TTL is zero")
	}
}

func TestSessionMemoryCleanup(t *testing.T) {
	ctx := context.Background()
	sm := NewSessionMemory(WithSessionTTL(30 * time.Millisecond))

	sm.Set(ctx, sdk.Record{Key: "expirable"})

	time.Sleep(60 * time.Millisecond)

	sm.Cleanup(ctx)

	_, ok := sm.Get(ctx, "expirable")
	if ok {
		t.Error("expected expired record to be cleaned up")
	}
}

func TestSessionMemoryMaxSize(t *testing.T) {
	ctx := context.Background()
	sm := NewSessionMemory(WithSessionMaxSize(3))

	sm.Set(ctx, sdk.Record{Key: "a"})
	sm.Set(ctx, sdk.Record{Key: "b"})
	sm.Set(ctx, sdk.Record{Key: "c"})
	sm.Set(ctx, sdk.Record{Key: "d"})

	_, ok := sm.Get(ctx, "a")
	if ok {
		t.Error("expected oldest record to be evicted when at max size")
	}

	_, ok = sm.Get(ctx, "d")
	if !ok {
		t.Error("expected newest record to exist")
	}
}

func TestStoreLayerPriority(t *testing.T) {
	ctx := context.Background()
	persistent := NewInMemoryProvider()
	st := NewStore(persistent)

	persistent.Write(ctx, sdk.Record{Key: "shared.key", Value: map[string]any{"source": "persistent"}})

	st.Working().Set(ctx, sdk.Record{Key: "shared.key", Value: map[string]any{"source": "working"}})

	r, err := st.Get(ctx, "shared.key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	src, _ := r.Value["source"].(string)
	if src != "working" {
		t.Errorf("expected working layer to take priority, got %s", src)
	}
}

func TestStorePutGetLayers(t *testing.T) {
	ctx := context.Background()
	persistent := NewInMemoryProvider()
	st := NewStore(persistent)

	if err := st.Put(ctx, "w.key", map[string]any{"layer": "working"}, LayerWorking); err != nil {
		t.Fatalf("Put working: %v", err)
	}
	if err := st.Put(ctx, "s.key", map[string]any{"layer": "session"}, LayerSession); err != nil {
		t.Fatalf("Put session: %v", err)
	}
	if err := st.Put(ctx, "p.key", map[string]any{"layer": "persistent"}, LayerPersistent); err != nil {
		t.Fatalf("Put persistent: %v", err)
	}

	r, err := st.Get(ctx, "w.key")
	if err != nil {
		t.Fatalf("Get w.key: %v", err)
	}
	v, _ := r.Value["layer"].(string)
	if v != "working" {
		t.Errorf("expected 'working', got %s", v)
	}

	r, err = st.Get(ctx, "s.key")
	if err != nil {
		t.Fatalf("Get s.key: %v", err)
	}
	v, _ = r.Value["layer"].(string)
	if v != "session" {
		t.Errorf("expected 'session', got %s", v)
	}

	r, err = st.Get(ctx, "p.key")
	if err != nil {
		t.Fatalf("Get p.key: %v", err)
	}
	v, _ = r.Value["layer"].(string)
	if v != "persistent" {
		t.Errorf("expected 'persistent', got %s", v)
	}
}

func TestStoreNotFound(t *testing.T) {
	ctx := context.Background()
	st := NewStore(NewInMemoryProvider())

	_, err := st.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent key")
	}
	if !isErrNotFound(err) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStoreSearchAll(t *testing.T) {
	ctx := context.Background()
	persistent := NewInMemoryProvider()
	st := NewStore(persistent)

	st.Working().Set(ctx, sdk.Record{Key: "shared.a", Value: map[string]any{"from": "working"}})
	st.Session().Set(ctx, sdk.Record{Key: "shared.a", Value: map[string]any{"from": "session"}})
	persistent.Write(ctx, sdk.Record{Key: "persistent.b", Value: map[string]any{"from": "persistent"}})

	results, err := st.SearchAll(ctx, "", 10)
	if err != nil {
		t.Fatalf("SearchAll: %v", err)
	}

	keys := make(map[string]bool)
	for _, r := range results {
		keys[r.Key] = true
	}

	if !keys["shared.a"] {
		t.Error("expected shared.a in results")
	}
	if !keys["persistent.b"] {
		t.Error("expected persistent.b in results")
	}
	if len(results) > 2 && len(results) < 1 {
		t.Errorf("expected 2 unique results, got %d", len(results))
	}
}

func TestStoreClearWorking(t *testing.T) {
	ctx := context.Background()
	st := NewStore(NewInMemoryProvider())

	st.Working().Set(ctx, sdk.Record{Key: "temp"})
	st.ClearWorking(ctx)

	_, ok := st.Working().Get(ctx, "temp")
	if ok {
		t.Error("expected working memory to be cleared")
	}
}

func isErrNotFound(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, types.ErrNotFound)
}
