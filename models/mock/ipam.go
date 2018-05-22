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
	copied := *config
	d.ipam = &copied
	return nil
}

func (d *driver) getIPAMConfig() (*sabakan.IPAMConfig, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.ipam == nil {
		return nil, errors.New("IPAMConfig is not set")
	}
	copied := *d.ipam
	return &copied, nil
}

type ipamDriver struct {
	*driver
}

func (d ipamDriver) PutConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	return d.putIPAMConfig(ctx, config)
}

func (d ipamDriver) GetConfig() (*sabakan.IPAMConfig, error) {
	return d.getIPAMConfig()
}
