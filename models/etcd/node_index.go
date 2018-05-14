package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/clientv3util"
	"github.com/cybozu-go/sabakan"
)

type rackIndexUsage struct {
	revision    int64
	usedIndices []uint
	indexMap    map[uint]bool
}

func (r rackIndexUsage) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(r.usedIndices)
	return data, err
}

func (r *rackIndexUsage) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &r.usedIndices)
	if err == nil {
		r.indexMap = make(map[uint]bool)
		for _, idx := range r.usedIndices {
			r.indexMap[idx] = true
		}
	}
	return err
}

func (r *rackIndexUsage) assign(m *sabakan.Machine, c *sabakan.IPAMConfig) error {
	var idx uint

OUT:
	switch m.Role {
	case "boot":
		idx = c.NodeIndexOffset
		if r.indexMap[idx] {
			return sabakan.ErrConflicted
		}
	default:
		for i := uint(0); i < c.MaxNodesInRack; i++ {
			idx = i + c.NodeIndexOffset + 1
			if !r.indexMap[idx] {
				break OUT
			}
		}
		return errors.New("no node index is available for new machine")
	}

	r.indexMap[idx] = true
	r.usedIndices = append(r.usedIndices, idx)
	m.IndexInRack = idx
	return nil
}

func (r *rackIndexUsage) release(m *sabakan.Machine) {
	if _, ok := r.indexMap[m.IndexInRack]; !ok {
		panic("inconsistent index map")
	}
	delete(r.indexMap, m.IndexInRack)

	used := make([]uint, 0, len(r.usedIndices)-1)
	for _, idx := range r.usedIndices {
		if idx == m.IndexInRack {
			continue
		}
		used = append(used, idx)
	}
	r.usedIndices = used
}

func (d *Driver) initializeNodeIndices(ctx context.Context, rack uint) error {
	var usage rackIndexUsage
	j, err := json.Marshal(usage)
	if err != nil {
		return err
	}

	key := d.indexInRackKey(rack)
	_, err = d.client.Txn(ctx).
		If(clientv3util.KeyMissing(key)).
		Then(clientv3.OpPut(key, string(j))).
		Else().
		Commit()

	return err
}

func (d *Driver) indexInRackKey(rack uint) string {
	return path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rack))
}

func (d *Driver) getRackIndexUsage(ctx context.Context, rack uint) (*rackIndexUsage, error) {
RETRY:
	key := d.indexInRackKey(rack)
	resp, err := d.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		err = d.initializeNodeIndices(ctx, rack)
		if err != nil {
			return nil, err
		}
		goto RETRY
	}

	kv := resp.Kvs[0]
	usage := new(rackIndexUsage)
	err = json.Unmarshal(kv.Value, usage)
	if err != nil {
		return nil, err
	}

	usage.revision = kv.ModRevision

	return usage, nil
}

func (d *Driver) assignNodeIndex(ctx context.Context, machines []*sabakan.Machine, config *sabakan.IPAMConfig) (map[uint]*rackIndexUsage, error) {
	usageMap := make(map[uint]*rackIndexUsage)
	for _, m := range machines {
		usage := usageMap[m.Rack]
		if usage == nil {
			u, err := d.getRackIndexUsage(ctx, m.Rack)
			if err != nil {
				return nil, err
			}
			usageMap[m.Rack] = u
			usage = u
		}

		err := usage.assign(m, config)
		if err != nil {
			return nil, err
		}
	}

	return usageMap, nil
}
