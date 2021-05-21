package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/cybozu-go/sabakan/v2"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/clientv3util"
)

func (d *driver) putIPAMConfig(ctx context.Context, config *sabakan.IPAMConfig) error {
	j, err := json.Marshal(config)
	if err != nil {
		return err
	}

	sj := string(j)
	tresp, err := d.client.Txn(ctx).
		If(clientv3util.KeyMissing(KeyMachines).WithPrefix()).
		Then(clientv3.OpPut(KeyIPAM, sj)).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !tresp.Succeeded {
		return errors.New("machines already exists")
	}

	d.addLog(ctx, time.Now(), tresp.Header.Revision, sabakan.AuditIPAM, "config", "put", sj)

	return nil
}

func (d *driver) getIPAMConfig() (*sabakan.IPAMConfig, error) {
	v := d.ipamConfig.Load()
	if v == nil {
		return nil, errors.New("IPAMConfig is not set")
	}

	return v.(*sabakan.IPAMConfig), nil
}

func (d *driver) handleIPAMConfig(ev *clientv3.Event) error {
	if ev.Type == clientv3.EventTypeDelete {
		return nil
	}

	config := new(sabakan.IPAMConfig)
	err := json.Unmarshal(ev.Kv.Value, config)
	if err != nil {
		return err
	}

	d.ipamConfig.Store(config)
	return nil
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
