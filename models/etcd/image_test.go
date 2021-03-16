package etcd

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan/v2"
)

func newTestImage(kernel, initrd string) io.Reader {

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	hdr := &tar.Header{
		Name: sabakan.ImageKernelFilename,
		Mode: 0644,
		Size: int64(len(kernel)),
	}
	err := tw.WriteHeader(hdr)
	if err != nil {
		panic(err)
	}
	tw.Write([]byte(kernel))

	hdr = &tar.Header{
		Name: sabakan.ImageInitrdFilename,
		Mode: 0644,
		Size: int64(len(initrd)),
	}
	err = tw.WriteHeader(hdr)
	if err != nil {
		panic(err)
	}
	tw.Write([]byte(initrd))
	tw.Close()
	return buf
}

func testImagePutIndex(t *testing.T, d *driver, index sabakan.ImageIndex, osName string) {
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatal(err)
	}

	key := path.Join(KeyImages, osName)
	_, err = d.client.Put(context.Background(), key, string(data))
	if err != nil {
		t.Fatal(err)
	}
}

func testImageGetIndex(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := os.MkdirTemp("", "sabakan-image-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	index, err := d.imageGetIndex(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if index == nil {
		t.Error("image index should not be nil")
	}
	if len(index) != 0 {
		t.Error("image index should be empty")
	}

	testIndex := sabakan.ImageIndex{
		&sabakan.Image{
			ID: "1234.5",
		},
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, testIndex, "coreos")

	index, err = d.imageGetIndex(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 2 {
		t.Fatal("wrong image index: ", len(index))
	}
	for i := 0; i < len(testIndex); i++ {
		if index[i].ID != testIndex[i].ID {
			t.Error("mismatch id", i, index[i].ID)
		}
		if index[i].Exists != false {
			t.Error("mismatch exists", i)
		}
	}

	err = os.MkdirAll(filepath.Join(tempdir, "images", "coreos", "1234.5"), 0755)
	if err != nil {
		t.Fatal(err)
	}
	index, err = d.imageGetIndex(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 2 {
		t.Fatal("wrong image index: ", len(index))
	}
	if index[0].Exists != true {
		t.Error("exists should be true")
	}
	if index[1].Exists != false {
		t.Error("exists should be false")
	}
}

func testImageGetInfoAll(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	testIndex := sabakan.ImageIndex{
		&sabakan.Image{
			ID: "1234.5",
		},
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, testIndex, "coreos")
	testImagePutIndex(t, d, testIndex, "ubuntu")

	images, err := d.imageGetInfoAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	found := make(map[string]int)
	for _, image := range images {
		found[image.ID]++
	}
	if found["1234.5"] != 2 || found["2234.6"] != 2 {
		t.Errorf("unexpected result from imageGetInfoAll(): %v", images)
	}
}

func testImageUpload(t *testing.T) {
	t.Parallel()

	archive := newTestImage("abcd", "efg")

	d, _ := testNewDriver(t)

	tempdir, err := os.MkdirTemp("", "sabakan-image-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	err = d.imageUpload(context.Background(), "coreos", "1234.5", archive)
	if err != nil {
		t.Fatal(err)
	}

	index, err := d.imageGetIndex(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != 1 {
		t.Error("index not registered")
	}

	if index[0].ID != "1234.5" {
		t.Error("ID mismatch", index[0].ID)
	}

	if !d.getImageDir("coreos").Exists("1234.5") {
		t.Error("image is not stored")
	}

	for i := 0; i < sabakan.MaxImages; i++ {
		archive = newTestImage("abcd", "efg")
		err = d.imageUpload(context.Background(), "coreos", fmt.Sprint(i), archive)
		if err != nil {
			t.Fatal(err)
		}
	}
	index, err = d.imageGetIndex(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(index) != sabakan.MaxImages {
		t.Error("index size should not exceed sabakan.MaxImages")
	}

	dels, _, err := d.imageGetDeletedWithRev(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(dels) != 1 {
		t.Fatal("deleted is not stored")
	}
	if dels[0] != "1234.5" {
		t.Error("wrong deleted image ID")
	}

	for i := 0; i < MaxDeleted; i++ {
		archive = newTestImage("abcd", "efg")
		err = d.imageUpload(context.Background(), "coreos", fmt.Sprint(i+10), archive)
		if err != nil {
			t.Fatal(err)
		}
	}
	dels, _, err = d.imageGetDeletedWithRev(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(dels) != MaxDeleted {
		t.Fatal("deleted should not exceed MaxDeleted")
	}
	if dels[0] != "0" {
		t.Error("wrong deleted image ID")
	}
}

func testImageDownload(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := os.MkdirTemp("", "sabakan-image-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	index := sabakan.ImageIndex{
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, index, "coreos")

	// case 1. ID is not in the index, no local copy.
	err = d.imageDownload(context.Background(), "coreos", "1234.5", io.Discard)
	if err != sabakan.ErrNotFound {
		t.Error(`err != sabakan.ErrNotFound`)
	}

	// case 2. ID is in the index, but no local copy.
	index = sabakan.ImageIndex{
		&sabakan.Image{
			ID: "1234.5",
		},
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, index, "coreos")
	err = d.imageDownload(context.Background(), "coreos", "1234.5", io.Discard)
	if err != sabakan.ErrNotFound {
		t.Error(`err != sabakan.ErrNotFound`)
	}

	// case 3. ID is in the index and a local copy exists.
	dir := d.getImageDir("coreos")
	err = dir.Extract(newTestImage("abc", "def"), "1234.5", []string{
		sabakan.ImageKernelFilename,
		sabakan.ImageInitrdFilename,
	})
	if err != nil {
		t.Fatal(err)
	}
	buf := new(bytes.Buffer)
	err = d.imageDownload(context.Background(), "coreos", "1234.5", buf)
	if err != nil {
		t.Fatal(err)
	}
	tr := tar.NewReader(buf)
	count := 0
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Error(err)
			break
		}
		expects := ""
		switch hdr.Name {
		case sabakan.ImageKernelFilename:
			count++
			expects = "abc"
		case sabakan.ImageInitrdFilename:
			count++
			expects = "def"
		default:
			t.Error("unexpected file in tar:", hdr.Name)
		}

		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != expects {
			t.Error(`string(data) != expects`, string(data), expects)
		}
	}
	if count != 2 {
		t.Error(`count != 2`)
	}

	// case 4. ID is not in the index, a local copy remains.
	index = sabakan.ImageIndex{
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, index, "coreos")
	err = d.imageDownload(context.Background(), "coreos", "1234.5", io.Discard)
	if err != sabakan.ErrNotFound {
		t.Error(`err != sabakan.ErrNotFound`)
	}
}

func testImageDelete(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := os.MkdirTemp("", "sabakan-image-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	index := sabakan.ImageIndex{
		&sabakan.Image{
			ID: "1234.5",
		},
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, index, "coreos")

	err = d.imageDelete(context.Background(), "coreos", "hoge")
	if err != sabakan.ErrNotFound {
		t.Error(`err != sabakan.ErrNotFound`)
	}

	err = d.imageDelete(context.Background(), "coreos", "1234.5")
	if err != nil {
		t.Fatal(err)
	}

	index2, err := d.imageGetIndex(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(index2) != 1 {
		t.Fatal(`len(index2) != 1`)
	}
	if index2[0].ID != "2234.6" {
		t.Error(`index2[0].ID != "2234.6"`, index2[0].ID)
	}

	deleted, _, err := d.imageGetDeletedWithRev(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted) != 1 {
		t.Fatal(`len(deleted) != 1`)
	}
	if deleted[0] != "1234.5" {
		t.Error(`deleted[0] != "1234.5"`)
	}

	index = sabakan.ImageIndex{}
	for i := 0; i < MaxDeleted; i++ {
		index = append(index, &sabakan.Image{
			ID: fmt.Sprint(i),
		})
	}
	testImagePutIndex(t, d, index, "coreos")

	for i := 0; i < MaxDeleted; i++ {
		err = d.imageDelete(context.Background(), "coreos", fmt.Sprint(i))
		if err != nil {
			t.Fatal(err)
		}
	}

	deleted, _, err = d.imageGetDeletedWithRev(context.Background(), "coreos")
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted) != MaxDeleted {
		t.Fatal(`len(deleted) != MaxDeleted`)
	}
	if deleted[0] != "0" {
		t.Error(`deleted[0] != "0"`)
	}
}

func testImageServeFile(t *testing.T) {
	t.Parallel()

	d, _ := testNewDriver(t)

	tempdir, err := os.MkdirTemp("", "sabakan-image-test")
	if err != nil {
		t.Fatal(err)
	}
	d.dataDir = tempdir
	defer os.RemoveAll(tempdir)

	index := sabakan.ImageIndex{
		&sabakan.Image{
			ID: "1234.5",
		},
		&sabakan.Image{
			ID: "2234.6",
		},
	}
	testImagePutIndex(t, d, index, "coreos")

	buf := new(bytes.Buffer)
	f := func(mt time.Time, content io.ReadSeeker) {
		buf.Reset()
		io.Copy(buf, content)
	}

	err = d.imageServeFile(context.Background(), "coreos", "kernel", f)
	if err != sabakan.ErrNotFound {
		t.Error(`err != sabakan.ErrNotFound`, err)
	}

	dir := d.getImageDir("coreos")
	err = dir.Extract(newTestImage("abc", "def"), "1234.5", []string{
		sabakan.ImageKernelFilename,
		sabakan.ImageInitrdFilename,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.imageServeFile(context.Background(), "coreos", sabakan.ImageKernelFilename, f)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "abc" {
		t.Error(`buf.String() != "abc"`, buf.String())
	}

	err = d.imageServeFile(context.Background(), "coreos", sabakan.ImageInitrdFilename, f)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "def" {
		t.Error(`buf.String() != "def"`, buf.String())
	}

	err = d.imageServeFile(context.Background(), "coreos", "no-such-file", f)
	if err == nil {
		t.Error("imageServeFile should return an error that causes an internal server error")
	}

	err = dir.Extract(newTestImage("zzzz", "3838"), "2234.6", []string{
		sabakan.ImageKernelFilename,
		sabakan.ImageInitrdFilename,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = d.imageServeFile(context.Background(), "coreos", sabakan.ImageKernelFilename, f)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "zzzz" {
		t.Error(`buf.String() != "zzzz"`, buf.String())
	}

	err = d.imageServeFile(context.Background(), "coreos", sabakan.ImageInitrdFilename, f)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "3838" {
		t.Error(`buf.String() != "3838"`, buf.String())
	}
}

func TestImage(t *testing.T) {
	t.Run("GetIndex", testImageGetIndex)
	t.Run("GetInfoAll", testImageGetInfoAll)
	t.Run("Upload", testImageUpload)
	t.Run("Download", testImageDownload)
	t.Run("Delete", testImageDelete)
	t.Run("ServeFile", testImageServeFile)
}
