package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shagston/routerpilot/sdk/types"
)

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	executionID := types.ExecutionID(r.URL.Query().Get("execution_id"))
	events := s.App.Events.EventsFiltered(executionID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleEventsStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	executionID := types.ExecutionID(r.URL.Query().Get("execution_id"))
	if r.URL.Query().Get("replay") != "false" {
		for _, event := range s.App.Events.EventsFiltered(executionID) {
			if err := writeSSEEvent(w, event); err != nil {
				return
			}
			flusher.Flush()
		}
	}

	ch, unsubscribe := s.App.Events.SubscribeChan(256)
	defer unsubscribe()

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			if executionID != "" && event.ExecutionID != executionID {
				continue
			}
			if err := writeSSEEvent(w, event); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeSSEEvent(w http.ResponseWriter, event types.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "event: %s\n", event.Type); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return err
	}
	return nil
}
