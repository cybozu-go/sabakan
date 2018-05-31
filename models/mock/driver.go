// Package mock implements mockup sabakan model for testing.
package mock

import (
	"context"
	"sync"

	"github.com/cybozu-go/sabakan"
)

// driver implements all interfaces for sabakan model.
type driver struct {
	mu        sync.Mutex
	storage   map[string][]byte
	machines  map[string]*sabakan.Machine
	leases    map[string]*leaseUsage
	ipam      *sabakan.IPAMConfig
	dhcp      *sabakan.DHCPConfig
	ignitions map[string]map[string]string
}

// NewModel returns sabakan.Model
func NewModel() sabakan.Model {
	d := &driver{
		storage:   make(map[string][]byte),
		machines:  make(map[string]*sabakan.Machine),
		leases:    make(map[string]*leaseUsage),
		ignitions: make(map[string]map[string]string),
	}
	return sabakan.Model{
		Runner:  d,
		Storage: d,
		Machine: d,
		IPAM:    ipamDriver{d},
		DHCP:    dhcpDriver{d},
		Image:   newImageDriver(),
	}
}

func (d *driver) Run(ctx context.Context, ch chan<- struct{}) error {
	ch <- struct{}{}
	return nil
}
