package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

// PutConfig implements sabakan.ConfigModel
func (d *Driver) PutConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	configKey := path.Join(d.prefix, KeyConfig)
	machinesKey := path.Join(d.prefix, KeyMachines)

	txnThenOps := []clientv3.Op{}
	txnThenOps = append(txnThenOps, clientv3.OpPut(configKey, string(j)))
	// delete all indices before put, because number of racks may be decreased in config
	txnThenOps = append(txnThenOps, clientv3.OpDelete(path.Join(d.prefix, KeyNodeIndices), clientv3.WithPrefix()))
	tresp, err := d.client.Txn(ctx).
		If(clientv3util.KeyMissing(machinesKey).WithPrefix()).
		Then(txnThenOps...).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		return errors.New("machines already exists")
	}

	// we can put keys outside transaction
	for rackIdx := 0; uint(rackIdx) < config.MaxRacks; rackIdx++ {
		for nodeIdx := 0; uint(nodeIdx) < config.MaxNodesInRack; nodeIdx++ {
			key := path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rackIdx), fmt.Sprintf("%02d", nodeIdx))
			_, err := d.client.Put(ctx, key, fmt.Sprint(nodeIdx))
			if err != nil {
				return err
			}
		}
	}

	return err
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
