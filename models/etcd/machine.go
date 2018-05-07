package etcd

import (
	"context"
	"encoding/json"
	"path"

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

	tresp, err := d.client.Txn(ctx).
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
		return sabakan.ErrConflicted
	}
	return nil
}

func (d *Driver) Query(ctx context.Context, q *sabakan.Query) ([]*sabakan.Machine, error) {
	var serials []string
	if len(q.Serial) > 0 {
		serials = []string{q.Serial}
	} else {
		serials = d.mi.query(q)
	}

	res := make([]*sabakan.Machine, 0, len(serials))
	for _, serial := range serials {
		key := path.Join(d.prefix, KeyMachines, serial)
		resp, err := d.client.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		if resp.Count == 0 {
			continue
		}

		var mj sabakan.MachineJson
		err = json.Unmarshal(resp.Kvs[0].Value, &mj)
		if err != nil {
			return nil, err
		}

		m := mj.ToMachine()
		if q.Match(m) {
			res = append(res, m)
		}
	}

	if len(res) == 0 {
		return nil, nil
	}

	return res, nil
}

func (d *Driver) Delete(ctx context.Context, serials []string) error {
	return nil
}
