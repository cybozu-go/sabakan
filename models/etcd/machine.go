package etcd

import (
	"context"
	"encoding/json"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	"github.com/cybozu-go/sabakan"
)

// Register implements sabakan.MachineModel
func (d *Driver) Register(ctx context.Context, machines []*sabakan.Machine) error {
	wmcs := make([]*sabakan.MachineJSON, len(machines))

	cfg, err := d.GetConfig(ctx)
	if err != nil {
		return err
	}

RETRY:
	err = d.assignNodeIndex(ctx, machines)
	if err != nil {
		return err
	}
	for i, rmc := range machines {
		cfg.GenerateIP(rmc)
		wmcs[i] = rmc.ToJSON()
	}

	// Put machines into etcd
	conflictMachinesIfOps := []clientv3.Cmp{}
	availableNodeIndexIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(d.prefix, KeyMachines, wmc.Serial)
		indexKey := d.getNodeIndexKey(*wmc.Rack, *wmc.NodeIndexInRack)
		conflictMachinesIfOps = append(conflictMachinesIfOps, clientv3util.KeyMissing(key))
		availableNodeIndexIfOps = append(availableNodeIndexIfOps, clientv3util.KeyExists(indexKey))
		j, err := json.Marshal(wmc)
		if err != nil {
			return err
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
		txnThenOps = append(txnThenOps, clientv3.OpDelete(indexKey))
	}

	tresp, err := d.client.Txn(ctx).
		If(
			availableNodeIndexIfOps...,
		).
		Then(
			clientv3.OpTxn(conflictMachinesIfOps, txnThenOps, nil),
		).
		Else().
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		goto RETRY
	}
	if !tresp.Responses[0].Response.(*etcdserverpb.ResponseOp_ResponseTxn).ResponseTxn.Succeeded {
		return sabakan.ErrConflicted
	}

	return nil
}

// Query implements sabakan.MachineModel
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

		var mj sabakan.MachineJSON
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

// Delete implements sabakan.MachineModel
func (d *Driver) Delete(ctx context.Context, serial string) error {
	machines, err := d.Query(ctx, sabakan.QueryBySerial(serial))
	if err != nil {
		return err
	}
	if len(machines) != 1 {
		return sabakan.ErrNotFound
	}
	indexKey := d.getNodeIndexKey(machines[0].Rack, machines[0].NodeIndexInRack)
	indexValue := encodeNodeIndex(machines[0].NodeIndexInRack)

	key := path.Join(d.prefix, KeyMachines, serial)

	resp, err := d.client.Txn(ctx).
		If(clientv3util.KeyExists(key)).
		Then(clientv3.OpDelete(key), clientv3.OpPut(indexKey, indexValue)).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		return sabakan.ErrNotFound
	}

	return nil
}
