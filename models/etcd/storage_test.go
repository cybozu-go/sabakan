package etcd

import (
	"context"
	"sort"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func TestStorage(t *testing.T) {
	t.Parallel()

	d, ch := testNewDriver(t)
	_, err := initializeTestData(d, ch)
	if err != nil {
		t.Fatal(err)
	}

	err = d.PutEncryptionKey(context.Background(), "12345678", "abcd-efgh", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutEncryptionKey(context.Background(), "12345678", "abcd-efgh", []byte("data"))
	if err != sabakan.ErrConflicted {
		t.Error("not conflicted: ", err)
	}
	err = d.PutEncryptionKey(context.Background(), "12345678", "asdf-hjkl", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	err = d.PutEncryptionKey(context.Background(), "123456789", "asdf-hjkl", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}

	data, err := d.GetEncryptionKey(context.Background(), "12345678", "abcd-efgh")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "data" {
		t.Errorf("invalid data: %s", string(data))
	}

	_, err = d.DeleteEncryptionKeys(context.Background(), "12345678")
	if err == nil {
		t.Error("encryption keys should be deleted only for non-retiring machines")
	}

	err = d.machineSetState(context.Background(), "12345678", sabakan.StateRetiring)
	if err != nil {
		t.Fatal(err)
	}

	paths, err := d.DeleteEncryptionKeys(context.Background(), "12345678")
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

	m3, err := d.machineGet(context.Background(), "12345678")
	if err != nil {
		t.Fatal(err)
	}
	if m3.Status.State != sabakan.StateRetired {
		t.Error(`m3.Status.State != sabakan.StateRetired`)
	}

	err = d.PutEncryptionKey(context.Background(), "12345678", "abcd-efgh", []byte("data"))
	if err == nil {
		t.Error("encryption keys should not be added to retired machines")
	}

	data, err = d.GetEncryptionKey(context.Background(), "12345678", "abcd-efgh")
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Errorf("not deleted: %s", string(data))
	}
}
