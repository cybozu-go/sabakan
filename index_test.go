package sabakan

import (
	"context"
	"encoding/json"
	"path"
	"testing"
)

func TestIndexing(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	ctx := context.Background()
	_, err := Indexing(ctx, etcd, prefix)
	if err != nil {
		t.Fatal("Failed to create index")
	}
}

func TestAddIndex(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	ctx := context.Background()
	mi, _ := Indexing(ctx, etcd, prefix)
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
}

func TestDeleteIndex(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	ctx := context.Background()
	mi, _ := Indexing(ctx, etcd, prefix)
	etcdClient := EtcdClient{etcd, prefix, mi}
	PostConfig(etcdClient)
	mcs, _ := PostMachines(etcdClient)

	for i := 0; i < 2; i++ {
		etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err := mi.AddIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to add index, ", err.Error())
		}
		etcd.Delete(context.Background(), path.Join(prefix, EtcdKeyMachines, mcs[i].Serial))
		err = mi.DeleteIndex(etcdResp.Kvs[0].Value)
		if err != nil {
			t.Fatal("Failed to delete index, ", err.Error())
		}
	}
}

func TestUpdateIndex(t *testing.T) {
	etcd, _ := newEtcdClient()
	defer etcd.Close()
	prefix := path.Join(*flagEtcdPrefix, t.Name())
	ctx := context.Background()
	mi, _ := Indexing(ctx, etcd, prefix)
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

	new, _ := json.Marshal(map[string]interface{}{
		"serial":              "1234abcd",
		"product":             "bbbb",
		"datacenter":          "cccc",
		"rack":                2,
		"node-number-of-rack": 1,
		"role":                "dddd",
		"cluster":             "eeee",
	})

	etcdResp, _ := etcd.Get(context.Background(), path.Join(prefix, EtcdKeyMachines, "1234abcd"))
	prev := etcdResp.Kvs[0].Value
	etcd.Put(context.Background(), path.Join(prefix, EtcdKeyMachines, "1234abcd"), string(new))
	err := mi.UpdateIndex(prev, new)
	if err != nil {
		t.Fatal("Failed to update index, ", err.Error())
	}
}
