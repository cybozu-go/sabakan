package mock

import (
	"context"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) putDHCPConfig(ctx context.Context, config *sabakan.DHCPConfig) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.dhcp = *config
	return nil
}

func (d *driver) getDHCPConfig(ctx context.Context) (*sabakan.DHCPConfig, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	copied := d.dhcp
	return &copied, nil
}

type dhcpDriver struct {
	*driver
}

func (d dhcpDriver) PutConfig(ctx context.Context, config *sabakan.DHCPConfig) error {
	return d.putDHCPConfig(ctx, config)
}

func (d dhcpDriver) GetConfig(ctx context.Context) (*sabakan.DHCPConfig, error) {
	return d.getDHCPConfig(ctx)
}
