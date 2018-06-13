package etcd

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan"
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
		t.Error("initial asset index should be empty")
	}

	_, err = d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.assetPut(context.Background(), "abc", "text/plain", "", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.assetPut(context.Background(), "xyz", "text/plain", "", strings.NewReader("bar"))
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

func testAssetGetInfo(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := ioutil.TempDir("", "sabakan-asset-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	_, err = d.assetGetInfo(context.Background(), "foo")
	if err != sabakan.ErrNotFound {
		t.Error("err != sabakan.ErrNotFound:", err)
	}

	_, err = d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	asset, err := d.assetGetInfo(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "foo" {
		t.Error("asset.Name != foo:", asset.Name)
	}
	if asset.ID != 1 {
		t.Error("asset.ID != 1:", asset.ID)
	}
	if asset.ContentType != "text/plain" {
		t.Error("asset.ContentType != text/plain:", asset.ContentType)
	}
	if asset.Sha256 != "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9" {
		t.Error("wrong Sha256:", asset.Sha256)
	}
	u := *d.advertiseURL
	u.Path = "/api/v1/assets/foo"
	if !reflect.DeepEqual(asset.URLs, []string{u.String()}) {
		t.Error("wrong URLs", asset.URLs)
	}
	if !asset.Exists {
		t.Error("file must exist locally")
	}

	// force local copy absent
	// TODO: this may cause watcher to panic, so stop it beforehand
	err = os.Remove(d.getAssetDir().Path(asset.ID))
	if err != nil {
		t.Fatal(err)
	}
	asset, err = d.assetGetInfo(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Exists {
		t.Error("file must not exist")
	}
}

func testAssetPut(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := ioutil.TempDir("", "sabakan-asset-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	// case 1. creation
	status, err := d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	if status.Status != http.StatusCreated {
		t.Error("status.Status != http.StatusCreated:", status.Status)
	}
	if status.ID != 1 {
		t.Error("status.ID != 1:", status.ID)
	}

	// check etcd data directly
	resp, err := d.client.Get(context.Background(), KeyAssets+"foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 1 {
		t.Fatal("asset not stored in etcd; len(resp.Kvs) != 1:", len(resp.Kvs))
	}
	asset := new(sabakan.Asset)
	err = json.Unmarshal(resp.Kvs[0].Value, asset)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "foo" {
		t.Error("asset.Name != foo:", asset.Name)
	}
	if asset.ID != 1 {
		t.Error("asset.ID != 1:", asset.ID)
	}
	if asset.ContentType != "text/plain" {
		t.Error("asset.ContentType != text/plain:", asset.ContentType)
	}
	if asset.Sha256 != "fcde2b2edba56bf408601fb721fe9b5c338d10ee429ea04fae5511b68fbf8fb9" {
		t.Error("wrong Sha256:", asset.Sha256)
	}
	u := *d.advertiseURL
	u.Path = "/api/v1/assets/foo"
	if !reflect.DeepEqual(asset.URLs, []string{u.String()}) {
		t.Error("wrong URLs", asset.URLs)
	}
	// asset.Exists in etcd has no meaning

	// check local file directly
	f1, err := os.Open(d.getAssetDir().Path(status.ID))
	if err != nil {
		t.Fatal(err)
	}
	defer f1.Close()
	buf, err := ioutil.ReadAll(f1)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != "bar" {
		t.Error("local copy corrupted")
	}

	// case 2. update
	status, err = d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("baz"))
	if err != nil {
		t.Fatal(err)
	}

	if status.Status != http.StatusOK {
		t.Error("status.Status != http.StatusOK:", status.Status)
	}
	if status.ID != 2 {
		t.Error("status.ID != 2:", status.ID)
	}

	resp, err = d.client.Get(context.Background(), KeyAssets+"foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 1 {
		t.Fatal("asset not stored in etcd; len(resp.Kvs) != 1:", len(resp.Kvs))
	}
	asset = new(sabakan.Asset)
	err = json.Unmarshal(resp.Kvs[0].Value, asset)
	if err != nil {
		t.Fatal(err)
	}
	if asset.Name != "foo" {
		t.Error("asset.Name != foo:", asset.Name)
	}
	if asset.ID != 2 {
		t.Error("asset.ID != 2:", asset.ID)
	}
	if asset.ContentType != "text/plain" {
		t.Error("asset.ContentType != text/plain:", asset.ContentType)
	}
	if asset.Sha256 != "baa5a0964d3320fbc0c6a922140453c8513ea24ab8fd0577034804a967248096" {
		t.Error("wrong Sha256:", asset.Sha256)
	}
	if !reflect.DeepEqual(asset.URLs, []string{u.String()}) {
		t.Error("wrong URLs", asset.URLs)
	}

	f2, err := os.Open(d.getAssetDir().Path(status.ID))
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()
	buf, err = ioutil.ReadAll(f2)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != "baz" {
		t.Error("local copy corrupted")
	}
}

type mockHandler struct {
	calledServeContent bool
	calledRedirect     bool
	err                error
	asset              *sabakan.Asset
	content            []byte
	redirectURL        string
}

func (h *mockHandler) ServeContent(asset *sabakan.Asset, content io.ReadSeeker) {
	h.calledServeContent = true
	h.asset = asset
	buf, err := ioutil.ReadAll(content)
	if err != nil {
		h.err = err
		return
	}
	h.content = buf
}

func (h *mockHandler) Redirect(url string) {
	h.calledRedirect = true
	h.redirectURL = url
}

func testAssetGet(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := ioutil.TempDir("", "sabakan-asset-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	h := new(mockHandler)
	err = d.assetGet(context.Background(), "foo", h)
	if err != sabakan.ErrNotFound {
		t.Error("err != sabakan.ErrNotFound:", err)
	}

	status, err := d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	h = new(mockHandler)
	err = d.assetGet(context.Background(), "foo", h)
	if err != nil {
		t.Fatal(err)
	}
	if h.err != nil {
		t.Fatal(err)
	}
	if !h.calledServeContent {
		t.Error("ServeContent() was not called")
	}
	if h.asset == nil || h.asset.ID != status.ID {
		t.Error("ServeContent() received wrong asset")
	}
	if string(h.content) != "bar" {
		t.Error("ServeContent() received wrong content")
	}

	// force local copy absent
	// TODO: this may cause watcher to panic, so stop it beforehand
	err = os.Remove(d.getAssetDir().Path(status.ID))
	if err != nil {
		t.Fatal(err)
	}

	h = new(mockHandler)
	err = d.assetGet(context.Background(), "foo", h)
	if err != nil {
		t.Fatal(err)
	}
	if h.err != nil {
		t.Fatal(err)
	}
	if !h.calledRedirect {
		t.Error("Redirect() was not called")
	}
	u := *d.advertiseURL
	u.Path = "/api/v1/assets/foo"
	// it's nonsense to redirect to myself; just for test
	if h.redirectURL != u.String() {
		t.Error("Redirect() received wrong URL:", h.redirectURL)
	}
}

func testAssetDelete(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := ioutil.TempDir("", "sabakan-asset-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	err = d.assetDelete(context.Background(), "foo")
	if err != sabakan.ErrNotFound {
		t.Error("err != sabakan.ErrNotFound:", err)
	}

	_, err = d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("bar"))
	if err != nil {
		t.Fatal(err)
	}

	err = d.assetDelete(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.assetGetInfo(context.Background(), "foo")
	if err != sabakan.ErrNotFound {
		t.Error("err != sabakan.ErrNotFound:", err)
	}

	// check etcd data directly
	resp, err := d.client.Get(context.Background(), KeyAssets+"foo")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Kvs) != 0 {
		t.Error("asset not deleted from etcd; len(resp.Kvs) != 0:", len(resp.Kvs))
	}

	// TODO: check local file directly; this needs watcher

	status, err := d.assetPut(context.Background(), "foo", "text/plain", "", strings.NewReader("baz"))
	if err != nil {
		t.Fatal(err)
	}
	if status.Status != http.StatusCreated {
		t.Error("status.Status != http.StatusCreated:", status.Status)
	}
}

func TestAsset(t *testing.T) {
	t.Run("GetIndex", testAssetGetIndex)
	t.Run("GetInfo", testAssetGetInfo)
	t.Run("Put", testAssetPut)
	t.Run("Get", testAssetGet)
	t.Run("Delete", testAssetDelete)
}
