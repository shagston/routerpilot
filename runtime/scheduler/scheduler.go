package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/shagston/routerpilot/sdk/events"
	"github.com/shagston/routerpilot/sdk/types"
)

type ExecutionFunc func(ctx context.Context, sched Schedule) error

type Scheduler struct {
	mu       sync.RWMutex
	schedules map[string]*Schedule
	dag      *DAG
	active   map[string]int

	execFn  ExecutionFunc
	pub     events.Publisher

	cronChecker  *cronTracker
	stopCh       chan struct{}
	wg           sync.WaitGroup
	log          *slog.Logger

	maxActive        int
	globalActive     int
	tickInterval     time.Duration
}

type cronTracker struct {
	entries map[string]*cronExpr
}

type Option func(*Scheduler)

func WithEventPublisher(pub events.Publisher) Option {
	return func(s *Scheduler) { s.pub = pub }
}

func WithMaxActive(n int) Option {
	return func(s *Scheduler) { s.maxActive = n }
}

func WithTickInterval(d time.Duration) Option {
	return func(s *Scheduler) { s.tickInterval = d }
}

func New(execFn ExecutionFunc, opts ...Option) *Scheduler {
	sch := &Scheduler{
		schedules:    make(map[string]*Schedule),
		dag:          NewDAG(),
		active:       make(map[string]int),
		execFn:       execFn,
		cronChecker:  &cronTracker{entries: make(map[string]*cronExpr)},
		stopCh:       make(chan struct{}),
		log:          slog.With("component", "scheduler"),
		maxActive:    10,
		tickInterval: time.Second,
	}
	for _, opt := range opts {
		opt(sch)
	}
	return sch
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.wg.Add(1)
	go s.tickLoop(ctx)
	s.log.Info("scheduler started")
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	close(s.stopCh)
	s.wg.Wait()
	s.log.Info("scheduler stopped")
	return nil
}

func (s *Scheduler) Register(sched Schedule) error {
	if sched.ID == "" {
		return fmt.Errorf("schedule ID is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.schedules[sched.ID]; exists {
		return fmt.Errorf("schedule %s already registered", sched.ID)
	}

	if sched.Type == TypeCron {
		expr, err := parseCron(sched.Expression)
		if err != nil {
			return fmt.Errorf("parse cron %q: %w", sched.Expression, err)
		}
		s.cronChecker.entries[sched.ID] = expr
	}

	for _, dep := range sched.Deps {
		if _, ok := s.schedules[dep]; !ok {
			if _, exists := s.schedules[dep]; !exists {
				continue
			}
		}
		if err := s.dag.AddEdge(sched.ID, dep); err != nil {
			return fmt.Errorf("dependency %s: %w", dep, err)
		}
	}

	sched.Status = StatusPending
	sched.CreatedAt = time.Now()
	sched.UpdatedAt = time.Now()
	s.schedules[sched.ID] = &sched

	s.emitEvent(sched.ID, "scheduler.registered")
	s.log.Info("schedule registered", "id", sched.ID, "type", sched.Type.String())

	return nil
}

func (s *Scheduler) Remove(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.schedules[id]; !ok {
		return fmt.Errorf("schedule %s not found", id)
	}

	delete(s.schedules, id)
	delete(s.cronChecker.entries, id)

	s.emitEvent(id, "scheduler.cancelled")
	s.log.Info("schedule removed", "id", id)

	return nil
}

func (s *Scheduler) Trigger(id string) error {
	s.mu.RLock()
	sched, ok := s.schedules[id]
	s.mu.RUnlock()

	if !ok {
		return fmt.Errorf("schedule %s not found", id)
	}

	go s.executeWithRetry(context.Background(), *sched)
	return nil
}

func (s *Scheduler) List() []Schedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Schedule, 0, len(s.schedules))
	for _, sched := range s.schedules {
		result = append(result, *sched)
	}
	return result
}

func (s *Scheduler) Get(id string) (Schedule, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sched, ok := s.schedules[id]
	if !ok {
		return Schedule{}, false
	}
	return *sched, true
}

func (s *Scheduler) OnEvent(ctx context.Context, event types.Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, sched := range s.schedules {
		if sched.Type != TypeEvent {
			continue
		}
		for _, et := range sched.EventTypes {
			if string(event.Type) == et {
				s.log.Debug("event trigger matched", "schedule", sched.ID, "event", et)
				go s.executeWithRetry(ctx, *sched)
				break
			}
		}
	}
}

func (s *Scheduler) tickLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case now := <-ticker.C:
			s.processTick(ctx, now)
		}
	}
}

func (s *Scheduler) processTick(ctx context.Context, now time.Time) {
	type trigger struct {
		sched Schedule
	}

	s.mu.RLock()
	if s.globalActive >= s.maxActive {
		s.mu.RUnlock()
		return
	}

	var triggers []trigger
	for _, sched := range s.schedules {
		if sched.Status == StatusCancelled {
			continue
		}
		if s.active[sched.ID] >= sched.MaxActive && sched.MaxActive > 0 {
			continue
		}
		if s.globalActive >= s.maxActive {
			break
		}

		var shouldTrigger bool
		switch sched.Type {
		case TypeCron:
			expr, ok := s.cronChecker.entries[sched.ID]
			if !ok {
				continue
			}
			next := expr.next(now.Add(-time.Second))
			shouldTrigger = !next.IsZero() && next.Sub(now) < time.Second

		case TypeInterval:
			shouldTrigger = sched.UpdatedAt.Add(sched.Interval).Before(now) || sched.UpdatedAt == sched.CreatedAt

		case TypeOneshot:
			shouldTrigger = !sched.RunAt.IsZero() && sched.RunAt.Before(now) && sched.Status == StatusPending

		default:
			continue
		}

		if shouldTrigger {
			sched.Status = StatusActive
			sched.UpdatedAt = now
			s.active[sched.ID]++
			s.globalActive++
			s.emitEvent(sched.ID, "scheduler.triggered")
			triggers = append(triggers, trigger{sched: *sched})
		}
	}
	s.mu.RUnlock()

	for _, t := range triggers {
		s.executeWithRetry(ctx, t.sched)
		s.mu.Lock()
		s.active[t.sched.ID]--
		s.globalActive--
		s.mu.Unlock()
	}
}

func (s *Scheduler) executeWithRetry(ctx context.Context, sched Schedule) {
	maxAttempts := 1
	var backoff time.Duration

	if sched.Retry != nil {
		maxAttempts = sched.Retry.MaxAttempts
		backoff = sched.Retry.Backoff
		if maxAttempts < 1 {
			maxAttempts = 1
		}
		if backoff <= 0 {
			backoff = time.Second
		}
	}

	s.emitEvent(sched.ID, "scheduler.started")

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			s.log.Debug("retrying schedule", "id", sched.ID, "attempt", attempt)
			time.Sleep(backoff)
			if backoff < 60*time.Second {
				backoff *= 2
			}
		}

		execCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		err := s.execFn(execCtx, sched)
		cancel()

		if err == nil {
			s.mu.Lock()
			if sched.Status == StatusActive {
				sched.Status = StatusCompleted
			}
			s.schedules[sched.ID] = &sched
			s.mu.Unlock()

			s.emitEvent(sched.ID, "scheduler.completed")
			return
		}
		lastErr = err
	}

	s.mu.Lock()
	sched.Status = StatusFailed
	s.schedules[sched.ID] = &sched
	s.mu.Unlock()

	s.emitEventWithError(sched.ID, "scheduler.failed", lastErr)
	s.log.Error("schedule failed", "id", sched.ID, "error", lastErr)
}

func (s *Scheduler) emitEvent(scheduleID, eventType string) {
	if s.pub == nil {
		return
	}
	s.pub.Publish(types.Event{
		Timestamp: time.Now(),
		Type:      types.EventType(eventType),
		Source:    "scheduler",
		Severity:  types.SeverityInfo,
		Priority:  types.PriorityNormal,
		Payload: map[string]any{
			"schedule_id": scheduleID,
		},
	})
}

func (s *Scheduler) emitEventWithError(scheduleID, eventType string, err error) {
	if s.pub == nil {
		return
	}
	s.pub.Publish(types.Event{
		Timestamp: time.Now(),
		Type:      types.EventType(eventType),
		Source:    "scheduler",
		Severity:  types.SeverityError,
		Priority:  types.PriorityNormal,
		Payload: map[string]any{
			"schedule_id": scheduleID,
			"error":       err.Error(),
		},
	})
}
