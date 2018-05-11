package etcd

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strconv"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
)

type assignedIndex struct {
	rack  uint
	index uint
}

func (d *Driver) getNodeIndexKey(machine *sabakan.Machine) (string, error) {
	switch machine.Role {
	case "worker":
		return d.getWorkerNodeIndexKey(machine.Rack, machine.NodeIndexInRack), nil
	case "boot":
		return d.getBootNodeIndexKey(machine.Rack), nil
	default:
		return "", errors.New("unknown role: " + machine.Role)
	}
}

func (d *Driver) getWorkerNodeIndexKey(rack, index uint) string {
	return path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rack), "worker", fmt.Sprintf("%02d", index))
}

func (d *Driver) getBootNodeIndexKey(rack uint) string {
	return path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rack), "boot")
}

func (d *Driver) getNodeIndicesInRackKey(rack uint) string {
	return path.Join(d.prefix, KeyNodeIndices, fmt.Sprint(rack), "worker") + "/"
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

func (d *Driver) assignNodeIndex(ctx context.Context, machines []*sabakan.Machine) error {
	bootServers := map[uint]*sabakan.Machine{}
	machinesGroupedByRack := map[uint][]*sabakan.Machine{}
	for _, m := range machines {
		switch m.Role {
		case "boot":
			if _, ok := bootServers[m.Rack]; ok {
				return sabakan.ErrConflicted
			}
			bootServers[m.Rack] = m
		case "worker":
			if _, ok := machinesGroupedByRack[m.Rack]; ok {
				machinesGroupedByRack[m.Rack] = append(machinesGroupedByRack[m.Rack], m)
			} else {
				machinesGroupedByRack[m.Rack] = []*sabakan.Machine{m}
			}
		default:
			return errors.New("unknown role: " + m.Role)
		}
	}

	for rack, ms := range machinesGroupedByRack {
		key := d.getNodeIndicesInRackKey(rack)
		resp, err := d.client.Get(
			ctx, key,
			clientv3.WithPrefix(),
			clientv3.WithLimit(int64(len(ms))),
			clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
		)
		if err != nil {
			return err
		}
		if len(resp.Kvs) < len(ms) {
			return errors.New("no node index is available for new machine")
		}

		for i, m := range ms {
			m.NodeIndexInRack, err = decodeNodeIndex(string(resp.Kvs[i].Value))
			if err != nil {
				return err
			}
		}
	}

	for rack, m := range bootServers {
		key := d.getBootNodeIndexKey(rack)
		resp, err := d.client.Get(ctx, key)
		if err != nil {
			return err
		}
		if len(resp.Kvs) == 0 {
			return fmt.Errorf("boot server has been registered for rack %d", rack)
		}

		m.NodeIndexInRack, err = decodeNodeIndex(string(resp.Kvs[0].Value))
		if err != nil {
			return err
		}
	}

	return nil
}
