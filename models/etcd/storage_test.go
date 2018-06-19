package etcd

import (
	"context"
	"sort"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func TestStorage(t *testing.T) {
	d, ch := testNewDriver(t)
	config := &testIPAMConfig
	err := d.putIPAMConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}
	<-ch
	m := sabakan.NewMachine(sabakan.MachineSpec{Serial: "1234"})
	m2 := sabakan.NewMachine(sabakan.MachineSpec{Serial: "12345"})
	err = d.machineRegister(context.Background(), []*sabakan.Machine{m, m2})
	if err != nil {
		t.Fatal(err)
	}
	<-ch

	err = d.PutEncryptionKey(context.Background(), "1234", "abcd-efgh", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutEncryptionKey(context.Background(), "1234", "abcd-efgh", []byte("data"))
	if err != sabakan.ErrConflicted {
		t.Error("not conflicted: ", err)
	}
	err = d.PutEncryptionKey(context.Background(), "1234", "asdf-hjkl", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutEncryptionKey(context.Background(), "12345", "asdf-hjkl", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := d.GetEncryptionKey(context.Background(), "1234", "abcd-efgh")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("invalid data: %s", string(data))
	}

	_, err = d.DeleteEncryptionKeys(context.Background(), "1234")
	if err == nil {
		t.Error("encryption keys should be deleted only for non-retiring machines")
	}

	err = d.machineSetState(context.Background(), "1234", sabakan.StateRetiring)
	if err != nil {
		t.Fatal(err)
	}

	paths, err := d.DeleteEncryptionKeys(context.Background(), "1234")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Error("wrong deleted paths", paths)
	}
	sort.Strings(paths)
	if paths[0] != "abcd-efgh" || paths[1] != "asdf-hjkl" {
		t.Error("wrong deleted paths", paths)
	}

	m3, err := d.machineGet(context.Background(), "1234")
	if err != nil {
		t.Fatal(err)
	}
	if m3.Status.State != sabakan.StateRetired {
		t.Error(`m3.Status.State != sabakan.StateRetired`)
	}

	err = d.PutEncryptionKey(context.Background(), "1234", "abcd-efgh", []byte("data"))
	if err == nil {
		t.Error("encryption keys should not be added to retired machines")
	}

	data, err = d.GetEncryptionKey(context.Background(), "1234", "abcd-efgh")
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Errorf("not deleted: %s", string(data))
	}
}
