package sabakan

import (
	"context"
	"encoding/json"
	"path"

	"github.com/coreos/etcd/clientv3"
)

// GetMachinesBySerial returns values of the etcd keys by serial
func GetMachinesBySerial(ctx context.Context, e *EtcdClient, ss []string) ([]Machine, error) {
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
					return nil, err
				}
				mcs = append(mcs, mc)
				break
			}
		}
	}
	return mcs, nil
}

// GetMachineByIPv4 returns type []Machine from the etcd and serial by IPv4
func GetMachinesByIPv4(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, []string{mi.IPv4[q]})
}

// GetMachineByIPv6 returns type []Machine from the etcd and serial by IPv6
func GetMachinesByIPv6(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, []string{mi.IPv6[q]})
}

// GetMachinesByProduct returns type []Machine from the etcd and serial by product
func GetMachinesByProduct(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, mi.Product[q])
}

// GetMachinesByDatacenter returns type []Machine from the etcd and serial by datacenter
func GetMachinesByDatacenter(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, mi.Datacenter[q])
}

// GetMachinesByRack returns type []Machine from the etcd and serial by rack
func GetMachinesByRack(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, mi.Rack[q])
}

// GetMachinesByRole returns type []Machine from the etcd and serial by role
func GetMachinesByRole(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, mi.Role[q])
}

// GetMachinesByCluster returns type []Machine from the etcd and serial by cluster
func GetMachinesByCluster(ctx context.Context, e *EtcdClient, q string) ([]Machine, error) {
	return GetMachinesBySerial(ctx, e, mi.Cluster[q])
}
