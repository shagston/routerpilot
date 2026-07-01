package memory

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	sdk "github.com/shagston/routerpilot/sdk/memory"
	"github.com/shagston/routerpilot/sdk/types"
)

type Store struct {
	working    *WorkingMemory
	session    *SessionMemory
	persistent sdk.Provider
	external   sdk.Provider
	log        *slog.Logger
}

func NewStore(persistent sdk.Provider) *Store {
	return &Store{
		working:    NewWorkingMemory(),
		session:    NewSessionMemory(),
		persistent: persistent,
		log:        slog.With("component", "memory"),
	}
}

func (s *Store) SetExternal(provider sdk.Provider) {
	s.external = provider
}

func (s *Store) Working() *WorkingMemory {
	return s.working
}

func (s *Store) Session() *SessionMemory {
	return s.session
}

func (s *Store) Get(ctx context.Context, key string) (sdk.Record, error) {
	if r, ok := s.working.Get(ctx, key); ok {
		return r, nil
	}

	if r, ok := s.session.Get(ctx, key); ok {
		return r, nil
	}

	if s.persistent != nil {
		r, err := s.persistent.Read(ctx, key)
		if err == nil {
			return r, nil
		}
		if !errors.Is(err, types.ErrNotFound) {
			s.log.Warn("persistent memory read error", "key", key, "error", err)
		}
	}

	if s.external != nil {
		r, err := s.external.Read(ctx, key)
		if err == nil {
			return r, nil
		}
		if !errors.Is(err, types.ErrNotFound) {
			s.log.Warn("external memory read error", "key", key, "error", err)
		}
	}

	return sdk.Record{}, fmt.Errorf("%w: key %q not found in any layer", types.ErrNotFound, key)
}

func (s *Store) Put(ctx context.Context, key string, value map[string]any, layer Layer) error {
	record := sdk.Record{
		Key:   key,
		Value: value,
	}

	switch layer {
	case LayerWorking:
		s.working.Set(ctx, record)
		return nil

	case LayerSession:
		s.session.Set(ctx, record)
		return nil

	case LayerPersistent:
		if s.persistent == nil {
			return fmt.Errorf("persistent provider not configured")
		}
		return s.persistent.Write(ctx, record)

	case LayerExternal:
		if s.external == nil {
			return fmt.Errorf("external provider not configured")
		}
		return s.external.Write(ctx, record)

	default:
		return fmt.Errorf("unknown memory layer: %v", layer)
	}
}

func (s *Store) Delete(ctx context.Context, key string, layer Layer) error {
	switch layer {
	case LayerWorking:
		s.working.Delete(ctx, key)
		return nil

	case LayerSession:
		s.session.Delete(ctx, key)
		return nil

	case LayerPersistent:
		if s.persistent == nil {
			return fmt.Errorf("persistent provider not configured")
		}
		return s.persistent.Write(ctx, sdk.Record{Key: key, Value: nil})
		// note: a real delete would need a Delete method on Provider

	case LayerExternal:
		if s.external == nil {
			return fmt.Errorf("external provider not configured")
		}
		return s.external.Write(ctx, sdk.Record{Key: key, Value: nil})

	default:
		return fmt.Errorf("unknown memory layer: %v", layer)
	}
}

func (s *Store) Search(ctx context.Context, prefix string, limit int, layer Layer) ([]sdk.Record, error) {
	switch layer {
	case LayerWorking:
		return s.working.Search(ctx, prefix, limit), nil

	case LayerSession:
		return s.session.Search(ctx, prefix, limit), nil

	case LayerPersistent:
		if s.persistent == nil {
			return nil, fmt.Errorf("persistent provider not configured")
		}
		return s.persistent.Search(ctx, prefix, limit)

	case LayerExternal:
		if s.external == nil {
			return nil, fmt.Errorf("external provider not configured")
		}
		return s.external.Search(ctx, prefix, limit)

	default:
		return nil, fmt.Errorf("unknown memory layer: %v", layer)
	}
}

func (s *Store) SearchAll(ctx context.Context, prefix string, limit int) ([]sdk.Record, error) {
	if limit < 1 {
		limit = 100
	}

	var results []sdk.Record
	seen := make(map[string]bool)

	for _, layer := range AllLayers {
		records, err := s.Search(ctx, prefix, limit-len(results), layer)
		if err != nil {
			continue
		}
		for _, r := range records {
			if !seen[r.Key] {
				seen[r.Key] = true
				results = append(results, r)
				if len(results) >= limit {
					return results, nil
				}
			}
		}
	}

	return results, nil
}

func (s *Store) ClearWorking(ctx context.Context) {
	s.working.Clear(ctx)
}

func (s *Store) ClearSession(ctx context.Context) {
	s.session.Clear(ctx)
}

func (s *Store) SessionCleanup(ctx context.Context) {
	s.session.Cleanup(ctx)
}
