package etcd

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/coreos/etcd/clientv3"
	"github.com/cybozu-go/sabakan"
)

var defaultTestConfig = sabakan.IPAMConfig{
	MaxRacks:        80,
	MaxNodesInRack:  28,
	NodeIPv4Offset:  "10.69.0.0/26",
	NodeRackShift:   6,
	NodeIndexOffset: 3,
	BMCIPv4Offset:   "10.72.17.0/27",
	BMCRackShift:    5,
	NodeIPPerNode:   3,
	BMCIPPerNode:    1,
}

func testPutConfig(t *testing.T) {
	d := testNewDriver(t)
	config := &defaultTestConfig
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), t.Name()+"/config")
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) != 1 {
		t.Error("config was not saved")
	}
	var actual sabakan.IPAMConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config, &actual) {
		t.Errorf("unexpected saved config %#v", actual)
	}

	resp, err = d.client.Get(context.Background(), t.Name()+"/node-indices", clientv3.WithPrefix())
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) != int(config.MaxRacks*(config.MaxNodesInRack+1)) {
		t.Errorf("number of node indices should be %d but %d", config.MaxRacks*(config.MaxNodesInRack+1), len(resp.Kvs))
	}

	resp, err = d.client.Get(context.Background(), t.Name()+"/node-indices/0/worker/04")
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) != 1 {
		t.Error("node index 0/worker/04 not found")
	}

	err = d.Register(context.Background(), []*sabakan.Machine{{Serial: "1234abcd", Role: "worker"}})
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutConfig(context.Background(), config)
	if err == nil {
		t.Error("should be failed, because some machine is already registered")
	}
}

func testGetConfig(t *testing.T) {
	d := testNewDriver(t)

	config := &defaultTestConfig

	bytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.client.Put(context.Background(), t.Name()+"/config", string(bytes))
	if err != nil {
		t.Fatal(err)
	}

	actual, err := d.GetConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, actual) {
		t.Errorf("unexpected loaded config %#v", actual)
	}
}

func TestConfig(t *testing.T) {
	t.Run("Put", testPutConfig)
	t.Run("Get", testGetConfig)
}
