package mock

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) Register(ctx context.Context, machines []*sabakan.Machine) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, m := range machines {
		d.machines[m.Serial] = m
	}
	return nil
}

func (d *driver) Query(ctx context.Context, q *Query) ([]*sabakan.Machine, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	res := make([]*sabakan.Machine)
	for _, m := range d.machines {
		if q.Match(m) {
			res = append(res, m)
		}
	}
	return res, nil
}

func (d *driver) Delete(ctx context.Context, serials []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, serial := range serials {
		delete(d.machines, serial)
	}

	return nil
}
