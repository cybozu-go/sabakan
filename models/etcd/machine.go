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
func (d *driver) Register(ctx context.Context, machines []*sabakan.Machine) error {
	wmcs := make([]*sabakan.Machine, len(machines))

	cfg, err := d.getIPAMConfig()
	if err != nil {
		return err
	}

RETRY:
	// Assign node indices temporarily
	usageMap, err := d.assignNodeIndex(ctx, machines, cfg)
	if err != nil {
		return err
	}
	for i, rmc := range machines {
		cfg.GenerateIP(rmc)
		wmcs[i] = rmc
	}

	// Put machines into etcd
	conflictMachinesIfOps := []clientv3.Cmp{}
	usageCASIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(d.prefix, KeyMachines, wmc.Serial)
		conflictMachinesIfOps = append(conflictMachinesIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			return err
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}
	for rack, usage := range usageMap {
		key := d.indexInRackKey(rack)
		j, err := json.Marshal(usage)
		if err != nil {
			return err
		}
		usageCASIfOps = append(usageCASIfOps, clientv3.Compare(clientv3.ModRevision(key), "=", usage.revision))
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}

	tresp, err := d.client.Txn(ctx).
		If(
			usageCASIfOps...,
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
		// usageCASIfOps evaluated to false; index usage was updated by another txn.
		goto RETRY
	}
	if !tresp.Responses[0].Response.(*etcdserverpb.ResponseOp_ResponseTxn).ResponseTxn.Succeeded {
		// conflictMachinesIfOps evaluated to false
		return sabakan.ErrConflicted
	}

	return nil
}

// Query implements sabakan.MachineModel
func (d *driver) Query(ctx context.Context, q *sabakan.Query) ([]*sabakan.Machine, error) {
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

		m := new(sabakan.Machine)
		err = json.Unmarshal(resp.Kvs[0].Value, m)
		if err != nil {
			return nil, err
		}

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
func (d *driver) Delete(ctx context.Context, serial string) error {
	machines, err := d.Query(ctx, sabakan.QueryBySerial(serial))
	if err != nil {
		return err
	}
	if len(machines) != 1 {
		return sabakan.ErrNotFound
	}

	m := machines[0]

	machineKey := path.Join(d.prefix, KeyMachines, serial)
	indexKey := d.indexInRackKey(m.Rack)

RETRY:
	usage, err := d.getRackIndexUsage(ctx, m.Rack)
	if err != nil {
		return err
	}
	usage.release(m)

	j, err := json.Marshal(usage)
	if err != nil {
		return err
	}

	resp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(indexKey), "=", usage.revision)).
		Then(
			clientv3.OpTxn(
				[]clientv3.Cmp{clientv3util.KeyExists(machineKey)},
				[]clientv3.Op{
					clientv3.OpDelete(machineKey),
					clientv3.OpPut(indexKey, string(j)),
				},
				nil),
		).
		Else().
		Commit()
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		// revision mismatch
		goto RETRY
	}

	if !resp.Responses[0].Response.(*etcdserverpb.ResponseOp_ResponseTxn).ResponseTxn.Succeeded {
		// KeyExists(machineKey) failed
		return sabakan.ErrNotFound
	}

	return nil
}
