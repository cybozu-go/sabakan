package etcd

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLastRev(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	dirName, err := ioutil.TempDir("", "sabakan-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = dirName
	defer os.RemoveAll(dirName)

	rev := d.loadLastRev()
	if rev != 0 {
		t.Error("initial revision must be zero")
	}

	err = d.saveLastRev(123)
	if err != nil {
		t.Fatal(err)
	}

	rev = d.loadLastRev()
	if rev != 123 {
		t.Error("saved revision cannot be loaded")
	}

	err = d.saveLastRev(0)
	if err != nil {
		t.Fatal(err)
	}

	rev = d.loadLastRev()
	if rev != 0 {
		t.Error("failed to reset the revision")
	}
}
