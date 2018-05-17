package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
)

func (d *driver) putIPAMConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configKey := path.Join(d.prefix, KeyConfig)
	machinesKey := path.Join(d.prefix, KeyMachines)

	tresp, err := d.client.Txn(ctx).
		If(clientv3util.KeyMissing(machinesKey).WithPrefix()).
		Then(clientv3.OpPut(configKey, string(j))).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		return errors.New("machines already exists")
	}

	return nil
}

func (d *driver) getIPAMConfig(ctx context.Context) (*sabakan.IPAMConfig, error) {
	key := path.Join(d.prefix, KeyConfig)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, nil
	}
	var config sabakan.IPAMConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
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
