package reticulum

import (
	"fmt"
	"sync"
)

var globalHub = &hub{
	transports: make(map[string]*ReticulumTransport),
}

type hub struct {
	mu         sync.RWMutex
	transports map[string]*ReticulumTransport
}

func (h *hub) register(tp *ReticulumTransport) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	id := tp.identity.ID
	if _, ok := h.transports[id]; ok {
		return fmt.Errorf("transport %s already registered", id)
	}
	h.transports[id] = tp
	return nil
}

func (h *hub) unregister(id string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.transports, id)
}

func (h *hub) resolve(id string) (*ReticulumTransport, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	tp, ok := h.transports[id]
	return tp, ok
}

func (h *hub) all() []*ReticulumTransport {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]*ReticulumTransport, 0, len(h.transports))
	for _, tp := range h.transports {
		result = append(result, tp)
	}
	return result
}
