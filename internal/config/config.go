package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Planner  PlannerConfig  `json:"planner"`
	Logging  LoggingConfig  `json:"logging"`
	Telegram TelegramConfig `json:"telegram"`
	Security SecurityConfig `json:"security"`
	System   SystemConfig   `json:"system"`
}

type ServerConfig struct {
	Port string `json:"port"`
	Host string `json:"host"`
}

type PlannerConfig struct {
	Type     string `json:"type"`
	APIKey   string `json:"api_key"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
}

type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type TelegramConfig struct {
	Token string `json:"token"`
}

type SecurityConfig struct {
	Risk        string   `json:"risk"`
	Permissions []string `json:"permissions"`
}

type SystemConfig struct {
	PluginDir string `json:"plugin_dir"`
}

func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: ":8080",
			Host: "",
		},
		Planner: PlannerConfig{
			Type: "simple",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		Telegram: TelegramConfig{},
		Security: SecurityConfig{
			Risk:        "low",
			Permissions: []string{"read", "write"},
		},
		System: SystemConfig{
			PluginDir: "plugins",
		},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	if path == "" {
		path = os.Getenv("ROUTERPILOT_CONFIG")
	}
	if path == "" {
		path = "routerpilot.json"
	}

	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("config: failed to parse %s: %w", path, err)
		}
	}

	cfg.overrideFromEnv()

	return cfg, nil
}

func (c *Config) overrideFromEnv() {
	if v := os.Getenv("ROUTERPILOT_PORT"); v != "" {
		if v[0] != ':' {
			v = ":" + v
		}
		c.Server.Port = v
	}
	if v := os.Getenv("ROUTERPILOT_PLANNER"); v != "" {
		c.Planner.Type = v
	}
	if v := os.Getenv("ROUTERPILOT_API_KEY"); v != "" {
		c.Planner.APIKey = v
	}
	if v := os.Getenv("ROUTERPILOT_ENDPOINT"); v != "" {
		c.Planner.Endpoint = v
	}
	if v := os.Getenv("ROUTERPILOT_MODEL"); v != "" {
		c.Planner.Model = v
	}
	if v := os.Getenv("ROUTERPILOT_LOG_LEVEL"); v != "" {
		c.Logging.Level = v
	}
	if v := os.Getenv("ROUTERPILOT_LOG_FORMAT"); v != "" {
		c.Logging.Format = v
	}
	if v := os.Getenv("ROUTERPILOT_TELEGRAM_TOKEN"); v != "" {
		c.Telegram.Token = v
	}
	if v := os.Getenv("ROUTERPILOT_RISK"); v != "" {
		c.Security.Risk = v
	}
	if v := os.Getenv("ROUTERPILOT_PERMISSIONS"); v != "" {
		c.Security.Permissions = strings.Split(v, ",")
	}
	if v := os.Getenv("ROUTERPILOT_PLUGIN_DIR"); v != "" {
		c.System.PluginDir = v
	}
}
