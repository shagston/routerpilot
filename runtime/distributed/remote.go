package distributed

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/shagston/routerpilot/sdk/capability"
	sdkTp "github.com/shagston/routerpilot/sdk/transport"
)

type RemoteResult struct {
	Output map[string]any `json:"output"`
	Status string         `json:"status"`
	Error  string         `json:"error,omitempty"`
}

type pendingRequest struct {
	resultCh chan RemoteResult
	created  time.Time
	timeout  time.Duration
}

type RemoteExecutor struct {
	mesh     *Mesh
	mu       sync.RWMutex
	pending  map[string]*pendingRequest
	nextID   uint64
}

func NewRemoteExecutor(mesh *Mesh) *RemoteExecutor {
	return &RemoteExecutor{
		mesh:    mesh,
		pending: make(map[string]*pendingRequest),
	}
}

func (re *RemoteExecutor) Execute(ctx context.Context, targetID string, provider capability.Provider, input map[string]any) (map[string]any, error) {
	reqID := re.nextRequestID()

	payload := map[string]any{
		"request_id":   reqID,
		"capability":   string(provider.ID()),
		"name":         provider.Info().Name,
		"input":        input,
		"timeout_ms":   provider.Info().Timeout,
		"correlation":  reqID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal capability request: %w", err)
	}

	timeout := re.resolveTimeout(provider)
	resultCh := make(chan RemoteResult, 1)

	re.mu.Lock()
	re.pending[reqID] = &pendingRequest{
		resultCh: resultCh,
		created:  time.Now(),
		timeout:  timeout,
	}
	re.mu.Unlock()

	defer func() {
		re.mu.Lock()
		delete(re.pending, reqID)
		re.mu.Unlock()
	}()

	env := sdkTp.Envelope{
		Source: re.mesh.localID,
		Target: targetID,
		Type:   "capability.request",
		Payload: payloadBytes,
		Metadata: map[string]string{
			"request_id":  reqID,
			"capability":  string(provider.ID()),
			"correlation": reqID,
		},
	}

	if err := re.mesh.trans.Endpoint().Send(ctx, env); err != nil {
		return nil, fmt.Errorf("send capability request: %w", err)
	}

	select {
	case result := <-resultCh:
		if result.Error != "" {
			return nil, fmt.Errorf("remote execution error: %s", result.Error)
		}
		return result.Output, nil

	case <-ctx.Done():
		return nil, ctx.Err()

	case <-time.After(timeout):
		return nil, fmt.Errorf("remote execution timeout after %v", timeout)
	}
}

func (re *RemoteExecutor) handleResponse(ctx context.Context, env sdkTp.Envelope) {
	var payload map[string]any
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		return
	}

	reqID, _ := payload["request_id"].(string)
	if reqID == "" {
		reqID = env.Metadata["request_id"]
	}
	if reqID == "" {
		return
	}

	re.mu.RLock()
	pending, ok := re.pending[reqID]
	re.mu.RUnlock()

	if !ok {
		return
	}

	result := RemoteResult{
		Status: "success",
	}

	if status, ok := payload["status"].(string); ok {
		result.Status = status
	}
	if errStr, ok := payload["error"].(string); ok {
		result.Error = errStr
		result.Status = "error"
	}
	if output, ok := payload["output"].(map[string]any); ok {
		result.Output = output
	}

	pending.resultCh <- result
}

func (re *RemoteExecutor) nextRequestID() string {
	re.mu.Lock()
	defer re.mu.Unlock()
	re.nextID++
	return fmt.Sprintf("rem-%s-%d", re.mesh.localID, re.nextID)
}

func (re *RemoteExecutor) resolveTimeout(provider capability.Provider) time.Duration {
	timeout := provider.Info().Timeout
	if timeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(timeout) * time.Millisecond
}


