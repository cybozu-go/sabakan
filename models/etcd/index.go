package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/cybozu-go/sabakan/v2"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	labelSep = "="
)

// machinesIndex is on-memory index of the etcd values
type machinesIndex struct {
	mux     sync.RWMutex
	Rack    map[string][]string
	Role    map[string][]string
	Labels  map[string][]string
	IPv4    map[string]string
	IPv6    map[string]string
	BMCType map[string][]string
	State   map[sabakan.MachineState][]string
}

func newMachinesIndex() *machinesIndex {
	return &machinesIndex{
		Rack:    make(map[string][]string),
		Role:    make(map[string][]string),
		Labels:  make(map[string][]string),
		IPv4:    make(map[string]string),
		IPv6:    make(map[string]string),
		BMCType: make(map[string][]string),
		State:   make(map[sabakan.MachineState][]string),
	}
}

func (mi *machinesIndex) init(ctx context.Context, client *clientv3.Client) error {
	resp, err := client.Get(ctx, KeyMachines, clientv3.WithPrefix())
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
	spec := &m.Spec
	mcrack := fmt.Sprint(spec.Rack)
	mi.Rack[mcrack] = append(mi.Rack[mcrack], spec.Serial)
	mi.Role[spec.Role] = append(mi.Role[spec.Role], spec.Serial)
	mi.BMCType[spec.BMC.Type] = append(mi.BMCType[spec.BMC.Type], spec.Serial)
	for _, ip := range spec.IPv4 {
		mi.IPv4[ip] = spec.Serial
	}
	for _, ip := range spec.IPv6 {
		mi.IPv6[ip] = spec.Serial
	}
	if len(spec.BMC.IPv4) > 0 {
		mi.IPv4[spec.BMC.IPv4] = spec.Serial
	}
	if len(spec.BMC.IPv6) > 0 {
		mi.IPv6[spec.BMC.IPv6] = spec.Serial
	}
	mi.State[m.Status.State] = append(mi.State[m.Status.State], spec.Serial)
	for k, v := range spec.Labels {
		labelKey := k + labelSep + v
		mi.Labels[labelKey] = append(mi.Labels[labelKey], spec.Serial)
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
	spec := &m.Spec
	mcrack := fmt.Sprint(spec.Rack)
	i := indexOf(mi.Rack[mcrack], spec.Serial)
	mi.Rack[mcrack] = append(mi.Rack[mcrack][:i], mi.Rack[mcrack][i+1:]...)
	i = indexOf(mi.Role[spec.Role], spec.Serial)
	mi.Role[spec.Role] = append(mi.Role[spec.Role][:i], mi.Role[spec.Role][i+1:]...)
	i = indexOf(mi.BMCType[spec.BMC.Type], spec.Serial)
	mi.BMCType[spec.BMC.Type] = append(mi.BMCType[spec.BMC.Type][:i], mi.BMCType[spec.BMC.Type][i+1:]...)
	for _, ip := range spec.IPv4 {
		delete(mi.IPv4, ip)
	}
	for _, ip := range spec.IPv6 {
		delete(mi.IPv6, ip)
	}
	delete(mi.IPv4, spec.BMC.IPv4)
	delete(mi.IPv6, spec.BMC.IPv6)

	i = indexOf(mi.State[m.Status.State], spec.Serial)
	mi.State[m.Status.State] = append(mi.State[m.Status.State][:i], mi.State[m.Status.State][i+1:]...)
	for k, v := range spec.Labels {
		labelKey := k + labelSep + v
		i = indexOf(mi.Labels[labelKey], spec.Serial)
		mi.Labels[labelKey] = append(mi.Labels[labelKey][:i], mi.Labels[labelKey][i+1:]...)
	}
}

// UpdateIndex updates target machine on the index
func (mi *machinesIndex) UpdateIndex(prevM *sabakan.Machine, newM *sabakan.Machine) {
	mi.mux.Lock()
	mi.deleteNoLock(prevM)
	mi.addNoLock(newM)
	mi.mux.Unlock()
}

func (mi *machinesIndex) query(q sabakan.Query) []string {
	mi.mux.RLock()
	defer mi.mux.RUnlock()

	res := make(map[string]struct{})

	for _, rack := range strings.Split(q.Rack(), ",") {
		for _, serial := range mi.Rack[rack] {
			res[serial] = struct{}{}
		}
	}
	for _, role := range strings.Split(q.Role(), ",") {
		for _, serial := range mi.Role[role] {
			res[serial] = struct{}{}
		}
	}
	for _, ipv4 := range strings.Split(q.IPv4(), ",") {
		if serial, ok := mi.IPv4[ipv4]; len(ipv4) > 0 && ok {
			res[serial] = struct{}{}
		}
	}
	for _, ipv6 := range strings.Split(q.IPv6(), ",") {
		if serial, ok := mi.IPv6[ipv6]; len(ipv6) > 0 && ok {
			res[serial] = struct{}{}
		}
	}
	for _, bmcType := range strings.Split(q.BMCType(), ",") {
		for _, serial := range mi.BMCType[bmcType] {
			res[serial] = struct{}{}
		}
	}
	for _, state := range strings.Split(q.State(), ",") {
		for _, serial := range mi.State[sabakan.MachineState(state)] {
			res[serial] = struct{}{}
		}
	}
	for _, labelKey := range q.Labels() {
		for _, serial := range mi.Labels[labelKey] {
			res[serial] = struct{}{}
		}
	}

	serials := make([]string, 0, len(res))
	for serial := range res {
		serials = append(serials, serial)
	}
	return serials
}

func decodeMachine(val []byte) (*sabakan.Machine, error) {
	var mc sabakan.Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return nil, err
	}
	return &mc, nil
}

func (d *driver) handleMachines(ev *clientv3.Event) error {
	switch {
	case ev.IsCreate():
		m, err := decodeMachine(ev.Kv.Value)
		if err != nil {
			return err
		}
		d.mi.AddIndex(m)
	case ev.IsModify():
		prevM, err := decodeMachine(ev.PrevKv.Value)
		if err != nil {
			return err
		}
		newM, err := decodeMachine(ev.Kv.Value)
		if err != nil {
			return err
		}
		d.mi.UpdateIndex(prevM, newM)
	default: // DELETE
		m, err := decodeMachine(ev.PrevKv.Value)
		if err != nil {
			return err
		}
		d.mi.DeleteIndex(m)
	}

	return nil
}
