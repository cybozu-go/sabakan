package etcd

import (
	"context"
	"encoding/json"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

func (d *Driver) getNodeIndex(ctx context.Context, machine *sabakan.Machine) (uint32, error) {
	key := path.Join(d.prefix, KeyNodeIndices, string(machine.Rack)) + "/"

	resp, err := d.client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}

	if len(resp.Kvs) == 0 {
		return 0, errors.New("no node index is available for new machine")
	}

	return 0, nil
}

func releaseNodeIndices(is []uint32) {
	return
}

// Register implements sabakan.MachineModel
func (d *Driver) Register(ctx context.Context, machines []*sabakan.Machine) error {

	wmcs := make([]*sabakan.MachineJSON, len(machines))
	nodeIndices := []uint32{}

	cfg, err := d.GetConfig(ctx)
	if err != nil {
		return err
	}

	for i, rmc := range machines {
		nodeIndex, err := d.getNodeIndex(ctx, rmc)
		if err != nil {
			releaseNodeIndices(nodeIndices)
			return err
		}
		nodeIndices = append(nodeIndices, nodeIndex)
		//rmc.NodeIndexInRack = nodeIndex

		cfg.GenerateIP(rmc)
		wmcs[i] = rmc.ToJSON()
	}

	// Put machines into etcd
	txnIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(d.prefix, KeyMachines, wmc.Serial)
		txnIfOps = append(txnIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			releaseNodeIndices(nodeIndices)
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
		releaseNodeIndices(nodeIndices)
		return err
	}
	if !tresp.Succeeded {
		releaseNodeIndices(nodeIndices)
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
	key := path.Join(d.prefix, KeyMachines, serial)

	resp, err := d.client.Delete(ctx, key)
	if err != nil {
		return err
	}

	if resp.Deleted == 0 {
		return sabakan.ErrNotFound
	}

	return nil
}
