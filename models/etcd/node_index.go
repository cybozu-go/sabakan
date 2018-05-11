package etcd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
	"github.com/cybozu-go/sabakan"
)

type assignedIndex struct {
	rack  uint
	index uint
}

func (d *Driver) getNodeIndexKey(rack, index uint) string {
	return path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rack), fmt.Sprintf("%02d", index))
}

func (d *Driver) getNodeIndicesInRackKey(rack uint) string {
	return path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rack)) + "/"
}

func encodeNodeIndex(index uint) string {
	return fmt.Sprint(index)
}

func decodeNodeIndex(indexString string) (uint, error) {
	index, err := strconv.Atoi(indexString)
	if err != nil {
		return uint(0), err
	}
	return uint(index), nil
}

func (d *Driver) assignNodeIndex(ctx context.Context, machine *sabakan.Machine) error {
	key := d.getNodeIndicesInRackKey(machine.Rack)
	resp, err := d.client.Get(ctx, key, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	if err != nil {
		return err
	}
	if len(resp.Kvs) == 0 {
		return errors.New("no node index is available for new machine")
	}

	// indices retrieved above may be deleted by others, so ensure index by delete
	for _, kv := range resp.Kvs {
		dresp, err := d.client.Delete(ctx, string(kv.Key))
		if err != nil {
			return err
		}
		if dresp.Deleted > 0 {
			nodeIndex, err := decodeNodeIndex(string(kv.Value))
			if err != nil {
				return err
			}
			machine.NodeIndexInRack = nodeIndex
			return nil
		}
	}

	return errors.New("no node index is available for new machine")
}

func (d *Driver) releaseNodeIndices(ctx context.Context, assignedIndices []assignedIndex) {
	for _, assigned := range assignedIndices {
		key := d.getNodeIndexKey(assigned.rack, assigned.index)
		_, err := d.client.Put(ctx, key, encodeNodeIndex(assigned.index))
		if err != nil {
			log.Error("failed to release node index", map[string]interface{}{
				log.FnError: err,
				"rack":      assigned.rack,
				"index":     assigned.index,
			})
		}
	}
}
