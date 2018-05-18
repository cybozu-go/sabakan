package etcd

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cybozu-go/sabakan"
)

var testDHCPConfig = sabakan.DHCPConfig{
	GatewayOffset: 100,
}

func testDHCPPutConfig(t *testing.T) {
	d := testNewDriver(t)
	config := &testDHCPConfig
	err := d.putDHCPConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), t.Name()+KeyDHCP)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) != 1 {
		t.Error("config was not saved")
	}
	var actual sabakan.DHCPConfig
	err = json.Unmarshal(resp.Kvs[0].Value, &actual)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config, &actual) {
		t.Errorf("unexpected saved config %#v", actual)
	}
}

func testDHCPGetConfig(t *testing.T) {
	d := testNewDriver(t)

	config := &testDHCPConfig

	bytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.client.Put(context.Background(), t.Name()+KeyDHCP, string(bytes))
	if err != nil {
		t.Fatal(err)
	}

	actual, err := d.getDHCPConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(config, actual) {
		t.Errorf("unexpected loaded config %#v", actual)
	}
}

func TestDHCP(t *testing.T) {
	t.Run("Put", testDHCPPutConfig)
	t.Run("Get", testDHCPGetConfig)
}
