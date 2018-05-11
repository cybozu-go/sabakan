package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func testConfigGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{m}

	config := &sabakan.DefaultTestConfig

	err := m.Config.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/config", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	var result sabakan.IPAMConfig
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&result, config) {
		t.Errorf("wrong config: %#v", result)
	}
}

func testConfigPut(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{m}

	bad := "{}"
	good := `
{
   "max-racks": 80,
   "max-nodes-in-rack": 28,
   "node-ipv4-offset": "10.69.0.0/26",
   "node-rack-shift": 6,
   "node-index-offset": 3,
   "bmc-ipv4-offset": "10.72.17.0/27",
   "bmc-rack-shift": 5,
   "node-ip-per-node": 3,
   "bmc-ip-per-node": 1
}
`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/config", strings.NewReader(bad))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode == http.StatusBadRequest")
	}

	r = httptest.NewRequest("PUT", "/api/v1/config", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	conf, err := m.Config.GetConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	expected := &sabakan.IPAMConfig{
		MaxRacks:        80,
		MaxNodesInRack:  28,
		NodeIPv4Offset:  "10.69.0.0/26",
		NodeRackShift:   6,
		NodeIndexOffset: 3,
		BMCIPv4Offset:   "10.72.17.0/27",
		BMCRackShift:    5,
		NodeIPPerNode:   3,
		BMCIPPerNode:    1,
	}
	if !reflect.DeepEqual(conf, expected) {
		t.Errorf("mismatch: %#v", conf)
	}

	r = httptest.NewRequest("PUT", "/api/v1/config", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	machine := &sabakan.Machine{
		Serial: "1234",
	}
	err = m.Machine.Register(context.Background(), []*sabakan.Machine{machine})
	if err != nil {
		t.Fatal(err)
	}

	r = httptest.NewRequest("PUT", "/api/v1/config", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode == http.StatusOK {
		t.Error("request must not succeed")
	}
}

func TestConfig(t *testing.T) {
	t.Run("Get", testConfigGet)
	t.Run("Put", testConfigPut)
}
