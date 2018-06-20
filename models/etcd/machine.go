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

func (d *driver) machineRegister(ctx context.Context, machines []*sabakan.Machine) error {
	cfg, err := d.getIPAMConfig()
	if err != nil {
		return err
	}
RETRY:
	// Assign node indices and addresses temporarily
	usageMap, err := d.assignNodeIndex(ctx, machines, cfg)
	if err != nil {
		return err
	}
	for _, m := range machines {
		cfg.GenerateIP(m)
	}

	tresp, err := d.machineDoRegister(ctx, machines, usageMap)
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		goto RETRY
	}
	if !tresp.Responses[0].GetResponseTxn().Succeeded {
		// inner If, i.e. conflictMachinesIfOps, evaluated to false
		return sabakan.ErrConflicted
	}

	return nil
}

func (d *driver) machineDoRegister(ctx context.Context, wmcs []*sabakan.Machine, usageMap map[uint]*rackIndexUsage) (*clientv3.TxnResponse, error) {
	// Put machines into etcd
	conflictMachinesIfOps := []clientv3.Cmp{}
	usageCASIfOps := []clientv3.Cmp{}
	txnThenOps := []clientv3.Op{}
	for _, wmc := range wmcs {
		key := path.Join(KeyMachines, wmc.Spec.Serial)
		conflictMachinesIfOps = append(conflictMachinesIfOps, clientv3util.KeyMissing(key))
		j, err := json.Marshal(wmc)
		if err != nil {
			return nil, err
		}
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}
	for rack, usage := range usageMap {
		key := d.indexInRackKey(rack)
		j, err := json.Marshal(usage)
		if err != nil {
			return nil, err
		}

		usageCASIfOps = append(usageCASIfOps, clientv3.Compare(clientv3.ModRevision(key), "=", usage.revision))
		txnThenOps = append(txnThenOps, clientv3.OpPut(key, string(j)))
	}

	return d.client.Txn(ctx).
		If(usageCASIfOps...).
		Then(clientv3.OpTxn(conflictMachinesIfOps, txnThenOps, nil)).
		Commit()
}

func (d *driver) machineGetWithRev(ctx context.Context, serial string) (*sabakan.Machine, int64, error) {
	key := KeyMachines + serial

	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, 0, err
	}

	if resp.Count == 0 {
		return nil, 0, sabakan.ErrNotFound
	}

	m := new(sabakan.Machine)
	err = json.Unmarshal(resp.Kvs[0].Value, m)
	if err != nil {
		return nil, 0, err
	}

	return m, resp.Kvs[0].ModRevision, nil
}

func (d *driver) machineGet(ctx context.Context, serial string) (*sabakan.Machine, error) {
	m, _, err := d.machineGetWithRev(ctx, serial)
	return m, err
}

func (d *driver) machineSetState(ctx context.Context, serial string, state sabakan.MachineState) error {
	key := KeyMachines + serial

RETRY:
	m, rev, err := d.machineGetWithRev(ctx, serial)
	if err != nil {
		return err
	}

	err = m.SetState(state)
	if err != nil {
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	tresp, err := d.client.Txn(ctx).
		If(clientv3.Compare(clientv3.ModRevision(key), "=", rev)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return err
	}
	if !tresp.Succeeded {
		goto RETRY
	}
	return nil
}

func (d *driver) machineQuery(ctx context.Context, q sabakan.Query) ([]*sabakan.Machine, error) {
	var serials []string

	switch {
	case q.IsEmpty():
		resp, err := d.client.Get(ctx, KeyMachines, clientv3.WithPrefix(), clientv3.WithKeysOnly())
		if err != nil {
			return nil, err
		}
		serials = make([]string, resp.Count)
		for i, kv := range resp.Kvs {
			serials[i] = string(kv.Key[len(KeyMachines):])
		}
	case len(q.Serial()) > 0:
		serials = []string{q.Serial()}
	default:
		serials = d.mi.query(q)
	}

	res := make([]*sabakan.Machine, 0, len(serials))
	for _, serial := range serials {
		key := path.Join(KeyMachines, serial)
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

func (d *driver) machineDelete(ctx context.Context, serial string) error {
RETRY:
	m, rev, err := d.machineGetWithRev(ctx, serial)
	if err != nil {
		return err
	}

	if m.Status.State != sabakan.StateRetired {
		return errors.New("non-retired machine cannot be deleted")
	}

	usage, err := d.getRackIndexUsage(ctx, m.Spec.Rack)
	if err != nil {
		return err
	}
	needUpdate := usage.release(m)
	if !needUpdate {
		return nil
	}

	resp, err := d.machineDoDelete(ctx, m, rev, usage)
	if err != nil {
		return err
	}

	if !resp.Succeeded {
		goto RETRY
	}

	return nil
}

func (d *driver) machineDoDelete(ctx context.Context, machine *sabakan.Machine, rev int64,
	usage *rackIndexUsage) (*clientv3.TxnResponse, error) {

	machineKey := KeyMachines + machine.Spec.Serial
	indexKey := d.indexInRackKey(machine.Spec.Rack)

	j, err := json.Marshal(usage)
	if err != nil {
		return nil, err
	}

	return d.client.Txn(ctx).
		If(
			clientv3.Compare(clientv3.ModRevision(machineKey), "=", rev),
			clientv3.Compare(clientv3.ModRevision(indexKey), "=", usage.revision),
		).
		Then(
			clientv3.OpDelete(machineKey),
			clientv3.OpPut(indexKey, string(j)),
		).
		Commit()
}

type machineDriver struct {
	*driver
}

// Register implements sabakan.MachineModel
func (d machineDriver) Register(ctx context.Context, machines []*sabakan.Machine) error {
	return d.machineRegister(ctx, machines)
}

// Get implements sabakan.MachineModel
func (d machineDriver) Get(ctx context.Context, serial string) (*sabakan.Machine, error) {
	return d.machineGet(ctx, serial)
}

// SetState implements sabakan.MachineModel
func (d machineDriver) SetState(ctx context.Context, serial string, state sabakan.MachineState) error {
	return d.machineSetState(ctx, serial, state)
}

// Query implements sabakan.MachineModel
func (d machineDriver) Query(ctx context.Context, query sabakan.Query) ([]*sabakan.Machine, error) {
	return d.machineQuery(ctx, query)
}

// Delete implements sabakan.MachineModel
func (d machineDriver) Delete(ctx context.Context, serial string) error {
	return d.machineDelete(ctx, serial)
}
