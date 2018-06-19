package mock

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) Register(ctx context.Context, machines []*sabakan.Machine) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, m := range machines {
		if _, ok := d.machines[m.Spec.Serial]; ok {
			return sabakan.ErrConflicted
		}
	}
	for _, m := range machines {
		d.machines[m.Spec.Serial] = m
	}
	return nil
}

func (d *driver) SetState(ctx context.Context, serial string, state sabakan.MachineState) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	m, ok := d.machines[serial]
	if !ok {
		return sabakan.ErrNotFound
	}
	return m.SetState(state)
}

func (d *driver) Query(ctx context.Context, q *sabakan.Query) ([]*sabakan.Machine, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	res := make([]*sabakan.Machine, 0)
	for _, m := range d.machines {
		if q.Match(&m.Spec) {
			res = append(res, m)
		}
	}
	return res, nil
}

func (d *driver) Delete(ctx context.Context, serial string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	_, ok := d.machines[serial]
	if !ok {
		return sabakan.ErrNotFound
	}

	delete(d.machines, serial)
	return nil
}
