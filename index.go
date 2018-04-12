package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
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
}

// MI is a variable of type MachinesIndex
var MI MachinesIndex

// Indexing is indexing MachineIndex
func Indexing(client *clientv3.Client, prefix string) {
	key := path.Join(prefix, EtcdKeyMachines)
	resp, err := client.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		log.ErrorExit(err)
	}
	if resp.Count == 0 {
		return
	}
	for _, m := range resp.Kvs {
		AddIndex(m.Value)
	}
}

func initMI() {
	if MI.Product == nil {
		MI.Product = map[string][]string{}
	}
	if MI.Datacenter == nil {
		MI.Datacenter = map[string][]string{}
	}
	if MI.Rack == nil {
		MI.Rack = map[string][]string{}
	}
	if MI.Role == nil {
		MI.Role = map[string][]string{}
	}
	if MI.Cluster == nil {
		MI.Cluster = map[string][]string{}
	}
	if MI.IPv4 == nil {
		MI.IPv4 = map[string]string{}
	}
	if MI.IPv6 == nil {
		MI.IPv6 = map[string]string{}
	}
}

// AddIndex adds new machine into the index
func AddIndex(val []byte) {
	var mc Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		log.ErrorExit(err)
	}
	initMI()

	MI.Product[mc.Product] = append(MI.Product[mc.Product], mc.Serial)
	MI.Datacenter[mc.Datacenter] = append(MI.Datacenter[mc.Datacenter], mc.Serial)
	mcrack := fmt.Sprint(mc.Rack)
	MI.Rack[mcrack] = append(MI.Rack[mcrack], mc.Serial)
	MI.Role[mc.Role] = append(MI.Role[mc.Role], mc.Serial)
	MI.Cluster[mc.Cluster] = append(MI.Cluster[mc.Cluster], mc.Serial)
	for _, ifn := range mc.Network {
		for k, v := range ifn.(map[string]interface{}) {
			if k == "ipv4" {
				for _, ip := range v.([]interface{}) {
					MI.IPv4[ip.(string)] = mc.Serial
				}
			}
			if k == "ipv6" {
				for _, ip := range v.([]interface{}) {
					MI.IPv6[ip.(string)] = mc.Serial
				}
			}
		}
	}
	for k, v := range mc.BMC {
		if k == "ipv4" {
			for _, ip := range v.([]interface{}) {
				MI.IPv4[ip.(string)] = mc.Serial
			}
		}
		if k == "ipv6" {
			for _, ip := range v.([]interface{}) {
				MI.IPv6[ip.(string)] = mc.Serial
			}
		}
	}
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
func DeleteIndex(val []byte) {
	var mc Machine
	err := json.Unmarshal(val, &mc)
	if err != nil {
		log.ErrorExit(err)
		return
	}
	initMI()

	i := indexOf(MI.Product[mc.Product], mc.Serial)
	MI.Product[mc.Product] = append(MI.Product[mc.Product][:i], MI.Product[mc.Product][i+1:]...)
	i = indexOf(MI.Datacenter[mc.Datacenter], mc.Serial)
	MI.Datacenter[mc.Datacenter] = append(MI.Datacenter[mc.Datacenter][:i], MI.Datacenter[mc.Datacenter][i+1:]...)
	mcrack := fmt.Sprint(mc.Rack)
	i = indexOf(MI.Rack[mcrack], mc.Serial)
	MI.Rack[mcrack] = append(MI.Rack[mcrack][:i], MI.Rack[mcrack][i+1:]...)
	i = indexOf(MI.Role[mc.Role], mc.Serial)
	MI.Role[mc.Role] = append(MI.Role[mc.Role][:i], MI.Role[mc.Role][i+1:]...)
	i = indexOf(MI.Cluster[mc.Cluster], mc.Serial)
	MI.Cluster[mc.Cluster] = append(MI.Cluster[mc.Cluster][:i], MI.Cluster[mc.Cluster][i+1:]...)
	for _, ifn := range mc.Network {
		for k, v := range ifn.(map[string]interface{}) {
			if k == "ipv4" {
				for _, ip := range v.([]interface{}) {
					delete(MI.IPv4, ip.(string))
				}
			}
			if k == "ipv6" {
				for _, ip := range v.([]interface{}) {
					delete(MI.IPv6, ip.(string))
				}
			}
		}
	}
	for k, v := range mc.BMC {
		if k == "ipv4" {
			for _, ip := range v.([]interface{}) {
				delete(MI.IPv4, ip.(string))
			}
		}
		if k == "ipv6" {
			for _, ip := range v.([]interface{}) {
				delete(MI.IPv6, ip.(string))
			}
		}
	}
}

// UpdateIndex updates target machine on theindex
func UpdateIndex(pval []byte, nval []byte) {
	DeleteIndex(pval)
	AddIndex(nval)
}
