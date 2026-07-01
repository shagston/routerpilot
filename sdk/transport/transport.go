package transport

import "context"

type Envelope struct {
	ID       string            `json:"id"`
	Source   string            `json:"source"`
	Target   string            `json:"target"`
	Type     string            `json:"type"`
	Payload  []byte            `json:"payload"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Endpoint interface {
	ID() string
	Send(ctx context.Context, env Envelope) error
	Receive(ctx context.Context) (<-chan Envelope, error)
	Close(ctx context.Context) error
}

type Router interface {
	Route(ctx context.Context, env Envelope) (string, error)
	AddRoute(pattern string, endpointID string) error
	RemoveRoute(pattern string) error
}

type Info struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type Discovery interface {
	Register(ctx context.Context, info Info) error
	Unregister(ctx context.Context, id string) error
	Lookup(ctx context.Context, svc string) ([]Info, error)
}

type Transport interface {
	Endpoint() Endpoint
	Router() Router
	Discovery() Discovery
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
