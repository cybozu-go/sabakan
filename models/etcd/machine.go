package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
	"github.com/pkg/errors"
)

type assignedIndex struct {
	rack  uint32
	index uint32
}

func (d *Driver) assignNodeIndex(ctx context.Context, machine *sabakan.Machine) error {
	key := path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(machine.Rack)) + "/"
	fmt.Println(key)

	resp, err := d.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}

	if len(resp.Kvs) == 0 {
		fmt.Println("foo")
		return errors.New("no node index is available for new machine")
	}

	for _, kv := range resp.Kvs {
		dresp, err := d.client.Delete(ctx, string(kv.Key))
		if err != nil {
			return err
		}
		if dresp.Deleted > 0 {
			nodeIndex, err := strconv.Atoi(string(kv.Value))
			if err != nil {
				return err
			}
			machine.NodeIndexInRack = uint32(nodeIndex)
			return nil
		}
	}

	return errors.New("no node index is available for new machine")
}

func (d *Driver) releaseNodeIndices(ctx context.Context, nodeIndices []assignedIndex) {
	for _, nodeIndex := range nodeIndices {
		key := path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(nodeIndex.rack), fmt.Sprintf("%02d", nodeIndex.index))
		_, err := d.client.Put(ctx, key, fmt.Sprint(nodeIndex.index))
		if err != nil {
			log.Error("failed to release node index", map[string]interface{}{
				log.FnError: err,
				"rack":      nodeIndex.rack,
				"index":     nodeIndex.index,
			})
		}
	}
}

// Register implements sabakan.MachineModel
func (d *Driver) Register(ctx context.Context, machines []*sabakan.Machine) error {

	wmcs := make([]*sabakan.MachineJSON, len(machines))
	nodeIndices := []assignedIndex{}

	cfg, err := d.GetConfig(ctx)
	if err != nil {
		return err
	}

	for i, rmc := range machines {
		err = d.assignNodeIndex(ctx, rmc)
		if err != nil {
			d.releaseNodeIndices(ctx, nodeIndices)
			return err
		}
		nodeIndices = append(nodeIndices, assignedIndex{rack: rmc.Rack, index: rmc.NodeIndexInRack})

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
			d.releaseNodeIndices(ctx, nodeIndices)
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
		d.releaseNodeIndices(ctx, nodeIndices)
		return err
	}
	if !tresp.Succeeded {
		d.releaseNodeIndices(ctx, nodeIndices)
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
	indexKey := path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(machines[0].Rack), fmt.Sprintf("%02d", machines[0].NodeIndexInRack))
	indexValue := fmt.Sprint(machines[0].NodeIndexInRack)

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
