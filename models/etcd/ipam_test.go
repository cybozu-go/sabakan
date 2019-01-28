package etcd

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/cybozu-go/sabakan"
)

var testIPAMConfig = sabakan.IPAMConfig{
	MaxNodesInRack:   28,
	NodeIPv4Pool:     "10.69.0.0/20",
	NodeIPv4Offset:   "",
	NodeRangeSize:    6,
	NodeRangeMask:    26,
	NodeIndexOffset:  3,
	NodeIPPerNode:    3,
	BMCIPv4Pool:      "10.72.16.0/20",
	BMCIPv4Offset:    "0.0.1.0",
	BMCRangeSize:     5,
	BMCRangeMask:     20,
	BMCGatewayOffset: 1,
}

func testIPAMPutConfig(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	resp, err := d.client.Get(context.Background(), KeyIPAM)
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

	err = d.machineRegister(context.Background(),
		[]*sabakan.Machine{sabakan.NewMachine(
			sabakan.MachineSpec{Serial: "1234abcd", Role: "worker"})})
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	err = d.putIPAMConfig(context.Background(), config)
	if err == nil {
		t.Error("should be failed, because some machine is already registered")
	}
	<-ch
}

func testIPAMGetConfig(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	config := &testIPAMConfig

	bytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.client.Put(context.Background(), KeyIPAM, string(bytes))
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	actual, err := d.getIPAMConfig()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(config, actual) {
		t.Errorf("unexpected loaded config %#v", actual)
	}
}

func TestIPAM(t *testing.T) {
	t.Run("Put", testIPAMPutConfig)
	t.Run("Get", testIPAMGetConfig)
}
