package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cybozu-go/sabakan"
	"github.com/cybozu-go/sabakan/models/mock"
)

func testHandleAPIV1(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := Server{Model: m}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/v1/config/ipam", nil)
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatal("resp.StatusCode != http.StatusForbidden:", resp.StatusCode)
	}
}

func testHandlePermission(t *testing.T) {
	t.Parallel()

	_, ipnet, err := net.ParseCIDR("127.0.0.1/32")
	if err != nil {
		t.Fatal(err)
	}
	s := Server{AllowedRemotes: []*net.IPNet{ipnet}}

	cases := []struct {
		remote string
		method string
		path   string
	}{
		{"10.69.0.4", "GET", "/api/v1/config/ipam"},
		{"127.0.0.1", "POST", "/api/v1/config/ipam"},
		{"10.69.0.4", "GET", "/api/v1/crypts/1234/abc"},
		{"10.69.0.4", "PUT", "/api/v1/crypts/1234/abc"},
		{"10.69.0.4", "GET", "/api/v1/boot/coreos/kernel"},
	}
	for _, c := range cases {
		remote := c.remote + ":11111"
		r := &http.Request{RemoteAddr: remote, Method: c.method, URL: &url.URL{Path: c.path}}
		if !s.hasPermission(r) {
			t.Errorf("!hasPermission(r) == false; r=%v", c)
		}
	}

	cases = []struct {
		remote string
		method string
		path   string
	}{
		{"10.69.0.4", "POST", "/api/v1/config/ipam"},
		{"10.69.0.4", "PUT", "/api/v1/images/coreos/123.456"},
		{"10.69.0.4", "DELETE", "/api/v1/crypts/1234"},
	}
	for _, c := range cases {
		remote := c.remote + ":11111"
		r := &http.Request{RemoteAddr: remote, Method: c.method, URL: &url.URL{Path: c.path}}
		if s.hasPermission(r) {
			t.Errorf("hasPermission(r) == true; r=%v", c)
		}
	}
}

func testAuditContext(t *testing.T) {
	t.Parallel()

	m := mock.NewModel()
	handler := newTestServer(m)

	good := `
{
   "max-nodes-in-rack": 28,
   "node-ipv4-pool": "10.69.0.0/20",
   "node-ipv4-range-size": 6,
   "node-ipv4-range-mask": 26,
   "node-index-offset": 3,
   "node-ip-per-node": 3,
   "bmc-ipv4-pool": "10.72.16.0/20",
   "bmc-ipv4-range-size": 5,
   "bmc-ipv4-range-mask": 20
}
`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(good))
	handler.ServeHTTP(w, r)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatal("request failed with " + http.StatusText(resp.StatusCode))
	}

	buf := new(bytes.Buffer)
	err := m.Log.Dump(context.Background(), time.Time{}, time.Time{}, buf)
	if err != nil {
		t.Fatal(err)
	}

	a := new(sabakan.AuditLog)
	err = json.Unmarshal(buf.Bytes(), a)
	if err != nil {
		t.Fatal(err)
	}

	if a.IP != "192.0.2.1" {
		t.Error(`a.IP != "192.0.2.1"`, a.IP)
	}
	if a.Category != sabakan.AuditIPAM {
		t.Error(`a.Category != sabakan.AuditIPAM`, a.Category)
	}
	if a.User != "" {
		t.Error(`a.User != ""`, a.User)
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", "/api/v1/config/ipam", strings.NewReader(good))
	r.Header.Set(HeaderSabactlUser, "cybozu")
	handler.ServeHTTP(w, r)

	buf.Reset()
	err = m.Log.Dump(context.Background(), time.Time{}, time.Time{}, buf)
	if err != nil {
		t.Fatal(err)
	}

	a = new(sabakan.AuditLog)
	err = json.Unmarshal(buf.Bytes(), a)
	if err != nil {
		t.Fatal(err)
	}

	if a.User != "cybozu" {
		t.Error(`a.User != "cybozu"`, a.User)
	}
}

func TestServeHTTP(t *testing.T) {
	t.Run("APIV1", testHandleAPIV1)
	t.Run("Permission", testHandlePermission)
	t.Run("AuditContext", testAuditContext)
}
