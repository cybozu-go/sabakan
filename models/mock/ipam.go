package mock

import (
	"context"

	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

type ipamDriver struct {
	driver *driver
	ipam   *sabakan.IPAMConfig
}

func newIPAMDriver(d *driver) *ipamDriver {
	return &ipamDriver{
		driver: d,
	}
}

func (d *ipamDriver) PutConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	d.driver.mu.Lock()
	defer d.driver.mu.Unlock()

	if len(d.driver.machines) > 0 {
		return errors.New("machines already exist")
	}
	copied := *config
	d.ipam = &copied
	return nil
}

func (d *ipamDriver) GetConfig() (*sabakan.IPAMConfig, error) {
	d.driver.mu.Lock()
	defer d.driver.mu.Unlock()
	if d.ipam == nil {
		return nil, errors.New("IPAMConfig is not set")
	}
	copied := *d.ipam
	return &copied, nil
}
