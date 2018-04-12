package sabakan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
)

// GetMachineBySerial returns a value of the etcd key by serial
func GetMachineBySerial(e *etcdClient, r *http.Request, s string) ([]byte, error) {
	key := path.Join(e.prefix, EtcdKeyMachines, s)
	resp, err := e.client.Get(r.Context(), key)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, fmt.Errorf("serial " + s + " " + ErrorMachineNotExists)
	}
	return resp.Kvs[0].Value, nil
}

// GetMachineByIPv4 returns type Machine from the etcd and serial by IPv4
func GetMachineByIPv4(e *etcdClient, r *http.Request, q string) ([]byte, error) {
	mc, err := GetMachineBySerial(e, r, MI.IPv4[q])
	if err != nil {
		return nil, err
	}
	return mc, nil
}

// GetMachineByIPv6 returns type Machine from the etcd and serial by IPv6
func GetMachineByIPv6(e *etcdClient, r *http.Request, q string) ([]byte, error) {
	mc, err := GetMachineBySerial(e, r, MI.IPv6[q])
	if err != nil {
		return nil, err
	}
	return mc, nil
}

// GetMachinesByProduct returns type []Machine from the etcd and serial by product
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

// GetMachinesByDatacenter returns type []Machine from the etcd and serial by datacenter
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

// GetMachinesByRack returns type []Machine from the etcd and serial by rack
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

// GetMachinesByRole returns type []Machine from the etcd and serial by role
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

// GetMachinesByCluster returns type []Machine from the etcd and serial by cluster
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
