package mock

import (
	"context"
	"sync"

	"github.com/cybozu-go/sabakan"
)

type machineDriver struct {
	driver *driver
	mu     sync.Mutex
}

func newMachineDriver(d *driver) *machineDriver {
	return &machineDriver{
		driver: d,
	}
}

func (d *machineDriver) Register(ctx context.Context, machines []*sabakan.Machine) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, m := range machines {
		if _, ok := d.driver.machines[m.Serial]; ok {
			return sabakan.ErrConflicted
		}
	}
	for _, m := range machines {
		d.driver.machines[m.Serial] = m
	}
	return nil
}

func (d *machineDriver) Query(ctx context.Context, q *sabakan.Query) ([]*sabakan.Machine, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	res := make([]*sabakan.Machine, 0)
	for _, m := range d.driver.machines {
		if q.Match(m) {
			res = append(res, m)
		}
	}
	return res, nil
}

func (d *machineDriver) Delete(ctx context.Context, serial string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.driver.machines[serial]
	if !ok {
		return sabakan.ErrNotFound
	}

	delete(d.driver.machines, serial)
	return nil
}
