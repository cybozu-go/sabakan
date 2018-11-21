package web

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func TestRetireDate(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	m.Machine.Register(context.Background(), []*sabakan.Machine{
		sabakan.NewMachine(sabakan.MachineSpec{
			Serial:     "1234abcd",
			Rack:       1,
			Role:       "worker",
			BMC:        sabakan.MachineBMC{Type: "IPMI-2.0"},
			RetireDate: time.Date(2018, time.November, 22, 1, 2, 3, 0, time.UTC),
		}),
	})

	expected := time.Date(2023, time.February, 28, 9, 9, 9, 345, time.UTC)
	input := []byte(expected.Format(time.RFC3339Nano))
	badInput := []byte(`"9999"`)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/retire-date/", bytes.NewReader(input))
	handler.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/api/v1/retire-date/1234abcd", bytes.NewReader(input))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatal("resp.StatusCode != http.StatusMethodNotAllowed:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/retire-date/ufuf", bytes.NewReader(input))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatal("resp.StatusCode != http.StatusNotFound:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/retire-date/1234abcd", bytes.NewReader(badInput))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatal("resp.StatusCode != http.StatusBadRequest:", resp.StatusCode)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/retire-date/1234abcd", bytes.NewReader(input))
	handler.ServeHTTP(w, r)
	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	stored, err := m.Machine.Get(context.Background(), "1234abcd")
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Spec.RetireDate.Equal(expected) {
		t.Error("retire-date was not set: ", stored.Spec.RetireDate)
	}
}
