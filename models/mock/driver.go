// Package mock implements mockup sabakan model for testing.
package mock

import (
	"sync"

	"github.com/cybozu-go/sabakan"
)

// driver implements all interfaces for sabakan model.
type driver struct {
	mu       sync.Mutex
	storage  map[string][]byte
	machines map[string]*sabakan.Machine
	config   sabakan.IPAMConfig
}

// NewModel returns sabakan.Model
func NewModel() sabakan.Model {
	d := &driver{
		storage:  make(map[string][]byte),
		machines: make(map[string]*sabakan.Machine),
	}
	return sabakan.Model{
		Storage: d,
		Machine: d,
		Config:  d,
	}
}
