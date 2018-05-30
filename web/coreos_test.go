package web

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cybozu-go/sabakan/models/mock"
)

func testHandleCoreOSKernel(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/boot/coreos/kernel", nil)
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
	r = httptest.NewRequest("GET", "/api/v1/boot/coreos/kernel", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}

	//
	archive = newTestImage("opqr", "stu")
	err = m.Image.Upload(context.Background(), "coreos", "5678", archive)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/coreos/kernel", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "opqr" {
		t.Error("wrong content")
	}
}

func testHandleCoreOSInitRD(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/boot/coreos/initrd.gz", nil)
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
	r = httptest.NewRequest("GET", "/api/v1/boot/coreos/initrd.gz", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}

	//
	archive = newTestImage("opqr", "stu")
	err = m.Image.Upload(context.Background(), "coreos", "5678", archive)
	if err != nil {
		t.Fatal(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/coreos/initrd.gz", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCore != http.StatusOK:", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "stu" {
		t.Error("wrong content")
	}
}

func TestHandleCoreOS(t *testing.T) {
	t.Run("kernel", testHandleCoreOSKernel)
	t.Run("initrd", testHandleCoreOSInitRD)
}
