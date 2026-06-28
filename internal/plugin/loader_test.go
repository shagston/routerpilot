package plugin

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/shagston/routerpilot/internal/registry"
)

func writePlugin(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoaderSkipsNonExistentDir(t *testing.T) {
	loader := NewLoader(filepath.Join(os.TempDir(), "does-not-exist-12345"))
	reg := registry.NewToolRegistry()
	if err := loader.LoadAll(context.Background(), reg); err != nil {
		t.Fatalf("expected nil for missing dir, got: %v", err)
	}
}

func TestLoaderSkipsInvalidBinary(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("test uses .exe extension")
	}

	dir := t.TempDir()
	writePlugin(t, dir, "not-a-plugin.exe", []byte("not a real binary"))

	loader := NewLoader(dir)
	reg := registry.NewToolRegistry()
	if err := loader.LoadAll(context.Background(), reg); err != nil {
		t.Fatalf("expected nil, got: %v", err)
	}

	if len(reg.List()) > 0 {
		t.Fatal("expected no tools from invalid binary")
	}
}
