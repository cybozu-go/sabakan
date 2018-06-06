package web

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
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

func newBrokenTestImage() io.Reader {
	kernel := []byte("abcd")

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
	tw.Write(kernel)
	tw.Close()
	return buf
}

func testHandleImageIndexGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/images/coreos", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}
	var data sabakan.ImageIndex
	err := json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Error("len(data) != 0:", len(data))
	}

	archive := newTestImage("abcd", "efgh")
	err = m.Image.Upload(context.Background(), "coreos", "1234", archive)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/images/coreos", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 1 {
		t.Fatal("len(data) != 1:", len(data))
	}
	if data[0].ID != "1234" {
		t.Error("data[0].ID != \"1234\":", data[0].ID)
	}
}

func testHandleImagesGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	archive := newTestImage("abcd", "efgh")
	err := m.Image.Upload(context.Background(), "coreos", "1234", archive)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/images/coreos/1234", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/images/coreos/xyz", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCore != http.StatusNotFound:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/images/coreos/!!!!", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("resp.StatusCore != http.StatusBadRequest:", resp.StatusCode)
	}
}

func testHandleImagesPut(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	archive := newTestImage("abcd", "efgh")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/images/coreos/1234", archive)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Fatal("resp.StatusCore != http.StatusCreated:", resp.StatusCode)
	}

	archive = newTestImage("abcd", "efgh")
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/images/coreos/1234", archive)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusConflict {
		t.Fatal("resp.StatusCore != http.StatusConflict:", resp.StatusCode)
	}

	archive = newBrokenTestImage()
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/images/coreos/4567", archive)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("resp.StatusCore != http.StatusBadRequest:", resp.StatusCode)
	}
}

func testHandleImagesDelete(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/images/coreos/1234", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCore != http.StatusNotFound:", resp.StatusCode)
	}

	archive := newTestImage("abcd", "efgh")
	err := m.Image.Upload(context.Background(), "coreos", "1234", archive)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/images/coreos/1234", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}
}

func TestHandleImages(t *testing.T) {
	t.Run("GetIndex", testHandleImageIndexGet)
	t.Run("Get", testHandleImagesGet)
	t.Run("Put", testHandleImagesPut)
	t.Run("Delete", testHandleImagesDelete)
}
