package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/cybozu-go/sabakan"
)

// machinesIndex is on-memory index of the etcd values
type machinesIndex struct {
	mux        sync.RWMutex
	Product    map[string][]string
	Datacenter map[string][]string
	Rack       map[string][]string
	Role       map[string][]string
	IPv4       map[string]string
	IPv6       map[string]string
}

func (mi *machinesIndex) init(ctx context.Context, client *clientv3.Client, prefix string) error {
	mi.Product = make(map[string][]string)
	mi.Datacenter = make(map[string][]string)
	mi.Rack = make(map[string][]string)
	mi.Role = make(map[string][]string)
	mi.IPv4 = make(map[string]string)
	mi.IPv6 = make(map[string]string)

	key := path.Join(prefix, KeyMachines)
	resp, err := client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if resp.Count == 0 {
		return nil
	}
	for _, m := range resp.Kvs {
		err := mi.AddIndex(m.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mi *machinesIndex) AddIndex(val []byte) error {
	mi.mux.Lock()
	defer mi.mux.Unlock()
	return mi.addNoLock(val)
}

func (mi *machinesIndex) addNoLock(val []byte) error {
	var mc sabakan.MachineJson
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return err
	}

	mi.Product[mc.Product] = append(mi.Product[mc.Product], mc.Serial)
	mi.Datacenter[mc.Datacenter] = append(mi.Datacenter[mc.Datacenter], mc.Serial)
	mcrack := fmt.Sprint(*mc.Rack)
	mi.Rack[mcrack] = append(mi.Rack[mcrack], mc.Serial)
	mi.Role[mc.Role] = append(mi.Role[mc.Role], mc.Serial)
	for _, ifn := range mc.Network {
		for _, v := range ifn.IPv4 {
			mi.IPv4[v] = mc.Serial
		}
		for _, v := range ifn.IPv6 {
			mi.IPv6[v] = mc.Serial
		}
	}
	for _, v := range mc.BMC.IPv4 {
		mi.IPv4[v] = mc.Serial
	}
	return nil
}

func indexOf(data []string, element string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}

func (mi *machinesIndex) DeleteIndex(val []byte) error {
	mi.mux.Lock()
	defer mi.mux.Unlock()
	return mi.deleteNoLock(val)
}

func (mi *machinesIndex) deleteNoLock(val []byte) error {
	var mc sabakan.MachineJson
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return err
	}

	i := indexOf(mi.Product[mc.Product], mc.Serial)
	mi.Product[mc.Product] = append(mi.Product[mc.Product][:i], mi.Product[mc.Product][i+1:]...)
	i = indexOf(mi.Datacenter[mc.Datacenter], mc.Serial)
	mi.Datacenter[mc.Datacenter] = append(mi.Datacenter[mc.Datacenter][:i], mi.Datacenter[mc.Datacenter][i+1:]...)
	mcrack := fmt.Sprint(*mc.Rack)
	i = indexOf(mi.Rack[mcrack], mc.Serial)
	mi.Rack[mcrack] = append(mi.Rack[mcrack][:i], mi.Rack[mcrack][i+1:]...)
	i = indexOf(mi.Role[mc.Role], mc.Serial)
	mi.Role[mc.Role] = append(mi.Role[mc.Role][:i], mi.Role[mc.Role][i+1:]...)
	for _, ifn := range mc.Network {
		for _, v := range ifn.IPv4 {
			mi.IPv4[v] = ""
		}
		for _, v := range ifn.IPv6 {
			mi.IPv4[v] = ""
		}
	}
	for _, v := range mc.BMC.IPv4 {
		mi.IPv4[v] = ""
	}
	return nil
}

// UpdateIndex updates target machine on the index
// BUG: should hold a lock while updating.
func (mi *machinesIndex) UpdateIndex(pval []byte, nval []byte) error {
	mi.mux.Lock()
	defer mi.mux.Unlock()

	err := mi.deleteNoLock(pval)
	if err != nil {
		return err
	}
	return mi.addNoLock(nval)
}

func (d *Driver) startWatching(ctx context.Context) error {
	err := d.mi.init(ctx, d.watcher, d.prefix)
	if err != nil {
		return err
	}

	key := path.Join(d.prefix, KeyMachines)
	rch := d.watcher.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT && ev.PrevKv != nil {
				err := d.mi.UpdateIndex(ev.PrevKv.Value, ev.Kv.Value)
				if err != nil {
					return err
				}
			}
			if ev.Type == mvccpb.PUT && ev.PrevKv == nil {
				err := d.mi.AddIndex(ev.Kv.Value)
				if err != nil {
					return err
				}
			}
			if ev.Type == mvccpb.DELETE {
				err := d.mi.DeleteIndex(ev.PrevKv.Value)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (mi *machineIndex) query(q *sabakan.Query) []string {
	mi.mux.RLock()
	defer mi.mux.RUnlock()

	res := make(map[string]struct{})

	for _, serial := range mi.Product[q.Product] {
		res[serial] = struct{}{}
	}
	for _, serial := range mi.Datacenter[q.Datacenter] {
		res[serial] = struct{}{}
	}
	for _, serial := range mi.Rack[q.Rack] {
		res[serial] = struct{}{}
	}
	for _, serial := range mi.Role[q.Role] {
		res[serial] = struct{}{}
	}
	if len(q.IPv4) > 0 {
		if serial, ok := mi.IPv4[q.IPv4]; ok {
			res[serial] = struct{}{}
		}
	}
	if len(q.IPv6) > 0 {
		if serial, ok := mi.IPv6[q.IPv6]; ok {
			res[serial] = struct{}{}
		}
	}

	serials := make([]string, 0, len(res))
	for serial := range res {
		serials = append(serials, serial)
	}
	return serials
}
