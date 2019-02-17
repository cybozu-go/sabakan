package web

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/sabakan/v2/models/mock"
)

func testStateGet(t *testing.T) {
	ctx := context.Background()
	m := mock.NewModel()
	handler := Server{Model: m}

	machine := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "123",
	})
	err := m.Machine.Register(ctx, []*sabakan.Machine{machine})
	if err != nil {
		t.Fatal(err)
	}

	testData := []struct {
		serial string
		status int
		state  string
	}{
		{"123", 200, "uninitialized"},
		{"not-exist", 404, ""},
	}

	for _, td := range testData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path.Join("/api/v1/state", td.serial), nil)
		handler.ServeHTTP(w, r)
		resp := w.Result()

		if resp.StatusCode != td.status {
			t.Error("wrong status code, expects:", td.status, ", actual:", resp.StatusCode)
		}
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		state := string(data)
		if len(td.state) != 0 && td.state != state {
			t.Error("wrong state, expects:", td.state, ", actual:", state)
		}
	}
}

func testStatePut(t *testing.T) {
	ctx := context.Background()
	m := mock.NewModel()
	handler := newTestServer(m)

	machine := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: "123",
	})
	err := m.Machine.Register(ctx, []*sabakan.Machine{machine})
	if err != nil {
		t.Fatal(err)
	}

	testData := []struct {
		state  string
		status int
	}{
		{"uninitialized", 200},
		{"healthy", 200},
		{"updating", 200},
		{"unhealthy", 500},
		{"uninitialized", 200},
		{"healthy", 200},
		{"unreachable", 200},
		{"healthy", 200},
		{"unreachable", 200},
		{"unhealthy", 500},
		{"healthy", 200},
		{"unhealthy", 200},
		{"healthy", 500},
		{"unreachable", 500},
		{"uninitialized", 200},
		{"healthy", 200},
		{"retired", 400},
		{"retiring", 200},
		{"healthy", 500},
	}

	for _, td := range testData {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("PUT", "/api/v1/state/123", strings.NewReader(td.state))
		handler.ServeHTTP(w, r)

		resp := w.Result()
		if resp.StatusCode != td.status {
			t.Error("wrong status code, state:", td.state, "expects:", td.status, ", actual:", resp.StatusCode)
		}

		stored, err := m.Machine.Get(ctx, "123")
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/state/456", strings.NewReader("healthy"))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Error("wrong status code, expects: NotFound, actual:", resp.StatusCode)
	}
}

func TestState(t *testing.T) {
	t.Run("Get", testStateGet)
	t.Run("Put", testStatePut)
}
