package etcd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func testRegister(t *testing.T) {
	d := testNewDriver(t)
	config := &sabakan.IPAMConfig{
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

	resp, err := d.client.Get(context.Background(), t.Name()+"/machines/1234abcd")
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

	err = d.Register(context.Background(), machines)
	if err != sabakan.ErrConflicted {
		t.Errorf("unexpected error: %v", err)
	}

}

func testQuery(t *testing.T) {
}

func TestMachine(t *testing.T) {
	t.Run("Register", testRegister)
	t.Run("Query", testQuery)
}
