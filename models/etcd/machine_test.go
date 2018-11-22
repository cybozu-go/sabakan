package etcd

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan"
)

func testRegister(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	machines, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := d.client.Get(context.Background(), KeyMachines+"12345679")
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
	if !saved.Spec.RetireDate.Equal(time.Date(2018, time.November, 22, 1, 2, 3, 0, time.UTC)) {
		t.Error("retire-date is not saved:", saved.Spec.RetireDate)
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
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

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
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

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

	q = sabakan.Query{"labels": "product=R630"}
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
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	m, err := d.machineGet(ctx, "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if m.Status.State != sabakan.StateUninitialized {
		t.Error("m.Status.State == sabakan.StateUninitialized:", m.Status.State)
	}
	err = d.machineSetState(ctx, "12345678", sabakan.StateHealthy)
	if err != nil {
		t.Fatal(err)
	}

	m, err = d.machineGet(ctx, "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if m.Status.State != sabakan.StateHealthy {
		t.Error("m.Status.State == sabakan.StateHealthy:", m.Status.State)
	}
}

func testAddLabels(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	err = d.machineAddLabels(context.Background(), "12345678", map[string]string{"datacenter": "heaven"})
	if err != nil {
		t.Fatal(err)
	}

	m, err := d.machineGet(context.Background(), "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if dc, ok := m.Spec.Labels["datacenter"]; !ok || dc != "heaven" {
		t.Error("wrong labels:", m.Spec.Labels)
	}

	err = d.machineAddLabels(context.Background(), "1111", map[string]string{"datacenter": "heaven"})
	if err != sabakan.ErrNotFound {
		if err != nil {
			t.Fatal(err)
		}
		t.Error("AddLabels succeeded for non-existing machine")
	}
}

func testDeleteLabel(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	err = d.machineDeleteLabel(context.Background(), "12345678", "product")
	if err != nil {
		t.Fatal(err)
	}

	m, err := d.machineGet(context.Background(), "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Spec.Labels["product"]; ok {
		t.Error("label was not deleted correctly:", m.Spec.Labels)
	}

	err = d.machineDeleteLabel(context.Background(), "12345678", "datacenter")
	if err != sabakan.ErrNotFound {
		if err != nil {
			t.Fatal(err)
		}
		t.Error("DeleteLabel succeeded for non-existing label")
	}

	err = d.machineDeleteLabel(context.Background(), "1111", "product")
	if err != sabakan.ErrNotFound {
		if err != nil {
			t.Fatal(err)
		}
		t.Error("DeleteLabel succeeded for non-existing machine")
	}
}

func testSetRetireDate(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	date := time.Date(2023, time.February, 28, 9, 9, 0, 0, time.UTC)
	err = d.machineSetRetireDate(context.Background(), "12345678", date)
	if err != nil {
		t.Fatal(err)
	}

	m, err := d.machineGet(context.Background(), "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if !m.Spec.RetireDate.Equal(date) {
		t.Error("retire-date was not set:", m.Spec.RetireDate)
	}
}

func testDelete(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	machines, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	// delete one
	err = d.machineDelete(context.Background(), "12345678")
	if err == nil {
		t.Error("non-retired machine should not be deleted")
	}

	err = d.machineSetState(context.Background(), "12345678", sabakan.StateRetiring)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	err = d.machineSetState(context.Background(), "12345678", sabakan.StateRetired)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	err = d.machineDelete(context.Background(), "12345678")
	if err != nil {
		t.Error(err)
	}
	<-ch

	// confirm deletion
	resp, err := d.client.Get(context.Background(), KeyMachines+"12345678")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 0 {
		t.Error("machine was not deleted")
	}

	// double delete
	err = d.machineDelete(context.Background(), "12345678")
	if err != sabakan.ErrNotFound {
		if err == nil {
			t.Error("delete succeeded for already deleted machine")
		} else {
			t.Fatal(err)
		}
	}

	// register after delete
	err = d.machineRegister(context.Background(), machines[:1])
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	resp, err = d.client.Get(context.Background(), KeyMachines+"12345678")
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
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	m, rev, err := d.machineGetWithRev(ctx, "12345678")
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
	t.Run("AddLabels", testAddLabels)
	t.Run("DeleteLabel", testDeleteLabel)
	t.Run("SetRetireDate", testSetRetireDate)
	t.Run("Delete", testDelete)
	t.Run("DeleteRace", testDeleteRace)
}
