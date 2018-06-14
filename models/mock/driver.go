// Package mock implements mockup sabakan model for testing.
package mock

import (
	"context"
	"sync"

	"github.com/cybozu-go/sabakan"
)

// driver implements all interfaces for sabakan model.
type driver struct {
	mu       sync.Mutex
	ipam     *sabakan.IPAMConfig
	machines map[string]*sabakan.Machine
}

// NewModel returns sabakan.Model
func NewModel() sabakan.Model {
	d := &driver{
		machines: make(map[string]*sabakan.Machine),
	}
	return sabakan.Model{
		Runner:   d,
		Storage:  newStorageDriver(),
		Machine:  d,
		IPAM:     ipamDriver{d},
		DHCP:     newDHCPDriver(d),
		Image:    newImageDriver(),
		Asset:    newAssetDriver(),
		Ignition: newIgnitionDriver(),
	}
}

func (d *driver) Run(ctx context.Context, ch chan<- struct{}) error {
	ch <- struct{}{}
	return nil
}
