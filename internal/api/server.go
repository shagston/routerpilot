package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/shagston/routerpilot/internal/app"
	"github.com/shagston/routerpilot/internal/config"
	"github.com/shagston/routerpilot/internal/webui"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
)

type Server struct {
	App    *app.App
	http   *http.Server
	mu     sync.Mutex
}

type IntentRequest struct {
	Intent  string         `json:"intent"`
	Args    map[string]any `json:"args"`
	Timeout time.Duration  `json:"timeout"`
}

type PlanRequest struct {
	Intent  string         `json:"intent"`
	Args    map[string]any `json:"args"`
}

func NewServer(a *app.App) *Server {
	return &Server{App: a}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", s.handleRoot)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/events", s.handleEvents)
	mux.HandleFunc("/events/stream", s.handleEventsStream)
	mux.HandleFunc("/intent", s.handleIntent)
	mux.HandleFunc("/plan", s.handlePlan)
	mux.HandleFunc("/tools", s.handleTools)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.Handle("/", webui.Handler())
	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"service": "RouterPilot",
		"version": "0.1.0",
		"endpoints": []string{
			"GET  /              - Web UI",
			"GET  /api           - API info",
			"GET  /health        - Health check",
			"POST /intent        - Execute an intent",
			"POST /plan          - Preview a plan without executing",
			"GET  /tools         - List available tools",
			"GET  /status        - Server status",
			"GET  /events        - List execution events",
			"GET  /events/stream - Stream execution events (SSE)",
			"GET  /ws            - WebSocket",
		},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleIntent(w http.ResponseWriter, r *http.Request) {
	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Intent == "" {
		writeError(w, http.StatusBadRequest, "Intent name is required")
		return
	}

	timeout := 5 * time.Minute
	if req.Timeout > 0 {
		timeout = req.Timeout
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	intent := sdkPlanner.Intent{
		Name:      req.Intent,
		Arguments: req.Args,
	}

	execution, err := s.App.ExecuteIntent(ctx, intent, false)
	if err != nil {
		var safetyErr *app.SafetyError
		if errors.As(err, &safetyErr) {
			writeJSON(w, http.StatusForbidden, map[string]string{
				"error":   "safety_confirmation_required",
				"message": "This intent is too risky for API execution. Please use CLI for manual confirmation.",
			})
			return
		}

		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, execution)
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	var req PlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Intent == "" {
		writeError(w, http.StatusBadRequest, "Intent name is required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	snapshot, plan, err := s.App.PreviewPlan(ctx, sdkPlanner.Intent{
		Name:      req.Intent,
		Arguments: req.Args,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"plan":     plan,
		"context":  snapshot,
	})
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.App.Registry.List())
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "running",
		"tools_cnt": len(s.App.Registry.List()),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		writeJSON(w, http.StatusOK, s.App.Config)
	case "PUT":
		var incoming config.Config
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}

		cfg := s.App.Config
		if v := incoming.Server.Port; v != "" {
			cfg.Server.Port = v
		}
		if v := incoming.Server.Host; v != "" {
			cfg.Server.Host = v
		}
		if v := incoming.Planner.Type; v != "" {
			cfg.Planner.Type = v
		}
		if v := incoming.Planner.APIKey; v != "" {
			cfg.Planner.APIKey = v
		}
		if v := incoming.Planner.Endpoint; v != "" {
			cfg.Planner.Endpoint = v
		}
		if v := incoming.Planner.Model; v != "" {
			cfg.Planner.Model = v
		}
		if v := incoming.Logging.Level; v != "" {
			cfg.Logging.Level = v
		}
		if v := incoming.Logging.Format; v != "" {
			cfg.Logging.Format = v
		}
		if v := incoming.Telegram.Token; v != "" {
			cfg.Telegram.Token = v
		}
		if v := incoming.Security.Risk; v != "" {
			cfg.Security.Risk = v
		}
		if len(incoming.Security.Permissions) > 0 {
			cfg.Security.Permissions = incoming.Security.Permissions
		}
		cfg.Security.ReadOnly = incoming.Security.ReadOnly
		cfg.Security.DryRun = incoming.Security.DryRun
		if v := incoming.System.PluginDir; v != "" {
			cfg.System.PluginDir = v
		}

		data, err := json.MarshalIndent(cfg, "", "  ")
		if err == nil {
			_ = os.WriteFile("routerpilot.json", data, 0644)
		}

		writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "config": cfg})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) Serve(addr string) error {
	s.mu.Lock()
	s.http = &http.Server{
		Addr:    addr,
		Handler: s.Routes(),
	}
	s.mu.Unlock()
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http == nil {
		return nil
	}
	return s.http.Shutdown(ctx)
}
