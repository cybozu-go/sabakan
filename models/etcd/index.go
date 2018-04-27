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
	mux        sync.Mutex
	Product    map[string][]string
	Datacenter map[string][]string
	Rack       map[string][]string
	Role       map[string][]string
	Cluster    map[string][]string
	IPv4       map[string]string
	IPv6       map[string]string
}

func (mi *machinesIndex) init(ctx context.Context, client *clientv3.Client, prefix string) error {
	mi.Product = map[string][]string{}
	mi.Datacenter = map[string][]string{}
	mi.Rack = map[string][]string{}
	mi.Role = map[string][]string{}
	mi.Cluster = map[string][]string{}
	mi.IPv4 = map[string]string{}
	mi.IPv6 = map[string]string{}

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

// AddIndex adds new machine into the index
func (mi *machinesIndex) AddIndex(val []byte) error {
	var mc sabakan.Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return err
	}
	mi.mux.Lock()
	defer mi.mux.Unlock()

	mi.Product[mc.Product] = append(mi.Product[mc.Product], mc.Serial)
	mi.Datacenter[mc.Datacenter] = append(mi.Datacenter[mc.Datacenter], mc.Serial)
	mcrack := fmt.Sprint(mc.Rack)
	mi.Rack[mcrack] = append(mi.Rack[mcrack], mc.Serial)
	mi.Role[mc.Role] = append(mi.Role[mc.Role], mc.Serial)
	mi.Cluster[mc.Cluster] = append(mi.Cluster[mc.Cluster], mc.Serial)
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

// DeleteIndex deletes a machine from the index
func (mi *machinesIndex) DeleteIndex(val []byte) error {
	var mc sabakan.Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return err
	}
	mi.mux.Lock()
	defer mi.mux.Unlock()

	i := indexOf(mi.Product[mc.Product], mc.Serial)
	mi.Product[mc.Product] = append(mi.Product[mc.Product][:i], mi.Product[mc.Product][i+1:]...)
	i = indexOf(mi.Datacenter[mc.Datacenter], mc.Serial)
	mi.Datacenter[mc.Datacenter] = append(mi.Datacenter[mc.Datacenter][:i], mi.Datacenter[mc.Datacenter][i+1:]...)
	mcrack := fmt.Sprint(mc.Rack)
	i = indexOf(mi.Rack[mcrack], mc.Serial)
	mi.Rack[mcrack] = append(mi.Rack[mcrack][:i], mi.Rack[mcrack][i+1:]...)
	i = indexOf(mi.Role[mc.Role], mc.Serial)
	mi.Role[mc.Role] = append(mi.Role[mc.Role][:i], mi.Role[mc.Role][i+1:]...)
	i = indexOf(mi.Cluster[mc.Cluster], mc.Serial)
	mi.Cluster[mc.Cluster] = append(mi.Cluster[mc.Cluster][:i], mi.Cluster[mc.Cluster][i+1:]...)
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
	err := mi.DeleteIndex(pval)
	if err != nil {
		return err
	}
	return mi.AddIndex(nval)
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
