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

func testConfigDHCPGet(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	config := &sabakan.DHCPConfig{
		GatewayOffset: 100,
	}

	err := m.DHCP.PutConfig(context.Background(), config)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v1/config/dhcp", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("resp.StatusCode != http.StatusOK:", resp.StatusCode)
	}

	var result sabakan.DHCPConfig
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(&result, config) {
		t.Errorf("wrong config: %#v", result)
	}
}

func testConfigDHCPPut(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	bad := "{}"
	bad2 := `
{
   "gateway-offset": 100,
   "lease-minutes": 30,
   "dns-servers": ["10.0.0.1", "::1"]
}
`
	good := `
{
   "gateway-offset": 100
}
`
	good2 := `
{
   "gateway-offset": 100,
   "lease-minutes": 30,
   "dns-servers": ["10.0.0.1", "10.0.0.2"]
}
`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/config/dhcp", strings.NewReader(bad))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest")
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/config/dhcp", strings.NewReader(bad2))
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Error("resp.StatusCode != http.StatusBadRequest")
	}

	r = httptest.NewRequest("PUT", "/api/v1/config/dhcp", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	conf, err := m.DHCP.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	expected := &sabakan.DHCPConfig{
		GatewayOffset: 100,
	}
	if !reflect.DeepEqual(conf, expected) {
		t.Errorf("mismatch: %#v", conf)
	}

	r = httptest.NewRequest("PUT", "/api/v1/config/dhcp", strings.NewReader(good))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	r = httptest.NewRequest("PUT", "/api/v1/config/dhcp", strings.NewReader(good2))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	resp = w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	conf, err = m.DHCP.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	expected = &sabakan.DHCPConfig{
		GatewayOffset: 100,
		LeaseMinutes:  30,
		DNSServers:    []string{"10.0.0.1", "10.0.0.2"},
	}
	if !reflect.DeepEqual(conf, expected) {
		t.Errorf("mismatch: %#v", conf)
	}
}

func TestConfigDHCP(t *testing.T) {
	t.Run("Get", testConfigDHCPGet)
	t.Run("Put", testConfigDHCPPut)
}
