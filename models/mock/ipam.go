package mock

import (
	"context"

	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

func (d *driver) putIPAMConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.machines) > 0 {
		return errors.New("machines already exist")
	}
	d.config = *config
	return nil
}

func (d *driver) getIPAMConfig(ctx context.Context) (*sabakan.IPAMConfig, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	copied := d.config
	return &copied, nil
}

type ipamDriver struct {
	*driver
}

func (d ipamDriver) PutConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	return d.putIPAMConfig(ctx, config)
}

func (d ipamDriver) GetConfig(ctx context.Context) (*sabakan.IPAMConfig, error) {
	return d.getIPAMConfig(ctx)
}
