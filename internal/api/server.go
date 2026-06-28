package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/shagston/routerpilot/internal/app"
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
	mux.HandleFunc("GET /", s.handleRoot)
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /events", s.handleEvents)
	mux.HandleFunc("GET /events/stream", s.handleEventsStream)
	mux.HandleFunc("POST /intent", s.handleIntent)
	mux.HandleFunc("POST /plan", s.handlePlan)
	mux.HandleFunc("GET /tools", s.handleTools)
	mux.HandleFunc("GET /status", s.handleStatus)
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
			"GET  /health         - Health check",
			"POST /intent         - Execute an intent",
			"POST /plan           - Preview a plan without executing",
			"GET  /tools          - List available tools",
			"GET  /status         - Server status",
			"GET  /events         - List execution events",
			"GET  /events/stream  - Stream execution events (SSE)",
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
