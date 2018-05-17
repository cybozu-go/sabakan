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

func newMachinesIndex() *machinesIndex {
	return &machinesIndex{
		Product:    make(map[string][]string),
		Datacenter: make(map[string][]string),
		Rack:       make(map[string][]string),
		Role:       make(map[string][]string),
		IPv4:       make(map[string]string),
		IPv6:       make(map[string]string),
	}
}

func (mi *machinesIndex) init(ctx context.Context, client *clientv3.Client, prefix string) error {
	key := path.Join(prefix, KeyMachines)
	resp, err := client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if resp.Count == 0 {
		return nil
	}

	f := func(val []byte) (*sabakan.Machine, error) {
		var mc sabakan.Machine
		err := json.Unmarshal(val, &mc)
		if err != nil {
			return nil, err
		}
		return &mc, nil
	}
	for _, kv := range resp.Kvs {
		m, err := f(kv.Value)
		if err != nil {
			return err
		}
		mi.AddIndex(m)
	}
	return nil
}

func (mi *machinesIndex) AddIndex(m *sabakan.Machine) {
	mi.mux.Lock()
	mi.addNoLock(m)
	mi.mux.Unlock()
}

func (mi *machinesIndex) addNoLock(m *sabakan.Machine) {
	mi.Product[m.Product] = append(mi.Product[m.Product], m.Serial)
	mi.Datacenter[m.Datacenter] = append(mi.Datacenter[m.Datacenter], m.Serial)
	mcrack := fmt.Sprint(m.Rack)
	mi.Rack[mcrack] = append(mi.Rack[mcrack], m.Serial)
	mi.Role[m.Role] = append(mi.Role[m.Role], m.Serial)
	for _, ifn := range m.Network {
		for _, v := range ifn.IPv4 {
			mi.IPv4[v] = m.Serial
		}
		for _, v := range ifn.IPv6 {
			mi.IPv6[v] = m.Serial
		}
	}
	if len(m.BMC.IPv4) > 0 {
		mi.IPv4[m.BMC.IPv4] = m.Serial
	}
	if len(m.BMC.IPv6) > 0 {
		mi.IPv6[m.BMC.IPv6] = m.Serial
	}
}

func indexOf(data []string, element string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	panic("element not found")
}

func (mi *machinesIndex) DeleteIndex(m *sabakan.Machine) {
	mi.mux.Lock()
	mi.deleteNoLock(m)
	mi.mux.Unlock()
}

func (mi *machinesIndex) deleteNoLock(m *sabakan.Machine) {
	i := indexOf(mi.Product[m.Product], m.Serial)
	mi.Product[m.Product] = append(mi.Product[m.Product][:i], mi.Product[m.Product][i+1:]...)
	i = indexOf(mi.Datacenter[m.Datacenter], m.Serial)
	mi.Datacenter[m.Datacenter] = append(mi.Datacenter[m.Datacenter][:i], mi.Datacenter[m.Datacenter][i+1:]...)
	mcrack := fmt.Sprint(m.Rack)
	i = indexOf(mi.Rack[mcrack], m.Serial)
	mi.Rack[mcrack] = append(mi.Rack[mcrack][:i], mi.Rack[mcrack][i+1:]...)
	i = indexOf(mi.Role[m.Role], m.Serial)
	mi.Role[m.Role] = append(mi.Role[m.Role][:i], mi.Role[m.Role][i+1:]...)
	for _, ifn := range m.Network {
		for _, v := range ifn.IPv4 {
			delete(mi.IPv4, v)
		}
		for _, v := range ifn.IPv6 {
			delete(mi.IPv6, v)
		}
	}
	delete(mi.IPv4, m.BMC.IPv4)
	delete(mi.IPv6, m.BMC.IPv6)
}

// UpdateIndex updates target machine on the index
func (mi *machinesIndex) UpdateIndex(prevM *sabakan.Machine, newM *sabakan.Machine) {
	mi.mux.Lock()
	mi.deleteNoLock(prevM)
	mi.addNoLock(newM)
	mi.mux.Unlock()
}

func (d *driver) startWatching(ctx context.Context) error {
	err := d.mi.init(ctx, d.watcher, d.prefix)
	if err != nil {
		return err
	}

	f := func(val []byte) (*sabakan.Machine, error) {
		var mc sabakan.Machine
		err := json.Unmarshal(val, &mc)
		if err != nil {
			return nil, err
		}
		return &mc, nil
	}

	key := path.Join(d.prefix, KeyMachines)
	rch := d.watcher.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithPrevKV())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT && ev.PrevKv != nil {
				prevM, err := f(ev.PrevKv.Value)
				if err != nil {
					panic(err)
				}
				newM, err := f(ev.Kv.Value)
				if err != nil {
					panic(err)
				}
				d.mi.UpdateIndex(prevM, newM)
			}
			if ev.Type == mvccpb.PUT && ev.PrevKv == nil {
				m, err := f(ev.Kv.Value)
				if err != nil {
					panic(err)
				}
				d.mi.AddIndex(m)
			}
			if ev.Type == mvccpb.DELETE {
				m, err := f(ev.PrevKv.Value)
				if err != nil {
					panic(err)
				}
				d.mi.DeleteIndex(m)
			}
		}
	}
	return nil
}

func (mi *machinesIndex) query(q *sabakan.Query) []string {
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
