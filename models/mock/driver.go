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
	storage  map[string][]byte
	log      *sabakan.AuditLog
}

// NewModel returns sabakan.Model
func NewModel() sabakan.Model {
	d := &driver{
		machines: make(map[string]*sabakan.Machine),
		storage:  make(map[string][]byte),
	}
	return sabakan.Model{
		Runner:       d,
		IPAM:         ipamDriver{d},
		Machine:      machineDriver{d},
		Storage:      d,
		DHCP:         newDHCPDriver(d),
		Image:        newImageDriver(),
		Asset:        newAssetDriver(),
		Ignition:     newIgnitionDriver(),
		Log:          logDriver{d},
		KernelParams: newKernelParamsDriver(),
	}
}

func (d *driver) Run(ctx context.Context, ch chan<- struct{}) error {
	ch <- struct{}{}
	return nil
}
