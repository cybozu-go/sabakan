package etcd

import (
	"context"
	"encoding/json"

	"path"

	"errors"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
)

func (d *Driver) Register(ctx context.Context, machines []*sabakan.Machine) error {

	wmcs := make([]*sabakan.MachineJson, len(machines))
	for i, rmc := range machines {
		var err error
		wmcs[i], err = d.generateIP(ctx, rmc)
		if err != nil {
			return err
		}
	}

	// Put machines into etcd
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(d.prefix, KeyMachines, wmc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			return err
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}

	tresp, err := d.Txn(ctx).
		If(
			txnIfOps...,
		).
		Then(
			txnThenOps...,
		).
		Else().
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		return errors.New("transaction failed")
	}
	return nil
}

func (d *Driver) Query(ctx context.Context, query *sabakan.Query) ([]*sabakan.Machine, error) {

}

func (d *Driver) Delete(ctx context.Context, serials []string) error {

}
