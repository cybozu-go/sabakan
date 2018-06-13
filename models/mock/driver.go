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
	machines map[string]*sabakan.Machine
	ipam     *sabakan.IPAMConfig
	dhcp     *sabakan.DHCPConfig
}

// NewModel returns sabakan.Model
func NewModel() sabakan.Model {
	d := &driver{
		machines: make(map[string]*sabakan.Machine),
	}
	return sabakan.Model{
		Runner:   d,
		Storage:  newStorageDriver(),
		Machine:  newMachineDriver(d),
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
