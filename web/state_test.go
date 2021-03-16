package web

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan/v2"
	"github.com/cybozu-go/sabakan/v2/models/mock"
)

type testTransition struct {
	state  string
	status int
}

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
		data, err := io.ReadAll(resp.Body)
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

func testStatePutFromUninitialized(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 200},
		{"healthy", 200},
		{"unhealthy", 500},
		{"unreachable", 500},
		{"updating", 500},
		{"retiring", 200},
		{"retired", 500},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		init := testTransition{"uninitialized", 200}
		setStateRequest(serial, "", init, handler, t)
		setStateRequest(serial, "uninitialized", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutFromHealthy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 500},
		{"healthy", 200},
		{"unhealthy", 200},
		{"unreachable", 200},
		{"updating", 200},
		{"retiring", 200},
		{"retired", 500},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		setStateRequest(serial, "", testTransition{"uninitialized", 200}, handler, t)
		setStateRequest(serial, "uninitialized", testTransition{"healthy", 200}, handler, t)
		setStateRequest(serial, "healthy", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutFromUnhealthy(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 500},
		{"healthy", 200},
		{"unhealthy", 200},
		{"unreachable", 200},
		{"updating", 500},
		{"retiring", 200},
		{"retired", 500},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		setStateRequest(serial, "", testTransition{"uninitialized", 200}, handler, t)
		setStateRequest(serial, "uninitialized", testTransition{"healthy", 200}, handler, t)
		setStateRequest(serial, "healthy", testTransition{"unhealthy", 200}, handler, t)
		setStateRequest(serial, "unhealthy", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutFromUnreachable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 500},
		{"healthy", 200},
		{"unhealthy", 500},
		{"unreachable", 200},
		{"updating", 500},
		{"retiring", 200},
		{"retired", 500},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		setStateRequest(serial, "", testTransition{"uninitialized", 200}, handler, t)
		setStateRequest(serial, "uninitialized", testTransition{"healthy", 200}, handler, t)
		setStateRequest(serial, "healthy", testTransition{"unreachable", 200}, handler, t)
		setStateRequest(serial, "unreachable", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutFromUpdating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 200},
		{"healthy", 500},
		{"unhealthy", 500},
		{"unreachable", 500},
		{"updating", 200},
		{"retiring", 500},
		{"retired", 500},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		setStateRequest(serial, "", testTransition{"uninitialized", 200}, handler, t)
		setStateRequest(serial, "uninitialized", testTransition{"healthy", 200}, handler, t)
		setStateRequest(serial, "healthy", testTransition{"updating", 200}, handler, t)
		setStateRequest(serial, "updating", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutFromRetiring(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 500},
		{"healthy", 500},
		{"unhealthy", 500},
		{"unreachable", 500},
		{"updating", 500},
		{"retiring", 200},
		{"retired", 200},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		setStateRequest(serial, "", testTransition{"uninitialized", 200}, handler, t)
		setStateRequest(serial, "uninitialized", testTransition{"healthy", 200}, handler, t)
		setStateRequest(serial, "healthy", testTransition{"retiring", 200}, handler, t)
		setStateRequest(serial, "retiring", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutFromRetired(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testData := []testTransition{
		{"uninitialized", 200},
		{"healthy", 500},
		{"unhealthy", 500},
		{"unreachable", 500},
		{"updating", 500},
		{"retiring", 500},
		{"retired", 200},
	}

	serial := "123"
	for _, td := range testData {
		m, handler := setupMock(ctx, serial, t)
		setStateRequest(serial, "", testTransition{"uninitialized", 200}, handler, t)
		setStateRequest(serial, "uninitialized", testTransition{"healthy", 200}, handler, t)
		setStateRequest(serial, "healthy", testTransition{"retiring", 200}, handler, t)
		setStateRequest(serial, "retiring", testTransition{"retired", 200}, handler, t)
		setStateRequest(serial, "retired", td, handler, t)

		stored, err := m.Machine.Get(ctx, serial)
		if err != nil {
			t.Fatal(err)
		}
		storedState := string(stored.Status.State)
		if td.status == http.StatusOK && td.state != storedState {
			t.Error("stored state is wrong, expect:", td.state, ", actual:", storedState)
		}
	}
}

func testStatePutRetiredWithEncryptionKey(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m, handler := setupMock(ctx, "123", t)

	setStateRequest("123", "", testTransition{"uninitialized", 200}, handler, t)
	setStateRequest("123", "uninitialized", testTransition{"healthy", 200}, handler, t)

	err := m.Storage.PutEncryptionKey(ctx, "123", "path", []byte("aaa"))
	if err != nil {
		t.Fatal(err)
	}
	setStateRequest("123", "healthy", testTransition{"retiring", 200}, handler, t)

	// when the machine still has encryption keys, the machine is not permitted to transition to retired
	setStateRequest("123", "retiring", testTransition{"retired", 400}, handler, t)
}

func testStatePutOnNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	_, handler := setupMock(ctx, "123", t)
	setStateRequest("456", "", testTransition{"uninitialized", 404}, handler, t)
}

func setStateRequest(serial, from string, td testTransition, handler *Server, t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/state/"+serial, strings.NewReader(td.state))
	handler.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != td.status {
		t.Error("wrong status code, from", from, "to", td.state, ". expects:", td.status, ", actual:", resp.StatusCode)
	}
}

func setupMock(ctx context.Context, serial string, t *testing.T) (sabakan.Model, *Server) {
	m := mock.NewModel()
	handler := newTestServer(m)
	machine := sabakan.NewMachine(sabakan.MachineSpec{
		Serial: serial,
	})
	err := m.Machine.Register(ctx, []*sabakan.Machine{machine})
	if err != nil {
		t.Fatal(err)
	}
	return m, handler
}

func TestState(t *testing.T) {
	t.Run("Get", testStateGet)
	t.Run("PutFromUninitialized", testStatePutFromUninitialized)
	t.Run("PutFromHealthy", testStatePutFromHealthy)
	t.Run("PutFromUnhealthy", testStatePutFromUnhealthy)
	t.Run("PutFromUnreachable", testStatePutFromUnreachable)
	t.Run("PutFromUpdating", testStatePutFromUpdating)
	t.Run("PutFromRetiring", testStatePutFromRetiring)
	t.Run("PutFromRetired", testStatePutFromRetired)
	t.Run("PutRetiredWithEncryptionKey", testStatePutRetiredWithEncryptionKey)
	t.Run("PutOnNotFound", testStatePutOnNotFound)
}
