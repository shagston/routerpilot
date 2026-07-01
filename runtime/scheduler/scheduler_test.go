package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})
	ctx := context.Background()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := s.Stop(ctx); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestRegisterAndList(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})

	if err := s.Register(CronSchedule("test-1", "0 * * * * *")); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := s.Register(IntervalSchedule("test-2", 5*time.Second)); err != nil {
		t.Fatalf("Register: %v", err)
	}

	schedules := s.List()
	if len(schedules) != 2 {
		t.Errorf("expected 2 schedules, got %d", len(schedules))
	}
}

func TestRegisterDuplicate(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})

	if err := s.Register(CronSchedule("dup", "0 * * * * *")); err != nil {
		t.Fatalf("First register: %v", err)
	}
	if err := s.Register(CronSchedule("dup", "*/5 * * * * *")); err == nil {
		t.Error("expected error for duplicate schedule")
	}
}

func TestRemoveSchedule(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})

	if err := s.Register(CronSchedule("remove-me", "0 * * * * *")); err != nil {
		t.Fatalf("Register: %v", err)
	}
	if err := s.Remove("remove-me"); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, ok := s.Get("remove-me"); ok {
		t.Error("expected schedule to be removed")
	}
}

func TestRemoveNotFound(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})

	if err := s.Remove("nonexistent"); err == nil {
		t.Error("expected error for removing nonexistent schedule")
	}
}

func TestTriggerSchedule(t *testing.T) {
	var executed atomic.Int32
	s := New(func(ctx context.Context, sched Schedule) error {
		executed.Add(1)
		return nil
	})

	if err := s.Register(CronSchedule("trigger-test", "0 * * * * *")); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := s.Trigger("trigger-test"); err != nil {
		t.Fatalf("Trigger: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 1 {
		t.Errorf("expected 1 execution, got %d", executed.Load())
	}
}

func TestIntervalSchedule(t *testing.T) {
	var executed atomic.Int32
	s := New(func(ctx context.Context, sched Schedule) error {
		executed.Add(1)
		return nil
	}, WithTickInterval(10*time.Millisecond))
	ctx := context.Background()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Stop(ctx)

	sch := IntervalSchedule("interval-test", 50*time.Millisecond)
	if err := s.Register(sch); err != nil {
		t.Fatalf("Register: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if executed.Load() < 2 {
		t.Errorf("expected at least 2 executions, got %d", executed.Load())
	}
}

func TestOneshotSchedule(t *testing.T) {
	var executed atomic.Int32
	s := New(func(ctx context.Context, sched Schedule) error {
		executed.Add(1)
		return nil
	}, WithTickInterval(10*time.Millisecond))
	ctx := context.Background()

	if err := s.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer s.Stop(ctx)

	sch := OneshotSchedule("oneshot-test", time.Now().Add(30*time.Millisecond))
	if err := s.Register(sch); err != nil {
		t.Fatalf("Register: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 1 {
		t.Errorf("expected 1 execution, got %d", executed.Load())
	}
}

func TestRetryPolicy(t *testing.T) {
	var attempts atomic.Int32
	s := New(func(ctx context.Context, sched Schedule) error {
		attempts.Add(1)
		if attempts.Load() < 3 {
			return context.DeadlineExceeded
		}
		return nil
	})

	sch := CronSchedule("retry-test", "0 * * * * *")
	sch.Retry = &RetryPolicy{
		MaxAttempts: 3,
		Backoff:     10 * time.Millisecond,
	}

	if err := s.Register(sch); err != nil {
		t.Fatalf("Register: %v", err)
	}

	if err := s.Trigger("retry-test"); err != nil {
		t.Fatalf("Trigger: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	if attempts.Load() != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts.Load())
	}
}

func TestScheduleGet(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})

	sch := CronSchedule("get-test", "0 * * * * *")
	if err := s.Register(sch); err != nil {
		t.Fatalf("Register: %v", err)
	}

	got, ok := s.Get("get-test")
	if !ok {
		t.Fatal("expected to find schedule")
	}
	if got.ID != "get-test" {
		t.Errorf("expected ID get-test, got %s", got.ID)
	}
}

func TestDAGCycleDetection(t *testing.T) {
	dag := NewDAG()

	if err := dag.AddEdge("a", "b"); err != nil {
		t.Fatalf("AddEdge a->b: %v", err)
	}
	if err := dag.AddEdge("b", "c"); err != nil {
		t.Fatalf("AddEdge b->c: %v", err)
	}
	if err := dag.AddEdge("c", "a"); err == nil {
		t.Error("expected error for cyclic dependency")
	}
}

func TestDAGTopologicalSort(t *testing.T) {
	dag := NewDAG()

	dag.AddNode("a")
	dag.AddNode("b")
	dag.AddNode("c")
	dag.AddEdge("a", "b")
	dag.AddEdge("a", "c")

	order, err := dag.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort: %v", err)
	}

	if len(order) != 3 {
		t.Errorf("expected 3 items, got %d", len(order))
	}
}

func TestCronNext(t *testing.T) {
	expr, err := parseCron("0 */5 * * * *")
	if err != nil {
		t.Fatalf("parseCron: %v", err)
	}

	base := time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
	next := expr.next(base)

	if next.Minute() != 5 {
		t.Errorf("expected minute 5 (next match after 00:00:00), got %d", next.Minute())
	}
	if next.Second() != 0 {
		t.Errorf("expected second 0, got %d", next.Second())
	}
}

func TestCronEverySecond(t *testing.T) {
	expr, err := parseCron("* * * * * *")
	if err != nil {
		t.Fatalf("parseCron: %v", err)
	}

	base := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	next := expr.next(base)

	expected := base.Add(time.Second)
	if !next.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, next)
	}
}

func TestCronSpecificSecond(t *testing.T) {
	expr, err := parseCron("30 * * * * *")
	if err != nil {
		t.Fatalf("parseCron: %v", err)
	}

	base := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	next := expr.next(base)

	if next.Second() != 30 {
		t.Errorf("expected second 30, got %d", next.Second())
	}
	if next.Minute() != 0 {
		t.Errorf("expected minute 0, got %d", next.Minute())
	}
}

func TestEmptyScheduleID(t *testing.T) {
	s := New(func(ctx context.Context, sched Schedule) error {
		return nil
	})

	if err := s.Register(Schedule{ID: ""}); err == nil {
		t.Error("expected error for empty schedule ID")
	}
}
