package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.Port != ":8080" {
		t.Fatalf("expected :8080, got %s", cfg.Server.Port)
	}
	if cfg.Planner.Type != "simple" {
		t.Fatalf("expected simple, got %s", cfg.Planner.Type)
	}
	if cfg.Logging.Level != "info" {
		t.Fatalf("expected info, got %s", cfg.Logging.Level)
	}
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != ":8080" {
		t.Fatalf("expected :8080, got %s", cfg.Server.Port)
	}
}

func TestLoad_File(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "routerpilot.json")
	if err := os.WriteFile(cfgPath, []byte(`{"server":{"port":":9999"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != ":9999" {
		t.Fatalf("expected :9999, got %s", cfg.Server.Port)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "routerpilot.json")
	if err := os.WriteFile(cfgPath, []byte(`{"server":{"port":":9999"},"logging":{"level":"debug"}}`), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("ROUTERPILOT_PORT", ":7777")

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != ":7777" {
		t.Fatalf("expected :7777 (env override), got %s", cfg.Server.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected debug (from file), got %s", cfg.Logging.Level)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestDefaultDryRun(t *testing.T) {
	cfg := Default()
	if cfg.Security.DryRun {
		t.Fatal("expected dry_run to default to false")
	}
	if cfg.Security.ReadOnly {
		t.Fatal("expected read_only to default to false")
	}
}

func TestLoadReadOnly(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "routerpilot.json")
	if err := os.WriteFile(cfgPath, []byte(`{"security":{"read_only":true}}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Security.ReadOnly {
		t.Fatal("expected read_only to be true")
	}
}

func TestLoadDryRun(t *testing.T) {
	tmp := t.TempDir()
	cfgPath := filepath.Join(tmp, "routerpilot.json")
	if err := os.WriteFile(cfgPath, []byte(`{"security":{"dry_run":true}}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Security.DryRun {
		t.Fatal("expected dry_run to be true")
	}
}

func TestLoadDryRunEnvOverride(t *testing.T) {
	t.Setenv("ROUTERPILOT_DRY_RUN", "true")
	t.Setenv("ROUTERPILOT_READ_ONLY", "true")

	cfg := Default()
	cfg.overrideFromEnv()

	if !cfg.Security.DryRun {
		t.Fatal("expected dry_run to be true from env")
	}
	if !cfg.Security.ReadOnly {
		t.Fatal("expected read_only to be true from env")
	}
}

func TestOverrideFromEnv(t *testing.T) {
	cfg := Default()

	t.Setenv("ROUTERPILOT_LOG_LEVEL", "debug")
	t.Setenv("ROUTERPILOT_LOG_FORMAT", "json")
	t.Setenv("ROUTERPILOT_PLANNER", "llm")
	t.Setenv("ROUTERPILOT_TELEGRAM_TOKEN", "test:token")
	t.Setenv("ROUTERPILOT_RISK", "high")

	cfg.overrideFromEnv()

	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected debug, got %s", cfg.Logging.Level)
	}
	if cfg.Logging.Format != "json" {
		t.Fatalf("expected json, got %s", cfg.Logging.Format)
	}
	if cfg.Planner.Type != "llm" {
		t.Fatalf("expected llm, got %s", cfg.Planner.Type)
	}
	if cfg.Telegram.Token != "test:token" {
		t.Fatalf("expected test:token, got %s", cfg.Telegram.Token)
	}
	if cfg.Security.Risk != "high" {
		t.Fatalf("expected high, got %s", cfg.Security.Risk)
	}
}
