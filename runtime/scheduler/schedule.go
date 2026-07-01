package scheduler

import "time"

type Type int

const (
	TypeCron    Type = 0
	TypeInterval Type = 1
	TypeOneshot Type = 2
	TypeEvent   Type = 3
)

func (t Type) String() string {
	switch t {
	case TypeCron:
		return "cron"
	case TypeInterval:
		return "interval"
	case TypeOneshot:
		return "oneshot"
	case TypeEvent:
		return "event"
	default:
		return "unknown"
	}
}

type Status int

const (
	StatusPending   Status = 0
	StatusActive    Status = 1
	StatusCompleted Status = 2
	StatusFailed    Status = 3
	StatusCancelled Status = 4
)

func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusActive:
		return "active"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

type RetryPolicy struct {
	MaxAttempts int
	Backoff     time.Duration
	MaxBackoff  time.Duration
}

type Schedule struct {
	ID          string
	Type        Type
	Priority    int
	Expression  string
	Interval    time.Duration
	RunAt       time.Time
	EventTypes  []string
	AgentID     string
	Capability  string
	Input       map[string]any
	Deps        []string
	Retry       *RetryPolicy
	MaxActive   int
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func CronSchedule(id, expression string) Schedule {
	return Schedule{
		ID:         id,
		Type:       TypeCron,
		Expression: expression,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func IntervalSchedule(id string, interval time.Duration) Schedule {
	return Schedule{
		ID:        id,
		Type:      TypeInterval,
		Interval:  interval,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func OneshotSchedule(id string, runAt time.Time) Schedule {
	return Schedule{
		ID:        id,
		Type:      TypeOneshot,
		RunAt:     runAt,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func EventSchedule(id string, eventTypes ...string) Schedule {
	return Schedule{
		ID:         id,
		Type:       TypeEvent,
		EventTypes: eventTypes,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}
