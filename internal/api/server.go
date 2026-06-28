package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/shagston/routerpilot/internal/app"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
)

type Server struct {
	App *app.App
}

type IntentRequest struct {
	Intent  string         `json:"intent"`
	Args    map[string]any `json:"args"`
	Timeout time.Duration  `json:"timeout"`
}

func NewServer(a *app.App) *Server {
	return &Server{App: a}
}

func (s *Server) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /events", s.handleEvents)
	mux.HandleFunc("GET /events/stream", s.handleEventsStream)
	mux.HandleFunc("POST /intent", s.handleIntent)
	mux.HandleFunc("GET /tools", s.handleTools)
	mux.HandleFunc("GET /status", s.handleStatus)
	return mux
}

func (s *Server) handleIntent(w http.ResponseWriter, r *http.Request) {
	var req IntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.Intent == "" {
		http.Error(w, "Intent name is required", http.StatusBadRequest)
		return
	}

	// Установка таймаута
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
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "safety_confirmation_required",
				"message": "This intent is too risky for API execution. Please use CLI for manual confirmation.",
			})
			return
		}
		
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execution)
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	tools := s.App.Registry.List()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tools)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]any{
		"status":    "running",
		"tools_cnt": len(s.App.Registry.List()),
		"timestamp": time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
