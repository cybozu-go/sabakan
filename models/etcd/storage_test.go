package etcd

import (
	"context"
	"sort"
	"testing"

	"github.com/cybozu-go/sabakan"
)

func TestStorage(t *testing.T) {
	d, _ := testNewDriver(t)

	err := d.PutEncryptionKey(context.Background(), "1234", "abcd-efgh", []byte("data"))
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

	data, err = d.GetEncryptionKey(context.Background(), "1234", "abcd-efgh")
	if err != nil {
		t.Fatal(err)
	}
	if data != nil {
		t.Errorf("not deleted: %s", string(data))
	}
}
