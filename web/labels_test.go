package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan/v3"
	"github.com/cybozu-go/sabakan/v3/models/mock"
)

func testLabelsPut(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "1234abcd",
			Rack:   1,
			Role:   "worker",
			BMC:    sabakan.MachineBMC{Type: "IPMI-2.0"},
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/labels/1234abcd/datacenter", strings.NewReader("heaven"))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Error("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	stored, err := m.Machine.Get(context.Background(), "1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stored.Spec.Labels, map[string]string{"datacenter": "heaven"}) {
		t.Error("stored labels are wrong:", stored.Spec.Labels)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/labels/1234abcd/"+strings.Repeat("too-long", 8), strings.NewReader("heaven"))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/labels/1234abcd/datacenter", strings.NewReader(strings.Repeat("too-long", 8)))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/labels/5678efgh/datacenter", strings.NewReader("heaven"))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}
}

func testLabelsDelete(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial: "1234abcd",
			Labels: map[string]string{"datacenter": "heaven"},
			Rack:   1,
			Role:   "worker",
			BMC:    sabakan.MachineBMC{Type: "IPMI-2.0"},
		}),
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/v1/labels/1234abcd/datacenter", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	stored, err := m.Machine.Get(context.Background(), "1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := stored.Spec.Labels["datacenter"]; ok {
		t.Error("label was not deleted correctly:", stored.Spec.Labels)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/labels/1234abcd/product", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", "/api/v1/labels/5678efgh/datacenter", nil)
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}
}

func TestLabels(t *testing.T) {
	t.Run("Put", testLabelsPut)
	t.Run("Delete", testLabelsDelete)
}
