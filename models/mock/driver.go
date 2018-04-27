// Package mock implements mockup sabakan model for testing.
package etcd

import (
	"sync"

	"github.com/cybozu-go/sabakan"
)

// driver implements all interfaces for sabakan model.
type driver struct {
	mu      sync.Mutex
	storage map[string][]byte
}

// NewModel returns sabakan.Model
func NewModel() sabakan.Model {
	d := &driver{
		storage: make(map[string][]byte),
	}
	return sabakan.Model{
		Storage: d,
	}
}
