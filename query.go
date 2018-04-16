package sabakan

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
)

// GetMachinesBySerial returns values of the etcd keys by serial
func GetMachinesBySerial(e *EtcdClient, ctx context.Context, ss []string) ([]Machine, error) {
	var mcs []Machine
	key := path.Join(e.Prefix, EtcdKeyMachines)
	resp, err := e.Client.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	for _, s := range ss {
		var mc Machine
		for _, m := range resp.Kvs {
			_, k := path.Split(string(m.Key))
			if k == s {
				err = json.Unmarshal(m.Value, &mc)
				if err != nil {
					e.MI.mux.Unlock()
					return nil, err
				}
				mcs = append(mcs, mc)
				break
			}
		}
	}
	if len(mcs) == 0 {
		return nil, fmt.Errorf(ErrorMachineNotExists)
	}
	return mcs, nil
}

// GetMachineBySerial returns a value of the etcd key by serial
func GetMachineBySerial(e *EtcdClient, ctx context.Context, q string) (Machine, error) {
	mcs, err := GetMachinesBySerial(e, ctx, []string{q})
	if err != nil {
		return Machine{}, err
	}
	return mcs[0], nil
}

// GetMachineByIPv4 returns type Machine from the etcd and serial by IPv4
func GetMachineByIPv4(e *EtcdClient, ctx context.Context, q string) (Machine, error) {
	mcs, err := GetMachinesBySerial(e, ctx, []string{e.MI.IPv4[q]})
	if err != nil {
		return Machine{}, err
	}
	return mcs[0], err
}

// GetMachineByIPv6 returns type Machine from the etcd and serial by IPv6
func GetMachineByIPv6(e *EtcdClient, ctx context.Context, q string) (Machine, error) {
	mcs, err := GetMachinesBySerial(e, ctx, []string{e.MI.IPv6[q]})
	if err != nil {
		return Machine{}, err
	}
	return mcs[0], err
}

// GetMachinesByProduct returns type []Machine from the etcd and serial by product
func GetMachinesByProduct(e *EtcdClient, ctx context.Context, q string) ([]Machine, error) {
	return GetMachinesBySerial(e, ctx, e.MI.Product[q])
}

// GetMachinesByDatacenter returns type []Machine from the etcd and serial by datacenter
func GetMachinesByDatacenter(e *EtcdClient, ctx context.Context, q string) ([]Machine, error) {
	return GetMachinesBySerial(e, ctx, e.MI.Datacenter[q])
}

// GetMachinesByRack returns type []Machine from the etcd and serial by rack
func GetMachinesByRack(e *EtcdClient, ctx context.Context, q string) ([]Machine, error) {
	return GetMachinesBySerial(e, ctx, e.MI.Rack[q])
}

// GetMachinesByRole returns type []Machine from the etcd and serial by role
func GetMachinesByRole(e *EtcdClient, ctx context.Context, q string) ([]Machine, error) {
	return GetMachinesBySerial(e, ctx, e.MI.Role[q])
}

// GetMachinesByCluster returns type []Machine from the etcd and serial by cluster
func GetMachinesByCluster(e *EtcdClient, ctx context.Context, q string) ([]Machine, error) {
	return GetMachinesBySerial(e, ctx, e.MI.Cluster[q])
}
