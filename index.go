package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/log"
)

type MachinesIndex struct {
	Product    map[string][]string
	Datacenter map[string][]string
	Rack       map[string][]string
	Role       map[string][]string
	Cluster    map[string][]string
	IPv4       map[string]string
	IPv6       map[string]string
}

var MI MachinesIndex

func Indexing(client *clientv3.Client, prefix string) {
	key := path.Join(prefix, EtcdKeyMachines)
	resp, err := client.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		log.ErrorExit(err)
	}
	if resp.Count == 0 {
		return
	}

	MI.Product = map[string][]string{}
	MI.Datacenter = map[string][]string{}
	MI.Rack = map[string][]string{}
	MI.Role = map[string][]string{}
	MI.Cluster = map[string][]string{}
	MI.IPv4 = map[string]string{}
	MI.IPv6 = map[string]string{}
	for _, m := range resp.Kvs {
		var mc Machine
		err := json.Unmarshal(m.Value, &mc)
		if err != nil {
			log.ErrorExit(err)
		}

		MI.Product[mc.Product] = append(MI.Product[mc.Product], mc.Serial)
		MI.Datacenter[mc.Datacenter] = append(MI.Datacenter[mc.Datacenter], mc.Serial)
		MI.Rack[fmt.Sprint(mc.Rack)] = append(MI.Rack[fmt.Sprint(mc.Rack)], mc.Serial)
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
}

func GetMachineBySerial(e *etcdClient, r *http.Request, s string) ([]byte, error) {
	key := path.Join(e.prefix, EtcdKeyMachines, s)
	resp, err := e.client.Get(r.Context(), key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, fmt.Errorf("serial " + s + ErrorMachineNotFound)
	}
	return resp.Kvs[0].Value, nil
}

func GetMachineByIPv4(e *etcdClient, r *http.Request, q string) ([]byte, error) {
	mc, err := GetMachineBySerial(e, r, MI.IPv4[q])
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func GetMachineByIPv6(e *etcdClient, r *http.Request, q string) ([]byte, error) {
	mc, err := GetMachineBySerial(e, r, MI.IPv6[q])
	if err != nil {
		return nil, err
	}
	return mc, nil
}

func GetMachinesByProduct(e *etcdClient, r *http.Request, q string) ([]Machine, error) {
	var mcs []Machine
	for _, mc := range MI.Product[q] {
		j, err := GetMachineBySerial(e, r, mc)
		if err != nil {
			return nil, err
		}

		var rmc Machine
		err = json.Unmarshal(j, &rmc)
		if err != nil {
			return nil, err
		}
		mcs = append(mcs, rmc)
	}
	return mcs, nil
}

func GetMachinesByDatacenter(e *etcdClient, r *http.Request, q string) ([]Machine, error) {
	var mcs []Machine
	for _, mc := range MI.Datacenter[q] {
		j, err := GetMachineBySerial(e, r, mc)
		if err != nil {
			return nil, err
		}

		var rmc Machine
		err = json.Unmarshal(j, &rmc)
		if err != nil {
			return nil, err
		}
		mcs = append(mcs, rmc)
	}
	return mcs, nil
}

func GetMachinesByRack(e *etcdClient, r *http.Request, q string) ([]Machine, error) {
	var mcs []Machine
	for _, mc := range MI.Rack[q] {
		j, err := GetMachineBySerial(e, r, mc)
		if err != nil {
			return nil, err
		}

		var rmc Machine
		err = json.Unmarshal(j, &rmc)
		if err != nil {
			return nil, err
		}
		mcs = append(mcs, rmc)
	}
	return mcs, nil
}

func GetMachinesByRole(e *etcdClient, r *http.Request, q string) ([]Machine, error) {
	var mcs []Machine
	for _, mc := range MI.Role[q] {
		j, err := GetMachineBySerial(e, r, mc)
		if err != nil {
			return nil, err
		}

		var rmc Machine
		err = json.Unmarshal(j, &rmc)
		if err != nil {
			return nil, err
		}
		mcs = append(mcs, rmc)
	}
	return mcs, nil
}

func GetMachinesByCluster(e *etcdClient, r *http.Request, q string) ([]Machine, error) {
	var mcs []Machine
	for _, mc := range MI.Cluster[q] {
		j, err := GetMachineBySerial(e, r, mc)
		if err != nil {
			return nil, err
		}

		var rmc Machine
		err = json.Unmarshal(j, &rmc)
		if err != nil {
			return nil, err
		}
		mcs = append(mcs, rmc)
	}
	return mcs, nil
}
