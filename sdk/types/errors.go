package types

import "errors"

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrTimeout           = errors.New("timeout")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrCapabilityMissing = errors.New("capability missing")
	ErrExecutionFailed   = errors.New("execution failed")
	ErrCancelled         = errors.New("cancelled")
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
)
