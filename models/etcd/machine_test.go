package etcd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func testRegister(t *testing.T) {
	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

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
	<-ch // wait for initialization of rack#0 node-indices
	<-ch

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
	if len(saved.Network) != int(testIPAMConfig.NodeIPPerNode) {
		t.Errorf("unexpected assigned IP addresses: %v", len(saved.Network))
	}
	if saved.IndexInRack != testIPAMConfig.NodeIndexOffset+2 {
		t.Errorf("node index of 2nd worker should be %v but %v", testIPAMConfig.NodeIndexOffset+2, saved.IndexInRack)
	}

	err = d.Register(context.Background(), machines)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}
	// no need to wait; failed registration does not modify etcd,
	// so it does not generate event

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
	<-ch

	resp, err = d.client.Get(context.Background(), t.Name()+KeyMachines+"/00000000")
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &saved)
	if err != nil {
		t.Fatal(err)
	}
	if saved.IndexInRack != testIPAMConfig.NodeIndexOffset {
		t.Errorf("node index of boot server should be %v but %v", testIPAMConfig.NodeIndexOffset, saved.IndexInRack)
	}

	err = d.Register(context.Background(), bootServer2)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}
}

func testQuery(t *testing.T) {
	d, ch := testNewDriver(t)

	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		&sabakan.Machine{Serial: "12345678", Product: "R630", Role: "worker"},
		&sabakan.Machine{Serial: "12345679", Product: "R630", Role: "worker"},
		&sabakan.Machine{Serial: "12345680", Product: "R730", Role: "worker"},
	}
	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

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
	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		&sabakan.Machine{
			Serial: "1234abcd",
			Role:   "worker",
		},
	}

	// register one
	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

	// delete one
	err = d.Delete(context.Background(), "1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	// confirm deletion
	resp, err := d.client.Get(context.Background(), t.Name()+KeyMachines+"/1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 0 {
		t.Error("machine was not deleted")
	}

	// double delete
	err = d.Delete(context.Background(), "1234abcd")
	if err != sabakan.ErrNotFound {
		if err == nil {
			t.Error("delete succeeded for already deleted machine")
		} else {
			t.Fatal(err)
		}
	}

	// register after delete
	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	resp, err = d.client.Get(context.Background(), t.Name()+KeyMachines+"/1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 1 {
		t.Error("failed to register machine after delete")
	}
}

func testRegisterRace(t *testing.T) {
	d, ch := testNewDriver(t)
	cfg := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{&sabakan.Machine{
		Serial: "1234abcd", Role: "worker",
	}}

RETRY:
	wmcs1, usageMap1, err := d.updateMachines(context.Background(), machines, cfg)
	if err != nil {
		t.Fatal(err)
	}
	wmcs2, usageMap2, err := d.updateMachines(context.Background(), machines, cfg)
	if err != nil {
		t.Fatal(err)
	}
	tresp2, err := d.doRegister(context.Background(), wmcs2, usageMap2)
	if err != nil {
		t.Fatal(err)
	}
	if !tresp2.Succeeded {
		goto RETRY
	}
	<-ch

	tresp1, err := d.doRegister(context.Background(), wmcs1, usageMap1)
	if err != nil {
		t.Fatal(err)
	}
	if tresp1.Succeeded {
		t.Error("update operations should be failed, if revision number has been changed")
	}
}

func testDeleteRace(t *testing.T) {
	d, ch := testNewDriver(t)
	cfg := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{&sabakan.Machine{
		Serial: "1234abcd", Role: "worker",
	}}

	err = d.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

	m := machines[0]

RETRY:
	usage1, err := d.getRackIndexUsage(context.Background(), m.Rack)
	if err != nil {
		t.Fatal(err)
	}
	usage1.release(m)

	usage2, err := d.getRackIndexUsage(context.Background(), m.Rack)
	if err != nil {
		t.Fatal(err)
	}
	usage2.release(m)

	resp2, err := d.doDelete(context.Background(), m, usage2)
	if err != nil {
		t.Fatal(err)
	}
	if !resp2.Succeeded {
		goto RETRY
	}
	<-ch

	resp1, err := d.doDelete(context.Background(), m, usage1)
	if err != nil {
		t.Fatal(err)
	}
	if resp1.Succeeded {
		t.Error("update operations should be failed, if revision number has been changed")
	}
}

func TestMachine(t *testing.T) {
	t.Run("Register", testRegister)
	t.Run("Query", testQuery)
	t.Run("Delete", testDelete)
	t.Run("RegisterRace", testRegisterRace)
	t.Run("DeleteRace", testDeleteRace)
}
