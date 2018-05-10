package etcd

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cybozu-go/cmd"
	"github.com/cybozu-go/sabakan"
)

func testRegister(t *testing.T) {
	d := testNewDriver(t)
	config := &sabakan.IPAMConfig{
		MaxRacks:       80,
		MaxNodesInRack: 28,
		NodeIPv4Offset: "10.0.0.0/24",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.10.0.0/24",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "1234abcd",
		},
		&sabakan.Machine{
			Serial: "5678efgh",
		},
	}

	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), t.Name()+"/machines/5678efgh")
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Kvs) != 1 {
		t.Error("machine was not saved")
	}

	var saved sabakan.MachineJSON
	err = json.Unmarshal(resp.Kvs[0].Value, &saved)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Network) != 3 {
		t.Errorf("unexpected assigned IP addresses: %v", len(saved.Network))
	}
	if *saved.NodeIndexInRack != uint32(1) {
		t.Errorf("node index of 2nd machine should be 1 but %v", *saved.NodeIndexInRack)
	}

	err = d.Register(context.Background(), machines)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}

}

func testQuery(t *testing.T) {
	d := testNewDriver(t)
	cmd.Go(d.Run)
	time.Sleep(1 * time.Millisecond)

	config := &sabakan.IPAMConfig{
		MaxRacks:       80,
		MaxNodesInRack: 28,
		NodeIPv4Offset: "10.0.0.0/24",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.10.0.0/24",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		&sabakan.Machine{Serial: "12345678", Product: "R630"},
		&sabakan.Machine{Serial: "12345679", Product: "R630"},
		&sabakan.Machine{Serial: "12345680", Product: "R730"},
	}
	time.Sleep(1 * time.Millisecond)
	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}

	q := sabakan.QueryBySerial("12345678")
	resp, err := d.Query(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected query result: %#v", resp)
	}
	if !q.Match(resp[0]) {
		t.Errorf("unexpected responsed machine: %#v", resp[0])
	}

	q = &sabakan.Query{Product: "R630"}
	resp, err = d.Query(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != 2 {
		t.Fatalf("unexpected query result: %#v", resp)
	}
	if !(q.Match(resp[0]) && q.Match(resp[1])) {
		t.Errorf("unexpected responsed machine: %#v", resp)
	}
}

func testDelete(t *testing.T) {
	d := testNewDriver(t)
	config := &sabakan.IPAMConfig{
		MaxRacks:       80,
		MaxNodesInRack: 28,
		NodeIPv4Offset: "10.0.0.0/24",
		NodeRackShift:  4,
		BMCIPv4Offset:  "10.10.0.0/24",
		BMCRackShift:   2,
		NodeIPPerNode:  3,
		BMCIPPerNode:   1,
	}
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "1234abcd",
		},
	}

	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}

	err = d.Delete(context.Background(), "1234abcd")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), t.Name()+"/machines/1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 0 {
		t.Error("machine was not deleted")
	}

	resp, err = d.client.Get(context.Background(), t.Name()+"/node-indices/0/00")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 1 {
		t.Error("node index was not released")
	}

	err = d.Delete(context.Background(), "1234abcd")
	if err != sabakan.ErrNotFound {
		if err == nil {
			t.Error("delete succeeded for already deleted machine")
		} else {
			t.Fatal(err)
		}
	}
}

func TestMachine(t *testing.T) {
	t.Run("Register", testRegister)
	t.Run("Query", testQuery)
	t.Run("Delete", testDelete)
}
