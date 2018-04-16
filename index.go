package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sync"

	"github.com/coreos/etcd/clientv3"
)

// MachinesIndex is on-memory index of the etcd values
type MachinesIndex struct {
	Product    map[string][]string
	Datacenter map[string][]string
	Rack       map[string][]string
	Role       map[string][]string
	Cluster    map[string][]string
	IPv4       map[string]string
	IPv6       map[string]string
	Wg         sync.WaitGroup
}

// Indexing is indexing machineIndex
func Indexing(client *clientv3.Client, prefix string) (MachinesIndex, error) {
	var mi MachinesIndex
	mi.Product = map[string][]string{}
	mi.Datacenter = map[string][]string{}
	mi.Rack = map[string][]string{}
	mi.Role = map[string][]string{}
	mi.Cluster = map[string][]string{}
	mi.IPv4 = map[string]string{}
	mi.IPv6 = map[string]string{}

	key := path.Join(prefix, EtcdKeyMachines)
	resp, err := client.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return mi, err
	}
	if resp.Count == 0 {
		return mi, nil
	}
	for _, m := range resp.Kvs {
		err := mi.AddIndex(m.Value)
		if err != nil {
			return mi, err
		}
	}
	return mi, nil
}

// AddIndex adds new machine into the index
func (mi *MachinesIndex) AddIndex(val []byte) error {
	var mc Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return err
	}

	mi.Product[mc.Product] = append(mi.Product[mc.Product], mc.Serial)
	mi.Datacenter[mc.Datacenter] = append(mi.Datacenter[mc.Datacenter], mc.Serial)
	mcrack := fmt.Sprint(mc.Rack)
	mi.Rack[mcrack] = append(mi.Rack[mcrack], mc.Serial)
	mi.Role[mc.Role] = append(mi.Role[mc.Role], mc.Serial)
	mi.Cluster[mc.Cluster] = append(mi.Cluster[mc.Cluster], mc.Serial)
	for _, ifn := range mc.Network {
		for k, v := range ifn.(map[string]interface{}) {
			if k == "ipv4" {
				for _, ip := range v.([]interface{}) {
					mi.IPv4[ip.(string)] = mc.Serial
				}
			}
			if k == "ipv6" {
				for _, ip := range v.([]interface{}) {
					mi.IPv6[ip.(string)] = mc.Serial
				}
			}
		}
	}
	for k, v := range mc.BMC {
		if k == "ipv4" {
			for _, ip := range v.([]interface{}) {
				mi.IPv4[ip.(string)] = mc.Serial
			}
		}
		if k == "ipv6" {
			for _, ip := range v.([]interface{}) {
				mi.IPv6[ip.(string)] = mc.Serial
			}
		}
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
func (mi *MachinesIndex) DeleteIndex(val []byte) error {
	var mc Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		return err
	}

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
		for k, v := range ifn.(map[string]interface{}) {
			if k == "ipv4" {
				for _, ip := range v.([]interface{}) {
					delete(mi.IPv4, ip.(string))
				}
			}
			if k == "ipv6" {
				for _, ip := range v.([]interface{}) {
					delete(mi.IPv6, ip.(string))
				}
			}
		}
	}
	for k, v := range mc.BMC {
		if k == "ipv4" {
			for _, ip := range v.([]interface{}) {
				delete(mi.IPv4, ip.(string))
			}
		}
		if k == "ipv6" {
			for _, ip := range v.([]interface{}) {
				delete(mi.IPv6, ip.(string))
			}
		}
	}
	return nil
}

// UpdateIndex updates target machine on the index
func (mi *MachinesIndex) UpdateIndex(pval []byte, nval []byte) error {
	err := mi.DeleteIndex(pval)
	if err != nil {
		return err
	}
	err = mi.AddIndex(nval)
	if err != nil {
		return err
	}
	return nil
}
