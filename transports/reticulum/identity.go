package reticulum

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
)

var globalIdentitySeq atomic.Uint64

type Identity struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Labels       map[string]string `json:"labels,omitempty"`
	Capabilities []string          `json:"capabilities,omitempty"`
}

func GenerateIdentity(name, version string) Identity {
	id := make([]byte, 8)
	rand.Read(id)
	seq := globalIdentitySeq.Add(1)
	return Identity{
		ID:      fmt.Sprintf("%x-%d", id, seq),
		Name:    name,
		Version: version,
		Labels:  make(map[string]string),
	}
}
