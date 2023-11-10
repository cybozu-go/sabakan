package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/sabakan/v3/models/mock"
)

func testHandleiPXE(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/boot/coreos/ipxe", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "chain") {
		t.Error("unexpected ipxe script:", string(body))
	}
}

func testHandleiPXEWithSerial(t *testing.T) {
	t.Parallel()

	machines := []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "2222abcd",
			Labels: map[string]string{
				"product":    "R630",
				"datacenter": "ty3",
			},
			Rack: 1,
			Role: "cs",
			BMC:  sabakan.MachineBMC{Type: "iDRAC-9"},
		}),
	}

	m := mock.NewModel()
	handler := newTestServer(m)
	err := m.Machine.Register(context.Background(), machines)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/boot/coreos/ipxe/2222abcd", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	tmpl := &sabakan.IgnitionTemplate{
		Version: sabakan.Ignition2_3,
	}
	err = m.Ignition.PutTemplate(context.Background(), "cs", "1.0.0", tmpl)
	if err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/kernel_params/coreos", strings.NewReader("console=ttyS0 coreos.autologin=ttyS0"))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/coreos/ipxe/2222abcd", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "kernel") {
		t.Error("unexpected ipxe script:", string(body))
	}
	if !strings.Contains(string(body), "console=ttyS0 coreos.autologin=ttyS0") {
		t.Error("kernel parameter is not contained", string(body))
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/boot/coreos/ipxe/1234abcd", nil)
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}
}

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

	data, err := io.ReadAll(resp.Body)
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "stu" {
		t.Error("wrong content")
	}
}

func TestHandleCoreOS(t *testing.T) {
	t.Run("iPXE", testHandleiPXE)
	t.Run("iPXEWithSerial", testHandleiPXEWithSerial)
	t.Run("kernel", testHandleCoreOSKernel)
	t.Run("initrd", testHandleCoreOSInitRD)
}
