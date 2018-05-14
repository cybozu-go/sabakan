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

// PutConfig implements sabakan.ConfigModel
func (d *Driver) PutConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
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

// GetConfig implements sabakan.ConfigModel
func (d *Driver) GetConfig(ctx context.Context) (*sabakan.IPAMConfig, error) {
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
