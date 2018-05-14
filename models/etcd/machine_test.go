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
	config := &defaultTestConfig
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "1234abcd",
			Role:   "worker",
		},
		&sabakan.Machine{
			Serial: "5678efgh",
			Role:   "worker",
		},
	}

	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), t.Name()+KeyMachines+"/5678efgh")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 1 {
		t.Error("machine was not saved")
	}

	var saved sabakan.Machine
	err = json.Unmarshal(resp.Kvs[0].Value, &saved)
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Network) != int(defaultTestConfig.NodeIPPerNode) {
		t.Errorf("unexpected assigned IP addresses: %v", len(saved.Network))
	}
	if saved.IndexInRack != defaultTestConfig.NodeIndexOffset+2 {
		t.Errorf("node index of 2nd worker should be %v but %v", defaultTestConfig.NodeIndexOffset+2, saved.IndexInRack)
	}

	err = d.Register(context.Background(), machines)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}

	bootServer := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "00000000",
			Role:   "boot",
		},
	}
	bootServer2 := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "00000001",
			Role:   "boot",
		},
	}

	err = d.Register(context.Background(), bootServer)
	if err != nil {
		t.Fatal(err)
	}

	resp, err = d.client.Get(context.Background(), t.Name()+KeyMachines+"/00000000")
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &saved)
	if err != nil {
		t.Fatal(err)
	}
	if saved.IndexInRack != defaultTestConfig.NodeIndexOffset {
		t.Errorf("node index of boot server should be %v but %v", defaultTestConfig.NodeIndexOffset, saved.IndexInRack)
	}

	err = d.Register(context.Background(), bootServer2)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}
}

func testQuery(t *testing.T) {
	d := testNewDriver(t)
	cmd.Go(d.Run)
	time.Sleep(1 * time.Millisecond)

	config := &defaultTestConfig
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		&sabakan.Machine{Serial: "12345678", Product: "R630", Role: "worker"},
		&sabakan.Machine{Serial: "12345679", Product: "R630", Role: "worker"},
		&sabakan.Machine{Serial: "12345680", Product: "R730", Role: "worker"},
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
	config := &defaultTestConfig
	err := d.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	machines := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "1234abcd",
			Role:   "worker",
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

	resp, err := d.client.Get(context.Background(), t.Name()+KeyMachines+"/1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 0 {
		t.Error("machine was not deleted")
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
