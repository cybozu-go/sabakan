package etcd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func testRegister(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		sabakan.NewMachine(
			sabakan.MachineSpec{
				Serial: "1234abcd",
				Role:   "worker",
			}),
		sabakan.NewMachine(
			sabakan.MachineSpec{
				Serial: "5678efgh",
				Role:   "worker",
			}),
	}

	err = d.machineRegister(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch // wait for initialization of rack#0 node-indices
	<-ch

	resp, err := d.client.Get(context.Background(), KeyMachines+"5678efgh")
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
	if len(saved.Spec.IPv4) != int(testIPAMConfig.NodeIPPerNode) {
		t.Errorf("unexpected assigned IP addresses: %v", len(saved.Spec.IPv4))
	}
	if saved.Spec.IndexInRack != testIPAMConfig.NodeIndexOffset+2 {
		t.Errorf("node index of 2nd worker should be %v but %v", testIPAMConfig.NodeIndexOffset+2, saved.Spec.IndexInRack)
	}

	err = d.machineRegister(context.Background(), machines)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}
	// no need to wait; failed registration does not modify etcd,
	// so it does not generate event

	bootServer := []*sabakan.Machine{
		sabakan.NewMachine(
			sabakan.MachineSpec{
				Serial: "00000000",
				Role:   "boot",
			}),
	}
	bootServer2 := []*sabakan.Machine{
		sabakan.NewMachine(
			sabakan.MachineSpec{
				Serial: "00000001",
				Role:   "boot",
			}),
	}

	err = d.machineRegister(context.Background(), bootServer)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	resp, err = d.client.Get(context.Background(), KeyMachines+"00000000")
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(resp.Kvs[0].Value, &saved)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Spec.IndexInRack != testIPAMConfig.NodeIndexOffset {
		t.Errorf("node index of boot server should be %v but %v", testIPAMConfig.NodeIndexOffset, saved.Spec.IndexInRack)
	}

	err = d.machineRegister(context.Background(), bootServer2)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}
}

func testGet(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345678"}),
	}
	err = d.machineRegister(ctx, machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	_, err = d.machineGet(ctx, "a")
	if err != sabakan.ErrNotFound {
		t.Error(`err != sabakan.ErrNotFound`)
	}

	m, err := d.machineGet(ctx, "12345678")
	if err != nil {
		t.Fatal(err)
	}

	if m.Spec.Serial != "12345678" {
		t.Error(`m.Spec.Serial != "12345678"`)
	}
}

func testQuery(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)

	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345678", Product: "R630", Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345679", Product: "R630", Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345680", Product: "R730", Role: "worker"}),
	}
	err = d.machineRegister(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

	q := sabakan.Query{"serial": "12345678"}
	resp, err := d.machineQuery(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected query result: %#v", resp)
	}
	if !q.Match(resp[0]) {
		t.Errorf("unexpected responsed machine: %#v", resp[0])
	}

	q = sabakan.Query{"product": "R630"}
	resp, err = d.machineQuery(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != 2 {
		t.Fatalf("unexpected query result: %#v", resp)
	}
	if !(q.Match(resp[0]) && q.Match(resp[1])) {
		t.Errorf("unexpected responsed machine: %#v", resp)
	}

	q = sabakan.Query{}
	resp, err = d.machineQuery(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp) != 3 {
		t.Fatalf("unexpected query result: %#v", resp)
	}
}

func testSetState(t *testing.T) {
	d, ch := testNewDriver(t)
	ctx := context.Background()

	config := &testIPAMConfig
	err := d.putIPAMConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345678", Product: "R630", Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345679", Product: "R630", Role: "worker"}),
		sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345680", Product: "R730", Role: "worker"}),
	}
	err = d.machineRegister(ctx, machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

	m, err := d.machineGet(ctx, "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if m.Status.State != sabakan.StateHealthy {
		t.Error("m.Status.State == sabakan.StateHealthy:", m.Status.State)
	}
	err = d.machineSetState(ctx, "12345678", sabakan.StateDead)
	if err != nil {
		t.Fatal(err)
	}

	m, err = d.machineGet(ctx, "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if m.Status.State != sabakan.StateDead {
		t.Error("m.Status.State == sabakan.StateDead:", m.Status.State)
	}
}

func testDelete(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{
		sabakan.NewMachine(
			sabakan.MachineSpec{
				Serial: "1234abcd",
				Role:   "worker",
			}),
	}

	// register one
	err = d.machineRegister(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

	// delete one
	err = d.machineDelete(context.Background(), "1234abcd")
	if err == nil {
		t.Error("non-retired machine should not be deleted")
	}

	err = d.machineSetState(context.Background(), "1234abcd", sabakan.StateRetiring)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	err = d.machineSetState(context.Background(), "1234abcd", sabakan.StateRetired)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	err = d.machineDelete(context.Background(), "1234abcd")
	if err != nil {
		t.Error(err)
	}
	<-ch

	// confirm deletion
	resp, err := d.client.Get(context.Background(), KeyMachines+"1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 0 {
		t.Error("machine was not deleted")
	}

	// double delete
	err = d.machineDelete(context.Background(), "1234abcd")
	if err != sabakan.ErrNotFound {
		if err == nil {
			t.Error("delete succeeded for already deleted machine")
		} else {
			t.Fatal(err)
		}
	}

	// register after delete
	err = d.machineRegister(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	resp, err = d.client.Get(context.Background(), KeyMachines+"1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 1 {
		t.Error("failed to register machine after delete")
	}
}

func testDeleteRace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	d, ch := testNewDriver(t)
	cfg := &testIPAMConfig
	err := d.putIPAMConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	machines := []*sabakan.Machine{sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "1234abcd", Role: "worker",
	})}

	// prepare data to be deleted
	err = d.machineRegister(ctx, machines)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	<-ch

	m, rev, err := d.machineGetWithRev(ctx, "1234abcd")
	if err != nil {
		t.Fatal(err)
	}

RETRY:
	// retrieve usage data #1 and #2 with revision, and update data on memory
	usage1, err := d.getRackIndexUsage(ctx, m.Spec.Rack)
	if err != nil {
		t.Fatal(err)
	}
	usage1.release(m)

	usage2, err := d.getRackIndexUsage(ctx, m.Spec.Rack)
	if err != nil {
		t.Fatal(err)
	}
	usage2.release(m)

	// update data#2 on etcd; this increments revision
	resp2, err := d.machineDoDelete(ctx, m, rev, usage2)
	if err != nil {
		t.Fatal(err)
	}
	if !resp2.Succeeded {
		goto RETRY
	}
	<-ch

	// try to update data#1 on etcd; this must fail
	resp1, err := d.machineDoDelete(ctx, m, rev, usage1)
	if err != nil {
		t.Fatal(err)
	}
	if resp1.Succeeded {
		t.Error("update operations should fail, if revision number has been changed")
	}
}

func TestMachine(t *testing.T) {
	t.Run("Register", testRegister)
	t.Run("Get", testGet)
	t.Run("Query", testQuery)
	t.Run("SetState", testSetState)
	t.Run("Delete", testDelete)
	t.Run("DeleteRace", testDeleteRace)
}
