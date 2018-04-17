package sabakan

import (
	"context"
	"path"
	"testing"
)

func TestGetMachinesBySerial(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	mcs, err := GetMachinesBySerial(ctx, &etcdClient, []string{"1234abcd"})
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
}

func TestGetMachineBySerial(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	_, err := GetMachineBySerial(ctx, &etcdClient, "1234abcd")
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
}

func TestGetMachineByIPv4(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	value := "10.0.0.33"
	_, err := GetMachineByIPv4(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "10.1.0.9"
	_, err = GetMachineByIPv4(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "0.0.0.0"
	_, err = GetMachineByIPv4(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Unknown error to get, ", err.Error())
	}
}

// IPv6 doesn't support yet.
func TestGetMachineByIPv6(t *testing.T) {
}

func TestGetMachinesByProduct(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	value := "R740"
	_, err := GetMachinesByProduct(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "R640"
	_, err = GetMachinesByProduct(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "BBBB"
	_, err = GetMachinesByProduct(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Unknown error to get, ", err.Error())
	}
}

func TestGetMachinesByDatacenter(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	value := "dc1"
	_, err := GetMachinesByDatacenter(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "dc"
	_, err = GetMachinesByDatacenter(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "ny"
	_, err = GetMachinesByDatacenter(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Unknown error to get, ", err.Error())
	}
}

func TestGetMachinesByRack(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	value := "2"
	_, err := GetMachinesByRack(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "22"
	_, err = GetMachinesByRack(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "aaa"
	_, err = GetMachinesByRack(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Unknown error to get, ", err.Error())
	}
}

func TestGetMachinesByRole(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	value := "boot"
	_, err := GetMachinesByRole(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "node"
	_, err = GetMachinesByRole(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "boott"
	_, err = GetMachinesByRole(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Unknown error to get, ", err.Error())
	}
}

func TestGetMachinesByCluster(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	mi, _ := Indexing(etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
	}

	ctx := context.Background()
	value := "apac"
	_, err := GetMachinesByCluster(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "emea"
	_, err = GetMachinesByCluster(ctx, &etcdClient, value)
	if err != nil {
		t.Fatal("Failed to get, ", err.Error())
	}
	value = "tokyo"
	_, err = GetMachinesByCluster(ctx, &etcdClient, value)
	if err == nil {
		t.Fatal("Unknown error to get, ", err.Error())
	}
}
