package etcd

import (
	"context"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func testAssetGetIndex(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := ioutil.TempDir("", "sabakan-asset-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	index, err := d.assetGetIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 0 {
		t.Error("asset index should be empty")
	}

	_, err = d.assetPut(context.Background(), "foo", "text/plain", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.assetPut(context.Background(), "abc", "text/plain", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.assetPut(context.Background(), "xyz", "text/plain", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	index, err = d.assetGetIndex(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// index must be sorted
	if !reflect.DeepEqual(index, []string{"abc", "foo", "xyz"}) {
		t.Error("wrong asset index:", index)
	}
}

func TestAsset(t *testing.T) {
	t.Run("GetIndex", testAssetGetIndex)
}
