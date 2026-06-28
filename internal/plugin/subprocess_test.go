package plugin

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetManifest(t *testing.T) {
	// Build the dns-plugin if not already built
	srcDir := filepath.Join("..", "..", "examples", "dns-plugin")
	pluginPath := filepath.Join(t.TempDir(), "dns-plugin.exe")

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", pluginPath, srcDir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build plugin: %v", err)
	}

	m, err := getManifest(context.Background(), pluginPath)
	if err != nil {
		t.Fatalf("getManifest failed: %v", err)
	}

	if m.ID != "dns.lookup" {
		t.Fatalf("expected dns.lookup, got %s", m.ID)
	}
	if m.Version != "0.1.0" {
		t.Fatalf("expected 0.1.0, got %s", m.Version)
	}
}

func TestSubprocessPluginLoadAndExecute(t *testing.T) {
	srcDir := filepath.Join("..", "..", "examples", "dns-plugin")
	pluginPath := filepath.Join(t.TempDir(), "dns-plugin.exe")

	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", pluginPath, srcDir)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("build plugin: %v", err)
	}

	plug, err := loadSubprocess(context.Background(), pluginPath)
	if err != nil {
		t.Fatalf("loadSubprocess failed: %v", err)
	}

	manifest := plug.Manifest()
	if manifest.ID != "dns.lookup" {
		t.Fatalf("expected dns.lookup, got %s", manifest.ID)
	}

	tool := plug.Tool()
	meta := tool.Metadata()
	if meta.Version != "0.1.0" {
		t.Fatalf("expected 0.1.0, got %s", meta.Version)
	}
}
