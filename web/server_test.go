package web

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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

func TestServeHTTP(t *testing.T) {
	t.Run("APIV1", testHandleAPIV1)
}

func TestHandlePermission(t *testing.T) {
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
		{"10.69.0.4", "GET", "/api/v1/crypts/abcd1234/virtio-pci-0000:00:05.0"},
		{"10.69.0.4", "PUT", "/api/v1/crypts/abcd1234/virtio-pci-0000:00:05.0"},
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
	}
	for _, c := range cases {
		remote := c.remote + ":11111"
		r := &http.Request{RemoteAddr: remote, Method: c.method, URL: &url.URL{Path: c.path}}
		if s.hasPermission(r) {
			t.Errorf("hasPermission(r) == true; r=%v", c)
		}
	}
}
