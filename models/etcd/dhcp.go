package etcd

import (
	"context"
	"encoding/json"
	"path"

	"github.com/cybozu-go/sabakan"
)

func (d *driver) putDHCPConfig(ctx context.Context, config *sabakan.DHCPConfig) error {
	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configKey := path.Join(d.prefix, KeyDHCP)

	_, err = d.client.Put(ctx, configKey, string(j))
	return err
}

func (d *driver) getDHCPConfig(ctx context.Context) (*sabakan.DHCPConfig, error) {
	key := path.Join(d.prefix, KeyDHCP)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	var config sabakan.DHCPConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
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
